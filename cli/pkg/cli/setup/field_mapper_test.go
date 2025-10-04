package setup

import (
	"testing"

	"github.com/cline/cli/pkg/cli/config"
	"github.com/cline/cli/pkg/generated"
)

// TestMapFieldToConfig_APIKeys tests that provider-specific API keys are mapped to APIKey field
func TestMapFieldToConfig_APIKeys(t *testing.T) {
	tests := []struct {
		name           string
		fieldName      string
		value          string
		expectedAPIKey string
	}{
		{
			name:           "Generic API key",
			fieldName:      "apiKey",
			value:          "test-api-key",
			expectedAPIKey: "test-api-key",
		},
		{
			name:           "Cerebras API key",
			fieldName:      "cerebrasApiKey",
			value:          "cerebras-test-key",
			expectedAPIKey: "cerebras-test-key",
		},
		{
			name:           "DeepSeek API key",
			fieldName:      "deepSeekApiKey",
			value:          "deepseek-test-key",
			expectedAPIKey: "deepseek-test-key",
		},
		{
			name:           "XAI API key",
			fieldName:      "xaiApiKey",
			value:          "xai-test-key",
			expectedAPIKey: "xai-test-key",
		},
		{
			name:           "Groq API key",
			fieldName:      "groqApiKey",
			value:          "groq-test-key",
			expectedAPIKey: "groq-test-key",
		},
		{
			name:           "OpenRouter API key",
			fieldName:      "openRouterApiKey",
			value:          "openrouter-test-key",
			expectedAPIKey: "openrouter-test-key",
		},
		{
			name:           "Anthropic API key",
			fieldName:      "anthropicApiKey",
			value:          "anthropic-test-key",
			expectedAPIKey: "anthropic-test-key",
		},
		{
			name:           "Gemini API key",
			fieldName:      "geminiApiKey",
			value:          "gemini-test-key",
			expectedAPIKey: "gemini-test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig := &config.ProviderConfig{}
			field := generated.ConfigField{Name: tt.fieldName}

			MapFieldToConfig(field, tt.value, providerConfig)

			if providerConfig.APIKey != tt.expectedAPIKey {
				t.Errorf("APIKey = %v, want %v", providerConfig.APIKey, tt.expectedAPIKey)
			}
		})
	}
}

// TestMapFieldToConfig_AWSFields tests that AWS multi-key fields go to ExtraConfig
func TestMapFieldToConfig_AWSFields(t *testing.T) {
	tests := []struct {
		name              string
		fieldName         string
		value             string
		expectedConfigKey string
	}{
		{
			name:              "AWS Access Key",
			fieldName:         "awsAccessKey",
			value:             "AKIAIOSFODNN7EXAMPLE",
			expectedConfigKey: "aws_access_key",
		},
		{
			name:              "AWS Secret Key",
			fieldName:         "awsSecretKey",
			value:             "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			expectedConfigKey: "aws_secret_key",
		},
		{
			name:              "AWS Region",
			fieldName:         "awsRegion",
			value:             "us-east-1",
			expectedConfigKey: "aws_region",
		},
		{
			name:              "AWS Session Token",
			fieldName:         "awsSessionToken",
			value:             "session-token-example",
			expectedConfigKey: "aws_session_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig := &config.ProviderConfig{}
			field := generated.ConfigField{Name: tt.fieldName}

			MapFieldToConfig(field, tt.value, providerConfig)

			if providerConfig.ExtraConfig == nil {
				t.Fatal("ExtraConfig is nil")
			}

			if got := providerConfig.ExtraConfig[tt.expectedConfigKey]; got != tt.value {
				t.Errorf("ExtraConfig[%s] = %v, want %v", tt.expectedConfigKey, got, tt.value)
			}

			// Verify APIKey was not set for AWS fields
			if providerConfig.APIKey != "" {
				t.Errorf("APIKey should be empty for AWS fields, got %v", providerConfig.APIKey)
			}
		})
	}
}

