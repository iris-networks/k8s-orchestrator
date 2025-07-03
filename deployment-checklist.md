# Kubernetes Deployment Checklist

This checklist outlines best practices for deploying applications to Kubernetes, covering both pre-deployment and post-deployment stages.

## Pre-Deployment

### Container Image Best Practices

- **Use Minimal Base Images**: Start with a minimal base image (e.g., Alpine, distroless) to reduce attack surface and image size.
- **Don't Run as Root**: Create a non-root user and group in your Dockerfile and use the `USER` instruction.
- **Scan Images for Vulnerabilities**: Integrate image scanning tools (e.g., Trivy, Anchore) into your CI/CD pipeline.
- **Use Specific Image Tags**: Avoid `:latest`. Use immutable tags like Git commit hashes or semantic versions.
- **Optimize Image Layers**: Structure your Dockerfile to leverage layer caching.

### Manifest and Configuration Best Practices

- **GitOps**: Store Kubernetes manifests in a Git repository for version control and collaboration.
- **Declarative Manifests**: Use `kubectl apply` with YAML files.
- **Validate Manifests**: Use tools like `kubeval`, `kube-score`, or `KubeLinter` to validate manifests before applying them.
- **Labels and Annotations**: Organize resources effectively.
- **Separate Configuration**: Use ConfigMaps and Secrets to externalize configuration.
- **Templating**: Use Helm or Kustomize for managing complex applications.

### Security Best Practices

- **RBAC**: Implement Role-Based Access Control with the principle of least privilege.
- **Network Policies**: Restrict traffic between pods and services.
- **Security Contexts**: Define privilege and access control settings for Pods and Containers.
- **Secrets Management**: Mount secrets as volumes, not environment variables. Consider using a secrets management tool.
- **Pod Security Admission**: Enforce Pod Security Standards.

### Resource Management Best Practices

- **Resource Requests and Limits**: Specify CPU and memory requests and limits for all containers.
- **Quality of Service (QoS)**: Understand QoS classes (Guaranteed, Burstable, BestEffort).
- **Resource Quotas**: Limit resource consumption per namespace.
- **Probes**: Implement liveness and readiness probes.

## Post-Deployment

### Verification and Health Checks

- **Check Deployment Status**: Use `kubectl get deployments` to verify replica status.
- **Monitor Rolling Updates**: Keep an eye on rolling updates for smooth transitions.
- **Test in Staging**: Thoroughly test all changes in a staging environment before production deployment.

### High Availability and Scalability

- **Multiple Replicas**: Run multiple instances of each component.
- **Pod Anti-Affinity**: Prevent pods of the same application from being scheduled on the same node.
- **Horizontal Pod Autoscaler (HPA)**: Automatically scale pods based on resource utilization.
- **Pod Disruption Budgets (PDBs)**: Prevent loss of all replicas during voluntary disruptions.
- **Topology Spread Constraints**: Distribute pods across failure domains.

### Monitoring, Logging, and Observability

- **Monitoring Pipeline**: Set up a monitoring pipeline (e.g., Prometheus, Grafana).
- **Logging**: Configure applications to write logs to `stdout` and `stderr`.
- **Alerting**: Set up alerts for critical events and resource usage.

### Deployment Strategy and Automation

- **Declarative Configuration**: Use declarative YAML files stored in Git.
- **GitOps**: Adopt a GitOps approach for deployments.
- **Deployment Strategies**: Choose the right deployment strategy (e.g., Rolling, Blue/Green, Canary).
