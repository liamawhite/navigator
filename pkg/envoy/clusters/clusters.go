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
// This provides generic Envoy cluster parsing without service mesh assumptions.
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
			ClusterType: v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
			Direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
			Port:        0,
			Subset:      "",
			ServiceFqdn: cluster.Name, // Use cluster name as-is
		}

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
		Health:         p.getHealthStatus(host),
		Priority:       p.getPriority(host),                       // Extract actual priority
		Weight:         p.getWeight(host),                         // Extract actual weight
		Metadata:       make(map[string]string),                   // Always initialize
		AddressType:    v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE, // Explicit default
		Address:        "unknown",                                 // Default
		HostIdentifier: "unknown",                                 // Default
		Locality:       p.getLocality(host),                       // Extract locality information
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
	} else {
		endpoint.Address = "unknown"
		endpoint.HostIdentifier = "unknown"
		endpoint.AddressType = v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE
	}

	// Note: Priority and weight are extracted from clusters API HostStatus.

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

// getWeight extracts the weight from HostStatus
func (p *Parser) getWeight(host *admin.HostStatus) uint32 {
	if host == nil {
		return 1 // Default weight is 1 according to Envoy docs
	}
	weight := host.GetWeight()
	if weight == 0 {
		return 1 // Default weight is 1 when not explicitly set (JSON unmarshaling sets 0 for unset fields)
	}
	return weight
}

// getPriority extracts the priority from HostStatus
func (p *Parser) getPriority(host *admin.HostStatus) uint32 {
	if host == nil {
		return 0 // Default priority is 0 according to Envoy docs
	}
	return host.GetPriority()
}

// getLocality extracts locality information from HostStatus
func (p *Parser) getLocality(host *admin.HostStatus) *v1alpha1.LocalityInfo {
	if host == nil {
		return nil
	}

	envoyLocality := host.GetLocality()
	if envoyLocality == nil {
		return nil
	}

	return &v1alpha1.LocalityInfo{
		Region: envoyLocality.Region,
		Zone:   envoyLocality.Zone,
	}
}
