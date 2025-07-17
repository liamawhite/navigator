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

package configdump

import (
	"fmt"
	"net"
	"strconv"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// parseEndpointsFromAny extracts endpoint configurations from protobuf Any
func (p *Parser) parseEndpointsFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	endpointDump := &admin.EndpointsConfigDump{}
	if err := configAny.UnmarshalTo(endpointDump); err != nil {
		return fmt.Errorf("failed to unmarshal endpoints config dump: %w", err)
	}

	// Extract dynamic endpoints (like istioctl)
	for _, e := range endpointDump.DynamicEndpointConfigs {
		if e.EndpointConfig != nil {
			var endpoint endpointv3.ClusterLoadAssignment
			if err := e.EndpointConfig.UnmarshalTo(&endpoint); err == nil {
				parsed.Endpoints = append(parsed.Endpoints, &endpoint)
			}
		}
	}

	// Extract static endpoints
	for _, e := range endpointDump.StaticEndpointConfigs {
		if e.EndpointConfig != nil {
			var endpoint endpointv3.ClusterLoadAssignment
			if err := e.EndpointConfig.UnmarshalTo(&endpoint); err == nil {
				parsed.Endpoints = append(parsed.Endpoints, &endpoint)
			}
		}
	}

	return nil
}

// summarizeEndpoint converts an endpoint config to an EndpointSummary
func (p *Parser) summarizeEndpoint(endpoints *endpointv3.ClusterLoadAssignment) *v1alpha1.EndpointSummary {
	if endpoints == nil {
		return nil
	}

	summary := &v1alpha1.EndpointSummary{
		ClusterName: endpoints.ClusterName,
	}

	for _, localityEndpoints := range endpoints.Endpoints {
		for _, lbEndpoint := range localityEndpoints.LbEndpoints {
			endpointInfo := &v1alpha1.EndpointInfo{
				Health:              lbEndpoint.HealthStatus.String(),
				LoadBalancingWeight: lbEndpoint.LoadBalancingWeight.GetValue(),
				Priority:            localityEndpoints.Priority,
				Weight:              localityEndpoints.LoadBalancingWeight.GetValue(),
			}

			if endpoint := lbEndpoint.GetEndpoint(); endpoint != nil {
				if addr := endpoint.Address; addr != nil {
					if sockAddr := addr.GetSocketAddress(); sockAddr != nil {
						endpointInfo.Address = sockAddr.Address
						endpointInfo.Port = sockAddr.GetPortValue()
					}
				}

				if endpoint.Hostname != "" {
					endpointInfo.HostIdentifier = endpoint.Hostname
				} else {
					endpointInfo.HostIdentifier = net.JoinHostPort(endpointInfo.Address, strconv.Itoa(int(endpointInfo.Port)))
				}
			}

			// Extract metadata as simple string map
			if lbEndpoint.Metadata != nil && lbEndpoint.Metadata.FilterMetadata != nil {
				endpointInfo.Metadata = make(map[string]string)
				for k, v := range lbEndpoint.Metadata.FilterMetadata {
					if v.Fields != nil {
						for fk, fv := range v.Fields {
							if fv.GetStringValue() != "" {
								endpointInfo.Metadata[fmt.Sprintf("%s.%s", k, fk)] = fv.GetStringValue()
							}
						}
					}
				}
			}

			summary.Endpoints = append(summary.Endpoints, endpointInfo)
		}
	}

	return summary
}
