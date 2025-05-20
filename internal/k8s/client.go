package k8s

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/k8sgo/internal/models"
	"github.com/k8sgo/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// Client represents a Kubernetes client
type Client struct {
	clientset *kubernetes.Clientset
	namespace string
	domain    string
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	// Get Kubernetes clientset
	clientset, err := utils.GetKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %v", err)
	}

	return &Client{
		clientset: clientset,
		namespace: "default",
		domain:    utils.GetDomainFromEnv(),
	}, nil
}

// CreateEnvironment creates a new user environment
func (c *Client) CreateEnvironment(req models.EnvironmentRequest) (*models.Environment, error) {
	// Set default values if not provided
	if req.Image == "" {
		req.Image = "accetto/ubuntu-vnc-xfce-firefox-g3"
	}

	if len(req.Ports) == 0 {
		req.Ports = []int{5901, 6901}
	}

	if req.VolumeSize == "" {
		req.VolumeSize = "1Gi"
	}

	log.Printf("Creating environment for user: %s", req.Username)

	// 1. Create namespace
	namespace, err := c.createNamespace(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace: %v", err)
	}

	// 2. Create PVC
	err = c.createPVC(req.Username, namespace, req.VolumeSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC: %v", err)
	}

	// 3. Create deployment
	err = c.createDeployment(req.Username, namespace, req.Image, req.Ports, req.VolumeMounts, req.Env)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %v", err)
	}

	// 4. Create service
	err = c.createService(req.Username, namespace, req.Ports)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %v", err)
	}

	// 5. Create ingress
	err = c.createIngress(req.Username, namespace, req.Ports[0])
	if err != nil {
		return nil, fmt.Errorf("failed to create ingress: %v", err)
	}

	// Return the environment details
	return &models.Environment{
		Username:     req.Username,
		Namespace:    namespace,
		Image:        req.Image,
		Ports:        req.Ports,
		VolumeSize:   req.VolumeSize,
		VolumeMounts: req.VolumeMounts,
		Env:          req.Env,
		Status:       "Running",
		URL:          fmt.Sprintf("http://%s.%s", req.Username, c.domain),
		CreatedAt:    time.Now().Format(time.RFC3339),
	}, nil
}

// ListEnvironments lists all user environments
func (c *Client) ListEnvironments() ([]models.Environment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List all namespaces with our app label
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "app=k8sgo",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	environments := make([]models.Environment, 0, len(namespaces.Items))

	// For each namespace, gather environment details
	for _, ns := range namespaces.Items {
		username := ns.Labels["username"]
		if username == "" {
			continue // Skip if no username label
		}

		// Get the deployment
		deployment, err := c.clientset.AppsV1().Deployments(ns.Name).Get(ctx, fmt.Sprintf("%s-desktop", username), metav1.GetOptions{})
		if err != nil {
			log.Printf("Warning: failed to get deployment for %s: %v", username, err)
			continue
		}

		// Get container details
		var image string
		var ports []int
		var env map[string]string
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			container := deployment.Spec.Template.Spec.Containers[0]
			image = container.Image

			ports = make([]int, len(container.Ports))
			for i, port := range container.Ports {
				ports[i] = int(port.ContainerPort)
			}

			env = make(map[string]string)
			for _, envVar := range container.Env {
				env[envVar.Name] = envVar.Value
			}
		}

		// Get PVC details
		pvc, err := c.clientset.CoreV1().PersistentVolumeClaims(ns.Name).Get(ctx, fmt.Sprintf("%s-data", username), metav1.GetOptions{})
		var volumeSize string
		if err == nil {
			quantity := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			volumeSize = quantity.String()
		} else {
			volumeSize = "unknown"
		}

		// Create environment
		environment := models.Environment{
			Username:   username,
			Namespace:  ns.Name,
			Image:      image,
			Ports:      ports,
			VolumeSize: volumeSize,
			Env:        env,
			Status:     string(deployment.Status.Conditions[0].Type),
			URL:        fmt.Sprintf("http://%s.%s", username, c.domain),
			CreatedAt:  ns.CreationTimestamp.Format(time.RFC3339),
		}

		environments = append(environments, environment)
	}

	return environments, nil
}

