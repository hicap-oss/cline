package generated

import (
	"strings"
	"testing"
)

// TestAllProvidersHaveDefinitions verifies every provider in AllProviders has a definition
func TestAllProvidersHaveDefinitions(t *testing.T) {
	definitions, err := GetProviderDefinitions()
	if err != nil {
		t.Fatalf("GetProviderDefinitions() error = %v", err)
	}

	for _, providerID := range AllProviders {
		t.Run(providerID, func(t *testing.T) {
			def, exists := definitions[providerID]
			if !exists {
				t.Errorf("provider %s exists in AllProviders but has no definition", providerID)
				return
			}

			// Verify basic definition properties
			if def.ID != providerID {
				t.Errorf("definition ID = %v, want %v", def.ID, providerID)
			}

			if def.Name == "" {
				t.Errorf("provider %s has empty Name", providerID)
			}

			if def.SetupInstructions == "" {
				t.Errorf("provider %s has empty SetupInstructions", providerID)
			}

			// Verify models map exists (can be empty for providers without model definitions)
			// Some providers may not have models defined in the TypeScript source
			if def.Models == nil {
				// This is acceptable - the models map is initialized in GetProviderDefinitions
				// but may be empty if no models are defined in the raw model definitions
			}
		})
	}
}

// TestRequiredFieldsMatchGenerator verifies required fields match the generator's expectations
func TestRequiredFieldsMatchGenerator(t *testing.T) {
	// Expected required fields for critical providers (from providerRequiredFields in generator)
	expectedRequired := map[string][]string{
		"cerebras":   {"cerebrasApiKey"},
		"deepseek":   {"deepSeekApiKey"},
		"xai":        {"xaiApiKey"},
		"groq":       {"groqApiKey"},
		"bedrock":    {"awsAccessKey", "awsSecretKey", "awsRegion"},
		"vertex":     {"vertexProjectId", "vertexRegion"},
		"ollama":     {"ollamaBaseUrl"},
		"lmstudio":   {"lmStudioBaseUrl"},
		"gemini":     {"geminiApiKey"},
		"openrouter": {"openRouterApiKey"},
		"openai":     {"openAiApiKey"},
	}

	definitions, err := GetProviderDefinitions()
	if err != nil {
		t.Fatalf("GetProviderDefinitions() error = %v", err)
	}

	for providerID, expectedFields := range expectedRequired {
		t.Run(providerID, func(t *testing.T) {
			def, exists := definitions[providerID]
			if !exists {
				t.Fatalf("provider %s not found in definitions", providerID)
			}

			// Create map of required field names for easy lookup
			actualRequired := make(map[string]bool)
			for _, field := range def.RequiredFields {
				actualRequired[field.Name] = true
			}

			// Verify all expected required fields are present
			for _, expectedField := range expectedFields {
				if !actualRequired[expectedField] {
					t.Errorf("provider %s missing required field %s", providerID, expectedField)
				}
			}

			// Verify count matches (allowing for general fields like requestTimeoutMs)
			// General fields are included for all providers, so actual may be higher
			if len(def.RequiredFields) < len(expectedFields) {
				t.Errorf("provider %s has %d required fields, expected at least %d",
					providerID, len(def.RequiredFields), len(expectedFields))
			}
		})
	}
}

