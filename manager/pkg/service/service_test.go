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
	"testing"
	"time"

	"github.com/liamawhite/navigator/manager/pkg/connections"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock config for testing
type mockConfig struct {
	port           int
	maxMessageSize int
}

func (m *mockConfig) GetPort() int {
	return m.port
}

func (m *mockConfig) GetMaxMessageSize() int {
	return m.maxMessageSize
}

func (m *mockConfig) Validate() error {
	return nil
}

// Mock connection manager for testing
type mockConnectionManager struct {
	connections map[string]bool
	states      map[string]*v1alpha1.ClusterState
	shouldFail  bool
}

func newMockConnectionManager() *mockConnectionManager {
	return &mockConnectionManager{
		connections: make(map[string]bool),
		states:      make(map[string]*v1alpha1.ClusterState),
	}
}

func (m *mockConnectionManager) RegisterConnection(clusterID string, stream v1alpha1.ManagerService_ConnectServer) error {
	if m.shouldFail {
		return status.Errorf(codes.AlreadyExists, "connection already exists")
	}

	if m.connections[clusterID] {
		return status.Errorf(codes.AlreadyExists, "connection already exists")
	}

	m.connections[clusterID] = true
	return nil
}

func (m *mockConnectionManager) UnregisterConnection(clusterID string) {
	delete(m.connections, clusterID)
	delete(m.states, clusterID)
}

func (m *mockConnectionManager) UpdateClusterState(clusterID string, clusterState *v1alpha1.ClusterState) error {
	if !m.connections[clusterID] {
		return status.Errorf(codes.NotFound, "connection not found")
	}

	m.states[clusterID] = clusterState
	return nil
}

func (m *mockConnectionManager) GetClusterState(clusterID string) (*v1alpha1.ClusterState, error) {
	state, exists := m.states[clusterID]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "cluster state not found")
	}
	return state, nil
}

func (m *mockConnectionManager) GetAllClusterStates() map[string]*v1alpha1.ClusterState {
	return m.states
}

func (m *mockConnectionManager) IsClusterConnected(clusterID string) bool {
	return m.connections[clusterID]
}

func (m *mockConnectionManager) GetActiveClusterCount() int {
	return len(m.connections)
}

func (m *mockConnectionManager) SendMessageToCluster(clusterID string, message *v1alpha1.ConnectResponse) error {
	if !m.connections[clusterID] {
		return status.Errorf(codes.NotFound, "connection not found")
	}
	return nil
}

// Read-optimized methods for ReadOptimizedConnectionManager interface
func (m *mockConnectionManager) ListAggregatedServices(namespace, clusterID string) []*connections.AggregatedService {
	// Simple mock implementation - return empty slice
	return []*connections.AggregatedService{}
}

func (m *mockConnectionManager) GetAggregatedService(serviceID string) (*connections.AggregatedService, bool) {
	// Simple mock implementation - return nil, false
	return nil, false
}

func (m *mockConnectionManager) GetAggregatedServiceInstance(instanceID string) (*connections.AggregatedServiceInstance, bool) {
	// Simple mock implementation - return nil, false
	return nil, false
}

func TestManagerService_processClusterIdentification(t *testing.T) {
	logger := logging.For("test")
	config := &mockConfig{port: 8080, maxMessageSize: 10485760}
	connectionManager := newMockConnectionManager()

	service, err := NewManagerService(config, connectionManager, logger)
	if err != nil {
		t.Fatalf("Failed to create manager service: %v", err)
	}

	tests := []struct {
		name        string
		req         *v1alpha1.ConnectRequest
		expectedID  string
		expectError bool
	}{
		{
			name: "valid cluster identification",
			req: &v1alpha1.ConnectRequest{
				Message: &v1alpha1.ConnectRequest_ClusterIdentification{
					ClusterIdentification: &v1alpha1.ClusterIdentification{
						ClusterId: "test-cluster",
					},
				},
			},
			expectedID:  "test-cluster",
			expectError: false,
		},
		{
			name: "empty cluster ID",
			req: &v1alpha1.ConnectRequest{
				Message: &v1alpha1.ConnectRequest_ClusterIdentification{
					ClusterIdentification: &v1alpha1.ClusterIdentification{
						ClusterId: "",
					},
				},
			},
			expectedID:  "",
			expectError: true,
		},
		{
			name: "nil cluster identification",
			req: &v1alpha1.ConnectRequest{
				Message: &v1alpha1.ConnectRequest_ClusterIdentification{
					ClusterIdentification: nil,
				},
			},
			expectedID:  "",
			expectError: true,
		},
		{
			name: "wrong message type",
			req: &v1alpha1.ConnectRequest{
				Message: &v1alpha1.ConnectRequest_ClusterState{
					ClusterState: &v1alpha1.ClusterState{},
				},
			},
			expectedID:  "",
			expectError: true,
		},
		{
			name: "empty message",
			req: &v1alpha1.ConnectRequest{
				Message: nil,
			},
			expectedID:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterID, err := service.processClusterIdentification(tt.req)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if clusterID != tt.expectedID {
				t.Errorf("Expected cluster ID '%s' but got '%s'", tt.expectedID, clusterID)
			}
		})
	}
}

