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
	"fmt"

	"github.com/liamawhite/navigator/manager/pkg/connections"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// convertAggregatedService converts an AggregatedService to the frontend API format
func convertAggregatedService(aggService *connections.AggregatedService) *frontendv1alpha1.Service {
	instances := make([]*frontendv1alpha1.ServiceInstance, 0, len(aggService.Instances))

	// Determine the overall proxy mode for this service
	serviceProxyMode := typesv1alpha1.ProxyMode_UNKNOWN_PROXY_MODE
	for _, aggInstance := range aggService.Instances {
		instance := convertAggregatedServiceInstance(aggInstance)
		instances = append(instances, instance)

		// Service proxy mode priority: ROUTER > SIDECAR > UNKNOWN_PROXY_MODE
		if aggInstance.ProxyMode == typesv1alpha1.ProxyMode_ROUTER {
			serviceProxyMode = typesv1alpha1.ProxyMode_ROUTER
		} else if aggInstance.ProxyMode == typesv1alpha1.ProxyMode_SIDECAR && serviceProxyMode == typesv1alpha1.ProxyMode_UNKNOWN_PROXY_MODE {
			serviceProxyMode = typesv1alpha1.ProxyMode_SIDECAR
		}
	}

	return &frontendv1alpha1.Service{
		Id:          aggService.ID,
		Name:        aggService.Name,
		Namespace:   aggService.Namespace,
		Instances:   instances,
		ClusterIps:  aggService.ClusterIPs,
		ExternalIps: aggService.ExternalIPs,
		ProxyMode:   serviceProxyMode,
	}
}

// convertAggregatedServiceInstance converts an AggregatedServiceInstance to the frontend API format
func convertAggregatedServiceInstance(aggInstance *connections.AggregatedServiceInstance) *frontendv1alpha1.ServiceInstance {
	return &frontendv1alpha1.ServiceInstance{
		InstanceId:   aggInstance.InstanceID,
		Ip:           aggInstance.IP,
		PodName:      aggInstance.PodName,
		Namespace:    aggInstance.Namespace,
		ClusterName:  aggInstance.ClusterName,
		EnvoyPresent: aggInstance.EnvoyPresent,
	}
}

// convertAggregatedServiceInstanceToDetail converts an AggregatedServiceInstance to the detailed frontend API format
func convertAggregatedServiceInstanceToDetail(aggInstance *connections.AggregatedServiceInstance) *frontendv1alpha1.ServiceInstanceDetail {
	// Convert container information
	containers := make([]*frontendv1alpha1.Container, len(aggInstance.Containers))
	for i, container := range aggInstance.Containers {
		containers[i] = &frontendv1alpha1.Container{
			Name:         container.Name,
			Image:        container.Image,
			Status:       container.Status,
			Ready:        container.Ready,
			RestartCount: container.RestartCount,
		}
	}

	return &frontendv1alpha1.ServiceInstanceDetail{
		InstanceId:     aggInstance.InstanceID,
		Ip:             aggInstance.IP,
		PodName:        aggInstance.PodName,
		Namespace:      aggInstance.Namespace,
		ClusterName:    aggInstance.ClusterName,
		EnvoyPresent:   aggInstance.EnvoyPresent,
		ServiceName:    fmt.Sprintf("%s:%s", aggInstance.Namespace, extractServiceNameFromInstanceID(aggInstance.InstanceID)),
		Containers:     containers,
		PodStatus:      aggInstance.PodStatus,
		NodeName:       aggInstance.NodeName,
		CreatedAt:      aggInstance.CreatedAt,
		Labels:         aggInstance.Labels,
		Annotations:    aggInstance.Annotations,
		IsEnvoyPresent: aggInstance.IsEnvoyPresent,
	}
}

// extractServiceNameFromInstanceID extracts the service name from an instance ID
// This is a helper function since we don't store service name directly in the instance
func extractServiceNameFromInstanceID(instanceID string) string {
	_, _, podName, err := parseInstanceID(instanceID)
	if err != nil {
		return "unknown"
	}
	// For now, return pod name as service identifier
	// In a more complete implementation, this would require a service lookup
	return podName
}
