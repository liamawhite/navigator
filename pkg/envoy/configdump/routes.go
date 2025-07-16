package configdump

import (
	"fmt"
	"regexp"
	"strings"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// determineRouteType categorizes a route name into one of three types
// isFromStaticConfig indicates if this route came from StaticRouteConfigs
func determineRouteType(routeName string, isFromStaticConfig bool) v1alpha1.RouteType {
	// If route came from StaticRouteConfigs, it's always STATIC
	if isFromStaticConfig {
		return v1alpha1.RouteType_STATIC
	}

	// PORT_BASED: Routes with just port numbers (e.g., "80", "443", "15010")
	if isPortOnly(routeName) {
		return v1alpha1.RouteType_PORT_BASED
	}

	// STATIC: Istio/Envoy internal routing patterns (for dynamic routes that are actually static)
	if isStaticRoute(routeName) {
		return v1alpha1.RouteType_STATIC
	}

	// SERVICE_SPECIFIC: Routes with service hostnames and ports (default for anything else)
	return v1alpha1.RouteType_SERVICE_SPECIFIC
}

// isPortOnly checks if the route name is just a port number
func isPortOnly(routeName string) bool {
	// Match exactly a number between 1 and 65535
	portRegex := regexp.MustCompile(`^[1-9]\d{0,4}$`)
	if !portRegex.MatchString(routeName) {
		return false
	}

	// Additional validation: check if it's a valid port number (1-65535)
	if len(routeName) > 5 {
		return false
	}

	// Check if the number is actually <= 65535
	if len(routeName) == 5 && routeName > "65535" {
		return false
	}

	return true
}

// isStaticRoute checks if the route name matches Istio/Envoy internal patterns
func isStaticRoute(routeName string) bool {
	// Empty or whitespace-only names are considered static
	if routeName == "" || strings.TrimSpace(routeName) == "" {
		return true
	}

	// Static route patterns commonly seen in Istio/Envoy
	staticPatterns := []string{
		"InboundPassthroughCluster",
		"BlackHoleCluster",
		"PassthroughCluster",
	}

	// Check exact matches for known static patterns
	for _, pattern := range staticPatterns {
		if routeName == pattern {
			return true
		}
	}

	// Check for inbound route patterns like "inbound|8080||"
	inboundRegex := regexp.MustCompile(`^inbound\|\d+\|\|.*$`)
	if inboundRegex.MatchString(routeName) {
		return true
	}

	// Check for outbound route patterns like "outbound|8080||"
	outboundRegex := regexp.MustCompile(`^outbound\|\d+\|\|.*$`)
	return outboundRegex.MatchString(routeName)
}

// parseRoutesFromAny extracts route configurations from protobuf Any
func (p *Parser) parseRoutesFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	routeDump := &admin.RoutesConfigDump{}
	if err := configAny.UnmarshalTo(routeDump); err != nil {
		return fmt.Errorf("failed to unmarshal routes config dump: %w", err)
	}

	// Track which routes are from static configs
	staticRouteNames := make(map[string]bool)

	// Extract dynamic routes (like istioctl)
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

	// Extract header modifications (simplified)
	for _, header := range routeConfig.ResponseHeadersToAdd {
		summary.ResponseHeadersToAdd = append(summary.ResponseHeadersToAdd, &v1alpha1.HeaderValueOption{
			Header: &v1alpha1.HeaderInfo{
				Key:   header.Header.Key,
				Value: header.Header.Value,
			},
			Append: header.GetAppendAction() == corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
		})
	}
	summary.ResponseHeadersToRemove = routeConfig.ResponseHeadersToRemove

	for _, header := range routeConfig.RequestHeadersToAdd {
		summary.RequestHeadersToAdd = append(summary.RequestHeadersToAdd, &v1alpha1.HeaderValueOption{
			Header: &v1alpha1.HeaderInfo{
				Key:   header.Header.Key,
				Value: header.Header.Value,
			},
			Append: header.GetAppendAction() == corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
		})
	}
	summary.RequestHeadersToRemove = routeConfig.RequestHeadersToRemove

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
