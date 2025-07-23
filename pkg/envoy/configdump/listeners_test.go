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
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcp_proxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_parseFilterChainMatch(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		filterChain  *listenerv3.FilterChainMatch
		expectedType string
		expectedSNI  []string
		expectedALPN []string
		expectedTLS  string
	}{
		{
			name: "SNI and ALPN matching",
			filterChain: &listenerv3.FilterChainMatch{
				ServerNames:          []string{"example.com", "*.example.com"},
				ApplicationProtocols: []string{"h2", "http/1.1"},
				TransportProtocol:    "tls",
			},
			expectedSNI:  []string{"example.com", "*.example.com"},
			expectedALPN: []string{"h2", "http/1.1"},
			expectedTLS:  "tls",
		},
		{
			name: "SNI only matching",
			filterChain: &listenerv3.FilterChainMatch{
				ServerNames: []string{"api.service.com"},
			},
			expectedSNI: []string{"api.service.com"},
		},
		{
			name: "ALPN only matching",
			filterChain: &listenerv3.FilterChainMatch{
				ApplicationProtocols: []string{"grpc"},
			},
			expectedALPN: []string{"grpc"},
		},
		{
			name: "Transport protocol only",
			filterChain: &listenerv3.FilterChainMatch{
				TransportProtocol: "raw_buffer",
			},
			expectedTLS: "raw_buffer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseFilterChainMatch(tt.filterChain)

			if tt.filterChain == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			filterChain := result.GetFilterChain()
			require.NotNil(t, filterChain)

			assert.Equal(t, tt.expectedSNI, filterChain.ServerNames)
			assert.Equal(t, tt.expectedALPN, filterChain.ApplicationProtocols)
			assert.Equal(t, tt.expectedTLS, filterChain.TransportProtocol)
		})
	}

	t.Run("handles nil input", func(t *testing.T) {
		result := parser.parseFilterChainMatch(nil)
		assert.Nil(t, result)
	})
}

func TestParser_parseRouteMatch(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name            string
		routeMatch      *route.RouteMatch
		expectedPath    *v1alpha1.PathMatchInfo
		expectedHeaders int
	}{
		{
			name: "exact path match",
			routeMatch: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Path{
					Path: "/api/v1/users",
				},
				CaseSensitive: wrapperspb.Bool(true),
			},
			expectedPath: &v1alpha1.PathMatchInfo{
				MatchType:     "exact",
				Path:          "/api/v1/users",
				CaseSensitive: true,
			},
		},
		{
			name: "prefix path match",
			routeMatch: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/api/",
				},
				CaseSensitive: wrapperspb.Bool(false),
			},
			expectedPath: &v1alpha1.PathMatchInfo{
				MatchType:     "prefix",
				Path:          "/api/",
				CaseSensitive: false,
			},
		},
		{
			name: "regex path match",
			routeMatch: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_SafeRegex{
					SafeRegex: &matcherv3.RegexMatcher{
						Regex: "/api/v[0-9]+/.*",
					},
				},
			},
			expectedPath: &v1alpha1.PathMatchInfo{
				MatchType:     "regex",
				Path:          "/api/v[0-9]+/.*",
				CaseSensitive: false,
			},
		},
		{
			name: "with header matches",
			routeMatch: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/",
				},
				Headers: []*route.HeaderMatcher{
					{
						Name: ":authority",
						HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
							ExactMatch: "example.com",
						},
					},
					{
						Name: "x-version",
						HeaderMatchSpecifier: &route.HeaderMatcher_PrefixMatch{
							PrefixMatch: "v1",
						},
					},
				},
			},
			expectedPath: &v1alpha1.PathMatchInfo{
				MatchType:     "prefix",
				Path:          "/",
				CaseSensitive: false,
			},
			expectedHeaders: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseRouteMatch(tt.routeMatch)

			require.NotNil(t, result)
			httpRoute := result.GetHttpRoute()
			require.NotNil(t, httpRoute)

			if tt.expectedPath != nil {
				assert.Equal(t, tt.expectedPath, httpRoute.PathMatch)
			}

			if tt.expectedHeaders > 0 {
				assert.Len(t, httpRoute.HeaderMatches, tt.expectedHeaders)
			}
		})
	}

	t.Run("handles nil input", func(t *testing.T) {
		result := parser.parseRouteMatch(nil)
		assert.Nil(t, result)
	})
}

