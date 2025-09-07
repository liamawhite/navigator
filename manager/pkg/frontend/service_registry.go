// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frontend

import (
	"context"
	"log/slog"
	"strings"

	"github.com/liamawhite/navigator/manager/pkg/providers"
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServiceRegistryService implements the frontend ServiceRegistryService
type ServiceRegistryService struct {
	frontendv1alpha1.UnimplementedServiceRegistryServiceServer
	connectionManager providers.ReadOptimizedConnectionManager
	proxyProvider     providers.ProxyConfigProvider
	istioProvider     providers.IstioResourcesProvider
	logger            *slog.Logger
}

// NewServiceRegistryService creates a new service registry service
func NewServiceRegistryService(connectionManager providers.ReadOptimizedConnectionManager, proxyProvider providers.ProxyConfigProvider, istioProvider providers.IstioResourcesProvider, logger *slog.Logger) *ServiceRegistryService {
	return &ServiceRegistryService{
		connectionManager: connectionManager,
		proxyProvider:     proxyProvider,
		istioProvider:     istioProvider,
		logger:            logger,
	}
}

// ListServices returns all services in the specified namespace and/or cluster
func (s *ServiceRegistryService) ListServices(ctx context.Context, req *frontendv1alpha1.ListServicesRequest) (*frontendv1alpha1.ListServicesResponse, error) {
	s.logger.Debug("listing services", "namespace", req.Namespace, "cluster_id", req.ClusterId)

	namespace := ""
	clusterID := ""

	if req.Namespace != nil {
		namespace = *req.Namespace
	}
	if req.ClusterId != nil {
		clusterID = *req.ClusterId
	}

	aggServices := s.connectionManager.ListAggregatedServices(namespace, clusterID)
	services := make([]*frontendv1alpha1.Service, 0, len(aggServices))

	for _, aggService := range aggServices {
		service := convertAggregatedService(aggService)
		services = append(services, service)
	}

	s.logger.Debug("listed services", "count", len(services))

	return &frontendv1alpha1.ListServicesResponse{
		Services: services,
	}, nil
}

// GetService returns detailed information about a specific service
func (s *ServiceRegistryService) GetService(ctx context.Context, req *frontendv1alpha1.GetServiceRequest) (*frontendv1alpha1.GetServiceResponse, error) {
	s.logger.Debug("getting service", "id", req.Id)

	aggService, exists := s.connectionManager.GetAggregatedService(req.Id)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "service not found: %s", req.Id)
	}

	service := convertAggregatedService(aggService)

	s.logger.Debug("got service", "id", req.Id, "instances", len(service.Instances))

	return &frontendv1alpha1.GetServiceResponse{
		Service: service,
	}, nil
}

// GetServiceInstance returns detailed information about a specific service instance
func (s *ServiceRegistryService) GetServiceInstance(ctx context.Context, req *frontendv1alpha1.GetServiceInstanceRequest) (*frontendv1alpha1.GetServiceInstanceResponse, error) {
	s.logger.Debug("getting service instance", "service_id", req.ServiceId, "instance_id", req.InstanceId)

	aggInstance, exists := s.connectionManager.GetAggregatedServiceInstance(req.InstanceId)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "service instance not found: %s", req.InstanceId)
	}

	instance := convertAggregatedServiceInstanceToDetail(aggInstance)

	s.logger.Debug("got service instance", "instance_id", req.InstanceId)

	return &frontendv1alpha1.GetServiceInstanceResponse{
		Instance: instance,
	}, nil
}

