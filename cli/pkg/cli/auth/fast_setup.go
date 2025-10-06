package auth

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cline/cli/pkg/cli/config"
	"github.com/cline/cli/pkg/cli/setup"
	"github.com/cline/cli/pkg/generated"
)

// FastSetup performs quick provider setup with provider ID and optional API key
func FastSetup(providerID, apiKey string) error {
	// Validate and prompt for missing params
	validatedProviderID, validatedAPIKey, err := validateAndPromptParams(providerID, apiKey)
	if err != nil {
		return err
	}

	// Create config manager and registry
	configManager, err := config.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	registry, err := config.NewProviderRegistry()
	if err != nil {
		return fmt.Errorf("failed to create provider registry: %w", err)
	}

	// Load existing config or create new one
	if _, err := configManager.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get provider definition
	def, err := registry.GetProviderDefinition(validatedProviderID)
	if err != nil {
		return fmt.Errorf("invalid provider '%s': %w", validatedProviderID, err)
	}

	fmt.Printf("Configuring %s...\n", def.Name)

	// Create provider config
	providerConfig := config.ProviderConfig{
		ID:          validatedProviderID,
		Name:        def.Name,
		ExtraConfig: make(map[string]string),
	}

	// Set the API key
	if err := setAPIKeyForProvider(def, validatedAPIKey, &providerConfig); err != nil {
		return err
	}

	// For providers with multiple required fields, collect the rest
	if len(def.RequiredFields) > 1 {
		fmt.Println("\nAdditional required configuration:")
		for _, field := range def.RequiredFields {
			// Skip the API key field since we already have it
			if isAPIKeyField(field.Name) {
				continue
			}

			value, err := promptForField(field, true)
			if err != nil {
				return err
			}

			setup.MapFieldToConfig(field, value, &providerConfig)
		}
	}

	// Validate all required fields
	if err := setup.ValidateRequiredFields(def.ID, providerConfig, def.RequiredFields); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Ask about optional configuration
	configureOptional := false
	if len(def.OptionalFields) > 0 {
		prompt := &survey.Confirm{
			Message: "Would you like to configure optional settings?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &configureOptional); err != nil {
			return fmt.Errorf("failed to get optional config choice: %w", err)
		}

		if configureOptional {
			fmt.Println("\nOptional configuration:")
			for _, field := range def.OptionalFields {
				value, err := promptForField(field, false)
				if err != nil {
					return err
				}

				if value != "" {
					switch field.Name {
					case "baseUrl":
						providerConfig.BaseURL = value
					default:
						providerConfig.ExtraConfig[field.Name] = value
					}
				}
			}
		}
	}

	// Select model
	if err := selectModelForProvider(def, &providerConfig); err != nil {
		return err
	}

	// Validate configuration
	if err := registry.ValidateProviderConfig(providerConfig); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Add to config
	if err := configManager.AddProvider(providerConfig); err != nil {
		return fmt.Errorf("failed to add provider: %w", err)
	}

	// Save configuration
	if err := configManager.Save(configManager.GetConfig()); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("\nSuccessfully configured %s!\n", def.Name)
	fmt.Printf("Configuration saved to %s\n", configPath)

	return nil
}

// validateAndPromptParams validates provider ID and API key, prompting for missing values
func validateAndPromptParams(providerID, apiKey string) (string, string, error) {
	// Validate provider ID
	if providerID == "" {
		return "", "", fmt.Errorf("provider ID is required")
	}

	// Normalize provider ID (trim spaces, lowercase)
	providerID = strings.TrimSpace(strings.ToLower(providerID))

	// Check if provider exists
	registry, err := config.NewProviderRegistry()
	if err != nil {
		return "", "", fmt.Errorf("failed to create provider registry: %w", err)
	}

	def, err := registry.GetProviderDefinition(providerID)
	if err != nil {
		// Provider not found, show suggestions
		return "", "", fmt.Errorf("provider '%s' not found. Use 'cline auth' to see available providers", providerID)
	}

	// If API key is missing, prompt for it
	if apiKey == "" {
		fmt.Printf("Configuring %s\n", def.Name)
		
		// Find the API key field
		var apiKeyField *generated.ConfigField
		for _, field := range def.RequiredFields {
			if isAPIKeyField(field.Name) {
				apiKeyField = &field
				break
			}
		}

		if apiKeyField != nil {
			value, err := promptForField(*apiKeyField, true)
			if err != nil {
				return "", "", fmt.Errorf("failed to get API key: %w", err)
			}
			apiKey = value
		} else {
			return "", "", fmt.Errorf("provider %s does not have an API key field", providerID)
		}
	}

	return providerID, apiKey, nil
}

