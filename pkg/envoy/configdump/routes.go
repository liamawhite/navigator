package configdump

import (
	"fmt"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// parseRoutesFromAny extracts route configurations from protobuf Any
func (p *Parser) parseRoutesFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	routeDump := &admin.RoutesConfigDump{}
	if err := configAny.UnmarshalTo(routeDump); err != nil {
		return fmt.Errorf("failed to unmarshal routes config dump: %w", err)
	}

	// Extract dynamic routes (like istioctl)
	for _, r := range routeDump.DynamicRouteConfigs {
		if r.RouteConfig != nil {
			var route routev3.RouteConfiguration
			if err := r.RouteConfig.UnmarshalTo(&route); err == nil {
				parsed.Routes = append(parsed.Routes, &route)
			}
		}
	}

	// Extract static routes
	for _, r := range routeDump.StaticRouteConfigs {
		if r.RouteConfig != nil {
			var route routev3.RouteConfiguration
			if err := r.RouteConfig.UnmarshalTo(&route); err == nil {
				parsed.Routes = append(parsed.Routes, &route)
			}
		}
	}

	return nil
}

// summarizeRouteConfig converts a RouteConfiguration to a RouteConfigSummary
func (p *Parser) summarizeRouteConfig(routeConfig *routev3.RouteConfiguration) *v1alpha1.RouteConfigSummary {
	if routeConfig == nil {
		return nil
	}

	summary := &v1alpha1.RouteConfigSummary{
		Name:                routeConfig.Name,
		InternalOnlyHeaders: routeConfig.InternalOnlyHeaders,
		ValidateClusters:    routeConfig.ValidateClusters.GetValue(),
	}

	// Extract header modifications (simplified)
	for _, header := range routeConfig.ResponseHeadersToAdd {
		summary.ResponseHeadersToAdd = append(summary.ResponseHeadersToAdd, &v1alpha1.HeaderValueOption{
			Header: &v1alpha1.HeaderInfo{
				Key:   header.Header.Key,
				Value: header.Header.Value,
			},
			Append: header.Append.GetValue(),
		})
	}
	summary.ResponseHeadersToRemove = routeConfig.ResponseHeadersToRemove

	for _, header := range routeConfig.RequestHeadersToAdd {
		summary.RequestHeadersToAdd = append(summary.RequestHeadersToAdd, &v1alpha1.HeaderValueOption{
			Header: &v1alpha1.HeaderInfo{
				Key:   header.Header.Key,
				Value: header.Header.Value,
			},
			Append: header.Append.GetValue(),
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
