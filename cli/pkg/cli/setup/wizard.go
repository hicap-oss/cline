package setup

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cline/cli/pkg/cli/config"
	"github.com/cline/cli/pkg/generated"
)

// SetupWizard handles the interactive setup process
type SetupWizard struct {
	configManager *config.ConfigManager
	registry      *config.ProviderRegistry
}

// NewSetupWizard creates a new setup wizard
func NewSetupWizard() (*SetupWizard, error) {
	configManager, err := config.NewConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	registry, err := config.NewProviderRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to create provider registry: %w", err)
	}

	return &SetupWizard{
		configManager: configManager,
		registry:      registry,
	}, nil
}

// Run runs the complete setup wizard
func (sw *SetupWizard) Run() error {
	fmt.Println("🚀 Welcome to Cline CLI Setup!")
	fmt.Println("This wizard will help you configure API providers for the Cline CLI.")
	fmt.Println()

	// Check if config already exists
	exists, err := config.ConfigExists()
	if err != nil {
		return fmt.Errorf("failed to check config existence: %w", err)
	}

	if exists {
		overwrite := false
		prompt := &survey.Confirm{
			Message: "Configuration already exists. Do you want to add more providers or reconfigure?",
			Default: true,
		}
		if err := survey.AskOne(prompt, &overwrite); err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}

		if !overwrite {
			fmt.Println("Setup cancelled.")
			return nil
		}
	}

	// Load existing config or create new one
	if _, err := sw.configManager.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Show provider selection menu
	for {
		// Always get the current config from ConfigManager
		cliConfig := sw.configManager.GetConfig()

		action, err := sw.showMainMenu(cliConfig)
		if err != nil {
			return err
		}

		switch action {
		case "add":
			if err := sw.addProvider(); err != nil {
				fmt.Printf("❌ Error adding provider: %v\n", err)
				continue
			}
		case "remove":
			if err := sw.removeProvider(); err != nil {
				fmt.Printf("❌ Error removing provider: %v\n", err)
				continue
			}
		case "list":
			sw.listConfiguredProviders()
		case "test":
			if err := sw.testProviders(); err != nil {
				fmt.Printf("❌ Error testing providers: %v\n", err)
				continue
			}
		case "default":
			if err := sw.setDefaultProvider(); err != nil {
				fmt.Printf("❌ Error setting default provider: %v\n", err)
				continue
			}
		case "save":
			if err := sw.saveAndExit(); err != nil {
				return err
			}
			return nil
		case "exit":
			return nil
		}
	}
}

// showMainMenu displays the main setup menu
func (sw *SetupWizard) showMainMenu(cliConfig *config.CLIConfig) (string, error) {
	options := []string{
		"Add a new provider",
		"Remove a provider",
		"List configured providers",
		"Test provider connections",
		"Set default provider",
		"Save configuration and exit",
		"Exit without saving",
	}

	var choice string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: options,
	}

	if err := survey.AskOne(prompt, &choice); err != nil {
		return "", fmt.Errorf("failed to get menu choice: %w", err)
	}

	switch choice {
	case options[0]:
		return "add", nil
	case options[1]:
		return "remove", nil
	case options[2]:
		return "list", nil
	case options[3]:
		return "test", nil
	case options[4]:
		return "default", nil
	case options[5]:
		return "save", nil
	case options[6]:
		return "exit", nil
	default:
		return "", fmt.Errorf("invalid choice")
	}
}

