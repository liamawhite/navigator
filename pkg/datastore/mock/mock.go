package mock

import (
	"context"

	"github.com/stretchr/testify/assert"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// Datastore is a simple mock implementation of ServiceDatastore for testing
type Datastore struct {
	Services map[string][]*v1alpha1.Service // namespace -> services
}

func (m *Datastore) ListServices(ctx context.Context, namespace string) ([]*v1alpha1.Service, error) {
	if namespace == "" {
		// Return all services from all namespaces
		var allServices []*v1alpha1.Service
		for _, services := range m.Services {
			allServices = append(allServices, services...)
		}
		return allServices, nil
	}

	services, exists := m.Services[namespace]
	if !exists {
		return []*v1alpha1.Service{}, nil
	}
	return services, nil
}

func (m *Datastore) GetService(ctx context.Context, id string) (*v1alpha1.Service, error) {
	for _, services := range m.Services {
		for _, service := range services {
			if service.Id == id {
				return service, nil
			}
		}
	}

	return nil, assert.AnError
}

func (m *Datastore) GetServiceInstance(ctx context.Context, serviceID, instanceID string) (*v1alpha1.ServiceInstanceDetail, error) {
	// Find the service instance in our mock data
	for _, services := range m.Services {
		for _, service := range services {
			if service.Id == serviceID {
				for _, instance := range service.Instances {
					if instance.InstanceId == instanceID {
						// Create a detailed response with mock data
						return &v1alpha1.ServiceInstanceDetail{
							InstanceId:     instance.InstanceId,
							Ip:             instance.Ip,
							Pod:            instance.Pod,
							Namespace:      instance.Namespace,
							ClusterName:    instance.ClusterName,
							IsEnvoyPresent: instance.IsEnvoyPresent,
							ServiceName:    service.Name,
							PodStatus:      "Running",
							CreatedAt:      "2023-01-01T00:00:00Z",
							Labels:         map[string]string{"app": service.Name, "version": "v1"},
							Annotations:    map[string]string{"deployment.kubernetes.io/revision": "1"},
							Containers: []*v1alpha1.ContainerInfo{
								{
									Name:         service.Name,
									Image:        "nginx:latest",
									Ready:        true,
									RestartCount: 0,
									Status:       "Running",
								},
							},
							NodeName: "node-1",
						}, nil
					}
				}
			}
		}
	}

	return nil, assert.AnError
}
