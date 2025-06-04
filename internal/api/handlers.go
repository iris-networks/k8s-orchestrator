package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shanurcsenitap/irisk8s/internal/k8s"
)


// SandboxHandler manages sandbox operations with Traefik integration
type SandboxHandler struct {
	k8sClient *k8s.ClientWithTraefik
}

// NewSandboxHandler creates a new sandbox handler with Traefik integration
func NewSandboxHandler(k8sClient *k8s.ClientWithTraefik) *SandboxHandler {
	return &SandboxHandler{
		k8sClient: k8sClient,
	}
}

// ListSandboxes lists all sandboxes with Traefik integration
// @Summary      List all sandboxes with Traefik routing
// @Description  Retrieves a list of all sandboxes with their status
// @Tags         sandbox
// @Accept       json
// @Produce      json
// @Success      200 {object} SandboxListResponse
// @Failure      500 {object} ErrorResponse
// @Router       /v1/sandboxes [get]
func (h *SandboxHandler) ListSandboxes(c *gin.Context) {
	ctx := c.Request.Context()

	sandboxes, err := h.k8sClient.ListSandboxes(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SandboxListResponse{
		Count:     len(sandboxes),
		Sandboxes: sandboxes,
	})
}


// CreateSandbox creates a new sandbox for a user with Traefik integration
// @Summary      Create a user sandbox with Traefik routing
// @Description  Creates a new containerized sandbox for a specific user with Traefik IngressRoutes
// @Tags         sandbox
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID"
// @Param        request body SandboxRequest false "Environment variables to pass to the container"
// @Success      201 {object} SandboxResponse
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /v1/sandbox/{userId} [post]
func (h *SandboxHandler) CreateSandbox(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	// Parse request body to get environment variables
	var request SandboxRequest
	if err := c.ShouldBindJSON(&request); err != nil && err.Error() != "EOF" {
		// Only return error if it's not an empty body
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Initialize empty maps if no env vars were provided
	if request.EnvVars == nil {
		request.EnvVars = make(map[string]string)
	}
	if request.NodeEnvVars == nil {
		request.NodeEnvVars = make(map[string]string)
	}

	err := h.k8sClient.CreateSandbox(userID, request.EnvVars, request.NodeEnvVars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Return response with VNC and API URLs
	vncURL := "https://" + userID + "-vnc.tryiris.dev"
	apiURL := "https://" + userID + "-api.tryiris.dev"

	c.JSON(http.StatusCreated, SandboxResponse{
		Response: Response{
			Message: "Sandbox created successfully",
			UserID:  userID,
		},
		VncURL: vncURL,
		ApiURL: apiURL,
	})
}

// DeleteSandbox deletes a user's sandbox with Traefik integration
// @Summary      Delete a user sandbox with Traefik routing
// @Description  Deletes a containerized sandbox for a specific user including Traefik IngressRoutes
// @Tags         sandbox
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID"
// @Success      200 {object} Response
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /v1/sandbox/{userId} [delete]
func (h *SandboxHandler) DeleteSandbox(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	err := h.k8sClient.DeleteSandbox(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Message: "Sandbox deleted successfully",
		UserID:  userID,
	})
}


// GetSandboxStatus gets the status of a sandbox by user ID with Traefik integration
// @Summary      Get the status of a user sandbox with Traefik routing
// @Description  Retrieves the status of a sandbox for a specific user with Traefik IngressRoutes
// @Tags         sandbox
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID"
// @Success      200 {object} SandboxStatusResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /v1/sandbox/{userId}/status [get]
func (h *SandboxHandler) GetSandboxStatus(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	ctx := c.Request.Context()
	sandbox, err := h.k8sClient.GetSandboxStatus(ctx, userID)
	if err != nil {
		// Check if the error is "not found"
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: fmt.Sprintf("No sandbox found for user ID: %s", userID),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// For Traefik integration, include the URLs
	vncURL := "https://" + userID + "-vnc.tryiris.dev"
	apiURL := "https://" + userID + "-api.tryiris.dev"

	c.JSON(http.StatusOK, SandboxStatusResponseWithURLs{
		SandboxStatusResponse: SandboxStatusResponse{
			UserID:    sandbox.UserID,
			Status:    sandbox.Status,
			CreatedAt: sandbox.CreatedAt,
			Exists:    true,
		},
		VncURL: vncURL,
		ApiURL: apiURL,
	})
}

// TriggerCleanup triggers the cleanup of sandboxes older than the specified duration with Traefik integration
// @Summary      Trigger cleanup of old sandboxes with Traefik routing
// @Description  Deletes all sandboxes that have been running for more than the specified minutes
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        minutes query int true "Age in minutes"
// @Param        auth query string true "Authentication token"
// @Success      200 {object} CleanupResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /v1/admin/cleanup [post]
func (h *SandboxHandler) TriggerCleanup(c *gin.Context) {
	// Get minutes from query parameter
	minutesStr := c.Query("minutes")
	if minutesStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Minutes parameter is required",
		})
		return
	}

	minutes, err := strconv.Atoi(minutesStr)
	if err != nil || minutes <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Minutes must be a positive integer",
		})
		return
	}

	// Get auth token from query parameter
	authToken := c.Query("auth")
	if authToken == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Auth token is required",
		})
		return
	}

	// Convert minutes to duration
	duration := time.Duration(minutes) * time.Minute

	// Trigger cleanup
	ctx := c.Request.Context()
	err = h.k8sClient.CleanupExpiredSandboxesByDuration(ctx, duration, authToken)
	if err != nil {
		// Check if the error is unauthorized
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Unauthorized: Invalid auth token",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CleanupResponse{
		Message:  "Cleanup triggered successfully",
		Duration: fmt.Sprintf("%d minutes", minutes),
	})
}