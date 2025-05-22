# Prerequisites for GKE Deployment

Before deploying the K8s Orchestrator to Google Kubernetes Engine (GKE), you need to install and configure several tools.

## Installing Required Tools

### Install gcloud CLI

If you haven't installed the Google Cloud SDK:

```bash
# Download and install the Google Cloud SDK
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
gcloud init  # Follow the initialization steps
```

### Install kubectl and GKE Auth Plugin for macOS

For GKE on macOS, you need both kubectl and the GKE auth plugin:

#### Option 1: Install using our script (Recommended)

We provide a script that automatically detects your server version and installs the compatible kubectl and GKE auth plugin:

```bash
# Run the script directly from the repository
bash scripts/install_kubectl.sh

# After installation is complete, the script will verify the versions
# The binaries will be installed to the appropriate location on your system
```

This script will:
- Check your current server version (if possible)
- Download and install the matching kubectl version
- Install the GKE auth plugin
- Add everything to your PATH if necessary

#### Option 2: Manual installation with Homebrew

```bash
# Install kubectl
brew install kubectl

# Install Google Cloud SDK (includes the auth plugin)
brew install --cask google-cloud-sdk

# After installation, update components
gcloud components update

# Install the GKE auth plugin specifically
gcloud components install gke-gcloud-auth-plugin

# Verify installations
kubectl version --client
gke-gcloud-auth-plugin --version
```

#### Option 3: Manual installation with curl

```bash
# For kubectl (replace X.Y.Z with your server version)
curl -LO "https://dl.k8s.io/release/vX.Y.Z/bin/darwin/arm64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl

# For GKE auth plugin, you'll still need to install gcloud
# and then run: gcloud components install gke-gcloud-auth-plugin
```

> **Important**: For optimal compatibility, make sure your kubectl version is within one minor version of your cluster's version (e.g., v1.31.x client works with v1.32.x server).

### Install Helm

```bash
# For macOS (using Homebrew)
brew install helm

# For Linux
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## Next Step

Once you have installed all the prerequisites, proceed to [Setup Google Cloud and GKE Cluster](02-setup-gcp-gke.md).