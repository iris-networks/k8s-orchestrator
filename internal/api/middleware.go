package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shanurcsenitap/irisk8s/internal/config"
)

// AuthMiddleware creates a middleware for API key authentication
func AuthMiddleware(cfg *config.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "API key is required"})
			c.Abort()
			return
		}

		if apiKey != cfg.APIKey {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
