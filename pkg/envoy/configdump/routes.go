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
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// determineRouteType categorizes a route name for generic Envoy deployments
// isFromStaticConfig indicates if this route came from StaticRouteConfigs
func determineRouteType(routeName string, isFromStaticConfig bool) v1alpha1.RouteType {
	// If route came from StaticRouteConfigs, it's always STATIC
	if isFromStaticConfig {
		return v1alpha1.RouteType_STATIC
	}

	// For generic Envoy deployments, use basic classification
	// Service mesh specific classification should be done by enrichment layers
	if routeName == "" {
		return v1alpha1.RouteType_STATIC
	}

	// Default to SERVICE_SPECIFIC for non-empty, non-static routes
	return v1alpha1.RouteType_SERVICE_SPECIFIC
}

// parseRoutesFromAny extracts route configurations from protobuf Any
func (p *Parser) parseRoutesFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	routeDump := &admin.RoutesConfigDump{}
	if err := configAny.UnmarshalTo(routeDump); err != nil {
		return fmt.Errorf("failed to unmarshal routes config dump: %w", err)
	}

	// Track which routes are from static configs
	staticRouteNames := make(map[string]bool)

	// Extract dynamic routes
	for i, r := range routeDump.DynamicRouteConfigs {
		if r.RouteConfig != nil {
			var route routev3.RouteConfiguration
			if err := r.RouteConfig.UnmarshalTo(&route); err == nil {
				parsed.Routes = append(parsed.Routes, &route)

				// Map route config to raw key for empty names
				key := route.Name
				if route.Name == "" {
					key = fmt.Sprintf("__empty_dynamic_%d", i)
				}
				parsed.RouteConfigToRawKey[&route] = key
			}
		}
	}

	// Extract static routes and mark them
	for i, r := range routeDump.StaticRouteConfigs {
		if r.RouteConfig != nil {
			var route routev3.RouteConfiguration
			if err := r.RouteConfig.UnmarshalTo(&route); err == nil {
				staticRouteNames[route.Name] = true
				parsed.Routes = append(parsed.Routes, &route)

				// Map route config to raw key for empty names
				key := route.Name
				if route.Name == "" {
					key = fmt.Sprintf("__empty_static_%d", i)
				}
				parsed.RouteConfigToRawKey[&route] = key
			}
		}
	}

	// Store static route names for later use in summarizeRouteConfig
	if parsed.StaticRouteNames == nil {
		parsed.StaticRouteNames = make(map[string]bool)
	}
	for name := range staticRouteNames {
		parsed.StaticRouteNames[name] = true
	}

	return nil
}

// summarizeRouteConfig converts a RouteConfiguration to a RouteConfigSummary
func (p *Parser) summarizeRouteConfig(routeConfig *routev3.RouteConfiguration, parsed *ParsedConfig) *v1alpha1.RouteConfigSummary {
	if routeConfig == nil {
		return nil
	}

	// Check if this route came from StaticRouteConfigs
	isFromStaticConfig := false
	if parsed != nil && parsed.StaticRouteNames != nil {
		isFromStaticConfig = parsed.StaticRouteNames[routeConfig.Name]
	}

	summary := &v1alpha1.RouteConfigSummary{
		Name:                routeConfig.Name,
		InternalOnlyHeaders: routeConfig.InternalOnlyHeaders,
		ValidateClusters:    routeConfig.ValidateClusters.GetValue(),
		Type:                determineRouteType(routeConfig.Name, isFromStaticConfig),
	}

	// Add raw config if available
	if parsed != nil && parsed.RawRoutes != nil {
		// Use the mapped key for this specific route config
		var key string
		if parsed.RouteConfigToRawKey != nil {
			if mappedKey, exists := parsed.RouteConfigToRawKey[routeConfig]; exists {
				key = mappedKey
			} else {
				key = routeConfig.Name // fallback to route name
			}
		} else {
			key = routeConfig.Name // fallback if mapping not available
		}

		if rawConfig, exists := parsed.RawRoutes[key]; exists {
			summary.RawConfig = rawConfig
		}
	}

	// Header modifications not included in simplified schema

	// Extract virtual hosts (basic version - can be expanded)
	for _, vh := range routeConfig.VirtualHosts {
		vhInfo := &v1alpha1.VirtualHostInfo{
			Name:    vh.Name,
			Domains: vh.Domains,
		}

		// Extract basic route information (simplified for now)
		for _, route := range vh.Routes {
			routeInfo := &v1alpha1.RouteInfo{
				Name: route.Name,
			}

			// Extract match information (basic)
			if match := route.Match; match != nil {
				routeInfo.Match = &v1alpha1.RouteMatchInfo{
					CaseSensitive: match.CaseSensitive.GetValue(),
				}

				// Extract path specifier
				switch ps := match.PathSpecifier.(type) {
				case *routev3.RouteMatch_Prefix:
					routeInfo.Match.PathSpecifier = "prefix"
					routeInfo.Match.Path = ps.Prefix
				case *routev3.RouteMatch_Path:
					routeInfo.Match.PathSpecifier = "path"
					routeInfo.Match.Path = ps.Path
				case *routev3.RouteMatch_SafeRegex:
					routeInfo.Match.PathSpecifier = "safe_regex"
					routeInfo.Match.Path = ps.SafeRegex.Regex
				}
			}

			// Extract action information (basic)
			switch action := route.Action.(type) {
			case *routev3.Route_Route:
				routeInfo.Action = &v1alpha1.RouteActionInfo{
					ActionType: "route",
				}

				// Extract cluster specifier
				switch cs := action.Route.ClusterSpecifier.(type) {
				case *routev3.RouteAction_Cluster:
					routeInfo.Action.Cluster = cs.Cluster
				case *routev3.RouteAction_WeightedClusters:
					for _, wc := range cs.WeightedClusters.Clusters {
						routeInfo.Action.WeightedClusters = append(routeInfo.Action.WeightedClusters, &v1alpha1.WeightedClusterInfo{
							Name:   wc.Name,
							Weight: wc.Weight.GetValue(),
						})
					}
				}

			case *routev3.Route_Redirect:
				routeInfo.Action = &v1alpha1.RouteActionInfo{
					ActionType: "redirect",
				}
			case *routev3.Route_DirectResponse:
				routeInfo.Action = &v1alpha1.RouteActionInfo{
					ActionType: "direct_response",
				}
			}

			vhInfo.Routes = append(vhInfo.Routes, routeInfo)
		}

		summary.VirtualHosts = append(summary.VirtualHosts, vhInfo)
	}

	return summary
}
