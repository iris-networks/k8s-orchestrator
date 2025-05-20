package utils

import (
	"os"
	"testing"
)

func TestGetDomainFromEnv(t *testing.T) {
	// Test default value
	os.Unsetenv("DOMAIN")
	domain := GetDomainFromEnv()
	if domain != "local.dev" {
		t.Errorf("Expected domain to be 'local.dev', got %s", domain)
	}

	// Test custom value
	os.Setenv("DOMAIN", "test.com")
	domain = GetDomainFromEnv()
	if domain != "test.com" {
		t.Errorf("Expected domain to be 'test.com', got %s", domain)
	}
}