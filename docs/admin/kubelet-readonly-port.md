# Disabling the Deprecated Kubelet Readonly Port

Google Kubernetes Engine (GKE) shows a warning regarding the Kubelet readonly port (10255):

> Note: The Kubelet readonly port (10255) is now deprecated. Please update your workloads to use the recommended alternatives.

This document explains how to address this warning.

## Understanding the Issue

The Kubelet readonly port (10255) was previously used for unauthenticated access to the Kubelet API. For security reasons, this port is being deprecated and will eventually be disabled by default.

## Checking for Usage

To check if your workloads are using the Kubelet readonly port:

```bash
# Check for pods accessing 10255 directly
kubectl get pods --all-namespaces -o json | jq '.items[] | select(.spec.containers[].env[]?.value | contains(":10255"))'

# Look for NetworkPolicies that might reference port 10255
kubectl get networkpolicies --all-namespaces -o json | jq '.items[] | select(.spec | tostring | contains("10255"))'

# Check custom metrics adapters or monitoring tools that might use the readonly port
kubectl get pods --all-namespaces -o json | jq '.items[] | select(.spec | tostring | contains("10255"))'
```

## Recommended Alternatives

Instead of using the Kubelet readonly port, use one of these alternatives:

1. **Kubelet HTTPS API (port 10250)** - Requires authentication:
   ```bash
   # Example of using the secure port with authentication
   kubectl proxy
   curl http://localhost:8001/api/v1/nodes/NODE_NAME/proxy/stats/summary
   ```

2. **Kubernetes Metrics API** - For resource metrics:
   ```bash
   # Install metrics-server if not already installed
   kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
   
   # Use the metrics API
   kubectl top nodes
   kubectl top pods
   ```

3. **Prometheus** - For comprehensive monitoring:
   ```bash
   # Install Prometheus using Helm
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm install prometheus prometheus-community/prometheus
   ```

## Disable the Readonly Port in GKE Autopilot

For GKE Autopilot clusters, Google manages the cluster configuration, including the Kubelet settings. The readonly port will be automatically disabled according to Google's deprecation timeline.

## Disable the Readonly Port in GKE Standard

For GKE Standard clusters, you can explicitly disable the readonly port:

```bash
# Create a new cluster with the readonly port disabled
gcloud container clusters create CLUSTER_NAME \
  --region REGION \
  --no-enable-legacy-authorization \
  --no-enable-insecure-kubelet-readonly-port

# Update an existing cluster
gcloud container clusters update CLUSTER_NAME \
  --region REGION \
  --no-enable-insecure-kubelet-readonly-port
```

## Updating Applications

If your applications are using the readonly port:

1. **Metrics Collection**: Switch to using the metrics API or Prometheus
2. **Health Checks**: Use the Kubernetes API or liveness/readiness probes
3. **Custom Monitoring**: Update to use authenticated endpoints via port 10250 or the Kubernetes API server

## Further Information

For more details, see the [official GKE documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/disable-kubelet-readonly-port) on this topic.