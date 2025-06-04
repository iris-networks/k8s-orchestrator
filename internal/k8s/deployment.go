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
func (c *Client) createDeployment(ctx context.Context, userID string, envVars map[string]string, hasNodeEnv bool) error {
	deploymentName := fmt.Sprintf("iris-%s-deployment", userID)

	// Create deployment
	var replicas int32 = 1

	// Convert map of environment variables to Kubernetes EnvVar slice
	var envVarSlice []corev1.EnvVar
	for key, value := range envVars {
		envVarSlice = append(envVarSlice, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	// Always include the USER_ID environment variable
	envVarSlice = append(envVarSlice, corev1.EnvVar{
		Name:  "USER_ID",
		Value: userID,
	})

	// Get volume mounts for the container from storage
	volumeMounts := c.getUserDataVolumeMounts()

	// Get volumes for the pod from storage
	volumes := []corev1.Volume{
		c.getUserDataVolume(userID),
	}

	// Add Node.js environment variables ConfigMap if needed
	if hasNodeEnv {
		volumeMounts = append(volumeMounts, c.getNodeEnvVolumeMount())
		volumes = append(volumes, c.getNodeEnvVolume(userID))
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
							Name:  "volume-permissions",
							Image: "busybox",
							Command: []string{
								"sh",
								"-c",
								"mkdir -p /home/nodeuser/.iris && chmod -R 777 /home/nodeuser/.iris && mkdir -p /home/headless/.mozilla/firefox && chmod -R 777 /home/headless/.mozilla/firefox && mkdir -p /home/vncuser/.config && chmod -R 777 /home/vncuser/.config && rm -f /home/vncuser/.config/google-chrome/Singleton*",
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
							Image: "shanurcsenitap/iris_agent:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
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
									corev1.ResourceMemory: resource.MustParse("4Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("2Gi"),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/health",
										Port: intstr.FromInt(3000),
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      5,
								PeriodSeconds:       15,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/health",
										Port: intstr.FromInt(3000),
									},
								},
								InitialDelaySeconds: 1,
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

	_, err := c.clientset.AppsV1().Deployments(c.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}