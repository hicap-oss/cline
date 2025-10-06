package models

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cline/cli/pkg/cli/config"
)

// ModelOption represents a formatted model option for display
type ModelOption struct {
	Number      int    // Display number (1-indexed)
	ModelID     string // Model identifier
	DisplayText string // Formatted display text
}

// FormatModelOption formats a single model for display
// Format: "model-id (context window)"
func FormatModelOption(modelID string, info config.ModelInfo) string {
	contextStr := formatContextWindow(info.ContextWindow)
	
	// Just show model ID and context window for a clean, concise display
	return fmt.Sprintf("%s %s", modelID, contextStr)
}

// formatContextWindow formats the context window for display
func formatContextWindow(contextWindow int) string {
	if contextWindow == 0 {
		return ""
	}
	
	// Convert to K format for readability
	if contextWindow >= 1000 {
		return fmt.Sprintf("(%dk context)", contextWindow/1000)
	}
	
	return fmt.Sprintf("(%d context)", contextWindow)
}

// FormatModelList converts a map of models to a sorted list of display options
func FormatModelList(models map[string]config.ModelInfo) []ModelOption {
	// Convert map to slice for sorting
	var options []ModelOption
	for modelID, info := range models {
		option := ModelOption{
			ModelID:     modelID,
			DisplayText: FormatModelOption(modelID, info),
		}
		options = append(options, option)
	}
	
	// Sort by relevance
	options = SortModelOptions(options, models)
	
	// Assign display numbers (1-indexed)
	for i := range options {
		options[i].Number = i + 1
	}
	
	return options
}

// SortModelOptions sorts model options by relevance
// Priority order:
// 1. Popular/recommended models (Claude, GPT-4, etc.)
// 2. Larger context windows
// 3. Alphabetical by model ID
func SortModelOptions(options []ModelOption, models map[string]config.ModelInfo) []ModelOption {
	sort.Slice(options, func(i, j int) bool {
		// Get model info for comparison
		modelI := models[options[i].ModelID]
		modelJ := models[options[j].ModelID]
		
		// Priority for popular models
		priorityI := getModelPriority(options[i].ModelID)
		priorityJ := getModelPriority(options[j].ModelID)
		
		if priorityI != priorityJ {
			return priorityI < priorityJ // Lower priority number = higher priority
		}
		
		// Sort by context window (larger first)
		if modelI.ContextWindow != modelJ.ContextWindow {
			return modelI.ContextWindow > modelJ.ContextWindow
		}
		
		// Finally, alphabetical by model ID
		return options[i].ModelID < options[j].ModelID
	})
	
	return options
}

// getModelPriority returns a priority score for sorting
// Lower numbers = higher priority
func getModelPriority(modelID string) int {
	modelLower := strings.ToLower(modelID)
	
	// Tier 1: Latest Claude and GPT models
	if strings.Contains(modelLower, "claude-sonnet-4") || strings.Contains(modelLower, "claude-4") {
		return 1
	}
	if strings.Contains(modelLower, "gpt-4o") {
		return 1
	}
	
	// Tier 2: Other Claude models
	if strings.Contains(modelLower, "claude-3.7") || strings.Contains(modelLower, "claude-3-7") {
		return 2
	}
	if strings.Contains(modelLower, "claude-3.5") || strings.Contains(modelLower, "claude-3-5") {
		return 3
	}
	if strings.Contains(modelLower, "claude") {
		return 4
	}
	
	// Tier 3: GPT models
	if strings.Contains(modelLower, "gpt-4") {
		return 5
	}
	if strings.Contains(modelLower, "gpt-3.5") {
		return 6
	}
	if strings.Contains(modelLower, "gpt") {
		return 7
	}
	
	// Tier 4: Popular open source models
	if strings.Contains(modelLower, "llama-3.3") || strings.Contains(modelLower, "llama3.3") {
		return 8
	}
	if strings.Contains(modelLower, "llama-3") || strings.Contains(modelLower, "llama3") {
		return 9
	}
	if strings.Contains(modelLower, "mixtral") {
		return 10
	}
	if strings.Contains(modelLower, "qwen") {
		return 11
	}
	if strings.Contains(modelLower, "deepseek") {
		return 12
	}
	
	// Tier 5: Everything else
	return 100
}

// PaginateModels splits models into pages for display
// Returns a slice of pages, where each page is a slice of ModelOptions
func PaginateModels(options []ModelOption, pageSize int) [][]ModelOption {
	if pageSize <= 0 {
		pageSize = 15 // Default page size
	}
	
	var pages [][]ModelOption
	for i := 0; i < len(options); i += pageSize {
		end := i + pageSize
		if end > len(options) {
			end = len(options)
		}
		pages = append(pages, options[i:end])
	}
	
	return pages
}

// FormatModelPage formats a page of models for display
func FormatModelPage(page []ModelOption, pageNum int, totalPages int) string {
	var sb strings.Builder
	
	if totalPages > 1 {
		sb.WriteString(fmt.Sprintf("\nAvailable models (page %d of %d):\n", pageNum, totalPages))
	} else {
		sb.WriteString("\nAvailable models:\n")
	}
	
	for _, option := range page {
		sb.WriteString(fmt.Sprintf("%d. %s\n", option.Number, option.DisplayText))
	}
	
	if totalPages > 1 && pageNum < totalPages {
		sb.WriteString(fmt.Sprintf("\nType 'next' to see more models, or select by number/ID\n"))
	}
	
	return sb.String()
}

// FindModelByNumberOrID finds a model by display number or model ID
// Returns the model ID and true if found, empty string and false otherwise
func FindModelByNumberOrID(input string, options []ModelOption) (string, bool) {
	// Try to parse as a number first
	var selectedNum int
	if _, err := fmt.Sscanf(input, "%d", &selectedNum); err == nil {
		// Input is a number, find by display number
		for _, option := range options {
			if option.Number == selectedNum {
				return option.ModelID, true
			}
		}
		return "", false
	}
	
	// Not a number, try to find by exact model ID match
	for _, option := range options {
		if option.ModelID == input {
			return option.ModelID, true
		}
	}
	
	// Try case-insensitive match
	inputLower := strings.ToLower(input)
	for _, option := range options {
		if strings.ToLower(option.ModelID) == inputLower {
			return option.ModelID, true
		}
	}
	
	return "", false
}
