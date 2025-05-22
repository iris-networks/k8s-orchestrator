#!/bin/bash
set -e

# Configuration
IMAGE_NAME="yourusername/k8sgo"  # Change to your Docker Hub username/repo
VERSION="1.0.0"                  # Change as needed
PLATFORMS="linux/amd64,linux/arm64"

# Ensure Docker buildx is available
if ! docker buildx version > /dev/null 2>&1; then
  echo "Error: Docker buildx is not available"
  exit 1
fi

# Use multiplatform-builder or create one if it doesn't exist
if ! docker buildx use multiplatform-builder > /dev/null 2>&1; then
  echo "Creating new buildx builder instance..."
  docker buildx create --name multiplatform-builder --use
fi

# Build and push the multi-architecture image
echo "Building and pushing multi-architecture image..."
docker buildx build \
  --platform ${PLATFORMS} \
  --tag ${IMAGE_NAME}:${VERSION} \
  --tag ${IMAGE_NAME}:latest \
  --push \
  .

echo "Multi-architecture image built and pushed successfully!"
echo "Image: ${IMAGE_NAME}:${VERSION} and ${IMAGE_NAME}:latest"