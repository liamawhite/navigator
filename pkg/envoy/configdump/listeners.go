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
	"strings"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
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

	// Extract filter chains
	filterChains := p.getFilterChains(listener)
	summary.FilterChains = make([]*v1alpha1.FilterChainSummary, 0, len(filterChains))

	for _, fc := range filterChains {
		fcSummary := &v1alpha1.FilterChainSummary{
			Name: fc.Name,
		}

		// Extract filter chain match
		if match := fc.FilterChainMatch; match != nil {
			fcSummary.Match = &v1alpha1.FilterChainMatchInfo{
				ServerNames:          match.ServerNames,
				TransportProtocol:    match.TransportProtocol,
				ApplicationProtocols: match.ApplicationProtocols,
				DestinationPort:      match.DestinationPort.GetValue(),
			}

			// Extract prefix ranges
			for _, pr := range match.PrefixRanges {
				cidr := fmt.Sprintf("%s/%d", pr.AddressPrefix, pr.GetPrefixLen().GetValue())
				fcSummary.Match.DirectSourcePrefixRanges = append(fcSummary.Match.DirectSourcePrefixRanges, cidr)
			}
		}

		// Extract TLS context info
		if fc.TransportSocket != nil {
			fcSummary.TlsContext = &v1alpha1.TLSContextInfo{
				CommonTlsContext: true, // Simplified - just indicate TLS is present
			}
		}

		// Extract filters
		fcSummary.Filters = make([]*v1alpha1.FilterSummary, 0, len(fc.Filters))
		for _, filter := range fc.Filters {
			filterSummary := &v1alpha1.FilterSummary{
				Name: filter.Name,
			}

			// Extract typed config for known filter types
			switch filter.Name {
			case HTTPConnectionManager:
				hcmSummary := p.extractHTTPConnectionManagerSummary(filter)
				if hcmSummary != nil {
					filterSummary.TypedConfig = &v1alpha1.FilterSummary_HttpConnectionManager{
						HttpConnectionManager: hcmSummary,
					}
				}
			case TCPProxy:
				tcpSummary := p.extractTCPProxySummary(filter)
				if tcpSummary != nil {
					filterSummary.TypedConfig = &v1alpha1.FilterSummary_TcpProxy{
						TcpProxy: tcpSummary,
					}
				}
			}

			fcSummary.Filters = append(fcSummary.Filters, filterSummary)
		}

		summary.FilterChains = append(summary.FilterChains, fcSummary)
	}

	// Determine listener type based on name, address, port, and use_original_dst
	summary.Type = p.determineListenerType(summary.Name, summary.Address, summary.Port, summary.UseOriginalDst)

	// Extract listener filters
	for _, lf := range listener.ListenerFilters {
		summary.ListenerFilters = append(summary.ListenerFilters, &v1alpha1.ListenerFilterSummary{
			Name:            lf.Name,
			TypedConfigType: lf.GetTypedConfig().GetTypeUrl(), // Just the type URL for simplicity
		})
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

// extractHTTPConnectionManagerSummary extracts HCM configuration
func (p *Parser) extractHTTPConnectionManagerSummary(filter *listenerv3.Filter) *v1alpha1.HTTPConnectionManagerSummary {
	hcmConfig := &hcm.HttpConnectionManager{}
	if err := filter.GetTypedConfig().UnmarshalTo(hcmConfig); err != nil {
		return nil
	}

	summary := &v1alpha1.HTTPConnectionManagerSummary{
		CodecType:                hcmConfig.CodecType.String(),
		UseRemoteAddress:         hcmConfig.UseRemoteAddress.GetValue(),
		XffNumTrustedHops:        hcmConfig.XffNumTrustedHops,
		Via:                      hcmConfig.Via,
		GenerateRequestId:        hcmConfig.GenerateRequestId.GetValue(),
		ForwardClientCertDetails: hcmConfig.ForwardClientCertDetails.String(),
		ServerName:               hcmConfig.ServerName,
	}

	// Extract timeouts
	if hcmConfig.StreamIdleTimeout != nil {
		summary.StreamIdleTimeout = hcmConfig.StreamIdleTimeout.String()
	}
	if hcmConfig.RequestTimeout != nil {
		summary.RequestTimeout = hcmConfig.RequestTimeout.String()
	}
	if hcmConfig.DrainTimeout != nil {
		summary.DrainTimeout = hcmConfig.DrainTimeout.String()
	}
	if hcmConfig.DelayedCloseTimeout != nil {
		summary.DelayedCloseTimeout = hcmConfig.DelayedCloseTimeout.String()
	}

	// Extract route configuration (basic version)
	if routeConfig := hcmConfig.GetRouteConfig(); routeConfig != nil {
		summary.RouteConfig = &v1alpha1.RouteConfigInfo{
			Name:                routeConfig.Name,
			InternalOnlyHeaders: routeConfig.InternalOnlyHeaders,
			ValidateClusters:    routeConfig.ValidateClusters.GetValue(),
		}
	}

	// Extract RDS configuration (simplified)
	if rds := hcmConfig.GetRds(); rds != nil {
		summary.Rds = &v1alpha1.RDSInfo{
			RouteConfigName: rds.RouteConfigName,
		}
	}

	// Extract HTTP filters (simplified)
	for _, httpFilter := range hcmConfig.HttpFilters {
		summary.HttpFilters = append(summary.HttpFilters, &v1alpha1.HTTPFilterSummary{
			Name:            httpFilter.Name,
			TypedConfigType: httpFilter.GetTypedConfig().GetTypeUrl(),
		})
	}

	return summary
}

// extractTCPProxySummary extracts TCP proxy configuration
func (p *Parser) extractTCPProxySummary(filter *listenerv3.Filter) *v1alpha1.TCPProxySummary {
	// Skip black hole clusters
	if strings.Contains(string(filter.GetTypedConfig().GetValue()), BlackHoleCluster) {
		return nil
	}

	tcpProxy := &tcp.TcpProxy{}
	if err := filter.GetTypedConfig().UnmarshalTo(tcpProxy); err != nil {
		return nil
	}

	summary := &v1alpha1.TCPProxySummary{
		StatPrefix:                      tcpProxy.StatPrefix,
		MaxConnectAttempts:              tcpProxy.MaxConnectAttempts.GetValue(),
		MaxDownstreamConnectionDuration: tcpProxy.MaxDownstreamConnectionDuration.String(),
	}

	// Extract cluster specifier
	switch cs := tcpProxy.ClusterSpecifier.(type) {
	case *tcp.TcpProxy_Cluster:
		summary.Cluster = cs.Cluster
	case *tcp.TcpProxy_WeightedClusters:
		for _, wc := range cs.WeightedClusters.Clusters {
			summary.WeightedClusters = append(summary.WeightedClusters, &v1alpha1.WeightedClusterInfo{
				Name:   wc.Name,
				Weight: wc.Weight,
			})
		}
	}

	// Extract timeouts
	if tcpProxy.IdleTimeout != nil {
		summary.IdleTimeout = tcpProxy.IdleTimeout.String()
	}
	if tcpProxy.DownstreamIdleTimeout != nil {
		summary.DownstreamIdleTimeout = tcpProxy.DownstreamIdleTimeout.String()
	}
	if tcpProxy.UpstreamIdleTimeout != nil {
		summary.UpstreamIdleTimeout = tcpProxy.UpstreamIdleTimeout.String()
	}

	return summary
}
