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
	"testing"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichRouteType(t *testing.T) {
	enrichFunc := enrichRouteType()

	tests := []struct {
		name        string
		routeName   string
		currentType v1alpha1.RouteType
		expected    v1alpha1.RouteType
		description string
	}{
		{
			name:        "preserves existing static type",
			routeName:   "some_route",
			currentType: v1alpha1.RouteType_STATIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Static routes should remain static regardless of name",
		},
		{
			name:        "preserves static with service name",
			routeName:   "backend.demo.svc.cluster.local:8080",
			currentType: v1alpha1.RouteType_STATIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Service-like name should not override static classification",
		},
		{
			name:        "service-specific route with pipe separator",
			routeName:   "backend|8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Pipe separators indicate service-specific routes in Istio",
		},
		{
			name:        "service route with complex pipe format",
			routeName:   "backend.demo.svc.cluster.local|8080|v1",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Complex pipe-separated service route with version",
		},
		{
			name:        "service route with namespace pipe",
			routeName:   "api.production|443",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Service route across namespaces",
		},
		{
			name:        "port-based route standard HTTP",
			routeName:   "80",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Standard HTTP port should remain port-based",
		},
		{
			name:        "port-based route standard HTTPS",
			routeName:   "443",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Standard HTTPS port should remain port-based",
		},
		{
			name:        "port-based route custom port",
			routeName:   "8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Custom application port should remain port-based",
		},
		{
			name:        "port-based route high port",
			routeName:   "65535",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Maximum port number should be port-based",
		},
		{
			name:        "port-based route Istio control plane",
			routeName:   "15010",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Istio discovery port should be port-based",
		},
		{
			name:        "static route - localhost with port",
			routeName:   "localhost:8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Localhost routes should be static",
		},
		{
			name:        "static route - localhost without port",
			routeName:   "localhost",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Localhost without port should be static",
		},
		{
			name:        "static route - 127.0.0.1 with port",
			routeName:   "127.0.0.1:8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "IPv4 loopback should be static",
		},
		{
			name:        "static route - 127.0.0.1 without port",
			routeName:   "127.0.0.1",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "IPv4 loopback without port should be static",
		},
		{
			name:        "static route - BlackHoleCluster",
			routeName:   "BlackHoleCluster",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Istio BlackHoleCluster should be static",
		},
		{
			name:        "static route - InboundPassthroughCluster",
			routeName:   "InboundPassthroughCluster",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Istio InboundPassthroughCluster should be static",
		},
		{
			name:        "static route - PassthroughCluster",
			routeName:   "PassthroughCluster",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Istio PassthroughCluster should be static",
		},
		{
			name:        "static route - local_agent",
			routeName:   "local_agent",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Local agent routes should be static",
		},
		{
			name:        "static route - admin",
			routeName:   "admin",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Admin routes should be static",
		},
		{
			name:        "static route - empty name",
			routeName:   "",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Empty route names should be static",
		},
		{
			name:        "static route - whitespace only",
			routeName:   "   \t\n  ",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Whitespace-only route names should be static",
		},
		{
			name:        "service-specific FQDN route",
			routeName:   "backend.demo.svc.cluster.local:8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "FQDN service routes should be service-specific",
		},
		{
			name:        "service-specific cross-namespace",
			routeName:   "api.production.svc.cluster.local:443",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Cross-namespace service should be service-specific",
		},
		{
			name:        "service-specific istio system",
			routeName:   "istiod.istio-system.svc.cluster.local:15010",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Istio system services should be service-specific",
		},
		{
			name:        "service-specific kube-system",
			routeName:   "kube-dns.kube-system.svc.cluster.local:53",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Kubernetes system services should be service-specific",
		},
		{
			name:        "service-specific external service",
			routeName:   "api.external.com:443",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "External service routes should be service-specific",
		},
		{
			name:        "default to service-specific",
			routeName:   "some_other_route",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Unknown route patterns should default to service-specific",
		},
		{
			name:        "complex service name",
			routeName:   "payment-service-v2.production",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Complex service names should be service-specific",
		},
		{
			name:        "UUID-like route name",
			routeName:   "550e8400-e29b-41d4-a716-446655440000",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "UUID-like names should be service-specific",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			route := &v1alpha1.RouteConfigSummary{
				Name: test.routeName,
				Type: test.currentType,
			}

			err := enrichFunc(route)
			require.NoError(t, err, test.description)
			assert.Equal(t, test.expected, route.Type, test.description)
		})
	}

	t.Run("handles nil route", func(t *testing.T) {
		err := enrichFunc(nil)
		assert.NoError(t, err)
	})
}

