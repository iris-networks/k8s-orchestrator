# Traefik Routing Configuration Guide for K8s Sandbox Platform

This guide provides detailed instructions for configuring Traefik to handle dynamic subdomain routing for the K8s Sandbox Platform.

## Overview

The K8s Sandbox Platform requires Traefik to route traffic to user sandboxes using dynamic subdomains:
- `{user}-vnc.pods.{domain}` → Port 6901 (VNC access)
- `{user}-api.pods.{domain}` → Port 3000 (HTTP/REST access)

## Prerequisites

- GKE Autopilot cluster with Traefik installed
- Domain name with DNS configured
- `kubectl` configured to use your cluster

## Step 1: Install Traefik with Helm

First, install Traefik using Helm:

```bash
# Add the Traefik Helm repository
helm repo add traefik https://traefik.github.io/charts
helm repo update

# Create a namespace for Traefik
kubectl create namespace traefik

# Install Traefik with appropriate values
helm install traefik traefik/traefik \
  --namespace=traefik \
  --set ingressClass.enabled=true \
  --set ingressClass.isDefaultClass=true \
  --set ports.websecure.tls.enabled=true \
  --set additionalArguments="{--providers.kubernetescrd,--providers.kubernetesingress}"
```

## Step 2: Configure DNS Records

After Traefik is deployed, get its external IP:

```bash
TRAEFIK_IP=$(kubectl get service traefik -n traefik -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
echo "Traefik External IP: ${TRAEFIK_IP}"
```

Configure your DNS records:

1. Create an A record for `api.${DOMAIN}` pointing to `${TRAEFIK_IP}`
2. Create a wildcard A record for `*.pods.${DOMAIN}` pointing to `${TRAEFIK_IP}`

## Step 3: Configure Traefik CRDs for the Sandbox Platform

There are two ways to configure routing in Traefik: using Kubernetes Ingress resources or using Traefik's own CRDs.

### Option 1: Using Kubernetes Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k8sgo-management-api
  namespace: k8sgo-system
  annotations:
    kubernetes.io/ingress.class: traefik
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: k8sgo
            port:
              number: 8080
```

### Option 2: Using Traefik CRDs (Recommended)

For more advanced routing, Traefik CRDs provide greater flexibility:

```yaml
# IngressRoute for the management API
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: k8sgo-management-api
  namespace: k8sgo-system
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`api.example.com`)
      kind: Rule
      services:
        - name: k8sgo
          port: 8080
  tls: {}
```

## Step 4: Dynamic Routing for User Sandboxes

The K8s Sandbox Platform needs to create dynamic routing rules for each user sandbox. This is handled by the application, which creates Traefik IngressRoute resources programmatically.

Here's an example of the IngressRoute resources that will be created for a user "test-user":

```yaml
# IngressRoute for VNC access
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: test-user-vnc
  namespace: user-sandboxes
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`test-user-vnc.pods.example.com`)
      kind: Rule
      services:
        - name: test-user
          port: 6901
  tls: {}

# IngressRoute for API access
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: test-user-api
  namespace: user-sandboxes
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`test-user-api.pods.example.com`)
      kind: Rule
      services:
        - name: test-user
          port: 3000
  tls: {}
```

## Step 5: TLS Configuration with Let's Encrypt

For secure connections, configure Traefik to use Let's Encrypt for automatic TLS certificate generation:

1. Install cert-manager:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.1/cert-manager.yaml
```

2. Create a ClusterIssuer for Let's Encrypt:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    email: admin@example.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
```

3. Update your IngressRoute to use TLS with the cert-manager issuer:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: k8sgo-management-api
  namespace: k8sgo-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`api.example.com`)
      kind: Rule
      services:
        - name: k8sgo
          port: 8080
  tls:
    secretName: api-tls-cert
```

## Step 6: Traefik Middleware for Enhanced Security

You can add middleware to your routes for additional security and functionality:

### Basic Authentication for the Management API

