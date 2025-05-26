.PHONY: build clean test run swagger docker-build docker-push

# Variables
APP_NAME := k8sgo
DOCKER_IMAGE := shanurcsenitap/irisk8s
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

# Push Docker image to Docker Hub
docker-push: docker-build
	@echo "Pushing Docker image $(DOCKER_FULL_IMAGE) to Docker Hub..."
	@docker push $(DOCKER_FULL_IMAGE)