// TestMapFieldToConfig_AWSBedrockMultipleKeys tests AWS Bedrock with all 3 required keys
func TestMapFieldToConfig_AWSBedrockMultipleKeys(t *testing.T) {
	providerConfig := &config.ProviderConfig{}

	// Map all three required AWS fields
	MapFieldToConfig(generated.ConfigField{Name: "awsAccessKey"}, "AKIAIOSFODNN7EXAMPLE", providerConfig)
	MapFieldToConfig(generated.ConfigField{Name: "awsSecretKey"}, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", providerConfig)
	MapFieldToConfig(generated.ConfigField{Name: "awsRegion"}, "us-west-2", providerConfig)

	// Verify all three keys are stored correctly
	if providerConfig.ExtraConfig == nil {
		t.Fatal("ExtraConfig is nil")
	}

	expectedKeys := map[string]string{
		"aws_access_key": "AKIAIOSFODNN7EXAMPLE",
		"aws_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"aws_region":     "us-west-2",
	}

	for key, expectedValue := range expectedKeys {
		if got := providerConfig.ExtraConfig[key]; got != expectedValue {
			t.Errorf("ExtraConfig[%s] = %v, want %v", key, got, expectedValue)
		}
	}

	// Verify we have exactly 3 keys (no extras)
	if len(providerConfig.ExtraConfig) != 3 {
		t.Errorf("ExtraConfig has %d keys, want 3", len(providerConfig.ExtraConfig))
	}
}

// TestMapFieldToConfig_VertexFields tests that Vertex fields are mapped correctly
func TestMapFieldToConfig_VertexFields(t *testing.T) {
	tests := []struct {
		name              string
		fieldName         string
		value             string
		expectedConfigKey string
	}{
		{
			name:              "Vertex Project ID",
			fieldName:         "vertexProjectId",
			value:             "my-gcp-project",
			expectedConfigKey: "vertex_project_id",
		},
		{
			name:              "Vertex Region",
			fieldName:         "vertexRegion",
			value:             "us-central1",
			expectedConfigKey: "vertex_region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig := &config.ProviderConfig{}
			field := generated.ConfigField{Name: tt.fieldName}

			MapFieldToConfig(field, tt.value, providerConfig)

			if providerConfig.ExtraConfig == nil {
				t.Fatal("ExtraConfig is nil")
			}

			if got := providerConfig.ExtraConfig[tt.expectedConfigKey]; got != tt.value {
				t.Errorf("ExtraConfig[%s] = %v, want %v", tt.expectedConfigKey, got, tt.value)
			}
		})
	}
}

// TestMapFieldToConfig_BaseURLs tests that base URL fields are mapped to BaseURL
func TestMapFieldToConfig_BaseURLs(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
	}{
		{
			name:      "Ollama Base URL",
			fieldName: "ollamaBaseUrl",
			value:     "http://localhost:11434",
		},
		{
			name:      "LM Studio Base URL",
			fieldName: "lmStudioBaseUrl",
			value:     "http://localhost:1234",
		},
		{
			name:      "OpenAI Base URL",
			fieldName: "openAiBaseUrl",
			value:     "https://api.openai.com/v1",
		},
		{
			name:      "LiteLLM Base URL",
			fieldName: "liteLLMBaseUrl",
			value:     "http://localhost:4000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig := &config.ProviderConfig{}
			field := generated.ConfigField{Name: tt.fieldName}

			MapFieldToConfig(field, tt.value, providerConfig)

			if providerConfig.BaseURL != tt.value {
				t.Errorf("BaseURL = %v, want %v", providerConfig.BaseURL, tt.value)
			}
		})
	}
}

// TestMapFieldToConfig_NilExtraConfig tests that ExtraConfig is initialized when nil
func TestMapFieldToConfig_NilExtraConfig(t *testing.T) {
	providerConfig := &config.ProviderConfig{
		ExtraConfig: nil,
	}

	field := generated.ConfigField{Name: "awsAccessKey"}
	MapFieldToConfig(field, "test-key", providerConfig)

	if providerConfig.ExtraConfig == nil {
		t.Fatal("ExtraConfig should be initialized, but is nil")
	}

	if got := providerConfig.ExtraConfig["aws_access_key"]; got != "test-key" {
		t.Errorf("ExtraConfig[aws_access_key] = %v, want test-key", got)
	}
}

// TestMapFieldToConfig_UnknownFields tests that unknown fields go to ExtraConfig
func TestMapFieldToConfig_UnknownFields(t *testing.T) {
	providerConfig := &config.ProviderConfig{}
	field := generated.ConfigField{Name: "customUnknownField"}
	value := "custom-value"

	MapFieldToConfig(field, value, providerConfig)

	if providerConfig.ExtraConfig == nil {
		t.Fatal("ExtraConfig is nil")
	}

	if got := providerConfig.ExtraConfig["customUnknownField"]; got != value {
		t.Errorf("ExtraConfig[customUnknownField] = %v, want %v", got, value)
	}
}

