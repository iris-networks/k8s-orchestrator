.PHONY: build run clean swagger test docker-build docker-push docker-multiplatform

# Default target
all: build

# Build the application
build:
	@echo "Building k8sgo..."
	@go build -o bin/k8sgo main.go

# Run the application
run:
	@echo "Running k8sgo..."
	@go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ docs/docs.go docs/swagger.json docs/swagger.yaml

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@./scripts/swagger.sh

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t shanurcsenitap/irisk8s:latest .

# Push Docker image to Docker Hub
docker-push: docker-build
	@echo "Pushing Docker image to Docker Hub..."
	docker push shanurcsenitap/irisk8s:latest

# Build and push multi-platform Docker image (linux/amd64, linux/arm64)
docker-multiplatform:
	@echo "Building and pushing multi-platform Docker image..."
	@./scripts/build-push-multiplatform.sh

# Build and push multi-platform Docker image with custom tag
docker-multiplatform-tag:
	@echo "Usage: make docker-multiplatform-tag TAG=<tag>"
	@./scripts/build-push-multiplatform.sh --tag $${TAG:-latest}