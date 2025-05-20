package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/k8sgo/internal/api"
)

// Run starts the server and returns an exit code
func Run() int {
	log.Println("Starting K8s orchestration service...")

	// Initialize the server
	server, err := api.NewServer()
	if err != nil {
		log.Printf("Failed to create server: %v", err)
		return 1
	}

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server failed: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	server.Shutdown()
	return 0
}