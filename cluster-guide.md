# Complete Guide to Kubernetes Cluster Management for K8sGo

This guide provides detailed instructions for managing your Kubernetes cluster with the k8sgo application, including deleting an existing cluster, creating a new one, and deploying all necessary components.

## 1. Deleting the Existing Cluster

### Delete using GKE (Google Kubernetes Engine)

```bash
# Get the current cluster name
CLUSTER_NAME=$(kubectl config current-context | cut -d'_' -f4)
REGION=$(kubectl config current-context | cut -d'_' -f3)
PROJECT=$(kubectl config current-context | cut -d'_' -f2)

# Delete the cluster
gcloud container clusters delete $CLUSTER_NAME --region=$REGION --project=$PROJECT --quiet
```

## 2. Creating a New Kubernetes Cluster

### Create using GKE

```bash
# Set your variables
PROJECT_ID="driven-seer-460401-p9"
CLUSTER_NAME="k8sgo-cluster"
REGION="us-central1-a"  # Choose your preferred region
MACHINE_TYPE="e2-standard-2"  # Adjust based on your needs
NODE_COUNT=2  # Adjust based on your needs

# Create a GKE cluster
gcloud container clusters create $CLUSTER_NAME \
  --project=$PROJECT_ID \
  --region=$REGION \
  --machine-type=$MACHINE_TYPE \
  --enable-autoscaling --min-nodes=1 --max-nodes=5 \
  --num-nodes=$NODE_COUNT \
  --enable-network-policy

# Get credentials for kubectl
gcloud container clusters get-credentials $CLUSTER_NAME --region=$REGION --project=$PROJECT_ID
```

## 3. Setting Up Artifact Registry for Container Images

Instead of using Container Registry (gcr.io) with manual pull secrets, we'll use Artifact Registry which provides better regional isolation and simpler authentication.

```bash
# Create an Artifact Registry repository in the same region as your GKE cluster
REGION="us-central1"  # Must match region prefix of your GKE cluster
PROJECT_ID="driven-seer-460401-p9"
REPOSITORY="k8sgo-repo"

gcloud artifacts repositories create $REPOSITORY \
  --project=$PROJECT_ID \
  --repository-format=docker \
  --location=$REGION \
  --description="K8sGo Docker images"

# Authenticate Docker with Artifact Registry
gcloud auth configure-docker $REGION-docker.pkg.dev
```

### Building and Pushing Images for GKE Compatibility

When building Docker images for GKE (especially from Mac), ensure you build for the correct architecture:

```bash
# Build image specifically for linux/amd64 architecture
docker build --platform linux/amd64 -t $REGION-docker.pkg.dev/$PROJECT_ID/$REPOSITORY/irisk8s:latest .

# Push to Artifact Registry
docker push $REGION-docker.pkg.dev/$PROJECT_ID/$REPOSITORY/irisk8s:latest
```

You can also use the provided Makefile or build-multiarch.sh script which handle this for you:

```bash
# Using Makefile
make docker-push

# Or using the build script
./build-multiarch.sh
```

### Image Pull Authentication

GKE nodes automatically have access to pull images from:
- Artifact Registry repositories in the same project and same region
- Container Registry (gcr.io) in the same project

No additional pull secrets are required as long as your:
1. Image is in the same Google Cloud project as your GKE cluster
2. Artifact Registry repository is in the same region as your GKE cluster

## 4. Deploying Traefik

```bash
# Create a .env file with your Cloudflare token
cat <<EOF > .env
CLOUDFLARE_ACCESS_TOKEN=your_cloudflare_token_here
EOF

# Run the Traefik deployment script
bash kubernetes/deploy-traefik.sh
```

The `deploy-traefik.sh` script will:
1. Create Cloudflare API token secret
2. Add and update Traefik Helm repository
3. Install/upgrade Traefik with custom values
4. Apply IngressRoute for api.tryiris.dev

## 5. Deploying Application Manifests
```bash
kubectl kustomize kubernetes/manifests/ | kubectl apply -f -
```

## 6. DNS Configuration

Configure your DNS records in Cloudflare to point to the Traefik LoadBalancer IP:

```bash
# Get the external IP of your Traefik LoadBalancer
TRAEFIK_IP=$(kubectl get svc traefik -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo "Configure these DNS records in Cloudflare pointing to $TRAEFIK_IP:"
echo "api.tryiris.dev → $TRAEFIK_IP"
echo "dashboard.tryiris.dev → $TRAEFIK_IP"
echo "*.tryiris.dev → $TRAEFIK_IP (for wildcard sandbox domains)"
```

## 7. Verifying Deployment

```bash
# Check deployment status
kubectl get pods

# Check service status
kubectl get svc

# Check ingress status
kubectl get ingress

# Check Traefik IngressRoute
kubectl get ingressroute

# Check logs
kubectl logs -l app=k8sgo
```

## 8. Testing the Application

After deployment, verify that the API is accessible:

```bash
# Test API access (within cluster)
kubectl run curl --image=curlimages/curl -i --rm --restart=Never -- curl k8sgo/health

# Test external access (after DNS propagation)
curl https://api.tryiris.dev/v1/sandbox/health

# Create a test sandbox
curl -X POST https://api.tryiris.dev/v1/sandbox/test-user

# Verify the sandbox subdomains are accessible
# VNC: https://test-user-vnc.tryiris.dev
# API: https://test-user-api.tryiris.dev

# Delete the test sandbox
curl -X DELETE https://api.tryiris.dev/v1/sandbox/test-user
```

## Troubleshooting

### Image Pull Issues
```bash
# Check pod events for image pull errors
kubectl describe pod [pod-name]

# Check if the image architecture matches your GKE nodes
# GKE nodes typically run linux/amd64 architecture
# If using Mac (especially with M1/M2/M3), ensure you build with --platform=linux/amd64

# Verify image in Artifact Registry
gcloud artifacts docker images list $REGION-docker.pkg.dev/$PROJECT_ID/$REPOSITORY/irisk8s

# Ensure service account has access to Artifact Registry
# GKE node service accounts have automatic access to same-project, same-region repositories
```

### Traefik Issues
```bash
# Check Traefik pods
kubectl get pods -n default | grep traefik

# View Traefik logs
kubectl logs -l app.kubernetes.io/name=traefik -n default

# Check if IngressRoutes are correctly defined
kubectl get ingressroute
kubectl describe ingressroute k8sgo-api
```

### DNS Issues
```bash
# Check if DNS records are properly configured in Cloudflare
# Verify they point to your Traefik LoadBalancer IP
nslookup api.tryiris.dev

# If you're using Cloudflare proxying, you'll see Cloudflare IPs
# Test direct access to LoadBalancer IP:
curl -k --resolve api.tryiris.dev:443:$(kubectl get svc traefik -o=jsonpath='{.status.loadBalancer.ingress[0].ip}') https://api.tryiris.dev/health
```

### RBAC Issues
```bash
# Check if RBAC is properly configured
kubectl auth can-i create ingressroute --as=system:serviceaccount:default:k8sgo-sa
```

### Application Issues
```bash
# Check application logs
kubectl logs -l app=k8sgo
```