# Troubleshooting

This guide provides solutions for common issues that may arise during deployment or operation of the K8s Orchestrator.

## Common Issues and Solutions

### Autopilot-Specific Issues

If you encounter issues with GKE Autopilot:

```bash
# Error: Pod scheduling failure due to resource requirements
# Solution: Ensure all pods have resource requests set
kubectl describe pod [pod-name]  # Check for resource-related events

# Warning: autopilot-default-resources-mutator
# Solution: This is normal, Autopilot is setting default resource requirements
# See: docs/admin/autopilot-resource-defaults.md

# Error: Unsupported features in Autopilot
# Solution: Check Autopilot limitations
# Common limitations:
# - DaemonSets require special permissions
# - Local persistent volumes are not supported
# - Some privileged operations are restricted
# - Custom node configurations are not available
gcloud container clusters describe k8s-orchestrator --region asia-southeast1 | grep autopilot

# Switch to Standard mode if Autopilot limitations are too restrictive
gcloud container clusters update k8s-orchestrator --region asia-southeast1 --no-enable-autopilot
```

### Kubelet Readonly Port Warning

If you see warnings about the deprecated Kubelet readonly port:

```
Note: The Kubelet readonly port (10255) is now deprecated. Please update your workloads to use the recommended alternatives.
```

This is a standard warning from GKE. In most cases, you don't need to do anything as your workloads likely don't use this port directly.

### GKE Auth Plugin and Kubectl Issues

If you encounter errors related to kubectl authentication or version compatibility:

```bash
# Error: "CRITICAL: ACTION REQUIRED: gke-gcloud-auth-plugin, which is needed for continued use of kubectl, was not found or is not executable."
# Solution for macOS: Install the GKE auth plugin
gcloud components install gke-gcloud-auth-plugin

# Error: "WARNING: version difference between client (X.Y) and server (A.B) exceeds the supported minor version skew of +/-1"
# Solution: Install the compatible kubectl version using our script

# For both issues, run our script to fix everything at once:
bash scripts/install_kubectl.sh
```

### Project ID and Quota Issues

If you encounter errors related to project ID or quota:

```bash
# Error: The value of ``core/project'' property is set to project number
# Solution: Set the project ID (not number)
gcloud projects list  # Get the PROJECT_ID
gcloud config set project YOUR_PROJECT_ID  # Use the ID, not the number

# Error: Quota issues or unexpected billing
# Solution: Set the quota project
gcloud auth application-default set-quota-project YOUR_PROJECT_ID

# Error: Authentication issues
# Solution: Re-login and select the correct account
gcloud auth login
gcloud config set account YOUR_EMAIL@DOMAIN.COM
```

### Ingress Issues

If your Traefik IngressRoute is not working properly:

```bash
# Check IngressRoute status
kubectl get ingressroute -l app.kubernetes.io/instance=k8s-orchestrator

# Check IngressRoute events
kubectl describe ingressroute -l app.kubernetes.io/instance=k8s-orchestrator

# Check Traefik middleware resources if used
kubectl get middleware -l app.kubernetes.io/instance=k8s-orchestrator

# Check Traefik Ingress Controller logs
kubectl logs -l app.kubernetes.io/name=traefik

# Check Traefik dashboard if enabled
kubectl port-forward $(kubectl get pods -l app.kubernetes.io/name=traefik -o name) 9000:9000
# Then open http://localhost:9000/dashboard/ in your browser
```

### Certificate Issues

If TLS certificates are not being issued with Traefik's built-in Let's Encrypt integration:

```bash
# Check Traefik logs for ACME (Let's Encrypt) related messages
kubectl logs -l app.kubernetes.io/name=traefik | grep -i acme

# Check if the ACME storage file exists in Traefik's persistent volume
kubectl exec -it $(kubectl get pods -l app.kubernetes.io/name=traefik -o name | head -n 1) -- ls -la /data

# Check Traefik's IngressRoute definitions to ensure TLS is properly configured
kubectl get ingressroute -o yaml | grep -A 10 tls

# Verify that Traefik can reach Let's Encrypt servers
kubectl exec -it $(kubectl get pods -l app.kubernetes.io/name=traefik -o name | head -n 1) -- wget -q -O- https://acme-v02.api.letsencrypt.org/directory

# For detailed certificate debugging, enable debug logs in Traefik
kubectl edit deployment traefik  # Add "--log.level=DEBUG" to args, then check logs again
```

### Pod Startup Issues

If pods are not starting properly:

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/instance=k8s-orchestrator

# Check pod events
kubectl describe pod -l app.kubernetes.io/instance=k8s-orchestrator

# Check container logs
kubectl logs -l app.kubernetes.io/instance=k8s-orchestrator
```

### User Environment Issues

If user environments are not being created:

```bash
# Check the orchestrator logs
kubectl logs -l app.kubernetes.io/instance=k8s-orchestrator

# Verify RBAC permissions
kubectl auth can-i create namespaces --as=system:serviceaccount:default:k8s-orchestrator
kubectl auth can-i create deployments --as=system:serviceaccount:default:k8s-orchestrator
kubectl auth can-i create ingress --as=system:serviceaccount:default:k8s-orchestrator
```

## Next Step

For cleanup instructions, proceed to [Cleanup](10-cleanup.md).