// GetProxyConfig retrieves the Envoy proxy configuration for a specific service instance
func (s *ServiceRegistryService) GetProxyConfig(ctx context.Context, req *frontendv1alpha1.GetProxyConfigRequest) (*frontendv1alpha1.GetProxyConfigResponse, error) {
	s.logger.Debug("getting proxy config", "service_id", req.ServiceId, "instance_id", req.InstanceId)

	// Parse instance ID to extract cluster, namespace, and pod name
	clusterID, namespace, podName, err := parseInstanceID(req.InstanceId)
	if err != nil {
		s.logger.Warn("invalid instance ID format", "instance_id", req.InstanceId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid instance ID format: %v", err)
	}

	// Verify the instance exists
	_, exists := s.connectionManager.GetAggregatedServiceInstance(req.InstanceId)
	if !exists {
		s.logger.Warn("service instance not found", "instance_id", req.InstanceId)
		return nil, status.Errorf(codes.NotFound, "service instance not found: %s", req.InstanceId)
	}

	// Request proxy configuration from the appropriate edge cluster
	proxyConfig, err := s.proxyProvider.GetProxyConfig(ctx, clusterID, namespace, podName)
	if err != nil {
		s.logger.Error("failed to get proxy config",
			"instance_id", req.InstanceId,
			"cluster_id", clusterID,
			"namespace", namespace,
			"pod_name", podName,
			"error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve proxy configuration: %v", err)
	}

	s.logger.Debug("got proxy config",
		"instance_id", req.InstanceId,
		"cluster_id", clusterID,
		"version", proxyConfig.Version)

	return &frontendv1alpha1.GetProxyConfigResponse{
		ProxyConfig: proxyConfig,
	}, nil
}

// GetIstioResources retrieves the Istio configuration resources for a specific service instance
func (s *ServiceRegistryService) GetIstioResources(ctx context.Context, req *frontendv1alpha1.GetIstioResourcesRequest) (*frontendv1alpha1.GetIstioResourcesResponse, error) {
	s.logger.Debug("getting istio resources", "service_id", req.ServiceId, "instance_id", req.InstanceId)

	// Parse instance ID to extract cluster, namespace, and pod name
	clusterID, namespace, _, err := parseInstanceID(req.InstanceId)
	if err != nil {
		s.logger.Warn("invalid instance ID format", "instance_id", req.InstanceId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid instance ID format: %v", err)
	}

	// Get the service instance to extract labels
	aggInstance, exists := s.connectionManager.GetAggregatedServiceInstance(req.InstanceId)
	if !exists {
		s.logger.Warn("service instance not found", "instance_id", req.InstanceId)
		return nil, status.Errorf(codes.NotFound, "service instance not found: %s", req.InstanceId)
	}

	// Convert to ServiceInstance for the istio provider
	serviceInstance := &backendv1alpha1.ServiceInstance{
		Labels: aggInstance.Labels,
	}

	// Request Istio resources from the appropriate cluster
	istioResources, err := s.istioProvider.GetIstioResourcesForWorkload(ctx, clusterID, namespace, serviceInstance)
	if err != nil {
		s.logger.Error("failed to get istio resources",
			"instance_id", req.InstanceId,
			"cluster_id", clusterID,
			"namespace", namespace,
			"error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve istio resources: %v", err)
	}

	s.logger.Debug("got istio resources",
		"instance_id", req.InstanceId,
		"cluster_id", clusterID,
		"gateways", len(istioResources.Gateways),
		"virtual_services", len(istioResources.VirtualServices),
		"destination_rules", len(istioResources.DestinationRules))

	return istioResources, nil
}

// parseInstanceID parses an instance ID in the format "cluster_id:namespace:pod_name"
// Returns cluster ID, namespace, pod name, and any error
func parseInstanceID(instanceID string) (clusterID, namespace, podName string, err error) {
	parts := strings.Split(instanceID, ":")
	if len(parts) != 3 {
		return "", "", "", status.Errorf(codes.InvalidArgument, "invalid instance ID format, expected 'cluster_id:namespace:pod_name', got: %s", instanceID)
	}

	clusterID = parts[0]
	namespace = parts[1]
	podName = parts[2]

	if clusterID == "" || namespace == "" || podName == "" {
		return "", "", "", status.Errorf(codes.InvalidArgument, "instance ID contains empty components: %s", instanceID)
	}

	return clusterID, namespace, podName, nil
}
