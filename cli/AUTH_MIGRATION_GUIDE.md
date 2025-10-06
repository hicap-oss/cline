# Auth Command Migration Guide

## Overview

The Cline CLI has consolidated authentication and provider configuration under the `cline auth` command. The `cline setup` command is now deprecated and will be removed in a future version.

## What's Changed

### New Unified Command: `cline auth`

The `cline auth` command now handles both:
1. **Cline Account Authentication** - Sign in to your Cline account
2. **API Provider Configuration** - Configure API providers (previously `cline setup`)

### Command Usage

#### Interactive Menu (No Arguments)
```bash
cline auth
```
Shows a menu with two options:
- Authenticate with Cline account
- Configure API provider

#### Fast Provider Setup (With Arguments)
```bash
# Configure provider with interactive prompts
cline auth anthropic

# Fast setup with provider and API key
cline auth anthropic sk-ant-xxx

# Another example
cline auth openrouter sk-or-xxx
```

### Migration Path

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| `cline setup` | `cline auth` | Interactive menu, then select "Configure API provider" |
| `cline setup` | `cline auth [provider]` | Direct provider configuration |
| N/A | `cline auth [provider] [key]` | New fast setup feature |

## Benefits of the New Command

1. **Unified Experience**: Single command for all authentication needs
2. **Fast Setup**: Configure providers with a single command
3. **Better Organization**: Clear separation between Cline auth and provider config
4. **Modern Design**: Clean output without emojis for better terminal compatibility

## Deprecated Features

The `cline setup` command is **deprecated** but still functional for backwards compatibility. It will show a deprecation notice when used.

### Timeline
- **Current**: `cline setup` works but shows deprecation notice
- **Future Release**: `cline setup` will be removed entirely

## Examples

### Example 1: Configure Anthropic Provider (Interactive)
```bash
$ cline auth
? What would you like to do?
  > Authenticate with Cline account
    Configure API provider

# Select "Configure API provider"
# Then follow the wizard prompts
```

### Example 2: Fast Setup with Anthropic
```bash
$ cline auth anthropic sk-ant-api03-xxx
Configuring Anthropic...
Using default model: claude-3-5-sonnet-20241022

Successfully configured Anthropic!
Configuration saved to /Users/username/Documents/cline-cli-config.json
```

### Example 3: Configure Without API Key
```bash
$ cline auth openrouter
Configuring OpenRouter
? apiKey (Your OpenRouter API key) * [hidden input]
Using default model: anthropic/claude-3.5-sonnet

Successfully configured OpenRouter!
Configuration saved to /Users/username/Documents/cline-cli-config.json
```

### Example 4: Authenticate with Cline Account
```bash
$ cline auth
? What would you like to do?
  > Authenticate with Cline account
    Configure API provider

# Browser opens for authentication
You are signed in!
```

## Features Preserved

All features from `cline setup` are preserved in `cline auth`:
- 35+ supported AI providers
- Secure API key storage
- Multiple provider configuration
- Model selection
- Provider testing
- Default provider setting
- Optional configuration

## Clean Terminal Output

The new command removes all emojis from output for:
- Better terminal compatibility
- Cleaner, more professional appearance
- Improved screen reader support
- Consistent experience across platforms

### Before (with emojis)
```
🚀 Welcome to Cline CLI Setup!
✅ Successfully configured Anthropic!
🎉 Setup complete!
```

### After (clean)
```
Welcome to Cline CLI Setup!
Successfully configured Anthropic!
Setup complete!
```

## Troubleshooting

### "Provider not found" Error
```bash
$ cline auth invalid-provider
Error: provider 'invalid-provider' not found. Use 'cline auth' to see available providers
```

**Solution**: Run `cline auth` without arguments to see the interactive menu and browse available providers.

### Multiple Providers Configuration
Both the old `cline setup` and new `cline auth` commands use the same configuration file, so you can:
- Use `cline auth` to add new providers
- Existing providers configured with `cline setup` remain accessible
- No migration of existing configurations needed

## Getting Help

```bash
# Show auth command help
cline auth --help

# Show general CLI help
cline --help

# Report issues
cline --help  # includes feedback information
```

## Summary

✅ **Do**: Use `cline auth` for all authentication and provider configuration
❌ **Don't**: Use `cline setup` (deprecated)

The migration is seamless - your existing configuration will continue to work with the new command.
