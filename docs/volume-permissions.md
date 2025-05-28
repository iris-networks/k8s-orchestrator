# Volume Permissions in Kubernetes

This document explains how persistent volume permissions are handled in the k8sgo application.

## Overview

The k8sgo application creates Kubernetes pods that run as `nodeuser` (UID 1001, GID 1001) and need to write to persistent volumes mounted at `/home/nodeuser/.iris`. By default, Kubernetes volume mounts often have permissions that prevent non-root users from writing to them.

## Solution Implementation

We've implemented two key strategies to ensure proper permissions:

### 1. PodSecurityContext

The pod specification includes a SecurityContext with:

```yaml
securityContext:
  fsGroup: 1001  # nodeuser's group ID
  runAsUser: 1001  # nodeuser's user ID
  runAsGroup: 1001
  fsGroupChangePolicy: OnRootMismatch  # More efficient permission changes
```

The `fsGroup` setting ensures the volume is accessible to the group, while `fsGroupChangePolicy: OnRootMismatch` makes permission changes more efficient by only applying them when necessary.

### 2. Init Container

An init container runs before the main application container to explicitly set permissions:

```yaml
initContainers:
- name: volume-permissions
  image: busybox:latest
  command:
  - sh
  - -c
  - mkdir -p /home/nodeuser/.iris && chown -R 1001:1001 /home/nodeuser/.iris
  volumeMounts:
  - name: user-data
    mountPath: /home/nodeuser/.iris
```

This init container ensures that:
1. The `.iris` directory exists
2. It has the correct ownership (nodeuser:nodeuser)

## Testing

When testing new deployments, you can verify permissions are correct by:

1. Connecting to the pod:
   ```bash
   kubectl exec -it <pod-name> -n user-sandboxes -- /bin/bash
   ```

2. Checking permissions:
   ```bash
   ls -la /home/nodeuser/.iris
   ```

3. Testing write access:
   ```bash
   sudo -u nodeuser touch /home/nodeuser/.iris/test.txt
   ```

## Troubleshooting

If permission issues persist:

1. Verify the UID/GID:
   ```bash
   kubectl exec -n user-sandboxes <pod-name> -- id nodeuser
   ```

2. Check volume mount permissions:
   ```bash
   kubectl exec -n user-sandboxes <pod-name> -- ls -la /home/nodeuser
   ```

3. Inspect the pod's security context:
   ```bash
   kubectl get pod <pod-name> -n user-sandboxes -o yaml | grep -A15 securityContext
   ```

## Storage Class Considerations

Different storage classes may handle permissions differently. If you're using a storage class that doesn't respect `fsGroup`, you may need to rely more heavily on the init container approach.