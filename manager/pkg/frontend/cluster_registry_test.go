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
	"time"

	"github.com/liamawhite/navigator/manager/pkg/connections"
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClusterRegistryConnectionManager for testing
type MockClusterRegistryConnectionManager struct {
	mock.Mock
}

func (m *MockClusterRegistryConnectionManager) RegisterConnection(clusterID string, stream backendv1alpha1.ManagerService_ConnectServer) error {
	args := m.Called(clusterID, stream)
	return args.Error(0)
}

func (m *MockClusterRegistryConnectionManager) UnregisterConnection(clusterID string) {
	m.Called(clusterID)
}

func (m *MockClusterRegistryConnectionManager) UpdateClusterState(clusterID string, clusterState *backendv1alpha1.ClusterState) error {
	args := m.Called(clusterID, clusterState)
	return args.Error(0)
}

func (m *MockClusterRegistryConnectionManager) UpdateCapabilities(clusterID string, capabilities *backendv1alpha1.EdgeCapabilities) error {
	args := m.Called(clusterID, capabilities)
	return args.Error(0)
}

func (m *MockClusterRegistryConnectionManager) GetClusterState(clusterID string) (*backendv1alpha1.ClusterState, error) {
	args := m.Called(clusterID)
	return args.Get(0).(*backendv1alpha1.ClusterState), args.Error(1)
}

func (m *MockClusterRegistryConnectionManager) GetAllClusterStates() map[string]*backendv1alpha1.ClusterState {
	args := m.Called()
	return args.Get(0).(map[string]*backendv1alpha1.ClusterState)
}

func (m *MockClusterRegistryConnectionManager) IsClusterConnected(clusterID string) bool {
	args := m.Called(clusterID)
	return args.Bool(0)
}

func (m *MockClusterRegistryConnectionManager) GetActiveClusterCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockClusterRegistryConnectionManager) SendMessageToCluster(clusterID string, message *backendv1alpha1.ConnectResponse) error {
	args := m.Called(clusterID, message)
	return args.Error(0)
}

func (m *MockClusterRegistryConnectionManager) ListAggregatedServices(namespace, clusterID string) []*connections.AggregatedService {
	args := m.Called(namespace, clusterID)
	return args.Get(0).([]*connections.AggregatedService)
}

func (m *MockClusterRegistryConnectionManager) GetAggregatedService(serviceID string) (*connections.AggregatedService, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*connections.AggregatedService), args.Bool(1)
}

func (m *MockClusterRegistryConnectionManager) GetAggregatedServiceInstance(instanceID string) (*connections.AggregatedServiceInstance, bool) {
	args := m.Called(instanceID)
	return args.Get(0).(*connections.AggregatedServiceInstance), args.Bool(1)
}

func (m *MockClusterRegistryConnectionManager) GetConnectionInfo() map[string]connections.ConnectionInfo {
	args := m.Called()
	return args.Get(0).(map[string]connections.ConnectionInfo)
}

