package k8s

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client is a Kubernetes client wrapper
type Client struct {
	clientset *kubernetes.Clientset
	namespace string
	domain    string
}

// SandboxInfo contains information about a sandbox
type SandboxInfo struct {
	UserID    string `json:"userId" example:"user123"`
	Status    string `json:"status" example:"Running"`
	CreatedAt string `json:"createdAt" example:"2023-04-20T12:00:00Z"`
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		home := homedir.HomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Default namespace and domain - should be configurable via environment variables
	namespace := "user-sandboxes"
	domain := "tryiris.dev"

	return &Client{
		clientset: clientset,
		namespace: namespace,
		domain:    domain,
	}, nil
}

// CreateSandbox creates a new sandbox for a user
func (c *Client) CreateSandbox(userID string) error {
	ctx := context.Background()

	// Create namespace if it doesn't exist
	if err := c.ensureNamespace(ctx); err != nil {
		return err
	}

	// Create PVC for user
	if err := c.createPVC(ctx, userID); err != nil {
		return err
	}

	// Create deployment
	if err := c.createDeployment(ctx, userID); err != nil {
		return err
	}

	// Create service
	if err := c.createService(ctx, userID); err != nil {
		return err
	}

	// Create ingress
	if err := c.createIngress(ctx, userID); err != nil {
		return err
	}

	log.Printf("Sandbox created for user: %s", userID)
	return nil
}

// DeleteSandbox deletes a user's sandbox
func (c *Client) DeleteSandbox(userID string) error {
	ctx := context.Background()

	// Delete ingress
	if err := c.clientset.NetworkingV1().Ingresses(c.namespace).Delete(ctx, 
		fmt.Sprintf("%s-ingress", userID), metav1.DeleteOptions{}); err != nil {
		log.Printf("Error deleting ingress: %v", err)
	}

	// Delete service
	if err := c.clientset.CoreV1().Services(c.namespace).Delete(ctx, 
		fmt.Sprintf("%s-service", userID), metav1.DeleteOptions{}); err != nil {
		log.Printf("Error deleting service: %v", err)
	}

	// Delete deployment
	if err := c.clientset.AppsV1().Deployments(c.namespace).Delete(ctx, 
		fmt.Sprintf("%s-deployment", userID), metav1.DeleteOptions{}); err != nil {
		log.Printf("Error deleting deployment: %v", err)
	}

	// Keep PVC for now (user data persistence)
	// Uncomment to delete PVC as well
	/*
	if err := c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).Delete(ctx, 
		fmt.Sprintf("%s-pvc", userID), metav1.DeleteOptions{}); err != nil {
		log.Printf("Error deleting PVC: %v", err)
	}
	*/

	log.Printf("Sandbox deleted for user: %s", userID)
	return nil
}

// ensureNamespace creates the namespace if it doesn't exist
func (c *Client) ensureNamespace(ctx context.Context) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, c.namespace, metav1.GetOptions{})
	if err == nil {
		// Namespace exists
		return nil
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.namespace,
		},
	}
	_, err = c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

// createPVC creates a persistent volume claim for the user
func (c *Client) createPVC(ctx context.Context, userID string) error {
	pvcName := fmt.Sprintf("%s-pvc", userID)
	
	// Check if PVC already exists
	_, err := c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err == nil {
		// PVC exists
		return nil
	}

	storageClassName := "standard-rwo" // Use the default storage class
	// Create PVC
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvcName,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: &storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		},
	}

	_, err = c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

// createDeployment creates a deployment for the user's sandbox
func (c *Client) createDeployment(ctx context.Context, userID string) error {
	deploymentName := fmt.Sprintf("%s-deployment", userID)
	
	// Create deployment
	var replicas int32 = 1
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "user-sandbox",
					"user": userID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "user-sandbox",
						"user": userID,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "sandbox",
							Image: "accetto/ubuntu-vnc-xfce-firefox-g3",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 6901,
									Name:          "vnc",
								},
								{
									ContainerPort: 3000,
									Name:          "http",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "user-data",
									MountPath: "/home/headless/Documents",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("2Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "user-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: fmt.Sprintf("%s-pvc", userID),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().Deployments(c.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

// createService creates a service for the user's sandbox
func (c *Client) createService(ctx context.Context, userID string) error {
	serviceName := fmt.Sprintf("%s-service", userID)
	
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "user-sandbox",
				"user": userID,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "vnc",
					Port:       6901,
					TargetPort: intstr.FromInt(6901),
				},
				{
					Name:       "http",
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
				},
			},
		},
	}

	_, err := c.clientset.CoreV1().Services(c.namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

// createIngress creates an ingress for the user's sandbox
func (c *Client) createIngress(ctx context.Context, userID string) error {
	ingressName := fmt.Sprintf("%s-ingress", userID)

	pathTypePrefix := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: ingressName,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "traefik",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("%s-vnc.%s", userID, c.domain),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: fmt.Sprintf("%s-service", userID),
											Port: networkingv1.ServiceBackendPort{
												Name: "vnc",
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Host: fmt.Sprintf("%s-api.%s", userID, c.domain),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: fmt.Sprintf("%s-service", userID),
											Port: networkingv1.ServiceBackendPort{
												Name: "http",
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

	_, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Create(ctx, ingress, metav1.CreateOptions{})
	return err
}

// ListSandboxes retrieves all sandboxes in the namespace
func (c *Client) ListSandboxes(ctx context.Context) ([]SandboxInfo, error) {
	// Get all deployments in the namespace with the app=user-sandbox label
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=user-sandbox",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	sandboxes := make([]SandboxInfo, 0, len(deployments.Items))
	for _, deployment := range deployments.Items {
		// Extract user ID from deployment labels
		userID := deployment.Labels["user"]
		if userID == "" {
			// Skip deployments without a user ID
			continue
		}

		// Check deployment status
		status := "Unknown"
		if deployment.Status.AvailableReplicas > 0 {
			status = "Running"
		} else if deployment.Status.UnavailableReplicas > 0 {
			status = "Unavailable"
		} else if deployment.Status.ReadyReplicas == 0 {
			status = "Pending"
		}

		// Get creation timestamp
		createdAt := deployment.CreationTimestamp.Format(metav1.RFC3339Micro)

		sandboxes = append(sandboxes, SandboxInfo{
			UserID:    userID,
			Status:    status,
			CreatedAt: createdAt,
		})
	}

	return sandboxes, nil
}