# Traefik Installation Guide

This guide explains how to set up Traefik for the k8sgo application to enable routing to the management API and dynamic subdomain provisioning for user sandboxes.

## Prerequisites

- Kubernetes cluster with admin access
- kubectl configured to access your cluster
- Helm v3 installed
- Cloudflare account with API token that has Zone:Read and DNS:Edit permissions
- Domain configured in Cloudflare (tryiris.dev)

## Phase 1: Install Traefik and Set Up Basic Routing

### Step 1: Create the Cloudflare API Token Secret

```bash
kubectl create secret generic cloudflare-api-token \
  --from-literal=api-token=9lePC6a6l8aawBtz-df4KnMTkCYxdGJpijdvbgtQ
```

### Step 2: Install Traefik Helm Chart

```bash
# Add the Traefik Helm repository
helm repo add traefik https://traefik.github.io/charts
helm repo update

# Install Traefik with custom values
helm install traefik traefik/traefik --values kubernetes/traefik/values.yaml
```

<!-- External-DNS section removed as it's no longer used -->

### Step 3: Create the IngressRoute for k8sgo API

```bash
# Apply the IngressRoute for api.tryiris.dev
kubectl apply -f kubernetes/traefik/ingress-route.yaml
```

## Phase 2: Set Up Wildcard DNS for User Sandboxes

The k8sgo application will automatically create IngressRoute resources for each user sandbox using the dynamic client in `internal/k8s/client_update.go`. Each user will get two subdomains:

- `{userID}-vnc.tryiris.dev` -> Port 6901 (VNC access)
- `{userID}-api.tryiris.dev` -> Port 3000 (HTTP/REST access)

### Implementation Details:

1. **DNS Configuration**:
   - You will need to manually configure DNS records in Cloudflare
   - The records should point to the Traefik ingress controller

2. **TLS Certificates**:
   - Traefik will use Let's Encrypt with Cloudflare DNS challenge to obtain TLS certificates
   - Wildcard certificates may be used to cover all user subdomains

3. **User Sandbox Creation**:
   - When a user sandbox is created, the k8sgo application will create:
     - Kubernetes Deployment
     - Kubernetes Service
     - Two Traefik IngressRoute resources (VNC and API)
   - You will need to manually configure DNS records as needed

4. **User Sandbox Deletion**:
   - When a user sandbox is deleted, the k8sgo application will delete:
     - Kubernetes Deployment
     - Kubernetes Service
     - Two Traefik IngressRoute resources (VNC and API)
   - You will need to manually delete DNS records as needed

## Testing the Setup

### Test Phase 1: Verify API Access

After deploying the k8sgo application, verify that the API is accessible at:
`https://api.tryiris.dev/v1/sandbox/{userId}`

### Test Phase 2: Verify User Sandbox Access

1. Create a user sandbox:
   ```
   curl -X POST https://api.tryiris.dev/v1/sandbox/test-user
   ```

2. Verify that the subdomains are accessible:
   - VNC: `https://test-user-vnc.tryiris.dev`
   - API: `https://test-user-api.tryiris.dev`

3. Delete the user sandbox:
   ```
   curl -X DELETE https://api.tryiris.dev/v1/sandbox/test-user
   ```

4. Verify that the subdomains are no longer accessible.