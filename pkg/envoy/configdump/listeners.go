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

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// parseListenersFromAny extracts listener configurations from protobuf Any
func (p *Parser) parseListenersFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	listenerDump := &admin.ListenersConfigDump{}
	if err := configAny.UnmarshalTo(listenerDump); err != nil {
		return fmt.Errorf("failed to unmarshal listeners config dump: %w", err)
	}

	// Extract dynamic listeners (like istioctl)
	for _, l := range listenerDump.DynamicListeners {
		// Only process listeners with active state
		if l.ActiveState != nil && l.ActiveState.Listener != nil {
			var listener listenerv3.Listener
			if err := l.ActiveState.Listener.UnmarshalTo(&listener); err == nil {
				parsed.Listeners = append(parsed.Listeners, &listener)
				// Raw configuration will be populated by extractRawListenerConfigs
			}
		}
	}

	// Extract static listeners
	for _, l := range listenerDump.StaticListeners {
		if l.Listener != nil {
			var listener listenerv3.Listener
			if err := l.Listener.UnmarshalTo(&listener); err == nil {
				parsed.Listeners = append(parsed.Listeners, &listener)
				// Raw configuration will be populated by extractRawListenerConfigs
			}
		}
	}

	return nil
}

// summarizeListener converts a Listener config to a ListenerSummary
func (p *Parser) summarizeListener(listener *listenerv3.Listener, parsed *ParsedConfig) *v1alpha1.ListenerSummary {
	if listener == nil {
		return nil
	}

	summary := &v1alpha1.ListenerSummary{
		Name:           listener.Name,
		UseOriginalDst: listener.UseOriginalDst.GetValue(),
	}

	// Extract address information
	if listener.Address != nil {
		if sockAddr := listener.Address.GetSocketAddress(); sockAddr != nil {
			summary.Address = sockAddr.Address
			summary.Port = sockAddr.GetPortValue()
		}
	}

	// Determine listener type based on name, address, port, and use_original_dst
	summary.Type = p.determineListenerType(summary.Name, summary.Address, summary.Port, summary.UseOriginalDst)

	// Store raw config for debugging
	if listener != nil {
		summary.RawConfig = listener.String()
	}

	// Use the raw JSON config that was extracted directly from the original config dump
	if rawJSON, exists := parsed.RawListeners[listener.Name]; exists {
		summary.RawConfig = rawJSON
	}

	return summary
}

// getFilterChains returns all filter chains for a listener
func (p *Parser) getFilterChains(l *listenerv3.Listener) []*listenerv3.FilterChain {
	res := l.FilterChains
	if l.DefaultFilterChain != nil {
		res = append(res, l.DefaultFilterChain)
	}
	return res
}

// determineListenerType determines the listener type based on name, address, port, and use_original_dst
func (p *Parser) determineListenerType(name, address string, port uint32, useOriginalDst bool) v1alpha1.ListenerType {
	// First: Check for virtual listeners by name (most reliable)
	if name == "virtualInbound" {
		return v1alpha1.ListenerType_VIRTUAL_INBOUND
	}
	if name == "virtualOutbound" {
		return v1alpha1.ListenerType_VIRTUAL_OUTBOUND
	}

	// Second: Check for well-known proxy-specific ports on 0.0.0.0
	if address == "0.0.0.0" {
		switch port {
		case 15090:
			// Prometheus metrics endpoint
			return v1alpha1.ListenerType_PROXY_METRICS
		case 15021:
			// Health check endpoint
			return v1alpha1.ListenerType_PROXY_HEALTHCHECK
		}

		// Check for virtual listener patterns on 0.0.0.0
		// Virtual outbound: typically port 15001 with use_original_dst
		if (port == 15001 || port == 15006) && useOriginalDst {
			return v1alpha1.ListenerType_VIRTUAL_OUTBOUND
		}

		// All other 0.0.0.0 listeners are port-based (generic port traffic)
		// This includes application ports like 80, 8080, etc.
		return v1alpha1.ListenerType_PORT_OUTBOUND
	}

	// Third: Specific IP addresses indicate outbound listeners (no inbound listeners in modern Istio)
	// Distinguish between service-specific and port-based patterns

	// Service-specific listeners: IP addresses that look like service cluster IPs
	// These typically have full service.namespace.svc.cluster.local resolution
	// Pattern: specific cluster IP with service port
	if p.isServiceSpecificListener(name, address, port) {
		return v1alpha1.ListenerType_SERVICE_OUTBOUND
	}

	// Port-based listeners: Generic port traffic patterns
	// These handle broader traffic patterns by port
	return v1alpha1.ListenerType_PORT_OUTBOUND
}

// isServiceSpecificListener determines if a listener is service-specific vs port-based
func (p *Parser) isServiceSpecificListener(name, address string, port uint32) bool {
	// Port-based listeners: 0.0.0.0 listeners that are not admin/virtual
	// These handle generic port traffic patterns
	if address == "0.0.0.0" {
		// Already handled virtual and admin listeners above, so these are port-based
		return false
	}

	// Service-specific listeners: Any specific IP address
	// These are for outbound connections to specific services
	return true
}

// Simplified listener processing - detailed filter analysis removed
// as it's not part of the current simplified schema
