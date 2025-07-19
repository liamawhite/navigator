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

package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/liamawhite/navigator/manager/pkg/connections"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ReadOptimizedConnectionManager extends ConnectionManager with read-optimized methods
type ReadOptimizedConnectionManager interface {
	ConnectionManager
	ListAggregatedServices(namespace, clusterID string) []*connections.AggregatedService
	GetAggregatedService(serviceID string) (*connections.AggregatedService, bool)
	GetAggregatedServiceInstance(instanceID string) (*connections.AggregatedServiceInstance, bool)
	GetConnectionInfo() map[string]connections.ConnectionInfo
}

// ProxyConfigProvider defines the interface for retrieving proxy configurations
type ProxyConfigProvider interface {
	GetProxyConfig(ctx context.Context, clusterID, namespace, podName string) (*types.ProxyConfig, error)
}

// FrontendService implements the frontend ServiceRegistryService
type FrontendService struct {
	frontendv1alpha1.UnimplementedServiceRegistryServiceServer
	connectionManager ReadOptimizedConnectionManager
	proxyProvider     ProxyConfigProvider
	logger            *slog.Logger
}

// NewFrontendService creates a new frontend service
func NewFrontendService(connectionManager ReadOptimizedConnectionManager, proxyProvider ProxyConfigProvider, logger *slog.Logger) *FrontendService {
	return &FrontendService{
		connectionManager: connectionManager,
		proxyProvider:     proxyProvider,
		logger:            logger,
	}
}

// ListServices returns all services in the specified namespace and/or cluster
func (f *FrontendService) ListServices(ctx context.Context, req *frontendv1alpha1.ListServicesRequest) (*frontendv1alpha1.ListServicesResponse, error) {
	f.logger.Debug("listing services", "namespace", req.Namespace, "cluster_id", req.ClusterId)

	namespace := ""
	clusterID := ""

	if req.Namespace != nil {
		namespace = *req.Namespace
	}
	if req.ClusterId != nil {
		clusterID = *req.ClusterId
	}

	aggServices := f.connectionManager.ListAggregatedServices(namespace, clusterID)
	services := make([]*frontendv1alpha1.Service, 0, len(aggServices))

	for _, aggService := range aggServices {
		service := f.convertAggregatedService(aggService)
		services = append(services, service)
	}

	f.logger.Debug("listed services", "count", len(services))

	return &frontendv1alpha1.ListServicesResponse{
		Services: services,
	}, nil
}

// GetService returns detailed information about a specific service
func (f *FrontendService) GetService(ctx context.Context, req *frontendv1alpha1.GetServiceRequest) (*frontendv1alpha1.GetServiceResponse, error) {
	f.logger.Debug("getting service", "id", req.Id)

	aggService, exists := f.connectionManager.GetAggregatedService(req.Id)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "service not found: %s", req.Id)
	}

	service := f.convertAggregatedService(aggService)

	f.logger.Debug("got service", "id", req.Id, "instances", len(service.Instances))

	return &frontendv1alpha1.GetServiceResponse{
		Service: service,
	}, nil
}

// GetServiceInstance returns detailed information about a specific service instance
func (f *FrontendService) GetServiceInstance(ctx context.Context, req *frontendv1alpha1.GetServiceInstanceRequest) (*frontendv1alpha1.GetServiceInstanceResponse, error) {
	f.logger.Debug("getting service instance", "service_id", req.ServiceId, "instance_id", req.InstanceId)

	aggInstance, exists := f.connectionManager.GetAggregatedServiceInstance(req.InstanceId)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "service instance not found: %s", req.InstanceId)
	}

	instance := f.convertAggregatedServiceInstanceToDetail(aggInstance)

	f.logger.Debug("got service instance", "instance_id", req.InstanceId)

	return &frontendv1alpha1.GetServiceInstanceResponse{
		Instance: instance,
	}, nil
}

