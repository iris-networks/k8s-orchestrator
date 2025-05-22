# Kubernetes Cluster Administration Guide

This guide provides common administrative operations for managing your Kubernetes cluster after deployment. These commands are useful for day-to-day cluster management, monitoring, and troubleshooting.

## Viewing Cluster Information

### Basic Cluster Information

```bash
# Get a comprehensive overview of your cluster
kubectl cluster-info

# Get detailed information about your cluster
gcloud container clusters describe k8s-orchestrator --region asia-southeast1
```

### Managing Contexts

```bash
# View your current kubectl context
kubectl config current-context

# List all available contexts
kubectl config get-contexts

# Switch to a different context if you have multiple clusters
kubectl config use-context CONTEXT_NAME
```

### Checking Node Status

```bash
# List all nodes in your cluster
kubectl get nodes

# Get detailed information about a specific node
kubectl describe node NODE_NAME

# Get information about your cluster's capacity and allocatable resources
kubectl get nodes -o=custom-columns=NAME:.metadata.name,CPU:.status.capacity.cpu,MEMORY:.status.capacity.memory,STATUS:.status.conditions[4].type

# Check node resource usage
kubectl top nodes
```

## Working with Namespaces

```bash
# List all namespaces
kubectl get namespaces

# Create a new namespace
kubectl create namespace NAMESPACE_NAME

# Set the default namespace for the current context
kubectl config set-context --current --namespace=NAMESPACE_NAME

# Delete a namespace and all its resources
kubectl delete namespace NAMESPACE_NAME
```

## Pod Management

```bash
# Get all pods in the current namespace
kubectl get pods

# Get all pods across all namespaces
kubectl get pods --all-namespaces

# Get pods with more details
kubectl get pods -o wide

# Describe a specific pod
kubectl describe pod POD_NAME

# View pod logs
kubectl logs POD_NAME

# View logs for a specific container in a pod
kubectl logs POD_NAME -c CONTAINER_NAME

# Stream logs from a pod
kubectl logs -f POD_NAME

# Execute a command in a pod
kubectl exec -it POD_NAME -- /bin/bash

# Check pod resource usage
kubectl top pods
```

## Deployment Management

```bash
# List all deployments
kubectl get deployments

# Scale a deployment
kubectl scale deployment DEPLOYMENT_NAME --replicas=N

# Edit a deployment
kubectl edit deployment DEPLOYMENT_NAME

# Restart a deployment (rolling restart)
kubectl rollout restart deployment DEPLOYMENT_NAME

# Check deployment status
kubectl rollout status deployment DEPLOYMENT_NAME

# View deployment history
kubectl rollout history deployment DEPLOYMENT_NAME

# Rollback to a previous deployment
kubectl rollout undo deployment DEPLOYMENT_NAME
kubectl rollout undo deployment DEPLOYMENT_NAME --to-revision=N
```

## Service Management

```bash
# List all services
kubectl get services

# Get detailed information about a service
kubectl describe service SERVICE_NAME

# Access a service (port forwarding)
kubectl port-forward service/SERVICE_NAME LOCAL_PORT:SERVICE_PORT
```

## Resource Management

```bash
# Get all resources in the current namespace
kubectl get all

# Get specific resource types
kubectl get deployments,services,pods

# Get resource usage
kubectl top pods
kubectl top nodes

# Get resource quotas
kubectl get resourcequota --all-namespaces

# Get limit ranges
kubectl get limitrange --all-namespaces
```

## Ingress Management

```bash
# List all ingresses
kubectl get ingress --all-namespaces

# Describe a specific ingress
kubectl describe ingress INGRESS_NAME -n NAMESPACE

# Get ingress IP address
kubectl get ingress INGRESS_NAME -n NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

## Certificate Management

```bash
# List all certificates
kubectl get certificates --all-namespaces

# Check certificate details
kubectl describe certificate CERTIFICATE_NAME -n NAMESPACE

# List all certificate requests
kubectl get certificaterequests --all-namespaces

# Check certificate issuer status
kubectl get clusterissuers
```

## Persistent Volume Management

```bash
# List all persistent volume claims
kubectl get pvc --all-namespaces

