package auth

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cline/cli/pkg/cli/config"
	"github.com/cline/cli/pkg/cli/models"
	"github.com/cline/cli/pkg/cli/setup"
	"github.com/cline/cli/pkg/generated"
)

// ProviderWizard handles the interactive provider configuration process
type ProviderWizard struct {
	configManager *config.ConfigManager
	registry      *config.ProviderRegistry
}

// NewProviderWizard creates a new provider configuration wizard
func NewProviderWizard() (*ProviderWizard, error) {
	configManager, err := config.NewConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	registry, err := config.NewProviderRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to create provider registry: %w", err)
	}

	return &ProviderWizard{
		configManager: configManager,
		registry:      registry,
	}, nil
}

// Run runs the complete provider configuration wizard
func (pw *ProviderWizard) Run() error {
	fmt.Println("Welcome to Cline API Provider Configuration!")
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
			fmt.Println("Configuration cancelled.")
			return nil
		}
	}

	// Load existing config or create new one
	if _, err := pw.configManager.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Show provider selection menu
	for {
		// Always get the current config from ConfigManager
		cliConfig := pw.configManager.GetConfig()

		action, err := pw.showMainMenu(cliConfig)
		if err != nil {
			return err
		}

		switch action {
		case "add":
			if err := pw.addProvider(); err != nil {
				fmt.Printf("Error adding provider: %v\n", err)
				continue
			}
		case "remove":
			if err := pw.removeProvider(); err != nil {
				fmt.Printf("Error removing provider: %v\n", err)
				continue
			}
		case "list":
			pw.listConfiguredProviders()
		case "test":
			if err := pw.testProviders(); err != nil {
				fmt.Printf("Error testing providers: %v\n", err)
				continue
			}
		case "default":
			if err := pw.setDefaultProvider(); err != nil {
				fmt.Printf("Error setting default provider: %v\n", err)
				continue
			}
		case "save":
			if err := pw.saveAndExit(); err != nil {
				return err
			}
			return nil
		case "exit":
			return nil
		}
	}
}

