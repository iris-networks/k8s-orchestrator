.PHONY: build clean test run swagger docker-build docker-push artifact-auth docker-all create-artifact-repo deploy create-secret

# Variables
APP_NAME := k8sgo
PROJECT_ID := driven-seer-460401-p9
REGION := us-central1
REPOSITORY := k8sgo-repo
DOCKER_IMAGE := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/irisk8s
DOCKER_TAG := $(shell git rev-parse --short HEAD)
DOCKER_FULL_IMAGE := $(DOCKER_IMAGE):$(DOCKER_TAG)
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


# Build Docker image (ensuring linux/amd64 architecture)
docker-build: 
	@echo "Building Docker image $(DOCKER_FULL_IMAGE) for linux/amd64..."
	@docker build --platform linux/amd64 -t $(DOCKER_FULL_IMAGE) .

# Create Artifact Registry repository if it doesn't exist
create-artifact-repo:
	@echo "Checking if Artifact Registry repository exists..."
	@if ! gcloud artifacts repositories describe $(REPOSITORY) --project=$(PROJECT_ID) --location=$(REGION) > /dev/null 2>&1; then \
		echo "Creating Artifact Registry repository..."; \
		gcloud artifacts repositories create $(REPOSITORY) --project=$(PROJECT_ID) --repository-format=docker --location=$(REGION) --description="K8sGo Docker images"; \
	else \
		echo "Artifact Registry repository already exists"; \
	fi

# Authenticate with Artifact Registry
artifact-auth: create-artifact-repo
	@echo "Authenticating with Google Artifact Registry..."
	@gcloud auth configure-docker $(REGION)-docker.pkg.dev

# Push Docker image to Google Artifact Registry
docker-push: artifact-auth docker-build
	@echo "Pushing Docker image $(DOCKER_FULL_IMAGE) to Google Artifact Registry..."
	@docker push $(DOCKER_FULL_IMAGE)

# All Docker tasks: generate swagger docs, build image, and push to Artifact Registry
docker-all: swagger docker-push
	@echo "Completed all Docker tasks: swagger docs, build, and push to Artifact Registry"

# Build, push and deploy with same tag
docker-deploy: docker-all deploy
	@echo "Completed build, push and deploy with tag: $(DOCKER_TAG)"


# Deploy application using kubectl with the current image tag
deploy:
	@echo "Deploying k8sgo with image tag: $(DOCKER_TAG)..."
	@sed -i '' 's|image: us-central1-docker.pkg.dev/driven-seer-460401-p9/k8sgo-repo/irisk8s:.*|image: $(DOCKER_IMAGE):$(DOCKER_TAG)|' kubernetes/manifests/deployment.yaml
	@kubectl apply -f kubernetes/manifests/deployment.yaml -n default
	@echo "Deployment complete with tag: $(DOCKER_TAG)"

# Create secret from .env file or provided values
create-secret:
	@if [ -f .env ]; then \
		echo "Loading secrets from .env file..."; \
		export $$(grep -v '^#' .env | xargs); \
		if [ -z "$$API_KEY" ] || [ -z "$$SANDBOX_TIMEOUT_MINUTES" ]; then \
			echo "Error: API_KEY and SANDBOX_TIMEOUT_MINUTES must be set in .env file"; \
			exit 1; \
		fi; \
		echo "Creating/updating secret from .env values..."; \
		kubectl delete secret k8sgo-secrets --ignore-not-found=true; \
		kubectl create secret generic k8sgo-secrets \
			--from-literal=API_KEY="$$API_KEY" \
			--from-literal=SANDBOX_TIMEOUT_MINUTES="$$SANDBOX_TIMEOUT_MINUTES"; \
	elif [ -n "$(API_KEY)" ] && [ -n "$(SANDBOX_TIMEOUT_MINUTES)" ]; then \
		echo "Creating/updating secret with provided values..."; \
		kubectl delete secret k8sgo-secrets --ignore-not-found=true; \
		kubectl create secret generic k8sgo-secrets \
			--from-literal=API_KEY=$(API_KEY) \
			--from-literal=SANDBOX_TIMEOUT_MINUTES=$(SANDBOX_TIMEOUT_MINUTES); \
	else \
		echo "Error: Either create a .env file or provide API_KEY and SANDBOX_TIMEOUT_MINUTES"; \
		echo "Usage: make create-secret API_KEY=your-key SANDBOX_TIMEOUT_MINUTES=30"; \
		echo "Or create a .env file with these values"; \
		exit 1; \
	fi

