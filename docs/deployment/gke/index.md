# Deploying K8s Orchestrator to Google Kubernetes Engine (GKE)

This guide walks you through the complete process of deploying the K8s Orchestrator service to Google Kubernetes Engine (GKE). The guide is divided into steps to make it easier to follow.

## Architecture Overview

The following diagram shows the high-level architecture of the K8s Orchestrator deployment on GKE:

```mermaid
graph TD
    subgraph "Google Cloud Platform"
        subgraph "Google Kubernetes Engine"
            subgraph "K8s Orchestrator Components"
                API[API Service] --> K8sClient[Kubernetes Client]
                API --> SubdomainManager[Subdomain Manager]
                API --> VolManager[Volume Manager]
            end

            subgraph "User Environment 1"
                Container1[Container\nWeb Server + VNC]
                PVC1[Persistent Volume]
                Container1 -- Mounts --> PVC1
            end

            subgraph "User Environment 2"
                Container2[Container\nWeb Server + VNC]
                PVC2[Persistent Volume]
                Container2 -- Mounts --> PVC2
            end

            K8sClient -- Manages --> Container1
            K8sClient -- Manages --> Container2
            VolManager -- Provisions --> PVC1
            VolManager -- Provisions --> PVC2
            SubdomainManager -- Configures --> TraefikRoute3000[Traefik IngressRoute:3000]
            SubdomainManager -- Configures --> TraefikRoute6901[Traefik IngressRoute:6901]
        end

        DNS[Cloud DNS]
        TraefikRoute3000 -- Routes to --> DNS
        TraefikRoute6901 -- Routes to --> DNS
    end

    DNS -- Resolves --> User1Web[user1.pods.tryiris.dev:3000]
    DNS -- Resolves --> User1VNC[user1.pods.tryiris.dev:6901]
    DNS -- Resolves --> User2Web[user2.pods.tryiris.dev:3000]
    DNS -- Resolves --> User2VNC[user2.pods.tryiris.dev:6901]

    User1[User 1] -- Accesses --> User1Web
    User1 -- Accesses --> User1VNC
    User2[User 2] -- Accesses --> User2Web
    User2 -- Accesses --> User2VNC

    Admin[Administrator] -- Manages via --> API
```

## Deployment Process Flow

```mermaid
sequenceDiagram
    participant User as You
    participant GCloud as gcloud CLI
    participant GKE as GKE Cluster
    participant Helm as Helm
    participant DNS as Cloud DNS

    User->>GCloud: Login & Configure
    User->>GCloud: Create GKE Cluster
    GCloud->>GKE: Provision Cluster
    User->>GCloud: Configure kubectl
    User->>GKE: Install Traefik with Let's Encrypt
    User->>DNS: Configure Domain & Subdomains
    User->>Helm: Customize values-gcp.yaml
    User->>Helm: Deploy Helm Chart
    Helm->>GKE: Create K8s Resources
    User->>GKE: Verify Deployment
    User->>GKE: Test User Environment Creation
```

## Prerequisites

Before you begin, ensure you have the following:

- Google Cloud Platform account with billing enabled
- `gcloud` CLI installed and configured
- `kubectl` installed
- `helm` (v3.2.0+) installed
- Domain name: tryiris.dev with pods.tryiris.dev for user environments

## Deployment Steps

Follow these steps to deploy the K8s Orchestrator to GKE:

1. [Setup Requirements](01-prerequisites.md) - Install and configure required tools
2. [Setup Google Cloud and GKE Cluster](02-setup-gcp-gke.md) - Create and configure your GKE cluster
3. [Setup Core Kubernetes Component](03-core-kubernetes-components.md) - Install Traefik with integrated Let's Encrypt
4. [Configure DNS](04-configure-dns.md) - Set up DNS records for your domain
5. [Deploy with Helm](05-deploy-with-helm.md) - Deploy the K8s Orchestrator using Helm
6. [Verify Deployment](06-verify-deployment.md) - Check that everything is working correctly
7. [Create and Access User Environments](07-user-environments.md) - Create user environments and access them
8. [Scale and Manage Deployment](08-scaling-management.md) - Scale and manage your deployment
9. [Troubleshooting](09-troubleshooting.md) - Common issues and their solutions
10. [Cleanup](10-cleanup.md) - Clean up resources when they're no longer needed