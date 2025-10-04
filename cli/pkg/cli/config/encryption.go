package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ConfigEncryptor handles encryption and decryption of sensitive configuration data
type ConfigEncryptor struct {
	gcm cipher.AEAD
}

// NewConfigEncryptor creates a new configuration encryptor
func NewConfigEncryptor() (*ConfigEncryptor, error) {
	key, err := getOrCreateEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &ConfigEncryptor{gcm: gcm}, nil
}

// EncryptAPIKey encrypts an API key for storage
func (ce *ConfigEncryptor) EncryptAPIKey(apiKey string) (string, error) {
	if apiKey == "" {
		return "", nil
	}

	// Generate a random nonce
	nonce := make([]byte, ce.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the API key
	ciphertext := ce.gcm.Seal(nonce, nonce, []byte(apiKey), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAPIKey decrypts an API key from storage
func (ce *ConfigEncryptor) DecryptAPIKey(encryptedKey string) (string, error) {
	if encryptedKey == "" {
		return "", nil
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	// Extract nonce
	nonceSize := ce.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := ce.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return string(plaintext), nil
}

// getOrCreateEncryptionKey gets or creates the encryption key
func getOrCreateEncryptionKey() ([]byte, error) {
	keyPath, err := getEncryptionKeyPath()
	if err != nil {
		return nil, err
	}

	// Try to read existing key
	if _, err := os.Stat(keyPath); err == nil {
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read encryption key: %w", err)
		}

		key, err := base64.StdEncoding.DecodeString(string(keyData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode encryption key: %w", err)
		}

		if len(key) != 32 {
			return nil, fmt.Errorf("invalid key length: expected 32, got %d", len(key))
		}

		return key, nil
	}

	// Generate new key
	key := make([]byte, 32) // 256-bit key
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Ensure key directory exists
	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	// Save key to file
	encodedKey := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(keyPath, []byte(encodedKey), 0600); err != nil {
		return nil, fmt.Errorf("failed to save encryption key: %w", err)
	}

	return key, nil
}

// getEncryptionKeyPath returns the path to the encryption key file
func getEncryptionKeyPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	keyDir := filepath.Join(homeDir, "Documents", "Cline", "CLI", ".keys")
	keyFile := filepath.Join(keyDir, "encryption.key")

	return keyFile, nil
}

// GenerateKeyFingerprint generates a fingerprint for the encryption key
func GenerateKeyFingerprint() (string, error) {
	keyPath, err := getEncryptionKeyPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", fmt.Errorf("encryption key does not exist")
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read encryption key: %w", err)
	}

	// Generate SHA256 hash of the key
	hash := sha256.Sum256(keyData)
	fingerprint := fmt.Sprintf("%x", hash[:8]) // First 8 bytes as hex

	return fingerprint, nil
}

// RotateEncryptionKey rotates the encryption key and re-encrypts all data
func RotateEncryptionKey(configManager *ConfigManager) error {
	// Load current config with old key
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config with old key: %w", err)
	}

	// Backup current key
	keyPath, err := getEncryptionKeyPath()
	if err != nil {
		return err
	}

	backupKeyPath := keyPath + ".backup"
	if err := copyFile(keyPath, backupKeyPath); err != nil {
		return fmt.Errorf("failed to backup encryption key: %w", err)
	}

	// Remove old key to force generation of new one
	if err := os.Remove(keyPath); err != nil {
		return fmt.Errorf("failed to remove old key: %w", err)
	}

	// Create new encryptor (will generate new key)
	newEncryptor, err := NewConfigEncryptor()
	if err != nil {
		// Restore backup key on failure
		copyFile(backupKeyPath, keyPath)
		return fmt.Errorf("failed to create new encryptor: %w", err)
	}

	// Update config manager with new encryptor
	configManager.encryptor = newEncryptor

	// Save config with new key
	if err := configManager.Save(config); err != nil {
		// Restore backup key on failure
		copyFile(backupKeyPath, keyPath)
		return fmt.Errorf("failed to save config with new key: %w", err)
	}

	// Remove backup key on success
	os.Remove(backupKeyPath)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0600)
}

// ValidateEncryption validates that encryption/decryption is working correctly
func ValidateEncryption() error {
	encryptor, err := NewConfigEncryptor()
	if err != nil {
		return fmt.Errorf("failed to create encryptor: %w", err)
	}

	testData := "test-api-key-12345"

	// Encrypt
	encrypted, err := encryptor.EncryptAPIKey(testData)
	if err != nil {
		return fmt.Errorf("failed to encrypt test data: %w", err)
	}

	// Decrypt
	decrypted, err := encryptor.DecryptAPIKey(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt test data: %w", err)
	}

	// Verify
	if decrypted != testData {
		return fmt.Errorf("encryption validation failed: expected %s, got %s", testData, decrypted)
	}

	return nil
}

// IsEncrypted checks if a string appears to be encrypted (base64 encoded)
func IsEncrypted(value string) bool {
	if value == "" {
		return false
	}

	// Try to decode as base64
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return false
	}

	// Check if it has the minimum length for encrypted data (nonce + some data)
	return len(decoded) >= 16
}

// GetEncryptionInfo returns information about the encryption setup
func GetEncryptionInfo() (map[string]interface{}, error) {
	keyPath, err := getEncryptionKeyPath()
	if err != nil {
		return nil, err
	}

	info := make(map[string]interface{})

	// Check if key exists
	if stat, err := os.Stat(keyPath); err == nil {
		info["key_exists"] = true
		info["key_path"] = keyPath
		info["key_size"] = stat.Size()
		info["key_modified"] = stat.ModTime()

		// Generate fingerprint
		if fingerprint, err := GenerateKeyFingerprint(); err == nil {
			info["key_fingerprint"] = fingerprint
		}
	} else {
		info["key_exists"] = false
		info["key_path"] = keyPath
	}

	// Test encryption
	if err := ValidateEncryption(); err == nil {
		info["encryption_working"] = true
	} else {
		info["encryption_working"] = false
		info["encryption_error"] = err.Error()
	}

	return info, nil
}
