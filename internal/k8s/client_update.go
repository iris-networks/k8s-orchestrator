package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Update the Client struct to include a dynamic client for Traefik CRDs
type ClientWithTraefik struct {
	Client
	dynamicClient dynamic.Interface
}

// NewClientWithTraefik creates a new Kubernetes client with Traefik CRD support
func NewClientWithTraefik() (*ClientWithTraefik, error) {
	// Create the base client
	baseClient, err := NewClient()
	if err != nil {
		return nil, err
	}

	// Get the config for dynamic client
	var config *rest.Config
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, err
		}
	}

	// Create the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ClientWithTraefik{
		Client:        *baseClient,
		dynamicClient: dynamicClient,
	}, nil
}

// createIngressRoute creates a Traefik IngressRoute CRD for the user's sandbox
func (c *ClientWithTraefik) createIngressRoute(ctx context.Context, userID string) error {
	// Create the VNC IngressRoute
	if err := c.createVncIngressRoute(ctx, userID); err != nil {
		return err
	}

	// Create the API IngressRoute
	if err := c.createApiIngressRoute(ctx, userID); err != nil {
		return err
	}

	return nil
}

// createVncIngressRoute creates the VNC IngressRoute for the user
func (c *ClientWithTraefik) createVncIngressRoute(ctx context.Context, userID string) error {
	// Get the IngressRoute GVR
	gvr := IngressRouteGVR()

	// Define the VNC IngressRoute
	ingressRoute := &IngressRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "traefik.io/v1alpha1",
			Kind:       "IngressRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-vnc", userID),
			Namespace: c.namespace,
			// External-DNS annotations have been removed
			Annotations: map[string]string{},
		},
		Spec: IngressRouteSpec{
			EntryPoints: []string{"websecure"},
			Routes: []Route{
				{
					Match: fmt.Sprintf("Host(`%s-vnc.%s`)", userID, c.domain),
					Kind:  "Rule",
					Services: []Service{
						{
							Name: fmt.Sprintf("%s-service", userID),
							Port: 6901,
						},
					},
				},
			},
			TLS: &TLS{
				CertResolver: "letsencrypt",
			},
		},
	}

	// Convert to unstructured for the dynamic client
	unstructuredObj, err := convertToUnstructured(ingressRoute)
	if err != nil {
		return err
	}

	// Create the IngressRoute
	_, err = c.dynamicClient.Resource(gvr).Namespace(c.namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	return err
}

// createApiIngressRoute creates the API IngressRoute for the user
func (c *ClientWithTraefik) createApiIngressRoute(ctx context.Context, userID string) error {
	// Get the IngressRoute GVR
	gvr := IngressRouteGVR()

	// Define the API IngressRoute
	ingressRoute := &IngressRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "traefik.io/v1alpha1",
			Kind:       "IngressRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-api", userID),
			Namespace: c.namespace,
			// External-DNS annotations have been removed
			Annotations: map[string]string{},
		},
		Spec: IngressRouteSpec{
			EntryPoints: []string{"websecure"},
			Routes: []Route{
				{
					Match: fmt.Sprintf("Host(`%s-api.%s`)", userID, c.domain),
					Kind:  "Rule",
					Services: []Service{
						{
							Name: fmt.Sprintf("%s-service", userID),
							Port: 3000,
						},
					},
				},
			},
			TLS: &TLS{
				CertResolver: "letsencrypt",
			},
		},
	}

	// Convert to unstructured for the dynamic client
	unstructuredObj, err := convertToUnstructured(ingressRoute)
	if err != nil {
		return err
	}

	// Create the IngressRoute
	_, err = c.dynamicClient.Resource(gvr).Namespace(c.namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	return err
}

// deleteIngressRoutes deletes the IngressRoutes for a user
func (c *ClientWithTraefik) deleteIngressRoutes(ctx context.Context, userID string) error {
	// Get the IngressRoute GVR
	gvr := IngressRouteGVR()

	// Delete the VNC IngressRoute
	err := c.dynamicClient.Resource(gvr).Namespace(c.namespace).Delete(ctx, fmt.Sprintf("%s-vnc", userID), metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error deleting VNC IngressRoute: %v", err)
	}

	// Delete the API IngressRoute
	err = c.dynamicClient.Resource(gvr).Namespace(c.namespace).Delete(ctx, fmt.Sprintf("%s-api", userID), metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Error deleting API IngressRoute: %v", err)
	}

	return nil
}

