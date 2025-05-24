#!/bin/bash
set -e

# Configuration
IMAGE_NAME="shanurcsenitap/irisk8s"  # Docker Hub username/repo
VERSION="amd64"                  # Specific version tag
PLATFORMS="linux/amd64"          # Only build for amd64

# Ensure Docker buildx is available
if ! docker buildx version > /dev/null 2>&1; then
  echo "Error: Docker buildx is not available"
  exit 1
fi

# Use builder or create one if it doesn't exist
if ! docker buildx use amd64-builder > /dev/null 2>&1; then
  echo "Creating new buildx builder instance..."
  docker buildx create --name amd64-builder --use
fi

# Build and push the image
echo "Building and pushing Linux AMD64 image..."
docker buildx build \
  --platform ${PLATFORMS} \
  --tag ${IMAGE_NAME}:${VERSION} \
  --tag ${IMAGE_NAME}:latest \
  --push \
  .

echo "Linux AMD64 image built and pushed successfully!"
echo "Image: ${IMAGE_NAME}:${VERSION} and ${IMAGE_NAME}:latest"