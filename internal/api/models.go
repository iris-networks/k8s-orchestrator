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

// ErrorResponse is the standard error response
// @Description Standard API error response
type ErrorResponse struct {
	// Error message
	Error string `json:"error" example:"User ID is required"`
}