// addProvider guides the user through adding a new provider
func (sw *SetupWizard) addProvider() error {
	fmt.Println("\n📋 Adding a new provider...")

	// Show provider selection
	providerID, err := sw.selectProvider()
	if err != nil {
		return err
	}

	// Get provider definition
	def, err := sw.registry.GetProviderDefinition(providerID)
	if err != nil {
		return err
	}

	fmt.Printf("\n🔧 Configuring %s\n", def.Name)
	fmt.Printf("📖 Setup instructions: %s\n\n", def.SetupInstructions)

	// Create provider config
	providerConfig := config.ProviderConfig{
		ID:          providerID,
		Name:        def.Name,
		ExtraConfig: make(map[string]string),
	}

	// Collect required fields
	if err := sw.collectRequiredFields(def, &providerConfig); err != nil {
		return err
	}

	// Collect optional fields
	if err := sw.collectOptionalFields(def, &providerConfig); err != nil {
		return err
	}

	// Select model
	if err := sw.selectModel(def, &providerConfig); err != nil {
		return err
	}

	// Validate configuration
	if err := sw.registry.ValidateProviderConfig(providerConfig); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Add to config
	if err := sw.configManager.AddProvider(providerConfig); err != nil {
		return fmt.Errorf("failed to add provider: %w", err)
	}

	fmt.Printf("✅ Successfully configured %s!\n", def.Name)
	return nil
}

// selectProvider shows provider selection interface
func (sw *SetupWizard) selectProvider() (string, error) {
	// Show selection method
	method := ""
	methodPrompt := &survey.Select{
		Message: "How would you like to choose a provider?",
		Options: []string{
			"View popular providers",
			"View all providers",
			"Browse by category",
			"Search providers",
			
		},
	}

	if err := survey.AskOne(methodPrompt, &method); err != nil {
		return "", fmt.Errorf("failed to get selection method: %w", err)
	}

	switch method {
	case "Browse by category":
		return sw.selectProviderByCategory()
	case "View popular providers":
		return sw.selectFromPopularProviders()
	case "Search providers":
		return sw.searchAndSelectProvider()
	case "View all providers":
		return sw.selectFromAllProviders()
	default:
		return "", fmt.Errorf("invalid selection method")
	}
}

// selectProviderByCategory shows providers grouped by category
func (sw *SetupWizard) selectProviderByCategory() (string, error) {
	categories := sw.registry.GetProvidersByCategory()

	// Select category
	categoryNames := make([]string, 0, len(categories))
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)

	var selectedCategory string
	categoryPrompt := &survey.Select{
		Message: "Select a category:",
		Options: categoryNames,
	}

	if err := survey.AskOne(categoryPrompt, &selectedCategory); err != nil {
		return "", fmt.Errorf("failed to get category: %w", err)
	}

	// Select provider from category
	providers := categories[selectedCategory]
	providerOptions := make([]string, len(providers))
	for i, providerID := range providers {
		providerOptions[i] = fmt.Sprintf("%s (%s)", sw.registry.GetProviderDisplayName(providerID), providerID)
	}

	var selectedProvider string
	providerPrompt := &survey.Select{
		Message: fmt.Sprintf("Select a provider from %s:", selectedCategory),
		Options: providerOptions,
	}

	if err := survey.AskOne(providerPrompt, &selectedProvider); err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Extract provider ID from selection (it's the last part in parentheses)
	// Format: "Display Name (provider-id)" or "Display Name (extra) (provider-id)"
	lastOpenParen := strings.LastIndex(selectedProvider, "(")
	lastCloseParen := strings.LastIndex(selectedProvider, ")")
	if lastOpenParen == -1 || lastCloseParen == -1 || lastCloseParen < lastOpenParen {
		return "", fmt.Errorf("invalid provider selection format")
	}
	providerID := strings.TrimSpace(selectedProvider[lastOpenParen+1 : lastCloseParen])

	return providerID, nil
}