1. Create a Secret with the credentials:

```bash
htpasswd -c auth admin
kubectl create secret generic admin-auth --from-file=auth -n k8sgo-system
```

2. Create a Middleware resource:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: admin-auth
  namespace: k8sgo-system
spec:
  basicAuth:
    secret: admin-auth
```

3. Reference the middleware in your IngressRoute:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: k8sgo-management-api
  namespace: k8sgo-system
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`api.example.com`)
      kind: Rule
      middlewares:
        - name: admin-auth
      services:
        - name: k8sgo
          port: 8080
  tls: {}
```

### Rate Limiting for User Sandboxes

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: rate-limit
  namespace: user-sandboxes
spec:
  rateLimit:
    average: 100
    burst: 50
```

Reference this middleware in your user sandbox IngressRoutes to prevent abuse.

## Implementation in the K8s Sandbox Platform

In the K8s Sandbox Platform, the following code is responsible for creating the necessary Kubernetes resources for routing:

```go
// Create Traefik IngressRoute for VNC access
vncIngressRoute := &v1alpha1.IngressRoute{
    ObjectMeta: metav1.ObjectMeta{
        Name:      fmt.Sprintf("%s-vnc", userId),
        Namespace: "user-sandboxes",
    },
    Spec: v1alpha1.IngressRouteSpec{
        EntryPoints: []string{"websecure"},
        Routes: []v1alpha1.Route{
            {
                Match: fmt.Sprintf("Host(`%s-vnc.pods.%s`)", userId, domain),
                Kind:  "Rule",
                Services: []v1alpha1.Service{
                    {
                        Name: userId,
                        Port: 6901,
                    },
                },
            },
        },
        TLS: &v1alpha1.TLS{},
    },
}

// Create Traefik IngressRoute for API access
apiIngressRoute := &v1alpha1.IngressRoute{
    ObjectMeta: metav1.ObjectMeta{
        Name:      fmt.Sprintf("%s-api", userId),
        Namespace: "user-sandboxes",
    },
    Spec: v1alpha1.IngressRouteSpec{
        EntryPoints: []string{"websecure"},
        Routes: []v1alpha1.Route{
            {
                Match: fmt.Sprintf("Host(`%s-api.pods.%s`)", userId, domain),
                Kind:  "Rule",
                Services: []v1alpha1.Service{
                    {
                        Name: userId,
                        Port: 3000,
                    },
                },
            },
        },
        TLS: &v1alpha1.TLS{},
    },
}

// Create the IngressRoutes
k8sClient.TraefikV1alpha1().IngressRoutes("user-sandboxes").Create(context.TODO(), vncIngressRoute, metav1.CreateOptions{})
k8sClient.TraefikV1alpha1().IngressRoutes("user-sandboxes").Create(context.TODO(), apiIngressRoute, metav1.CreateOptions{})
```

## Troubleshooting

### Check Traefik Logs

```bash
kubectl logs -n traefik -l app.kubernetes.io/name=traefik
```

### Verify IngressRoute Configuration

```bash
kubectl get ingressroutes -A
kubectl describe ingressroute <ingressroute-name> -n <namespace>
```

### Check Traefik Dashboard (if enabled)

```bash
# Port-forward to the Traefik dashboard
kubectl port-forward -n traefik $(kubectl get pods -n traefik -l app.kubernetes.io/name=traefik -o name) 9000:9000
```

Then access the dashboard at http://localhost:9000/dashboard/

## Conclusion

Proper Traefik configuration is essential for the K8s Sandbox Platform to function correctly. The dynamic subdomain routing allows each user to have their own isolated sandbox environment with unique access points.

For more information, refer to:
- [Traefik Kubernetes CRD Documentation](https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/)
- [Traefik Kubernetes Ingress Documentation](https://doc.traefik.io/traefik/routing/providers/kubernetes-ingress/)
- [Traefik Middleware Documentation](https://doc.traefik.io/traefik/middlewares/overview/)