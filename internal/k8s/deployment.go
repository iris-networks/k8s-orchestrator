package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// createDeployment creates a deployment for the user's sandbox
func (c *Client) createDeployment(ctx context.Context, userID string) error {
	deploymentName := fmt.Sprintf("%s-deployment", userID)

	// Create deployment
	var replicas int32 = 1

	// Only include the USER_ID environment variable
	var envVarSlice []corev1.EnvVar
	envVarSlice = append(envVarSlice, corev1.EnvVar{
		Name:  "USER_ID",
		Value: userID,
	})

	// Get volume mounts for the container from storage
	volumeMounts := c.getUserDataVolumeMounts()

	// Get volumes for the pod from storage
	volumes := []corev1.Volume{
		c.getUserDataVolume(userID),
		{
			Name: "shm-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMediumMemory,
					SizeLimit: resource.NewQuantity(512*1024*1024, resource.BinarySI), // 512 MiB
				},
			},
		},
	}

	// Get image tag from configmap
	imageTag, err := c.getImageTagFromConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get image tag from configmap: %v", err)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Labels: map[string]string{
				"app":  "user-sandbox",
				"user": userID,
			},
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
					// Add init container to set correct permissions on the volume
					InitContainers: []corev1.Container{
						{
							Name:            "volume-permissions",
							Image:           "busybox",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								"chmod -R 777 /config && rm -f /config/browser/user-data/Singleton* && wait",
							},
							VolumeMounts: c.getUserDataVolumeMounts(),
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: func() *int64 {
									var uid int64 = 0 // Run as root to set permissions
									return &uid
								}(),
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "sandbox",
							Image: fmt.Sprintf("us-central1-docker.pkg.dev/driven-seer-460401-p9/iris-repo/iris_agent:%s", imageTag),
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeUnconfined,
								},
							},
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
							Env:          envVarSlice,
							VolumeMounts: volumeMounts,
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
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/health",
										Port: intstr.FromInt(3000),
									},
								},
								InitialDelaySeconds: 3,
								TimeoutSeconds:      2,
								PeriodSeconds:       3,
								SuccessThreshold:    1,
								FailureThreshold:    10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/health",
										Port: intstr.FromInt(3000),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      1,
								PeriodSeconds:       3,
								SuccessThreshold:    1,
								FailureThreshold:    2,
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	_, err = c.clientset.AppsV1().Deployments(c.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

// getImageTagFromConfigMap retrieves the container image tag from the app-config configmap
func (c *Client) getImageTagFromConfigMap(ctx context.Context) (string, error) {
	configMap, err := c.clientset.CoreV1().ConfigMaps("user-sandboxes").Get(ctx, "app-config", metav1.GetOptions{})
	if err != nil {
		return "latest", fmt.Errorf("failed to get configmap: %v", err)
	}

	imageTag, exists := configMap.Data["container-image-tag"]
	if !exists {
		return "latest", fmt.Errorf("container-image-tag not found in configmap")
	}

	return imageTag, nil
}