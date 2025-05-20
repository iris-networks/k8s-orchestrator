.PHONY: build run clean swagger test

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