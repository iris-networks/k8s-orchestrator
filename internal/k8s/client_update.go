package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
				CertResolver: "cloudflare",
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
				CertResolver: "cloudflare",
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

// CreateSandbox creates a new sandbox for a user with Traefik IngressRoutes
func (c *ClientWithTraefik) CreateSandbox(userID string) error {
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