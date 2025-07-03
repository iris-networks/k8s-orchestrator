package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Default configuration values
const (
	// DefaultSandboxTimeoutMinutes is the default duration in minutes after which a sandbox will be automatically deleted
	DefaultSandboxTimeoutMinutes = 30
	// DefaultAPIKey is the default API key for securing endpoints
	DefaultAPIKey = "default-secret-key"
	// SecretMountPath is the directory where secrets are mounted
	SecretMountPath = "/etc/config"
)

// Configuration holds all configurable parameters for the application
type Configuration struct {
	// SandboxTimeoutDuration is the duration after which a sandbox will be automatically deleted
	SandboxTimeoutDuration time.Duration
	// APIKey is the secret key for authenticating requests
	APIKey string
}

// GetConfig returns the application configuration, populated from environment variables or defaults
func GetConfig() *Configuration {
	config := &Configuration{
		SandboxTimeoutDuration: time.Duration(DefaultSandboxTimeoutMinutes) * time.Minute,
	}

	// Override from environment if available
	if envTimeout := readSecret("SANDBOX_TIMEOUT_MINUTES"); envTimeout != "" {
		if minutes, err := strconv.Atoi(envTimeout); err == nil && minutes > 0 {
			config.SandboxTimeoutDuration = time.Duration(minutes) * time.Minute
		}
	}

	// Get API key from environment or use default
	if apiKey := readSecret("API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	} else {
		config.APIKey = DefaultAPIKey
	}

	return config
}

func readSecret(key string) string {
	// Check environment variable first
	if value := os.Getenv(key); value != "" {
		return value
	}

	// Fallback to reading from the secret file
	secretPath := filepath.Join(SecretMountPath, key)
	if content, err := os.ReadFile(secretPath); err == nil {
		return strings.TrimSpace(string(content))
	}

	return ""
}
