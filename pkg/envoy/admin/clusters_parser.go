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

package admin

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// ClusterEndpointInfo represents live endpoint information from Envoy's /clusters endpoint
type ClusterEndpointInfo struct {
	ClusterName string
	Endpoints   []*v1alpha1.EndpointInfo
}

// ClustersResponse represents the JSON response from Envoy's /clusters?format=json endpoint
type ClustersResponse struct {
	ClusterStatuses []ClusterStatus `json:"cluster_statuses"`
}

// ClusterStatus represents a single cluster in the response
type ClusterStatus struct {
	Name         string       `json:"name"`
	AddedViaAPI  bool         `json:"added_via_api"`
	HostStatuses []HostStatus `json:"host_statuses"`
}

// HostStatus represents a single host/endpoint in a cluster
type HostStatus struct {
	Address      SocketAddress `json:"address"`
	HealthStatus HealthStatus  `json:"health_status"`
	Weight       uint32        `json:"weight"`
	Locality     Locality      `json:"locality"`
}

// SocketAddress represents the network address of an endpoint
type SocketAddress struct {
	SocketAddress *SocketAddressDetail `json:"socket_address"`
	Pipe          *PipeAddress         `json:"pipe"`
}

// SocketAddressDetail contains the actual IP and port
type SocketAddressDetail struct {
	Address   string `json:"address"`
	PortValue uint32 `json:"port_value"`
}

// PipeAddress represents a Unix domain socket path
type PipeAddress struct {
	Path string `json:"path"`
}

// HealthStatus represents the health status of an endpoint
type HealthStatus struct {
	EDSHealthStatus string `json:"eds_health_status"`
}

// Locality represents the locality information of an endpoint
type Locality struct {
	Region string `json:"region"`
	Zone   string `json:"zone"`
}

// ParseClustersOutput parses the output from Envoy's /clusters admin endpoint
// This extracts live endpoint information with health status and connection data
func ParseClustersOutput(clustersOutput string) ([]*ClusterEndpointInfo, error) {
	if clustersOutput == "" {
		return nil, nil
	}

	// Parse JSON response
	var response ClustersResponse
	if err := json.Unmarshal([]byte(clustersOutput), &response); err != nil {
		return nil, err
	}

	var clusterEndpoints []*ClusterEndpointInfo

	// Process each cluster in the response
	for _, cluster := range response.ClusterStatuses {
		if len(cluster.HostStatuses) == 0 {
			continue // Skip clusters with no endpoints
		}

		clusterInfo := &ClusterEndpointInfo{
			ClusterName: cluster.Name,
			Endpoints:   make([]*v1alpha1.EndpointInfo, 0, len(cluster.HostStatuses)),
		}

		// Process each endpoint in the cluster
		for _, host := range cluster.HostStatuses {
			var address string
			var port uint32
			var hostIdentifier string

			if host.Address.SocketAddress != nil {
				// TCP socket address
				address = host.Address.SocketAddress.Address
				port = host.Address.SocketAddress.PortValue
				hostIdentifier = address + ":" + strconv.Itoa(int(port))
			} else if host.Address.Pipe != nil {
				// Unix domain socket
				address = host.Address.Pipe.Path
				port = 0 // Pipes don't have ports
				hostIdentifier = "unix://" + address
			} else {
				// Skip endpoints with unknown address types
				continue
			}

			endpointInfo := &v1alpha1.EndpointInfo{
				Address:             address,
				Port:                port,
				HostIdentifier:      hostIdentifier,
				Health:              host.HealthStatus.EDSHealthStatus,
				Weight:              host.Weight,
				LoadBalancingWeight: host.Weight,
				Metadata:            make(map[string]string),
			}

			// Add locality information if available
			if host.Locality.Region != "" {
				endpointInfo.Metadata["region"] = host.Locality.Region
			}
			if host.Locality.Zone != "" {
				endpointInfo.Metadata["zone"] = host.Locality.Zone
			}

			clusterInfo.Endpoints = append(clusterInfo.Endpoints, endpointInfo)
		}

		clusterEndpoints = append(clusterEndpoints, clusterInfo)
	}

	return clusterEndpoints, nil
}

