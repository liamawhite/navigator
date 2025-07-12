package kubeconfig

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/datastore"
)

// Ensure datastore implements the ServiceDatastore interface
var _ types.ServiceDatastore = (*datastore)(nil)

type datastore struct {
	client    kubernetes.Interface
	namespace string
}

func New(kubeconfigPath string) (*datastore, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &datastore{
		client: client,
	}, nil
}

func (d *datastore) ListServices(ctx context.Context, namespace string) ([]*v1alpha1.Service, error) {
	// Use all namespaces if namespace is empty
	targetNamespace := namespace
	if targetNamespace == "" {
		targetNamespace = metav1.NamespaceAll
	}

	services, err := d.client.CoreV1().Services(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var result []*v1alpha1.Service
	for _, svc := range services.Items {
		endpoints, err := d.getEndpointsForService(ctx, svc.Name, svc.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get endpoints for service %s: %w", svc.Name, err)
		}

		result = append(result, &v1alpha1.Service{
			Id:        fmt.Sprintf("%s:%s", svc.Namespace, svc.Name),
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Instances: endpoints,
		})
	}

	return result, nil
}

func (d *datastore) GetService(ctx context.Context, id string) (*v1alpha1.Service, error) {
	// Parse namespace:name from ID
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid service ID format: %s (expected namespace:name)", id)
	}
	namespace, name := parts[0], parts[1]

	svc, err := d.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", name, err)
	}

	endpoints, err := d.getEndpointsForService(ctx, svc.Name, svc.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints for service %s: %w", svc.Name, err)
	}

	return &v1alpha1.Service{
		Id:        id,
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Instances: endpoints,
	}, nil
}

func (d *datastore) getEndpointsForService(ctx context.Context, name, namespace string) ([]*v1alpha1.ServiceInstance, error) {
	endpoints, err := d.client.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// Return empty list if endpoints don't exist (service might not have any pods yet)
		return []*v1alpha1.ServiceInstance{}, nil
	}

	var instances []*v1alpha1.ServiceInstance
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			podName := ""
			podNamespace := namespace
			if address.TargetRef != nil && address.TargetRef.Kind == "Pod" {
				podName = address.TargetRef.Name
				if address.TargetRef.Namespace != "" {
					podNamespace = address.TargetRef.Namespace
				}
			}

			hasProxySidecar := false
			if podName != "" {
				hasProxySidecar, err = d.checkForProxySidecar(ctx, podName, podNamespace)
				if err != nil {
					// Log warning but don't fail the request
					// In production, you might want to use a proper logger here
					hasProxySidecar = false
				}
			}

			instances = append(instances, &v1alpha1.ServiceInstance{
				Ip:              address.IP,
				Pod:             podName,
				Namespace:       podNamespace,
				HasProxySidecar: hasProxySidecar,
			})
		}
	}

	return instances, nil
}

// checkForProxySidecar checks if a pod has an Istio proxy sidecar container
func (d *datastore) checkForProxySidecar(ctx context.Context, podName, namespace string) (bool, error) {
	pod, err := d.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get pod %s: %w", podName, err)
	}

	// Check all containers in the pod for Istio proxy
	for _, container := range pod.Spec.Containers {
		if container.Name == "istio-proxy" || strings.HasPrefix(container.Image, "istio/proxyv2") {
			return true, nil
		}
	}

	// Check init containers as well
	for _, container := range pod.Spec.InitContainers {
		if container.Name == "istio-proxy" || strings.HasPrefix(container.Image, "istio/proxyv2") {
			return true, nil
		}
	}

	return false, nil
}
