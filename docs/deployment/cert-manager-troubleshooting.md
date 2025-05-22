# Cert-Manager Troubleshooting Guide

## Common Issues with cert-manager Webhook Validation

### Issue: Webhook TLS Certificate Validation Error

When applying ClusterIssuer or Certificate resources, you may encounter this error:

```
Error from server (InternalError): error when creating "manifests/letsencrypt-prod.yaml": Internal error occurred: failed calling webhook "webhook.cert-manager.io": failed to call webhook: Post "https://cert-manager-webhook.cert-manager.svc:443/validate?timeout=30s": tls: failed to verify certificate: x509: certificate signed by unknown authority
```

This error occurs when the Kubernetes API server can't validate against the cert-manager webhook because it doesn't trust the webhook's TLS certificate.

### Solutions

#### Solution 1: Wait Longer (Recommended First Step)

The cert-manager webhook needs time to fully register and initialize after installation:

```bash
# Wait for cert-manager components to be ready
echo "Waiting for cert-manager controller to be ready..."
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s

echo "Waiting for cert-manager webhook to be ready..."
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=webhook \
  --timeout=120s

echo "Waiting for cert-manager cainjector to be ready..."
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=cainjector \
  --timeout=120s

# Wait additional time for webhook API registration
echo "Waiting 60 seconds for cert-manager webhook to fully initialize..."
sleep 60
```

After waiting, try applying your resources again.

#### Solution 2: Restart the Webhook Pod

If waiting doesn't resolve the issue, try restarting the cert-manager webhook pod:

```bash
# Get the webhook pod name
WEBHOOK_POD=$(kubectl get pods -n cert-manager -l app.kubernetes.io/component=webhook -o jsonpath='{.items[0].metadata.name}')

# Delete the pod (it will be automatically recreated)
kubectl delete pod -n cert-manager $WEBHOOK_POD

# Wait for the new pod to be ready
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=webhook \
  --timeout=120s

# Wait additional time for webhook API registration
sleep 30
```

#### Solution 3: Temporarily Patch the Webhook to Ignore Validation Errors

If the above solutions don't work, you can temporarily patch the webhook to bypass validation:

```bash
# Patch the webhook to ignore validation failures
kubectl patch validatingwebhookconfiguration cert-manager-webhook --type='json' -p='[{"op": "replace", "path": "/webhooks/0/failurePolicy", "value": "Ignore"}]'

# Apply your resource
kubectl apply -f manifests/letsencrypt-prod.yaml

# Restore the webhook to its original state
kubectl patch validatingwebhookconfiguration cert-manager-webhook --type='json' -p='[{"op": "replace", "path": "/webhooks/0/failurePolicy", "value": "Fail"}]'
```

This solution temporarily modifies the webhook to ignore validation failures, allowing you to apply your resources. After applying, restore the webhook to its original state to maintain security.

#### Solution 4: Check for Clock Skew

If you're using a custom Kubernetes cluster, ensure there's no significant clock skew between nodes:

```bash
# Check system time on each node
kubectl get nodes -o wide | awk '{print $1}' | grep -v NAME | xargs -I{} kubectl debug {} -it --image=ubuntu -- date
```

#### Solution 5: Reinstall cert-manager

If all else fails, you may need to completely uninstall and reinstall cert-manager:

```bash
# Uninstall cert-manager
kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml

# Wait for resources to be removed
sleep 30

# Reinstall cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml

# Wait for cert-manager components to be ready (follow Solution 1)
```

## Preventive Measures in Deployment Scripts

To prevent this issue in your automated deployments, add these safeguards to your scripts:

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml

# Wait for cert-manager components to be ready
echo "Waiting for cert-manager controller to be ready..."
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s

echo "Waiting for cert-manager webhook to be ready..."
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=webhook \
  --timeout=120s

echo "Waiting for cert-manager cainjector to be ready..."
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=cainjector \
  --timeout=120s

# Additional delay to ensure the webhook API is fully registered
echo "Waiting 60 seconds for cert-manager webhook to fully initialize..."
sleep 60

# Apply the ClusterIssuer with auto-retry and fallback to webhook patching
echo "Applying ClusterIssuer..."
if ! kubectl apply -f manifests/letsencrypt-prod.yaml; then
  echo "First attempt failed. Waiting 30 more seconds..."
  sleep 30
  
  if ! kubectl apply -f manifests/letsencrypt-prod.yaml; then
    echo "Second attempt failed. Temporarily patching webhook to ignore validation errors..."
    kubectl patch validatingwebhookconfiguration cert-manager-webhook --type='json' -p='[{"op": "replace", "path": "/webhooks/0/failurePolicy", "value": "Ignore"}]'
    
    echo "Applying ClusterIssuer with validation ignored..."
    kubectl apply -f manifests/letsencrypt-prod.yaml
    
    echo "Restoring webhook configuration..."
    kubectl patch validatingwebhookconfiguration cert-manager-webhook --type='json' -p='[{"op": "replace", "path": "/webhooks/0/failurePolicy", "value": "Fail"}]'
  fi
fi

# Verify the ClusterIssuer was created
kubectl get clusterissuer letsencrypt-prod -o wide
```

This script automatically tries multiple approaches in sequence until one succeeds.

## Additional Troubleshooting Commands

If you're still encountering issues, these commands may help diagnose the problem:

```bash
# Check cert-manager pod logs
kubectl logs -n cert-manager -l app.kubernetes.io/component=webhook
kubectl logs -n cert-manager -l app.kubernetes.io/component=controller
kubectl logs -n cert-manager -l app.kubernetes.io/component=cainjector

# Check webhook configuration
kubectl get validatingwebhookconfiguration cert-manager-webhook -o yaml

# Check API service registration
kubectl get apiservice v1beta1.webhook.cert-manager.io -o yaml

# Check if the webhook endpoint is reachable
kubectl run -it --rm debug --image=curlimages/curl:7.86.0 --restart=Never -- \
  curl -k https://cert-manager-webhook.cert-manager.svc:443/validate
```