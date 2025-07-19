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
	"testing"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestManager_RegisterConnection(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Test successful registration
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for first registration")

	// Test duplicate registration
	err = manager.RegisterConnection("cluster1", nil)
	assert.Error(t, err, "Expected error for duplicate registration")

	// Test different cluster registration
	err = manager.RegisterConnection("cluster2", nil)
	assert.NoError(t, err, "Expected no error for different cluster registration")
}

func TestManager_UnregisterConnection(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Register connection
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for registration")

	// Verify connection exists
	assert.True(t, manager.IsClusterConnected("cluster1"), "Expected cluster to be connected")

	// Unregister connection
	manager.UnregisterConnection("cluster1")

	// Verify connection is removed
	assert.False(t, manager.IsClusterConnected("cluster1"), "Expected cluster to be disconnected")

	// Unregister non-existent connection (should not panic)
	assert.NotPanics(t, func() {
		manager.UnregisterConnection("non-existent")
	}, "Unregistering non-existent connection should not panic")
}

func TestManager_UpdateClusterState(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Test update without connection
	clusterState := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{Name: "test-service", Namespace: "default"},
		},
	}

	err := manager.UpdateClusterState("cluster1", clusterState)
	assert.Error(t, err, "Expected error for update without connection")

	// Register connection
	err = manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for registration")

	// Test successful update
	err = manager.UpdateClusterState("cluster1", clusterState)
	assert.NoError(t, err, "Expected no error for cluster state update")

	// Verify state was updated
	retrievedState, err := manager.GetClusterState("cluster1")
	assert.NoError(t, err, "Expected no error retrieving cluster state")
	assert.Len(t, retrievedState.Services, 1, "Expected 1 service in cluster state")
	assert.Equal(t, "test-service", retrievedState.Services[0].Name, "Service name should match")
}

func TestManager_GetClusterState(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Test get state without connection
	_, err := manager.GetClusterState("cluster1")
	assert.Error(t, err, "Expected error for get state without connection")

	// Register connection
	err = manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for registration")

	// Test get state without cluster state
	_, err = manager.GetClusterState("cluster1")
	assert.Error(t, err, "Expected error for get state without cluster state")

	// Update cluster state
	clusterState := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{Name: "test-service", Namespace: "default"},
		},
	}

	err = manager.UpdateClusterState("cluster1", clusterState)
	assert.NoError(t, err, "Expected no error for cluster state update")

	// Test successful get state
	retrievedState, err := manager.GetClusterState("cluster1")
	assert.NoError(t, err, "Expected no error retrieving cluster state")
	assert.Len(t, retrievedState.Services, 1, "Expected 1 service in retrieved state")
}

func TestManager_GetAllClusterStates(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Test empty result
	states := manager.GetAllClusterStates()
	assert.Empty(t, states, "Expected empty result for no connections")

	// Register multiple connections
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for cluster1 registration")

	err = manager.RegisterConnection("cluster2", nil)
	assert.NoError(t, err, "Expected no error for cluster2 registration")

	// Update cluster states
	clusterState1 := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{Name: "service1", Namespace: "default"},
		},
	}

	clusterState2 := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{Name: "service2", Namespace: "default"},
		},
	}

	err = manager.UpdateClusterState("cluster1", clusterState1)
	assert.NoError(t, err, "Expected no error for cluster1 state update")

	err = manager.UpdateClusterState("cluster2", clusterState2)
	assert.NoError(t, err, "Expected no error for cluster2 state update")

	// Test get all states
	states = manager.GetAllClusterStates()
	assert.Len(t, states, 2, "Expected 2 cluster states")
	assert.Contains(t, states, "cluster1", "Expected cluster1 state to be present")
	assert.Contains(t, states, "cluster2", "Expected cluster2 state to be present")
}

func TestManager_GetConnectionInfo(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Register connection
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for registration")

	// Update cluster state
	clusterState := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{Name: "service1", Namespace: "default"},
			{Name: "service2", Namespace: "default"},
		},
	}

	err = manager.UpdateClusterState("cluster1", clusterState)
	assert.NoError(t, err, "Expected no error for cluster state update")

	// Get connection info
	info := manager.GetConnectionInfo()
	assert.Len(t, info, 1, "Expected 1 connection info")

	clusterInfo, exists := info["cluster1"]
	assert.True(t, exists, "Expected cluster1 info to exist")
	assert.Equal(t, "cluster1", clusterInfo.ClusterID, "Expected cluster ID to match")
	assert.Equal(t, 2, clusterInfo.ServiceCount, "Expected service count to be 2")
	assert.False(t, clusterInfo.ConnectedAt.IsZero(), "Expected ConnectedAt to be set")
	assert.False(t, clusterInfo.LastUpdate.IsZero(), "Expected LastUpdate to be set")
}

func TestManager_GetActiveClusterCount(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Test initial count
	count := manager.GetActiveClusterCount()
	assert.Equal(t, 0, count, "Expected 0 active clusters initially")

	// Register connections
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for cluster1 registration")

	err = manager.RegisterConnection("cluster2", nil)
	assert.NoError(t, err, "Expected no error for cluster2 registration")

	// Test count after registrations
	count = manager.GetActiveClusterCount()
	assert.Equal(t, 2, count, "Expected 2 active clusters after registrations")

	// Unregister one connection
	manager.UnregisterConnection("cluster1")

	// Test count after unregistration
	count = manager.GetActiveClusterCount()
	assert.Equal(t, 1, count, "Expected 1 active cluster after unregistration")
}
