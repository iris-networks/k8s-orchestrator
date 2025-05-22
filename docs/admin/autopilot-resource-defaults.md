# Understanding GKE Autopilot Resource Defaults

When deploying applications to GKE Autopilot, you may see warnings like:

```
Warning: autopilot-default-resources-mutator:Autopilot updated Deployment cert-manager/cert-manager-webhook: defaulted unspecified 'cpu' resource for containers [cert-manager-webhook] (see http://g.co/gke/autopilot-defaults).
```

These warnings are normal and expected. This document explains what they mean and why you don't need to worry about them.

## What These Warnings Mean

GKE Autopilot requires that all containers specify resource requests and limits. If you deploy workloads that don't specify these values, Autopilot automatically assigns default values.

These warnings are simply informing you that:

1. The deployment did not have complete resource specifications
2. Autopilot automatically applied default values
3. The deployment was successful with these applied defaults

## Why This Happens with Common Tools

Many popular Kubernetes tools like NGINX Ingress Controller and cert-manager have deployment manifests that don't fully specify resource requirements for all containers. When you install these tools in Autopilot, the cluster automatically adds the missing specifications.

Common examples:

- **NGINX Ingress Controller**: Some sidecar containers may not have CPU or memory specifications
- **cert-manager**: The webhook and cainjector containers often lack complete resource specifications
- **Prometheus/Grafana**: Some components may not specify resource requests

## Default Resource Values

Autopilot typically applies the following defaults:

| Resource | Default Request | Default Limit |
|----------|----------------|--------------|
| CPU      | 250m (0.25 cores) | 1 core     |
| Memory   | 256Mi          | 512Mi        |

These values are designed to be appropriate for most sidecar and helper containers.

## Do You Need to Take Action?

**No action required**. These warnings are informational only. Your deployments are working correctly with the automatically applied resource specifications.

If you want to eliminate the warnings, you can:

1. Use Helm charts that allow specifying resources for all containers
2. Apply custom resource specifications to override the defaults
3. Create PodPatches that apply your desired resource specifications

## Impact on Billing

The automatically applied resource requests do affect your billing in Autopilot, as Autopilot bills based on requested resources. However, the default values are relatively small and designed to be cost-effective for most workloads.

## Further Information

For more details on Autopilot resource management, see the [official GKE documentation](https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-resource-requests).