.PHONY: docker-build docker-push

# Variables
APP_NAME := k8sgo
DOCKER_IMAGE := shanurcsenitap/irisk8s
DOCKER_TAG := latest
DOCKER_FULL_IMAGE := $(DOCKER_IMAGE):$(DOCKER_TAG)
PLATFORM := linux/amd64

# Build Docker image for linux/amd64
docker-build: 
	@echo "Building Docker image $(DOCKER_FULL_IMAGE) for $(PLATFORM)..."
	@docker build --platform $(PLATFORM) -t $(DOCKER_FULL_IMAGE) .

# Push Docker image to Docker Hub
docker-push: docker-build
	@echo "Pushing Docker image $(DOCKER_FULL_IMAGE) to Docker Hub..."
	@docker push $(DOCKER_FULL_IMAGE)