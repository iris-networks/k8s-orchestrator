package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k8sgo/internal/k8s"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	
	_ "github.com/k8sgo/docs" // Import swagger docs
)

// Server represents the API server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	k8sClient  *k8s.Client
}

// NewServer creates a new API server
func NewServer() (*Server, error) {
	// Create Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, err
	}

	// Create gin router
	router := gin.Default()

	// Setup routes
	server := &Server{
		router:    router,
		k8sClient: k8sClient,
		httpServer: &http.Server{
			Addr:    ":8080",
			Handler: router,
		},
	}

	server.setupRoutes()
	return server, nil
}

// setupRoutes configures all API endpoints
func (s *Server) setupRoutes() {
	// API versioning
	v1 := s.router.Group("/api/v1")
	{
		// Environment endpoints
		env := v1.Group("/environments")
		{
			env.POST("", s.createEnvironment)
			env.GET("", s.listEnvironments)
			env.GET("/:username", s.getEnvironment)
			env.DELETE("/:username", s.deleteEnvironment)
			env.PUT("/:username", s.updateEnvironment)
		}
		
		// Health check endpoint
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Swagger documentation - make sure this is at the root level, not under /api/v1
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// Start launches the HTTP server
func (s *Server) Start() error {
	log.Println("Server started on :8080")
	log.Println("Swagger documentation available at http://localhost:8080/swagger/index.html")
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}