func TestManagerService_processClusterStateUpdate(t *testing.T) {
	logger := logging.For("test")
	config := &mockConfig{port: 8080, maxMessageSize: 10485760}
	connectionManager := newMockConnectionManager()

	service, err := NewManagerService(config, connectionManager, logger)
	if err != nil {
		t.Fatalf("Failed to create manager service: %v", err)
	}

	// Register a connection first
	_ = connectionManager.RegisterConnection("test-cluster", nil)

	tests := []struct {
		name        string
		clusterID   string
		req         *v1alpha1.ConnectRequest
		expectError bool
	}{
		{
			name:      "valid cluster state update",
			clusterID: "test-cluster",
			req: &v1alpha1.ConnectRequest{
				Message: &v1alpha1.ConnectRequest_ClusterState{
					ClusterState: &v1alpha1.ClusterState{
						Services: []*v1alpha1.Service{
							{Name: "test-service", Namespace: "default"},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:      "nil cluster state",
			clusterID: "test-cluster",
			req: &v1alpha1.ConnectRequest{
				Message: &v1alpha1.ConnectRequest_ClusterState{
					ClusterState: nil,
				},
			},
			expectError: true,
		},
		{
			name:      "wrong message type",
			clusterID: "test-cluster",
			req: &v1alpha1.ConnectRequest{
				Message: &v1alpha1.ConnectRequest_ClusterIdentification{
					ClusterIdentification: &v1alpha1.ClusterIdentification{
						ClusterId: "test-cluster",
					},
				},
			},
			expectError: true,
		},
		{
			name:      "empty message",
			clusterID: "test-cluster",
			req: &v1alpha1.ConnectRequest{
				Message: nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.processClusterStateUpdate(tt.clusterID, tt.req)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			// Verify state was updated
			if tt.req.Message != nil {
				if clusterStateMsg, ok := tt.req.Message.(*v1alpha1.ConnectRequest_ClusterState); ok {
					if clusterStateMsg.ClusterState != nil {
						state, err := connectionManager.GetClusterState(tt.clusterID)
						if err != nil {
							t.Errorf("Expected cluster state to be saved but got error: %v", err)
						}
						if len(state.Services) != len(clusterStateMsg.ClusterState.Services) {
							t.Error("Cluster state was not updated correctly")
						}
					}
				}
			}
		})
	}
}

func TestManagerService_StartStop(t *testing.T) {
	logger := logging.For("test")
	config := &mockConfig{port: 0, maxMessageSize: 10485760} // Use port 0 to get a random available port
	connectionManager := newMockConnectionManager()

	service, err := NewManagerService(config, connectionManager, logger)
	if err != nil {
		t.Fatalf("Failed to create manager service: %v", err)
	}

	// Test start
	err = service.Start()
	if err != nil {
		t.Errorf("Expected no error starting service, got: %v", err)
	}

	// Test double start
	err = service.Start()
	if err == nil {
		t.Error("Expected error starting service twice")
	}

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test stop
	err = service.Stop()
	if err != nil {
		t.Errorf("Expected no error stopping service, got: %v", err)
	}

	// Test double stop
	err = service.Stop()
	if err != nil {
		t.Errorf("Expected no error stopping service twice, got: %v", err)
	}
}
