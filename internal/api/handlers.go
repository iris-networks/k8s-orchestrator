package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shanurcsenitap/irisk8s/internal/k8s"
)

// SandboxHandler manages sandbox operations with the base Kubernetes client
type SandboxHandler struct {
	k8sClient *k8s.Client
}

// NewSandboxHandler creates a new sandbox handler with the base Kubernetes client
func NewSandboxHandler(k8sClient *k8s.Client) *SandboxHandler {
	return &SandboxHandler{
		k8sClient: k8sClient,
	}
}

// ListSandboxes lists all sandboxes
// @Summary      List all sandboxes
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

// SandboxHandlerWithTraefik manages sandbox operations with Traefik integration
type SandboxHandlerWithTraefik struct {
	k8sClient *k8s.ClientWithTraefik
}

// NewSandboxHandlerWithTraefik creates a new sandbox handler with Traefik integration
func NewSandboxHandlerWithTraefik(k8sClient *k8s.ClientWithTraefik) *SandboxHandlerWithTraefik {
	return &SandboxHandlerWithTraefik{
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
func (h *SandboxHandlerWithTraefik) ListSandboxes(c *gin.Context) {
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

// CreateSandbox creates a new sandbox for a user
// @Summary      Create a user sandbox
// @Description  Creates a new containerized sandbox for a specific user
// @Tags         sandbox
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID"
// @Param        request body SandboxRequest false "Environment variables to pass to the container"
// @Success      201 {object} Response
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

	c.JSON(http.StatusCreated, Response{
		Message: "Sandbox created successfully",
		UserID:  userID,
	})
}

// DeleteSandbox deletes a user's sandbox
// @Summary      Delete a user sandbox
// @Description  Deletes a containerized sandbox for a specific user
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
func (h *SandboxHandlerWithTraefik) CreateSandbox(c *gin.Context) {
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
func (h *SandboxHandlerWithTraefik) DeleteSandbox(c *gin.Context) {
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

// GetSandboxStatus gets the status of a sandbox by user ID
// @Summary      Get the status of a user sandbox
// @Description  Retrieves the status of a sandbox for a specific user
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

	c.JSON(http.StatusOK, SandboxStatusResponse{
		UserID:    sandbox.UserID,
		Status:    sandbox.Status,
		CreatedAt: sandbox.CreatedAt,
		Exists:    true,
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
func (h *SandboxHandlerWithTraefik) GetSandboxStatus(c *gin.Context) {
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