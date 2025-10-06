package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cline/cli/pkg/cli/config"
)

// OllamaFetcher implements ModelFetcher for Ollama
type OllamaFetcher struct{}

// ollamaResponse represents the API response from Ollama /api/tags endpoint
type ollamaResponse struct {
	Models []ollamaModel `json:"models"`
}

// ollamaModel represents a single model from Ollama's API
type ollamaModel struct {
	Name       string                 `json:"name"`
	Model      string                 `json:"model"`
	ModifiedAt string                 `json:"modified_at"`
	Size       int64                  `json:"size"`
	Digest     string                 `json:"digest"`
	Details    map[string]interface{} `json:"details"`
}

// FetchModels retrieves available models from Ollama API
func (f *OllamaFetcher) FetchModels(apiKey string, baseURL string) (map[string]config.ModelInfo, error) {
	// Use provided baseURL or default to localhost
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	
	// Ensure baseURL doesn't end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	endpoint := fmt.Sprintf("%s/api/tags", baseURL)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
	var apiResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert to ModelInfo map, using model name as the key
	models := make(map[string]config.ModelInfo)
	seenModels := make(map[string]bool)
	
	for _, model := range apiResp.Models {
		// Use the name field as the model ID, deduplicate
		modelID := model.Name
		if seenModels[modelID] {
			continue
		}
		seenModels[modelID] = true
		
		// Ollama doesn't provide detailed model info via API, so we create basic entries
		// The context window and other details would need to be inferred from model name
		// or fetched from a separate endpoint
		modelInfo := config.ModelInfo{
			Description: fmt.Sprintf("Ollama model: %s", modelID),
			// Ollama models typically have varying context windows
			// We'll set a reasonable default that users can override
			ContextWindow: 4096, // Conservative default
			MaxTokens:     2048, // Conservative default
			SupportsImages: false, // Would need to check model capabilities
			InputPrice:    0, // Local models are free
			OutputPrice:   0, // Local models are free
		}
		
		// Try to infer context window from common model names
		modelInfo.ContextWindow = inferOllamaContextWindow(modelID)
		
		models[modelID] = modelInfo
	}
	
	return models, nil
}

// inferOllamaContextWindow tries to infer context window size from model name
func inferOllamaContextWindow(modelName string) int {
	nameLower := strings.ToLower(modelName)
	
	// Check for specific context window indicators in model name
	if strings.Contains(nameLower, "32k") || strings.Contains(nameLower, "32000") {
		return 32768
	}
	if strings.Contains(nameLower, "16k") || strings.Contains(nameLower, "16000") {
		return 16384
	}
	if strings.Contains(nameLower, "8k") || strings.Contains(nameLower, "8000") {
		return 8192
	}
	
	// Model-specific defaults based on known models
	switch {
	case strings.Contains(nameLower, "llama3") || strings.Contains(nameLower, "llama-3"):
		return 8192
	case strings.Contains(nameLower, "llama2") || strings.Contains(nameLower, "llama-2"):
		return 4096
	case strings.Contains(nameLower, "mistral"):
		return 8192
	case strings.Contains(nameLower, "mixtral"):
		return 32768
	case strings.Contains(nameLower, "codellama"):
		return 16384
	case strings.Contains(nameLower, "phi"):
		return 2048
	case strings.Contains(nameLower, "gemma"):
		return 8192
	case strings.Contains(nameLower, "qwen"):
		return 32768
	case strings.Contains(nameLower, "deepseek"):
		return 16384
	default:
		return 4096 // Conservative default
	}
}
