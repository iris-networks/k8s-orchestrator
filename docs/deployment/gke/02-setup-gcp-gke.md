# Setting Up Google Cloud and GKE Cluster

This guide walks you through setting up your Google Cloud Platform account and creating a GKE Autopilot cluster.

## Install Google Cloud SDK (if not already installed)

If you haven't installed the Google Cloud SDK:

```bash
# Download and install the Google Cloud SDK
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
```

## Login and Set Project

```bash
# Login to your Google Cloud account
gcloud auth login

# List available projects
gcloud projects list

# Set your project ID (use the PROJECT_ID from the list, not the PROJECT_NUMBER)
gcloud config set project YOUR_PROJECT_ID

# Set the quota project for Application Default Credentials (important for billing)
gcloud auth application-default set-quota-project YOUR_PROJECT_ID

# Verify your configuration
gcloud config list
```

## Enable Required APIs

```bash
# Enable the Kubernetes Engine API
gcloud services enable container.googleapis.com
```

## Create a GKE Autopilot Cluster

Create a GKE Autopilot cluster in Singapore region:

```bash
# Create a GKE Autopilot cluster in Singapore region
gcloud container clusters create-auto k8s-orchestrator \
  --region asia-southeast1 \
  --release-channel=regular
```

Autopilot automatically manages the infrastructure, including:
- Node provisioning and scaling
- Security and node upgrades
- Resource optimization

## Set Resource Quotas

After cluster creation, set resource quotas to limit usage and costs:

```bash
# Create a cluster-wide resource quota to prevent unexpected billing
kubectl create -f - <<EOF
apiVersion: v1
kind: ResourceQuota
metadata:
  name: cluster-resource-quota
  namespace: default
spec:
  hard:
    requests.cpu: "8"      # Maximum 8 cores across the entire cluster
    requests.memory: 32Gi  # Maximum 32GB memory across the entire cluster
    limits.cpu: "16"       # Maximum burst capacity of 16 cores
    limits.memory: 64Gi    # Maximum burst capacity of 64GB memory
    pods: "50"             # Maximum 50 pods total
EOF
```

## Configure kubectl to Use the GKE Autopilot Cluster

After creating your cluster for the first time, you need to get the Kubernetes credentials:

```bash
# Get credentials for your GKE Autopilot cluster
# This configures kubectl to use your new cluster
gcloud container clusters get-credentials k8s-orchestrator --region asia-southeast1
```

This command:
1. Downloads the cluster's credentials
2. Creates or updates your kubeconfig file (~/.kube/config)
3. Sets your current context to the new cluster

Verify that kubectl is properly configured:

```bash
# Check if kubectl is properly configured
kubectl cluster-info

# List all nodes in your cluster
kubectl get nodes
```

If these commands succeed, you're ready to proceed. If you encounter any errors, make sure:
- You're logged in with gcloud (`gcloud auth login`)
- The cluster creation has completed
- You're using the correct project ID

> **Note**: For common administrative operations and commands to manage your cluster after deployment, please see the [Kubernetes Cluster Administration Guide](../admin/kubernetes-operations.md)

## Next Step

Once your GKE cluster is set up, proceed to [Setup Core Kubernetes Components](03-core-kubernetes-components.md).