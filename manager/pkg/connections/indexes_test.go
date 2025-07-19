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

func TestManager_ReadOptimizedIndexes(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Register connection
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for registration")

	// Create test cluster state
	clusterState := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{
				Name:      "service1",
				Namespace: "default",
				Instances: []*v1alpha1.ServiceInstance{
					{
						Ip:           "10.0.0.1",
						PodName:      "pod1",
						EnvoyPresent: true,
					},
					{
						Ip:           "10.0.0.2",
						PodName:      "pod2",
						EnvoyPresent: false,
					},
				},
			},
			{
				Name:      "service2",
				Namespace: "kube-system",
				Instances: []*v1alpha1.ServiceInstance{
					{
						Ip:           "10.0.0.3",
						PodName:      "pod3",
						EnvoyPresent: false,
					},
				},
			},
		},
	}

	// Update cluster state
	err = manager.UpdateClusterState("cluster1", clusterState)
	assert.NoError(t, err, "Expected no error for cluster state update")

	// Test ListAggregatedServices - all services
	services := manager.ListAggregatedServices("", "")
	assert.Len(t, services, 2, "Expected 2 services")

	// Test ListAggregatedServices - filter by namespace
	services = manager.ListAggregatedServices("default", "")
	assert.Len(t, services, 1, "Expected 1 service in default namespace")
	assert.Equal(t, "service1", services[0].Name, "Expected service1")

	// Test ListAggregatedServices - filter by cluster
	services = manager.ListAggregatedServices("", "cluster1")
	assert.Len(t, services, 2, "Expected 2 services in cluster1")

	// Test ListAggregatedServices - filter by namespace and cluster
	services = manager.ListAggregatedServices("kube-system", "cluster1")
	assert.Len(t, services, 1, "Expected 1 service in kube-system namespace and cluster1")
	assert.Equal(t, "service2", services[0].Name, "Expected service2")

	// Test GetAggregatedService
	service, exists := manager.GetAggregatedService("default:service1")
	assert.True(t, exists, "Expected service to exist")
	assert.Equal(t, "service1", service.Name, "Expected service name to match")
	assert.Equal(t, "default", service.Namespace, "Expected namespace to match")
	assert.Len(t, service.Instances, 2, "Expected 2 instances")

	// Test GetAggregatedServiceInstance
	instance, exists := manager.GetAggregatedServiceInstance("cluster1:default:pod1")
	assert.True(t, exists, "Expected instance to exist")
	assert.Equal(t, "pod1", instance.PodName, "Expected pod name to match")
	assert.Equal(t, "cluster1", instance.ClusterName, "Expected cluster name to match")
	assert.Equal(t, "10.0.0.1", instance.IP, "Expected IP to match")
	assert.True(t, instance.EnvoyPresent, "Expected EnvoyPresent to be true")

	// Test GetServiceInstances
	instances := manager.GetServiceInstances("default:service1")
	assert.Len(t, instances, 2, "Expected 2 instances for service1")

	// Test non-existent service
	_, exists = manager.GetAggregatedService("nonexistent:service")
	assert.False(t, exists, "Expected service to not exist")

	// Test non-existent instance
	_, exists = manager.GetAggregatedServiceInstance("nonexistent:instance")
	assert.False(t, exists, "Expected instance to not exist")
}

func TestManager_MultiClusterAggregation(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Register two clusters
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for cluster1 registration")

	err = manager.RegisterConnection("cluster2", nil)
	assert.NoError(t, err, "Expected no error for cluster2 registration")

	// Create cluster state for cluster1
	clusterState1 := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{
				Name:      "web-service",
				Namespace: "default",
				Instances: []*v1alpha1.ServiceInstance{
					{
						Ip:           "10.1.0.1",
						PodName:      "web-pod1",
						EnvoyPresent: true,
					},
				},
			},
		},
	}

	// Create cluster state for cluster2
	clusterState2 := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{
				Name:      "web-service",
				Namespace: "default",
				Instances: []*v1alpha1.ServiceInstance{
					{
						Ip:           "10.2.0.1",
						PodName:      "web-pod2",
						EnvoyPresent: false,
					},
				},
			},
		},
	}

	// Update cluster states
	err = manager.UpdateClusterState("cluster1", clusterState1)
	assert.NoError(t, err, "Expected no error for cluster1 state update")

	err = manager.UpdateClusterState("cluster2", clusterState2)
	assert.NoError(t, err, "Expected no error for cluster2 state update")

	// Test aggregated service - should have instances from both clusters
	service, exists := manager.GetAggregatedService("default:web-service")
	assert.True(t, exists, "Expected aggregated service to exist")
	assert.Len(t, service.Instances, 2, "Expected 2 instances across clusters")

	// Test cluster mapping
	cluster1Instances := service.ClusterMap["cluster1"]
	cluster2Instances := service.ClusterMap["cluster2"]

	assert.Len(t, cluster1Instances, 1, "Expected 1 instance in cluster1")
	assert.Len(t, cluster2Instances, 1, "Expected 1 instance in cluster2")

	// Test filtering by cluster
	services := manager.ListAggregatedServices("", "cluster1")
	assert.Len(t, services, 1, "Expected 1 service in cluster1")

	// Test that unregistering a cluster removes its data
	manager.UnregisterConnection("cluster1")

	service, exists = manager.GetAggregatedService("default:web-service")
	assert.True(t, exists, "Expected service to still exist after cluster1 removal")
	assert.Len(t, service.Instances, 1, "Expected 1 instance after cluster1 removal")
	assert.Equal(t, "cluster2", service.Instances[0].ClusterName, "Expected remaining instance to be from cluster2")
}

func TestManager_EmptyIndexes(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Test with no connections
	services := manager.ListAggregatedServices("", "")
	assert.Empty(t, services, "Expected empty services list with no connections")

	service, exists := manager.GetAggregatedService("default:test")
	assert.False(t, exists, "Expected service to not exist with no connections")
	assert.Nil(t, service, "Expected nil service with no connections")

	instance, exists := manager.GetAggregatedServiceInstance("cluster1:default:pod1")
	assert.False(t, exists, "Expected instance to not exist with no connections")
	assert.Nil(t, instance, "Expected nil instance with no connections")

	instances := manager.GetServiceInstances("default:test")
	assert.Empty(t, instances, "Expected empty instances list with no connections")
}

func TestManager_FilteringEdgeCases(t *testing.T) {
	logger := logging.For("test")
	manager := NewManager(logger)

	// Register connection and add state
	err := manager.RegisterConnection("cluster1", nil)
	assert.NoError(t, err, "Expected no error for registration")

	clusterState := &v1alpha1.ClusterState{
		Services: []*v1alpha1.Service{
			{
				Name:      "service1",
				Namespace: "default",
				Instances: []*v1alpha1.ServiceInstance{
					{Ip: "10.0.0.1", PodName: "pod1", EnvoyPresent: true},
				},
			},
		},
	}

	err = manager.UpdateClusterState("cluster1", clusterState)
	assert.NoError(t, err, "Expected no error for cluster state update")

	// Test filtering by non-existent namespace
	services := manager.ListAggregatedServices("nonexistent", "")
	assert.Empty(t, services, "Expected empty result for non-existent namespace")

	// Test filtering by non-existent cluster
	services = manager.ListAggregatedServices("", "nonexistent")
	assert.Empty(t, services, "Expected empty result for non-existent cluster")

	// Test filtering by non-existent namespace and cluster
	services = manager.ListAggregatedServices("nonexistent", "nonexistent")
	assert.Empty(t, services, "Expected empty result for non-existent namespace and cluster")
}
