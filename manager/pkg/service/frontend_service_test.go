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
	"testing"

	"github.com/liamawhite/navigator/manager/pkg/connections"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockConnectionManager for testing
type MockConnectionManager struct {
	mock.Mock
}

func (m *MockConnectionManager) RegisterConnection(clusterID string, stream v1alpha1.ManagerService_ConnectServer) error {
	args := m.Called(clusterID, stream)
	return args.Error(0)
}

func (m *MockConnectionManager) UnregisterConnection(clusterID string) {
	m.Called(clusterID)
}

func (m *MockConnectionManager) UpdateClusterState(clusterID string, clusterState *v1alpha1.ClusterState) error {
	args := m.Called(clusterID, clusterState)
	return args.Error(0)
}

func (m *MockConnectionManager) GetClusterState(clusterID string) (*v1alpha1.ClusterState, error) {
	args := m.Called(clusterID)
	return args.Get(0).(*v1alpha1.ClusterState), args.Error(1)
}

func (m *MockConnectionManager) GetAllClusterStates() map[string]*v1alpha1.ClusterState {
	args := m.Called()
	return args.Get(0).(map[string]*v1alpha1.ClusterState)
}

func (m *MockConnectionManager) IsClusterConnected(clusterID string) bool {
	args := m.Called(clusterID)
	return args.Bool(0)
}

func (m *MockConnectionManager) GetActiveClusterCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockConnectionManager) SendMessageToCluster(clusterID string, message *v1alpha1.ConnectResponse) error {
	args := m.Called(clusterID, message)
	return args.Error(0)
}

// Read-optimized methods
func (m *MockConnectionManager) ListAggregatedServices(namespace, clusterID string) []*connections.AggregatedService {
	args := m.Called(namespace, clusterID)
	return args.Get(0).([]*connections.AggregatedService)
}

func (m *MockConnectionManager) GetAggregatedService(serviceID string) (*connections.AggregatedService, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*connections.AggregatedService), args.Bool(1)
}

func (m *MockConnectionManager) GetAggregatedServiceInstance(instanceID string) (*connections.AggregatedServiceInstance, bool) {
	args := m.Called(instanceID)
	return args.Get(0).(*connections.AggregatedServiceInstance), args.Bool(1)
}

func (m *MockConnectionManager) GetConnectionInfo() map[string]connections.ConnectionInfo {
	args := m.Called()
	return args.Get(0).(map[string]connections.ConnectionInfo)
}

// MockProxyService for testing
type MockProxyService struct {
	mock.Mock
}

func (m *MockProxyService) GetProxyConfig(ctx context.Context, clusterID, namespace, podName string) (*types.ProxyConfig, error) {
	args := m.Called(ctx, clusterID, namespace, podName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ProxyConfig), args.Error(1)
}

func (m *MockProxyService) HandleProxyConfigResponse(response *v1alpha1.ProxyConfigResponse) error {
	args := m.Called(response)
	return args.Error(0)
}

func (m *MockProxyService) GetPendingRequestCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockProxyService) CleanupExpiredRequests() {
	m.Called()
}

// MockIstioService for testing
type MockIstioService struct {
	mock.Mock
}

func (m *MockIstioService) GetIstioResourcesForWorkload(ctx context.Context, clusterID, namespace string, labels map[string]string) (*frontendv1alpha1.GetIstioResourcesResponse, error) {
	args := m.Called(ctx, clusterID, namespace, labels)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*frontendv1alpha1.GetIstioResourcesResponse), args.Error(1)
}

