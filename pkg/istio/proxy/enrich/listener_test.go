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

func TestEnrichListenerType(t *testing.T) {
	enrichFunc := enrichListenerType()

	tests := []struct {
		name           string
		listenerName   string
		address        string
		port           uint32
		useOriginalDst bool
		expectedType   v1alpha1.ListenerType
		description    string
	}{
		{
			name:           "virtual inbound listener",
			listenerName:   "virtualInbound",
			address:        "0.0.0.0",
			port:           15006,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_INBOUND,
			description:    "Istio virtual inbound listener for sidecar",
		},
		{
			name:           "virtual outbound listener",
			listenerName:   "virtualOutbound",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Istio virtual outbound listener for sidecar",
		},
		{
			name:           "prometheus metrics listener",
			listenerName:   "metrics",
			address:        "0.0.0.0",
			port:           15090,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PROXY_METRICS,
			description:    "Prometheus metrics endpoint",
		},
		{
			name:           "health check listener",
			listenerName:   "health",
			address:        "0.0.0.0",
			port:           15021,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PROXY_HEALTHCHECK,
			description:    "Health check endpoint for readiness/liveness",
		},
		{
			name:           "virtual outbound with port 15001",
			listenerName:   "0.0.0.0_15001",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Standard virtual outbound on port 15001",
		},
		{
			name:           "virtual outbound with port 15006",
			listenerName:   "0.0.0.0_15006",
			address:        "0.0.0.0",
			port:           15006,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Alternative virtual outbound on port 15006",
		},
		{
			name:           "port-based outbound HTTP",
			listenerName:   "0.0.0.0_80",
			address:        "0.0.0.0",
			port:           80,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Port-based HTTP outbound listener",
		},
		{
			name:           "port-based outbound HTTPS",
			listenerName:   "0.0.0.0_443",
			address:        "0.0.0.0",
			port:           443,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Port-based HTTPS outbound listener",
		},
		{
			name:           "port-based outbound custom port",
			listenerName:   "0.0.0.0_8080",
			address:        "0.0.0.0",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Port-based custom port outbound listener",
		},
		{
			name:           "service-specific outbound",
			listenerName:   "10.96.1.100_8080",
			address:        "10.96.1.100",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Service-specific outbound listener for ClusterIP",
		},
		{
			name:           "kubernetes API service",
			listenerName:   "10.96.0.1_443",
			address:        "10.96.0.1",
			port:           443,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Kubernetes API server service listener",
		},
		{
			name:           "DNS service outbound",
			listenerName:   "10.96.0.10_53",
			address:        "10.96.0.10",
			port:           53,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "DNS service outbound listener",
		},
		{
			name:           "istio control plane listener",
			listenerName:   "10.96.245.191_15010",
			address:        "10.96.245.191",
			port:           15010,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Istio discovery service listener",
		},
		{
			name:           "backend service listener",
			listenerName:   "10.96.173.1_8080",
			address:        "10.96.173.1",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Backend application service listener",
		},
		{
			name:           "database service listener",
			listenerName:   "10.96.200.50_5432",
			address:        "10.96.200.50",
			port:           5432,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Database service listener",
		},
		{
			name:           "monitoring service listener",
			listenerName:   "10.96.100.25_9090",
			address:        "10.96.100.25",
			port:           9090,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Prometheus monitoring service listener",
		},
		{
			name:           "headless service listener",
			listenerName:   "172.16.1.10_3306",
			address:        "172.16.1.10",
			port:           3306,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "MySQL headless service listener",
		},
		{
			name:           "external IP listener",
			listenerName:   "192.168.1.100_443",
			address:        "192.168.1.100",
			port:           443,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "External IP service listener",
		},
		{
			name:           "wildcard with OriginalDst disabled",
			listenerName:   "catch_all",
			address:        "0.0.0.0",
			port:           9999,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Generic wildcard listener without OriginalDst",
		},
		{
			name:           "unknown virtual listener name",
			listenerName:   "unknownVirtual",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Unknown virtual name should be classified by port and OriginalDst",
		},
		{
			name:           "localhost IPv4 address",
			listenerName:   "127.0.0.1_8080",
			address:        "127.0.0.1",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Localhost IPv4 service listener",
		},
		{
			name:           "IPv6 loopback address",
			listenerName:   "::1_8080",
			address:        "::1",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "IPv6 loopback service listener",
		},
		{
			name:           "high port number",
			listenerName:   "0.0.0.0_65535",
			address:        "0.0.0.0",
			port:           65535,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Maximum port number listener",
		},
		{
			name:           "standard HTTP alternative port",
			listenerName:   "0.0.0.0_8080",
			address:        "0.0.0.0",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Alternative HTTP port listener",
		},
		{
			name:           "HTTPS alternative port",
			listenerName:   "0.0.0.0_8443",
			address:        "0.0.0.0",
			port:           8443,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Alternative HTTPS port listener",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listener := &v1alpha1.ListenerSummary{
				Name:           test.listenerName,
				Address:        test.address,
				Port:           test.port,
				UseOriginalDst: test.useOriginalDst,
			}

			err := enrichFunc(listener)
			require.NoError(t, err, test.description)
			assert.Equal(t, test.expectedType, listener.Type, test.description)
		})
	}

	t.Run("handles nil listener", func(t *testing.T) {
		err := enrichFunc(nil)
		assert.NoError(t, err)
	})
}

