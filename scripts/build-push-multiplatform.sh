#!/bin/bash
set -e

# Script to build and push multi-platform Docker images
# Supports: linux/amd64, linux/arm64

# Default values
IMAGE_NAME="shanurcsenitap/irisk8s"
IMAGE_TAG="latest"

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -t|--tag) IMAGE_TAG="$2"; shift ;;
        -i|--image) IMAGE_NAME="$2"; shift ;;
        -h|--help) 
            echo "Usage: $0 [-t|--tag TAG] [-i|--image IMAGE_NAME]"
            echo "Builds and pushes a multi-platform Docker image"
            echo ""
            echo "Options:"
            echo "  -t, --tag       Docker image tag (default: latest)"
            echo "  -i, --image     Docker image name (default: shanurcsenitap/irisk8s)"
            echo "  -h, --help      Show this help message"
            exit 0
            ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

echo "Building multi-platform Docker image: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "Platforms: linux/amd64, linux/arm64"

# Create a new builder instance if it doesn't exist
if ! docker buildx inspect multiplatform-builder &>/dev/null; then
    echo "Creating new buildx builder: multiplatform-builder"
    docker buildx create --name multiplatform-builder --driver docker-container --use
else
    echo "Using existing buildx builder: multiplatform-builder"
    docker buildx use multiplatform-builder
fi

# Build and push multi-platform images
echo "Building and pushing multi-platform image..."
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag "${IMAGE_NAME}:${IMAGE_TAG}" \
    --push \
    .

echo "Successfully built and pushed ${IMAGE_NAME}:${IMAGE_TAG} for multiple platforms"

# List the pushed image with supported platforms
echo "Verifying pushed images..."
docker buildx imagetools inspect "${IMAGE_NAME}:${IMAGE_TAG}"

echo "Done!"