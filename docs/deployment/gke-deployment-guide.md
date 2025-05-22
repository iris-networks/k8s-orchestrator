# Deploying K8s Sandbox Platform to Google Kubernetes Engine (GKE)

This guide provides step-by-step instructions for deploying the K8s Sandbox Platform to Google Kubernetes Engine (GKE).

## Prerequisites

Before you begin, ensure you have the following:

- Google Cloud Platform account with billing enabled
- `gcloud` CLI installed and configured
- `kubectl` installed and configured
- Docker installed (for building and pushing images)
- Domain name with DNS configured (for Traefik ingress)

## Step 1: Set Up Environment Variables

Set up environment variables to make the deployment process easier:

```bash
# GCP Project and region
export PROJECT_ID=$(gcloud config get-value project)
export REGION=asia-southeast1
export ZONE=${REGION}-a

# Cluster configuration
export CLUSTER_NAME=k8sgo-cluster
export MACHINE_TYPE=e2-standard-2

# Application configuration
export DOCKER_IMAGE=shanurcsenitap/irisk8s:latest
export DOMAIN=tryiris.dev
```

## Step 2: Create a GKE Cluster

Create a GKE cluster with the necessary configurations:

```bash
gcloud container clusters create ${CLUSTER_NAME} \
  --project=${PROJECT_ID} \
  --zone=${ZONE} \
  --machine-type=${MACHINE_TYPE} \
  --num-nodes=3 \
  --network=default \
  --enable-ip-alias \
  --cluster-version=latest
```

After the cluster is created, configure `kubectl` to use the cluster:

```bash
gcloud container clusters get-credentials ${CLUSTER_NAME} --zone=${ZONE} --project=${PROJECT_ID}
```

## Step 3: Build and Push the Docker Image

Build and push the Docker image using the Makefile:

```bash
# Build and push the Docker image to Docker Hub
make docker-all
```

Alternatively, build and push manually:

```bash
# Build the Docker image
docker build -t ${DOCKER_IMAGE} .

# Push the Docker image to Docker Hub
docker push ${DOCKER_IMAGE}
```

## Step 4: Install Traefik Ingress Controller

Install Traefik using Helm:

```bash
# Add the Traefik Helm repository
helm repo add traefik https://traefik.github.io/charts
helm repo update

# Create a namespace for Traefik
kubectl create namespace traefik

# Install Traefik
helm install traefik traefik/traefik \
  --namespace=traefik \
  --set ingressClass.enabled=true \
  --set ingressClass.isDefaultClass=true \
  --set ports.websecure.tls.enabled=true
```

## Step 5: Create Kubernetes Namespaces

Create the necessary namespaces:

```bash
# Create namespace for the K8s Sandbox Platform
kubectl create namespace k8sgo-system

# Create namespace for user sandboxes
kubectl create namespace user-sandboxes
```

## Step 6: Configure RBAC for the K8s Sandbox Platform

Create a service account and role binding to allow the K8s Sandbox Platform to manage resources:

```bash
# Create a ClusterRole, ServiceAccount, and ClusterRoleBinding
kubectl apply -f kubernetes/deployment.yaml
```

## Step 7: Deploy the K8s Sandbox Platform

Deploy the K8s Sandbox Platform to the cluster:

```bash
# Update the image in the deployment file
sed -i '' "s|image: shanurcsenitap/irisk8s:latest|image: ${DOCKER_IMAGE}|g" kubernetes/deployment.yaml

# Deploy the application
kubectl apply -f kubernetes/deployment.yaml
```

## Step 8: Configure DNS for Traefik

Get the external IP address of the Traefik service:

```bash
TRAEFIK_IP=$(kubectl get service traefik -n traefik -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
echo "Traefik External IP: ${TRAEFIK_IP}"
```

Update your DNS records:

1. Create an A record for `api.${DOMAIN}` pointing to `${TRAEFIK_IP}`
2. Create a wildcard A record for `*.pods.${DOMAIN}` pointing to `${TRAEFIK_IP}`

## Step 9: Configure Certificates (Optional)

For production deployments, you should configure TLS certificates. You can use Let's Encrypt with cert-manager:

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.1/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=ready pod -l app=cert-manager -n cert-manager --timeout=60s

