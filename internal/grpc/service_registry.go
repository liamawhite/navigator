package grpc

import (
	"context"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/datastore"
)

// ServiceRegistryServer implements the ServiceRegistryService gRPC interface.
type ServiceRegistryServer struct {
	v1alpha1.UnimplementedServiceRegistryServiceServer
	datastore datastore.ServiceDatastore
}

// NewServiceRegistryServer creates a new ServiceRegistryServer with the given datastore.
func NewServiceRegistryServer(ds datastore.ServiceDatastore) *ServiceRegistryServer {
	return &ServiceRegistryServer{
		datastore: ds,
	}
}

// ListServices returns all services in the specified namespace, or all namespaces if not specified.
func (s *ServiceRegistryServer) ListServices(ctx context.Context, req *v1alpha1.ListServicesRequest) (*v1alpha1.ListServicesResponse, error) {
	namespace := req.GetNamespace()

	services, err := s.datastore.ListServices(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.ListServicesResponse{
		Services: services,
	}, nil
}

// GetService returns detailed information about a specific service.
func (s *ServiceRegistryServer) GetService(ctx context.Context, req *v1alpha1.GetServiceRequest) (*v1alpha1.GetServiceResponse, error) {
	service, err := s.datastore.GetService(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.GetServiceResponse{
		Service: service,
	}, nil
}
