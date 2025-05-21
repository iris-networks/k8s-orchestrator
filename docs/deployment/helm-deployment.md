# Deploying with Helm

This guide provides instructions for deploying the K8s Orchestrator service using Helm, which is the recommended method for both local development and production deployments.

## Prerequisites

- Kubernetes cluster (local or cloud-based)
- Helm 3.2.0+ installed
- `kubectl` configured to access your cluster

## Quick Start

### Local Development (Docker Desktop/Minikube)

```bash
# Navigate to the helm chart directory
cd helm/k8s-orchestrator

# Install the chart with local values
helm install k8s-orchestrator . -f values-local.yaml
```

### Production Deployment on AWS EKS

```bash
# Navigate to the helm chart directory
cd helm/k8s-orchestrator

# Customize values-aws.yaml to match your environment
# - Update domain name
# - Set appropriate resource limits
# - Configure storage classes as needed

# The Docker image is now hosted on Docker Hub as shanurcsenitap/irisk8s:latest
# This is already configured in the values files

# Install the chart with AWS values
helm install k8s-orchestrator . -f values-aws.yaml
```

### Production Deployment on Google Cloud GKE

```bash
# Navigate to the helm chart directory
cd helm/k8s-orchestrator

# Customize values-gcp.yaml to match your environment
# - Update domain name
# - Set appropriate resource limits
# - Configure storage classes as needed

# The Docker image is now hosted on Docker Hub as shanurcsenitap/irisk8s:latest
# This is already configured in the values files

# Install the chart with GCP values
helm install k8s-orchestrator . -f values-gcp.yaml
```

## Configuring the Helm Chart

The Helm chart provides extensive configuration options through its values file. Here are the most important settings you might want to customize:

### General Settings

- `replicaCount`: Number of replicas of the orchestrator service
- `image.repository`: Docker image repository (default: shanurcsenitap/irisk8s)
- `image.tag`: Docker image tag (default: latest)
- `image.pullPolicy`: Image pull policy (default: Always)
- `env.DOMAIN`: Domain for user subdomains (e.g., "pods.tryiris.dev")

### Cloud Provider Settings

The chart supports cloud-specific configurations:

#### AWS EKS

```yaml
cloudProvider:
  aws:
    enabled: true
    ingressClass: "alb"
    annotations:
      # ALB-specific annotations
    storageClass: "gp2"
```

#### Google Cloud GKE

```yaml
cloudProvider:
  gcp:
    enabled: true
    ingressClass: "nginx"
    annotations:
      # NGINX ingress annotations
    storageClass: "standard-rwo"
```

### User Environment Settings

```yaml
userEnvironments:
  defaultImage: "accetto/ubuntu-vnc-xfce-firefox-g3"
  defaultPorts: [5901, 6901]
  defaultVolumeSize: "5Gi"
  storageClass: "gp2"  # Override for user PVCs
```

### Security and Networking

```yaml
networkPolicies:
  enabled: true  # Enable Kubernetes NetworkPolicies

rbac:
  create: true  # Create necessary RBAC resources
  rules:
    # Permissions for the orchestrator service
```

## Upgrading

To upgrade an existing deployment:

```bash
# Update your values file as needed
helm upgrade k8s-orchestrator . -f your-values.yaml
```

## Uninstalling

To remove the deployment:

```bash
helm uninstall k8s-orchestrator
```

Note that this will not delete the PersistentVolumeClaims or namespaces created for user environments. To fully clean up:

```bash
# Delete namespaces created for users
kubectl get namespaces -l app=k8sgo | grep user- | awk '{print $1}' | xargs kubectl delete namespace
```

## Troubleshooting

### Common Issues

1. **Ingress not working**:
   - Ensure your Ingress controller is properly installed
   - Check annotations are correct for your Ingress controller
   - Verify DNS is properly configured

2. **Permission errors**:
   - The service account needs specific permissions to create resources
   - Check the RBAC configuration in values.yaml

3. **PVC creation fails**:
   - Ensure the specified storage class exists in your cluster
   - Check if the storage class supports the requested volume size

### Debugging

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/instance=k8s-orchestrator

# View pod logs
kubectl logs -l app.kubernetes.io/instance=k8s-orchestrator

# Describe resources for more details
kubectl describe deployment k8s-orchestrator
```