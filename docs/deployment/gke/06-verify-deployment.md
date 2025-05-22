# Verifying the Deployment

This guide walks you through verifying that your K8s Orchestrator deployment is working correctly.

## Check Deployment Status

```bash
# Check the pods
kubectl get pods -l app.kubernetes.io/instance=k8s-orchestrator

# Check the service
kubectl get service -l app.kubernetes.io/instance=k8s-orchestrator

# Check the ingress
kubectl get ingressroute -l app.kubernetes.io/instance=k8s-orchestrator
```

All pods should be in the `Running` state, and the service should have an external IP address. The IngressRoute should be configured correctly.

## Test the API

```bash
# Wait for the Ingress to get an IP and for DNS to propagate
kubectl get ingressroute -l app.kubernetes.io/instance=k8s-orchestrator

# Test the API health endpoint
curl -k https://api.pods.tryiris.dev/api/v1/health
```

You should get a response indicating that the API is healthy. If you're using a self-signed certificate or Let's Encrypt is still provisioning a certificate, you may need to use the `-k` flag to skip certificate validation.

## User Environment Provisioning Flow

```mermaid
sequenceDiagram
    participant Admin as Administrator
    participant API as API Service
    participant K8s as Kubernetes API
    participant DNS as DNS
    participant User as End User

    Admin->>API: POST /environments (username: "user1")
    API->>K8s: Create Namespace
    API->>K8s: Create PVC
    API->>K8s: Create Deployment (VNC + Web Server)
    API->>K8s: Create Service
    API->>K8s: Create Traefik IngressRoute for Port 3000
    API->>K8s: Create Traefik IngressRoute for Port 6901
    API->>DNS: Configure Subdomain
    API->>Admin: Return Success

    User->>DNS: Request user1.tryiris.dev:3000
    DNS->>K8s: Route to Web Server
    K8s->>User: Serve Web Application

    User->>DNS: Request user1.tryiris.dev:6901
    DNS->>K8s: Route to VNC Interface
    K8s->>User: Serve VNC Desktop
```

## Next Step

Once you have verified that your deployment is working correctly, proceed to [Create and Access User Environments](07-user-environments.md).