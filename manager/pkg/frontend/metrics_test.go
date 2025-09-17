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

	"github.com/liamawhite/navigator/manager/pkg/connections"
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
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

func (m *MockMeshMetricsProvider) GetServiceConnections(ctx context.Context, clusterID string, req *frontendv1alpha1.GetServiceConnectionsRequest, proxyMode typesv1alpha1.ProxyMode) (*typesv1alpha1.ServiceGraphMetrics, error) {
	args := m.Called(ctx, clusterID, req, proxyMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*typesv1alpha1.ServiceGraphMetrics), args.Error(1)
}