func TestParser_parseHeaderMatch(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name           string
		headerMatcher  *route.HeaderMatcher
		expectedName   string
		expectedType   string
		expectedValue  string
		expectedInvert bool
	}{
		{
			name: "exact header match",
			headerMatcher: &route.HeaderMatcher{
				Name: ":authority",
				HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
					ExactMatch: "example.com",
				},
			},
			expectedName:  ":authority",
			expectedType:  "exact",
			expectedValue: "example.com",
		},
		{
			name: "prefix header match",
			headerMatcher: &route.HeaderMatcher{
				Name: "user-agent",
				HeaderMatchSpecifier: &route.HeaderMatcher_PrefixMatch{
					PrefixMatch: "Mozilla",
				},
			},
			expectedName:  "user-agent",
			expectedType:  "prefix",
			expectedValue: "Mozilla",
		},
		{
			name: "suffix header match",
			headerMatcher: &route.HeaderMatcher{
				Name: "accept",
				HeaderMatchSpecifier: &route.HeaderMatcher_SuffixMatch{
					SuffixMatch: "json",
				},
			},
			expectedName:  "accept",
			expectedType:  "suffix",
			expectedValue: "json",
		},
		{
			name: "regex header match",
			headerMatcher: &route.HeaderMatcher{
				Name: "x-custom",
				HeaderMatchSpecifier: &route.HeaderMatcher_SafeRegexMatch{
					SafeRegexMatch: &matcherv3.RegexMatcher{
						Regex: "^v[0-9]+$",
					},
				},
			},
			expectedName:  "x-custom",
			expectedType:  "regex",
			expectedValue: "^v[0-9]+$",
		},
		{
			name: "present header match",
			headerMatcher: &route.HeaderMatcher{
				Name: "authorization",
				HeaderMatchSpecifier: &route.HeaderMatcher_PresentMatch{
					PresentMatch: true,
				},
			},
			expectedName:  "authorization",
			expectedType:  "present",
			expectedValue: "",
		},
		{
			name: "inverted header match",
			headerMatcher: &route.HeaderMatcher{
				Name: "x-debug",
				HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
					ExactMatch: "true",
				},
				InvertMatch: true,
			},
			expectedName:   "x-debug",
			expectedType:   "exact",
			expectedValue:  "true",
			expectedInvert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseHeaderMatch(tt.headerMatcher)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedType, result.MatchType)
			assert.Equal(t, tt.expectedValue, result.Value)
			assert.Equal(t, tt.expectedInvert, result.InvertMatch)
		})
	}

	t.Run("handles nil input", func(t *testing.T) {
		result := parser.parseHeaderMatch(nil)
		assert.Nil(t, result)
	})
}

func TestParser_parseRouteAction(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name            string
		routeAction     *route.RouteAction
		expectedType    string
		expectedCluster string
		expectedWeight  uint32
	}{
		{
			name: "single cluster destination",
			routeAction: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: "backend-service",
				},
			},
			expectedType:    "cluster",
			expectedCluster: "backend-service",
		},
		{
			name: "weighted cluster destination",
			routeAction: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_WeightedClusters{
					WeightedClusters: &route.WeightedCluster{
						Clusters: []*route.WeightedCluster_ClusterWeight{
							{
								Name:   "service-v1",
								Weight: wrapperspb.UInt32(80),
							},
							{
								Name:   "service-v2",
								Weight: wrapperspb.UInt32(20),
							},
						},
					},
				},
			},
			expectedType:    "weighted_cluster",
			expectedCluster: "service-v1",
			expectedWeight:  80,
		},
		{
			name: "cluster header destination",
			routeAction: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_ClusterHeader{
					ClusterHeader: "x-target-cluster",
				},
			},
			expectedType: "cluster_header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseRouteAction(tt.routeAction)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType, result.DestinationType)
			assert.Equal(t, tt.expectedCluster, result.ClusterName)
			if tt.expectedWeight > 0 {
				assert.Equal(t, tt.expectedWeight, result.Weight)
			}
		})
	}

	t.Run("handles nil input", func(t *testing.T) {
		result := parser.parseRouteAction(nil)
		assert.Nil(t, result)
	})
}

func TestParser_parseTcpProxy(t *testing.T) {
	parser := NewParser()

	// Create TCP proxy config
	tcpProxyConfig := &tcp_proxy.TcpProxy{
		ClusterSpecifier: &tcp_proxy.TcpProxy_Cluster{
			Cluster: "tcp-backend",
		},
	}
	tcpProxyAny, err := anypb.New(tcpProxyConfig)
	require.NoError(t, err)

	// Create weighted TCP proxy config
	tcpProxyWeightedConfig := &tcp_proxy.TcpProxy{
		ClusterSpecifier: &tcp_proxy.TcpProxy_WeightedClusters{
			WeightedClusters: &tcp_proxy.TcpProxy_WeightedCluster{
				Clusters: []*tcp_proxy.TcpProxy_WeightedCluster_ClusterWeight{
					{
						Name:   "tcp-service-v1",
						Weight: 70,
					},
					{
						Name:   "tcp-service-v2",
						Weight: 30,
					},
				},
			},
		},
	}
	tcpProxyWeightedAny, err := anypb.New(tcpProxyWeightedConfig)
	require.NoError(t, err)

	tests := []struct {
		name            string
		filter          *listenerv3.Filter
		expectedType    string
		expectedCluster string
		expectedWeight  uint32
	}{
		{
			name: "single TCP cluster",
			filter: &listenerv3.Filter{
				Name: "envoy.filters.network.tcp_proxy",
				ConfigType: &listenerv3.Filter_TypedConfig{
					TypedConfig: tcpProxyAny,
				},
			},
			expectedType:    "tcp_cluster",
			expectedCluster: "tcp-backend",
		},
		{
			name: "weighted TCP clusters",
			filter: &listenerv3.Filter{
				Name: "envoy.filters.network.tcp_proxy",
				ConfigType: &listenerv3.Filter_TypedConfig{
					TypedConfig: tcpProxyWeightedAny,
				},
			},
			expectedType:    "tcp_weighted_cluster",
			expectedCluster: "tcp-service-v1",
			expectedWeight:  70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseTcpProxy(tt.filter)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType, result.DestinationType)
			assert.Equal(t, tt.expectedCluster, result.ClusterName)
			if tt.expectedWeight > 0 {
				assert.Equal(t, tt.expectedWeight, result.Weight)
			}
		})
	}

	t.Run("handles filter without config", func(t *testing.T) {
		filter := &listenerv3.Filter{
			Name: "envoy.filters.network.tcp_proxy",
		}
		result := parser.parseTcpProxy(filter)
		assert.Nil(t, result)
	})
}

