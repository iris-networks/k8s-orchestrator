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

// getUserDataVolume returns a volume for user data linked to the user's PVC
func (c *Client) getUserDataVolume(userID string) corev1.Volume {
	return corev1.Volume{
		Name: "user-data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf("%s-pvc", userID),
			},
		},
	}
}

// getUserDataVolumeMounts returns volume mounts for the user data volume
func (c *Client) getUserDataVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "user-data",
			MountPath: "/home/nodeuser/.iris",
		},
		{
			Name:      "user-data",
			MountPath: "/home/vncuser/.config",
		},
		{
			Name:      "user-data",
			MountPath: "/home/vncuser/.config/chromium",
			SubPath:   "chromium-data",
		},
	}
}

// This function has been removed as we no longer support .env file volume mounts

// These functions have been removed as we no longer support .env file volume mounts
// or environment variable passing from the API.