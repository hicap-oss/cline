package setup

import (
	"fmt"

	"github.com/cline/cli/pkg/cli/config"
	"github.com/cline/cli/pkg/generated"
)

// MapFieldToConfig maps a configuration field to the appropriate location in ProviderConfig
// This handles the complex mapping logic for multi-key providers like AWS Bedrock
func MapFieldToConfig(field generated.ConfigField, value string, providerConfig *config.ProviderConfig) {
	fieldName := field.Name

	// Explicit mapping based on field names and their purposes
	switch fieldName {
	// Generic API key field (used by providers without specific prefix)
	case "apiKey":
		providerConfig.APIKey = value

	// Provider-specific API keys
	case "anthropicApiKey", "cerebrasApiKey", "deepSeekApiKey", "xaiApiKey",
		"groqApiKey", "geminiApiKey", "openAiApiKey", "openAiNativeApiKey",
		"openRouterApiKey", "qwenApiKey", "doubaoApiKey", "mistralApiKey",
		"fireworksApiKey", "huggingFaceApiKey", "moonshotApiKey",
		"sambanovaApiKey", "sapAiCoreApiKey", "basetenApiKey",
		"nebiusApiKey", "askSageApiKey", "togetherApiKey",
		"liteLLMApiKey", "difyApiKey", "zaiApiKey", "requestyApiKey":
		providerConfig.APIKey = value

	// AWS Bedrock requires multiple keys stored in ExtraConfig
	case "awsAccessKey":
		if providerConfig.ExtraConfig == nil {
			providerConfig.ExtraConfig = make(map[string]string)
		}
		providerConfig.ExtraConfig["aws_access_key"] = value

	case "awsSecretKey":
		if providerConfig.ExtraConfig == nil {
			providerConfig.ExtraConfig = make(map[string]string)
		}
		providerConfig.ExtraConfig["aws_secret_key"] = value

	case "awsSessionToken":
		if providerConfig.ExtraConfig == nil {
			providerConfig.ExtraConfig = make(map[string]string)
		}
		providerConfig.ExtraConfig["aws_session_token"] = value

	case "awsRegion":
		if providerConfig.ExtraConfig == nil {
			providerConfig.ExtraConfig = make(map[string]string)
		}
		providerConfig.ExtraConfig["aws_region"] = value

	// Vertex AI fields
	case "vertexProjectId":
		if providerConfig.ExtraConfig == nil {
			providerConfig.ExtraConfig = make(map[string]string)
		}
		providerConfig.ExtraConfig["vertex_project_id"] = value

	case "vertexRegion":
		if providerConfig.ExtraConfig == nil {
			providerConfig.ExtraConfig = make(map[string]string)
		}
		providerConfig.ExtraConfig["vertex_region"] = value

	// Base URL fields (local providers like Ollama, LM Studio)
	case "ollamaBaseUrl", "lmStudioBaseUrl", "openAiBaseUrl",
		"liteLLMBaseUrl", "fireworksBaseUrl":
		providerConfig.BaseURL = value

	// All other fields go into ExtraConfig
	default:
		if providerConfig.ExtraConfig == nil {
			providerConfig.ExtraConfig = make(map[string]string)
		}
		// Use field name as-is for extra config
		providerConfig.ExtraConfig[fieldName] = value
	}
}

// ValidateRequiredFields ensures all required fields for a provider have been collected
func ValidateRequiredFields(providerID string, providerConfig config.ProviderConfig, requiredFields []generated.ConfigField) error {
	// Check each required field has a non-empty value
	for _, field := range requiredFields {
		hasValue := false

		switch field.Name {
		case "apiKey", "anthropicApiKey", "cerebrasApiKey", "deepSeekApiKey",
			"xaiApiKey", "groqApiKey", "geminiApiKey", "openAiApiKey",
			"openAiNativeApiKey", "openRouterApiKey", "qwenApiKey",
			"doubaoApiKey", "mistralApiKey", "fireworksApiKey",
			"huggingFaceApiKey", "moonshotApiKey", "sambanovaApiKey",
			"sapAiCoreApiKey", "basetenApiKey", "nebiusApiKey",
			"askSageApiKey", "togetherApiKey", "liteLLMApiKey",
			"difyApiKey", "zaiApiKey", "requestyApiKey":
			hasValue = providerConfig.APIKey != ""

		case "awsAccessKey":
			hasValue = providerConfig.ExtraConfig["aws_access_key"] != ""

		case "awsSecretKey":
			hasValue = providerConfig.ExtraConfig["aws_secret_key"] != ""

		case "awsRegion":
			hasValue = providerConfig.ExtraConfig["aws_region"] != ""

		case "vertexProjectId":
			hasValue = providerConfig.ExtraConfig["vertex_project_id"] != ""

		case "vertexRegion":
			hasValue = providerConfig.ExtraConfig["vertex_region"] != ""

		case "ollamaBaseUrl", "lmStudioBaseUrl":
			hasValue = providerConfig.BaseURL != ""

		default:
			// Check in ExtraConfig
			value, exists := providerConfig.ExtraConfig[field.Name]
			hasValue = exists && value != ""
		}

		if !hasValue {
			return fmt.Errorf("required field '%s' is missing or empty", field.Name)
		}
	}

	return nil
}
