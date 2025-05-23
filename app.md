# Kubernetes Sandbox Platform Requirements

## Overview
Build a Golang service that manages per-user containerized sandboxes using Kubernetes APIs.

## Core Functionality
- **Container Management**: Create/delete user sandboxes using `accetto/ubuntu-vnc-xfce-firefox-g3` image
- **Persistent Storage**: Attach user-specific persistent volumes that survive container restarts
- **Dynamic Subdomains**: Provision unique subdomains per user using Traefik
- **REST API**: Expose endpoints for container lifecycle management

## Technical Specifications

### Container Configuration
- **Image**: `accetto/ubuntu-vnc-xfce-firefox-g3`
- **Exposed Ports**: 
  - Port 6901 (VNC server)
  - Port 3000 (HTTP server)

### Storage Requirements
- One persistent volume per user
- Volumes must persist across container restarts
- Automatic reattachment when user returns

### Networking
- **Subdomains per user**: 2 subdomains required
  - `{user}-vnc.{domain}` → Port 6901 (VNC access)
  - `{user}-api.{domain}` → Port 3000 (HTTP/REST access)
- **Ingress**: Use Traefik for dynamic subdomain provisioning

### API Endpoints
- `POST /v1/sandbox/{userId}` - Create user sandbox
- `DELETE /v1/sandbox/{userId}` - Delete user sandbox

### Kubernetes Resources to Manage
- Deployments (container lifecycle)
- Services (port exposure)
- PersistentVolumeClaims (user storage)
- Ingress/IngressRoute (Traefik routing)

### Service Deployment
- **Containerization**: Package the Golang service as a Docker container
- **Kubernetes Deployment**: Deploy the service as a Kubernetes deployment with appropriate resource limits
- **Service Exposure**: Create Kubernetes Service (ClusterIP/LoadBalancer) to expose the management API
- **External Access**: Configure Traefik ingress for external access at `api.{domain}`
- **RBAC**: Configure ServiceAccount with necessary permissions to manage Kubernetes resources

### API Endpoints
- `POST /v1/sandbox/{userId}` - Create user sandbox
- `DELETE /v1/sandbox/{userId}` - Delete user sandbox
- **Access via**: `https://api.{domain}/v1/sandbox/{userId}`

## Expected Deliverables
1. **Dockerfile** for containerizing the Golang service
2. Golang application with Kubernetes client integration
3. REST API server with create/delete endpoints
4. Persistent volume management logic
5. Traefik ingress configuration for dynamic subdomains
6. **Complete Kubernetes manifests**:
   - Deployment (for the management service)
   - Service (to expose the API)
   - ServiceAccount + RBAC (for Kubernetes API permissions)
   - Ingress/IngressRoute (for external access)
7. Container lifecycle management (start, stop, restart with volume reattachment)