// IsValidKubernetesName validates if a name conforms to Kubernetes service naming rules
// This follows DNS-1035 label naming convention used by Kubernetes for services
// Valid: lowercase alphanumeric characters, '-', must start with a letter, end with alphanumeric
// Invalid: uppercase, '_', other special chars, start with number, start/end with '-'
func IsValidKubernetesName(name string) (bool, string) {
	// Check if the name is empty
	if name == "" {
		return false, "Name cannot be empty"
	}

	// Check maximum length (63 chars per DNS label)
	if len(name) > 63 {
		return false, "Name must be 63 characters or less"
	}

	// DNS-1035 label validation regex pattern (for service names)
	// Using the exact regex from Kubernetes: '[a-z]([-a-z0-9]*[a-z0-9])?'
	pattern := regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)

	// Check against DNS-1035 pattern
	if !pattern.MatchString(name) {
		return false, "Name must consist of lower case alphanumeric characters or '-', " +
			"start with an alphabetic character, and end with an alphanumeric character"
	}

	return true, ""
}

// CreateSandbox creates a new sandbox for a user with Traefik IngressRoutes
func (c *ClientWithTraefik) CreateSandbox(userID string, envVars map[string]string, nodeEnvVars map[string]string) error {
	ctx := context.Background()

	// Validate service name first
	valid, errMsg := IsValidKubernetesName(userID)
	if !valid {
		return fmt.Errorf("invalid user ID for Kubernetes service: %s", errMsg)
	}

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

	// Create Traefik IngressRoutes
	if err := c.createIngressRoute(ctx, userID); err != nil {
		return err
	}

	log.Printf("Sandbox created for user: %s", userID)
	return nil
}

// DeleteSandbox deletes a user's sandbox
func (c *ClientWithTraefik) DeleteSandbox(userID string) error {
	ctx := context.Background()

	// Delete Traefik IngressRoutes
	if err := c.deleteIngressRoutes(ctx, userID); err != nil {
		log.Printf("Error deleting IngressRoutes: %v", err)
	}

	// Try to delete service
	if err := c.clientset.CoreV1().Services(c.namespace).Delete(ctx,
		fmt.Sprintf("%s-service", userID), metav1.DeleteOptions{}); err != nil {
		log.Printf("Error deleting service: %v", err)
	}

	// Try possible deployment name patterns
	deploymentPatterns := []string{
		fmt.Sprintf("%s-deployment", userID),
	}

	deploymentDeleted := false
	for _, depName := range deploymentPatterns {
		log.Printf("Trying to delete deployment: %s", depName)
		err := c.clientset.AppsV1().Deployments(c.namespace).Delete(ctx, depName, metav1.DeleteOptions{})
		if err == nil {
			log.Printf("Successfully deleted deployment: %s", depName)
			deploymentDeleted = true
			break
		} else {
			log.Printf("Failed to delete deployment %s: %v", depName, err)
		}
	}

	if !deploymentDeleted {
		log.Printf("Warning: Could not delete any deployment for userID: %s", userID)
	}

	// Delete Node.js environment ConfigMap using the correct name format
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

	log.Printf("Sandbox deletion process completed for user: %s", userID)
	return nil
}

// ListSandboxes retrieves all sandboxes in the namespace
func (c *ClientWithTraefik) ListSandboxes(ctx context.Context) ([]SandboxInfo, error) {
	// Reuse the base client's implementation
	return c.Client.ListSandboxes(ctx)
}

// GetSandboxStatus retrieves the status of a specific sandbox by user ID
func (c *ClientWithTraefik) GetSandboxStatus(ctx context.Context, userID string) (*SandboxInfo, error) {
	// Reuse the base client's implementation
	return c.Client.GetSandboxStatus(ctx, userID)
}

// Helper function to convert a struct to unstructured.Unstructured
func convertToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	
	unstructuredObj := &unstructured.Unstructured{}
	err = json.Unmarshal(data, unstructuredObj)
	if err != nil {
		return nil, err
	}
	
	return unstructuredObj, nil
}