// TestFieldCategorization tests that fields are categorized correctly
func TestFieldCategorization(t *testing.T) {
	configFields, err := GetConfigFields()
	if err != nil {
		t.Fatalf("GetConfigFields() error = %v", err)
	}

	tests := []struct {
		fieldName        string
		expectedCategory string
	}{
		{"cerebrasApiKey", "cerebras"},
		{"deepSeekApiKey", "deepseek"},
		{"xaiApiKey", "xai"},
		{"groqApiKey", "groq"},
		{"awsAccessKey", "aws"},
		{"awsSecretKey", "aws"},
		{"awsRegion", "aws"},
		{"vertexProjectId", "vertex"},
		{"vertexRegion", "vertex"},
		{"ollamaBaseUrl", "ollama"},
		{"lmStudioBaseUrl", "lmstudio"},
		{"geminiApiKey", "gemini"},
		{"requestTimeoutMs", "general"},
	}

	// Create map for quick lookup
	fieldMap := make(map[string]ConfigField)
	for _, field := range configFields {
		fieldMap[field.Name] = field
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field, exists := fieldMap[tt.fieldName]
			if !exists {
				t.Fatalf("field %s not found in config fields", tt.fieldName)
			}

			if field.Category != tt.expectedCategory {
				t.Errorf("field %s has category %s, want %s",
					tt.fieldName, field.Category, tt.expectedCategory)
			}
		})
	}
}

// TestDefaultModelIDs verifies default model IDs are present and valid
func TestDefaultModelIDs(t *testing.T) {
	definitions, err := GetProviderDefinitions()
	if err != nil {
		t.Fatalf("GetProviderDefinitions() error = %v", err)
	}

	// Providers that should have default model IDs (non-dynamic providers with known models)
	providersWithDefaults := []string{
		"cerebras",
		"deepseek",
		"xai",
		"gemini",
		"anthropic",
		"bedrock",
		"vertex",
		"mistral",
		"openai-native",
		"doubao",
	}

	for _, providerID := range providersWithDefaults {
		t.Run(providerID, func(t *testing.T) {
			def, exists := definitions[providerID]
			if !exists {
				t.Fatalf("provider %s not found", providerID)
			}

			if def.DefaultModelID == "" {
				t.Errorf("provider %s has empty DefaultModelID", providerID)
			}

			// Verify default model ID does not contain :1m suffix
			if strings.Contains(def.DefaultModelID, ":1m") {
				t.Errorf("provider %s DefaultModelID contains :1m suffix: %s",
					providerID, def.DefaultModelID)
			}

			// Verify default model exists in models map (if models map is not empty)
			if len(def.Models) > 0 {
				// Check for exact match or with version suffix (e.g., bedrock models may have :0 suffix)
				exactMatch := false
				for modelID := range def.Models {
					if modelID == def.DefaultModelID || modelID == def.DefaultModelID+":0" {
						exactMatch = true
						break
					}
				}
				if !exactMatch {
					t.Errorf("provider %s DefaultModelID %s (or with :0 suffix) not found in Models map",
						providerID, def.DefaultModelID)
				}
			}
		})
	}
}