func TestFrontendService_ListServices(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	// Create test data
	testService := &connections.AggregatedService{
		ID:        "default:test-service",
		Name:      "test-service",
		Namespace: "default",
		Instances: []*connections.AggregatedServiceInstance{
			{
				InstanceID:   "cluster1:default:pod1",
				IP:           "10.0.0.1",
				PodName:      "pod1",
				Namespace:    "default",
				ClusterName:  "cluster1",
				EnvoyPresent: true,
			},
		},
	}

	mockConnManager.On("ListAggregatedServices", "", "").Return([]*connections.AggregatedService{testService})

	mockProxyService := new(MockProxyService)
	mockIstioService := new(MockIstioService)
	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test ListServices
	req := &frontendv1alpha1.ListServicesRequest{}
	resp, err := frontendService.ListServices(context.Background(), req)

	assert.NoError(t, err, "ListServices should not return error")
	assert.Len(t, resp.Services, 1, "Should return 1 service")
	assert.Equal(t, "default:test-service", resp.Services[0].Id, "Service ID should match")
	assert.Equal(t, "test-service", resp.Services[0].Name, "Service name should match")
	assert.Equal(t, "default", resp.Services[0].Namespace, "Service namespace should match")
	assert.Len(t, resp.Services[0].Instances, 1, "Should have 1 instance")

	mockConnManager.AssertExpectations(t)
}

func TestFrontendService_GetService(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	// Create test data
	testService := &connections.AggregatedService{
		ID:        "default:test-service",
		Name:      "test-service",
		Namespace: "default",
		Instances: []*connections.AggregatedServiceInstance{
			{
				InstanceID:   "cluster1:default:pod1",
				IP:           "10.0.0.1",
				PodName:      "pod1",
				Namespace:    "default",
				ClusterName:  "cluster1",
				EnvoyPresent: true,
			},
		},
	}

	mockConnManager.On("GetAggregatedService", "default:test-service").Return(testService, true)

	mockProxyService := new(MockProxyService)
	mockIstioService := new(MockIstioService)
	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetService
	req := &frontendv1alpha1.GetServiceRequest{Id: "default:test-service"}
	resp, err := frontendService.GetService(context.Background(), req)

	assert.NoError(t, err, "GetService should not return error")
	assert.Equal(t, "default:test-service", resp.Service.Id, "Service ID should match")
	assert.Equal(t, "test-service", resp.Service.Name, "Service name should match")
	assert.Len(t, resp.Service.Instances, 1, "Should have 1 instance")

	mockConnManager.AssertExpectations(t)
}

func TestFrontendService_GetService_NotFound(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	mockConnManager.On("GetAggregatedService", "default:nonexistent").Return((*connections.AggregatedService)(nil), false)

	mockProxyService := new(MockProxyService)
	mockIstioService := new(MockIstioService)
	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetService with non-existent service
	req := &frontendv1alpha1.GetServiceRequest{Id: "default:nonexistent"}
	resp, err := frontendService.GetService(context.Background(), req)

	assert.Error(t, err, "GetService should return error for non-existent service")
	assert.Nil(t, resp, "Response should be nil for non-existent service")

	// Check that it's a NotFound error
	st, ok := status.FromError(err)
	assert.True(t, ok, "Error should be a gRPC status error")
	assert.Equal(t, codes.NotFound, st.Code(), "Error code should be NotFound")

	mockConnManager.AssertExpectations(t)
}

func TestFrontendService_GetServiceInstance(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	// Create test data
	testInstance := &connections.AggregatedServiceInstance{
		InstanceID:   "cluster1:default:pod1",
		IP:           "10.0.0.1",
		PodName:      "pod1",
		Namespace:    "default",
		ClusterName:  "cluster1",
		EnvoyPresent: true,
	}

	mockConnManager.On("GetAggregatedServiceInstance", "cluster1:default:pod1").Return(testInstance, true)

	mockProxyService := new(MockProxyService)
	mockIstioService := new(MockIstioService)
	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetServiceInstance
	req := &frontendv1alpha1.GetServiceInstanceRequest{
		ServiceId:  "default:test-service",
		InstanceId: "cluster1:default:pod1",
	}
	resp, err := frontendService.GetServiceInstance(context.Background(), req)

	assert.NoError(t, err, "GetServiceInstance should not return error")
	assert.Equal(t, "cluster1:default:pod1", resp.Instance.InstanceId, "Instance ID should match")
	assert.Equal(t, "10.0.0.1", resp.Instance.Ip, "Instance IP should match")
	assert.Equal(t, "pod1", resp.Instance.PodName, "Pod name should match")
	assert.Equal(t, "cluster1", resp.Instance.ClusterName, "Cluster name should match")
	assert.True(t, resp.Instance.EnvoyPresent, "Envoy present should be true")

	mockConnManager.AssertExpectations(t)
}