func TestIsPortOnlyRoute(t *testing.T) {
	tests := []struct {
		name        string
		route       string
		expected    bool
		description string
	}{
		{
			name:        "valid standard HTTP port",
			route:       "80",
			expected:    true,
			description: "Standard HTTP port should be valid",
		},
		{
			name:        "valid standard HTTPS port",
			route:       "443",
			expected:    true,
			description: "Standard HTTPS port should be valid",
		},
		{
			name:        "valid custom port",
			route:       "8080",
			expected:    true,
			description: "Common custom port should be valid",
		},
		{
			name:        "maximum valid port",
			route:       "65535",
			expected:    true,
			description: "Maximum TCP port should be valid",
		},
		{
			name:        "minimum valid port",
			route:       "1",
			expected:    true,
			description: "Minimum valid port should be valid",
		},
		{
			name:        "common database port",
			route:       "5432",
			expected:    true,
			description: "PostgreSQL port should be valid",
		},
		{
			name:        "common mongodb port",
			route:       "27017",
			expected:    true,
			description: "MongoDB port should be valid",
		},
		{
			name:        "istio discovery port",
			route:       "15010",
			expected:    true,
			description: "Istio discovery port should be valid",
		},
		{
			name:        "istio proxy admin port",
			route:       "15000",
			expected:    true,
			description: "Istio proxy admin port should be valid",
		},
		{
			name:        "invalid port zero",
			route:       "0",
			expected:    false,
			description: "Port zero is not valid for application use",
		},
		{
			name:        "invalid port too large",
			route:       "65536",
			expected:    false,
			description: "Port above 65535 is invalid",
		},
		{
			name:        "invalid port way too large",
			route:       "99999",
			expected:    false,
			description: "Very large port numbers are invalid",
		},
		{
			name:        "too long numeric string",
			route:       "123456",
			expected:    false,
			description: "More than 5 digits cannot be valid port",
		},
		{
			name:        "contains letters",
			route:       "8080a",
			expected:    false,
			description: "Alphanumeric strings are not ports",
		},
		{
			name:        "contains special characters",
			route:       "8080-",
			expected:    false,
			description: "Special characters invalidate port numbers",
		},
		{
			name:        "leading zeros",
			route:       "08080",
			expected:    true,
			description: "Leading zeros should still be valid port",
		},
		{
			name:        "port with spaces",
			route:       " 8080 ",
			expected:    false,
			description: "Spaces invalidate port detection",
		},
		{
			name:        "empty string",
			route:       "",
			expected:    false,
			description: "Empty string is not a port",
		},
		{
			name:        "port-like with colon",
			route:       "8080:tcp",
			expected:    false,
			description: "Port specification with protocol is not just a port",
		},
		{
			name:        "negative number",
			route:       "-8080",
			expected:    false,
			description: "Negative numbers are not valid ports",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isPortOnlyRoute(test.route)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestIsIstioStaticRoute(t *testing.T) {
	tests := []struct {
		name        string
		route       string
		expected    bool
		description string
	}{
		{
			name:        "empty route",
			route:       "",
			expected:    true,
			description: "Empty routes should be considered static",
		},
		{
			name:        "whitespace only route",
			route:       "   ",
			expected:    true,
			description: "Whitespace-only routes should be static",
		},
		{
			name:        "tab and newline whitespace",
			route:       "\t\n  \r",
			expected:    true,
			description: "Various whitespace characters should be static",
		},
		{
			name:        "BlackHoleCluster static pattern",
			route:       "BlackHoleCluster",
			expected:    true,
			description: "Istio BlackHoleCluster should be static",
		},
		{
			name:        "InboundPassthroughCluster static pattern",
			route:       "InboundPassthroughCluster",
			expected:    true,
			description: "Istio InboundPassthroughCluster should be static",
		},
		{
			name:        "PassthroughCluster static pattern",
			route:       "PassthroughCluster",
			expected:    true,
			description: "Istio PassthroughCluster should be static",
		},
		{
			name:        "local_agent static pattern",
			route:       "local_agent",
			expected:    true,
			description: "Local agent routes should be static",
		},
		{
			name:        "admin static pattern",
			route:       "admin",
			expected:    true,
			description: "Admin routes should be static",
		},
		{
			name:        "localhost with port",
			route:       "localhost:8080",
			expected:    true,
			description: "Localhost routes with ports should be static",
		},
		{
			name:        "localhost without port",
			route:       "localhost",
			expected:    true,
			description: "Localhost routes without ports should be static",
		},
		{
			name:        "127.0.0.1 with port",
			route:       "127.0.0.1:8080",
			expected:    true,
			description: "IPv4 loopback with port should be static",
		},
		{
			name:        "127.0.0.1 without port",
			route:       "127.0.0.1",
			expected:    true,
			description: "IPv4 loopback without port should be static",
		},
		{
			name:        "127.0.0.1 with path",
			route:       "127.0.0.1:8080/admin",
			expected:    true,
			description: "Localhost with path should be static",
		},
		{
			name:        "localhost with path",
			route:       "localhost/health",
			expected:    true,
			description: "Localhost with path should be static",
		},
		{
			name:        "regular service route",
			route:       "backend_service",
			expected:    false,
			description: "Regular service names should not be static",
		},
		{
			name:        "kubernetes service FQDN",
			route:       "backend.demo.svc.cluster.local:8080",
			expected:    false,
			description: "Kubernetes FQDNs should not be static",
		},
		{
			name:        "external service domain",
			route:       "api.external.com:443",
			expected:    false,
			description: "External domains should not be static",
		},
		{
			name:        "istio system service",
			route:       "istiod.istio-system.svc.cluster.local:15010",
			expected:    false,
			description: "Istio system services should not be static",
		},
		{
			name:        "port-only route",
			route:       "8080",
			expected:    false,
			description: "Port-only routes should not be static",
		},
		{
			name:        "cluster IP address",
			route:       "10.96.1.100:8080",
			expected:    false,
			description: "Cluster IP addresses should not be static",
		},
		{
			name:        "pipe-separated route",
			route:       "backend|8080",
			expected:    false,
			description: "Pipe-separated routes should not be static",
		},
		{
			name:        "service with version",
			route:       "backend-v2.production",
			expected:    false,
			description: "Versioned services should not be static",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isIstioStaticRoute(test.route)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestInferIstioRouteType(t *testing.T) {
	tests := []struct {
		name        string
		routeName   string
		currentType v1alpha1.RouteType
		expected    v1alpha1.RouteType
		description string
	}{
		{
			name:        "preserves static classification",
			routeName:   "backend.demo.svc.cluster.local:8080",
			currentType: v1alpha1.RouteType_STATIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Static routes should always remain static",
		},
		{
			name:        "preserves static even with pipe separator",
			routeName:   "backend|8080",
			currentType: v1alpha1.RouteType_STATIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Static classification overrides pipe detection",
		},
		{
			name:        "pipe separator service-specific",
			routeName:   "backend.demo.svc.cluster.local|8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Pipe separators indicate service-specific routing",
		},
		{
			name:        "complex pipe separator with version",
			routeName:   "backend.demo.svc.cluster.local|8080|v1",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Complex pipe patterns are service-specific",
		},
		{
			name:        "port-only route detection",
			routeName:   "80",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Standard HTTP port should be port-based",
		},
		{
			name:        "port-only HTTPS",
			routeName:   "443",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Standard HTTPS port should be port-based",
		},
		{
			name:        "port-only custom",
			routeName:   "8080",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Custom ports should be port-based",
		},
		{
			name:        "port-only istio control",
			routeName:   "15010",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_PORT_BASED,
			description: "Istio control plane ports should be port-based",
		},
		{
			name:        "static localhost detection",
			routeName:   "localhost:8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Localhost routes should be static",
		},
		{
			name:        "static 127.0.0.1 detection",
			routeName:   "127.0.0.1:9090",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Loopback IP routes should be static",
		},
		{
			name:        "static BlackHoleCluster",
			routeName:   "BlackHoleCluster",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Istio BlackHoleCluster should be static",
		},
		{
			name:        "static PassthroughCluster",
			routeName:   "PassthroughCluster",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Istio PassthroughCluster should be static",
		},
		{
			name:        "static InboundPassthroughCluster",
			routeName:   "InboundPassthroughCluster",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Istio InboundPassthroughCluster should be static",
		},
		{
			name:        "static local_agent",
			routeName:   "local_agent",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Local agent routes should be static",
		},
		{
			name:        "static admin",
			routeName:   "admin",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Admin routes should be static",
		},
		{
			name:        "static empty route",
			routeName:   "",
			currentType: v1alpha1.RouteType_SERVICE_SPECIFIC,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Empty routes should be static",
		},
		{
			name:        "static whitespace route",
			routeName:   "   \t\n  ",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_STATIC,
			description: "Whitespace-only routes should be static",
		},
		{
			name:        "service-specific FQDN",
			routeName:   "backend.demo.svc.cluster.local:8080",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Kubernetes FQDN routes should be service-specific",
		},
		{
			name:        "service-specific external domain",
			routeName:   "api.external.com:443",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "External service domains should be service-specific",
		},
		{
			name:        "service-specific istio system",
			routeName:   "istiod.istio-system.svc.cluster.local:15010",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Istio system services should be service-specific",
		},
		{
			name:        "service-specific unknown pattern",
			routeName:   "some-unknown-service-name",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Unknown patterns should default to service-specific",
		},
		{
			name:        "service-specific complex name",
			routeName:   "payment-service-v2.production.company.internal",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "Complex service names should be service-specific",
		},
		{
			name:        "service-specific UUID-like",
			routeName:   "550e8400-e29b-41d4-a716-446655440000",
			currentType: v1alpha1.RouteType_PORT_BASED,
			expected:    v1alpha1.RouteType_SERVICE_SPECIFIC,
			description: "UUID-like identifiers should be service-specific",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := inferIstioRouteType(test.routeName, test.currentType)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}
