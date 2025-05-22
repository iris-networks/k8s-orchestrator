# Creating and Accessing User Environments

This guide walks you through creating and accessing user environments in the K8s Orchestrator.

## Create a User Environment

```bash
# Create a user environment via the API
# Note: In Autopilot, resource requests are required for all pods
curl -X POST https://api.pods.tryiris.dev/environments \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "image": "accetto/ubuntu-vnc-xfce-firefox-g3",
    "ports": [3000, 6901],
    "resources": {
      "requests": {
        "cpu": "250m",
        "memory": "256Mi"
      },
      "limits": {
        "cpu": "500m",
        "memory": "512Mi"
      }
    }
  }'
```

### Resource Allocation in Autopilot

Autopilot charges based on the resources requested by pods. Here's a breakdown of resource classes and approximate costs:

| Resource Class | vCPU       | Memory     | Approximate Cost* |
|---------------|------------|------------|-------------------|
| Small         | 0.25 - 1   | 0.5 - 4 GB | $40-60/mo         |
| Medium        | 1 - 4      | 4 - 16 GB  | $150-250/mo       |
| Large         | 4 - 8      | 16 - 32 GB | $500-900/mo       |

*Costs are approximate and vary by region

#### Our Enhanced VNC Environment Resources

For VNC environments, we've configured resources for better performance:

```yaml
resources:
  requests:
    cpu: 500m        # 0.5 vCPU
    memory: 512Mi    # 512 MB RAM
  limits:
    cpu: 1000m       # 1 vCPU
    memory: 1024Mi   # 1 GB RAM
```

This configuration provides a good balance of performance and cost. With our resource quotas, each user environment can scale up to 4 CPU cores and 8GB memory if needed.

## Access the User Environment

The user environment provides two main interfaces, both accessible via HTTPS:

1. **Web Server (Port 3000)**: Access the user's web server
   - URL: `https://testuser.pods.tryiris.dev:3000`
   - This is where the user's web application is served

2. **VNC Web Interface (Port 6901)**: Access the VNC desktop environment
   - URL: `https://testuser.pods.tryiris.dev:6901`
   - Provides a web-based VNC interface to the virtual desktop
   - Credentials are typically `headless:headless` (default for the VNC image)

## Verify the Created Resources

```bash
# Check the namespace
kubectl get namespace user-testuser

# Check resources in the user's namespace
kubectl get all -n user-testuser

# Check the persistent volume claim
kubectl get pvc -n user-testuser

# Check the ingress
kubectl get ingressroute -n user-testuser
```

## Next Step

Once you have created and accessed user environments, proceed to [Scale and Manage Deployment](08-scaling-management.md).