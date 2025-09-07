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

func (m *MockConnectionManager) UpdateCapabilities(clusterID string, capabilities *backendv1alpha1.EdgeCapabilities) error {
	args := m.Called(clusterID, capabilities)
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

func TestServiceRegistryService_ListServices(t *testing.T) {
	mockConnManager := &MockConnectionManager{}
	mockProxyService := &MockProxyService{}
	mockIstioService := &MockIstioService{}

	service := NewServiceRegistryService(mockConnManager, mockProxyService, mockIstioService, logging.For("test"))

	// Mock data
	aggregatedServices := []*connections.AggregatedService{
		{
			ID:        "test-namespace:test-service",
			Name:      "test-service",
			Namespace: "test-namespace",
			Instances: []*connections.AggregatedServiceInstance{},
		},
	}

	mockConnManager.On("ListAggregatedServices", "", "").Return(aggregatedServices)

	req := &frontendv1alpha1.ListServicesRequest{}
	resp, err := service.ListServices(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Services, 1)
	assert.Equal(t, "test-service", resp.Services[0].Name)
	assert.Equal(t, "test-namespace", resp.Services[0].Namespace)

	mockConnManager.AssertExpectations(t)
}

func TestServiceRegistryService_GetService_Success(t *testing.T) {
	mockConnManager := &MockConnectionManager{}
	mockProxyService := &MockProxyService{}
	mockIstioService := &MockIstioService{}

	service := NewServiceRegistryService(mockConnManager, mockProxyService, mockIstioService, logging.For("test"))

	// Mock data
	aggregatedService := &connections.AggregatedService{
		ID:        "test-namespace:test-service",
		Name:      "test-service",
		Namespace: "test-namespace",
		Instances: []*connections.AggregatedServiceInstance{},
	}

	mockConnManager.On("GetAggregatedService", "test-namespace:test-service").Return(aggregatedService, true)

	req := &frontendv1alpha1.GetServiceRequest{Id: "test-namespace:test-service"}
	resp, err := service.GetService(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-service", resp.Service.Name)
	assert.Equal(t, "test-namespace", resp.Service.Namespace)

	mockConnManager.AssertExpectations(t)
}

func TestServiceRegistryService_GetService_NotFound(t *testing.T) {
	mockConnManager := &MockConnectionManager{}
	mockProxyService := &MockProxyService{}
	mockIstioService := &MockIstioService{}

	service := NewServiceRegistryService(mockConnManager, mockProxyService, mockIstioService, logging.For("test"))

	// Mock returning not found
	var nilService *connections.AggregatedService
	mockConnManager.On("GetAggregatedService", "nonexistent:service").Return(nilService, false)

	req := &frontendv1alpha1.GetServiceRequest{Id: "nonexistent:service"}
	resp, err := service.GetService(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check that it's a NotFound error
	statusErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, statusErr.Code())

	mockConnManager.AssertExpectations(t)
}
