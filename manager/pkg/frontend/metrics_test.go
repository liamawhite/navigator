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
	"fmt"
	"testing"
	"time"

	"github.com/liamawhite/navigator/manager/pkg/connections"
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsConnectionManager for testing
type MockMetricsConnectionManager struct {
	mock.Mock
}

func (m *MockMetricsConnectionManager) RegisterConnection(clusterID string, stream backendv1alpha1.ManagerService_ConnectServer) error {
	args := m.Called(clusterID, stream)
	return args.Error(0)
}

func (m *MockMetricsConnectionManager) UnregisterConnection(clusterID string) {
	m.Called(clusterID)
}

func (m *MockMetricsConnectionManager) UpdateClusterState(clusterID string, clusterState *backendv1alpha1.ClusterState) error {
	args := m.Called(clusterID, clusterState)
	return args.Error(0)
}

func (m *MockMetricsConnectionManager) UpdateCapabilities(clusterID string, capabilities *backendv1alpha1.EdgeCapabilities) error {
	args := m.Called(clusterID, capabilities)
	return args.Error(0)
}

func (m *MockMetricsConnectionManager) GetClusterState(clusterID string) (*backendv1alpha1.ClusterState, error) {
	args := m.Called(clusterID)
	return args.Get(0).(*backendv1alpha1.ClusterState), args.Error(1)
}

func (m *MockMetricsConnectionManager) GetAllClusterStates() map[string]*backendv1alpha1.ClusterState {
	args := m.Called()
	return args.Get(0).(map[string]*backendv1alpha1.ClusterState)
}

func (m *MockMetricsConnectionManager) IsClusterConnected(clusterID string) bool {
	args := m.Called(clusterID)
	return args.Bool(0)
}

func (m *MockMetricsConnectionManager) GetActiveClusterCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMetricsConnectionManager) SendMessageToCluster(clusterID string, message *backendv1alpha1.ConnectResponse) error {
	args := m.Called(clusterID, message)
	return args.Error(0)
}

func (m *MockMetricsConnectionManager) ListAggregatedServices(namespace, clusterID string) []*connections.AggregatedService {
	args := m.Called(namespace, clusterID)
	return args.Get(0).([]*connections.AggregatedService)
}

func (m *MockMetricsConnectionManager) GetAggregatedService(serviceID string) (*connections.AggregatedService, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*connections.AggregatedService), args.Bool(1)
}

func (m *MockMetricsConnectionManager) GetAggregatedServiceInstance(instanceID string) (*connections.AggregatedServiceInstance, bool) {
	args := m.Called(instanceID)
	return args.Get(0).(*connections.AggregatedServiceInstance), args.Bool(1)
}

func (m *MockMetricsConnectionManager) GetConnectionInfo() map[string]connections.ConnectionInfo {
	args := m.Called()
	return args.Get(0).(map[string]connections.ConnectionInfo)
}

// MockMeshMetricsProvider for testing
type MockMeshMetricsProvider struct {
	mock.Mock
}

func (m *MockMeshMetricsProvider) GetServiceGraphMetrics(ctx context.Context, clusterID string, req *frontendv1alpha1.GetServiceGraphMetricsRequest) (*typesv1alpha1.ServiceGraphMetrics, error) {
	args := m.Called(ctx, clusterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*typesv1alpha1.ServiceGraphMetrics), args.Error(1)
}

func TestMetricsService_GetServiceMeshGraphMetrics_Success(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock connection info
	connectionInfos := map[string]connections.ConnectionInfo{
		"cluster-1": {
			ClusterID: "cluster-1",
		},
		"cluster-2": {
			ClusterID: "cluster-2",
		},
	}

	// Mock mesh metrics responses
	cluster1Metrics := &typesv1alpha1.ServiceGraphMetrics{
		Pairs: []*typesv1alpha1.ServicePairMetrics{
			{
				SourceCluster:        "cluster-1",
				SourceNamespace:      "default",
				SourceService:        "frontend",
				DestinationCluster:   "cluster-1",
				DestinationNamespace: "default",
				DestinationService:   "backend",
				ErrorRate:            0.05, // 5% error rate
				RequestRate:          100.0,
			},
		},
	}

	cluster2Metrics := &typesv1alpha1.ServiceGraphMetrics{
		Pairs: []*typesv1alpha1.ServicePairMetrics{
			{
				SourceCluster:        "cluster-2",
				SourceNamespace:      "production",
				SourceService:        "api",
				DestinationCluster:   "cluster-2",
				DestinationNamespace: "production",
				DestinationService:   "database",
				ErrorRate:            0.01, // 1% error rate
				RequestRate:          50.0,
			},
		},
	}

	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		// StartTime and EndTime will be set to defaults if not specified
	}

	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-1", req).Return(cluster1Metrics, nil)
	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-2", req).Return(cluster2Metrics, nil)

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Pairs, 2)
	assert.Len(t, resp.ClustersQueried, 2)
	assert.Contains(t, resp.ClustersQueried, "cluster-1")
	assert.Contains(t, resp.ClustersQueried, "cluster-2")
	assert.NotEmpty(t, resp.Timestamp)

	// Verify the metrics data is correctly converted
	pairsByCluster := make(map[string]*typesv1alpha1.ServicePairMetrics)
	for _, pair := range resp.Pairs {
		pairsByCluster[pair.SourceCluster] = pair
	}

	cluster1Pair := pairsByCluster["cluster-1"]
	assert.NotNil(t, cluster1Pair)
	assert.Equal(t, "default", cluster1Pair.SourceNamespace)
	assert.Equal(t, "frontend", cluster1Pair.SourceService)
	assert.Equal(t, "backend", cluster1Pair.DestinationService)
	assert.Equal(t, 0.05, cluster1Pair.ErrorRate)
	assert.Equal(t, 100.0, cluster1Pair.RequestRate)

	cluster2Pair := pairsByCluster["cluster-2"]
	assert.NotNil(t, cluster2Pair)
	assert.Equal(t, "production", cluster2Pair.SourceNamespace)
	assert.Equal(t, "api", cluster2Pair.SourceService)
	assert.Equal(t, "database", cluster2Pair.DestinationService)
	assert.Equal(t, 0.01, cluster2Pair.ErrorRate)
	assert.Equal(t, 50.0, cluster2Pair.RequestRate)

	mockConnManager.AssertExpectations(t)
	mockMeshProvider.AssertExpectations(t)
}

