package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cline/cli/pkg/cli/config"
)

// OpenRouterFetcher implements ModelFetcher for OpenRouter
type OpenRouterFetcher struct{}

// openRouterResponse represents the API response from OpenRouter
type openRouterResponse struct {
	Data []openRouterModel `json:"data"`
}

// openRouterModel represents a single model from OpenRouter's API
type openRouterModel struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	Description   *string                   `json:"description"`
	ContextLength *int                      `json:"context_length"`
	TopProvider   *openRouterTopProvider    `json:"top_provider"`
	Architecture  *openRouterArchitecture   `json:"architecture"`
	Pricing       *openRouterPricing        `json:"pricing"`
}

type openRouterTopProvider struct {
	MaxCompletionTokens *int `json:"max_completion_tokens"`
}

type openRouterArchitecture struct {
	Modality flexibleStringArray `json:"modality"`
}

// flexibleStringArray handles JSON values that could be either a string or []string
type flexibleStringArray []string

func (f *flexibleStringArray) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*f = arr
		return nil
	}
	
	// Try to unmarshal as single string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*f = []string{str}
		return nil
	}
	
	// If both fail, return empty array
	*f = []string{}
	return nil
}

type openRouterPricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// FetchModels retrieves available models from OpenRouter API
func (f *OpenRouterFetcher) FetchModels(apiKey string, baseURL string) (map[string]config.ModelInfo, error) {
	// Use the standard OpenRouter endpoint
	endpoint := "https://openrouter.ai/api/v1/models"
	
	// Create HTTP request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authorization header if API key is provided
	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}
	
	// Make the request with timeout
	client := createHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var apiResp openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert to ModelInfo map
	models := make(map[string]config.ModelInfo)
	for _, model := range apiResp.Data {
		modelInfo := convertOpenRouterModel(model)
		models[model.ID] = modelInfo
	}
	
	return models, nil
}

// convertOpenRouterModel converts an OpenRouter API model to ModelInfo
func convertOpenRouterModel(model openRouterModel) config.ModelInfo {
	info := config.ModelInfo{
		Description: safeString(model.Description),
	}
	
	// Set context window
	if model.ContextLength != nil {
		info.ContextWindow = *model.ContextLength
	}
	
	// Set max tokens
	if model.TopProvider != nil && model.TopProvider.MaxCompletionTokens != nil {
		info.MaxTokens = *model.TopProvider.MaxCompletionTokens
	}
	
	// Check if supports images
	if model.Architecture != nil {
		for _, modality := range model.Architecture.Modality {
			if modality == "image" {
				info.SupportsImages = true
				break
			}
		}
	}
	
	// Parse pricing (OpenRouter returns prices as strings, multiply by 1M to get per-token price)
	if model.Pricing != nil {
		if promptPrice, err := strconv.ParseFloat(model.Pricing.Prompt, 64); err == nil {
			info.InputPrice = promptPrice * 1_000_000
		}
		if completionPrice, err := strconv.ParseFloat(model.Pricing.Completion, 64); err == nil {
			info.OutputPrice = completionPrice * 1_000_000
		}
	}
	
	return info
}

// safeString safely dereferences a string pointer, returning empty string if nil
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
