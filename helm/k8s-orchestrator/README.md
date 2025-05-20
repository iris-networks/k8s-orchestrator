# K8s Orchestrator Helm Chart

This Helm chart deploys the K8s Orchestrator service, which creates isolated Kubernetes environments for users with persistent storage and dynamic subdomains, using VNC for desktop access.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Ingress controller installed (nginx, alb, etc.)
- Cert-manager (for TLS certificates)
- DNS setup for subdomains

## Usage

### Add the Repository

```bash
# This is a placeholder - replace with actual repository info once published
helm repo add k8s-orchestrator https://your-helm-repo.example.com
helm repo update
```

### Install the Chart

```bash
# Create a values.yaml file with your configurations
helm install k8s-orchestrator k8s-orchestrator/k8s-orchestrator \
  --namespace default \
  --create-namespace \
  -f values.yaml
```

### Installing on AWS EKS

For AWS EKS, use values like:

```yaml
# values-aws.yaml
cloudProvider:
  aws:
    enabled: true
    ingressClass: "alb"
    annotations:
      alb.ingress.kubernetes.io/scheme: "internet-facing"
      alb.ingress.kubernetes.io/target-type: "ip"
      alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}, {"HTTPS": 443}]'
      alb.ingress.kubernetes.io/ssl-redirect: "443"
      external-dns.alpha.kubernetes.io/hostname: "api.your-domain.com" 
    storageClass: "gp2"

env:
  DOMAIN: "your-domain.com"

ingress:
  hosts:
    - host: api.your-domain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: k8s-orchestrator-tls
      hosts:
        - api.your-domain.com
```

```bash
helm install k8s-orchestrator k8s-orchestrator/k8s-orchestrator \
  --namespace default \
  -f values-aws.yaml
```

### Installing on Google Cloud GKE

For GCP GKE, use values like:

```yaml
# values-gcp.yaml
cloudProvider:
  gcp:
    enabled: true
    ingressClass: "nginx"
    annotations:
      nginx.ingress.kubernetes.io/proxy-connect-timeout: "3600"
      nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
      nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
      external-dns.alpha.kubernetes.io/hostname: "api.your-domain.com" 
    storageClass: "standard-rwo"

env:
  DOMAIN: "your-domain.com"

ingress:
  hosts:
    - host: api.your-domain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: k8s-orchestrator-tls
      hosts:
        - api.your-domain.com
```

```bash
helm install k8s-orchestrator k8s-orchestrator/k8s-orchestrator \
  --namespace default \
  -f values-gcp.yaml
```

## Parameters

### Common Parameters

| Name                | Description                                                        | Value                       |
|---------------------|--------------------------------------------------------------------|------------------------------|
| `replicaCount`      | Number of replicas                                                | `2`                          |
| `image.repository`  | Container image repository                                        | `k8s-orchestrator`           |
| `image.tag`         | Container image tag                                               | `latest`                     |
| `image.pullPolicy`  | Container image pull policy                                       | `IfNotPresent`               |
| `env.DOMAIN`        | Domain for user environments                                      | `local.dev`                  |

### Cloud Provider Configuration

| Name                                   | Description                                      | Value     |
|----------------------------------------|--------------------------------------------------|-----------|
| `cloudProvider.aws.enabled`            | Enable AWS specific configuration                | `false`   |
| `cloudProvider.aws.ingressClass`       | AWS ALB ingress class                            | `alb`     |
| `cloudProvider.aws.storageClass`       | AWS EBS storage class                            | `gp2`     |
| `cloudProvider.gcp.enabled`            | Enable GCP specific configuration                | `false`   |
| `cloudProvider.gcp.ingressClass`       | GCP ingress class                                | `nginx`   |
| `cloudProvider.gcp.storageClass`       | GCP storage class                                | `standard-rwo` |

### User Environment Configuration

| Name                                   | Description                                      | Value     |
|----------------------------------------|--------------------------------------------------|-----------|
| `userEnvironments.defaultImage`        | Default container image for user environments    | `accetto/ubuntu-vnc-xfce-firefox-g3` |
| `userEnvironments.defaultPorts`        | Default ports for user environments              | `[5901, 6901]` |
| `userEnvironments.defaultVolumeSize`   | Default volume size for user environments        | `1Gi`    |
| `userEnvironments.storageClass`        | Storage class for user PVCs (overrides cloud provider setting) | `""` |

## Uninstalling the Chart

```bash
helm uninstall k8s-orchestrator --namespace default
```

## Limitations

- The service account needs cluster-level permissions to create namespaces and other resources
- For production use, make sure to set appropriate resource limits
- For highly available setups, use multiple replicas and set up proper anti-affinity rules