// GetEnvironment gets a specific user environment
func (c *Client) GetEnvironment(username string) (*models.Environment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get namespace
	namespaceName := fmt.Sprintf("user-%s", username)
	ns, err := c.clientset.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %v", err)
	}

	// Get deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespaceName).Get(ctx, fmt.Sprintf("%s-desktop", username), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("deployment not found: %v", err)
	}

	// Get container details
	var image string
	var ports []int
	var env map[string]string
	var volumeMounts []models.VolumeMount

	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := deployment.Spec.Template.Spec.Containers[0]
		image = container.Image

		ports = make([]int, len(container.Ports))
		for i, port := range container.Ports {
			ports[i] = int(port.ContainerPort)
		}

		env = make(map[string]string)
		for _, envVar := range container.Env {
			env[envVar.Name] = envVar.Value
		}

		volumeMounts = make([]models.VolumeMount, len(container.VolumeMounts))
		for i, vm := range container.VolumeMounts {
			volumeMounts[i] = models.VolumeMount{
				Name:      vm.Name,
				MountPath: vm.MountPath,
			}
		}
	}

	// Get PVC details
	pvc, err := c.clientset.CoreV1().PersistentVolumeClaims(namespaceName).Get(ctx, fmt.Sprintf("%s-data", username), metav1.GetOptions{})
	var volumeSize string
	if err == nil {
		quantity := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		volumeSize = quantity.String()
	} else {
		volumeSize = "unknown"
	}

	// Create environment
	environment := &models.Environment{
		Username:     username,
		Namespace:    namespaceName,
		Image:        image,
		Ports:        ports,
		VolumeSize:   volumeSize,
		VolumeMounts: volumeMounts,
		Env:          env,
		Status:       string(deployment.Status.Conditions[0].Type),
		URL:          fmt.Sprintf("http://%s.%s", username, c.domain),
		CreatedAt:    ns.CreationTimestamp.Format(time.RFC3339),
	}

	return environment, nil
}

// DeleteEnvironment deletes a user environment
func (c *Client) DeleteEnvironment(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	namespaceName := fmt.Sprintf("user-%s", username)

	// Delete namespace (this will cascade delete all resources in it)
	err := c.clientset.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %v", err)
	}

	return nil
}

// UpdateEnvironment updates a user environment
func (c *Client) UpdateEnvironment(req models.EnvironmentRequest) (*models.Environment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespaceName := fmt.Sprintf("user-%s", req.Username)

	// Check if namespace exists
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %v", err)
	}

	deploymentName := fmt.Sprintf("%s-desktop", req.Username)

	// Update deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespaceName).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("deployment not found: %v", err)
	}

	// Update container properties if provided
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := &deployment.Spec.Template.Spec.Containers[0]

		// Update image if provided
		if req.Image != "" {
			container.Image = req.Image
		}

		// Update ports if provided
		if len(req.Ports) > 0 {
			containerPorts := make([]corev1.ContainerPort, len(req.Ports))
			for i, port := range req.Ports {
				containerPorts[i] = corev1.ContainerPort{
					ContainerPort: int32(port),
					Protocol:      corev1.ProtocolTCP,
				}
			}
			container.Ports = containerPorts
		}

		// Update environment variables if provided
		if len(req.Env) > 0 {
			env := make([]corev1.EnvVar, 0, len(req.Env))
			for key, value := range req.Env {
				env = append(env, corev1.EnvVar{
					Name:  key,
					Value: value,
				})
			}
			container.Env = env
		}

		// Update volume mounts if provided
		if len(req.VolumeMounts) > 0 {
			volumeMounts := make([]corev1.VolumeMount, len(req.VolumeMounts))
			for i, vm := range req.VolumeMounts {
				volumeMounts[i] = corev1.VolumeMount{
					Name:      vm.Name,
					MountPath: vm.MountPath,
				}
			}
			container.VolumeMounts = volumeMounts
		}
	}

	// Update the deployment
	_, err = c.clientset.AppsV1().Deployments(namespaceName).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment: %v", err)
	}

	// Update service if ports changed
	if len(req.Ports) > 0 {
		serviceName := fmt.Sprintf("%s-svc", req.Username)
		service, err := c.clientset.CoreV1().Services(namespaceName).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("service not found: %v", err)
		}

		// Update service ports
		servicePorts := make([]corev1.ServicePort, len(req.Ports))
		for i, port := range req.Ports {
			servicePorts[i] = corev1.ServicePort{
				Name:       fmt.Sprintf("port-%d", port),
				Port:       int32(port),
				TargetPort: intstr.FromInt(port),
				Protocol:   corev1.ProtocolTCP,
			}
		}
		service.Spec.Ports = servicePorts

		_, err = c.clientset.CoreV1().Services(namespaceName).Update(ctx, service, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to update service: %v", err)
		}
	}

	// Update PVC size if provided (note: this is more complex and may not be supported by all storage classes)
	if req.VolumeSize != "" {
		// This is a placeholder for PVC resize which typically requires storage class that supports volume expansion
		// In many cases, this requires creating a new PVC and migrating data which is beyond this example
		log.Printf("Warning: PVC resize requested but not implemented")
	}

	// Return the updated environment
	return c.GetEnvironment(req.Username)
}