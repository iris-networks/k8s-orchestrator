#!/bin/bash
set -e

# Colors for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Checking for existing Kubernetes cluster...${NC}"

# Check if Docker Desktop is running
if ! docker info > /dev/null 2>&1; then
  echo -e "${RED}Error: Docker Desktop is not running. Please start Docker Desktop.${NC}"
  exit 1
fi

# Check for docker-desktop context and switch to it
CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "none")
if kubectl config get-contexts | grep -q "docker-desktop"; then
  if [ "$CURRENT_CONTEXT" != "docker-desktop" ]; then
    echo -e "${YELLOW}Switching to docker-desktop context...${NC}"
    kubectl config use-context docker-desktop
  fi
else
  echo -e "${RED}Docker Desktop context not found in kubectl config.${NC}"
  echo -e "${YELLOW}Available contexts:${NC}"
  kubectl config get-contexts
  exit 1
fi

# Check if Kubernetes is running in Docker Desktop
if ! kubectl cluster-info > /dev/null 2>&1; then
  echo -e "${YELLOW}Kubernetes is not running in Docker Desktop.${NC}"
  echo -e "${YELLOW}Please enable Kubernetes in Docker Desktop:${NC}"
  echo "1. Open Docker Desktop"
  echo "2. Go to Settings/Preferences"
  echo "3. Select Kubernetes from the left sidebar"
  echo "4. Check 'Enable Kubernetes'"
  echo "5. Click 'Apply & Restart'"
  echo "6. Wait for Kubernetes to start (this may take several minutes)"
  echo "7. Once Kubernetes is running, run this script again"
  exit 1
fi

echo -e "${GREEN}Docker Desktop Kubernetes is running!${NC}"

# Install Traefik instead of Nginx
echo -e "${GREEN}Installing Traefik Ingress Controller...${NC}"

# Create traefik namespace
kubectl create namespace traefik --dry-run=client -o yaml | kubectl apply -f -

# Create CRDs for Traefik
echo -e "${YELLOW}Creating Traefik CRDs...${NC}"
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.10/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml

# Create RBAC for Traefik
echo -e "${YELLOW}Creating Traefik RBAC resources...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-ingress-controller
  namespace: traefik
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: traefik-ingress-controller
rules:
  - apiGroups: ['']
    resources: ['services', 'endpoints', 'secrets']
    verbs: ['get', 'list', 'watch']
  - apiGroups: ['extensions', 'networking.k8s.io']
    resources: ['ingresses', 'ingressclasses']
    verbs: ['get', 'list', 'watch']
  - apiGroups: ['extensions', 'networking.k8s.io']
    resources: ['ingresses/status']
    verbs: ['update']
  - apiGroups: ['traefik.containo.us', 'traefik.io']
    resources: ['middlewares', 'middlewaretcps', 'ingressroutes', 'ingressroutetcps', 'ingressrouteudps', 'tlsoptions', 'tlsstores', 'serverstransports']
    verbs: ['get', 'list', 'watch']
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: traefik-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik-ingress-controller
subjects:
  - kind: ServiceAccount
    name: traefik-ingress-controller
    namespace: traefik
EOF

# Deploy Traefik with NodePort
echo -e "${YELLOW}Deploying Traefik...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: traefik
  namespace: traefik
  labels:
    app: traefik
spec:
  replicas: 1
  selector:
    matchLabels:
      app: traefik
  template:
    metadata:
      labels:
        app: traefik
    spec:
      serviceAccountName: traefik-ingress-controller
      containers:
        - name: traefik
          image: traefik:v2.10
          args:
            - --api.insecure=true
            - --accesslog
            - --entrypoints.web.address=:80
            - --entrypoints.websecure.address=:443
            - --providers.kubernetescrd
            - --providers.kubernetesingress
          ports:
            - name: web
              containerPort: 80
            - name: websecure
              containerPort: 443
            - name: admin
              containerPort: 8080
          volumeMounts:
            - name: data
              mountPath: /data
      volumes:
        - name: data
          emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: traefik
  namespace: traefik
