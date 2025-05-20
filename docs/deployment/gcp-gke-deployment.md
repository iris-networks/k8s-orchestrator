# Deploying to Google Kubernetes Engine (GKE)

This guide provides step-by-step instructions for deploying the K8s Orchestrator service on Google Kubernetes Engine (GKE).

## Prerequisites

- Google Cloud SDK (`gcloud`) installed and configured
- `kubectl` installed and configured
- `helm` installed (v3+)
- Docker installed (for building and pushing images)
- An active Google Cloud account with a project and billing enabled

## Infrastructure Setup

### 1. Set up Environment Variables

```bash
# Set your GCP project ID
export PROJECT_ID=$(gcloud config get-value project)
export REGION=us-central1
export ZONE=us-central1-a
export CLUSTER_NAME=k8s-orchestrator-cluster
```

### 2. Create a GKE Cluster

```bash
# Create the GKE cluster
gcloud container clusters create ${CLUSTER_NAME} \
  --project=${PROJECT_ID} \
  --zone=${ZONE} \
  --num-nodes=3 \
  --machine-type=e2-standard-2 \
  --network=default \
  --enable-autoscaling \
  --min-nodes=2 \
  --max-nodes=5 \
  --enable-ip-alias \
  --cluster-version=latest

# Get credentials for the cluster
gcloud container clusters get-credentials ${CLUSTER_NAME} --zone=${ZONE} --project=${PROJECT_ID}
```

### 3. Install Required Components

#### a. Nginx Ingress Controller

```bash
# Add the Nginx Ingress Helm repository
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Install Nginx Ingress Controller
helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.loadBalancerIP="" \
  --set controller.service.type=LoadBalancer

# Get the load balancer IP
export INGRESS_IP=$(kubectl get service nginx-ingress-ingress-nginx-controller \
  -n ingress-nginx \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "Ingress IP: ${INGRESS_IP}"
```

#### b. External DNS (for subdomain management)

```bash
# Create a service account for External DNS
gcloud iam service-accounts create external-dns \
  --display-name "External DNS"

# Add the necessary IAM bindings for DNS management
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member serviceAccount:external-dns@${PROJECT_ID}.iam.gserviceaccount.com \
  --role roles/dns.admin

# Create a Kubernetes service account and bind it to the GCP service account
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[external-dns/external-dns]" \
  external-dns@${PROJECT_ID}.iam.gserviceaccount.com

# Install External DNS
kubectl create namespace external-dns

kubectl apply -f - << EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
  namespace: external-dns
  annotations:
    iam.gke.io/gcp-service-account: external-dns@${PROJECT_ID}.iam.gserviceaccount.com
EOF

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install external-dns bitnami/external-dns \
  --namespace external-dns \
  --set provider=google \
  --set google.project=${PROJECT_ID} \
  --set serviceAccount.create=false \
  --set serviceAccount.name=external-dns \
  --set domainFilters[0]=your-domain.com
```

#### c. Cert Manager (for TLS certificates)

```bash
# Install Cert Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Create a ClusterIssuer for Let's Encrypt
cat > letsencrypt-prod.yaml << EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

kubectl apply -f letsencrypt-prod.yaml
```

### 4. Configure Storage Classes

GKE comes with a default storage class, but you might want to create a specific one:

```bash
cat > standard-rwo.yaml << EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard-rwo
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-standard
  fstype: ext4
reclaimPolicy: Delete
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
EOF

kubectl apply -f standard-rwo.yaml
```

## Application Deployment

### 1. Configure DNS in Cloud DNS

```bash
# Create a Cloud DNS zone if you don't already have one
gcloud dns managed-zones create your-domain-zone \
  --dns-name="your-domain.com." \
  --description="Zone for your-domain.com"

# Note the nameservers
gcloud dns managed-zones describe your-domain-zone \
  --format="value(nameServers)"

# Update your domain's name servers at your registrar with the nameservers provided
```

### 2. Build and Push the Docker Image

```bash
# Configure Docker to use the gcloud command-line tool as a credential helper
gcloud auth configure-docker

# Build and tag the image
docker build -t gcr.io/${PROJECT_ID}/k8s-orchestrator:latest .

# Push the image
docker push gcr.io/${PROJECT_ID}/k8s-orchestrator:latest
```

### 3. Deploy the Application

Create a deployment YAML file:

```bash
cat > k8s-orchestrator-gke.yaml << EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-orchestrator-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-orchestrator-role
rules:
- apiGroups: [""]
  resources: ["namespaces", "services", "persistentvolumeclaims", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8s-orchestrator-binding
subjects:
- kind: ServiceAccount
  name: k8s-orchestrator-sa
  namespace: default
roleRef:
  kind: ClusterRole
  name: k8s-orchestrator-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-orchestrator
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: k8s-orchestrator
  template:
    metadata:
      labels:
        app: k8s-orchestrator
    spec:
      serviceAccountName: k8s-orchestrator-sa
      containers:
      - name: k8s-orchestrator
        image: gcr.io/${PROJECT_ID}/k8s-orchestrator:latest
        imagePullPolicy: Always
        env:
        - name: DOMAIN
          value: "your-domain.com"
        - name: PORT
          value: "8080"
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: k8s-orchestrator
  namespace: default
spec:
  selector:
    app: k8s-orchestrator
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k8s-orchestrator
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    external-dns.alpha.kubernetes.io/hostname: "api.your-domain.com"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - api.your-domain.com
    secretName: k8s-orchestrator-tls
  rules:
  - host: api.your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: k8s-orchestrator
            port:
              number: 80
EOF

# Apply the configuration
kubectl apply -f k8s-orchestrator-gke.yaml
```

## Code Modifications

### 1. Update the Ingress Creation Logic

Modify `internal/k8s/resources.go` to use the appropriate annotations for Nginx Ingress Controller:

```go
func (c *Client) createIngress(username, namespace string, port int) error {
    ingressName := fmt.Sprintf("%s-ingress", username)
    host := fmt.Sprintf("%s.%s", username, c.domain)
    pathType := networkingv1.PathTypePrefix
    
    ingressClassName := "nginx"
    
    ingress := &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name: ingressName,
            Labels: map[string]string{
                "app":      "k8sgo",
                "username": username,
            },
            Annotations: map[string]string{
                "nginx.ingress.kubernetes.io/proxy-connect-timeout": "3600",
                "nginx.ingress.kubernetes.io/proxy-read-timeout": "3600",
                "nginx.ingress.kubernetes.io/proxy-send-timeout": "3600",
                "nginx.ingress.kubernetes.io/websocket-services": fmt.Sprintf("%s-svc", username),
                "nginx.ingress.kubernetes.io/ssl-redirect": "true",
                "external-dns.alpha.kubernetes.io/hostname": fmt.Sprintf("%s.%s", username, c.domain),
                "cert-manager.io/cluster-issuer": "letsencrypt-prod",
            },
        },
        Spec: networkingv1.IngressSpec{
            IngressClassName: &ingressClassName,
            TLS: []networkingv1.IngressTLS{
                {
                    Hosts: []string{host},
                    SecretName: fmt.Sprintf("%s-tls", username),
                },
            },
            Rules: []networkingv1.IngressRule{
                {
                    Host: host,
                    IngressRuleValue: networkingv1.IngressRuleValue{
                        HTTP: &networkingv1.HTTPIngressRuleValue{
                            Paths: []networkingv1.HTTPIngressPath{
                                {
                                    Path:     "/",
                                    PathType: &pathType,
                                    Backend: networkingv1.IngressBackend{
                                        Service: &networkingv1.IngressServiceBackend{
                                            Name: fmt.Sprintf("%s-svc", username),
                                            Port: networkingv1.ServiceBackendPort{
                                                Number: int32(port),
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    _, err := c.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
    return err
}
```

### 2. Update Storage Class in PVC Creation

Modify `internal/k8s/resources.go` to use the appropriate GCE PD storage class:

```go
func (c *Client) createPVC(username, namespace, size string) error {
    pvcName := fmt.Sprintf("%s-data", username)
    
    storageClassName := "standard-rwo"
    
    pvc := &corev1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{
            Name: pvcName,
            Labels: map[string]string{
                "app":      "k8sgo",
                "username": username,
            },
        },
        Spec: corev1.PersistentVolumeClaimSpec{
            StorageClassName: &storageClassName,
            AccessModes: []corev1.PersistentVolumeAccessMode{
                corev1.ReadWriteOnce,
            },
            Resources: corev1.ResourceRequirements{
                Requests: corev1.ResourceList{
                    corev1.ResourceStorage: resource.MustParse(size),
                },
            },
        },
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    _, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
    return err
}
```

## Monitoring and Logging

### 1. Set up Google Cloud Operations (formerly Stackdriver)

GKE integrates with Google Cloud Operations. Enable it for your cluster:

```bash
gcloud container clusters update ${CLUSTER_NAME} \
  --zone=${ZONE} \
  --enable-stackdriver-kubernetes \
  --project=${PROJECT_ID}
```

### 2. Install Prometheus and Grafana (Optional)

```bash
# Create monitoring namespace
kubectl create namespace monitoring

# Add the Prometheus Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/prometheus \
  --namespace monitoring \
  --set alertmanager.persistentVolume.storageClass=standard-rwo \
  --set server.persistentVolume.storageClass=standard-rwo

# Install Grafana
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm install grafana grafana/grafana \
  --namespace monitoring \
  --set persistence.enabled=true \
  --set persistence.storageClassName=standard-rwo \
  --set persistence.size=10Gi
```