// selectFromPopularProviders shows popular providers
func (sw *SetupWizard) selectFromPopularProviders() (string, error) {
	popular := sw.registry.GetPopularProviders()

	providerOptions := make([]string, len(popular))
	for i, providerID := range popular {
		providerOptions[i] = fmt.Sprintf("%s (%s)", sw.registry.GetProviderDisplayName(providerID), providerID)
	}

	var selectedProvider string
	prompt := &survey.Select{
		Message: "Select a popular provider:",
		Options: providerOptions,
	}

	if err := survey.AskOne(prompt, &selectedProvider); err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Extract provider ID from selection (it's the last part in parentheses)
	lastOpenParen := strings.LastIndex(selectedProvider, "(")
	lastCloseParen := strings.LastIndex(selectedProvider, ")")
	if lastOpenParen == -1 || lastCloseParen == -1 || lastCloseParen < lastOpenParen {
		return "", fmt.Errorf("invalid provider selection format")
	}
	providerID := strings.TrimSpace(selectedProvider[lastOpenParen+1 : lastCloseParen])

	return providerID, nil
}

// searchAndSelectProvider allows searching for providers
func (sw *SetupWizard) searchAndSelectProvider() (string, error) {
	var query string
	searchPrompt := &survey.Input{
		Message: "Enter search term (provider name, company, etc.):",
	}

	if err := survey.AskOne(searchPrompt, &query); err != nil {
		return "", fmt.Errorf("failed to get search query: %w", err)
	}

	matches := sw.registry.SearchProviders(query)
	if len(matches) == 0 {
		return "", fmt.Errorf("no providers found matching '%s'", query)
	}

	providerOptions := make([]string, len(matches))
	for i, providerID := range matches {
		providerOptions[i] = fmt.Sprintf("%s (%s)", sw.registry.GetProviderDisplayName(providerID), providerID)
	}

	var selectedProvider string
	prompt := &survey.Select{
		Message: fmt.Sprintf("Found %d providers matching '%s':", len(matches), query),
		Options: providerOptions,
	}

	if err := survey.AskOne(prompt, &selectedProvider); err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Extract provider ID from selection (it's the last part in parentheses)
	lastOpenParen := strings.LastIndex(selectedProvider, "(")
	lastCloseParen := strings.LastIndex(selectedProvider, ")")
	if lastOpenParen == -1 || lastCloseParen == -1 || lastCloseParen < lastOpenParen {
		return "", fmt.Errorf("invalid provider selection format")
	}
	providerID := strings.TrimSpace(selectedProvider[lastOpenParen+1 : lastCloseParen])

	return providerID, nil
}

// selectFromAllProviders shows all available providers
func (sw *SetupWizard) selectFromAllProviders() (string, error) {
	allProviders := sw.registry.GetAllProviders()

	providerOptions := make([]string, len(allProviders))
	for i, providerID := range allProviders {
		providerOptions[i] = fmt.Sprintf("%s (%s)", sw.registry.GetProviderDisplayName(providerID), providerID)
	}

	var selectedProvider string
	prompt := &survey.Select{
		Message:  "Select a provider:",
		Options:  providerOptions,
		PageSize: 15,
	}

	if err := survey.AskOne(prompt, &selectedProvider); err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Extract provider ID from selection (it's the last part in parentheses)
	lastOpenParen := strings.LastIndex(selectedProvider, "(")
	lastCloseParen := strings.LastIndex(selectedProvider, ")")
	if lastOpenParen == -1 || lastCloseParen == -1 || lastCloseParen < lastOpenParen {
		return "", fmt.Errorf("invalid provider selection format")
	}
	providerID := strings.TrimSpace(selectedProvider[lastOpenParen+1 : lastCloseParen])

	return providerID, nil
}

// collectRequiredFields collects required configuration fields
func (sw *SetupWizard) collectRequiredFields(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
	if len(def.RequiredFields) == 0 {
		return nil
	}

	fmt.Println("📝 Required configuration:")

	for _, field := range def.RequiredFields {
		value, err := sw.promptForField(field, true)
		if err != nil {
			return err
		}

		// Use the proper field mapper to handle complex multi-key providers
		MapFieldToConfig(field, value, providerConfig)
	}

	// Validate all required fields were collected
	if err := ValidateRequiredFields(def.ID, *providerConfig, def.RequiredFields); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// collectOptionalFields collects optional configuration fields
func (sw *SetupWizard) collectOptionalFields(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
	if len(def.OptionalFields) == 0 {
		return nil
	}

	configureOptional := false
	prompt := &survey.Confirm{
		Message: "Would you like to configure optional settings?",
		Default: false,
	}

	if err := survey.AskOne(prompt, &configureOptional); err != nil {
		return fmt.Errorf("failed to get optional config choice: %w", err)
	}

	if !configureOptional {
		return nil
	}

	fmt.Println("⚙️  Optional configuration:")

	for _, field := range def.OptionalFields {
		value, err := sw.promptForField(field, false)
		if err != nil {
			return err
		}

		if value != "" {
			// Map to provider config fields
			switch field.Name {
			case "baseUrl":
				providerConfig.BaseURL = value
			default:
				providerConfig.ExtraConfig[field.Name] = value
			}
		}
	}

	return nil
}

// promptForField prompts for a single configuration field
func (sw *SetupWizard) promptForField(field generated.ConfigField, required bool) (string, error) {
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

// selectModel helps user select a model for the provider
func (sw *SetupWizard) selectModel(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
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
		useDefault := true
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Use default model '%s'?", def.DefaultModelID),
			Default: true,
		}

		if err := survey.AskOne(prompt, &useDefault); err != nil {
			return fmt.Errorf("failed to get model choice: %w", err)
		}

		if useDefault {
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
			return nil
		}
	}

	// Show model selection
	modelIDs := make([]string, 0, len(def.Models))
	for modelID := range def.Models {
		modelIDs = append(modelIDs, modelID)
	}
	sort.Strings(modelIDs)

	modelOptions := make([]string, len(modelIDs))
	for i, modelID := range modelIDs {
		modelInfo := def.Models[modelID]
		description := modelID
		if modelInfo.Description != "" {
			description += fmt.Sprintf(" - %s", modelInfo.Description)
		}
		if modelInfo.ContextWindow > 0 {
			description += fmt.Sprintf(" (%dk context)", modelInfo.ContextWindow/1000)
		}
		modelOptions[i] = description
	}

	var selectedModel string
	prompt := &survey.Select{
		Message:  "Select a model:",
		Options:  modelOptions,
		PageSize: 10,
	}

	if err := survey.AskOne(prompt, &selectedModel); err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	// Extract model ID (first part before " - ")
	modelID := strings.Split(selectedModel, " - ")[0]
	modelID = strings.Split(modelID, " (")[0] // Remove context info

	providerConfig.ModelID = modelID
	if modelInfo, exists := def.Models[modelID]; exists {
		providerConfig.ModelInfo = config.ModelInfo{
			MaxTokens:      modelInfo.MaxTokens,
			ContextWindow:  modelInfo.ContextWindow,
			SupportsImages: modelInfo.SupportsImages,
			InputPrice:     modelInfo.InputPrice,
			OutputPrice:    modelInfo.OutputPrice,
			Description:    modelInfo.Description,
		}
	}

	return nil
}

// removeProvider removes a configured provider
func (sw *SetupWizard) removeProvider() error {
	cliConfig := sw.configManager.GetConfig()
	if len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	// List configured providers
	providerIDs := make([]string, 0, len(cliConfig.Providers))
	for id := range cliConfig.Providers {
		providerIDs = append(providerIDs, id)
	}
	sort.Strings(providerIDs)

	providerOptions := make([]string, len(providerIDs))
	for i, id := range providerIDs {
		provider := cliConfig.Providers[id]
		providerOptions[i] = fmt.Sprintf("%s (%s)", provider.Name, id)
	}

	var selectedProvider string
	prompt := &survey.Select{
		Message: "Select provider to remove:",
		Options: providerOptions,
	}

	if err := survey.AskOne(prompt, &selectedProvider); err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Extract provider ID
	parts := strings.Split(selectedProvider, "(")
	if len(parts) < 2 {
		return fmt.Errorf("invalid provider selection")
	}
	providerID := strings.TrimSuffix(parts[1], ")")

	// Confirm removal
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to remove %s?", providerID),
		Default: false,
	}

	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirm {
		fmt.Println("Removal cancelled.")
		return nil
	}

	// Remove provider
	if err := sw.configManager.RemoveProvider(providerID); err != nil {
		return fmt.Errorf("failed to remove provider: %w", err)
	}

	fmt.Printf("✅ Removed provider %s\n", providerID)
	return nil
}

// listConfiguredProviders lists all configured providers
func (sw *SetupWizard) listConfiguredProviders() {
	cliConfig := sw.configManager.GetConfig()
	if cliConfig == nil || len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured.")
		return
	}

	fmt.Println("\n📋 Configured providers:")
	for id, provider := range cliConfig.Providers {
		status := ""
		if id == cliConfig.DefaultProvider {
			status = " (default)"
		}

		fmt.Printf("  • %s (%s)%s\n", provider.Name, id, status)
		if provider.ModelID != "" {
			fmt.Printf("    Model: %s\n", provider.ModelID)
		}
		if provider.BaseURL != "" {
			fmt.Printf("    Base URL: %s\n", provider.BaseURL)
		}
	}
	fmt.Println()
}

// testProviders tests provider connections
func (sw *SetupWizard) testProviders() error {
	cliConfig := sw.configManager.GetConfig()
	if len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured to test.")
		return nil
	}

	fmt.Println("🧪 Testing provider connections...")
	fmt.Println("Note: This is a basic configuration validation. Full API testing requires actual API calls.")

	for id, provider := range cliConfig.Providers {
		fmt.Printf("Testing %s (%s)... ", provider.Name, id)

		// Basic validation
		if err := sw.registry.ValidateProviderConfig(provider); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			fmt.Printf("✅ Configuration valid\n")
		}
	}

	return nil
}

