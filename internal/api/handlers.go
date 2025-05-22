package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shanurcsenitap/irisk8s/internal/k8s"
)

// SandboxHandler manages sandbox operations
type SandboxHandler struct {
	k8sClient *k8s.Client
}

// NewSandboxHandler creates a new sandbox handler
func NewSandboxHandler(k8sClient *k8s.Client) *SandboxHandler {
	return &SandboxHandler{
		k8sClient: k8sClient,
	}
}

// CreateSandbox creates a new sandbox for a user
// @Summary      Create a user sandbox
// @Description  Creates a new containerized sandbox for a specific user
// @Tags         sandbox
// @Accept       json
// @Produce      json
// @Param        userId path string true "User ID"
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

	err := h.k8sClient.CreateSandbox(userID)
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