package localenv

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// Fixtures manages deployment and service creation in a Kubernetes cluster
type Fixtures struct {
	client    kubernetes.Interface
	namespace string
}

// NewFixtures creates a new fixtures manager
func NewFixtures(client kubernetes.Interface, namespace string) *Fixtures {
	return &Fixtures{
		client:    client,
		namespace: namespace,
	}
}

// CreateNamespace creates the fixtures namespace
func (f *Fixtures) CreateNamespace(ctx context.Context) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.namespace,
		},
	}

	_, err := f.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace %s: %w", f.namespace, err)
	}

	return nil
}

// DeleteNamespace removes the fixtures namespace and all its resources
func (f *Fixtures) DeleteNamespace(ctx context.Context) error {
	err := f.client.CoreV1().Namespaces().Delete(ctx, f.namespace, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", f.namespace, err)
	}

	return nil
}

// CreateWebService creates a simple web service with deployment and service
func (f *Fixtures) CreateWebService(ctx context.Context, name string, replicas int32) error {
	// Create deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "ghcr.io/liamawhite/microservice:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVICE_NAME",
									Value: name,
								},
								{
									Name:  "PORT",
									Value: "8080",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := f.client.AppsV1().Deployments(f.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment %s: %w", name, err)
	}

	// Create service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err = f.client.CoreV1().Services(f.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service %s: %w", name, err)
	}

	return nil
}

// CreateHeadlessService creates a headless service (ClusterIP: None)
func (f *Fixtures) CreateHeadlessService(ctx context.Context, name string) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.namespace,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err := f.client.CoreV1().Services(f.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create headless service %s: %w", name, err)
	}

	return nil
}

// CreateExternalService creates a service with external endpoints
func (f *Fixtures) CreateExternalService(ctx context.Context, name string, externalIPs []string) error {
	// Create service without selector (external service)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err := f.client.CoreV1().Services(f.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create external service %s: %w", name, err)
	}

	// Create endpoint slice manually
	endpoints := make([]discoveryv1.Endpoint, len(externalIPs))
	for i, ip := range externalIPs {
		endpoints[i] = discoveryv1.Endpoint{
			Addresses: []string{ip},
			Conditions: discoveryv1.EndpointConditions{
				Ready: func() *bool { b := true; return &b }(),
			},
		}
	}

	endpointSlice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-abc123",
			Namespace: f.namespace,
			Labels: map[string]string{
				discoveryv1.LabelServiceName: name,
			},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints:   endpoints,
		Ports: []discoveryv1.EndpointPort{
			{
				Port:     func() *int32 { p := int32(8080); return &p }(),
				Protocol: func() *corev1.Protocol { p := corev1.ProtocolTCP; return &p }(),
			},
		},
	}

	_, err = f.client.DiscoveryV1().EndpointSlices(f.namespace).Create(ctx, endpointSlice, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create endpoint slice for service %s: %w", name, err)
	}

	return nil
}

// CreateTopologyService creates a single service in a microservice topology
func (f *Fixtures) CreateTopologyService(ctx context.Context, serviceName string, replicas int32, nextService string) error {
	// Create deployment with environment variables for chaining
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: f.namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": serviceName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": serviceName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  serviceName,
							Image: "ghcr.io/liamawhite/microservice:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVICE_NAME",
									Value: serviceName,
								},
								{
									Name:  "PORT",
									Value: "8080",
								},
							},
						},
					},
				},
			},
		},
	}

	// Add next service in chain as environment variable
	if nextService != "" {
		deployment.Spec.Template.Spec.Containers[0].Env = append(
			deployment.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name:  "NEXT_SERVICE",
				Value: fmt.Sprintf("%s.%s.svc.cluster.local:8080", nextService, f.namespace),
			},
		)
	}

	_, err := f.client.AppsV1().Deployments(f.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment %s: %w", serviceName, err)
	}

	// Create service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: f.namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": serviceName,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err = f.client.CoreV1().Services(f.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service %s: %w", serviceName, err)
	}

	return nil
}

// WaitForDeployment waits for a deployment to be ready
func (f *Fixtures) WaitForDeployment(ctx context.Context, name string) error {
	return waitForCondition(ctx, func() (bool, error) {
		deployment, err := f.client.AppsV1().Deployments(f.namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas, nil
	})
}

// WaitForService waits for a service to have endpoint slices
func (f *Fixtures) WaitForService(ctx context.Context, name string) error {
	return waitForCondition(ctx, func() (bool, error) {
		endpointSlices, err := f.client.DiscoveryV1().EndpointSlices(f.namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", discoveryv1.LabelServiceName, name),
		})
		if err != nil {
			return false, err
		}

		for _, endpointSlice := range endpointSlices.Items {
			for _, endpoint := range endpointSlice.Endpoints {
				if endpoint.Conditions.Ready != nil && *endpoint.Conditions.Ready {
					return true, nil
				}
			}
		}

		return false, nil
	})
}

// waitForCondition polls a condition until it returns true or context is canceled
func waitForCondition(ctx context.Context, condition func() (bool, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ready, err := condition()
			if err != nil {
				return err
			}
			if ready {
				return nil
			}
			// Wait before checking again
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
				continue
			}
		}
	}
}