spec:
  type: NodePort
  ports:
    - port: 80
      name: web
      targetPort: web
      nodePort: 30080
    - port: 443
      name: websecure
      targetPort: websecure
      nodePort: 30443
    - port: 8080
      name: admin
      targetPort: admin
      nodePort: 30808
  selector:
    app: traefik
EOF

# Wait for Traefik deployment to be ready
echo -e "${YELLOW}Waiting for Traefik to be ready...${NC}"
kubectl wait --namespace traefik --for=condition=Available=True --timeout=180s deploy/traefik || {
  echo -e "${RED}Traefik deployment is not available yet, checking pods status...${NC}"
  kubectl get pods -n traefik
  
  # Wait for the pods to be ready
  echo -e "${YELLOW}Waiting for Traefik pods to be ready...${NC}"
  kubectl wait --namespace traefik --for=condition=Ready --timeout=180s pod --selector=app=traefik || {
    echo -e "${RED}Failed to wait for Traefik pods. Continuing anyway.${NC}"
  }
}

# Install cert-manager for SSL certificates
echo -e "${GREEN}Installing cert-manager...${NC}"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
echo -e "${YELLOW}Waiting for cert-manager to be ready...${NC}"
sleep 10

# Wait for the cert-manager namespace to be active
echo -e "${YELLOW}Waiting for cert-manager to be available...${NC}"
kubectl wait --namespace cert-manager --for=condition=Available=True --timeout=180s deploy --all || {
  echo -e "${RED}Cert-manager deployments are not available yet, checking pods status...${NC}"
  kubectl get pods -n cert-manager
  
  # Wait for the pods to be ready
  echo -e "${YELLOW}Waiting for cert-manager pods to be ready...${NC}"
  kubectl wait --namespace cert-manager --for=condition=Ready --timeout=180s pod --all || {
    echo -e "${RED}Failed to wait for cert-manager pods. Continuing anyway.${NC}"
  }
}

# Create self-signed cluster issuer
echo -e "${GREEN}Creating self-signed cluster issuer...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
EOF

# Create a default IngressRoute for Traefik
echo -e "${GREEN}Creating default IngressRoute for Traefik dashboard...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: traefik-dashboard
  namespace: traefik
spec:
  entryPoints:
    - web
  routes:
    - match: Host(\`traefik.local.dev\`)
      kind: Rule
      services:
        - name: traefik
          port: 8080
EOF

# Add local.dev domains to /etc/hosts
echo -e "${GREEN}Adding local.dev domains to /etc/hosts...${NC}"
echo -e "${YELLOW}NOTE: This will require sudo access to modify /etc/hosts${NC}"
if ! grep -q "# Kubernetes local.dev domains" /etc/hosts; then
  echo "
# Kubernetes local.dev domains
127.0.0.1 local.dev
127.0.0.1 traefik.local.dev" | sudo tee -a /etc/hosts
  echo -e "${GREEN}Added local.dev entries to /etc/hosts${NC}"
else
  echo -e "${YELLOW}local.dev entries already exist in /etc/hosts${NC}"
fi

echo -e "${YELLOW}NOTE: You need to add entries for your specific environments to /etc/hosts manually:${NC}"
echo -e "${YELLOW}For example: 127.0.0.1 yourusername.local.dev${NC}"

# Print information
echo ""
echo -e "${GREEN}Setup complete! Docker Desktop Kubernetes is ready with Traefik and cert-manager.${NC}"
echo -e "${GREEN}You can access the Kubernetes API at: $(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')${NC}"
echo ""
echo -e "${GREEN}Traefik dashboard is available at: http://traefik.local.dev:30808${NC}"
echo -e "${GREEN}Your system is now configured for the k8s-service.${NC}"
echo -e "${GREEN}Access services at yourusername.local.dev:30080 (HTTP) or yourusername.local.dev:30443 (HTTPS)${NC}"

# Note about service ports
echo ""
echo -e "${YELLOW}IMPORTANT:${NC}"
echo -e "${YELLOW}1. NodePort services are exposed on ports 30080 (HTTP), 30443 (HTTPS)${NC}"
echo -e "${YELLOW}2. Update your application to use these ports when accessing services${NC}"
echo -e "${YELLOW}3. For production deployment, consider using LoadBalancer type and proper DNS${NC}"