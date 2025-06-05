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
)

// Configuration holds all configurable parameters for the application
type Configuration struct {
	// SandboxTimeoutDuration is the duration after which a sandbox will be automatically deleted
	SandboxTimeoutDuration time.Duration
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

	return config
}