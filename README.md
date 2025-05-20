# Kubernetes User Environment Service

A NestJS service that creates isolated Kubernetes environments for users with persistent storage and dynamic subdomains. Built with:

- NestJS framework
- Kubernetes JavaScript client (@kubernetes/client-node)
- Kind for local testing
- VNC containers for GUI access

## Features

- Creates isolated user environments in Kubernetes using namespaces
- Provisions persistent storage via PVCs for each user
- Generates and configures subdomain access for each environment
- Supports destroying and recreating containers without data loss
- Provides VNC-based GUI access to environments
- OpenAPI/Swagger documentation for API endpoints

## Requirements

- Node.js (v16+)
- npm or yarn
- Docker
- Kind (Kubernetes in Docker)
- kubectl

## Project Structure

```
k8s-service/
├── src/
│   ├── kubernetes/         # Kubernetes API integration
│   ├── environments/       # User environment management
│   └── config/             # Configuration module
├── kind-config.yaml        # Kind cluster configuration
├── setup-kind.sh           # Setup script for local testing
└── .env.example            # Environment variables template
```

## Getting Started

### Clone the repository

```bash
git clone <repository-url>
cd k8s-service
```

### Setup local Kubernetes with Kind

1. Make sure you have Kind installed:
   ```bash
   brew install kind  # macOS with Homebrew
   ```

2. Run the setup script to create a Kind cluster with necessary components:
   ```bash
   ./setup-kind.sh
   ```

3. Set up local DNS entries for testing:
   Add the following entry to your `/etc/hosts` file:
   ```
   127.0.0.1  alice.local.dev bob.local.dev
   ```

### Configure the environment

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit the `.env` file with appropriate values:
   - Set `KUBECONFIG` to the path output from the setup script
   - Adjust other settings as needed

### Install dependencies and run the application

```bash
npm install
npm run start:dev
```

The API will be available at http://localhost:3000

## API Documentation

### Swagger UI

The API is documented using Swagger/OpenAPI. You can access the interactive documentation at:

```
http://localhost:3000/api
```

The Swagger UI allows you to:
- View all available endpoints
- Understand the required parameters and expected responses
- Test API calls directly from the browser

### Environments API

- `GET /environments` - List all environments
- `GET /environments?username=alice` - List environments for a specific user
- `GET /environments/:id` - Get details about a specific environment
- `POST /environments` - Create a new environment
- `DELETE /environments/:id` - Delete an environment
- `PUT /environments/:id/restart` - Restart an environment

### Example API Usage

**Create a new environment:**

```bash
curl -X POST http://localhost:3000/environments \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "storageSize": "5Gi"}'
```

**List all environments:**

```bash
curl http://localhost:3000/environments
```

**Delete an environment:**

```bash
curl -X DELETE http://localhost:3000/environments/some-environment-id
```

## Environment Access

After creating an environment for a user (e.g., "alice"), the user can access their environment via:

- VNC GUI: https://alice.local.dev

## Production Considerations

For production deployment, consider:

1. Using a real database for environment persistence
2. Implementing proper authentication and authorization
3. Setting up proper DNS with a wildcard certificate from Let's Encrypt
4. Configuring resource limits and quotas
5. Implementing monitoring and logging

## Troubleshooting

If you encounter issues:

1. Check the NestJS logs
2. Examine Kubernetes resources with kubectl:
   ```bash
   kubectl get ns
   kubectl get all -n user-alice
   kubectl describe pod -n user-alice <pod-name>
   ```
3. Check ingress configuration:
   ```bash
   kubectl get ingress -A
   kubectl describe ingress -n user-alice
   ```