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

// Package clusters provides utilities for parsing Envoy clusters API responses.
// This follows the same approach as istioctl's proxy-config endpoints command.
package clusters

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"istio.io/istio/pkg/util/protomarshal"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// Wrapper is a wrapper around the Envoy Clusters admin response
// It provides protobuf marshaling/unmarshaling support
type Wrapper struct {
	*admin.Clusters
}

// MarshalJSON is a custom marshaller to handle protobuf pain
func (w *Wrapper) MarshalJSON() ([]byte, error) {
	return protomarshal.Marshal(w)
}

// UnmarshalJSON is a custom unmarshaller to handle protobuf pain
func (w *Wrapper) UnmarshalJSON(b []byte) error {
	cd := &admin.Clusters{}
	err := protomarshal.UnmarshalAllowUnknown(b, cd)
	*w = Wrapper{cd}
	return err
}

// Parser handles parsing of Envoy clusters API responses
type Parser struct{}

// NewParser creates a new Envoy clusters parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseJSON parses a raw Envoy clusters API JSON response into endpoint summaries
func (p *Parser) ParseJSON(rawClustersResponse string) ([]*v1alpha1.EndpointSummary, error) {
	wrapper := &Wrapper{}
	if err := wrapper.UnmarshalJSON([]byte(rawClustersResponse)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clusters response: %w", err)
	}

	var summaries []*v1alpha1.EndpointSummary

	// Parse each cluster and extract endpoint information
	for _, cluster := range wrapper.ClusterStatuses {
		if len(cluster.HostStatuses) == 0 {
			continue
		}

		summary := &v1alpha1.EndpointSummary{
			ClusterName: cluster.Name,
			ClusterType: p.inferClusterType(cluster.Name),
		}

		// Parse cluster name components (format: direction|port|subset|servicefqdn)
		p.parseClusterName(cluster.Name, summary)

		for _, host := range cluster.HostStatuses {
			endpointInfo := p.convertHostToEndpoint(host)
			if endpointInfo != nil {
				summary.Endpoints = append(summary.Endpoints, endpointInfo)
			}
		}

		if len(summary.Endpoints) > 0 {
			summaries = append(summaries, summary)
		}
	}

	return summaries, nil
}

// convertHostToEndpoint converts an Envoy HostStatus to our EndpointInfo
func (p *Parser) convertHostToEndpoint(host *admin.HostStatus) *v1alpha1.EndpointInfo {
	if host == nil {
		return nil
	}

	endpoint := &v1alpha1.EndpointInfo{
		Health:              p.getHealthStatus(host),
		Priority:            host.Priority,                             // Extract from HostStatus
		Weight:              0,                                         // Default
		LoadBalancingWeight: 0,                                         // Default
		Metadata:            make(map[string]string),                   // Always initialize
		AddressType:         v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE, // Explicit default
		Address:             "unknown",                                 // Default
		HostIdentifier:      "unknown",                                 // Default
	}

	// Extract address and port
	if addr := host.Address.GetSocketAddress(); addr != nil {
		endpoint.Address = addr.Address
		endpoint.Port = addr.GetPortValue()
		endpoint.HostIdentifier = net.JoinHostPort(endpoint.Address, strconv.Itoa(int(endpoint.Port)))

		// Check if the socket address is actually a pipe/unix socket path
		if strings.HasPrefix(addr.Address, "./") || strings.HasPrefix(addr.Address, "/") || strings.Contains(addr.Address, "/socket") {
			endpoint.AddressType = v1alpha1.AddressType_PIPE_ADDRESS
			endpoint.HostIdentifier = endpoint.Address
		} else {
			endpoint.AddressType = v1alpha1.AddressType_SOCKET_ADDRESS
		}
	} else if pipe := host.Address.GetPipe(); pipe != nil {
		endpoint.Address = "unix://" + pipe.Path
		endpoint.HostIdentifier = endpoint.Address
		endpoint.AddressType = v1alpha1.AddressType_PIPE_ADDRESS
	} else if internal := host.Address.GetEnvoyInternalAddress(); internal != nil {
		endpoint.AddressType = v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS
		switch an := internal.GetAddressNameSpecifier().(type) {
		case *core.EnvoyInternalAddress_ServerListenerName:
			endpoint.Address = fmt.Sprintf("envoy://%s/%s", an.ServerListenerName, internal.EndpointId)
			endpoint.HostIdentifier = endpoint.Address
		default:
			// Fallback for other address name specifier types
			endpoint.Address = internal.EndpointId
			endpoint.HostIdentifier = internal.EndpointId
		}
	} else {
		endpoint.Address = "unknown"
		endpoint.HostIdentifier = "unknown"
		endpoint.AddressType = v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE
	}

	// Note: Priority is available from clusters API HostStatus.
	// Weight and load balancing weight are not directly available and would need to come from EDS config dump.

	return endpoint
}

