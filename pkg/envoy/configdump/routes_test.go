package configdump

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

func TestDetermineRouteType(t *testing.T) {
	tests := []struct {
		name               string
		routeName          string
		isFromStaticConfig bool
		expectedType       v1alpha1.RouteType
	}{
		// Port-based routes (dynamic)
		{
			name:               "port 80",
			routeName:          "80",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_PORT_BASED,
		},
		{
			name:               "port 443",
			routeName:          "443",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_PORT_BASED,
		},
		{
			name:               "port 15010",
			routeName:          "15010",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_PORT_BASED,
		},
		{
			name:               "port 15014",
			routeName:          "15014",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_PORT_BASED,
		},

		// Service-specific routes (dynamic)
		{
			name:               "backend service",
			routeName:          "backend.demo.svc.cluster.local:8080",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		{
			name:               "database service",
			routeName:          "database.demo.svc.cluster.local:8080",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		{
			name:               "frontend service",
			routeName:          "frontend.demo.svc.cluster.local:8080",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		{
			name:               "kube-dns service",
			routeName:          "kube-dns.kube-system.svc.cluster.local:9153",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},
		{
			name:               "istio-ingressgateway service",
			routeName:          "istio-ingressgateway.istio-system.svc.cluster.local:15021",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC,
		},

		// Static routes (from static config array)
		{
			name:               "static route with any name",
			routeName:          "whatever-name",
			isFromStaticConfig: true,
			expectedType:       v1alpha1.RouteType_STATIC,
		},
		{
			name:               "static route empty name",
			routeName:          "",
			isFromStaticConfig: true,
			expectedType:       v1alpha1.RouteType_STATIC,
		},

		// Static routes (dynamic but matching patterns)
		{
			name:               "InboundPassthroughCluster",
			routeName:          "InboundPassthroughCluster",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_STATIC,
		},
		{
			name:               "inbound route pattern",
			routeName:          "inbound|8080||",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_STATIC,
		},

		// Edge cases (dynamic)
		{
			name:               "empty route name (dynamic)",
			routeName:          "",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_STATIC, // empty names are static
		},
		{
			name:               "whitespace route name (dynamic)",
			routeName:          "   ",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_STATIC, // whitespace names are static
		},
		{
			name:               "unknown pattern",
			routeName:          "some-unknown-pattern",
			isFromStaticConfig: false,
			expectedType:       v1alpha1.RouteType_SERVICE_SPECIFIC, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineRouteType(tt.routeName, tt.isFromStaticConfig)
			assert.Equal(t, tt.expectedType, result, "Route %q (static=%v) should be categorized as %v, got %v", tt.routeName, tt.isFromStaticConfig, tt.expectedType, result)
		})
	}
}

func TestIsPortOnly(t *testing.T) {
	tests := []struct {
		name      string
		routeName string
		expected  bool
	}{
		{"port 80", "80", true},
		{"port 443", "443", true},
		{"port 8080", "8080", true},
		{"port 15010", "15010", true},
		{"port 65535", "65535", true},
		{"port 1", "1", true},
		{"leading zero not allowed", "080", false},
		{"zero not allowed", "0", false},
		{"too large", "65536", false},
		{"with colon", "80:8080", false},
		{"service name", "backend.demo.svc.cluster.local:8080", false},
		{"empty", "", false},
		{"letters", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPortOnly(tt.routeName)
			assert.Equal(t, tt.expected, result, "isPortOnly(%q) should return %v", tt.routeName, tt.expected)
		})
	}
}

func TestIsStaticRoute(t *testing.T) {
	tests := []struct {
		name      string
		routeName string
		expected  bool
	}{
		{"InboundPassthroughCluster", "InboundPassthroughCluster", true},
		{"BlackHoleCluster", "BlackHoleCluster", true},
		{"PassthroughCluster", "PassthroughCluster", true},
		{"inbound pattern", "inbound|8080||", true},
		{"inbound pattern with suffix", "inbound|8080||some-suffix", true},
		{"outbound pattern", "outbound|8080||", true},
		{"outbound pattern with suffix", "outbound|8080||backend.demo.svc.cluster.local", true},
		{"port only", "80", false},
		{"service name", "backend.demo.svc.cluster.local:8080", false},
		{"unknown pattern", "some-random-name", false},
		{"empty", "", true},
		{"whitespace", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isStaticRoute(tt.routeName)
			assert.Equal(t, tt.expected, result, "isStaticRoute(%q) should return %v", tt.routeName, tt.expected)
		})
	}
}
