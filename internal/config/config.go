package config

import "time"

// Config holds all configuration values for the application
var Config = struct {
	// Cleanup settings
	Cleanup struct {
		// ExpirationTime is the duration after which a sandbox will be automatically deleted
		ExpirationTime time.Duration
		// AuthToken is the auth token for external cleanup trigger
		AuthToken string
	}
}{
	Cleanup: struct {
		ExpirationTime time.Duration
		AuthToken      string
	}{
		// Default values
		ExpirationTime: 30 * time.Minute,
		AuthToken:      "k8s-auto-cleanup-token",
	},
}