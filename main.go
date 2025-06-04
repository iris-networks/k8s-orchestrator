package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/shanurcsenitap/irisk8s/docs"
	"github.com/shanurcsenitap/irisk8s/internal/api"
	"github.com/shanurcsenitap/irisk8s/internal/k8s"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)


func main() {
	// Initialize Kubernetes client with Traefik support
	k8sClient, err := k8s.NewClientWithTraefik()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Start the auto cleanup service to delete sandboxes after 15 minutes
	k8sClient.StartAutoCleanupService(context.Background())

	// Initialize router
	router := gin.Default()

	// Register routes
	api.RegisterRoutes(router, k8sClient)

	// Swagger documentation
	url := ginSwagger.URL("/swagger/doc.json") // The URL pointing to API definition
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}