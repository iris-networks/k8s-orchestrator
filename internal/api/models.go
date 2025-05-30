package api

import "github.com/shanurcsenitap/irisk8s/internal/k8s"

// Response is the standard success response
// @Description Standard API success response
type Response struct {
	// Response message
	Message string `json:"message" example:"Sandbox created successfully"`
	// User ID
	UserID string `json:"userId" example:"user123"`
}

// SandboxRequest represents a request to create a new sandbox.
// @Description Request to create a new sandbox with optional environment variables.
type SandboxRequest struct {
	// EnvVars contains key-value pairs of environment variables to pass to the container.
	// Example: {"DEBUG":"true", "API_KEY":"abc123"}
	EnvVars map[string]string `json:"envVars,omitempty" example:"{\"DEBUG\":\"true\", \"API_KEY\":\"abc123\"}"`

	// NodeEnvVars contains environment variables specifically for the Node.js process.
	// These will be passed only to the Node.js process running as nodeuser.
	// Example: {"NODE_ENV":"production", "PORT":"3000"}
	NodeEnvVars map[string]string `json:"nodeEnvVars,omitempty" example:"{\"NODE_ENV\":\"production\", \"PORT\":\"3000\"}"`
}

// SandboxResponse is the response for sandbox creation with Traefik integration
// @Description Sandbox creation response with URLs
type SandboxResponse struct {
	// Embed the standard response
	Response
	// VNC URL for the sandbox
	VncURL string `json:"vncUrl" example:"https://user123-vnc.tryiris.dev"`
	// API URL for the sandbox
	ApiURL string `json:"apiUrl" example:"https://user123-api.tryiris.dev"`
}

// SandboxListResponse is the response for listing all sandboxes
// @Description List of all sandboxes
type SandboxListResponse struct {
	// Count of sandboxes
	Count int `json:"count" example:"3"`
	// List of sandboxes
	Sandboxes []k8s.SandboxInfo `json:"sandboxes"`
}

// SandboxStatusResponse is the response for checking a sandbox's status
// @Description Response for sandbox status check
type SandboxStatusResponse struct {
	// User ID
	UserID string `json:"userId" example:"user123"`
	// Sandbox status
	Status string `json:"status" example:"Running"`
	// Created timestamp
	CreatedAt string `json:"createdAt" example:"2023-04-20T12:00:00Z"`
	// Whether the sandbox exists
	Exists bool `json:"exists" example:"true"`
}

// SandboxStatusResponseWithURLs is the response for checking a sandbox's status with Traefik integration
// @Description Response for sandbox status check with URLs
type SandboxStatusResponseWithURLs struct {
	// Embed the standard status response
	SandboxStatusResponse
	// VNC URL for the sandbox
	VncURL string `json:"vncUrl" example:"https://user123-vnc.tryiris.dev"`
	// API URL for the sandbox
	ApiURL string `json:"apiUrl" example:"https://user123-api.tryiris.dev"`
}

// ErrorResponse is the standard error response
// @Description Standard API error response
type ErrorResponse struct {
	// Error message
	Error string `json:"error" example:"User ID is required"`
}