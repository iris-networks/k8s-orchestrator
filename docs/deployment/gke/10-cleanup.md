# Cleaning Up

This guide walks you through cleaning up resources when they're no longer needed.

## Delete a User Environment

```bash
# Delete a user environment via the API
curl -X DELETE https://api.pods.tryiris.dev/environments/testuser
```

## Uninstall the Helm Release

```bash
# Uninstall the Helm release
helm uninstall k8s-orchestrator
```

## Delete User Namespaces

```bash
# List all user namespaces
kubectl get namespaces -l app=k8s-orchestrator

# Delete all user namespaces
kubectl get namespaces -l app=k8s-orchestrator | grep user- | awk '{print $1}' | xargs kubectl delete namespace
```

## Delete the GKE Cluster

```bash
# Delete the GKE cluster
gcloud container clusters delete k8s-orchestrator --region asia-southeast1
```

## Verify Cleanup

To ensure all resources have been cleaned up, check:

1. Google Cloud Console > Kubernetes Engine > Clusters (should show no clusters)
2. Google Cloud Console > Kubernetes Engine > Workloads (should show no workloads)
3. Google Cloud Console > Kubernetes Engine > Services & Ingress (should show no services)
4. Google Cloud Console > Compute Engine > VM instances (should show no VMs)
5. Google Cloud Console > Billing > Reports (to verify billing has stopped for GKE)

## Next Step

You've completed the GKE deployment guide. To deploy again, start with [Prerequisites](01-prerequisites.md).