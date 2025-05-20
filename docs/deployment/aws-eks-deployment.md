# Deploying to AWS EKS

This guide provides step-by-step instructions for deploying the K8s Orchestrator service on Amazon EKS (Elastic Kubernetes Service).

## Prerequisites

- AWS CLI installed and configured with appropriate permissions
- `kubectl` installed and configured
- `eksctl` installed
- `helm` installed (v3+)
- Docker installed (for building and pushing images)
- An active AWS account with appropriate permissions

## Infrastructure Setup

### 1. Create an EKS Cluster

If you don't already have an EKS cluster, you can create one using `eksctl`:

```bash
eksctl create cluster \
  --name k8s-orchestrator-cluster \
  --region us-west-2 \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 2 \
  --nodes-max 5 \
  --with-oidc \
  --managed
```

### 2. Configure IAM Permissions

Create an IAM role for the K8s Orchestrator service account:

```bash
eksctl create iamserviceaccount \
  --cluster=k8s-orchestrator-cluster \
  --namespace=default \
  --name=k8s-orchestrator-sa \
  --attach-policy-arn=arn:aws:iam::aws:policy/AmazonEKSClusterPolicy \
  --approve
```

### 3. Install Required Components

#### a. AWS Load Balancer Controller

```bash
helm repo add eks https://aws.github.io/eks-charts
helm repo update

eksctl utils associate-iam-oidc-provider \
    --region us-west-2 \
    --cluster k8s-orchestrator-cluster \
    --approve

# Download IAM policy
curl -o iam-policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/main/docs/install/iam_policy.json

# Create policy
aws iam create-policy \
    --policy-name AWSLoadBalancerControllerIAMPolicy \
    --policy-document file://iam-policy.json

# Create service account
eksctl create iamserviceaccount \
  --cluster=k8s-orchestrator-cluster \
  --namespace=kube-system \
  --name=aws-load-balancer-controller \
  --attach-policy-arn=arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):policy/AWSLoadBalancerControllerIAMPolicy \
  --approve

# Install controller
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=k8s-orchestrator-cluster \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

#### b. External DNS (for subdomain management)

```bash
# Create IAM policy for ExternalDNS
cat > external-dns-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets"
      ],
      "Resource": [
        "arn:aws:route53:::hostedzone/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "route53:ListHostedZones",
        "route53:ListResourceRecordSets"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

# Create policy
aws iam create-policy \
    --policy-name ExternalDNSPolicy \
    --policy-document file://external-dns-policy.json

# Create service account
eksctl create iamserviceaccount \
  --cluster=k8s-orchestrator-cluster \
  --namespace=kube-system \
  --name=external-dns \
  --attach-policy-arn=arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):policy/ExternalDNSPolicy \
  --approve

# Install ExternalDNS
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install external-dns bitnami/external-dns \
  --namespace kube-system \
  --set provider=aws \
  --set aws.zoneType=public \
  --set domainFilters[0]=your-domain.com \
  --set serviceAccount.create=false \
  --set serviceAccount.name=external-dns
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
          class: alb
EOF

kubectl apply -f letsencrypt-prod.yaml
```

### 4. Configure Storage Classes

EBS CSI driver is required for dynamic volume provisioning:

```bash
eksctl create iamserviceaccount \
  --name ebs-csi-controller-sa \
  --namespace kube-system \
  --cluster k8s-orchestrator-cluster \
  --attach-policy-arn arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy \
  --approve \
  --role-only \
  --role-name AmazonEKS_EBS_CSI_DriverRole

eksctl create addon \
  --name aws-ebs-csi-driver \
  --cluster k8s-orchestrator-cluster \
  --service-account-role-arn arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/AmazonEKS_EBS_CSI_DriverRole \
  --force
```

## Application Deployment

### 1. Configure DNS in Route53

Create a hosted zone in Route53 if you don't already have one:

```bash
aws route53 create-hosted-zone \
  --name your-domain.com \
  --caller-reference $(date +%s)
```

Update your domain's name servers at your registrar with the NS records provided.

### 2. Build and Push the Docker Image

Create an ECR repository and push the image:

```bash
# Create ECR repository
aws ecr create-repository --repository-name k8s-orchestrator

# Log in to ECR
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin $(aws sts get-caller-identity --query Account --output text).dkr.ecr.us-west-2.amazonaws.com

# Build and tag the image
docker build -t $(aws sts get-caller-identity --query Account --output text).dkr.ecr.us-west-2.amazonaws.com/k8s-orchestrator:latest .

