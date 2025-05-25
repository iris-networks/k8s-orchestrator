#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored messages
print_message() {
  echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
  print_error "kubectl is not installed. Please install it first."
  exit 1
fi

# Check if helm is installed
if ! command -v helm &> /dev/null; then
  print_error "helm is not installed. Please install it first."
  exit 1
fi

# Create the cloudflare-api-token secret
print_message "Creating Cloudflare API token secret..."
kubectl create secret generic cloudflare-api-token \
  --from-literal=api-token=$(grep CLOUDFLARE_ACCESS_TOKEN .env | cut -d '=' -f2) \
  --dry-run=client -o yaml | kubectl apply -f -

# Add the Traefik Helm repository
print_message "Adding Traefik Helm repository..."
helm repo add traefik https://traefik.github.io/charts
helm repo update

# Install or upgrade Traefik
print_message "Installing/upgrading Traefik with custom values..."
helm upgrade --install traefik traefik/traefik \
  --values kubernetes/traefik/values.yaml

# Wait for Traefik to be ready
print_message "Waiting for Traefik pods to be ready..."
kubectl rollout status deployment/traefik

# External-DNS deployment has been removed
print_message "External-DNS is no longer used, skipping..."

# Apply the IngressRoute for k8sgo API
print_message "Creating IngressRoute for api.tryiris.dev..."
kubectl apply -f kubernetes/traefik/ingress-route.yaml

print_message "Deployment completed successfully!"
print_message "Your Traefik dashboard should be accessible at: https://dashboard.tryiris.dev"
print_message "Your k8sgo API will be accessible at: https://api.tryiris.dev"
print_message "Wait a few minutes for DNS records to propagate."