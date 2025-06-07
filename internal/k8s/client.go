package k8s

import (
	"path/filepath"

	"github.com/shanurcsenitap/irisk8s/internal/config"
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
	config    *config.Configuration
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	var k8sConfig *rest.Config
	var err error

	// Try in-cluster config first
	k8sConfig, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		home := homedir.HomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	// Default namespace and domain - should be configurable via environment variables
	namespace := "user-sandboxes"
	domain := "tryiris.dev"

	// Get application configuration
	appConfig := config.GetConfig()

	// Set the ResourceExpirationTime variable for backward compatibility
	ResourceExpirationTime = appConfig.SandboxTimeoutDuration

	return &Client{
		clientset: clientset,
		namespace: namespace,
		domain:    domain,
		config:    appConfig,
	}, nil
}