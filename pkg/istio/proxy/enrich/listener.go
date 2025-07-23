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

package enrich

import (
	"strings"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// enrichListenerType classifies listener type based on Istio-specific patterns
func enrichListenerType(proxyMode v1alpha1.ProxyMode) func(*v1alpha1.ListenerSummary) error {
	return func(listener *v1alpha1.ListenerSummary) error {
		if listener == nil {
			return nil
		}

		listener.Type = inferIstioListenerType(listener.Name, listener.Address, listener.Port, listener.UseOriginalDst, proxyMode)

		// Enrich matches and destinations with Istio-specific information
		if err := enrichListenerMatchDestination()(listener); err != nil {
			return err
		}

		return nil
	}
}

// inferIstioListenerType applies Istio-specific listener type detection
func inferIstioListenerType(name, address string, port uint32, useOriginalDst bool, proxyMode v1alpha1.ProxyMode) v1alpha1.ListenerType {
	// Check for Istio virtual listeners by name
	if name == "virtualInbound" {
		return v1alpha1.ListenerType_VIRTUAL_INBOUND
	}
	if name == "virtualOutbound" {
		return v1alpha1.ListenerType_VIRTUAL_OUTBOUND
	}

	// Check for Istio-specific ports on 0.0.0.0
	if address == "0.0.0.0" {
		switch port {
		case 15090:
			// Prometheus metrics endpoint
			return v1alpha1.ListenerType_PROXY_METRICS
		case 15021:
			// Health check endpoint
			return v1alpha1.ListenerType_PROXY_HEALTHCHECK
		case 15001, 15006:
			// Istio virtual outbound listeners
			if useOriginalDst {
				return v1alpha1.ListenerType_VIRTUAL_OUTBOUND
			}
		}

		// Gateway-specific logic: 0.0.0.0 listeners without useOriginalDst
		if proxyMode == v1alpha1.ProxyMode_GATEWAY && !useOriginalDst {
			return v1alpha1.ListenerType_GATEWAY_INBOUND
		}

		// Other 0.0.0.0 listeners are port-based (for sidecars)
		return v1alpha1.ListenerType_PORT_OUTBOUND
	}

	// Specific IP addresses in Istio typically indicate service-specific outbound listeners
	return v1alpha1.ListenerType_SERVICE_OUTBOUND
}

// enrichListenerMatchDestination enriches listener matches and destinations with Istio-specific information
func enrichListenerMatchDestination() func(*v1alpha1.ListenerSummary) error {
	return func(listener *v1alpha1.ListenerSummary) error {
		if listener == nil {
			return nil
		}

		// Enrich rules with Istio-specific information
		for _, rule := range listener.Rules {
			if rule == nil {
				continue
			}

			// Enrich destination with Istio service FQDN information
			if rule.Destination != nil && rule.Destination.ClusterName != "" {
				enrichDestinationWithIstioInfo(rule.Destination)
			}

			// Classify match types based on listener type and Istio patterns
			if rule.Match != nil {
				enrichMatchWithIstioInfo(rule.Match, listener)
			}
		}

		return nil
	}
}

// enrichDestinationWithIstioInfo enriches destination with Istio-specific service information
func enrichDestinationWithIstioInfo(destination *v1alpha1.ListenerDestination) {
	if destination == nil || destination.ClusterName == "" {
		return
	}

	// Parse Istio cluster names (e.g., "outbound|80|v1|myservice.mynamespace.svc.cluster.local")
	_, port, _, serviceFQDN := ParseClusterNameComponents(destination.ClusterName)
	if serviceFQDN != "" {
		destination.ServiceFqdn = serviceFQDN
	}
	if port > 0 {
		destination.Port = port
	}

	// Handle special Istio destinations
	switch {
	case strings.Contains(destination.ClusterName, "PassthroughCluster"):
		destination.DestinationType = "passthrough"
	case strings.Contains(destination.ClusterName, "BlackHoleCluster"):
		destination.DestinationType = "blackhole"
	case strings.HasPrefix(destination.ClusterName, "inbound"):
		destination.DestinationType = "inbound"
	case strings.HasPrefix(destination.ClusterName, "outbound"):
		destination.DestinationType = "outbound"
	}
}

// enrichMatchWithIstioInfo enriches match with Istio-specific context information
func enrichMatchWithIstioInfo(match *v1alpha1.ListenerMatch, listener *v1alpha1.ListenerSummary) {
	if match == nil || listener == nil {
		return
	}

	// Use type switch to handle different match types
	switch matchType := match.MatchType.(type) {
	case *v1alpha1.ListenerMatch_HttpRoute:
		enrichHttpRouteMatch(matchType.HttpRoute, listener)
	case *v1alpha1.ListenerMatch_FilterChain:
		enrichFilterChainMatch(matchType.FilterChain, listener)
	case *v1alpha1.ListenerMatch_TcpProxy:
		enrichTcpProxyMatch(matchType.TcpProxy, listener)
	}
}

// enrichHttpRouteMatch enriches HTTP route matches with Istio context
func enrichHttpRouteMatch(httpRoute *v1alpha1.HttpRouteMatch, listener *v1alpha1.ListenerSummary) {
	if httpRoute == nil {
		return
	}

	// Enrich path matches with Istio context
	if httpRoute.PathMatch != nil && httpRoute.PathMatch.Path != "" {
		// Identify common Istio routing patterns
		path := httpRoute.PathMatch.Path
		switch {
		case strings.HasPrefix(path, "/stats"):
			httpRoute.PathMatch.MatchType = "istio_stats"
		case strings.HasPrefix(path, "/health"):
			httpRoute.PathMatch.MatchType = "istio_health"
		case strings.HasPrefix(path, "/ready"):
			httpRoute.PathMatch.MatchType = "istio_ready"
		}
	}

	// Enrich header matches with Istio context
	for _, headerMatch := range httpRoute.HeaderMatches {
		if headerMatch.Name == ":authority" || headerMatch.Name == "host" {
			// Authority/Host headers in Istio often contain service FQDNs
			if strings.Contains(headerMatch.Value, ".svc.cluster.local") {
				headerMatch.Name = "istio_service_host"
			}
		}
	}
}

// enrichFilterChainMatch enriches filter chain matches with Istio context
func enrichFilterChainMatch(filterChain *v1alpha1.FilterChainMatch, listener *v1alpha1.ListenerSummary) {
	if filterChain == nil {
		return
	}

	// Enrich SNI matching patterns
	for i, serverName := range filterChain.ServerNames {
		if strings.Contains(serverName, ".svc.cluster.local") {
			filterChain.ServerNames[i] = "istio_service_" + serverName
		}
	}

	// Classify transport protocol for Istio context
	if filterChain.TransportProtocol == "tls" {
		// Check for Istio-specific TLS patterns
		for _, serverName := range filterChain.ServerNames {
			if strings.HasSuffix(serverName, ".istio-system.svc.cluster.local") {
				filterChain.TransportProtocol = "istio_mtls"
				break
			}
		}
	}
}

// enrichTcpProxyMatch enriches TCP proxy matches with Istio context
func enrichTcpProxyMatch(tcpProxy *v1alpha1.TcpProxyMatch, listener *v1alpha1.ListenerSummary) {
	if tcpProxy == nil {
		return
	}

	// Parse cluster name for Istio-specific information
	if tcpProxy.ClusterName != "" {
		_, _, _, serviceFQDN := ParseClusterNameComponents(tcpProxy.ClusterName)
		if serviceFQDN != "" && strings.Contains(serviceFQDN, ".svc.cluster.local") {
			// This is an Istio service cluster
			tcpProxy.ClusterName = "istio_service_" + tcpProxy.ClusterName
		}
	}
}
