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

package connections

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// Manager manages active connections and cluster state
type Manager struct {
	logger *slog.Logger

	// Connection management (protected by mu)
	mu          sync.RWMutex
	connections map[string]*Connection // cluster_id -> connection

	// Read-optimized indexes (atomic pointer for lock-free reads)
	// This allows multiple goroutines to read service data simultaneously
	// without blocking each other or blocking writers. Writers atomically
	// replace the entire index structure, ensuring readers always see
	// either the complete old or complete new version.
	indexes atomic.Pointer[ReadOptimizedIndexes]
}

// NewManager creates a new connection manager
func NewManager(logger *slog.Logger) *Manager {
	m := &Manager{
		logger:      logger,
		connections: make(map[string]*Connection),
	}

	// Initialize empty indexes
	m.indexes.Store(&ReadOptimizedIndexes{
		Services:            make(map[string]*AggregatedService),
		ServicesByNamespace: make(map[string][]*AggregatedService),
		ServicesByCluster:   make(map[string][]*AggregatedService),
		Instances:           make(map[string]*AggregatedServiceInstance),
		InstancesByService:  make(map[string][]*AggregatedServiceInstance),
	})

	return m
}

// RegisterConnection attempts to register a new connection for a cluster
func (m *Manager) RegisterConnection(clusterID string, stream v1alpha1.ManagerService_ConnectServer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if cluster already has an active connection
	if existing, exists := m.connections[clusterID]; exists {
		m.logger.Warn("connection rejected - cluster already has active connection",
			"cluster_id", clusterID,
			"existing_connected_at", existing.ConnectedAt)
		return fmt.Errorf("cluster %s already has an active connection", clusterID)
	}

	// Register new connection
	connection := &Connection{
		ClusterID:   clusterID,
		ConnectedAt: time.Now(),
		LastUpdate:  time.Now(),
		Stream:      stream,
	}

	m.connections[clusterID] = connection

	m.logger.Info("connection registered",
		"cluster_id", clusterID,
		"connected_at", connection.ConnectedAt)

	return nil
}

// UnregisterConnection removes a connection for a cluster
func (m *Manager) UnregisterConnection(clusterID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if connection, exists := m.connections[clusterID]; exists {
		delete(m.connections, clusterID)

		// Rebuild read-optimized indexes after removing cluster
		m.rebuildIndexes()

		duration := time.Since(connection.ConnectedAt)
		m.logger.Info("connection unregistered",
			"cluster_id", clusterID,
			"connected_duration", duration)
	}
}

// UpdateClusterState updates the cluster state for a connection
func (m *Manager) UpdateClusterState(clusterID string, clusterState *v1alpha1.ClusterState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	connection, exists := m.connections[clusterID]
	if !exists {
		return fmt.Errorf("no active connection for cluster %s", clusterID)
	}

	connection.ClusterState = clusterState
	connection.LastUpdate = time.Now()

	// Rebuild read-optimized indexes
	m.rebuildIndexes()

	m.logger.Debug("cluster state updated",
		"cluster_id", clusterID,
		"services", len(clusterState.Services),
		"last_update", connection.LastUpdate)

	return nil
}

// GetClusterState returns the current cluster state for a cluster
func (m *Manager) GetClusterState(clusterID string) (*v1alpha1.ClusterState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connection, exists := m.connections[clusterID]
	if !exists {
		return nil, fmt.Errorf("no active connection for cluster %s", clusterID)
	}

	if connection.ClusterState == nil {
		return nil, fmt.Errorf("no cluster state available for cluster %s", clusterID)
	}

	return connection.ClusterState, nil
}

// GetAllClusterStates returns cluster states for all connected clusters
func (m *Manager) GetAllClusterStates() map[string]*v1alpha1.ClusterState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*v1alpha1.ClusterState)

	for clusterID, connection := range m.connections {
		if connection.ClusterState != nil {
			result[clusterID] = connection.ClusterState
		}
	}

	return result
}

// GetConnectionInfo returns information about active connections
func (m *Manager) GetConnectionInfo() map[string]ConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]ConnectionInfo)

	for clusterID, connection := range m.connections {
		serviceCount := 0
		if connection.ClusterState != nil {
			serviceCount = len(connection.ClusterState.Services)
		}

		result[clusterID] = ConnectionInfo{
			ClusterID:     clusterID,
			ConnectedAt:   connection.ConnectedAt,
			LastUpdate:    connection.LastUpdate,
			ServiceCount:  serviceCount,
			StateReceived: connection.ClusterState != nil,
		}
	}

	return result
}

// IsClusterConnected checks if a cluster has an active connection
func (m *Manager) IsClusterConnected(clusterID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.connections[clusterID]
	return exists
}

// GetActiveClusterCount returns the number of active cluster connections
func (m *Manager) GetActiveClusterCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.connections)
}

// SendMessageToCluster sends a message to a specific cluster
func (m *Manager) SendMessageToCluster(clusterID string, message *v1alpha1.ConnectResponse) error {
	m.mu.RLock()
	connection, exists := m.connections[clusterID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cluster %s is not connected", clusterID)
	}

	if err := connection.Stream.Send(message); err != nil {
		m.logger.Error("failed to send message to cluster", "cluster_id", clusterID, "error", err)
		return fmt.Errorf("failed to send message to cluster %s: %w", clusterID, err)
	}

	m.logger.Debug("message sent to cluster", "cluster_id", clusterID)
	return nil
}
