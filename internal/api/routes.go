package api

import (
	"github.com/gin-gonic/gin"
	"github.com/shanurcsenitap/irisk8s/internal/k8s"
)

// RegisterRoutes registers all API routes with the base Kubernetes client
func RegisterRoutes(router *gin.Engine, k8sClient *k8s.Client) {
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
		}

		// List sandboxes endpoint
		v1.GET("/sandboxes", sandboxHandler.ListSandboxes)
	}
}

// RegisterRoutesWithTraefik registers all API routes with the Traefik-enabled Kubernetes client
func RegisterRoutesWithTraefik(router *gin.Engine, k8sClient *k8s.ClientWithTraefik) {
	// Create handlers
	sandboxHandler := NewSandboxHandlerWithTraefik(k8sClient)

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
		}

		// List sandboxes endpoint
		v1.GET("/sandboxes", sandboxHandler.ListSandboxes)
	}
}