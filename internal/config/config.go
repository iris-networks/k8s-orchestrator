package config

import (
	"os"
	"strconv"
	"time"
)

// Default configuration values
const (
	// DefaultSandboxTimeoutMinutes is the default duration in minutes after which a sandbox will be automatically deleted
	DefaultSandboxTimeoutMinutes = 30
	// DefaultAPIKey is the default API key for securing endpoints
	DefaultAPIKey = "default-secret-key"
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
	if envTimeout := os.Getenv("SANDBOX_TIMEOUT_MINUTES"); envTimeout != "" {
		if minutes, err := strconv.Atoi(envTimeout); err == nil && minutes > 0 {
			config.SandboxTimeoutDuration = time.Duration(minutes) * time.Minute
		}
	}

	// Get API key from environment or use default
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	} else {
		config.APIKey = DefaultAPIKey
	}

	return config
}