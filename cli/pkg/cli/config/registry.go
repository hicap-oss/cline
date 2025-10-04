package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cline/cli/pkg/generated"
)

// ProviderRegistry manages available providers and their definitions
type ProviderRegistry struct {
	definitions map[string]generated.ProviderDefinition
	configFields []generated.ConfigField
	modelDefinitions map[string]map[string]generated.ModelInfo
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() (*ProviderRegistry, error) {
	// Load provider definitions from generated code
	definitions, err := generated.GetProviderDefinitions()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider definitions: %w", err)
	}

	configFields, err := generated.GetConfigFields()
	if err != nil {
		return nil, fmt.Errorf("failed to load config fields: %w", err)
	}

	modelDefinitions, err := generated.GetModelDefinitions()
	if err != nil {
		return nil, fmt.Errorf("failed to load model definitions: %w", err)
	}

	return &ProviderRegistry{
		definitions: definitions,
		configFields: configFields,
		modelDefinitions: modelDefinitions,
	}, nil
}

// GetAllProviders returns all available provider IDs
func (pr *ProviderRegistry) GetAllProviders() []string {
	providers := make([]string, 0, len(pr.definitions))
	for id := range pr.definitions {
		providers = append(providers, id)
	}
	sort.Strings(providers)
	return providers
}

// GetProviderDefinition returns the definition for a specific provider
func (pr *ProviderRegistry) GetProviderDefinition(providerID string) (*generated.ProviderDefinition, error) {
	def, exists := pr.definitions[providerID]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerID)
	}
	return &def, nil
}

// GetProvidersByCategory returns providers grouped by category
func (pr *ProviderRegistry) GetProvidersByCategory() map[string][]string {
	categories := map[string][]string{
		"Major Cloud Providers": {"anthropic", "openai-native", "gemini", "bedrock", "vertex"},
		"Aggregators": {"openrouter", "litellm", "together", "fireworks"},
		"Local/Self-Hosted": {"ollama", "lmstudio"},
		"Specialized": {"deepseek", "qwen", "mistral", "xai", "cerebras", "groq"},
		"Enterprise": {"sapaicore", "asksage", "vercel-ai-gateway"},
		"Other": {},
	}

	// Add any providers not in predefined categories to "Other"
	allProviders := pr.GetAllProviders()
	categorized := make(map[string]bool)
	
	for _, providerList := range categories {
		for _, provider := range providerList {
			categorized[provider] = true
		}
	}

	for _, provider := range allProviders {
		if !categorized[provider] {
			categories["Other"] = append(categories["Other"], provider)
		}
	}

	// Remove empty categories
	for category, providers := range categories {
		if len(providers) == 0 {
			delete(categories, category)
		} else {
			sort.Strings(providers)
		}
	}

	return categories
}

// GetPopularProviders returns a list of popular/recommended providers
func (pr *ProviderRegistry) GetPopularProviders() []string {
	return []string{
		"cline",
		"openrouter",
		"openai", 
		"anthropic",
		"xai",
		"ollama",
		"gemini",
		"deepseek",
		"groq",
		"cerebras",
	}
}

// SearchProviders searches for providers by name or description
func (pr *ProviderRegistry) SearchProviders(query string) []string {
	query = strings.ToLower(query)
	var matches []string

	for id, def := range pr.definitions {
		// Search in ID
		if strings.Contains(strings.ToLower(id), query) {
			matches = append(matches, id)
			continue
		}

		// Search in name
		if strings.Contains(strings.ToLower(def.Name), query) {
			matches = append(matches, id)
			continue
		}

		// Search in setup instructions
		if strings.Contains(strings.ToLower(def.SetupInstructions), query) {
			matches = append(matches, id)
			continue
		}
	}

	sort.Strings(matches)
	return matches
}

// GetRequiredFields returns required configuration fields for a provider
func (pr *ProviderRegistry) GetRequiredFields(providerID string) ([]generated.ConfigField, error) {
	def, err := pr.GetProviderDefinition(providerID)
	if err != nil {
		return nil, err
	}

	return def.RequiredFields, nil
}

// GetOptionalFields returns optional configuration fields for a provider
func (pr *ProviderRegistry) GetOptionalFields(providerID string) ([]generated.ConfigField, error) {
	def, err := pr.GetProviderDefinition(providerID)
	if err != nil {
		return nil, err
	}

	return def.OptionalFields, nil
}

// GetProviderModels returns available models for a provider
func (pr *ProviderRegistry) GetProviderModels(providerID string) (map[string]generated.ModelInfo, error) {
	def, err := pr.GetProviderDefinition(providerID)
	if err != nil {
		return nil, err
	}

	return def.Models, nil
}

// GetDefaultModel returns the default model for a provider
func (pr *ProviderRegistry) GetDefaultModel(providerID string) (string, error) {
	def, err := pr.GetProviderDefinition(providerID)
	if err != nil {
		return "", err
	}

	return def.DefaultModelID, nil
}

// ValidateProviderConfig validates a provider configuration
func (pr *ProviderRegistry) ValidateProviderConfig(config ProviderConfig) error {
	def, err := pr.GetProviderDefinition(config.ID)
	if err != nil {
		return err
	}

	// Check required fields
	for _, field := range def.RequiredFields {
		switch field.Name {
		case "apiKey":
			if config.APIKey == "" {
				return fmt.Errorf("API key is required for provider %s", config.ID)
			}
		case "baseUrl":
			if config.BaseURL == "" && field.Required {
				return fmt.Errorf("base URL is required for provider %s", config.ID)
			}
		}
	}

	// Validate model ID if models are defined
	if len(def.Models) > 0 && config.ModelID != "" {
		if _, exists := def.Models[config.ModelID]; !exists {
			return fmt.Errorf("model %s not found for provider %s", config.ModelID, config.ID)
		}
	}

	return nil
}

