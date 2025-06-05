#!/bin/bash
set -e

# Load environment variables from .env file
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
else
  echo "Error: .env file not found"
  echo "Please create a .env file based on .env.example"
  exit 1
fi

# Check if namespace is provided or use user-sandboxes as default
NAMESPACE=${1:-user-sandboxes}

echo "Creating GCR pull secret in namespace: $NAMESPACE"

# Create namespace if it doesn't exist
kubectl get namespace $NAMESPACE > /dev/null 2>&1 || kubectl create namespace $NAMESPACE

# Generate the Docker config JSON
if [ "$USE_GCLOUD_AUTH" = "true" ]; then
  # Using gcloud authentication
  echo "Using gcloud authentication"

  # Check if gcloud is authenticated
  if ! gcloud auth print-access-token &> /dev/null; then
    echo "Error: Not authenticated with gcloud"
    echo "Please run 'gcloud auth login' first"
    exit 1
  fi

  # Create Docker config JSON with gcloud token
  ACCESS_TOKEN=$(gcloud auth print-access-token)
  DOCKER_CONFIG="{\"auths\":{\"gcr.io\":{\"auth\":\"$(echo -n oauth2accesstoken:$ACCESS_TOKEN | base64)\"}}}"
  DOCKER_CONFIG_B64=$(echo -n $DOCKER_CONFIG | base64 | tr -d '\n')

elif [ -n "$GCR_JSON_KEY" ]; then
  # Using JSON key authentication
  echo "Using JSON key authentication"

  # Create Docker config JSON with JSON key
  DOCKER_CONFIG="{\"auths\":{\"gcr.io\":{\"auth\":\"$(echo -n _json_key:$GCR_JSON_KEY | base64)\"}}}"
  DOCKER_CONFIG_B64=$(echo -n $DOCKER_CONFIG | base64 | tr -d '\n')

else
  echo "Error: No authentication method specified"
  echo "Please set either USE_GCLOUD_AUTH=true or provide GCR_JSON_KEY in .env file"
  exit 1
fi

# Create the secret file from template
TEMPLATE_FILE="kubernetes/manifests/gcr-pullsecret.yaml.template"
SECRET_FILE="kubernetes/manifests/gcr-pullsecret.yaml"

if [ -f "$TEMPLATE_FILE" ]; then
  echo "Generating secret file from template"
  cat "$TEMPLATE_FILE" | sed "s|\${GCR_DOCKERCONFIGJSON}|$DOCKER_CONFIG_B64|g" > "$SECRET_FILE"

  # Apply the secret to the namespace
  kubectl apply -f "$SECRET_FILE" -n $NAMESPACE
else
  echo "Template file not found: $TEMPLATE_FILE"
  echo "Creating secret directly"

  # Create or update the secret directly
  if [ "$USE_GCLOUD_AUTH" = "true" ]; then
    kubectl create secret docker-registry gcr-pull-secret \
      --docker-server=gcr.io \
      --docker-username=oauth2accesstoken \
      --docker-password="$ACCESS_TOKEN" \
      --docker-email=gcr@example.com \
      --namespace=$NAMESPACE \
      --dry-run=client -o yaml | kubectl apply -f -
  else
    kubectl create secret docker-registry gcr-pull-secret \
      --docker-server=gcr.io \
      --docker-username=_json_key \
      --docker-password="$GCR_JSON_KEY" \
      --docker-email=gcr@example.com \
      --namespace=$NAMESPACE \
      --dry-run=client -o yaml | kubectl apply -f -
  fi
fi

# Add the secret to the default service account in the namespace
kubectl patch serviceaccount default -n $NAMESPACE -p '{"imagePullSecrets": [{"name": "gcr-pull-secret"}]}'

echo "Secret created and attached to default service account in namespace: $NAMESPACE"