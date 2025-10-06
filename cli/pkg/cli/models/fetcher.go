package models

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cline/cli/pkg/cli/config"
	"github.com/cline/cli/pkg/generated"
)

// ModelFetcher defines the interface for fetching models from various providers
type ModelFetcher interface {
	// FetchModels retrieves available models from the provider's API
	// Returns a map of model ID to ModelInfo for easy lookup
	FetchModels(apiKey string, baseURL string) (map[string]config.ModelInfo, error)
}

// httpTimeout is the maximum time to wait for API responses
const httpTimeout = 10 * time.Second

// createHTTPClient creates a configured HTTP client with timeout
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: httpTimeout,
	}
}

// createHTTPClientWithContext creates an HTTP client with a context for cancellation
func createHTTPClientWithContext(ctx context.Context) *http.Client {
	return &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			// Add context-aware transport if needed
		},
	}
}

// GetModelFetcher returns the appropriate ModelFetcher implementation for a provider
// Returns nil if the provider doesn't support dynamic model fetching
func GetModelFetcher(providerID string) ModelFetcher {
	switch providerID {
	case "openrouter":
		return &OpenRouterFetcher{}
	case "ollama":
		return &OllamaFetcher{}
	case "openai", "openai-native", "groq":
		return &OpenAICompatibleFetcher{}
	default:
		return nil
	}
}

// FetchModelsForProvider fetches models for a given provider definition
// This is a high-level convenience function that handles provider detection and fallback
func FetchModelsForProvider(def *generated.ProviderDefinition, apiKey string, baseURL string) (map[string]config.ModelInfo, error) {
	// Check if provider supports dynamic model fetching
	if !def.HasDynamicModels {
		return getHardcodedModels(def), nil
	}

	// Get the appropriate fetcher
	fetcher := GetModelFetcher(def.ID)
	if fetcher == nil {
		return getHardcodedModels(def), nil
	}

	// Try to fetch models from API
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	models, err := fetchWithContext(ctx, fetcher, apiKey, baseURL)
	if err != nil {
		// On error, fall back to hardcoded models
		return getHardcodedModels(def), fmt.Errorf("failed to fetch models from API: %w", err)
	}

	// If no models were returned, fall back to hardcoded models
	if len(models) == 0 {
		return getHardcodedModels(def), fmt.Errorf("API returned no models")
	}

	return models, nil
}

// fetchWithContext wraps the fetcher call with context support
func fetchWithContext(ctx context.Context, fetcher ModelFetcher, apiKey string, baseURL string) (map[string]config.ModelInfo, error) {
	// Create a channel to receive the result
	type result struct {
		models map[string]config.ModelInfo
		err    error
	}
	resultChan := make(chan result, 1)

	// Run the fetch in a goroutine
	go func() {
		models, err := fetcher.FetchModels(apiKey, baseURL)
		resultChan <- result{models: models, err: err}
	}()

	// Wait for either the result or context cancellation
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("request timeout after %v", httpTimeout)
	case res := <-resultChan:
		return res.models, res.err
	}
}

// getHardcodedModels extracts hardcoded models from a provider definition
func getHardcodedModels(def *generated.ProviderDefinition) map[string]config.ModelInfo {
	models := make(map[string]config.ModelInfo)
	
	for modelID, model := range def.Models {
		models[modelID] = config.ModelInfo{
			MaxTokens:      model.MaxTokens,
			ContextWindow:  model.ContextWindow,
			SupportsImages: model.SupportsImages,
			InputPrice:     model.InputPrice,
			OutputPrice:    model.OutputPrice,
			Description:    model.Description,
		}
	}
	
	return models
}