// showMainMenu displays the main provider configuration menu
func (pw *ProviderWizard) showMainMenu(cliConfig *config.CLIConfig) (string, error) {
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
func (pw *ProviderWizard) addProvider() error {
	fmt.Println("\nAdding a new provider...")

	// Show provider selection
	providerID, err := pw.selectProvider()
	if err != nil {
		return err
	}

	// Get provider definition
	def, err := pw.registry.GetProviderDefinition(providerID)
	if err != nil {
		return err
	}

	fmt.Printf("\nConfiguring %s\n", def.Name)
	fmt.Printf("Setup instructions: %s\n\n", def.SetupInstructions)

	// Create provider config
	providerConfig := config.ProviderConfig{
		ID:          providerID,
		Name:        def.Name,
		ExtraConfig: make(map[string]string),
	}

	// Collect required fields
	if err := pw.collectRequiredFields(def, &providerConfig); err != nil {
		return err
	}

	// Collect optional fields
	if err := pw.collectOptionalFields(def, &providerConfig); err != nil {
		return err
	}

	// Select model
	if err := pw.selectModel(def, &providerConfig); err != nil {
		return err
	}

	// Validate configuration
	if err := pw.registry.ValidateProviderConfig(providerConfig); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Add to config
	if err := pw.configManager.AddProvider(providerConfig); err != nil {
		return fmt.Errorf("failed to add provider: %w", err)
	}

	fmt.Printf("Successfully configured %s!\n", def.Name)
	return nil
}

// selectProvider shows provider selection interface
func (pw *ProviderWizard) selectProvider() (string, error) {
	// Show selection method
	method := ""
	methodPrompt := &survey.Select{
		Message: "How would you like to choose a provider?",
		Options: []string{
			"View popular providers",
			"View all providers",
			"Search providers",
		},
	}

	if err := survey.AskOne(methodPrompt, &method); err != nil {
		return "", fmt.Errorf("failed to get selection method: %w", err)
	}

	switch method {
	case "View popular providers":
		return pw.selectFromPopularProviders()
	case "Search providers":
		return pw.searchAndSelectProvider()
	case "View all providers":
		return pw.selectFromAllProviders()
	default:
		return "", fmt.Errorf("invalid selection method")
	}
}

// selectFromPopularProviders shows popular providers
func (pw *ProviderWizard) selectFromPopularProviders() (string, error) {
	popular := pw.registry.GetPopularProviders()

	providerOptions := make([]string, len(popular))
	for i, providerID := range popular {
		providerOptions[i] = fmt.Sprintf("%s (%s)", pw.registry.GetProviderDisplayName(providerID), providerID)
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
func (pw *ProviderWizard) searchAndSelectProvider() (string, error) {
	var query string
	searchPrompt := &survey.Input{
		Message: "Enter search term (provider name, company, etc.):",
	}

	if err := survey.AskOne(searchPrompt, &query); err != nil {
		return "", fmt.Errorf("failed to get search query: %w", err)
	}

	matches := pw.registry.SearchProviders(query)
	if len(matches) == 0 {
		return "", fmt.Errorf("no providers found matching '%s'", query)
	}

	providerOptions := make([]string, len(matches))
	for i, providerID := range matches {
		providerOptions[i] = fmt.Sprintf("%s (%s)", pw.registry.GetProviderDisplayName(providerID), providerID)
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
func (pw *ProviderWizard) selectFromAllProviders() (string, error) {
	allProviders := pw.registry.GetAllProviders()

	providerOptions := make([]string, len(allProviders))
	for i, providerID := range allProviders {
		providerOptions[i] = fmt.Sprintf("%s (%s)", pw.registry.GetProviderDisplayName(providerID), providerID)
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
func (pw *ProviderWizard) collectRequiredFields(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
	if len(def.RequiredFields) == 0 {
		return nil
	}

	fmt.Println("Required configuration:")

	for _, field := range def.RequiredFields {
		value, err := pw.promptForField(field, true)
		if err != nil {
			return err
		}

		// Use the proper field mapper to handle complex multi-key providers
		setup.MapFieldToConfig(field, value, providerConfig)
	}

	// Validate all required fields
	if err := setup.ValidateRequiredFields(def.ID, *providerConfig, def.RequiredFields); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// collectOptionalFields collects optional configuration fields
func (pw *ProviderWizard) collectOptionalFields(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
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

	fmt.Println("Optional configuration:")

	for _, field := range def.OptionalFields {
		value, err := pw.promptForField(field, false)
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
func (pw *ProviderWizard) promptForField(field generated.ConfigField, required bool) (string, error) {
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

// showModelList fetches and displays available models for a provider
func (pw *ProviderWizard) showModelList(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
	var modelMap map[string]config.ModelInfo
	var fetchErr error

	// Try to fetch models from API if supported
	if def.SupportsModelListing {
		fmt.Println("Fetching models from API...")
		
		apiKey := providerConfig.APIKey
		baseURL := providerConfig.BaseURL
		
		modelMap, fetchErr = models.FetchModelsForProvider(def, apiKey, baseURL)
		
		if fetchErr != nil {
			fmt.Printf("Failed to fetch models from API: %v\n", fetchErr)
			fmt.Println("Showing hardcoded models instead...")
		}
	} else {
		// Use hardcoded models for providers that don't support listing
		modelMap = make(map[string]config.ModelInfo)
		for modelID, modelInfo := range def.Models {
			modelMap[modelID] = config.ModelInfo{
				MaxTokens:      modelInfo.MaxTokens,
				ContextWindow:  modelInfo.ContextWindow,
				SupportsImages: modelInfo.SupportsImages,
				InputPrice:     modelInfo.InputPrice,
				OutputPrice:    modelInfo.OutputPrice,
				Description:    modelInfo.Description,
			}
		}
	}

	// If we have no models at all, show error
	if len(modelMap) == 0 {
		return fmt.Errorf("no models available for this provider")
	}

	// Format and sort models
	modelOptions := models.FormatModelList(modelMap)
	
	// Paginate if needed
	const pageSize = 15
	pages := models.PaginateModels(modelOptions, pageSize)
	currentPage := 0

	// Display pages with pagination
	for currentPage < len(pages) {
		page := pages[currentPage]
		
		// Display the current page
		pageText := models.FormatModelPage(page, currentPage+1, len(pages))
		fmt.Println(pageText)

		// Get user input for selection or pagination
		var selection string
		prompt := &survey.Input{
			Message: "Select by number or enter model ID (or 'next' for more, 'back' for previous):",
		}

		if err := survey.AskOne(prompt, &selection); err != nil {
			return fmt.Errorf("failed to get selection: %w", err)
		}

		selection = strings.TrimSpace(selection)

		// Handle pagination commands
		if strings.ToLower(selection) == "next" || strings.ToLower(selection) == "n" {
			if currentPage < len(pages)-1 {
				currentPage++
				continue
			} else {
				fmt.Println("Already on last page.")
				continue
			}
		}

		if strings.ToLower(selection) == "back" || strings.ToLower(selection) == "prev" || strings.ToLower(selection) == "b" {
			if currentPage > 0 {
				currentPage--
				continue
			} else {
				fmt.Println("Already on first page.")
				continue
			}
		}

		// Try to find model by number or ID
		modelID, found := models.FindModelByNumberOrID(selection, modelOptions)
		if !found {
			fmt.Printf("Invalid selection '%s'. Please try again.\n", selection)
			continue
		}

		// Set the selected model
		providerConfig.ModelID = modelID
		if modelInfo, exists := modelMap[modelID]; exists {
			providerConfig.ModelInfo = modelInfo
		}

		fmt.Printf("Selected model: %s\n", modelID)
		return nil
	}

	return nil
}

// selectModel helps user select a model for the provider
func (pw *ProviderWizard) selectModel(def *generated.ProviderDefinition, providerConfig *config.ProviderConfig) error {
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

	// Interactive model selection with support for "list" command
	for {
		var modelInput string
		prompt := &survey.Input{
			Message: "Enter model ID or type 'list' to see available models:",
		}

		if err := survey.AskOne(prompt, &modelInput); err != nil {
			return fmt.Errorf("failed to get model ID: %w", err)
		}

		modelInput = strings.TrimSpace(modelInput)

		// Check if user wants to list models
		if strings.ToLower(modelInput) == "list" {
			// showModelList handles the complete selection process
			// If it returns nil, a model was successfully selected
			if err := pw.showModelList(def, providerConfig); err != nil {
				fmt.Printf("Error displaying models: %v\n", err)
				fmt.Println("Continuing with manual model entry...")
				continue
			}
			// Model was selected in showModelList, we can return
			return nil
		}

		// If empty and provider has dynamic models but no hardcoded ones, allow empty
		if modelInput == "" {
			if def.HasDynamicModels && len(def.Models) == 0 {
				providerConfig.ModelID = ""
				return nil
			}
			fmt.Println("Model ID cannot be empty. Type 'list' to see available models.")
			continue
		}

		// Validate the model exists (if we have models)
		if len(def.Models) > 0 {
			if _, exists := def.Models[modelInput]; !exists {
				fmt.Printf("Model '%s' not found. Type 'list' to see available models.\n", modelInput)
				continue
			}
		}

		// Set the selected model
		providerConfig.ModelID = modelInput
		if modelInfo, exists := def.Models[modelInput]; exists {
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

// removeProvider removes a configured provider
func (pw *ProviderWizard) removeProvider() error {
	cliConfig := pw.configManager.GetConfig()
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

	// Extract provider ID from selection (it's the last part in parentheses)
	lastOpenParen := strings.LastIndex(selectedProvider, "(")
	lastCloseParen := strings.LastIndex(selectedProvider, ")")
	if lastOpenParen == -1 || lastCloseParen == -1 || lastCloseParen < lastOpenParen {
		return fmt.Errorf("invalid provider selection format")
	}
	providerID := strings.TrimSpace(selectedProvider[lastOpenParen+1 : lastCloseParen])

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
	if err := pw.configManager.RemoveProvider(providerID); err != nil {
		return fmt.Errorf("failed to remove provider: %w", err)
	}

	fmt.Printf("Removed provider %s\n", providerID)
	return nil
}

// listConfiguredProviders lists all configured providers
func (pw *ProviderWizard) listConfiguredProviders() {
	cliConfig := pw.configManager.GetConfig()
	if cliConfig == nil || len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured.")
		return
	}

	fmt.Println("\nConfigured providers:")
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
func (pw *ProviderWizard) testProviders() error {
	cliConfig := pw.configManager.GetConfig()
	if len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured to test.")
		return nil
	}

	fmt.Println("Testing provider connections...")
	fmt.Println("Note: This is a basic configuration validation. Full API testing requires actual API calls.")

	for id, provider := range cliConfig.Providers {
		fmt.Printf("Testing %s (%s)... ", provider.Name, id)

		// Basic validation
		if err := pw.registry.ValidateProviderConfig(provider); err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Printf("Configuration valid\n")
		}
	}

	return nil
}

// setDefaultProvider sets the default provider
func (pw *ProviderWizard) setDefaultProvider() error {
	cliConfig := pw.configManager.GetConfig()
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

	// Extract provider ID from selection (it's the last part in parentheses)
	lastOpenParen := strings.LastIndex(selectedProvider, "(")
	lastCloseParen := strings.LastIndex(selectedProvider, ")")
	if lastOpenParen == -1 || lastCloseParen == -1 || lastCloseParen < lastOpenParen {
		return fmt.Errorf("invalid provider selection format")
	}
	providerID := strings.TrimSpace(selectedProvider[lastOpenParen+1 : lastCloseParen])

	// Set default provider
	if err := pw.configManager.SetDefaultProvider(providerID); err != nil {
		return fmt.Errorf("failed to set default provider: %w", err)
	}

	fmt.Printf("Set %s as default provider\n", providerID)
	return nil
}

// saveAndExit saves the configuration and exits
func (pw *ProviderWizard) saveAndExit() error {
	cliConfig := pw.configManager.GetConfig()
	if len(cliConfig.Providers) == 0 {
		fmt.Println("No providers configured. Nothing to save.")
		return nil
	}

	// Validate configuration
	if err := pw.configManager.Validate(cliConfig); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save configuration
	if err := pw.configManager.Save(cliConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("Configuration saved to %s\n", configPath)
	fmt.Printf("Setup complete! You can now use the Cline CLI with %d configured provider(s).\n", len(cliConfig.Providers))

	return nil
}