// GetModelsByCapability returns models that support specific capabilities
func (pr *ProviderRegistry) GetModelsByCapability(providerID string, capability string) ([]string, error) {
	models, err := pr.GetProviderModels(providerID)
	if err != nil {
		return nil, err
	}

	var matching []string
	for modelID, modelInfo := range models {
		switch capability {
		case "images":
			if modelInfo.SupportsImages {
				matching = append(matching, modelID)
			}
		case "prompt_cache":
			if modelInfo.SupportsPromptCache {
				matching = append(matching, modelID)
			}
		case "large_context":
			if modelInfo.ContextWindow >= 100000 {
				matching = append(matching, modelID)
			}
		case "free":
			if modelInfo.InputPrice == 0 && modelInfo.OutputPrice == 0 {
				matching = append(matching, modelID)
			}
		}
	}

	sort.Strings(matching)
	return matching, nil
}

// GetProviderStats returns statistics about providers
func (pr *ProviderRegistry) GetProviderStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalProviders := len(pr.definitions)
	totalModels := 0
	providersWithModels := 0
	providersWithDynamicModels := 0
	totalConfigFields := len(pr.configFields)

	for _, def := range pr.definitions {
		if len(def.Models) > 0 {
			providersWithModels++
			totalModels += len(def.Models)
		}
		if def.HasDynamicModels {
			providersWithDynamicModels++
		}
	}

	stats["total_providers"] = totalProviders
	stats["total_models"] = totalModels
	stats["providers_with_models"] = providersWithModels
	stats["providers_with_dynamic_models"] = providersWithDynamicModels
	stats["total_config_fields"] = totalConfigFields

	return stats
}

// GetProviderComparison returns a comparison of providers
func (pr *ProviderRegistry) GetProviderComparison(providerIDs []string) (map[string]interface{}, error) {
	comparison := make(map[string]interface{})

	for _, providerID := range providerIDs {
		def, err := pr.GetProviderDefinition(providerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get definition for provider %s: %w", providerID, err)
		}

		providerInfo := map[string]interface{}{
			"name": def.Name,
			"setup_instructions": def.SetupInstructions,
			"has_dynamic_models": def.HasDynamicModels,
			"model_count": len(def.Models),
			"required_fields": len(def.RequiredFields),
			"optional_fields": len(def.OptionalFields),
		}

		// Add model capabilities summary
		if len(def.Models) > 0 {
			supportsImages := 0
			supportsPromptCache := 0
			freeModels := 0
			largeContextModels := 0

			for _, model := range def.Models {
				if model.SupportsImages {
					supportsImages++
				}
				if model.SupportsPromptCache {
					supportsPromptCache++
				}
				if model.InputPrice == 0 && model.OutputPrice == 0 {
					freeModels++
				}
				if model.ContextWindow >= 100000 {
					largeContextModels++
				}
			}

			providerInfo["models_with_images"] = supportsImages
			providerInfo["models_with_prompt_cache"] = supportsPromptCache
			providerInfo["free_models"] = freeModels
			providerInfo["large_context_models"] = largeContextModels
		}

		comparison[providerID] = providerInfo
	}

	return comparison, nil
}

// GetRecommendedProvider returns a recommended provider based on criteria
func (pr *ProviderRegistry) GetRecommendedProvider(criteria map[string]interface{}) (string, error) {
	needsImages := false
	needsFree := false
	needsLargeContext := false
	needsLocal := false

	if val, ok := criteria["images"]; ok {
		needsImages = val.(bool)
	}
	if val, ok := criteria["free"]; ok {
		needsFree = val.(bool)
	}
	if val, ok := criteria["large_context"]; ok {
		needsLargeContext = val.(bool)
	}
	if val, ok := criteria["local"]; ok {
		needsLocal = val.(bool)
	}

	// Score providers based on criteria
	scores := make(map[string]int)

	for providerID, def := range pr.definitions {
		score := 0

		// Local preference
		if needsLocal {
			if providerID == "ollama" || providerID == "lmstudio" {
				score += 10
			} else {
				continue // Skip non-local providers if local is required
			}
		}

		// Check model capabilities
		for _, model := range def.Models {
			if needsImages && model.SupportsImages {
				score += 3
			}
			if needsFree && model.InputPrice == 0 && model.OutputPrice == 0 {
				score += 2
			}
			if needsLargeContext && model.ContextWindow >= 100000 {
				score += 2
			}
		}

		// Bonus for popular providers
		popular := pr.GetPopularProviders()
		for _, p := range popular {
			if p == providerID {
				score += 1
				break
			}
		}

		scores[providerID] = score
	}

	// Find highest scoring provider
	var bestProvider string
	var bestScore int

	for providerID, score := range scores {
		if score > bestScore {
			bestScore = score
			bestProvider = providerID
		}
	}

	if bestProvider == "" {
		return "", fmt.Errorf("no provider matches the specified criteria")
	}

	return bestProvider, nil
}

// IsValidProvider checks if a provider ID is valid
func (pr *ProviderRegistry) IsValidProvider(providerID string) bool {
	_, exists := pr.definitions[providerID]
	return exists
}

// GetProviderDisplayName returns the display name for a provider
func (pr *ProviderRegistry) GetProviderDisplayName(providerID string) string {
	if def, exists := pr.definitions[providerID]; exists {
		return def.Name
	}
	return providerID
}
