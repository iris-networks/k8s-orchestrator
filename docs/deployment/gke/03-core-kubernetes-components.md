# Setting Up Core Kubernetes Component: Traefik

This guide walks you through setting up Traefik as the sole ingress controller and certificate manager for the K8s Orchestrator.

## Install Traefik Ingress Controller with Built-in Certificate Management

```bash
# Create the cluster role binding (required for Traefik)
kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin \
  --user $(gcloud config get-value account)

# Install Traefik CRDs first
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v2.10/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml

# Install Traefik Ingress Controller with Helm
helm repo add traefik https://traefik.github.io/charts
helm repo update

# Use the predefined Traefik values file from our repository
# The file is located at helm/k8s-orchestrator/traefik-values.yaml
# You can review the values with: cat /path/to/helm/k8s-orchestrator/traefik-values.yaml

# Install Traefik with the predefined values
helm install traefik traefik/traefik -f traefik-values.yaml

# Wait for the Traefik controller to be ready
kubectl wait --for=condition=ready pod \
  --selector=app.kubernetes.io/name=traefik \
  --timeout=120s
```

> **Note**: When installing Traefik on Autopilot, you may see warnings like:
> ```
> Warning: autopilot-default-resources-mutator:Autopilot updated Deployment...
> ```
> These warnings are normal and expected. Autopilot automatically assigns resource requirements to containers that don't specify them. See [Understanding Autopilot Resource Defaults](../admin/autopilot-resource-defaults.md) for details.

## Verify Traefik Installation

```bash
# Check that Traefik pods are running
kubectl get pods -l app.kubernetes.io/name=traefik

# Check Traefik service
kubectl get service traefik

# Get the external IP address for DNS configuration
TRAEFIK_IP=$(kubectl get service traefik -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo "Traefik external IP: $TRAEFIK_IP"
```

## Next Step

Once you have set up Traefik, proceed to [Configure DNS](04-configure-dns.md).