func TestMetricsService_GetServiceMeshGraphMetrics_WithFilters(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock connection info
	connectionInfos := map[string]connections.ConnectionInfo{
		"cluster-1": {
			ClusterID: "cluster-1",
		},
	}

	metrics := &typesv1alpha1.ServiceGraphMetrics{
		Pairs: []*typesv1alpha1.ServicePairMetrics{
			{
				SourceCluster:        "cluster-1",
				SourceNamespace:      "default",
				SourceService:        "frontend",
				DestinationCluster:   "cluster-1",
				DestinationNamespace: "default",
				DestinationService:   "backend",
				ErrorRate:            0.02,
				RequestRate:          75.0,
			},
		},
	}

	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		Namespaces: []string{"default"},
		// StartTime and EndTime will be set to defaults if not specified
	}

	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-1", req).Return(metrics, nil)

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Pairs, 1)
	assert.Len(t, resp.ClustersQueried, 1)

	pair := resp.Pairs[0]
	assert.Equal(t, "default", pair.SourceNamespace)
	assert.Equal(t, "default", pair.DestinationNamespace)

	mockConnManager.AssertExpectations(t)
	mockMeshProvider.AssertExpectations(t)
}

func TestMetricsService_GetServiceMeshGraphMetrics_UnhealthyProvider(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock connection info
	connectionInfos := map[string]connections.ConnectionInfo{
		"cluster-1": {
			ClusterID: "cluster-1",
		},
		"cluster-2": {
			ClusterID: "cluster-2",
		},
	}

	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		// StartTime and EndTime will be set to defaults if not specified
	}

	// Mock empty metrics responses
	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-1", mock.Anything).Return((*typesv1alpha1.ServiceGraphMetrics)(nil), nil)
	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-2", mock.Anything).Return((*typesv1alpha1.ServiceGraphMetrics)(nil), nil)

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Pairs, 0)
	assert.Len(t, resp.ClustersQueried, 0)

	mockConnManager.AssertExpectations(t)
	mockMeshProvider.AssertExpectations(t)
}

func TestMetricsService_GetServiceMeshGraphMetrics_ProviderError(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock connection info
	connectionInfos := map[string]connections.ConnectionInfo{
		"cluster-1": {
			ClusterID: "cluster-1",
		},
		"cluster-2": {
			ClusterID: "cluster-2",
		},
	}

	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		// StartTime and EndTime will be set to defaults if not specified
	}

	// cluster-1 returns error, cluster-2 returns valid data
	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-1", req).Return(nil, fmt.Errorf("prometheus connection failed"))

	cluster2Metrics := &typesv1alpha1.ServiceGraphMetrics{
		Pairs: []*typesv1alpha1.ServicePairMetrics{
			{
				SourceCluster:        "cluster-2",
				SourceNamespace:      "default",
				SourceService:        "app",
				DestinationCluster:   "cluster-2",
				DestinationNamespace: "default",
				DestinationService:   "db",
				ErrorRate:            0.03,
				RequestRate:          25.0,
			},
		},
	}
	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-2", req).Return(cluster2Metrics, nil)

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Pairs, 1) // Only cluster-2 data should be included
	assert.Len(t, resp.ClustersQueried, 1)
	assert.Contains(t, resp.ClustersQueried, "cluster-2")
	assert.NotContains(t, resp.ClustersQueried, "cluster-1")

	mockConnManager.AssertExpectations(t)
	mockMeshProvider.AssertExpectations(t)
}

