# Scaling and Managing the Deployment

This guide walks you through scaling and managing your K8s Orchestrator deployment.

## Cost Control in Autopilot

In Autopilot mode, you can't directly control node count, but you can control costs through the following methods:

### 1. ResourceQuotas

We always use ResourceQuotas to limit total CPU and memory usage:

```bash
# We already created a default ResourceQuota during cluster setup
# For each user namespace, we add an additional ResourceQuota to limit their resources:
kubectl create -f - <<EOF
apiVersion: v1
kind: ResourceQuota
metadata:
  name: user-compute-resources
  namespace: user-environment-namespace
spec:
  hard:
    requests.cpu: "2"      # Maximum 2 cores per user
    requests.memory: 4Gi   # Maximum 4GB memory per user
    limits.cpu: "4"        # Maximum 4 cores burst per user
    limits.memory: 8Gi     # Maximum 8GB memory burst per user
    pods: "5"              # Maximum 5 pods per user
EOF
```

Check existing quotas:

```bash
kubectl get resourcequota --all-namespaces
```

### 2. LimitRanges

We apply LimitRanges to every namespace to enforce resource constraints:

```bash
# Apply a LimitRange to each user namespace to ensure all containers have appropriate limits
kubectl create -f - <<EOF
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
  namespace: user-environment-namespace
spec:
  limits:
  - default:
      cpu: 500m         # Default limit of 0.5 cores
      memory: 512Mi     # Default limit of 512MB RAM
    defaultRequest:
      cpu: 250m         # Default request of 0.25 cores
      memory: 256Mi     # Default request of 256MB RAM
    max:
      cpu: "2"          # Maximum of 2 cores per container
      memory: 4Gi       # Maximum of 4GB RAM per container
    min:
      cpu: 100m         # Minimum of 0.1 cores per container
      memory: 64Mi      # Minimum of 64MB RAM per container
    type: Container
EOF
```

Verify the LimitRange:

```bash
kubectl get limitrange -n user-environment-namespace
```

### 3. Pod Resource Specifications

Always set appropriate requests/limits in pod specs for all deployments.

### 4. Budget Notifications

We always set up Google Cloud budget alerts:

```bash
# Get your billing account ID
gcloud billing accounts list
```

```bash
# Set up a budget alert via gcloud (REQUIRED for all clusters)
gcloud billing budgets create \
  --billing-account=YOUR_BILLING_ACCOUNT_ID \
  --display-name="GKE Budget - k8s-orchestrator" \
  --budget-amount=500 \
  --threshold-rules=threshold-percent=0.5,basis=current_spend \
  --threshold-rules=threshold-percent=0.8,basis=current_spend \
  --threshold-rules=threshold-percent=1.0,basis=current_spend \
  --email=your-team-email@tryiris.dev
```

## Scaling the Orchestrator Service

```bash
# Scale the deployment manually (pod count only in Autopilot)
kubectl scale deployment k8s-orchestrator --replicas=3

# Or update the Helm values and upgrade
helm upgrade k8s-orchestrator . -f values-gcp-autopilot.yaml
```

## Upgrading the Deployment

```bash
# Pull the latest image
docker pull shanurcsenitap/irisk8s:latest

# Update your values file if needed
# Then upgrade the Helm release
helm upgrade k8s-orchestrator . -f values-gcp-autopilot.yaml
```

## Next Step

For troubleshooting information, proceed to [Troubleshooting](09-troubleshooting.md).