func TestParser_parseListenerFilters(t *testing.T) {
	parser := NewParser()

	// Create HTTP connection manager config
	routeConfig := &route.RouteConfiguration{
		Name: "test-route",
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "test-vhost",
				Domains: []string{"*"},
				Routes: []*route.Route{
					{
						Match: &route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{
								Prefix: "/api",
							},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: "api-service",
								},
							},
						},
					},
				},
			},
		},
	}

	hcmConfig := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		},
		HttpFilters: []*hcm.HttpFilter{
			{Name: "envoy.filters.http.router"},
		},
	}
	hcmAny, err := anypb.New(hcmConfig)
	require.NoError(t, err)

	// Create listener with filter chains
	listener := &listenerv3.Listener{
		Name: "test-listener",
		FilterChains: []*listenerv3.FilterChain{
			{
				FilterChainMatch: &listenerv3.FilterChainMatch{
					ServerNames:          []string{"example.com"},
					ApplicationProtocols: []string{"h2", "http/1.1"},
					TransportProtocol:    "tls",
				},
				Filters: []*listenerv3.Filter{
					{
						Name: "envoy.filters.network.http_connection_manager",
						ConfigType: &listenerv3.Filter_TypedConfig{
							TypedConfig: hcmAny,
						},
					},
				},
			},
		},
	}

	rules, filterChains := parser.parseListenerFilters(listener)

	// Should have rules from filter chain match and HTTP route match
	require.Len(t, rules, 2)

	// First rule should be filter chain match with TCP destination
	firstRule := rules[0]
	require.NotNil(t, firstRule.Match)
	filterChainMatch := firstRule.Match.GetFilterChain()
	require.NotNil(t, filterChainMatch)
	assert.Equal(t, []string{"example.com"}, filterChainMatch.ServerNames)
	assert.Equal(t, []string{"h2", "http/1.1"}, filterChainMatch.ApplicationProtocols)
	assert.Equal(t, "tls", filterChainMatch.TransportProtocol)
	require.NotNil(t, firstRule.Destination)
	assert.Equal(t, "cluster", firstRule.Destination.DestinationType)
	assert.Equal(t, "api-service", firstRule.Destination.ClusterName)

	// Second rule should be HTTP route match with route destination
	secondRule := rules[1]
	require.NotNil(t, secondRule.Match)
	httpRouteMatch := secondRule.Match.GetHttpRoute()
	require.NotNil(t, httpRouteMatch)
	require.NotNil(t, httpRouteMatch.PathMatch)
	assert.Equal(t, "prefix", httpRouteMatch.PathMatch.MatchType)
	assert.Equal(t, "/api", httpRouteMatch.PathMatch.Path)
	require.NotNil(t, secondRule.Destination)
	assert.Equal(t, "cluster", secondRule.Destination.DestinationType)
	assert.Equal(t, "api-service", secondRule.Destination.ClusterName)

	// Should have filter chain summary
	require.NotNil(t, filterChains)
	assert.Equal(t, uint32(1), filterChains.TotalChains)
	assert.Len(t, filterChains.NetworkFilters, 1)
	assert.Len(t, filterChains.HttpFilters, 1)
	assert.False(t, filterChains.TlsContext) // No transport socket in this test
}

func TestParser_parseListenerFilters_EmptyListener(t *testing.T) {
	parser := NewParser()

	t.Run("handles nil listener", func(t *testing.T) {
		rules, filterChains := parser.parseListenerFilters(nil)
		assert.Nil(t, rules)
		assert.Equal(t, &v1alpha1.FilterChainSummary{}, filterChains)
	})

	t.Run("handles listener without filter chains", func(t *testing.T) {
		listener := &listenerv3.Listener{
			Name: "empty-listener",
		}
		rules, filterChains := parser.parseListenerFilters(listener)
		assert.Nil(t, rules)
		assert.Equal(t, &v1alpha1.FilterChainSummary{}, filterChains)
	})
}
