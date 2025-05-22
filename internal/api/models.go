package api

// Response is the standard success response
// @Description Standard API success response
type Response struct {
	// Response message
	Message string `json:"message" example:"Sandbox created successfully"`
	// User ID
	UserID string `json:"userId" example:"user123"`
}

// ErrorResponse is the standard error response
// @Description Standard API error response
type ErrorResponse struct {
	// Error message
	Error string `json:"error" example:"User ID is required"`
}