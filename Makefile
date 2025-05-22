.PHONY: all build clean test run swagger docker-build docker-push docker-buildx docker-buildx-push help

# Variables
APP_NAME := k8sgo
DOCKER_IMAGE := shanurcsenitap/irisk8s
DOCKER_TAG := latest
DOCKER_FULL_IMAGE := $(DOCKER_IMAGE):$(DOCKER_TAG)
PLATFORMS := linux/amd64,linux/arm64

# Go related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

# Build the Go binary
build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(GOBIN)/$(APP_NAME) .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN)/*
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	@$(GOBIN)/$(APP_NAME)

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g internal/api/handlers.go

# Build Docker image
docker-build: swagger
	@echo "Building Docker image $(DOCKER_FULL_IMAGE)..."
	@docker build -t $(DOCKER_FULL_IMAGE) .

# Push Docker image to Docker Hub
docker-push: docker-build
	@echo "Pushing Docker image $(DOCKER_FULL_IMAGE) to Docker Hub..."
	@docker push $(DOCKER_FULL_IMAGE)

# Build multi-architecture Docker image
docker-buildx:
	@echo "Building multi-architecture Docker image $(DOCKER_FULL_IMAGE)..."
	@docker buildx use multiplatform-builder 2>/dev/null || docker buildx create --name multiplatform-builder --use
	@docker buildx build --platform $(PLATFORMS) -t $(DOCKER_FULL_IMAGE) --load .

# Push multi-architecture Docker image to Docker Hub
docker-buildx-push: swagger
	@echo "Building and pushing multi-architecture Docker image $(DOCKER_FULL_IMAGE) to Docker Hub..."
	@docker buildx use multiplatform-builder 2>/dev/null || docker buildx create --name multiplatform-builder --use
	@docker buildx build --platform $(PLATFORMS) -t $(DOCKER_FULL_IMAGE) --push .

# Build and push Docker image
docker-all: docker-build docker-push

# Build and push multi-architecture Docker image
docker-all-multiarch: docker-buildx-push

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Setup development environment
setup: deps
	@echo "Setting up development environment..."
	@mkdir -p $(GOBIN)
	@go install github.com/swaggo/swag/cmd/swag@latest

# Help command
help:
	@echo "Available commands:"
	@echo "  make build              - Build the application"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make test               - Run tests"
	@echo "  make run                - Build and run the application"
	@echo "  make swagger            - Generate Swagger documentation"
	@echo "  make docker-build       - Build Docker image (single architecture)"
	@echo "  make docker-push        - Push Docker image to Docker Hub (single architecture)"
	@echo "  make docker-all         - Build and push Docker image (single architecture)"
	@echo "  make docker-buildx      - Build multi-architecture Docker image"
	@echo "  make docker-buildx-push - Build and push multi-architecture Docker image"
	@echo "  make docker-all-multiarch - Build and push multi-architecture Docker image"
	@echo "  make deps               - Install dependencies"
	@echo "  make setup              - Setup development environment"

# Default target
all: clean build