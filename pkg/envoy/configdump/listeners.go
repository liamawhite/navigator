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
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcp_proxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
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

	// Set default type - enrichment layers will apply specific classification
	summary.Type = v1alpha1.ListenerType_UNKNOWN_LISTENER_TYPE

	// Store raw config for debugging
	if listener != nil {
		summary.RawConfig = listener.String()
	}

	// Use the raw JSON config that was extracted directly from the original config dump
	if rawJSON, exists := parsed.RawListeners[listener.Name]; exists {
		summary.RawConfig = rawJSON
	}

	// Parse listener filters for rules and filter chain info
	rules, filterChains := p.parseListenerFilters(listener)
	summary.Rules = rules
	summary.FilterChains = filterChains

	return summary
}

// parseListenerFilters extracts matched rules and filter chain info from listener
func (p *Parser) parseListenerFilters(listener *listenerv3.Listener) ([]*v1alpha1.ListenerRule, *v1alpha1.FilterChainSummary) {
	if listener == nil || len(listener.FilterChains) == 0 {
		return nil, &v1alpha1.FilterChainSummary{}
	}

	var allRules []*v1alpha1.ListenerRule
	var httpFilters []*v1alpha1.FilterInfo
	var networkFilters []*v1alpha1.FilterInfo
	hasTLS := false

	for _, filterChain := range listener.FilterChains {
		// Check for TLS context
		if filterChain.TransportSocket != nil {
			hasTLS = true
		}

		// Parse filter chain match (SNI, ALPN, etc.)
		var filterChainMatch *v1alpha1.ListenerMatch
		if filterChain.FilterChainMatch != nil {
			filterChainMatch = p.parseFilterChainMatch(filterChain.FilterChainMatch)
		}

		// Process network filters
		for _, filter := range filterChain.Filters {
			networkFilters = append(networkFilters, &v1alpha1.FilterInfo{
				Name: filter.Name,
				Type: "network",
			})

			// Parse HTTP Connection Manager - creates HTTP route rules
			if filter.Name == "envoy.filters.network.http_connection_manager" {
				httpRules, httpFilterInfos := p.parseHttpConnectionManagerRules(filter)
				httpFilters = append(httpFilters, httpFilterInfos...)

				// If we have filter chain matching criteria, create a filter chain match rule first
				// This shows the L4 matching (SNI, ALPN, etc.) that happens before HTTP routing
				if filterChainMatch != nil {
					// Find the first HTTP route destination to pair with the filter chain match
					var firstHttpDest *v1alpha1.ListenerDestination
					if len(httpRules) > 0 && httpRules[0].Destination != nil {
						firstHttpDest = httpRules[0].Destination
					}

					// Create filter chain match rule first
					fcRule := &v1alpha1.ListenerRule{
						Match:       filterChainMatch,
						Destination: firstHttpDest,
					}
					allRules = append(allRules, fcRule)
				}

				// Then add HTTP route rules
				allRules = append(allRules, httpRules...)
			}

			// Parse TCP Proxy - pairs filter chain match with TCP destination
			if filter.Name == "envoy.filters.network.tcp_proxy" {
				tcpDest := p.parseTcpProxy(filter)
				if tcpDest != nil {
					// Create TCP rule: pair filter chain match with TCP destination
					rule := &v1alpha1.ListenerRule{
						Match:       filterChainMatch, // May be nil for implicit matches
						Destination: tcpDest,
					}
					allRules = append(allRules, rule)
				}
			}
		}
	}

	filterChainSummary := &v1alpha1.FilterChainSummary{
		TotalChains:    uint32(len(listener.FilterChains)), // #nosec G115 - listener filter chains count is bounded in practice
		HttpFilters:    httpFilters,
		NetworkFilters: networkFilters,
		TlsContext:     hasTLS,
	}

	return allRules, filterChainSummary
}

