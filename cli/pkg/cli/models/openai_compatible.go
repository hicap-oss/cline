package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cline/cli/pkg/cli/config"
)

// OpenAICompatibleFetcher implements ModelFetcher for OpenAI-compatible APIs
// This works for OpenAI, Groq, and other providers that follow the OpenAI API spec
type OpenAICompatibleFetcher struct{}

// openAIResponse represents the API response from OpenAI-compatible /v1/models endpoint
type openAIResponse struct {
	Object string          `json:"object"`
	Data   []openAIModel   `json:"data"`
}

// openAIModel represents a single model from OpenAI-compatible API
type openAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// FetchModels retrieves available models from OpenAI-compatible API
func (f *OpenAICompatibleFetcher) FetchModels(apiKey string, baseURL string) (map[string]config.ModelInfo, error) {
	// Use provided baseURL or default to OpenAI
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	
	// Ensure baseURL doesn't end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	endpoint := fmt.Sprintf("%s/v1/models", baseURL)
	
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
	var apiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert to ModelInfo map
	models := make(map[string]config.ModelInfo)
	for _, model := range apiResp.Data {
		modelInfo := enrichOpenAIModel(model.ID, baseURL)
		models[model.ID] = modelInfo
	}
	
	return models, nil
}

// enrichOpenAIModel creates ModelInfo with enriched data for known models
func enrichOpenAIModel(modelID string, baseURL string) config.ModelInfo {
	info := config.ModelInfo{
		Description: fmt.Sprintf("Model: %s", modelID),
		SupportsImages: false,
		InputPrice:  0,
		OutputPrice: 0,
	}
	
	// Detect provider from baseURL or model ID
	isGroq := strings.Contains(baseURL, "groq.com") || strings.Contains(baseURL, "groq")
	isOpenAI := strings.Contains(baseURL, "openai.com") || strings.Contains(baseURL, "openai")
	
	// Apply provider-specific or model-specific metadata
	if isGroq {
		enrichGroqModel(&info, modelID)
	} else if isOpenAI {
		enrichOpenAIStandardModel(&info, modelID)
	} else {
		// Generic OpenAI-compatible provider
		enrichGenericModel(&info, modelID)
	}
	
	return info
}

// enrichGroqModel adds Groq-specific model metadata
func enrichGroqModel(info *config.ModelInfo, modelID string) {
	switch {
	case strings.Contains(modelID, "llama-3.3-70b"):
		info.ContextWindow = 128000
		info.MaxTokens = 32768
		info.Description = "Meta Llama 3.3 70B"
		info.InputPrice = 0.59
		info.OutputPrice = 0.79
	case strings.Contains(modelID, "llama-3.1-8b"):
		info.ContextWindow = 128000
		info.MaxTokens = 8000
		info.Description = "Meta Llama 3.1 8B"
		info.InputPrice = 0.05
		info.OutputPrice = 0.08
	case strings.Contains(modelID, "mixtral-8x7b"):
		info.ContextWindow = 32768
		info.MaxTokens = 32768
		info.Description = "Mixtral 8x7B"
		info.InputPrice = 0.24
		info.OutputPrice = 0.24
	case strings.Contains(modelID, "gemma2-9b"):
		info.ContextWindow = 8192
		info.MaxTokens = 8192
		info.Description = "Google Gemma 2 9B"
		info.InputPrice = 0.20
		info.OutputPrice = 0.20
	default:
		info.ContextWindow = 8192
		info.MaxTokens = 4096
	}
}

// enrichOpenAIStandardModel adds OpenAI-specific model metadata
func enrichOpenAIStandardModel(info *config.ModelInfo, modelID string) {
	switch {
	case strings.Contains(modelID, "gpt-4o"):
		info.ContextWindow = 128000
		info.MaxTokens = 16384
		info.Description = "GPT-4 Optimized"
		info.SupportsImages = true
		info.InputPrice = 2.50
		info.OutputPrice = 10.00
	case strings.Contains(modelID, "gpt-4-turbo"):
		info.ContextWindow = 128000
		info.MaxTokens = 4096
		info.Description = "GPT-4 Turbo"
		info.SupportsImages = true
		info.InputPrice = 10.00
		info.OutputPrice = 30.00
	case strings.Contains(modelID, "gpt-4"):
		info.ContextWindow = 8192
		info.MaxTokens = 4096
		info.Description = "GPT-4"
		info.InputPrice = 30.00
		info.OutputPrice = 60.00
	case strings.Contains(modelID, "gpt-3.5-turbo"):
		info.ContextWindow = 16385
		info.MaxTokens = 4096
		info.Description = "GPT-3.5 Turbo"
		info.InputPrice = 0.50
		info.OutputPrice = 1.50
	case strings.Contains(modelID, "o1"):
		info.ContextWindow = 200000
		info.MaxTokens = 100000
		info.Description = "OpenAI o1"
		info.InputPrice = 15.00
		info.OutputPrice = 60.00
	default:
		info.ContextWindow = 8192
		info.MaxTokens = 4096
	}
}

// enrichGenericModel adds generic defaults for unknown OpenAI-compatible providers
func enrichGenericModel(info *config.ModelInfo, modelID string) {
	// Try to infer from model name patterns
	modelLower := strings.ToLower(modelID)
	
	switch {
	case strings.Contains(modelLower, "gpt-4"):
		info.ContextWindow = 128000
		info.MaxTokens = 4096
		info.Description = "GPT-4 compatible model"
	case strings.Contains(modelLower, "gpt-3.5"):
		info.ContextWindow = 16385
		info.MaxTokens = 4096
		info.Description = "GPT-3.5 compatible model"
	case strings.Contains(modelLower, "claude"):
		info.ContextWindow = 200000
		info.MaxTokens = 8192
		info.Description = "Claude compatible model"
	case strings.Contains(modelLower, "llama"):
		info.ContextWindow = 8192
		info.MaxTokens = 4096
		info.Description = "Llama compatible model"
	case strings.Contains(modelLower, "mistral") || strings.Contains(modelLower, "mixtral"):
		info.ContextWindow = 32768
		info.MaxTokens = 8192
		info.Description = "Mistral compatible model"
	default:
		info.ContextWindow = 8192
		info.MaxTokens = 4096
		info.Description = fmt.Sprintf("OpenAI-compatible model: %s", modelID)
	}
}
