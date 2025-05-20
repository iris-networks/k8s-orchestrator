package utils

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// GetKubernetesClient creates a Kubernetes clientset
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try to use in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		// Use the current context in kubeconfig
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

	return clientset, nil
}

// GetDomainFromEnv gets the domain name from environment variables or defaults to local.dev
func GetDomainFromEnv() string {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return "local.dev"
	}
	return domain
}