# Create a ClusterIssuer for Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@${DOMAIN}
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
EOF
```

Update the Ingress resource to use TLS:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k8sgo
  namespace: k8sgo-system
  annotations:
    kubernetes.io/ingress.class: traefik
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.${DOMAIN}
    secretName: k8sgo-tls
  rules:
  - host: api.${DOMAIN}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: k8sgo
            port:
              number: 80
EOF
```

## Step 10: Verify the Deployment

Check if the deployment was successful:

```bash
# Check if the deployment is running
kubectl get deployment -n k8sgo-system

# Check if the service is running
kubectl get service -n k8sgo-system

# Check if the ingress is configured
kubectl get ingress -n k8sgo-system

# Check the logs
kubectl logs -l app=k8sgo -n k8sgo-system
```

## Step 11: Test the API

Test the API endpoints:

```bash
# Create a user sandbox
curl -X POST https://api.${DOMAIN}/v1/sandbox/test-user

# Delete a user sandbox
curl -X DELETE https://api.${DOMAIN}/v1/sandbox/test-user
```

After creating a sandbox, you should be able to access it at:
- VNC interface: `https://test-user-vnc.pods.${DOMAIN}`
- API interface: `https://test-user-api.pods.${DOMAIN}`

## Troubleshooting

### Pod Creation Issues

If pods are not being created:

```bash
# Check the pod status
kubectl get pods -n user-sandboxes

# Check the logs for the K8s Sandbox Platform
kubectl logs -l app=k8sgo -n k8sgo-system
```

### Ingress Issues

If the ingress is not working:

```bash
# Check the Traefik logs
kubectl logs -l app.kubernetes.io/name=traefik -n traefik

# Check the ingress status
kubectl describe ingress -n k8sgo-system
```

### Storage Issues

If persistent volumes are not being created:

```bash
# Check the persistent volume claims
kubectl get pvc -n user-sandboxes

# Check the persistent volumes
kubectl get pv
```

## Scaling the Platform

To scale the platform horizontally:

```bash
# Scale the deployment
kubectl scale deployment k8sgo -n k8sgo-system --replicas=3
```

To increase node capacity:

```bash
# Add more nodes to the GKE cluster
gcloud container clusters resize ${CLUSTER_NAME} \
  --zone=${ZONE} \
  --num-nodes=5
```

## Cleanup

To clean up the resources:

```bash
# Delete the K8s Sandbox Platform
kubectl delete -f kubernetes/deployment.yaml

# Delete the namespaces
kubectl delete namespace k8sgo-system
kubectl delete namespace user-sandboxes

# Delete the GKE cluster
gcloud container clusters delete ${CLUSTER_NAME} --zone=${ZONE}
```

## Maintenance

### Updating the Application

To update the application:

1. Build and push a new Docker image:
   ```bash
   make docker-all
   ```

2. Update the deployment:
   ```bash
   kubectl set image deployment/k8sgo k8sgo=${DOCKER_IMAGE} -n k8sgo-system
   ```

### Monitoring

Set up monitoring using Google Cloud Monitoring and Logging:

1. Enable the Google Cloud Monitoring API:
   ```bash
   gcloud services enable monitoring.googleapis.com
   ```

2. View logs in the Google Cloud Console:
   ```bash
   gcloud logging read "resource.type=k8s_container AND resource.labels.namespace_name=k8sgo-system" --limit 10
   ```

## Security Considerations

1. **RBAC**: The application requires specific permissions to manage Kubernetes resources. Review the RBAC configuration to ensure it follows the principle of least privilege.

2. **Network Security**: Consider implementing network policies to restrict traffic between namespaces.

3. **Secrets Management**: Use Kubernetes Secrets or Google Secret Manager to store sensitive information.

4. **Container Security**: Regularly scan container images for vulnerabilities using tools like Container Registry Vulnerability Scanning.

5. **Pod Security**: Consider implementing Pod Security Policies to enforce security best practices.

## Conclusion

You have successfully deployed the K8s Sandbox Platform to Google Kubernetes Engine. The platform allows users to create and manage isolated sandboxes with persistent storage and unique subdomains.

For more information, refer to the following resources:

- [GKE Documentation](https://cloud.google.com/kubernetes-engine/docs)
- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)