package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	_, err = c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

// createNodeEnvConfigMap creates a ConfigMap for Node.js specific environment variables
func (c *Client) createNodeEnvConfigMap(ctx context.Context, userID string, nodeEnvVars map[string]string) error {
	configMapName := fmt.Sprintf("%s-node-env", userID)

	// Convert the environment variables to a .env file format
	envFileContent := ""
	for key, value := range nodeEnvVars {
		envFileContent += fmt.Sprintf("%s=%s\n", key, value)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
		},
		Data: map[string]string{
			"node.env": envFileContent,
		},
	}

	_, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Create(ctx, configMap, metav1.CreateOptions{})
	return err
}