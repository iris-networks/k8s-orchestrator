package api

import (
	"github.com/gin-gonic/gin"
	"github.com/shanurcsenitap/irisk8s/internal/k8s"
)


// RegisterRoutes registers all API routes with the Kubernetes client
func RegisterRoutes(router *gin.Engine, k8sClient *k8s.ClientWithTraefik) {
	// Create handlers
	sandboxHandler := NewSandboxHandler(k8sClient)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Sandbox endpoints
		sandbox := v1.Group("/sandbox")
		{
			sandbox.POST("/:userId", sandboxHandler.CreateSandbox)
			sandbox.DELETE("/:userId", sandboxHandler.DeleteSandbox)
			sandbox.GET("/:userId/status", sandboxHandler.GetSandboxStatus)
		}

		// List sandboxes endpoint
		v1.GET("/sandboxes", sandboxHandler.ListSandboxes)

		// Admin endpoints
		admin := v1.Group("/admin")
		{
			admin.POST("/cleanup", sandboxHandler.TriggerCleanup)
		}
	}
}