// MergeClusterEndpointsWithConfig merges live cluster endpoint data with static endpoint configuration
func MergeClusterEndpointsWithConfig(staticEndpoints []*v1alpha1.EndpointSummary, liveEndpoints []*ClusterEndpointInfo) []*v1alpha1.EndpointSummary {
	if len(liveEndpoints) == 0 {
		return staticEndpoints
	}

	// Create a map of live endpoints by cluster name for quick lookup
	liveEndpointMap := make(map[string]*ClusterEndpointInfo)
	for _, le := range liveEndpoints {
		liveEndpointMap[le.ClusterName] = le
	}

	// Merge with existing static endpoints
	mergedEndpoints := make([]*v1alpha1.EndpointSummary, 0, len(staticEndpoints))
	processedClusters := make(map[string]bool)

	// First, update existing static endpoints with live data
	for _, staticEndpoint := range staticEndpoints {
		if liveEndpoint, exists := liveEndpointMap[staticEndpoint.ClusterName]; exists {
			// Parse cluster name if static config is missing fields
			direction := staticEndpoint.Direction
			port := staticEndpoint.Port
			subset := staticEndpoint.Subset
			serviceFqdn := staticEndpoint.ServiceFqdn

			if direction == v1alpha1.ClusterDirection_UNSPECIFIED || port == 0 || serviceFqdn == "" {
				parsedDirection, parsedPort, parsedSubset, parsedServiceFqdn := parseClusterName(staticEndpoint.ClusterName)
				if direction == v1alpha1.ClusterDirection_UNSPECIFIED {
					direction = parsedDirection
				}
				if port == 0 {
					port = parsedPort
				}
				if subset == "" {
					subset = parsedSubset
				}
				if serviceFqdn == "" {
					serviceFqdn = parsedServiceFqdn
				}
			}

			// Merge live endpoint data into static config
			mergedEndpoint := &v1alpha1.EndpointSummary{
				ClusterName: staticEndpoint.ClusterName,
				ClusterType: staticEndpoint.ClusterType,
				Direction:   direction,
				Port:        port,
				Subset:      subset,
				ServiceFqdn: serviceFqdn,
				Endpoints:   liveEndpoint.Endpoints, // Use live endpoint data
			}
			mergedEndpoints = append(mergedEndpoints, mergedEndpoint)
			processedClusters[staticEndpoint.ClusterName] = true
		} else {
			// Keep static endpoint as-is if no live data available, but parse cluster name if needed
			endpoint := staticEndpoint
			if staticEndpoint.Direction == v1alpha1.ClusterDirection_UNSPECIFIED || staticEndpoint.Port == 0 || staticEndpoint.ServiceFqdn == "" {
				direction, port, subset, serviceFqdn := parseClusterName(staticEndpoint.ClusterName)
				endpoint = &v1alpha1.EndpointSummary{
					ClusterName: staticEndpoint.ClusterName,
					ClusterType: staticEndpoint.ClusterType,
					Direction:   direction,
					Port:        port,
					Subset:      subset,
					ServiceFqdn: serviceFqdn,
					Endpoints:   staticEndpoint.Endpoints,
				}
			}
			mergedEndpoints = append(mergedEndpoints, endpoint)
		}
	}

	// Add any live endpoints that don't have static configuration
	for clusterName, liveEndpoint := range liveEndpointMap {
		if !processedClusters[clusterName] {
			// Parse cluster name for Istio format
			direction, port, subset, serviceFqdn := parseClusterName(clusterName)

			newEndpoint := &v1alpha1.EndpointSummary{
				ClusterName: clusterName,
				Endpoints:   liveEndpoint.Endpoints,
				ClusterType: inferClusterType(clusterName),
				Direction:   direction,
				Port:        port,
				Subset:      subset,
				ServiceFqdn: serviceFqdn,
			}
			mergedEndpoints = append(mergedEndpoints, newEndpoint)
		}
	}

	return mergedEndpoints
}

// inferClusterType attempts to determine cluster type from cluster name
func inferClusterType(clusterName string) v1alpha1.ClusterType {
	clusterName = strings.ToLower(clusterName)

	if strings.Contains(clusterName, "outbound") || strings.Contains(clusterName, "inbound") {
		return v1alpha1.ClusterType_CLUSTER_EDS
	}

	if strings.Contains(clusterName, "static") {
		return v1alpha1.ClusterType_CLUSTER_STATIC
	}

	return v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE
}

// parseClusterName parses Istio cluster names in the format: direction|port|subset|serviceFQDN
// Examples:
//   - "outbound|8080||backend.demo.svc.cluster.local"
//   - "inbound|8080||"
//   - "outbound|443|v1|api.example.com"
func parseClusterName(clusterName string) (direction v1alpha1.ClusterDirection, port uint32, subset string, serviceFqdn string) {
	parts := strings.Split(clusterName, "|")

	// Default values
	direction = v1alpha1.ClusterDirection_UNSPECIFIED
	port = 0
	subset = ""
	serviceFqdn = ""

	if len(parts) >= 1 {
		// Parse direction
		switch strings.ToLower(parts[0]) {
		case "inbound":
			direction = v1alpha1.ClusterDirection_INBOUND
		case "outbound":
			direction = v1alpha1.ClusterDirection_OUTBOUND
		}
	}

	if len(parts) >= 2 {
		// Parse port
		if portValue, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
			port = uint32(portValue)
		}
	}

	if len(parts) >= 3 {
		// Parse subset (can be empty)
		subset = parts[2]
	}

	if len(parts) >= 4 {
		// Parse service FQDN
		serviceFqdn = parts[3]
	}

	return direction, port, subset, serviceFqdn
}
