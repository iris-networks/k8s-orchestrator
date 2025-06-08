package k8s

import (
	"context"
	"fmt"
	"log"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ContainerStatus contains detailed information about a container's status
type ContainerStatus struct {
	Name         string `json:"name" example:"sandbox"`
	Ready        bool   `json:"ready" example:"true"`
	State        string `json:"state" example:"running"`
	RestartCount int32  `json:"restartCount" example:"0"`
	Image        string `json:"image" example:"us-central1-docker.pkg.dev/driven-seer-460401-p9/iris-repo/iris_agent:latest"`
	Message      string `json:"message,omitempty" example:""`
	Reason       string `json:"reason,omitempty" example:""`
}

// SandboxInfo contains information about a sandbox
type SandboxInfo struct {
	UserID           string            `json:"userId" example:"user123"`
	Status           string            `json:"status" example:"Running"`
	CreatedAt        string            `json:"createdAt" example:"2023-04-20T12:00:00Z"`
	PodName          string            `json:"podName,omitempty" example:"user123-deployment-5d8b9c7b8f-2p8x7"`
	PodPhase         string            `json:"podPhase,omitempty" example:"Running"`
	PodConditions    []string          `json:"podConditions,omitempty" example:"[\"PodScheduled\",\"Initialized\",\"ContainersReady\",\"Ready\"]"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty"`
	InitContainerStatuses []ContainerStatus `json:"initContainerStatuses,omitempty"`
	Message          string            `json:"message,omitempty" example:""`
	Reason           string            `json:"reason,omitempty" example:""`
}

// CreateSandbox creates a new sandbox for a user
func (c *Client) CreateSandbox(userID string) error {
	// Previous parameters for environment variables have been removed
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

	// No longer deleting Node.js environment ConfigMap as it's no longer created

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

// ListSandboxes retrieves all sandboxes in the namespace
func (c *Client) ListSandboxes(ctx context.Context) ([]SandboxInfo, error) {
	// Get all deployments in the namespace without filtering by label
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	sandboxes := make([]SandboxInfo, 0, len(deployments.Items))
	for _, deployment := range deployments.Items {
		// For each deployment, try to extract a user ID
		// First try from the "user" label
		userID := deployment.Labels["user"]

		// If not found, check if the deployment name follows the pattern "{userId}-deployment"
		if userID == "" {
			// Try to extract from deployment name (format: {userId}-deployment)
			name := deployment.Name
			if strings.HasSuffix(name, "-deployment") {
				userID = strings.TrimSuffix(name, "-deployment")
			}
		}

		// If we still couldn't determine the user ID, skip this deployment
		if userID == "" {
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

// GetSandboxStatus retrieves the status of a specific sandbox by user ID
func (c *Client) GetSandboxStatus(ctx context.Context, userID string) (*SandboxInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Try to get the deployment for this user
	deploymentName := fmt.Sprintf("%s-deployment", userID)
	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("sandbox not found for user ID %s: %w", userID, err)
	}

	// Get creation timestamp
	createdAt := deployment.CreationTimestamp.Format(metav1.RFC3339Micro)

	// Create base sandbox info
	sandboxInfo := &SandboxInfo{
		UserID:    userID,
		CreatedAt: createdAt,
	}

	// Check deployment status
	if deployment.Status.AvailableReplicas > 0 {
		sandboxInfo.Status = "Running"
	} else if deployment.Status.UnavailableReplicas > 0 {
		sandboxInfo.Status = "Unavailable"
	} else if deployment.Status.ReadyReplicas == 0 {
		sandboxInfo.Status = "Pending"
	} else {
		sandboxInfo.Status = "Unknown"
	}

	// Get the pods associated with this deployment
	labelSelector := fmt.Sprintf("app=user-sandbox,user=%s", userID)
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		// If we can't get pods, just return the basic info
		return sandboxInfo, nil
	}

	// If no pods found, return basic info
	if len(pods.Items) == 0 {
		sandboxInfo.Message = "No pods found for this deployment"
		return sandboxInfo, nil
	}

	// Get the newest pod (most likely to be the active one)
	var newestPod *corev1.Pod
	for i := range pods.Items {
		if newestPod == nil || pods.Items[i].CreationTimestamp.After(newestPod.CreationTimestamp.Time) {
			newestPod = &pods.Items[i]
		}
	}

	if newestPod != nil {
		// Add pod details
		sandboxInfo.PodName = newestPod.Name
		sandboxInfo.PodPhase = string(newestPod.Status.Phase)

		// Add pod conditions
		podConditions := []string{}
		for _, condition := range newestPod.Status.Conditions {
			if condition.Status == "True" {
				podConditions = append(podConditions, string(condition.Type))
			}
		}
		sandboxInfo.PodConditions = podConditions

		// Add message and reason if present
		if newestPod.Status.Message != "" {
			sandboxInfo.Message = newestPod.Status.Message
		}
		if newestPod.Status.Reason != "" {
			sandboxInfo.Reason = newestPod.Status.Reason
		}

		// Process container statuses
		containerStatuses := []ContainerStatus{}
		for _, cs := range newestPod.Status.ContainerStatuses {
			state := "unknown"
			message := ""
			reason := ""

			if cs.State.Running != nil {
				state = "running"
			} else if cs.State.Waiting != nil {
				state = "waiting"
				message = cs.State.Waiting.Message
				reason = cs.State.Waiting.Reason
			} else if cs.State.Terminated != nil {
				state = "terminated"
				message = cs.State.Terminated.Message
				reason = cs.State.Terminated.Reason
			}

			containerStatuses = append(containerStatuses, ContainerStatus{
				Name:         cs.Name,
				Ready:        cs.Ready,
				State:        state,
				RestartCount: cs.RestartCount,
				Image:        cs.Image,
				Message:      message,
				Reason:       reason,
			})
		}
		sandboxInfo.ContainerStatuses = containerStatuses

		// Process init container statuses
		initContainerStatuses := []ContainerStatus{}
		for _, cs := range newestPod.Status.InitContainerStatuses {
			state := "unknown"
			message := ""
			reason := ""

			if cs.State.Running != nil {
				state = "running"
			} else if cs.State.Waiting != nil {
				state = "waiting"
				message = cs.State.Waiting.Message
				reason = cs.State.Waiting.Reason
			} else if cs.State.Terminated != nil {
				state = "terminated"
				message = cs.State.Terminated.Message
				reason = cs.State.Terminated.Reason
				if cs.State.Terminated.ExitCode == 0 {
					reason = "Completed"
				}
			}

			initContainerStatuses = append(initContainerStatuses, ContainerStatus{
				Name:         cs.Name,
				Ready:        cs.Ready,
				State:        state,
				RestartCount: cs.RestartCount,
				Image:        cs.Image,
				Message:      message,
				Reason:       reason,
			})
		}
		sandboxInfo.InitContainerStatuses = initContainerStatuses

		// Update overall status based on more detailed pod information
		if newestPod.Status.Phase == "Pending" {
			// Check if we're waiting on image pull
			for _, cs := range newestPod.Status.ContainerStatuses {
				if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ImagePullBackOff" {
					sandboxInfo.Status = "ImagePullBackOff"
					break
				} else if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ErrImagePull" {
					sandboxInfo.Status = "ErrImagePull"
					break
				} else if cs.State.Waiting != nil && cs.State.Waiting.Reason == "PodInitializing" {
					sandboxInfo.Status = "Initializing"
					break
				} else if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ContainerCreating" {
					sandboxInfo.Status = "ContainerCreating"
					break
				}
			}

			// Check init containers
			for _, cs := range newestPod.Status.InitContainerStatuses {
				if cs.State.Waiting != nil {
					sandboxInfo.Status = "InitContainerWaiting"
					break
				} else if cs.State.Running != nil {
					sandboxInfo.Status = "InitContainerRunning"
					break
				}
			}
		} else if newestPod.Status.Phase == "Running" {
			// Check if all containers are ready
			allReady := true
			for _, cs := range newestPod.Status.ContainerStatuses {
				if !cs.Ready {
					allReady = false
					break
				}
			}

			if !allReady {
				sandboxInfo.Status = "NotAllContainersReady"
			}
		}
	}

	return sandboxInfo, nil
}