func TestClusterRegistryService_ListClusters(t *testing.T) {
	mockConnManager := &MockClusterRegistryConnectionManager{}
	service := NewClusterRegistryService(mockConnManager, logging.For("test"))

	// Mock connection info data
	now := time.Now()
	connectionInfos := map[string]connections.ConnectionInfo{
		"cluster-1": {
			ClusterID:     "cluster-1",
			ConnectedAt:   now.Add(-5 * time.Minute),
			LastUpdate:    now.Add(-10 * time.Second), // Healthy (< 30s)
			ServiceCount:  15,
			StateReceived: true,
		},
		"cluster-2": {
			ClusterID:     "cluster-2",
			ConnectedAt:   now.Add(-10 * time.Minute),
			LastUpdate:    now.Add(-2 * time.Minute), // Stale (< 5min but > 30s)
			ServiceCount:  8,
			StateReceived: true,
		},
		"cluster-3": {
			ClusterID:     "cluster-3",
			ConnectedAt:   now.Add(-30 * time.Minute),
			LastUpdate:    now.Add(-10 * time.Minute), // Disconnected (> 5min)
			ServiceCount:  0,
			StateReceived: true,
		},
		"cluster-4": {
			ClusterID:     "cluster-4",
			ConnectedAt:   now.Add(-1 * time.Minute),
			LastUpdate:    time.Time{}, // Zero time
			ServiceCount:  0,
			StateReceived: false, // Initializing
		},
	}

	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.ListClustersRequest{}
	resp, err := service.ListClusters(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Clusters, 4)

	// Create a map for easier testing
	clustersByID := make(map[string]*frontendv1alpha1.ClusterSyncInfo)
	for _, cluster := range resp.Clusters {
		clustersByID[cluster.ClusterId] = cluster
	}

	// Test cluster-1 (healthy)
	cluster1 := clustersByID["cluster-1"]
	assert.NotNil(t, cluster1)
	assert.Equal(t, "cluster-1", cluster1.ClusterId)
	assert.Equal(t, int32(15), cluster1.ServiceCount)
	assert.Equal(t, frontendv1alpha1.SyncStatus_SYNC_STATUS_HEALTHY, cluster1.SyncStatus)

	// Test cluster-2 (stale)
	cluster2 := clustersByID["cluster-2"]
	assert.NotNil(t, cluster2)
	assert.Equal(t, "cluster-2", cluster2.ClusterId)
	assert.Equal(t, int32(8), cluster2.ServiceCount)
	assert.Equal(t, frontendv1alpha1.SyncStatus_SYNC_STATUS_STALE, cluster2.SyncStatus)

	// Test cluster-3 (disconnected)
	cluster3 := clustersByID["cluster-3"]
	assert.NotNil(t, cluster3)
	assert.Equal(t, "cluster-3", cluster3.ClusterId)
	assert.Equal(t, int32(0), cluster3.ServiceCount)
	assert.Equal(t, frontendv1alpha1.SyncStatus_SYNC_STATUS_DISCONNECTED, cluster3.SyncStatus)

	// Test cluster-4 (initializing)
	cluster4 := clustersByID["cluster-4"]
	assert.NotNil(t, cluster4)
	assert.Equal(t, "cluster-4", cluster4.ClusterId)
	assert.Equal(t, int32(0), cluster4.ServiceCount)
	assert.Equal(t, frontendv1alpha1.SyncStatus_SYNC_STATUS_INITIALIZING, cluster4.SyncStatus)

	mockConnManager.AssertExpectations(t)
}

func TestClusterRegistryService_ListClusters_Empty(t *testing.T) {
	mockConnManager := &MockClusterRegistryConnectionManager{}
	service := NewClusterRegistryService(mockConnManager, logging.For("test"))

	// Mock empty connection info
	connectionInfos := make(map[string]connections.ConnectionInfo)
	mockConnManager.On("GetConnectionInfo").Return(connectionInfos)

	req := &frontendv1alpha1.ListClustersRequest{}
	resp, err := service.ListClusters(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Clusters, 0)

	mockConnManager.AssertExpectations(t)
}

func TestConvertConnectionInfoToClusterSyncInfo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		connInfo connections.ConnectionInfo
		expected frontendv1alpha1.SyncStatus
	}{
		{
			name: "healthy status",
			connInfo: connections.ConnectionInfo{
				ClusterID:     "test-cluster",
				ConnectedAt:   now.Add(-1 * time.Hour),
				LastUpdate:    now.Add(-15 * time.Second), // < 30s
				ServiceCount:  10,
				StateReceived: true,
			},
			expected: frontendv1alpha1.SyncStatus_SYNC_STATUS_HEALTHY,
		},
		{
			name: "stale status",
			connInfo: connections.ConnectionInfo{
				ClusterID:     "test-cluster",
				ConnectedAt:   now.Add(-1 * time.Hour),
				LastUpdate:    now.Add(-2 * time.Minute), // > 30s but < 5min
				ServiceCount:  5,
				StateReceived: true,
			},
			expected: frontendv1alpha1.SyncStatus_SYNC_STATUS_STALE,
		},
		{
			name: "disconnected status",
			connInfo: connections.ConnectionInfo{
				ClusterID:     "test-cluster",
				ConnectedAt:   now.Add(-1 * time.Hour),
				LastUpdate:    now.Add(-10 * time.Minute), // > 5min
				ServiceCount:  0,
				StateReceived: true,
			},
			expected: frontendv1alpha1.SyncStatus_SYNC_STATUS_DISCONNECTED,
		},
		{
			name: "initializing status",
			connInfo: connections.ConnectionInfo{
				ClusterID:     "test-cluster",
				ConnectedAt:   now.Add(-1 * time.Minute),
				LastUpdate:    now.Add(-10 * time.Second),
				ServiceCount:  0,
				StateReceived: false, // No state received yet
			},
			expected: frontendv1alpha1.SyncStatus_SYNC_STATUS_INITIALIZING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertConnectionInfoToClusterSyncInfo(tt.connInfo)

			assert.Equal(t, tt.connInfo.ClusterID, result.ClusterId)
			assert.Equal(t, tt.expected, result.SyncStatus)
			assert.Equal(t, int32(tt.connInfo.ServiceCount), result.ServiceCount) //nolint:gosec // G115: test overflow conversion is intentional
		})
	}
}

