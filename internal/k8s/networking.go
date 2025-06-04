package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// createService creates a service for the user's sandbox
func (c *Client) createService(ctx context.Context, userID string) error {
	serviceName := fmt.Sprintf("iris-%s-service", userID)

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