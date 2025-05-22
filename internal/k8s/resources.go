package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/k8sgo/internal/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// createNamespace creates a namespace for a user
func (c *Client) createNamespace(username string) (string, error) {
	namespaceName := fmt.Sprintf("user-%s", username)
	
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				"app":      "k8sgo",
				"username": username,
			},
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	
	return namespaceName, nil
}

// createPVC creates a persistent volume claim for a user
func (c *Client) createPVC(username, namespace, size string) error {
	pvcName := fmt.Sprintf("%s-data", username)
	
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvcName,
			Labels: map[string]string{
				"app":      "k8sgo",
				"username": username,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

// createDeployment creates a deployment for a user
func (c *Client) createDeployment(username, namespace, image string, ports []int, volumeMounts []models.VolumeMount, envVars map[string]string) error {
	deploymentName := fmt.Sprintf("%s-desktop", username)
	
	// Convert ports to container ports
	containerPorts := make([]corev1.ContainerPort, len(ports))
	for i, port := range ports {
		containerPorts[i] = corev1.ContainerPort{
			ContainerPort: int32(port),
			Protocol:      corev1.ProtocolTCP,
		}
	}
	
	// Convert environment variables
	env := make([]corev1.EnvVar, 0, len(envVars))
	for key, value := range envVars {
		env = append(env, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}
	
	// Default volume mount if none provided
	k8sVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: "/home/headless/Documents",
		},
	}
	
	// Add custom volume mounts if provided
	if len(volumeMounts) > 0 {
		k8sVolumeMounts = make([]corev1.VolumeMount, len(volumeMounts))
		for i, vm := range volumeMounts {
			k8sVolumeMounts[i] = corev1.VolumeMount{
				Name:      vm.Name,
				MountPath: vm.MountPath,
			}
		}
	}
	
	// Create deployment
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Labels: map[string]string{
				"app":      "k8sgo",
				"username": username,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":      "k8sgo",
					"username": username,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":      "k8sgo",
						"username": username,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "desktop",
							Image: image,
							Ports: containerPorts,
							VolumeMounts: k8sVolumeMounts,
							Env:   env,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: fmt.Sprintf("%s-data", username),
								},
							},
						},
					},
				},
			},
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

// createService creates a service for a user
func (c *Client) createService(username, namespace string, ports []int) error {
	serviceName := fmt.Sprintf("%s-svc", username)
	
	// Create service ports
	servicePorts := make([]corev1.ServicePort, len(ports))
	for i, port := range ports {
		servicePorts[i] = corev1.ServicePort{
			Name:       fmt.Sprintf("port-%d", port),
			Port:       int32(port),
			TargetPort: intstr.FromInt(port),
			Protocol:   corev1.ProtocolTCP,
		}
	}
	
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"app":      "k8sgo",
				"username": username,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":      "k8sgo",
				"username": username,
			},
			Ports: servicePorts,
			Type:  corev1.ServiceTypeClusterIP,
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

// createIngress creates an ingress for a user
func (c *Client) createIngress(username, namespace string, ports []int) error {
	serviceName := fmt.Sprintf("%s-svc", username)
	host := fmt.Sprintf("%s.%s", username, c.domain)

	// Create a separate ingress for each port
	for _, port := range ports {
		ingressName := fmt.Sprintf("%s-ingress-%d", username, port)
		pathType := networkingv1.PathTypePrefix

		ingressClassName := "nginx"
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressName,
				Labels: map[string]string{
					"app":      "k8sgo",
					"username": username,
					"port":     fmt.Sprintf("%d", port),
				},
				Annotations: map[string]string{
					"nginx.ingress.kubernetes.io/proxy-connect-timeout": "3600",
					"nginx.ingress.kubernetes.io/proxy-read-timeout": "3600",
					"nginx.ingress.kubernetes.io/proxy-send-timeout": "3600",
					"nginx.ingress.kubernetes.io/websocket-services": serviceName,
					"nginx.ingress.kubernetes.io/ssl-redirect": "true",
					"cert-manager.io/cluster-issuer": "letsencrypt-prod",
				},
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &ingressClassName,
				TLS: []networkingv1.IngressTLS{
					{
						Hosts:      []string{host},
						SecretName: fmt.Sprintf("%s-tls-%d", username, port),
					},
				},
				Rules: []networkingv1.IngressRule{
					{
						Host: host,
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path:     "/",
										PathType: &pathType,
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: serviceName,
												Port: networkingv1.ServiceBackendPort{
													Number: int32(port),
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := c.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
		cancel()
		if err != nil {
			return err
		}
	}

	return nil
}