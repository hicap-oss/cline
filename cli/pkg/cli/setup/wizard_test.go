package setup

import (
	"testing"

	"github.com/cline/cli/pkg/cli/config"
	"github.com/cline/cli/pkg/generated"
)

func TestNewSetupWizard(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create setup wizard: %v", err)
	}

	if wizard == nil {
		t.Fatal("Setup wizard is nil")
	}

	if wizard.configManager == nil {
		t.Fatal("Config manager is nil")
	}

	if wizard.registry == nil {
		t.Fatal("Provider registry is nil")
	}
}

func TestProviderSelection(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create setup wizard: %v", err)
	}

	// Test that we can get all providers
	allProviders := wizard.registry.GetAllProviders()
	if len(allProviders) == 0 {
		t.Fatal("No providers found")
	}

	// Test that we can get popular providers
	popularProviders := wizard.registry.GetPopularProviders()
	if len(popularProviders) == 0 {
		t.Fatal("No popular providers found")
	}

	// Test that popular providers are a subset of all providers
	allProvidersMap := make(map[string]bool)
	for _, provider := range allProviders {
		allProvidersMap[provider] = true
	}

	for _, provider := range popularProviders {
		if !allProvidersMap[provider] {
			t.Fatalf("Popular provider %s not found in all providers", provider)
		}
	}
}

func TestProviderDefinitions(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create setup wizard: %v", err)
	}

	// Test that we can get provider definitions
	allProviders := wizard.registry.GetAllProviders()
	for _, providerID := range allProviders[:5] { // Test first 5 providers
		def, err := wizard.registry.GetProviderDefinition(providerID)
		if err != nil {
			t.Fatalf("Failed to get definition for provider %s: %v", providerID, err)
		}

		if def.ID != providerID {
			t.Fatalf("Provider ID mismatch: expected %s, got %s", providerID, def.ID)
		}

		if def.Name == "" {
			t.Fatalf("Provider %s has empty name", providerID)
		}

		if def.SetupInstructions == "" {
			t.Fatalf("Provider %s has empty setup instructions", providerID)
		}
	}
}

func TestProviderValidation(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create setup wizard: %v", err)
	}

	// Test validation with empty config (should fail for most providers)
	emptyConfig := config.ProviderConfig{
		ID:   "anthropic",
		Name: "Anthropic",
	}

	err = wizard.registry.ValidateProviderConfig(emptyConfig)
	if err == nil {
		t.Fatal("Expected validation to fail for empty config")
	}

	// Test validation with valid config
	validConfig := config.ProviderConfig{
		ID:     "anthropic",
		Name:   "Anthropic",
		APIKey: "test-key",
	}

	err = wizard.registry.ValidateProviderConfig(validConfig)
	if err != nil {
		t.Fatalf("Expected validation to pass for valid config: %v", err)
	}
}

func TestProviderCategories(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create setup wizard: %v", err)
	}

	categories := wizard.registry.GetProvidersByCategory()
	if len(categories) == 0 {
		t.Fatal("No provider categories found")
	}

	// Check that all providers are categorized
	allProviders := wizard.registry.GetAllProviders()
	categorizedProviders := make(map[string]bool)

	for _, providers := range categories {
		for _, provider := range providers {
			categorizedProviders[provider] = true
		}
	}

	for _, provider := range allProviders {
		if !categorizedProviders[provider] {
			t.Fatalf("Provider %s is not categorized", provider)
		}
	}
}

func TestProviderSearch(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create setup wizard: %v", err)
	}

	// Test search functionality
	testCases := []struct {
		query    string
		expected []string
	}{
		{"anthropic", []string{"anthropic"}},
		{"openai", []string{"openai", "openai-native"}},
		{"google", []string{"gemini", "vertex"}},
	}

	for _, tc := range testCases {
		results := wizard.registry.SearchProviders(tc.query)
		
		// Check that all expected providers are found
		resultMap := make(map[string]bool)
		for _, result := range results {
			resultMap[result] = true
		}

		for _, expected := range tc.expected {
			if !resultMap[expected] {
				t.Fatalf("Expected provider %s not found in search results for query '%s'", expected, tc.query)
			}
		}
	}
}

func TestGeneratedProviderData(t *testing.T) {
	// Test that generated provider data is accessible
	providers := generated.AllProviders
	if len(providers) == 0 {
		t.Fatal("No providers found in generated data")
	}

	// Test that we can get provider definitions
	definitions, err := generated.GetProviderDefinitions()
	if err != nil {
		t.Fatalf("Failed to get provider definitions: %v", err)
	}

	if len(definitions) == 0 {
		t.Fatal("No provider definitions found")
	}

	// Test that all providers have definitions
	for _, providerID := range providers {
		if _, exists := definitions[providerID]; !exists {
			t.Fatalf("No definition found for provider %s", providerID)
		}
	}

	// Test that we can get config fields
	configFields, err := generated.GetConfigFields()
	if err != nil {
		t.Fatalf("Failed to get config fields: %v", err)
	}

	if len(configFields) == 0 {
		t.Fatal("No config fields found")
	}

	// Test that we can get model definitions
	modelDefinitions, err := generated.GetModelDefinitions()
	if err != nil {
		t.Fatalf("Failed to get model definitions: %v", err)
	}

	if len(modelDefinitions) == 0 {
		t.Fatal("No model definitions found")
	}
}
