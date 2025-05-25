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

## 3. Deploying Traefik

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

## 4. Deploying Application Manifests
kubectl kustomize kubernetes/manifests/ | kubectl apply -f -


## 5. Verifying Deployment

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

## 6. Testing the Application

After deployment, verify that the API is accessible:

```bash
# Test API access
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

### Traefik Issues
```bash
# Check Traefik pods
kubectl get pods -n default | grep traefik

# View Traefik logs
kubectl logs -l app.kubernetes.io/name=traefik -n default
```

### DNS Issues
```bash
# Check if DNS records are properly configured in Cloudflare
# You'll need to manually check your Cloudflare dashboard
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