// setDefaultProvider sets the default provider
func (sw *SetupWizard) setDefaultProvider() error {
	cliConfig := sw.configManager.GetConfig()
	if len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}

	// List configured providers
	providerIDs := make([]string, 0, len(cliConfig.Providers))
	for id := range cliConfig.Providers {
		providerIDs = append(providerIDs, id)
	}
	sort.Strings(providerIDs)

	providerOptions := make([]string, len(providerIDs))
	for i, id := range providerIDs {
		provider := cliConfig.Providers[id]
		status := ""
		if id == cliConfig.DefaultProvider {
			status = " (current default)"
		}
		providerOptions[i] = fmt.Sprintf("%s (%s)%s", provider.Name, id, status)
	}

	var selectedProvider string
	prompt := &survey.Select{
		Message: "Select default provider:",
		Options: providerOptions,
	}

	if err := survey.AskOne(prompt, &selectedProvider); err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Extract provider ID
	parts := strings.Split(selectedProvider, "(")
	if len(parts) < 2 {
		return fmt.Errorf("invalid provider selection")
	}
	providerID := strings.TrimSuffix(strings.Split(parts[1], ")")[0], "")

	// Set default provider
	if err := sw.configManager.SetDefaultProvider(providerID); err != nil {
		return fmt.Errorf("failed to set default provider: %w", err)
	}

	fmt.Printf("✅ Set %s as default provider\n", providerID)
	return nil
}

// saveAndExit saves the configuration and exits
func (sw *SetupWizard) saveAndExit() error {
	cliConfig := sw.configManager.GetConfig()
	if len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured. Nothing to save.")
		return nil
	}

	// Validate configuration
	if err := sw.configManager.Validate(cliConfig); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save configuration
	if err := sw.configManager.Save(cliConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("✅ Configuration saved to %s\n", configPath)
	fmt.Printf("🎉 Setup complete! You can now use the Cline CLI with %d configured provider(s).\n", len(cliConfig.Providers))

	return nil
}
