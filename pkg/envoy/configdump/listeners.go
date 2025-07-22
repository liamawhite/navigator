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

	// Extract dynamic listeners
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

	// For generic Envoy deployments, use basic type classification
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

// determineListenerType determines the listener type for generic Envoy deployments
func (p *Parser) determineListenerType(name, address string, port uint32, useOriginalDst bool) v1alpha1.ListenerType {
	// Basic classification for generic Envoy deployments
	// Service mesh specific logic should be handled by enrichment layers

	// Check for 0.0.0.0 listeners (typically inbound or catch-all)
	if address == "0.0.0.0" {
		if useOriginalDst {
			return v1alpha1.ListenerType_VIRTUAL_OUTBOUND
		}
		return v1alpha1.ListenerType_PORT_OUTBOUND
	}

	// Specific IP addresses typically indicate outbound listeners
	return v1alpha1.ListenerType_SERVICE_OUTBOUND
}

// Simplified listener processing - detailed filter analysis removed
// as it's not part of the current simplified schema