# Push the image
docker push $(aws sts get-caller-identity --query Account --output text).dkr.ecr.us-west-2.amazonaws.com/k8s-orchestrator:latest
```

### 3. Deploy the Application

Create a deployment YAML file:

```bash
cat > k8s-orchestrator-aws.yaml << EOF
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
        image: $(aws sts get-caller-identity --query Account --output text).dkr.ecr.us-west-2.amazonaws.com/k8s-orchestrator:latest
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
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}, {"HTTPS": 443}]'
    alb.ingress.kubernetes.io/ssl-redirect: '443'
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
kubectl apply -f k8s-orchestrator-aws.yaml
```

## Code Modifications

### 1. Update the Ingress Creation Logic

Modify `internal/k8s/resources.go` to use the appropriate annotations for AWS ALB Ingress Controller:

```go
func (c *Client) createIngress(username, namespace string, port int) error {
    ingressName := fmt.Sprintf("%s-ingress", username)
    host := fmt.Sprintf("%s.%s", username, c.domain)
    pathType := networkingv1.PathTypePrefix
    
    ingressClassName := "alb"
    
    ingress := &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name: ingressName,
            Labels: map[string]string{
                "app":      "k8sgo",
                "username": username,
            },
            Annotations: map[string]string{
                "alb.ingress.kubernetes.io/scheme": "internet-facing",
                "alb.ingress.kubernetes.io/target-type": "ip",
                "alb.ingress.kubernetes.io/listen-ports": "[{\"HTTP\": 80}, {\"HTTPS\": 443}]",
                "alb.ingress.kubernetes.io/ssl-redirect": "443",
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

Modify `internal/k8s/resources.go` to use the appropriate EBS storage class:

```go
func (c *Client) createPVC(username, namespace, size string) error {
    pvcName := fmt.Sprintf("%s-data", username)
    
    storageClassName := "gp2"
    
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

### 1. Install Prometheus and Grafana

```bash
# Add the Prometheus Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/prometheus \
  --namespace monitoring \
  --create-namespace

# Install Grafana
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm install grafana grafana/grafana \
  --namespace monitoring \
  --set persistence.enabled=true \
  --set persistence.size=10Gi
```

### 2. Install AWS CloudWatch Agent

```bash
# Create namespace
kubectl create namespace amazon-cloudwatch

# Create ConfigMap
kubectl create configmap cluster-info \
  --from-literal=cluster.name=k8s-orchestrator-cluster \
  --from-literal=logs.region=us-west-2 -n amazon-cloudwatch

# Apply CloudWatch Agent
kubectl apply -f https://raw.githubusercontent.com/aws-samples/amazon-cloudwatch-container-insights/latest/k8s-deployment-manifest-templates/deployment-mode/daemonset/container-insights-monitoring/quickstart/cwagent-fluentd-quickstart.yaml
```

## Security Considerations

1. **Network Policies**: Implement network policies to restrict traffic between namespaces:

```bash
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

2. **Secrets Management**: For production, use AWS Secrets Manager or AWS Parameter Store:

```bash
# Install Secrets Store CSI Driver
helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm repo update
helm install csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver \
  --namespace kube-system \
  --set syncSecret.enabled=true

# Install AWS Provider
kubectl apply -f https://raw.githubusercontent.com/aws/secrets-store-csi-driver-provider-aws/main/deployment/aws-provider-installer.yaml
```

## Cost Optimization

1. Scale down resources during low-usage periods:

```bash
# Install Keda for event-driven autoscaling
helm repo add kedacore https://kedacore.github.io/charts
helm repo update
helm install keda kedacore/keda --namespace keda --create-namespace
```

2. Set up AWS Spot instances for worker nodes:

```bash
eksctl create nodegroup \
  --cluster k8s-orchestrator-cluster \
  --name k8s-orchestrator-spot \
  --node-type m5.large \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 5 \
  --capacity-type SPOT
```

## Backup and Disaster Recovery

Set up regular backups of PVCs using Velero:

```bash
# Install Velero with restic
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.5.0 \
  --bucket velero-eks-backup \
  --backup-location-config region=us-west-2 \
  --snapshot-location-config region=us-west-2 \
  --use-restic \
  --secret-file ./credentials-velero
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

5. **Check ALB controller logs**:
   ```bash
   kubectl logs -n kube-system -l app.kubernetes.io/name=aws-load-balancer-controller
   ```