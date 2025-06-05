.PHONY: build clean test run swagger docker-build docker-push gcr-auth docker-all create-gcr-secrets

# Variables
APP_NAME := k8sgo
PROJECT_ID := driven-seer-460401-p9
DOCKER_IMAGE := gcr.io/$(PROJECT_ID)/irisk8s
DOCKER_TAG := latest
DOCKER_FULL_IMAGE := $(DOCKER_IMAGE):$(DOCKER_TAG)
PLATFORM := linux/amd64
BIN_DIR := ./bin


# Build the Go binary
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	@$(BIN_DIR)/$(APP_NAME)



# Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g internal/api/handlers.go


# Build Docker image for linux/amd64
docker-build: 
	@echo "Building Docker image $(DOCKER_FULL_IMAGE) for $(PLATFORM)..."
	@docker build --platform $(PLATFORM) -t $(DOCKER_FULL_IMAGE) .

# Authenticate with Google Cloud for GCR access
gcr-auth:
	@echo "Authenticating with Google Cloud for GCR access..."
	@gcloud auth configure-docker gcr.io

# Push Docker image to Google Container Registry (GCR)
docker-push: gcr-auth docker-build
	@echo "Pushing Docker image $(DOCKER_FULL_IMAGE) to Google Container Registry..."
	@docker push $(DOCKER_FULL_IMAGE)

# All Docker tasks: generate swagger docs, build image, and push to GCR
docker-all: swagger docker-push
	@echo "Completed all Docker tasks: swagger docs, build, and push to GCR"

# Create GCR pull secrets in both default and user-sandboxes namespaces
# Note: Secrets using gcloud auth tokens expire after ~1 hour
create-gcr-secrets:
	@echo "Creating GCR pull secrets in default and user-sandboxes namespaces..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
	fi
	@echo "Creating secret in default namespace..."
	@./scripts/create-gcr-secret.sh default
	@echo "Creating secret in user-sandboxes namespace..."
	@./scripts/create-gcr-secret.sh user-sandboxes
	@echo "GCR pull secrets created successfully in both namespaces."
	@echo "NOTE: If using gcloud auth, these secrets will expire after ~1 hour."
	@echo "      Run this command again to refresh the tokens when needed."