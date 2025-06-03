package k8s

import (
	"context"
	"fmt"
	"log"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SandboxInfo contains information about a sandbox
type SandboxInfo struct {
	UserID    string `json:"userId" example:"user123"`
	Status    string `json:"status" example:"Running"`
	CreatedAt string `json:"createdAt" example:"2023-04-20T12:00:00Z"`
}

// CreateSandbox creates a new sandbox for a user
func (c *Client) CreateSandbox(userID string, envVars map[string]string, nodeEnvVars map[string]string) error {
	ctx := context.Background()

	// Create namespace if it doesn't exist
	if err := c.ensureNamespace(ctx); err != nil {
		return err
	}

	// Create PVC for user
	if err := c.createPVC(ctx, userID); err != nil {
		return err
	}

	// Create ConfigMap for Node.js environment variables if provided
	if len(nodeEnvVars) > 0 {
		if err := c.createNodeEnvConfigMap(ctx, userID, nodeEnvVars); err != nil {
			return err
		}
	}

	// Create deployment with environment variables
	if err := c.createDeployment(ctx, userID, envVars, len(nodeEnvVars) > 0); err != nil {
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

	// Delete Node.js environment ConfigMap if it exists
	if err := c.clientset.CoreV1().ConfigMaps(c.namespace).Delete(ctx,
		fmt.Sprintf("%s-node-env", userID), metav1.DeleteOptions{}); err != nil {
		log.Printf("Error deleting Node.js env ConfigMap: %v", err)
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

	return &SandboxInfo{
		UserID:    userID,
		Status:    status,
		CreatedAt: createdAt,
	}, nil
}