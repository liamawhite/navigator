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