// TestFieldFilteringByCategory tests that getFieldsByProvider uses category field correctly
func TestFieldFilteringByCategory(t *testing.T) {
	configFields, err := GetConfigFields()
	if err != nil {
		t.Fatalf("GetConfigFields() error = %v", err)
	}

	tests := []struct {
		providerID    string
		shouldInclude []string // Fields that should be included
		shouldExclude []string // Fields that should be excluded
	}{
		{
			providerID:    "cerebras",
			shouldInclude: []string{"cerebrasApiKey"},
			shouldExclude: []string{"deepSeekApiKey", "xaiApiKey", "groqApiKey"},
		},
		{
			providerID:    "deepseek",
			shouldInclude: []string{"deepSeekApiKey"},
			shouldExclude: []string{"cerebrasApiKey", "xaiApiKey", "groqApiKey"},
		},
		{
			providerID:    "bedrock",
			shouldInclude: []string{"awsAccessKey", "awsSecretKey", "awsRegion"},
			shouldExclude: []string{"cerebrasApiKey", "vertexProjectId"},
		},
		{
			providerID:    "vertex",
			shouldInclude: []string{"vertexProjectId", "vertexRegion"},
			shouldExclude: []string{"awsAccessKey", "cerebrasApiKey"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.providerID, func(t *testing.T) {
			// Get required fields for this provider
			requiredFields := getFieldsByProvider(tt.providerID, configFields, true)
			optionalFields := getFieldsByProvider(tt.providerID, configFields, false)

			allFields := append(requiredFields, optionalFields...)

			// Create map for quick lookup
			fieldMap := make(map[string]bool)
			for _, field := range allFields {
				fieldMap[field.Name] = true
			}

			// Verify included fields are present
			for _, fieldName := range tt.shouldInclude {
				if !fieldMap[fieldName] {
					t.Errorf("provider %s should include field %s but doesn't",
						tt.providerID, fieldName)
				}
			}

			// Verify excluded fields are not present
			for _, fieldName := range tt.shouldExclude {
				if fieldMap[fieldName] {
					t.Errorf("provider %s should exclude field %s but includes it",
						tt.providerID, fieldName)
				}
			}
		})
	}
}

// TestCriticalProviders tests specific providers in detail
func TestCriticalProviders(t *testing.T) {
	definitions, err := GetProviderDefinitions()
	if err != nil {
		t.Fatalf("GetProviderDefinitions() error = %v", err)
	}

	t.Run("Cerebras", func(t *testing.T) {
		def := definitions["cerebras"]

		// Verify required fields
		if len(def.RequiredFields) == 0 {
			t.Error("Cerebras should have required fields")
		}

		// Verify has cerebrasApiKey
		hasApiKey := false
		for _, field := range def.RequiredFields {
			if field.Name == "cerebrasApiKey" {
				hasApiKey = true
				break
			}
		}
		if !hasApiKey {
			t.Error("Cerebras should require cerebrasApiKey")
		}

		// Verify default model
		if def.DefaultModelID == "" {
			t.Error("Cerebras should have a default model")
		}

		// Verify has models
		if len(def.Models) == 0 {
			t.Error("Cerebras should have models defined")
		}
	})

	t.Run("DeepSeek", func(t *testing.T) {
		def := definitions["deepseek"]

		// Verify required fields
		hasApiKey := false
		for _, field := range def.RequiredFields {
			if field.Name == "deepSeekApiKey" {
				hasApiKey = true
				break
			}
		}
		if !hasApiKey {
			t.Error("DeepSeek should require deepSeekApiKey")
		}

		// Verify default model
		if def.DefaultModelID == "" {
			t.Error("DeepSeek should have a default model")
		}
	})

	t.Run("XAI", func(t *testing.T) {
		def := definitions["xai"]

		// Verify required fields
		hasApiKey := false
		for _, field := range def.RequiredFields {
			if field.Name == "xaiApiKey" {
				hasApiKey = true
				break
			}
		}
		if !hasApiKey {
			t.Error("XAI should require xaiApiKey")
		}

		// Verify default model
		if def.DefaultModelID == "" {
			t.Error("XAI should have a default model")
		}
	})

	t.Run("Groq", func(t *testing.T) {
		def := definitions["groq"]

		// Verify required fields
		hasApiKey := false
		for _, field := range def.RequiredFields {
			if field.Name == "groqApiKey" {
				hasApiKey = true
				break
			}
		}
		if !hasApiKey {
			t.Error("Groq should require groqApiKey")
		}

		// Verify default model
		if def.DefaultModelID == "" {
			t.Error("Groq should have a default model")
		}
	})

	t.Run("AWS Bedrock", func(t *testing.T) {
		def := definitions["bedrock"]

		// Verify all 3 AWS keys are required
		requiredKeys := []string{"awsAccessKey", "awsSecretKey", "awsRegion"}
		fieldMap := make(map[string]bool)
		for _, field := range def.RequiredFields {
			fieldMap[field.Name] = true
		}

		for _, key := range requiredKeys {
			if !fieldMap[key] {
				t.Errorf("Bedrock should require %s", key)
			}
		}

		// Verify default model
		if def.DefaultModelID == "" {
			t.Error("Bedrock should have a default model")
		}

		// Verify default model doesn't have :1m suffix
		if strings.Contains(def.DefaultModelID, ":1m") {
			t.Errorf("Bedrock DefaultModelID should not contain :1m suffix: %s", def.DefaultModelID)
		}
	})
}

// TestProviderCount verifies we have the expected number of providers
func TestProviderCount(t *testing.T) {
	if len(AllProviders) < 30 {
		t.Errorf("Expected at least 30 providers, got %d", len(AllProviders))
	}

	definitions, err := GetProviderDefinitions()
	if err != nil {
		t.Fatalf("GetProviderDefinitions() error = %v", err)
	}

	if len(definitions) != len(AllProviders) {
		t.Errorf("AllProviders has %d entries but definitions has %d",
			len(AllProviders), len(definitions))
	}
}

// TestModelInfo verifies model info structure for critical models
func TestModelInfo(t *testing.T) {
	definitions, err := GetProviderDefinitions()
	if err != nil {
		t.Fatalf("GetProviderDefinitions() error = %v", err)
	}

	t.Run("Cerebras models", func(t *testing.T) {
		def := definitions["cerebras"]
		if len(def.Models) == 0 {
			t.Fatal("Cerebras has no models")
		}

		// Check a known model exists
		model, exists := def.Models["qwen-3-coder-480b-free"]
		if !exists {
			t.Error("Cerebras should have qwen-3-coder-480b-free model")
			return
		}

		// Verify model has required fields
		if model.ContextWindow == 0 {
			t.Error("Model should have context window")
		}
	})

	t.Run("DeepSeek models", func(t *testing.T) {
		def := definitions["deepseek"]
		if len(def.Models) == 0 {
			t.Fatal("DeepSeek has no models")
		}

		// Check a known model exists
		model, exists := def.Models["deepseek-chat"]
		if !exists {
			t.Error("DeepSeek should have deepseek-chat model")
			return
		}

		// Verify prompt cache support
		if !model.SupportsPromptCache {
			t.Error("DeepSeek models should support prompt cache")
		}
	})
}

// TestDynamicModelProviders verifies providers with dynamic model discovery
func TestDynamicModelProviders(t *testing.T) {
	definitions, err := GetProviderDefinitions()
	if err != nil {
		t.Fatalf("GetProviderDefinitions() error = %v", err)
	}

	dynamicProviders := []string{
		"openrouter",
		"ollama",
		"lmstudio",
		"together",
		"openai-native",
	}

	for _, providerID := range dynamicProviders {
		t.Run(providerID, func(t *testing.T) {
			def, exists := definitions[providerID]
			if !exists {
				t.Fatalf("provider %s not found", providerID)
			}

			if !def.HasDynamicModels {
				t.Errorf("provider %s should have HasDynamicModels = true", providerID)
			}
		})
	}
}

// TestFieldTypes verifies field types are set correctly
func TestFieldTypes(t *testing.T) {
	configFields, err := GetConfigFields()
	if err != nil {
		t.Fatalf("GetConfigFields() error = %v", err)
	}

	tests := []struct {
		fieldName    string
		expectedType string
	}{
		{"cerebrasApiKey", "password"},
		{"deepSeekApiKey", "password"},
		{"xaiApiKey", "password"},
		{"groqApiKey", "password"},
		{"awsAccessKey", "password"},
		{"awsSecretKey", "password"},
		{"awsRegion", "select"},
		{"ollamaBaseUrl", "url"},
		{"lmStudioBaseUrl", "url"},
	}

	// Create map for quick lookup
	fieldMap := make(map[string]ConfigField)
	for _, field := range configFields {
		fieldMap[field.Name] = field
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field, exists := fieldMap[tt.fieldName]
			if !exists {
				t.Fatalf("field %s not found", tt.fieldName)
			}

			if field.FieldType != tt.expectedType {
				t.Errorf("field %s has type %s, want %s",
					tt.fieldName, field.FieldType, tt.expectedType)
			}
		})
	}
}