// GetProxyConfig retrieves the Envoy proxy configuration for a specific service instance
func (f *FrontendService) GetProxyConfig(ctx context.Context, req *frontendv1alpha1.GetProxyConfigRequest) (*frontendv1alpha1.GetProxyConfigResponse, error) {
	f.logger.Debug("getting proxy config", "service_id", req.ServiceId, "instance_id", req.InstanceId)

	// Parse instance ID to extract cluster, namespace, and pod name
	clusterID, namespace, podName, err := parseInstanceID(req.InstanceId)
	if err != nil {
		f.logger.Warn("invalid instance ID format", "instance_id", req.InstanceId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid instance ID format: %v", err)
	}

	// Verify the instance exists
	_, exists := f.connectionManager.GetAggregatedServiceInstance(req.InstanceId)
	if !exists {
		f.logger.Warn("service instance not found", "instance_id", req.InstanceId)
		return nil, status.Errorf(codes.NotFound, "service instance not found: %s", req.InstanceId)
	}

	// Request proxy configuration from the appropriate edge cluster
	proxyConfig, err := f.proxyProvider.GetProxyConfig(ctx, clusterID, namespace, podName)
	if err != nil {
		f.logger.Error("failed to get proxy config",
			"instance_id", req.InstanceId,
			"cluster_id", clusterID,
			"namespace", namespace,
			"pod_name", podName,
			"error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve proxy configuration: %v", err)
	}

	f.logger.Debug("got proxy config",
		"instance_id", req.InstanceId,
		"cluster_id", clusterID,
		"version", proxyConfig.Version)

	return &frontendv1alpha1.GetProxyConfigResponse{
		ProxyConfig: proxyConfig,
	}, nil
}

// ListClusters returns sync state information for all connected clusters
func (f *FrontendService) ListClusters(ctx context.Context, req *frontendv1alpha1.ListClustersRequest) (*frontendv1alpha1.ListClustersResponse, error) {
	f.logger.Debug("listing clusters")

	connectionInfos := f.connectionManager.GetConnectionInfo()
	clusters := make([]*frontendv1alpha1.ClusterSyncInfo, 0, len(connectionInfos))

	for _, connInfo := range connectionInfos {
		cluster := f.convertConnectionInfoToClusterSyncInfo(connInfo)
		clusters = append(clusters, cluster)
	}

	f.logger.Debug("listed clusters", "count", len(clusters))

	return &frontendv1alpha1.ListClustersResponse{
		Clusters: clusters,
	}, nil
}

// convertAggregatedService converts an AggregatedService to the frontend API format
func (f *FrontendService) convertAggregatedService(aggService *connections.AggregatedService) *frontendv1alpha1.Service {
	instances := make([]*frontendv1alpha1.ServiceInstance, 0, len(aggService.Instances))

	for _, aggInstance := range aggService.Instances {
		instance := f.convertAggregatedServiceInstance(aggInstance)
		instances = append(instances, instance)
	}

	return &frontendv1alpha1.Service{
		Id:        aggService.ID,
		Name:      aggService.Name,
		Namespace: aggService.Namespace,
		Instances: instances,
	}
}

// convertAggregatedServiceInstance converts an AggregatedServiceInstance to the frontend API format
func (f *FrontendService) convertAggregatedServiceInstance(aggInstance *connections.AggregatedServiceInstance) *frontendv1alpha1.ServiceInstance {
	return &frontendv1alpha1.ServiceInstance{
		InstanceId:   aggInstance.InstanceID,
		Ip:           aggInstance.IP,
		PodName:      aggInstance.PodName,
		Namespace:    aggInstance.Namespace,
		ClusterName:  aggInstance.ClusterName,
		EnvoyPresent: aggInstance.EnvoyPresent,
	}
}

// convertAggregatedServiceInstanceToDetail converts an AggregatedServiceInstance to the detailed frontend API format
func (f *FrontendService) convertAggregatedServiceInstanceToDetail(aggInstance *connections.AggregatedServiceInstance) *frontendv1alpha1.ServiceInstanceDetail {
	// Convert container information
	containers := make([]*frontendv1alpha1.Container, len(aggInstance.Containers))
	for i, container := range aggInstance.Containers {
		containers[i] = &frontendv1alpha1.Container{
			Name:         container.Name,
			Image:        container.Image,
			Status:       container.Status,
			Ready:        container.Ready,
			RestartCount: container.RestartCount,
		}
	}

	return &frontendv1alpha1.ServiceInstanceDetail{
		InstanceId:     aggInstance.InstanceID,
		Ip:             aggInstance.IP,
		PodName:        aggInstance.PodName,
		Namespace:      aggInstance.Namespace,
		ClusterName:    aggInstance.ClusterName,
		EnvoyPresent:   aggInstance.EnvoyPresent,
		ServiceName:    fmt.Sprintf("%s:%s", aggInstance.Namespace, extractServiceNameFromInstanceID(aggInstance.InstanceID)),
		Containers:     containers,
		PodStatus:      aggInstance.PodStatus,
		NodeName:       aggInstance.NodeName,
		CreatedAt:      aggInstance.CreatedAt,
		Labels:         aggInstance.Labels,
		Annotations:    aggInstance.Annotations,
		IsEnvoyPresent: aggInstance.IsEnvoyPresent,
	}
}

// parseInstanceID parses an instance ID in the format "cluster_id:namespace:pod_name"
// Returns cluster ID, namespace, pod name, and any error
func parseInstanceID(instanceID string) (clusterID, namespace, podName string, err error) {
	parts := strings.Split(instanceID, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid instance ID format, expected 'cluster_id:namespace:pod_name', got: %s", instanceID)
	}

	clusterID = parts[0]
	namespace = parts[1]
	podName = parts[2]

	if clusterID == "" || namespace == "" || podName == "" {
		return "", "", "", fmt.Errorf("instance ID contains empty components: %s", instanceID)
	}

	return clusterID, namespace, podName, nil
}

// extractServiceNameFromInstanceID extracts the service name from an instance ID
// This is a helper function since we don't store service name directly in the instance
func extractServiceNameFromInstanceID(instanceID string) string {
	_, _, podName, err := parseInstanceID(instanceID)
	if err != nil {
		return "unknown"
	}
	// For now, return pod name as service identifier
	// In a more complete implementation, this would require a service lookup
	return podName
}

// convertConnectionInfoToClusterSyncInfo converts a ConnectionInfo to the frontend API format
func (f *FrontendService) convertConnectionInfoToClusterSyncInfo(connInfo connections.ConnectionInfo) *frontendv1alpha1.ClusterSyncInfo {
	// Safe conversion from int to int32 to avoid overflow
	var serviceCount int32
	if connInfo.ServiceCount > 2147483647 { // max int32
		serviceCount = 2147483647
	} else if connInfo.ServiceCount < -2147483648 { // min int32
		serviceCount = -2147483648
	} else {
		serviceCount = int32(connInfo.ServiceCount) // #nosec G115 - bounds checked above
	}

	return &frontendv1alpha1.ClusterSyncInfo{
		ClusterId:    connInfo.ClusterID,
		ConnectedAt:  connInfo.ConnectedAt.Format(time.RFC3339),
		LastUpdate:   connInfo.LastUpdate.Format(time.RFC3339),
		ServiceCount: serviceCount,
		SyncStatus:   f.computeSyncStatus(connInfo.LastUpdate),
	}
}

// computeSyncStatus determines the sync health based on last update time
func (f *FrontendService) computeSyncStatus(lastUpdate time.Time) frontendv1alpha1.SyncStatus {
	timeSince := time.Since(lastUpdate)

	switch {
	case timeSince < 30*time.Second:
		return frontendv1alpha1.SyncStatus_SYNC_STATUS_HEALTHY
	case timeSince < 5*time.Minute:
		return frontendv1alpha1.SyncStatus_SYNC_STATUS_STALE
	default:
		return frontendv1alpha1.SyncStatus_SYNC_STATUS_DISCONNECTED
	}
}
