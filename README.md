# K8s Go - Kubernetes Desktop Orchestrator

K8s Go is a service that creates isolated Kubernetes environments for users with persistent storage and dynamic subdomains. Each user gets their own container with VNC access on a virtual desktop.

## Features

- Create isolated user environments in Kubernetes
- Provision persistent storage via PVCs
- Generate and configure subdomain access
- Support for destroying and recreating containers without data loss
- RESTful API with Swagger documentation

## Environment Components

Each user environment includes:
- A VNC container (using noVNC)
- Persistent storage for user data
- Network isolation from other users
- Auto-generated subdomain

## Prerequisites

- Go 1.21 or later
- Docker Desktop with Kubernetes enabled
- kubectl configured to access your cluster
- Proper DNS configuration for local testing (add entries to /etc/hosts)

## Installation

### Option 1: Use Docker Hub Image (Recommended)

The application is available as a pre-built multi-platform Docker image on Docker Hub (supports linux/amd64 and linux/arm64):

```bash
docker pull shanurcsenitap/irisk8s:latest
```

### Option 2: Clone the repository

```bash
git clone https://github.com/yourusername/k8sgo.git
cd k8sgo
```

### Build from source

```bash
make build
```

### Build and push Docker images

#### Build and push for a single platform:
```bash
make docker-build   # Build the Docker image
make docker-push    # Build and push the Docker image to Docker Hub
```

#### Build and push multi-platform images (linux/amd64, linux/arm64):
```bash
make docker-multiplatform            # Build and push with the 'latest' tag
make docker-multiplatform-tag TAG=v1.0.0  # Build and push with a custom tag
```

### Generate Swagger documentation

```bash
make swagger
```

## Running the Service

### Run directly

```bash
make run
```

### Run with Docker

```bash
docker-compose up -d
```

### Use pre-built Docker Hub image

The application is available as a pre-built Docker image on Docker Hub:

```bash
docker pull shanurcsenitap/irisk8s:latest
```

You can also use this image directly in your docker-compose.yml or Kubernetes deployments.

### Deploy with Helm (Recommended for Production)

For Kubernetes deployments, we recommend using Helm:

```bash
# Navigate to the helm chart directory
cd helm/k8s-orchestrator

# For local development
helm install k8s-orchestrator . -f values-local.yaml

# For AWS EKS
helm install k8s-orchestrator . -f values-aws.yaml

# For GCP GKE
helm install k8s-orchestrator . -f values-gcp.yaml
```

For detailed deployment instructions, see:
- [Helm Deployment](docs/deployment/helm-deployment.md)
- [AWS EKS Deployment](docs/deployment/aws-eks-deployment.md)
- [GCP GKE Deployment](docs/deployment/gcp-gke-deployment.md)

## Local Development Setup

1. Install dependencies
   ```bash
   go mod tidy
   ```

2. Generate Swagger documentation
   ```bash
   make swagger
   ```

3. Start the server
   ```bash
   make run
   ```

4. Access the Swagger documentation at http://localhost:8080/swagger/index.html

## API Usage Examples

### Create a new environment

```bash
curl -X POST http://localhost:8080/api/v1/environments \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "image": "accetto/ubuntu-vnc-xfce-firefox-g3",
    "ports": [5901, 6901],
    "volumeSize": "2Gi"
  }'
```

### Access the environment

After creating an environment, add an entry to your `/etc/hosts` file:

```
127.0.0.1 testuser.local.dev
```

Then access the VNC interface at: http://testuser.local.dev

## API Endpoints

The API provides the following endpoints:

- `POST /api/v1/environments` - Create a new environment
- `GET /api/v1/environments` - List all environments
- `GET /api/v1/environments/{username}` - Get a specific environment
- `DELETE /api/v1/environments/{username}` - Delete an environment
- `PUT /api/v1/environments/{username}` - Update an environment

## Project Structure

```
.
├── cmd/
│   └── server/            # Application entry point
├── internal/
│   ├── api/               # API handlers and server configuration
│   ├── k8s/               # Kubernetes client implementation
│   └── models/            # Data models
├── docs/                  # Generated Swagger documentation
├── pkg/
│   └── utils/             # Utility functions
├── scripts/               # Helper scripts
├── main.go                # Main entry point
├── Dockerfile             # Docker build configuration
├── docker-compose.yml     # Docker Compose configuration
└── Makefile               # Build automation
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.