#!/bin/bash
set -e

# Create Kind cluster
echo "Creating Kind cluster..."
kind create cluster --name k8s-service-cluster --config kind-config.yaml

echo "Installing Nginx Ingress Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Wait for ingress controller to be ready
echo "Waiting for Ingress controller to be ready..."
# Sleep to allow the resources to be created
sleep 10

# Wait for the ingress-nginx namespace to be active
echo "Waiting for ingress-nginx namespace to be active..."
kubectl wait --namespace ingress-nginx --for=condition=Available=True --timeout=90s deploy/ingress-nginx-controller || {
  echo "Ingress controller deployment is not available yet, checking pods status..."
  kubectl get pods -n ingress-nginx
  
  # Wait for the pods to be ready
  echo "Waiting for ingress-nginx pods to be ready..."
  kubectl wait --namespace ingress-nginx --for=condition=Ready --timeout=180s pod --selector=app.kubernetes.io/component=controller || {
    echo "Failed to wait for ingress-nginx pods. Continuing anyway."
  }
}

# Install cert-manager for SSL certificates
echo "Installing cert-manager..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
echo "Waiting for cert-manager to be ready..."
# Sleep to allow the namespace to be created
sleep 10

# Wait for the cert-manager namespace to be active
echo "Waiting for cert-manager to be available..."
kubectl wait --namespace cert-manager --for=condition=Available=True --timeout=120s deploy --all || {
  echo "Cert-manager deployments are not available yet, checking pods status..."
  kubectl get pods -n cert-manager
  
  # Wait for the pods to be ready
  echo "Waiting for cert-manager pods to be ready..."
  kubectl wait --namespace cert-manager --for=condition=Ready --timeout=180s pod --all || {
    echo "Failed to wait for cert-manager pods. Continuing anyway."
  }
}

# Create self-signed cluster issuer
echo "Creating self-signed cluster issuer..."
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
EOF

# Print information
echo "Setup complete! Kind cluster is ready with Nginx Ingress and cert-manager."
echo "You can access the Kubernetes API at: $(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')"

# Get the current KUBECONFIG file content
KUBECONFIG_CONTENT=$(kubectl config view --raw)

# Create a temp file with the KUBECONFIG content
KUBECONFIG_FILE="/tmp/kind-k8s-service-cluster-config"
echo "${KUBECONFIG_CONTENT}" > "${KUBECONFIG_FILE}"

echo "To use this cluster with the k8s-service, set your KUBECONFIG environment variable to:"
echo "${KUBECONFIG_FILE}"
echo ""
echo "You can do this by adding to your .env file:"
echo "KUBECONFIG=${KUBECONFIG_FILE}"