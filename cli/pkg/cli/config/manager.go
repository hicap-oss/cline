package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// CLIConfig represents the complete CLI configuration
type CLIConfig struct {
	Version         string                    `yaml:"version"`
	EncryptionNote  string                    `yaml:"encryption_note"`
	DefaultProvider string                    `yaml:"default_provider"`
	Providers       map[string]ProviderConfig `yaml:"providers"`
	CreatedAt       time.Time                 `yaml:"created_at"`
	UpdatedAt       time.Time                 `yaml:"updated_at"`
}

// ProviderConfig represents a configured API provider
type ProviderConfig struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	APIKey      string            `yaml:"api_key"` // encrypted
	BaseURL     string            `yaml:"base_url,omitempty"`
	ModelID     string            `yaml:"model_id"`
	ModelInfo   ModelInfo         `yaml:"model_info"`
	ExtraConfig map[string]string `yaml:"extra_config,omitempty"`
}

// ModelInfo represents model capabilities and pricing
type ModelInfo struct {
	MaxTokens        int     `yaml:"max_tokens,omitempty"`
	ContextWindow    int     `yaml:"context_window,omitempty"`
	SupportsImages   bool    `yaml:"supports_images"`
	InputPrice       float64 `yaml:"input_price,omitempty"`
	OutputPrice      float64 `yaml:"output_price,omitempty"`
	Description      string  `yaml:"description,omitempty"`
}

// ConfigManager handles configuration file operations
type ConfigManager struct {
	configPath string
	encryptor  *ConfigEncryptor
	config     *CLIConfig
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	encryptor, err := NewConfigEncryptor()
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	return &ConfigManager{
		configPath: configPath,
		encryptor:  encryptor,
	}, nil
}

// GetConfigPath returns the configuration file path
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use ~/.cline to match where cline-core looks for configuration
	configDir := filepath.Join(homeDir, ".cline")
	configFile := filepath.Join(configDir, "config.yaml")

	return configFile, nil
}

// EnsureConfigDirectory creates the config directory if it doesn't exist
func EnsureConfigDirectory() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}

// Load loads configuration from file
func (cm *ConfigManager) Load() (*CLIConfig, error) {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Create and store default config if file doesn't exist
		cm.config = cm.createDefaultConfig()
		return cm.config, nil
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config CLIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Decrypt API keys
	for id, provider := range config.Providers {
		if provider.APIKey != "" {
			decryptedKey, err := cm.encryptor.DecryptAPIKey(provider.APIKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt API key for provider %s: %w", id, err)
			}
			provider.APIKey = decryptedKey
			config.Providers[id] = provider
		}
	}

	cm.config = &config
	return &config, nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save(config *CLIConfig) error {
	// Ensure config directory exists
	if err := EnsureConfigDirectory(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a copy for encryption
	configCopy := *config
	configCopy.Providers = make(map[string]ProviderConfig)

	// Encrypt API keys
	for id, provider := range config.Providers {
		providerCopy := provider
		if provider.APIKey != "" {
			encryptedKey, err := cm.encryptor.EncryptAPIKey(provider.APIKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt API key for provider %s: %w", id, err)
			}
			providerCopy.APIKey = encryptedKey
		}
		configCopy.Providers[id] = providerCopy
	}

	// Update timestamps
	configCopy.UpdatedAt = time.Now()
	if configCopy.CreatedAt.IsZero() {
		configCopy.CreatedAt = configCopy.UpdatedAt
	}

	// Set encryption note
	configCopy.EncryptionNote = "API keys in this file are encrypted for security"

	// Marshal to YAML
	data, err := yaml.Marshal(&configCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.config = config
	return nil
}

// Validate validates the configuration
func (cm *ConfigManager) Validate(config *CLIConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.DefaultProvider != "" {
		if _, exists := config.Providers[config.DefaultProvider]; !exists {
			return fmt.Errorf("default provider %s not found in providers", config.DefaultProvider)
		}
	}

	// Validate each provider
	for id, provider := range config.Providers {
		if provider.ID != id {
			return fmt.Errorf("provider ID mismatch: %s != %s", provider.ID, id)
		}

		if provider.Name == "" {
			return fmt.Errorf("provider %s has empty name", id)
		}

		if provider.APIKey == "" {
			return fmt.Errorf("provider %s has empty API key", id)
		}

		if provider.ModelID == "" {
			return fmt.Errorf("provider %s has empty model ID", id)
		}
	}

	return nil
}

// BackupConfig creates a backup of the existing configuration
func (cm *ConfigManager) BackupConfig() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// No config to backup
		return nil
	}

	backupPath := cm.configPath + ".backup." + time.Now().Format("20060102-150405")
	
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// RestoreConfig restores configuration from a backup file
func (cm *ConfigManager) RestoreConfig(backupPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to restore config file: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *CLIConfig {
	return cm.config
}

// SetDefaultProvider sets the default provider
func (cm *ConfigManager) SetDefaultProvider(providerID string) error {
	if cm.config == nil {
		return fmt.Errorf("no config loaded")
	}

	if _, exists := cm.config.Providers[providerID]; !exists {
		return fmt.Errorf("provider %s not found", providerID)
	}

	cm.config.DefaultProvider = providerID
	return nil
}

// AddProvider adds a new provider to the configuration
func (cm *ConfigManager) AddProvider(provider ProviderConfig) error {
	if cm.config == nil {
		cm.config = cm.createDefaultConfig()
	}

	if cm.config.Providers == nil {
		cm.config.Providers = make(map[string]ProviderConfig)
	}

	cm.config.Providers[provider.ID] = provider
	return nil
}

// RemoveProvider removes a provider from the configuration
func (cm *ConfigManager) RemoveProvider(providerID string) error {
	if cm.config == nil {
		return fmt.Errorf("no config loaded")
	}

	if _, exists := cm.config.Providers[providerID]; !exists {
		return fmt.Errorf("provider %s not found", providerID)
	}

	delete(cm.config.Providers, providerID)

	// Clear default provider if it was the removed one
	if cm.config.DefaultProvider == providerID {
		cm.config.DefaultProvider = ""
	}

	return nil
}

// createDefaultConfig creates a default configuration
func (cm *ConfigManager) createDefaultConfig() *CLIConfig {
	return &CLIConfig{
		Version:         "1.0.0",
		EncryptionNote:  "API keys in this file are encrypted for security",
		DefaultProvider: "",
		Providers:       make(map[string]ProviderConfig),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// ConfigExists checks if a configuration file exists
func ConfigExists() (bool, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