// parseFilterChainMatch parses filter chain matching criteria (SNI, ALPN, etc.)
func (p *Parser) parseFilterChainMatch(match *listenerv3.FilterChainMatch) *v1alpha1.ListenerMatch {
	if match == nil {
		return nil
	}

	filterChainMatch := &v1alpha1.FilterChainMatch{
		TransportProtocol: match.TransportProtocol,
	}

	// Parse SNI matching
	if len(match.ServerNames) > 0 {
		filterChainMatch.ServerNames = match.ServerNames
	}

	// Parse ALPN matching
	if len(match.ApplicationProtocols) > 0 {
		filterChainMatch.ApplicationProtocols = match.ApplicationProtocols
	}

	return &v1alpha1.ListenerMatch{
		MatchType: &v1alpha1.ListenerMatch_FilterChain{
			FilterChain: filterChainMatch,
		},
	}
}

// parseHttpConnectionManagerRules parses HTTP connection manager and returns paired rules
func (p *Parser) parseHttpConnectionManagerRules(filter *listenerv3.Filter) ([]*v1alpha1.ListenerRule, []*v1alpha1.FilterInfo) {
	var httpRules []*v1alpha1.ListenerRule
	var httpFilters []*v1alpha1.FilterInfo

	if filter.ConfigType == nil {
		return httpRules, httpFilters
	}

	var hcmConfig hcm.HttpConnectionManager
	if typedConfig := filter.GetTypedConfig(); typedConfig != nil {
		if err := typedConfig.UnmarshalTo(&hcmConfig); err != nil {
			return httpRules, httpFilters
		}
	}

	// Parse HTTP filters
	for _, httpFilter := range hcmConfig.HttpFilters {
		httpFilters = append(httpFilters, &v1alpha1.FilterInfo{
			Name: httpFilter.Name,
			Type: "http",
		})
	}

	// Parse route configuration
	var routeConfig *route.RouteConfiguration
	switch hcmConfig.GetRouteSpecifier().(type) {
	case *hcm.HttpConnectionManager_RouteConfig:
		routeConfig = hcmConfig.GetRouteConfig()
	case *hcm.HttpConnectionManager_Rds:
		// For RDS, we can't parse routes without additional context
		return httpRules, httpFilters
	}

	if routeConfig != nil {
		// Parse virtual hosts and routes, creating paired rules
		for _, vhost := range routeConfig.VirtualHosts {
			for _, route := range vhost.Routes {
				// Parse route match
				var match *v1alpha1.ListenerMatch
				if route.Match != nil {
					match = p.parseRouteMatch(route.Match)
				}

				// Parse route action (destination)
				var dest *v1alpha1.ListenerDestination
				if route.GetRoute() != nil {
					dest = p.parseRouteAction(route.GetRoute())
				}

				// Create rule if we have either match or destination
				if match != nil || dest != nil {
					rule := &v1alpha1.ListenerRule{
						Match:       match,
						Destination: dest,
					}
					httpRules = append(httpRules, rule)
				}
			}
		}
	}

	return httpRules, httpFilters
}

// parseRouteMatch parses HTTP route matching criteria
func (p *Parser) parseRouteMatch(match *route.RouteMatch) *v1alpha1.ListenerMatch {
	if match == nil {
		return nil
	}

	httpRouteMatch := &v1alpha1.HttpRouteMatch{}

	// Parse path matching
	var pathMatch *v1alpha1.PathMatchInfo
	switch match.GetPathSpecifier().(type) {
	case *route.RouteMatch_Path:
		pathMatch = &v1alpha1.PathMatchInfo{
			MatchType:     "exact",
			Path:          match.GetPath(),
			CaseSensitive: match.CaseSensitive.GetValue(),
		}
	case *route.RouteMatch_Prefix:
		pathMatch = &v1alpha1.PathMatchInfo{
			MatchType:     "prefix",
			Path:          match.GetPrefix(),
			CaseSensitive: match.CaseSensitive.GetValue(),
		}
	case *route.RouteMatch_SafeRegex:
		pathMatch = &v1alpha1.PathMatchInfo{
			MatchType:     "regex",
			Path:          match.GetSafeRegex().Regex,
			CaseSensitive: match.CaseSensitive.GetValue(),
		}
	}
	httpRouteMatch.PathMatch = pathMatch

	// Parse header matching
	for _, headerMatch := range match.Headers {
		headerMatchInfo := p.parseHeaderMatch(headerMatch)
		if headerMatchInfo != nil {
			httpRouteMatch.HeaderMatches = append(httpRouteMatch.HeaderMatches, headerMatchInfo)
		}
	}

	return &v1alpha1.ListenerMatch{
		MatchType: &v1alpha1.ListenerMatch_HttpRoute{
			HttpRoute: httpRouteMatch,
		},
	}
}

