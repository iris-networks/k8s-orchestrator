# Kubernetes Sandbox Platform

A Kubernetes-based service that provides isolated, browser-accessible desktop environments for users, powering https://agent.tryiris.dev.

## Why This Was Built

This platform was built to:
- Provide isolated sandbox environments for security testing and training
- Enable users to access desktop environments through browsers without local software installation
- Create disposable, isolated Linux environments that automatically clean up after use
- Support educational and training scenarios requiring isolated workspaces
- Facilitate safe web browsing in containerized environments

## Exposed Ports

- **API Server**: 8080 (container), mapped to 80 (service)
- **Per Sandbox VNC**: Port 6901 via unique subdomain (e.g., user123-vnc.tryiris.dev)
- **Per Sandbox HTTP**: Port 3000 via unique subdomain (e.g., user123-api.tryiris.dev)

## Features

- **Container Management**: Create/delete user sandboxes using `accetto/ubuntu-vnc-xfce-firefox-g3` image
- **Persistent Storage**: Attach user-specific persistent volumes that survive container restarts
- **Dynamic Subdomains**: Provision unique subdomains per user via Traefik
- **Auto-Cleanup**: Sandboxes are automatically removed after 15 minutes of inactivity
- **REST API**: Simple endpoints for container lifecycle management

## Prerequisites

- Kubernetes cluster with Traefik ingress controller
- Docker
- Go 1.21+
- `kubectl` configured with appropriate permissions

## Installation

### Clone the repository

```bash
git clone https://github.com/shanurcsenitap/irisk8s.git
cd irisk8s
```

### Install dependencies

```bash
make deps
```

### Build the application

```bash
make build
```

### Generate Swagger documentation

```bash
make swagger
```

## Usage

### Building and Deploying

```bash
# Build and push Docker image
make docker-all

# Deploy to Kubernetes
kubectl apply -f kubernetes/manifests/
```

### Running Locally

```bash
make run
```

The API will be available at http://localhost:8080
Swagger documentation will be available at http://localhost:8080/swagger/index.html

### Creating a Sandbox

```bash
# Create a sandbox for user "user123"
curl -X POST http://localhost:8080/v1/sandbox/user123

# The response will include a URL to access the sandbox via VNC web interface
```

### Managing Sandboxes

```bash
# List all sandboxes
curl http://localhost:8080/v1/sandboxes

# Get status of a specific sandbox
curl http://localhost:8080/v1/sandbox/user123/status

# Delete a sandbox
curl -X DELETE http://localhost:8080/v1/sandbox/user123
```

## API Endpoints

### Sandbox Management
- `POST /v1/sandbox/{userId}` - Create user sandbox
- `DELETE /v1/sandbox/{userId}` - Delete user sandbox
- `GET /v1/sandbox/{userId}/status` - Get sandbox status
- `GET /v1/sandboxes` - List all sandboxes

### Administration
- `POST /v1/admin/cleanup?minutes={minutes}&auth={authToken}` - Cleanup sandboxes older than specified minutes
  - `minutes`: Age threshold in minutes
  - `auth`: Authentication token (required)

## Deployment

For full deployment instructions to Google Kubernetes Engine, see the [GKE Deployment Guide](docs/deployment/gke-deployment-guide.md).

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.