// getHealthStatus converts Envoy health status to string
func (p *Parser) getHealthStatus(host *admin.HostStatus) string {
	if host.HealthStatus == nil {
		return "UNKNOWN"
	}

	status := host.HealthStatus.GetEdsHealthStatus()
	return core.HealthStatus_name[int32(status)]
}

// GetOutlierCheckStatus returns whether the outlier check failed
func (p *Parser) GetOutlierCheckStatus(host *admin.HostStatus) bool {
	if host.HealthStatus == nil {
		return false
	}
	return host.HealthStatus.GetFailedOutlierCheck()
}

// inferClusterType infers the cluster type from cluster name patterns
// This follows similar logic to how istioctl and Envoy classify clusters
func (p *Parser) inferClusterType(clusterName string) v1alpha1.ClusterType {
	if clusterName == "" {
		return v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE
	}

	// Istio cluster name patterns:
	// - outbound|<port>|<subset>|<service_fqdn> -> EDS (service discovery)
	// - inbound|<port>|<subset>|<service_fqdn> -> EDS (service discovery)
	// - Static clusters (prometheus_stats, agent, xds-grpc, sds-grpc, etc.) -> STATIC
	// - DNS-based external services -> STRICT_DNS or LOGICAL_DNS

	// Check for Istio sidecar patterns (most common)
	if strings.HasPrefix(clusterName, "outbound|") || strings.HasPrefix(clusterName, "inbound|") {
		// These are typically EDS clusters for service discovery
		parts := strings.Split(clusterName, "|")
		if len(parts) >= 4 {
			serviceFqdn := parts[3]
			// If it looks like a Kubernetes service FQDN, it's EDS
			if strings.Contains(serviceFqdn, ".svc.cluster.local") {
				return v1alpha1.ClusterType_CLUSTER_EDS
			}
			// External services might be DNS-based
			if strings.Contains(serviceFqdn, ".") && !strings.Contains(serviceFqdn, ".svc.cluster.local") {
				return v1alpha1.ClusterType_CLUSTER_STRICT_DNS
			}
		}
		return v1alpha1.ClusterType_CLUSTER_EDS // Default for outbound/inbound patterns
	}

	// Static clusters (common Istio/Envoy internal clusters)
	staticClusters := []string{
		"prometheus_stats",
		"agent",
		"sds-grpc",
		"xds-grpc",
		"zipkin",
		"jaeger",
		"envoy_accesslog_service",
	}

	for _, staticCluster := range staticClusters {
		if clusterName == staticCluster {
			return v1alpha1.ClusterType_CLUSTER_STATIC
		}
	}

	// If cluster name contains an IP address, it's likely static
	if strings.Contains(clusterName, ".") && !strings.Contains(clusterName, "svc.cluster.local") {
		// Simple heuristic: if it looks like an external domain, it's DNS-based
		if strings.Count(clusterName, ".") >= 2 {
			return v1alpha1.ClusterType_CLUSTER_STRICT_DNS
		}
		return v1alpha1.ClusterType_CLUSTER_STATIC
	}

	// Default fallback
	return v1alpha1.ClusterType_CLUSTER_EDS
}

// parseClusterName parses Istio cluster names in the format: direction|port|subset|servicefqdn
func (p *Parser) parseClusterName(clusterName string, summary *v1alpha1.EndpointSummary) {
	// Default values
	summary.Direction = v1alpha1.ClusterDirection_UNSPECIFIED
	summary.Port = 0
	summary.Subset = ""
	summary.ServiceFqdn = ""

	// Split by pipe character
	parts := strings.Split(clusterName, "|")
	if len(parts) != 4 {
		// Not in expected format, leave defaults
		return
	}

	// Parse direction
	switch strings.ToLower(parts[0]) {
	case "inbound":
		summary.Direction = v1alpha1.ClusterDirection_INBOUND
	case "outbound":
		summary.Direction = v1alpha1.ClusterDirection_OUTBOUND
	default:
		summary.Direction = v1alpha1.ClusterDirection_UNSPECIFIED
	}

	// Parse port
	if port, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
		summary.Port = uint32(port)
	}

	// Parse subset (may be empty)
	summary.Subset = parts[2]

	// Parse service FQDN
	summary.ServiceFqdn = parts[3]
}