func TestFrontendService_GetProxyConfig(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)
	mockProxyService := new(MockProxyService)

	// Create test data
	testInstance := &connections.AggregatedServiceInstance{
		InstanceID:   "cluster1:default:pod1",
		IP:           "10.0.0.1",
		PodName:      "pod1",
		Namespace:    "default",
		ClusterName:  "cluster1",
		EnvoyPresent: true,
	}

	testProxyConfig := &types.ProxyConfig{
		Version:   "1.0.0",
		Bootstrap: &types.BootstrapSummary{},
	}

	mockConnManager.On("GetAggregatedServiceInstance", "cluster1:default:pod1").Return(testInstance, true)
	mockProxyService.On("GetProxyConfig", mock.Anything, "cluster1", "default", "pod1").Return(testProxyConfig, nil)

	mockIstioService := new(MockIstioService)
	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetProxyConfig
	req := &frontendv1alpha1.GetProxyConfigRequest{
		ServiceId:  "default:test-service",
		InstanceId: "cluster1:default:pod1",
	}
	resp, err := frontendService.GetProxyConfig(context.Background(), req)

	assert.NoError(t, err, "GetProxyConfig should not return error")
	assert.NotNil(t, resp, "Response should not be nil")
	assert.NotNil(t, resp.ProxyConfig, "Proxy config should not be nil")
	assert.Equal(t, "1.0.0", resp.ProxyConfig.Version, "Proxy config version should match")

	mockConnManager.AssertExpectations(t)
	mockProxyService.AssertExpectations(t)
}

func TestFrontendService_GetIstioResources(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	// Create test data
	testInstance := &connections.AggregatedServiceInstance{
		InstanceID:   "cluster1:default:pod1",
		IP:           "10.0.0.1",
		PodName:      "pod1",
		Namespace:    "default",
		ClusterName:  "cluster1",
		EnvoyPresent: true,
		Labels:       map[string]string{"app": "test", "version": "v1"},
	}

	testIstioResources := &frontendv1alpha1.GetIstioResourcesResponse{
		VirtualServices: []*types.VirtualService{
			{
				Name:      "test-vs",
				Namespace: "default",
				Hosts:     []string{"test.example.com"},
			},
		},
		Gateways: []*types.Gateway{
			{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
		},
	}

	mockConnManager.On("GetAggregatedServiceInstance", "cluster1:default:pod1").Return(testInstance, true)
	mockProxyService := new(MockProxyService)
	mockIstioService := new(MockIstioService)
	mockIstioService.On("GetIstioResourcesForWorkload", mock.Anything, "cluster1", "default", testInstance.Labels).Return(testIstioResources, nil)

	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetIstioResources
	req := &frontendv1alpha1.GetIstioResourcesRequest{
		ServiceId:  "default:test-service",
		InstanceId: "cluster1:default:pod1",
	}
	resp, err := frontendService.GetIstioResources(context.Background(), req)

	// Verify the response
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.VirtualServices, 1)
	assert.Equal(t, "test-vs", resp.VirtualServices[0].Name)
	assert.Len(t, resp.Gateways, 1)
	assert.Equal(t, "test-gateway", resp.Gateways[0].Name)

	// Verify all mocks were called as expected
	mockConnManager.AssertExpectations(t)
	mockIstioService.AssertExpectations(t)
}

func TestFrontendService_GetIstioResources_InstanceNotFound(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	mockConnManager.On("GetAggregatedServiceInstance", "cluster1:default:nonexistent").Return((*connections.AggregatedServiceInstance)(nil), false)

	mockProxyService := new(MockProxyService)
	mockIstioService := new(MockIstioService)
	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetIstioResources with non-existent instance
	req := &frontendv1alpha1.GetIstioResourcesRequest{
		ServiceId:  "default:test-service",
		InstanceId: "cluster1:default:nonexistent",
	}
	resp, err := frontendService.GetIstioResources(context.Background(), req)

	// Verify error response
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check that it's a NotFound error
	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, grpcErr.Code())

	// Verify all mocks were called as expected
	mockConnManager.AssertExpectations(t)
}
