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

package clusters

import (
	"encoding/json"
	"fmt"
	"net"
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
	Address      AdminAddress `json:"address"`
	HealthStatus HealthStatus `json:"health_status"`
	Weight       uint32       `json:"weight"`
	Locality     Locality     `json:"locality"`
}

// AdminAddress represents the network address of an endpoint from admin interface
type AdminAddress struct {
	SocketAddress        *SocketAddressDetail        `json:"socket_address"`
	Pipe                 *PipeAddress                `json:"pipe"`
	EnvoyInternalAddress *EnvoyInternalAddressDetail `json:"envoy_internal_address"`
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

// EnvoyInternalAddressDetail represents an Envoy internal address (ambient mode)
type EnvoyInternalAddressDetail struct {
	ServerListenerName string `json:"server_listener_name"`
	EndpointId         string `json:"endpoint_id"`
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

// ParseClustersAdminOutput parses the output from Envoy's /clusters admin endpoint
// This extracts live endpoint information with health status and connection data
func ParseClustersAdminOutput(clustersOutput string) ([]*ClusterEndpointInfo, error) {
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
			endpointInfo := convertAdminHostToEndpoint(host)
			if endpointInfo != nil {
				clusterInfo.Endpoints = append(clusterInfo.Endpoints, endpointInfo)
			}
		}

		if len(clusterInfo.Endpoints) > 0 {
			clusterEndpoints = append(clusterEndpoints, clusterInfo)
		}
	}

	return clusterEndpoints, nil
}

// ConvertToEndpointSummaries converts ClusterEndpointInfo directly to EndpointSummary format
// This bypasses the need for merging with config dump data and provides clusters-only endpoint information
func ConvertToEndpointSummaries(clusterEndpoints []*ClusterEndpointInfo) []*v1alpha1.EndpointSummary {
	var summaries []*v1alpha1.EndpointSummary

	for _, clusterInfo := range clusterEndpoints {
		if len(clusterInfo.Endpoints) == 0 {
			continue // Skip clusters with no endpoints
		}

		// Parse cluster name components for Istio format
		direction, port, subset, serviceFqdn := parseClusterNameComponents(clusterInfo.ClusterName)

		summary := &v1alpha1.EndpointSummary{
			ClusterName: clusterInfo.ClusterName,
			ClusterType: inferClusterTypeFromName(clusterInfo.ClusterName),
			Direction:   direction,
			Port:        port,
			Subset:      subset,
			ServiceFqdn: serviceFqdn,
			Endpoints:   clusterInfo.Endpoints,
		}

		summaries = append(summaries, summary)
	}

	return summaries
}

// convertAdminHostToEndpoint converts an admin HostStatus to our EndpointInfo with proper address type detection
func convertAdminHostToEndpoint(host HostStatus) *v1alpha1.EndpointInfo {
	endpointInfo := &v1alpha1.EndpointInfo{
		Health:              host.HealthStatus.EDSHealthStatus,
		Weight:              host.Weight,
		LoadBalancingWeight: host.Weight,
		Metadata:            make(map[string]string),
		AddressType:         v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE, // Default
		Address:             "unknown",                                 // Default
		HostIdentifier:      "unknown",                                 // Default
	}

	// Handle different address types
	if host.Address.SocketAddress != nil {
		// TCP socket address
		addr := host.Address.SocketAddress
		endpointInfo.Address = addr.Address
		endpointInfo.Port = addr.PortValue
		endpointInfo.HostIdentifier = net.JoinHostPort(endpointInfo.Address, strconv.Itoa(int(endpointInfo.Port)))

		// Check if the socket address is actually a pipe/unix socket path
		if strings.HasPrefix(addr.Address, "./") || strings.HasPrefix(addr.Address, "/") || strings.Contains(addr.Address, "/socket") {
			endpointInfo.AddressType = v1alpha1.AddressType_PIPE_ADDRESS
			endpointInfo.HostIdentifier = endpointInfo.Address
		} else {
			endpointInfo.AddressType = v1alpha1.AddressType_SOCKET_ADDRESS
		}
	} else if host.Address.Pipe != nil {
		// Unix domain socket
		endpointInfo.Address = "unix://" + host.Address.Pipe.Path
		endpointInfo.Port = 0 // Pipes don't have ports
		endpointInfo.HostIdentifier = endpointInfo.Address
		endpointInfo.AddressType = v1alpha1.AddressType_PIPE_ADDRESS
	} else if host.Address.EnvoyInternalAddress != nil {
		// Envoy internal address (ambient mode)
		internal := host.Address.EnvoyInternalAddress
		endpointInfo.Address = internal.EndpointId
		endpointInfo.Port = 0 // Internal addresses don't have traditional ports
		endpointInfo.HostIdentifier = fmt.Sprintf("envoy://%s/%s", internal.ServerListenerName, internal.EndpointId)
		endpointInfo.AddressType = v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS
	} else {
		// Unknown address type
		endpointInfo.Address = "unknown"
		endpointInfo.HostIdentifier = "unknown"
		endpointInfo.AddressType = v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE
		return nil // Skip endpoints with unknown address types
	}

	// Add locality information if available
	if host.Locality.Region != "" {
		endpointInfo.Metadata["region"] = host.Locality.Region
	}
	if host.Locality.Zone != "" {
		endpointInfo.Metadata["zone"] = host.Locality.Zone
	}

	return endpointInfo
}

// MergeWithStaticConfig merges live cluster endpoint data with static endpoint configuration
func MergeWithStaticConfig(staticEndpoints []*v1alpha1.EndpointSummary, liveEndpoints []*ClusterEndpointInfo) []*v1alpha1.EndpointSummary {
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
				parsedDirection, parsedPort, parsedSubset, parsedServiceFqdn := parseClusterNameComponents(staticEndpoint.ClusterName)
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
				direction, port, subset, serviceFqdn := parseClusterNameComponents(staticEndpoint.ClusterName)
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
			direction, port, subset, serviceFqdn := parseClusterNameComponents(clusterName)

			newEndpoint := &v1alpha1.EndpointSummary{
				ClusterName: clusterName,
				Endpoints:   liveEndpoint.Endpoints,
				ClusterType: inferClusterTypeFromName(clusterName),
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

// inferClusterTypeFromName attempts to determine cluster type from cluster name
func inferClusterTypeFromName(clusterName string) v1alpha1.ClusterType {
	clusterName = strings.ToLower(clusterName)

	if strings.Contains(clusterName, "outbound") || strings.Contains(clusterName, "inbound") {
		return v1alpha1.ClusterType_CLUSTER_EDS
	}

	if strings.Contains(clusterName, "static") {
		return v1alpha1.ClusterType_CLUSTER_STATIC
	}

	return v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE
}

// parseClusterNameComponents parses Istio cluster names in the format: direction|port|subset|serviceFQDN
// Examples:
//   - "outbound|8080||backend.demo.svc.cluster.local"
//   - "inbound|8080||"
//   - "outbound|443|v1|api.example.com"
func parseClusterNameComponents(clusterName string) (direction v1alpha1.ClusterDirection, port uint32, subset string, serviceFqdn string) {
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
