package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ensureNamespace creates the namespace if it doesn't exist
func (c *Client) ensureNamespace(ctx context.Context) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, c.namespace, metav1.GetOptions{})
	if err == nil {
		// Namespace exists
		return nil
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.namespace,
		},
	}
	_, err = c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

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

// createDeployment creates a deployment for the user's sandbox
func (c *Client) createDeployment(ctx context.Context, userID string, envVars map[string]string, hasNodeEnv bool) error {
	deploymentName := fmt.Sprintf("%s-deployment", userID)

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

	// Create volume mounts for the container
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "user-data",
			MountPath: "/home/nodeuser/.iris",
		},
	}

	// Create volumes for the pod
	volumes := []corev1.Volume{
		{
			Name: "user-data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("%s-pvc", userID),
				},
			},
		},
	}

	// Add Node.js environment variables ConfigMap if needed
	if hasNodeEnv {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "node-env",
			MountPath: "/app/.env",
			SubPath:   "node.env",
		})

		volumes = append(volumes, corev1.Volume{
			Name: "node-env",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-node-env", userID),
					},
				},
			},
		})
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
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
					Containers: []corev1.Container{
						{
							Name:  "sandbox",
							Image: "shanurcsenitap/iris_agent:latest",
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

// createService creates a service for the user's sandbox
func (c *Client) createService(ctx context.Context, userID string) error {
	serviceName := fmt.Sprintf("%s-service", userID)
	
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "user-sandbox",
				"user": userID,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "vnc",
					Port:       6901,
					TargetPort: intstr.FromInt(6901),
				},
				{
					Name:       "http",
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
				},
			},
		},
	}

	_, err := c.clientset.CoreV1().Services(c.namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

// createIngress creates an ingress for the user's sandbox
func (c *Client) createIngress(ctx context.Context, userID string) error {
	ingressName := fmt.Sprintf("%s-ingress", userID)

	pathTypePrefix := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: ingressName,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "traefik",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("%s-vnc.%s", userID, c.domain),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: fmt.Sprintf("%s-service", userID),
											Port: networkingv1.ServiceBackendPort{
												Name: "vnc",
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Host: fmt.Sprintf("%s-api.%s", userID, c.domain),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: fmt.Sprintf("%s-service", userID),
											Port: networkingv1.ServiceBackendPort{
												Name: "http",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Create(ctx, ingress, metav1.CreateOptions{})
	return err
}