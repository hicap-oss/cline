# CLI Model Listing Feature

## Overview

The Cline CLI now supports interactive model listing during provider configuration. When prompted for a model ID, users can type `list` to see all available models and select by number or name.

## Usage

### Setup Wizard

```bash
./bin/cline setup
```

When configuring a provider, you'll see:

```
Enter model ID or type 'list' to see available models:
> list
```

### Auth Wizard

```bash
./bin/cline auth
```

Select "Configure API provider" and follow the same process.

## Supported Providers

### Providers with Dynamic Model Fetching
These providers fetch models from their API in real-time:

- **OpenRouter** - Fetches from https://openrouter.ai/api/v1/models
- **Ollama** - Fetches from {baseUrl}/api/tags
- **OpenAI** - Fetches from {baseUrl}/v1/models
- **OpenAI Native** - Fetches from {baseUrl}/v1/models
- **Groq** - Fetches from {baseUrl}/v1/models

### Providers with Static Models
These providers use hardcoded model lists:

- **Anthropic**
- **Cerebras**
- **X AI (Grok)**
- **AWS Bedrock**

## Features

### 1. Model Listing

Type `list` to see available models:

```
Available models:
1. claude-sonnet-4-5-20250929 - Claude Sonnet 4.5 (200k context)
2. gpt-4o - GPT-4 Optimized (128k context)
3. llama-3.3-70b - Llama 3.3 70B (128k context)
...
```

### 2. Selection Methods

**By Number:**
```
Select by number or enter model ID:
> 1
```

**By Model ID:**
```
Select by number or enter model ID:
> claude-sonnet-4-5-20250929
```

### 3. Pagination

For providers with many models (e.g., OpenRouter with 100+ models), the list is paginated:

```
Available models (page 1 of 7):
1. claude-sonnet-4-5-20250929 - Claude Sonnet 4.5 (200k context)
...
15. llama-3.1-8b-instant - Llama 3.1 8B Instant (131k context)

Type 'next' to see more models, or select by number/ID
> next
```

Navigation commands:
- `next` - Go to next page
- `back` / `prev` / `previous` - Go to previous page

### 4. Smart Sorting

Models are sorted by relevance:
1. Popular models (Claude 4, GPT-4o)
2. Larger context windows
3. Alphabetical order

### 5. Automatic Fallback

If API fetching fails (timeout, network error, invalid API key), the wizard automatically falls back to hardcoded models:

```
Fetching models from API...
Failed to fetch models from API: request timeout after 10s
Showing hardcoded models instead...
```

## Technical Details

### API Timeout

All API requests have a 10-second timeout to prevent hanging. On timeout:
- Error message is displayed
- Wizard falls back to hardcoded models
- User can complete configuration without API access

### Model Data

For dynamic providers, the following information is fetched:
- Model ID
- Description
- Context window size
- Max output tokens
- Input/output pricing
- Image support capabilities

### Error Handling

The wizard handles errors gracefully:
- Network failures → Fallback to hardcoded models
- Invalid API keys → Fallback to hardcoded models  
- Timeout errors → Fallback to hardcoded models
- Invalid model selection → Prompt user to try again

## Implementation Files

### Core Package Structure

```
cli/pkg/cli/models/
├── fetcher.go              # Core interfaces and factory
├── openrouter.go           # OpenRouter API client
├── ollama.go               # Ollama API client
├── openai_compatible.go    # Generic OpenAI-compatible client
└── formatter.go            # Display formatting utilities
```

### Modified Files

- `cli/pkg/generated/providers.go` - Added SupportsModelListing and ModelListEndpoint fields
- `cli/pkg/cli/setup/wizard.go` - Updated selectModel() with list support
- `cli/pkg/cli/auth/provider_wizard.go` - Updated selectModel() with list support

## Examples

### OpenRouter Configuration

```bash
$ ./bin/cline setup
? What would you like to do? Add a new provider
? How would you like to choose a provider? View popular providers
? Select a popular provider: OpenRouter (openrouter)

Configuring OpenRouter
Setup instructions: Get your API key from https://openrouter.ai/keys

Required configuration:
? openRouterApiKey * ****************************

? Would you like to configure optional settings? No
? Enter model ID or type 'list' to see available models: list

Fetching models from API...

Available models (page 1 of 7):
1. anthropic/claude-sonnet-4-5 - Claude Sonnet 4.5 (200k context)
2. anthropic/claude-4-opus - Claude 4 Opus (200k context)
3. openai/gpt-4o - GPT-4 Optimized (128k context)
...

Select by number or enter model ID (or 'next' for more, 'back' for previous):
> 1

Selected model: anthropic/claude-sonnet-4-5
Successfully configured OpenRouter!
```

### Anthropic Configuration (Static Models)

```bash
$ ./bin/cline auth
? What would you like to do? Configure API provider
? Select a popular provider: Anthropic (Claude) (anthropic)

Configuring Anthropic (Claude)
Setup instructions: Get your API key from https://console.anthropic.com/

Required configuration:
? apiKey * ****************************

? Use default model 'claude-sonnet-4-5-20250929'? No
? Enter model ID or type 'list' to see available models: list

Available models:
1. claude-sonnet-4-5-20250929 - (200k context)
2. claude-sonnet-4-5-20250929:1m - (1000k context)
3. claude-opus-4-1-20250805 - (200k context)
...

Select by number or enter model ID:
> 2

Selected model: claude-sonnet-4-5-20250929:1m
Successfully configured Anthropic (Claude)!
```

## Testing

To test the feature manually:

1. **Test OpenRouter (API fetching):**
   ```bash
   ./bin/cline auth
   # Select OpenRouter, enter API key, type 'list'
   # Verify models are fetched from API
   # Test selection by number and by model ID
   ```

2. **Test Anthropic (Static models):**
   ```bash
   ./bin/cline setup
   # Select Anthropic, enter API key, type 'list'
   # Verify hardcoded models are shown
   # Test selection
   ```

3. **Test Offline Scenario:**
   ```bash
   ./bin/cline setup
   # Select OpenRouter, enter invalid API key, type 'list'
   # Verify fallback to hardcoded models works
   # Complete configuration successfully
   ```

4. **Test Pagination:**
   ```bash
   ./bin/cline setup
   # Select OpenRouter with valid key, type 'list'
   # Test 'next' and 'back' navigation
   # Select model from different pages
   ```

## Future Enhancements

Potential improvements for future iterations:

1. **Caching** - Cache fetched models to avoid repeated API calls
2. **Filtering** - Add ability to filter models by capabilities (e.g., "show only vision models")
3. **Search** - Add search within model list (e.g., "search for claude")
4. **Recommendations** - Highlight recommended models with visual indicators
5. **Comparison** - Show side-by-side comparison of selected models
6. **Cost Calculator** - Estimate usage costs based on typical workloads
