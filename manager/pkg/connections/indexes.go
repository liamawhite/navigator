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

// rebuildIndexes rebuilds the read-optimized indexes from current cluster states
// Must be called with m.mu.Lock() held
func (m *Manager) rebuildIndexes() {
	// Create new indexes
	newIndexes := &ReadOptimizedIndexes{
		Services:            make(map[string]*AggregatedService),
		ServicesByNamespace: make(map[string][]*AggregatedService),
		ServicesByCluster:   make(map[string][]*AggregatedService),
		Instances:           make(map[string]*AggregatedServiceInstance),
		InstancesByService:  make(map[string][]*AggregatedServiceInstance),
	}

	// Process all cluster states
	for clusterID, connection := range m.connections {
		if connection.ClusterState == nil {
			continue
		}

		var clusterServices []*AggregatedService

		// Process each service in the cluster
		for _, service := range connection.ClusterState.Services {
			serviceID := service.Namespace + ":" + service.Name

			// Get or create aggregated service
			aggService, exists := newIndexes.Services[serviceID]
			if !exists {
				aggService = &AggregatedService{
					ID:          serviceID,
					Name:        service.Name,
					Namespace:   service.Namespace,
					Instances:   make([]*AggregatedServiceInstance, 0),
					ClusterMap:  make(map[string][]*AggregatedServiceInstance),
					ClusterIPs:  make(map[string]string),
					ExternalIPs: make(map[string]string),
				}
				newIndexes.Services[serviceID] = aggService
			}

			// Add cluster IP if present
			if service.ClusterIp != "" {
				aggService.ClusterIPs[clusterID] = service.ClusterIp
			}

			// Add external IP if present
			if service.ExternalIp != "" {
				aggService.ExternalIPs[clusterID] = service.ExternalIp
			}

			var clusterInstances []*AggregatedServiceInstance

			// Process each instance in the service
			for _, instance := range service.Instances {
				instanceID := clusterID + ":" + service.Namespace + ":" + instance.PodName

				// Convert backend containers to manager containers
				containers := make([]Container, len(instance.Containers))
				for i, backendContainer := range instance.Containers {
					containers[i] = Container{
						Name:         backendContainer.Name,
						Image:        backendContainer.Image,
						Status:       backendContainer.Status,
						Ready:        backendContainer.Ready,
						RestartCount: backendContainer.RestartCount,
					}
				}

				aggInstance := &AggregatedServiceInstance{
					InstanceID:     instanceID,
					IP:             instance.Ip,
					PodName:        instance.PodName,
					Namespace:      service.Namespace,
					ClusterName:    clusterID,
					EnvoyPresent:   instance.EnvoyPresent,
					Containers:     containers,
					PodStatus:      instance.PodStatus,
					NodeName:       instance.NodeName,
					CreatedAt:      instance.CreatedAt,
					Labels:         instance.Labels,
					Annotations:    instance.Annotations,
					IsEnvoyPresent: instance.EnvoyPresent,
				}

				// Add to global instances index
				newIndexes.Instances[instanceID] = aggInstance

				// Add to service instances
				aggService.Instances = append(aggService.Instances, aggInstance)
				clusterInstances = append(clusterInstances, aggInstance)

				// Add to instances by service index
				newIndexes.InstancesByService[serviceID] = append(newIndexes.InstancesByService[serviceID], aggInstance)
			}

			// Add cluster instances to service cluster map
			if len(clusterInstances) > 0 {
				aggService.ClusterMap[clusterID] = clusterInstances
			}

			clusterServices = append(clusterServices, aggService)
		}

		// Add services to cluster index
		if len(clusterServices) > 0 {
			newIndexes.ServicesByCluster[clusterID] = clusterServices
		}
	}

	// Build namespace index
	for _, service := range newIndexes.Services {
		newIndexes.ServicesByNamespace[service.Namespace] = append(newIndexes.ServicesByNamespace[service.Namespace], service)
	}

	// Atomically update the indexes
	m.indexes.Store(newIndexes)
}

// ListAggregatedServices returns services filtered by namespace and/or cluster
func (m *Manager) ListAggregatedServices(namespace, clusterID string) []*AggregatedService {
	indexes := m.indexes.Load()
	if indexes == nil {
		return nil
	}

	var services []*AggregatedService

	if clusterID != "" {
		// Filter by cluster first
		clusterServices := indexes.ServicesByCluster[clusterID]
		if namespace != "" {
			// Also filter by namespace
			for _, service := range clusterServices {
				if service.Namespace == namespace {
					services = append(services, service)
				}
			}
		} else {
			services = clusterServices
		}
	} else if namespace != "" {
		// Filter by namespace only
		services = indexes.ServicesByNamespace[namespace]
	} else {
		// Return all services
		services = make([]*AggregatedService, 0, len(indexes.Services))
		for _, service := range indexes.Services {
			services = append(services, service)
		}
	}

	return services
}

// GetAggregatedService returns a specific service by ID
func (m *Manager) GetAggregatedService(serviceID string) (*AggregatedService, bool) {
	indexes := m.indexes.Load()
	if indexes == nil {
		return nil, false
	}

	service, exists := indexes.Services[serviceID]
	return service, exists
}

// GetAggregatedServiceInstance returns a specific service instance by ID
func (m *Manager) GetAggregatedServiceInstance(instanceID string) (*AggregatedServiceInstance, bool) {
	indexes := m.indexes.Load()
	if indexes == nil {
		return nil, false
	}

	instance, exists := indexes.Instances[instanceID]
	return instance, exists
}

// GetServiceInstances returns all instances for a specific service
func (m *Manager) GetServiceInstances(serviceID string) []*AggregatedServiceInstance {
	indexes := m.indexes.Load()
	if indexes == nil {
		return nil
	}

	return indexes.InstancesByService[serviceID]
}
