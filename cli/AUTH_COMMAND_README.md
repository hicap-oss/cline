# Cline CLI Auth Command

## Overview

The `cline auth` command is the unified interface for authentication and API provider configuration in the Cline CLI. It replaces the deprecated `cline setup` command and provides enhanced functionality with a cleaner, more modern user experience.

## Features

- **Dual Functionality**: Handle both Cline account authentication and API provider configuration
- **Fast Setup**: Configure providers with a single command line
- **Interactive Menu**: Choose between authentication options when no arguments provided
- **Clean Output**: No emojis for better terminal compatibility
- **35+ Providers**: Support for all major AI providers
- **Secure Storage**: Encrypted API key storage

## Command Syntax

```bash
cline auth [provider] [key]
```

## Usage Modes

### 1. Interactive Menu (No Arguments)

```bash
cline auth
```

Displays a menu with two options:
1. Authenticate with Cline account
2. Configure API provider

### 2. Provider Configuration (One Argument)

```bash
cline auth anthropic
```

Configures a specific provider, prompting interactively for the API key and other required fields.

### 3. Fast Setup (Two Arguments)

```bash
cline auth anthropic sk-ant-api03-xxx
```

Performs quick setup with both provider ID and API key provided on the command line.

## Examples

### Authenticate with Cline Account

```bash
$ cline auth
? What would you like to do?
  > Authenticate with Cline account
    Configure API provider

Authenticating with Cline...
You are signed in!
```

### Configure Anthropic (Interactive)

```bash
$ cline auth anthropic
Configuring Anthropic
? apiKey (Your Anthropic API key) * [hidden input]
Using default model: claude-3-5-sonnet-20241022

Successfully configured Anthropic!
Configuration saved to /Users/username/Documents/cline-cli-config.json
```

### Fast Setup with OpenRouter

```bash
$ cline auth openrouter sk-or-v1-xxx
Configuring OpenRouter...
Using default model: anthropic/claude-3.5-sonnet

Successfully configured OpenRouter!
Configuration saved to /Users/username/Documents/cline-cli-config.json
```

### Configure Provider with Multiple Required Fields

Some providers require additional configuration beyond the API key:

```bash
$ cline auth bedrock
Configuring AWS Bedrock
? awsAccessKey (AWS Access Key ID) * [hidden input]
? awsSecretKey (AWS Secret Access Key) * [hidden input]
? awsRegion (AWS region) us-east-1
Using default model: anthropic.claude-3-5-sonnet-20241022-v2:0

Successfully configured AWS Bedrock!
Configuration saved to /Users/username/Documents/cline-cli-config.json
```

## Supported Providers

The auth command supports 35+ AI providers including:

- **Anthropic** - Claude models (claude-3-5-sonnet, claude-3-opus, etc.)
- **OpenRouter** - Access to multiple providers
- **OpenAI** - GPT models
- **Google** - Gemini models
- **AWS Bedrock** - Enterprise AI models
- **Azure** - OpenAI on Azure
- **Ollama** - Local models
- **LM Studio** - Local models
- **And many more...**

Run `cline auth` and select "Configure API provider" to browse all available providers.

## Configuration File

All configurations are saved to:
```
/Users/[username]/Documents/cline-cli-config.json
```

The configuration file stores:
- API keys (encrypted)
- Model selections
- Base URLs (if customized)
- Default provider
- Additional provider-specific settings

## Advanced Usage

### Multiple Provider Configuration

Configure multiple providers for different use cases:

```bash
# Configure Anthropic
cline auth anthropic sk-ant-xxx

# Configure OpenRouter as backup
cline auth openrouter sk-or-xxx

# Configure local Ollama for offline work
cline auth ollama
```

### Setting Default Provider

Use the interactive wizard to set a default provider:

```bash
$ cline auth
? What would you like to do? Configure API provider

# In the wizard menu:
? What would you like to do? Set default provider
? Select default provider: Anthropic (anthropic)

Set anthropic as default provider
```

### Testing Provider Connections

```bash
$ cline auth
? What would you like to do? Configure API provider

# In the wizard menu:
? What would you like to do? Test provider connections

Testing provider connections...
Note: This is a basic configuration validation. Full API testing requires actual API calls.
Testing Anthropic (anthropic)... Configuration valid
Testing OpenRouter (openrouter)... Configuration valid
```

## Error Handling

### Invalid Provider

```bash
$ cline auth invalid-provider
Error: provider 'invalid-provider' not found. Use 'cline auth' to see available providers
```

### Missing Required Fields

```bash
$ cline auth anthropic
Configuring Anthropic
? apiKey (Your Anthropic API key) * [press Enter to skip]
Error: field apiKey is required
```

## Command Options

The auth command inherits global options from the CLI:

```bash
cline auth --verbose          # Enable verbose output
cline auth --help             # Show help message
cline auth --output-format=json  # JSON output (where applicable)
```

## Comparison with Setup Command

| Feature | `cline setup` (deprecated) | `cline auth` (new) |
|---------|---------------------------|-------------------|
| Provider configuration | ✅ | ✅ |
| Cline account auth | ❌ | ✅ |
| Fast setup | ❌ | ✅ |
| Emojis in output | ✅ | ❌ (removed) |
| Interactive menu | ✅ | ✅ |
| Multiple providers | ✅ | ✅ |
| Status | Deprecated | Active |

## Security Notes

- API keys are never displayed in plaintext in terminal output
- Configuration file stores keys securely
- Passwords fields use hidden input in interactive prompts
- Configuration file permissions should be set appropriately

## Troubleshooting

### Command Not Found

Ensure the Cline CLI is properly installed and in your PATH:

```bash
which cline
cline --version
```

### Configuration File Issues

Check if configuration file exists and is readable:

```bash
cat ~/Documents/cline-cli-config.json
```

### Provider Not Working

1. Verify API key is correct
2. Test the provider: Use the wizard's "Test provider connections" option
3. Check provider status on their website
4. Ensure you have credits/quota available

## Migration from Setup Command

See [AUTH_MIGRATION_GUIDE.md](./AUTH_MIGRATION_GUIDE.md) for detailed migration instructions.

Quick summary:
- Replace `cline setup` with `cline auth`
- All existing configurations remain compatible
- No data migration needed

## Getting Help

```bash
# Show auth command help
cline auth --help

# Show all available commands
cline --help

# View migration guide
cat cli/AUTH_MIGRATION_GUIDE.md
```

## Architecture

The auth command is implemented across several modules:

- `cli/pkg/cli/auth.go` - Main command routing
- `cli/pkg/cli/auth/menu.go` - Interactive menu
- `cli/pkg/cli/auth/provider_wizard.go` - Provider configuration wizard
- `cli/pkg/cli/auth/fast_setup.go` - Fast setup functionality

## Contributing

When adding new features to the auth command:

1. Ensure backwards compatibility
2. Follow the no-emoji output policy
3. Add examples to this documentation
4. Update the migration guide if needed
5. Test all command variations

## Changelog

### Version 1.0 (Current)
- Initial release of unified auth command
- Fast setup feature for provider configuration
- Removed all emojis from output
- Deprecated setup command
- Added interactive menu for dual functionality

## License

See main Cline CLI license.