# List all persistent volumes
kubectl get pv

# Describe a specific PVC
kubectl describe pvc PVC_NAME -n NAMESPACE
```

## Config Management

```bash
# List all config maps
kubectl get configmaps

# List all secrets
kubectl get secrets

# Create a config map from a file
kubectl create configmap CONFIG_NAME --from-file=PATH_TO_FILE

# Create a secret from a file
kubectl create secret generic SECRET_NAME --from-file=PATH_TO_FILE
```

## Troubleshooting

### Pod Issues

```bash
# Check pod status
kubectl get pods -n NAMESPACE

# Get detailed information about a pod
kubectl describe pod POD_NAME -n NAMESPACE

# Check pod logs
kubectl logs POD_NAME -n NAMESPACE

# Check previous container logs if a container has restarted
kubectl logs POD_NAME -n NAMESPACE --previous

# Check pod events
kubectl get events -n NAMESPACE --sort-by='.metadata.creationTimestamp'
```

### Node Issues

```bash
# Check node status
kubectl get nodes

# Get detailed information about a node
kubectl describe node NODE_NAME

# Check node resource usage
kubectl top nodes
```

### Networking Issues

```bash
# Test connectivity from a pod
kubectl exec -it POD_NAME -n NAMESPACE -- ping SERVICE_NAME.NAMESPACE.svc.cluster.local

# Check DNS resolution from a pod
kubectl exec -it POD_NAME -n NAMESPACE -- nslookup SERVICE_NAME.NAMESPACE.svc.cluster.local

# Test HTTP endpoints from a pod
kubectl exec -it POD_NAME -n NAMESPACE -- curl -v SERVICE_NAME.NAMESPACE.svc.cluster.local:PORT
```

## GKE-Specific Operations

### Autopilot Cluster Management

```bash
# View Autopilot settings
gcloud container clusters describe k8s-orchestrator --region asia-southeast1 | grep -A 10 autopilot

# List all GKE clusters
gcloud container clusters list

# Resize an Autopilot cluster (not directly supported, controlled via resource requests)
# Instead, update ResourceQuotas to control maximum resource consumption

# Convert Autopilot to Standard mode (if needed)
gcloud container clusters update k8s-orchestrator --region asia-southeast1 --no-enable-autopilot
```

### Cluster Authentication

```bash
# Get credentials for a specific cluster
gcloud container clusters get-credentials k8s-orchestrator --region asia-southeast1

# Update kubeconfig with credentials for all clusters in a project
gcloud container clusters list --format="value(name,zone)" | while read -r name zone; do
  gcloud container clusters get-credentials "$name" --zone "$zone"
done
```

### Cloud Logging and Monitoring

```bash
# View GKE cluster logs in Cloud Logging
gcloud logging read "resource.type=k8s_cluster" --limit=10

# View specific pod logs
gcloud logging read "resource.type=k8s_pod AND resource.labels.pod_name=POD_NAME" --limit=10

# View GKE audit logs
gcloud logging read "logName:projects/PROJECT_ID/logs/cloudaudit.googleapis.com%2Factivity AND resource.type=k8s_cluster" --limit=10
```

## User Environment Management

```bash
# List all user namespaces
kubectl get namespaces -l app=k8s-orchestrator

# Check resources in a user's namespace
kubectl get all -n user-USERNAME

# Check persistent volume claims for a user
kubectl get pvc -n user-USERNAME

# Check ingress for a user
kubectl get ingress -n user-USERNAME

# Delete a user's namespace (removes all resources)
kubectl delete namespace user-USERNAME
```

## Helm Chart Management

```bash
# List all Helm releases
helm list

# Get values for a release
helm get values k8s-orchestrator

# Upgrade a release with new values
helm upgrade k8s-orchestrator ./helm/k8s-orchestrator -f values-gcp-autopilot.yaml

# Rollback a release
helm rollback k8s-orchestrator REVISION_NUMBER

# Uninstall a release
helm uninstall k8s-orchestrator
```

Remember to replace placeholders like `POD_NAME`, `NAMESPACE`, `NODE_NAME`, etc. with your actual resource names.