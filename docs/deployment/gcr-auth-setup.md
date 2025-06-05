# Setting Up GCR Authentication for Kubernetes

This guide explains how to set up authentication for pulling images from Google Container Registry (GCR) in your Kubernetes cluster.

## Prerequisites
- Google Cloud SDK (gcloud) installed
- kubectl installed and configured to access your cluster
- Access to the GCP project containing your GCR images

## Secure Secret Management with Environment Variables

We use environment variables to securely manage credentials for GCR access. This keeps sensitive data out of your Git repository.

### 1. Set Up Environment File

1. Create a `.env` file based on the provided `.env.example`:
   ```bash
   cp .env.example .env
   ```

2. Edit the `.env` file with your credentials:
   ```
   # GCR Authentication Credentials
   GCR_PROJECT_ID=driven-seer-460401-p9
   GCR_JSON_KEY=
   # If using gcloud auth, leave GCR_JSON_KEY empty and set:
   USE_GCLOUD_AUTH=true
   ```

   **Note:** The `.env` file is automatically excluded from Git via `.gitignore`

### 2. Authentication Options

You can authenticate with GCR in two ways:

#### Option A: Using gcloud Authentication (Recommended for Development)

1. Log in with gcloud:
   ```bash
   gcloud auth login
   ```

2. In your `.env` file, set:
   ```
   USE_GCLOUD_AUTH=true
   ```

#### Option B: Using Service Account JSON Key (Recommended for Production)

1. Create a service account with permissions to pull from GCR:
   ```bash
   # Replace PROJECT_ID with your actual GCP project ID
   PROJECT_ID=driven-seer-460401-p9

   # Create service account
   gcloud iam service-accounts create gcr-puller --display-name "GCR Pull Access"

   # Grant the service account permission to pull from GCR
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member serviceAccount:gcr-puller@$PROJECT_ID.iam.gserviceaccount.com \
     --role roles/storage.objectViewer
   ```

2. Create and download a JSON key:
   ```bash
   # Create and download the key
   gcloud iam service-accounts keys create gcr-key.json \
     --iam-account gcr-puller@$PROJECT_ID.iam.gserviceaccount.com
   ```

3. Copy the contents of the JSON key file to the `GCR_JSON_KEY` environment variable in your `.env` file.

### 3. Create Kubernetes Secret

Run our script to create the secret in your Kubernetes cluster:

```bash
# Create secret in default namespace
./scripts/create-gcr-secret.sh

# Or specify a namespace
./scripts/create-gcr-secret.sh user-sandboxes
```

This script:
1. Loads credentials from your `.env` file
2. Creates the Kubernetes secret
3. Patches the default ServiceAccount to use the secret

## Troubleshooting

If you encounter "ImagePullBackOff" errors:

1. Check the pod events:
   ```bash
   kubectl describe pod [pod-name]
   ```

2. Verify the secret is correctly referenced in your ServiceAccount:
   ```bash
   kubectl describe serviceaccount default -n [namespace]
   ```

3. Check if your authentication credentials are valid:
   ```bash
   # For gcloud auth
   gcloud auth print-access-token

   # For JSON key
   echo $GCR_JSON_KEY | jq
   ```

4. For temporary testing, you can try manually authenticating Docker with GCR:
   ```bash
   # Using gcloud
   gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin https://gcr.io

   # Using JSON key
   echo $GCR_JSON_KEY | docker login -u _json_key --password-stdin https://gcr.io
   ```