func TestConvertConnectionInfoToClusterSyncInfo_ServiceCountOverflow(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		serviceCount  int
		expectedInt32 int32
	}{
		{
			name:          "normal count",
			serviceCount:  100,
			expectedInt32: 100,
		},
		{
			name:          "max int32",
			serviceCount:  2147483647,
			expectedInt32: 2147483647,
		},
		{
			name:          "overflow positive",
			serviceCount:  2147483648, // max int32 + 1
			expectedInt32: 2147483647, // should be capped to max
		},
		{
			name:          "large overflow",
			serviceCount:  9999999999,
			expectedInt32: 2147483647, // should be capped to max
		},
		{
			name:          "negative underflow",
			serviceCount:  -2147483649, // min int32 - 1
			expectedInt32: -2147483648, // should be capped to min
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connInfo := connections.ConnectionInfo{
				ClusterID:     "test-cluster",
				ConnectedAt:   now,
				LastUpdate:    now,
				ServiceCount:  tt.serviceCount,
				StateReceived: true,
			}

			result := convertConnectionInfoToClusterSyncInfo(connInfo)
			assert.Equal(t, tt.expectedInt32, result.ServiceCount)
		})
	}
}

func TestComputeSyncStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		stateReceived  bool
		lastUpdateAgo  time.Duration
		expectedStatus frontendv1alpha1.SyncStatus
	}{
		{
			name:           "initializing - no state received",
			stateReceived:  false,
			lastUpdateAgo:  time.Second, // doesn't matter
			expectedStatus: frontendv1alpha1.SyncStatus_SYNC_STATUS_INITIALIZING,
		},
		{
			name:           "healthy - 10 seconds ago",
			stateReceived:  true,
			lastUpdateAgo:  10 * time.Second,
			expectedStatus: frontendv1alpha1.SyncStatus_SYNC_STATUS_HEALTHY,
		},
		{
			name:           "healthy - exactly 30 seconds ago",
			stateReceived:  true,
			lastUpdateAgo:  30 * time.Second,
			expectedStatus: frontendv1alpha1.SyncStatus_SYNC_STATUS_STALE, // 30s is not < 30s
		},
		{
			name:           "stale - 1 minute ago",
			stateReceived:  true,
			lastUpdateAgo:  1 * time.Minute,
			expectedStatus: frontendv1alpha1.SyncStatus_SYNC_STATUS_STALE,
		},
		{
			name:           "stale - exactly 5 minutes ago",
			stateReceived:  true,
			lastUpdateAgo:  5 * time.Minute,
			expectedStatus: frontendv1alpha1.SyncStatus_SYNC_STATUS_DISCONNECTED, // 5min is not < 5min
		},
		{
			name:           "disconnected - 10 minutes ago",
			stateReceived:  true,
			lastUpdateAgo:  10 * time.Minute,
			expectedStatus: frontendv1alpha1.SyncStatus_SYNC_STATUS_DISCONNECTED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connInfo := connections.ConnectionInfo{
				StateReceived: tt.stateReceived,
				LastUpdate:    now.Add(-tt.lastUpdateAgo),
			}

			result := computeSyncStatus(connInfo)
			assert.Equal(t, tt.expectedStatus, result)
		})
	}
}