// parseHeaderMatch parses HTTP header matching criteria
func (p *Parser) parseHeaderMatch(headerMatch *route.HeaderMatcher) *v1alpha1.HeaderMatchInfo {
	if headerMatch == nil {
		return nil
	}

	headerMatchInfo := &v1alpha1.HeaderMatchInfo{
		Name:        headerMatch.Name,
		InvertMatch: headerMatch.InvertMatch,
	}

	switch headerMatch.GetHeaderMatchSpecifier().(type) {
	case *route.HeaderMatcher_StringMatch:
		stringMatch := headerMatch.GetStringMatch()
		switch stringMatch.GetMatchPattern().(type) {
		case *matcher.StringMatcher_Exact:
			headerMatchInfo.MatchType = "exact"
			headerMatchInfo.Value = stringMatch.GetExact()
		case *matcher.StringMatcher_Prefix:
			headerMatchInfo.MatchType = "prefix"
			headerMatchInfo.Value = stringMatch.GetPrefix()
		case *matcher.StringMatcher_Suffix:
			headerMatchInfo.MatchType = "suffix"
			headerMatchInfo.Value = stringMatch.GetSuffix()
		case *matcher.StringMatcher_SafeRegex:
			headerMatchInfo.MatchType = "regex"
			headerMatchInfo.Value = stringMatch.GetSafeRegex().Regex
		}
	case *route.HeaderMatcher_PresentMatch:
		headerMatchInfo.MatchType = "present"
		headerMatchInfo.Value = ""
	}

	return headerMatchInfo
}

// parseRouteAction parses HTTP route action for destination information
func (p *Parser) parseRouteAction(routeAction *route.RouteAction) *v1alpha1.ListenerDestination {
	if routeAction == nil {
		return nil
	}

	switch routeAction.GetClusterSpecifier().(type) {
	case *route.RouteAction_Cluster:
		return &v1alpha1.ListenerDestination{
			DestinationType: "cluster",
			ClusterName:     routeAction.GetCluster(),
		}
	case *route.RouteAction_WeightedClusters:
		// For weighted clusters, return the first one (or we could return all)
		weightedClusters := routeAction.GetWeightedClusters()
		if len(weightedClusters.Clusters) > 0 {
			first := weightedClusters.Clusters[0]
			return &v1alpha1.ListenerDestination{
				DestinationType: "weighted_cluster",
				ClusterName:     first.Name,
				Weight:          first.Weight.GetValue(),
			}
		}
	case *route.RouteAction_ClusterHeader:
		return &v1alpha1.ListenerDestination{
			DestinationType: "cluster_header",
		}
	}

	return nil
}

// parseTcpProxy parses TCP proxy filter for destination information
func (p *Parser) parseTcpProxy(filter *listenerv3.Filter) *v1alpha1.ListenerDestination {
	if filter.ConfigType == nil {
		return nil
	}

	var tcpProxyConfig tcp_proxy.TcpProxy
	if typedConfig := filter.GetTypedConfig(); typedConfig != nil {
		if err := typedConfig.UnmarshalTo(&tcpProxyConfig); err != nil {
			return nil
		}
	}

	switch tcpProxyConfig.GetClusterSpecifier().(type) {
	case *tcp_proxy.TcpProxy_Cluster:
		return &v1alpha1.ListenerDestination{
			DestinationType: "tcp_cluster",
			ClusterName:     tcpProxyConfig.GetCluster(),
		}
	case *tcp_proxy.TcpProxy_WeightedClusters:
		// For weighted clusters, return the first one
		weightedClusters := tcpProxyConfig.GetWeightedClusters()
		if len(weightedClusters.Clusters) > 0 {
			first := weightedClusters.Clusters[0]
			return &v1alpha1.ListenerDestination{
				DestinationType: "tcp_weighted_cluster",
				ClusterName:     first.Name,
				Weight:          first.Weight,
			}
		}
	}

	return nil
}