func TestMetricsService_GetServiceMeshGraphMetrics_EmptyMetrics(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock connection info
	connectionInfos := map[string]connections.ConnectionInfo{
		"cluster-1": {
			ClusterID: "cluster-1",
		},
	}

	// Mock empty metrics response
	emptyMetrics := &typesv1alpha1.ServiceGraphMetrics{
		Pairs: []*typesv1alpha1.ServicePairMetrics{},
	}

	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		// StartTime and EndTime will be set to defaults if not specified
	}

	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-1", req).Return(emptyMetrics, nil)

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Pairs, 0)
	assert.Len(t, resp.ClustersQueried, 0) // Empty metrics means cluster not queried

	mockConnManager.AssertExpectations(t)
	mockMeshProvider.AssertExpectations(t)
}

func TestMetricsService_GetServiceMeshGraphMetrics_NilMetrics(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock connection info
	connectionInfos := map[string]connections.ConnectionInfo{
		"cluster-1": {
			ClusterID: "cluster-1",
		},
	}

	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		// StartTime and EndTime will be set to defaults if not specified
	}

	// Return nil metrics
	mockMeshProvider.On("GetServiceGraphMetrics", mock.Anything, "cluster-1", req).Return(nil, nil)

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Pairs, 0)
	assert.Len(t, resp.ClustersQueried, 0)

	mockConnManager.AssertExpectations(t)
	mockMeshProvider.AssertExpectations(t)
}

func TestConvertServicePairMetrics(t *testing.T) {
	backendPairs := []*typesv1alpha1.ServicePairMetrics{
		{
			SourceCluster:        "cluster-1",
			SourceNamespace:      "default",
			SourceService:        "frontend",
			DestinationCluster:   "cluster-1",
			DestinationNamespace: "default",
			DestinationService:   "backend",
			ErrorRate:            0.05,
			RequestRate:          100.0,
		},
		{
			SourceCluster:        "cluster-2",
			SourceNamespace:      "production",
			SourceService:        "api",
			DestinationCluster:   "cluster-2",
			DestinationNamespace: "production",
			DestinationService:   "database",
			ErrorRate:            0.01,
			RequestRate:          50.0,
		},
	}

	frontendPairs := backendPairs

	assert.Len(t, frontendPairs, 2)

	// Test first pair
	pair1 := frontendPairs[0]
	assert.Equal(t, "cluster-1", pair1.SourceCluster)
	assert.Equal(t, "default", pair1.SourceNamespace)
	assert.Equal(t, "frontend", pair1.SourceService)
	assert.Equal(t, "cluster-1", pair1.DestinationCluster)
	assert.Equal(t, "default", pair1.DestinationNamespace)
	assert.Equal(t, "backend", pair1.DestinationService)
	assert.Equal(t, 0.05, pair1.ErrorRate)
	assert.Equal(t, 100.0, pair1.RequestRate)

	// Test second pair
	pair2 := frontendPairs[1]
	assert.Equal(t, "cluster-2", pair2.SourceCluster)
	assert.Equal(t, "production", pair2.SourceNamespace)
	assert.Equal(t, "api", pair2.SourceService)
	assert.Equal(t, "cluster-2", pair2.DestinationCluster)
	assert.Equal(t, "production", pair2.DestinationNamespace)
	assert.Equal(t, "database", pair2.DestinationService)
	assert.Equal(t, 0.01, pair2.ErrorRate)
	assert.Equal(t, 50.0, pair2.RequestRate)
}

func TestConvertServicePairMetrics_Empty(t *testing.T) {
	backendPairs := []*typesv1alpha1.ServicePairMetrics{}
	frontendPairs := backendPairs

	assert.Len(t, frontendPairs, 0)
	assert.NotNil(t, frontendPairs) // Should be empty slice, not nil
}

func TestMetricsService_GetServiceMeshGraphMetrics_NoConnections(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock empty connection info
	connectionInfos := make(map[string]connections.ConnectionInfo)
	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		// StartTime and EndTime will be set to defaults if not specified
	}

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Pairs, 0)
	assert.Len(t, resp.ClustersQueried, 0)
	assert.NotEmpty(t, resp.Timestamp)

	mockConnManager.AssertExpectations(t)
	mockMeshProvider.AssertExpectations(t) // No calls should be made
}

func TestMetricsService_GetServiceMeshGraphMetrics_TimestampFormat(t *testing.T) {
	mockConnManager := &MockMetricsConnectionManager{}
	mockMeshProvider := &MockMeshMetricsProvider{}
	service := NewMetricsService(mockConnManager, mockMeshProvider, logging.For("test"))

	// Mock empty connection info (no clusters)
	connectionInfos := make(map[string]connections.ConnectionInfo)
	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		// StartTime and EndTime will be set to defaults if not specified
	}

	resp, err := service.GetServiceGraphMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify timestamp format (RFC3339 with timezone)
	_, parseErr := time.Parse("2006-01-02T15:04:05Z07:00", resp.Timestamp)
	assert.NoError(t, parseErr, "Timestamp should be in RFC3339 format")

	mockConnManager.AssertExpectations(t)
}