func TestInferIstioListenerType(t *testing.T) {
	tests := []struct {
		name           string
		listenerName   string
		address        string
		port           uint32
		useOriginalDst bool
		expectedType   v1alpha1.ListenerType
		description    string
	}{
		{
			name:           "virtual inbound by name",
			listenerName:   "virtualInbound",
			address:        "192.168.1.1",
			port:           9999,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_INBOUND,
			description:    "Name takes precedence over other attributes",
		},
		{
			name:           "virtual outbound by name",
			listenerName:   "virtualOutbound",
			address:        "10.0.0.1",
			port:           1234,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Name takes precedence over other attributes",
		},
		{
			name:           "metrics port 15090",
			listenerName:   "stats",
			address:        "0.0.0.0",
			port:           15090,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PROXY_METRICS,
			description:    "Prometheus metrics port classification",
		},
		{
			name:           "health check port 15021",
			listenerName:   "health_check",
			address:        "0.0.0.0",
			port:           15021,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PROXY_HEALTHCHECK,
			description:    "Health check port classification",
		},
		{
			name:           "virtual outbound port 15001 with OriginalDst",
			listenerName:   "outbound_capture",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Port 15001 with OriginalDst should be virtual outbound",
		},
		{
			name:           "virtual outbound port 15006 with OriginalDst",
			listenerName:   "inbound_capture",
			address:        "0.0.0.0",
			port:           15006,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Port 15006 with OriginalDst should be virtual outbound",
		},
		{
			name:           "port-based without OriginalDst on 15001",
			listenerName:   "direct_listener",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Port 15001 without OriginalDst should be port-based",
		},
		{
			name:           "port-based without OriginalDst on 15006",
			listenerName:   "direct_listener",
			address:        "0.0.0.0",
			port:           15006,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Port 15006 without OriginalDst should be port-based",
		},
		{
			name:           "wildcard HTTP port",
			listenerName:   "http_listener",
			address:        "0.0.0.0",
			port:           80,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "HTTP port on wildcard should be port-based",
		},
		{
			name:           "wildcard HTTPS port",
			listenerName:   "https_listener",
			address:        "0.0.0.0",
			port:           443,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "HTTPS port on wildcard should be port-based",
		},
		{
			name:           "wildcard custom port",
			listenerName:   "custom_listener",
			address:        "0.0.0.0",
			port:           9090,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Custom port on wildcard should be port-based",
		},
		{
			name:           "service IP listener",
			listenerName:   "service_listener",
			address:        "10.96.1.100",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Specific IP should be service-specific",
		},
		{
			name:           "cluster IP listener",
			listenerName:   "backend_service",
			address:        "172.20.50.100",
			port:           3000,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Cluster IP should be service-specific",
		},
		{
			name:           "external IP listener",
			listenerName:   "external_service",
			address:        "203.0.113.10",
			port:           443,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "External IP should be service-specific",
		},
		{
			name:           "IPv6 address listener",
			listenerName:   "ipv6_service",
			address:        "2001:db8::1",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "IPv6 address should be service-specific",
		},
		{
			name:           "localhost listener",
			listenerName:   "local_service",
			address:        "127.0.0.1",
			port:           5000,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Localhost should be service-specific",
		},
		{
			name:           "empty name with wildcard",
			listenerName:   "",
			address:        "0.0.0.0",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Empty name with wildcard should be port-based",
		},
		{
			name:           "empty name with specific IP",
			listenerName:   "",
			address:        "10.0.0.1",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Empty name with specific IP should be service-specific",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := inferIstioListenerType(test.listenerName, test.address, test.port, test.useOriginalDst)
			assert.Equal(t, test.expectedType, result, test.description)
		})
	}
}