// setAPIKeyForProvider sets the API key in the provider config
func setAPIKeyForProvider(def *generated.ProviderDefinition, apiKey string, providerConfig *config.ProviderConfig) error {
	// Find the API key field
	for _, field := range def.RequiredFields {
		if isAPIKeyField(field.Name) {
			setup.MapFieldToConfig(field, apiKey, providerConfig)
			return nil
		}
	}

	return fmt.Errorf("provider %s does not have an API key field", def.ID)
}

// isAPIKeyField checks if a field name represents an API key
func isAPIKeyField(fieldName string) bool {
	lower := strings.ToLower(fieldName)
	return strings.Contains(lower, "apikey") || 
		strings.Contains(lower, "api_key") ||
		fieldName == "apiKey" ||
		fieldName == "key"
}

// promptForField prompts for a single configuration field
func promptForField(field generated.ConfigField, required bool) (string, error) {
	message := field.Name
	if field.Comment != "" {
		message += fmt.Sprintf(" (%s)", field.Comment)
	}
	if required {
		message += " *"
	}

	var value string
	var prompt survey.Prompt

	switch field.FieldType {
	case "password":
		prompt = &survey.Password{
			Message: message,
		}
	case "select":
		// For select fields, we'd need to define options based on the field
		// For now, treat as input
		prompt = &survey.Input{
			Message: message,
			Default: field.Placeholder,
		}
	default:
		prompt = &survey.Input{
			Message: message,
			Default: field.Placeholder,
		}
	}

	if err := survey.AskOne(prompt, &value); err != nil {
		return "", fmt.Errorf("failed to get field %s: %w", field.Name, err)
	}

	if required && value == "" {
		return "", fmt.Errorf("field %s is required", field.Name)
	}

	return value, nil
}

// selectModelForProvider helps user select a model for the provider
func selectModelForProvider(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
	if len(def.Models) == 0 {
		if def.HasDynamicModels {
			// For providers with dynamic models, ask for model ID
			var modelID string
			prompt := &survey.Input{
				Message: "Enter model ID (or leave empty for default):",
			}

			if err := survey.AskOne(prompt, &modelID); err != nil {
				return fmt.Errorf("failed to get model ID: %w", err)
			}

			providerConfig.ModelID = modelID
		}
		return nil
	}

	// Use default model if available
	if def.DefaultModelID != "" {
		providerConfig.ModelID = def.DefaultModelID
		if modelInfo, exists := def.Models[def.DefaultModelID]; exists {
			providerConfig.ModelInfo = config.ModelInfo{
				MaxTokens:      modelInfo.MaxTokens,
				ContextWindow:  modelInfo.ContextWindow,
				SupportsImages: modelInfo.SupportsImages,
				InputPrice:     modelInfo.InputPrice,
				OutputPrice:    modelInfo.OutputPrice,
				Description:    modelInfo.Description,
			}
		}
		fmt.Printf("Using default model: %s\n", def.DefaultModelID)
		return nil
	}

	// If no default, use the first available model
	for modelID, modelInfo := range def.Models {
		providerConfig.ModelID = modelID
		providerConfig.ModelInfo = config.ModelInfo{
			MaxTokens:      modelInfo.MaxTokens,
			ContextWindow:  modelInfo.ContextWindow,
			SupportsImages: modelInfo.SupportsImages,
			InputPrice:     modelInfo.InputPrice,
			OutputPrice:    modelInfo.OutputPrice,
			Description:    modelInfo.Description,
		}
		fmt.Printf("Using model: %s\n", modelID)
		break
	}

	return nil
}
