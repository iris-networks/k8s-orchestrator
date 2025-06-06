#!/bin/bash
set -e

# Configuration
PROJECT_ID="driven-seer-460401-p9"  # Google Cloud Project ID
REGION="us-central1"  # Same region as your GKE cluster
REPOSITORY="k8sgo-repo"  # Repository name in Artifact Registry
IMAGE_NAME="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY}/irisk8s"  # Artifact Registry image path
VERSION="latest"  # Version tag

# Create Artifact Registry repository if it doesn't exist
if ! gcloud artifacts repositories describe ${REPOSITORY} --project=${PROJECT_ID} --location=${REGION} > /dev/null 2>&1; then
  echo "Creating Artifact Registry repository..."
  gcloud artifacts repositories create ${REPOSITORY} \
    --project=${PROJECT_ID} \
    --repository-format=docker \
    --location=${REGION} \
    --description="K8sGo Docker images"
fi

# Authenticate with Artifact Registry
echo "Authenticating with Google Artifact Registry..."
gcloud auth configure-docker ${REGION}-docker.pkg.dev

# Build and push the image for linux/amd64
echo "Building and pushing Docker image for linux/amd64..."
docker build --platform linux/amd64 -t ${IMAGE_NAME}:${VERSION} .
docker push ${IMAGE_NAME}:${VERSION}

echo "Docker image built and pushed successfully to Google Artifact Registry!"
echo "Image: ${IMAGE_NAME}:${VERSION}"