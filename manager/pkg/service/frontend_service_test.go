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
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
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

func (m *MockConnectionManager) RegisterConnection(clusterID string, stream backendv1alpha1.ManagerService_ConnectServer) error {
	args := m.Called(clusterID, stream)
	return args.Error(0)
}

func (m *MockConnectionManager) UnregisterConnection(clusterID string) {
	m.Called(clusterID)
}

func (m *MockConnectionManager) UpdateClusterState(clusterID string, clusterState *backendv1alpha1.ClusterState) error {
	args := m.Called(clusterID, clusterState)
	return args.Error(0)
}

func (m *MockConnectionManager) GetClusterState(clusterID string) (*backendv1alpha1.ClusterState, error) {
	args := m.Called(clusterID)
	return args.Get(0).(*backendv1alpha1.ClusterState), args.Error(1)
}

func (m *MockConnectionManager) GetAllClusterStates() map[string]*backendv1alpha1.ClusterState {
	args := m.Called()
	return args.Get(0).(map[string]*backendv1alpha1.ClusterState)
}

func (m *MockConnectionManager) IsClusterConnected(clusterID string) bool {
	args := m.Called(clusterID)
	return args.Bool(0)
}

func (m *MockConnectionManager) GetActiveClusterCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockConnectionManager) SendMessageToCluster(clusterID string, message *backendv1alpha1.ConnectResponse) error {
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

func (m *MockProxyService) HandleProxyConfigResponse(response *backendv1alpha1.ProxyConfigResponse) error {
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

func (m *MockIstioService) GetIstioResourcesForWorkload(ctx context.Context, clusterID, namespace string, instance *backendv1alpha1.ServiceInstance) (*frontendv1alpha1.GetIstioResourcesResponse, error) {
	args := m.Called(ctx, clusterID, namespace, instance)
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
	mockIstioService.On("GetIstioResourcesForWorkload", mock.Anything, "cluster1", "default", mock.MatchedBy(func(instance *backendv1alpha1.ServiceInstance) bool {
		return instance.Labels["app"] == "test"
	})).Return(testIstioResources, nil)

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

func TestFrontendService_GetIstioResources_WithPeerAuthentications(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	// Create test data with peer authentications
	testInstance := &connections.AggregatedServiceInstance{
		InstanceID:   "cluster1:production:web-pod",
		IP:           "10.0.0.2",
		PodName:      "web-pod",
		Namespace:    "production",
		ClusterName:  "cluster1",
		EnvoyPresent: true,
		Labels:       map[string]string{"app": "web", "version": "v1", "tier": "frontend"},
	}

	testIstioResources := &frontendv1alpha1.GetIstioResourcesResponse{
		VirtualServices: []*types.VirtualService{
			{
				Name:      "web-vs",
				Namespace: "production",
				Hosts:     []string{"web.production.svc.cluster.local"},
			},
		},
		Gateways: []*types.Gateway{
			{
				Name:      "web-gateway",
				Namespace: "production",
				Selector:  map[string]string{"app": "web"},
			},
		},
		RequestAuthentications: []*types.RequestAuthentication{
			{
				Name:      "web-request-auth",
				Namespace: "production",
				Selector: &types.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
		},
		PeerAuthentications: []*types.PeerAuthentication{
			{
				Name:      "web-peer-auth-strict",
				Namespace: "production",
				Selector: &types.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "version": "v1"},
				},
			},
			{
				Name:      "frontend-tier-peer-auth",
				Namespace: "production",
				Selector: &types.WorkloadSelector{
					MatchLabels: map[string]string{"tier": "frontend"},
				},
			},
			{
				Name:      "default-peer-auth",
				Namespace: "production",
				Selector:  nil, // nil selector matches all workloads in namespace
			},
		},
	}

	mockConnManager.On("GetAggregatedServiceInstance", "cluster1:production:web-pod").Return(testInstance, true)
	mockProxyService := new(MockProxyService)
	mockIstioService := new(MockIstioService)
	mockIstioService.On("GetIstioResourcesForWorkload", mock.Anything, "cluster1", "production", mock.MatchedBy(func(instance *backendv1alpha1.ServiceInstance) bool {
		return instance.Labels["app"] == "web" && instance.Labels["version"] == "v1" && instance.Labels["tier"] == "frontend"
	})).Return(testIstioResources, nil)

	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetIstioResources
	req := &frontendv1alpha1.GetIstioResourcesRequest{
		ServiceId:  "production:web-service",
		InstanceId: "cluster1:production:web-pod",
	}
	resp, err := frontendService.GetIstioResources(context.Background(), req)

	// Verify the response
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify VirtualServices
	assert.Len(t, resp.VirtualServices, 1)
	assert.Equal(t, "web-vs", resp.VirtualServices[0].Name)

	// Verify Gateways
	assert.Len(t, resp.Gateways, 1)
	assert.Equal(t, "web-gateway", resp.Gateways[0].Name)

	// Verify RequestAuthentications
	assert.Len(t, resp.RequestAuthentications, 1)
	assert.Equal(t, "web-request-auth", resp.RequestAuthentications[0].Name)
	assert.Equal(t, "production", resp.RequestAuthentications[0].Namespace)
	assert.NotNil(t, resp.RequestAuthentications[0].Selector)
	assert.Equal(t, "web", resp.RequestAuthentications[0].Selector.MatchLabels["app"])

	// Verify PeerAuthentications - this is the key test for our new functionality
	assert.Len(t, resp.PeerAuthentications, 3)

	// Find specific peer authentications by name
	var strictPeerAuth, tierPeerAuth, defaultPeerAuth *types.PeerAuthentication
	for _, pa := range resp.PeerAuthentications {
		switch pa.Name {
		case "web-peer-auth-strict":
			strictPeerAuth = pa
		case "frontend-tier-peer-auth":
			tierPeerAuth = pa
		case "default-peer-auth":
			defaultPeerAuth = pa
		}
	}

	// Verify specific peer authentication with multi-label selector
	assert.NotNil(t, strictPeerAuth)
	assert.Equal(t, "production", strictPeerAuth.Namespace)
	assert.NotNil(t, strictPeerAuth.Selector)
	assert.Equal(t, "web", strictPeerAuth.Selector.MatchLabels["app"])
	assert.Equal(t, "v1", strictPeerAuth.Selector.MatchLabels["version"])

	// Verify tier-based peer authentication
	assert.NotNil(t, tierPeerAuth)
	assert.Equal(t, "production", tierPeerAuth.Namespace)
	assert.NotNil(t, tierPeerAuth.Selector)
	assert.Equal(t, "frontend", tierPeerAuth.Selector.MatchLabels["tier"])

	// Verify default peer authentication (nil selector)
	assert.NotNil(t, defaultPeerAuth)
	assert.Equal(t, "production", defaultPeerAuth.Namespace)
	assert.Nil(t, defaultPeerAuth.Selector) // nil selector matches all workloads

	// Verify all mocks were called as expected
	mockConnManager.AssertExpectations(t)
	mockIstioService.AssertExpectations(t)
}

func TestFrontendService_GetIstioResources_WithWasmPlugins(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	// Create test data with WasmPlugins
	testInstance := &connections.AggregatedServiceInstance{
		InstanceID:   "cluster1:production:web-pod",
		IP:           "10.0.0.2",
		PodName:      "web-pod",
		Namespace:    "production",
		ClusterName:  "cluster1",
		EnvoyPresent: true,
		Labels:       map[string]string{"app": "web", "version": "v1", "tier": "frontend"},
	}

	testIstioResources := &frontendv1alpha1.GetIstioResourcesResponse{
		VirtualServices: []*types.VirtualService{
			{
				Name:      "web-vs",
				Namespace: "production",
				Hosts:     []string{"web.production.svc.cluster.local"},
			},
		},
		Gateways: []*types.Gateway{
			{
				Name:      "web-gateway",
				Namespace: "production",
				Selector:  map[string]string{"app": "web"},
			},
		},
		WasmPlugins: []*types.WasmPlugin{
			{
				Name:      "auth-wasm-plugin",
				Namespace: "production",
				Selector: &types.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
				TargetRefs: nil,
			},
			{
				Name:      "metrics-wasm-plugin",
				Namespace: "production",
				Selector: &types.WorkloadSelector{
					MatchLabels: map[string]string{"tier": "frontend"},
				},
				TargetRefs: nil,
			},
			{
				Name:      "gateway-wasm-plugin",
				Namespace: "istio-system",
				Selector:  nil,
				TargetRefs: []*types.PolicyTargetReference{
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "web-gateway",
					},
				},
			},
			{
				Name:       "global-wasm-plugin",
				Namespace:  "istio-system",
				Selector:   nil,
				TargetRefs: nil,
			},
		},
		RequestAuthentications: []*types.RequestAuthentication{
			{
				Name:      "web-request-auth",
				Namespace: "production",
				Selector: &types.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
		},
	}

	// Set up mock expectations
	mockConnManager.On("GetAggregatedServiceInstance", "cluster1:production:web-pod").Return(testInstance, true)

	// Mock proxy service (not called by GetIstioResources)
	mockProxyService := new(MockProxyService)

	// Mock istio service to filter WasmPlugins for the workload
	mockIstioService := new(MockIstioService)
	mockIstioService.On("GetIstioResourcesForWorkload",
		context.Background(),
		"cluster1",
		"production",
		mock.MatchedBy(func(instance *backendv1alpha1.ServiceInstance) bool {
			return instance.Labels["app"] == "web" &&
				instance.Labels["version"] == "v1" &&
				instance.Labels["tier"] == "frontend"
		})).Return(testIstioResources, nil)

	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetIstioResources
	req := &frontendv1alpha1.GetIstioResourcesRequest{
		ServiceId:  "production:web-service",
		InstanceId: "cluster1:production:web-pod",
	}
	resp, err := frontendService.GetIstioResources(context.Background(), req)

	// Verify the response
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify VirtualServices
	assert.Len(t, resp.VirtualServices, 1)
	assert.Equal(t, "web-vs", resp.VirtualServices[0].Name)

	// Verify Gateways
	assert.Len(t, resp.Gateways, 1)
	assert.Equal(t, "web-gateway", resp.Gateways[0].Name)

	// Verify RequestAuthentications (regression test)
	assert.Len(t, resp.RequestAuthentications, 1)
	assert.Equal(t, "web-request-auth", resp.RequestAuthentications[0].Name)

	// Verify WasmPlugins - this is the key test for our new functionality
	assert.Len(t, resp.WasmPlugins, 4)

	// Find specific wasm plugins by name
	var authPlugin, metricsPlugin, gatewayPlugin, globalPlugin *types.WasmPlugin
	for _, wp := range resp.WasmPlugins {
		switch wp.Name {
		case "auth-wasm-plugin":
			authPlugin = wp
		case "metrics-wasm-plugin":
			metricsPlugin = wp
		case "gateway-wasm-plugin":
			gatewayPlugin = wp
		case "global-wasm-plugin":
			globalPlugin = wp
		}
	}

	// Verify app-specific wasm plugin
	assert.NotNil(t, authPlugin)
	assert.Equal(t, "production", authPlugin.Namespace)
	assert.NotNil(t, authPlugin.Selector)
	assert.Equal(t, "web", authPlugin.Selector.MatchLabels["app"])
	assert.Nil(t, authPlugin.TargetRefs)

	// Verify tier-specific wasm plugin
	assert.NotNil(t, metricsPlugin)
	assert.Equal(t, "production", metricsPlugin.Namespace)
	assert.NotNil(t, metricsPlugin.Selector)
	assert.Equal(t, "frontend", metricsPlugin.Selector.MatchLabels["tier"])
	assert.Nil(t, metricsPlugin.TargetRefs)

	// Verify gateway-targeted wasm plugin from istio-system
	assert.NotNil(t, gatewayPlugin)
	assert.Equal(t, "istio-system", gatewayPlugin.Namespace)
	assert.Nil(t, gatewayPlugin.Selector)
	assert.Len(t, gatewayPlugin.TargetRefs, 1)
	assert.Equal(t, "Gateway", gatewayPlugin.TargetRefs[0].Kind)
	assert.Equal(t, "web-gateway", gatewayPlugin.TargetRefs[0].Name)

	// Verify global wasm plugin from root namespace
	assert.NotNil(t, globalPlugin)
	assert.Equal(t, "istio-system", globalPlugin.Namespace)
	assert.Nil(t, globalPlugin.Selector)
	assert.Nil(t, globalPlugin.TargetRefs)

	// Verify all mocks were called as expected
	mockConnManager.AssertExpectations(t)
	mockProxyService.AssertExpectations(t)
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

func TestFrontendService_GetIstioResources_WithServiceEntries(t *testing.T) {
	logger := logging.For("test")
	mockConnManager := new(MockConnectionManager)

	// Create test data with ServiceEntries
	testInstance := &connections.AggregatedServiceInstance{
		InstanceID:   "cluster1:default:app-pod",
		IP:           "10.0.0.3",
		PodName:      "app-pod",
		Namespace:    "default",
		ClusterName:  "cluster1",
		EnvoyPresent: true,
		Labels:       map[string]string{"app": "myapp", "version": "v1"},
	}

	testIstioResources := &frontendv1alpha1.GetIstioResourcesResponse{
		VirtualServices: []*types.VirtualService{
			{
				Name:      "app-vs",
				Namespace: "default",
				Hosts:     []string{"myapp.default.svc.cluster.local"},
			},
		},
		ServiceEntries: []*types.ServiceEntry{
			{
				Name:      "external-api",
				Namespace: "istio-system",
				ExportTo:  []string{"*"},
			},
			{
				Name:      "shared-database",
				Namespace: "default",
				ExportTo:  []string{".", "production"},
			},
			{
				Name:      "team-service",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
		},
	}

	// Set up mock expectations
	mockConnManager.On("GetAggregatedServiceInstance", "cluster1:default:app-pod").Return(testInstance, true)

	// Mock proxy service (not called by GetIstioResources)
	mockProxyService := new(MockProxyService)

	// Mock istio service to filter ServiceEntries for the workload
	mockIstioService := new(MockIstioService)
	mockIstioService.On("GetIstioResourcesForWorkload",
		context.Background(),
		"cluster1",
		"default",
		mock.MatchedBy(func(instance *backendv1alpha1.ServiceInstance) bool {
			return instance.Labels["app"] == "myapp" &&
				instance.Labels["version"] == "v1"
		})).Return(testIstioResources, nil)

	frontendService := NewFrontendService(mockConnManager, mockProxyService, mockIstioService, logger)

	// Test GetIstioResources
	req := &frontendv1alpha1.GetIstioResourcesRequest{
		ServiceId:  "default:myapp-service",
		InstanceId: "cluster1:default:app-pod",
	}

	resp, err := frontendService.GetIstioResources(context.Background(), req)

	// Verify the response
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify VirtualServices
	assert.Len(t, resp.VirtualServices, 1)
	assert.Equal(t, "app-vs", resp.VirtualServices[0].Name)

	// Verify ServiceEntries - this is the key test for our new functionality
	assert.Len(t, resp.ServiceEntries, 3)

	// Test each service entry
	serviceEntryNames := make([]string, len(resp.ServiceEntries))
	for i, se := range resp.ServiceEntries {
		serviceEntryNames[i] = se.Name
	}
	assert.Contains(t, serviceEntryNames, "external-api")
	assert.Contains(t, serviceEntryNames, "shared-database")
	assert.Contains(t, serviceEntryNames, "team-service")

	// Verify specific ServiceEntry details
	for _, se := range resp.ServiceEntries {
		switch se.Name {
		case "external-api":
			assert.Equal(t, "istio-system", se.Namespace)
			assert.Equal(t, []string{"*"}, se.ExportTo)
		case "shared-database":
			assert.Equal(t, "default", se.Namespace)
			assert.Equal(t, []string{".", "production"}, se.ExportTo)
		case "team-service":
			assert.Equal(t, "production", se.Namespace)
			assert.Equal(t, []string{"default", "staging"}, se.ExportTo)
		}
	}

	// Verify all mocks were called as expected
	mockConnManager.AssertExpectations(t)
	mockProxyService.AssertExpectations(t)
	mockIstioService.AssertExpectations(t)
}