// TestValidateRequiredFields_AllProviders tests validation for representative providers
func TestValidateRequiredFields_AllProviders(t *testing.T) {
	tests := []struct {
		name           string
		providerID     string
		providerConfig config.ProviderConfig
		requiredFields []generated.ConfigField
		wantErr        bool
		errorContains  string
	}{
		{
			name:       "Cerebras - valid",
			providerID: "cerebras",
			providerConfig: config.ProviderConfig{
				APIKey: "cerebras-key",
			},
			requiredFields: []generated.ConfigField{
				{Name: "cerebrasApiKey", Required: true},
			},
			wantErr: false,
		},
		{
			name:       "Cerebras - missing API key",
			providerID: "cerebras",
			providerConfig: config.ProviderConfig{
				APIKey: "",
			},
			requiredFields: []generated.ConfigField{
				{Name: "cerebrasApiKey", Required: true},
			},
			wantErr:       true,
			errorContains: "cerebrasApiKey",
		},
		{
			name:       "DeepSeek - valid",
			providerID: "deepseek",
			providerConfig: config.ProviderConfig{
				APIKey: "deepseek-key",
			},
			requiredFields: []generated.ConfigField{
				{Name: "deepSeekApiKey", Required: true},
			},
			wantErr: false,
		},
		{
			name:       "XAI - valid",
			providerID: "xai",
			providerConfig: config.ProviderConfig{
				APIKey: "xai-key",
			},
			requiredFields: []generated.ConfigField{
				{Name: "xaiApiKey", Required: true},
			},
			wantErr: false,
		},
		{
			name:       "Groq - valid",
			providerID: "groq",
			providerConfig: config.ProviderConfig{
				APIKey: "groq-key",
			},
			requiredFields: []generated.ConfigField{
				{Name: "groqApiKey", Required: true},
			},
			wantErr: false,
		},
		{
			name:       "Ollama - valid base URL",
			providerID: "ollama",
			providerConfig: config.ProviderConfig{
				BaseURL: "http://localhost:11434",
			},
			requiredFields: []generated.ConfigField{
				{Name: "ollamaBaseUrl", Required: true},
			},
			wantErr: false,
		},
		{
			name:       "Ollama - missing base URL",
			providerID: "ollama",
			providerConfig: config.ProviderConfig{
				BaseURL: "",
			},
			requiredFields: []generated.ConfigField{
				{Name: "ollamaBaseUrl", Required: true},
			},
			wantErr:       true,
			errorContains: "ollamaBaseUrl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredFields(tt.providerID, tt.providerConfig, tt.requiredFields)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want to contain %v", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateRequiredFields_AWSBedrock tests that all 3 AWS keys are validated
func TestValidateRequiredFields_AWSBedrock(t *testing.T) {
	requiredFields := []generated.ConfigField{
		{Name: "awsAccessKey", Required: true},
		{Name: "awsSecretKey", Required: true},
		{Name: "awsRegion", Required: true},
	}

	tests := []struct {
		name           string
		providerConfig config.ProviderConfig
		wantErr        bool
		errorContains  string
	}{
		{
			name: "All AWS keys present",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"aws_access_key": "AKIAIOSFODNN7EXAMPLE",
					"aws_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					"aws_region":     "us-east-1",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing access key",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"aws_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					"aws_region":     "us-east-1",
				},
			},
			wantErr:       true,
			errorContains: "awsAccessKey",
		},
		{
			name: "Missing secret key",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"aws_access_key": "AKIAIOSFODNN7EXAMPLE",
					"aws_region":     "us-east-1",
				},
			},
			wantErr:       true,
			errorContains: "awsSecretKey",
		},
		{
			name: "Missing region",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"aws_access_key": "AKIAIOSFODNN7EXAMPLE",
					"aws_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
			wantErr:       true,
			errorContains: "awsRegion",
		},
		{
			name: "Empty access key value",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"aws_access_key": "",
					"aws_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					"aws_region":     "us-east-1",
				},
			},
			wantErr:       true,
			errorContains: "awsAccessKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredFields("bedrock", tt.providerConfig, requiredFields)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want to contain %v", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateRequiredFields_Vertex tests Vertex with 2 required fields
func TestValidateRequiredFields_Vertex(t *testing.T) {
	requiredFields := []generated.ConfigField{
		{Name: "vertexProjectId", Required: true},
		{Name: "vertexRegion", Required: true},
	}

	tests := []struct {
		name           string
		providerConfig config.ProviderConfig
		wantErr        bool
		errorContains  string
	}{
		{
			name: "Both Vertex fields present",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"vertex_project_id": "my-gcp-project",
					"vertex_region":     "us-central1",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing project ID",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"vertex_region": "us-central1",
				},
			},
			wantErr:       true,
			errorContains: "vertexProjectId",
		},
		{
			name: "Missing region",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"vertex_project_id": "my-gcp-project",
				},
			},
			wantErr:       true,
			errorContains: "vertexRegion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredFields("vertex", tt.providerConfig, requiredFields)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want to contain %v", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestValidateRequiredFields_EmptyValues tests that empty string values are detected as missing
func TestValidateRequiredFields_EmptyValues(t *testing.T) {
	tests := []struct {
		name           string
		providerConfig config.ProviderConfig
		requiredFields []generated.ConfigField
		wantErr        bool
		errorContains  string
	}{
		{
			name: "Empty API key",
			providerConfig: config.ProviderConfig{
				APIKey: "",
			},
			requiredFields: []generated.ConfigField{
				{Name: "cerebrasApiKey", Required: true},
			},
			wantErr:       true,
			errorContains: "cerebrasApiKey",
		},
		{
			name: "Empty base URL",
			providerConfig: config.ProviderConfig{
				BaseURL: "",
			},
			requiredFields: []generated.ConfigField{
				{Name: "ollamaBaseUrl", Required: true},
			},
			wantErr:       true,
			errorContains: "ollamaBaseUrl",
		},
		{
			name: "Empty AWS field in ExtraConfig",
			providerConfig: config.ProviderConfig{
				ExtraConfig: map[string]string{
					"aws_access_key": "",
				},
			},
			requiredFields: []generated.ConfigField{
				{Name: "awsAccessKey", Required: true},
			},
			wantErr:       true,
			errorContains: "awsAccessKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredFields("test-provider", tt.providerConfig, tt.requiredFields)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want to contain %v", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