## Security Considerations

### 1. Network Policies

Implement network policies to restrict traffic between namespaces:

```bash
# Apply the Calico CNI plugin for network policy support if not already enabled
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml

cat > network-policy.yaml << EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
spec:
  podSelector: {}
  policyTypes:
  - Ingress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-k8s-orchestrator
spec:
  podSelector:
    matchLabels:
      app: k8s-orchestrator
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 8080
EOF

kubectl apply -f network-policy.yaml
```

### 2. Secrets Management

For production, use Google Secret Manager:

```bash
# Create a service account for Secret Manager
gcloud iam service-accounts create secret-manager-sa \
  --display-name "Secret Manager SA"

# Add the necessary IAM bindings for secret management
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member serviceAccount:secret-manager-sa@${PROJECT_ID}.iam.gserviceaccount.com \
  --role roles/secretmanager.secretAccessor

# Create a Kubernetes service account and bind it to the GCP service account
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[default/k8s-orchestrator-sa]" \
  secret-manager-sa@${PROJECT_ID}.iam.gserviceaccount.com

# Install Secret Manager CSI Driver
kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/secrets-store-csi-driver-provider-gcp/main/deploy/provider-gcp-plugin.yaml
```

## Cost Optimization

### 1. Set up GKE Autopilot (Fully Managed Kubernetes)

Consider using GKE Autopilot to optimize costs and management:

```bash
# Create an Autopilot cluster
gcloud container clusters create-auto ${CLUSTER_NAME}-auto \
  --project=${PROJECT_ID} \
  --region=${REGION} \
  --enable-ip-alias
```

### 2. Configure Horizontal Pod Autoscaler

```bash
cat > hpa.yaml << EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: k8s-orchestrator-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: k8s-orchestrator
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
EOF

kubectl apply -f hpa.yaml
```

## Backup and Disaster Recovery

### 1. Set up Backup for GKE

```bash
# Enable the Backup for GKE API
gcloud services enable gkebackup.googleapis.com

# Create a backup plan
gcloud beta container backup-restore backup-plans create k8s-orchestrator-backup \
  --cluster=${CLUSTER_NAME} \
  --location=${ZONE} \
  --backup-schedule="0 2 * * *" \
  --backup-retain-days=14 \
  --include-secrets \
  --include-volume-data \
  --all-namespaces
```

### 2. Manual PVC Snapshots

```bash
gcloud compute snapshots create ${USERNAME}-data-snapshot \
  --project=${PROJECT_ID} \
  --source-disk=${PVC_DISK_NAME} \
  --zone=${ZONE}
```

## Troubleshooting

1. **Check pod status**:
   ```bash
   kubectl get pods -l app=k8s-orchestrator
   ```

2. **View pod logs**:
   ```bash
   kubectl logs -l app=k8s-orchestrator
   ```

3. **Check service status**:
   ```bash
   kubectl get svc k8s-orchestrator
   ```

4. **Test ingress status**:
   ```bash
   kubectl get ingress k8s-orchestrator
   ```

5. **Check Nginx Ingress logs**:
   ```bash
   kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller
   ```

6. **Access Cloud Logging for detailed logs**:
   ```bash
   gcloud logging read "resource.type=k8s_container AND resource.labels.namespace_name=default AND resource.labels.container_name=k8s-orchestrator"
   ```

## Advanced Configuration

### 1. Regional Cluster for High Availability

For production workloads, consider creating a regional cluster:

```bash
gcloud container clusters create ${CLUSTER_NAME}-regional \
  --project=${PROJECT_ID} \
  --region=${REGION} \
  --num-nodes=2 \
  --node-locations=${REGION}-a,${REGION}-b,${REGION}-c \
  --machine-type=e2-standard-2
```

### 2. VPC-native Cluster for Better Network Performance

```bash
gcloud container clusters create ${CLUSTER_NAME}-vpc-native \
  --project=${PROJECT_ID} \
  --zone=${ZONE} \
  --network=default \
  --enable-ip-alias \
  --create-subnetwork name=gke-subnet,range=10.0.0.0/24 \
  --cluster-ipv4-cidr=10.10.0.0/16 \
  --services-ipv4-cidr=10.11.0.0/16
```

### 3. Private Cluster for Enhanced Security

```bash
gcloud container clusters create ${CLUSTER_NAME}-private \
  --project=${PROJECT_ID} \
  --zone=${ZONE} \
  --enable-private-nodes \
  --enable-private-endpoint \
  --master-ipv4-cidr=172.16.0.0/28 \
  --enable-ip-alias
```