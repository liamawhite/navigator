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
	// Test with default sidecar proxy mode
	enrichFunc := enrichListenerType(v1alpha1.ProxyMode_SIDECAR)

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

func TestEnrichListenerTypeGateway(t *testing.T) {
	// Test with gateway proxy mode
	enrichFunc := enrichListenerType(v1alpha1.ProxyMode_GATEWAY)

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
			name:           "gateway HTTP inbound",
			listenerName:   "0.0.0.0_80",
			address:        "0.0.0.0",
			port:           80,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_GATEWAY_INBOUND,
			description:    "Gateway HTTP listener should be GATEWAY_INBOUND",
		},
		{
			name:           "gateway HTTPS inbound",
			listenerName:   "0.0.0.0_443",
			address:        "0.0.0.0",
			port:           443,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_GATEWAY_INBOUND,
			description:    "Gateway HTTPS listener should be GATEWAY_INBOUND",
		},
		{
			name:           "gateway custom port inbound",
			listenerName:   "0.0.0.0_8080",
			address:        "0.0.0.0",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_GATEWAY_INBOUND,
			description:    "Gateway custom port listener should be GATEWAY_INBOUND",
		},
		{
			name:           "gateway with OriginalDst still virtual outbound",
			listenerName:   "0.0.0.0_15001",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Gateway with OriginalDst should still be virtual outbound",
		},
		{
			name:           "gateway metrics still proxy metrics",
			listenerName:   "metrics",
			address:        "0.0.0.0",
			port:           15090,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PROXY_METRICS,
			description:    "Gateway metrics listener should still be PROXY_METRICS",
		},
		{
			name:           "gateway health check still proxy health",
			listenerName:   "health",
			address:        "0.0.0.0",
			port:           15021,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_PROXY_HEALTHCHECK,
			description:    "Gateway health check listener should still be PROXY_HEALTHCHECK",
		},
		{
			name:           "gateway service-specific still service outbound",
			listenerName:   "10.96.1.100_8080",
			address:        "10.96.1.100",
			port:           8080,
			useOriginalDst: false,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Gateway service-specific listener should still be SERVICE_OUTBOUND",
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
}

func TestInferIstioListenerType(t *testing.T) {
	tests := []struct {
		name           string
		listenerName   string
		address        string
		port           uint32
		useOriginalDst bool
		proxyMode      v1alpha1.ProxyMode
		expectedType   v1alpha1.ListenerType
		description    string
	}{
		{
			name:           "virtual inbound by name",
			listenerName:   "virtualInbound",
			address:        "192.168.1.1",
			port:           9999,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_SIDECAR,
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

	// Add proxy mode to all existing tests (default to SIDECAR for backward compatibility)
	for i := range tests {
		if tests[i].proxyMode == v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE {
			tests[i].proxyMode = v1alpha1.ProxyMode_SIDECAR
		}
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := inferIstioListenerType(test.listenerName, test.address, test.port, test.useOriginalDst, test.proxyMode)
			assert.Equal(t, test.expectedType, result, test.description)
		})
	}
}

func TestInferIstioListenerTypeGateway(t *testing.T) {
	tests := []struct {
		name           string
		listenerName   string
		address        string
		port           uint32
		useOriginalDst bool
		proxyMode      v1alpha1.ProxyMode
		expectedType   v1alpha1.ListenerType
		description    string
	}{
		{
			name:           "gateway HTTP inbound",
			listenerName:   "0.0.0.0_80",
			address:        "0.0.0.0",
			port:           80,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_GATEWAY,
			expectedType:   v1alpha1.ListenerType_GATEWAY_INBOUND,
			description:    "Gateway 0.0.0.0 listener without OriginalDst should be GATEWAY_INBOUND",
		},
		{
			name:           "gateway HTTPS inbound",
			listenerName:   "0.0.0.0_443",
			address:        "0.0.0.0",
			port:           443,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_GATEWAY,
			expectedType:   v1alpha1.ListenerType_GATEWAY_INBOUND,
			description:    "Gateway HTTPS listener should be GATEWAY_INBOUND",
		},
		{
			name:           "gateway custom port inbound",
			listenerName:   "gateway_8080",
			address:        "0.0.0.0",
			port:           8080,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_GATEWAY,
			expectedType:   v1alpha1.ListenerType_GATEWAY_INBOUND,
			description:    "Gateway custom port should be GATEWAY_INBOUND",
		},
		{
			name:           "sidecar same config still port outbound",
			listenerName:   "0.0.0.0_80",
			address:        "0.0.0.0",
			port:           80,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_SIDECAR,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Sidecar with same config should still be PORT_OUTBOUND",
		},
		{
			name:           "gateway with OriginalDst still virtual outbound",
			listenerName:   "0.0.0.0_15001",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			proxyMode:      v1alpha1.ProxyMode_GATEWAY,
			expectedType:   v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
			description:    "Gateway with OriginalDst should be VIRTUAL_OUTBOUND",
		},
		{
			name:           "gateway metrics port still proxy metrics",
			listenerName:   "metrics",
			address:        "0.0.0.0",
			port:           15090,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_GATEWAY,
			expectedType:   v1alpha1.ListenerType_PROXY_METRICS,
			description:    "Gateway metrics port should still be PROXY_METRICS",
		},
		{
			name:           "gateway health port still proxy health",
			listenerName:   "health",
			address:        "0.0.0.0",
			port:           15021,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_GATEWAY,
			expectedType:   v1alpha1.ListenerType_PROXY_HEALTHCHECK,
			description:    "Gateway health port should still be PROXY_HEALTHCHECK",
		},
		{
			name:           "gateway service IP still service outbound",
			listenerName:   "backend_service",
			address:        "10.96.1.100",
			port:           8080,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_GATEWAY,
			expectedType:   v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:    "Gateway service IP should still be SERVICE_OUTBOUND",
		},
		{
			name:           "unknown proxy mode defaults to port outbound",
			listenerName:   "0.0.0.0_80",
			address:        "0.0.0.0",
			port:           80,
			useOriginalDst: false,
			proxyMode:      v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			expectedType:   v1alpha1.ListenerType_PORT_OUTBOUND,
			description:    "Unknown proxy mode should default to PORT_OUTBOUND",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := inferIstioListenerType(test.listenerName, test.address, test.port, test.useOriginalDst, test.proxyMode)
			assert.Equal(t, test.expectedType, result, test.description)
		})
	}
}

func TestEnrichListenerMatchDestination(t *testing.T) {
	enrichFunc := enrichListenerMatchDestination()

	tests := []struct {
		name              string
		listener          *v1alpha1.ListenerSummary
		expectedDestCount int
		expectedDest0Type string
		expectedDest0FQDN string
		expectedDest0Port uint32
		expectedMatchType string
		expectedPathMatch string
		description       string
	}{
		{
			name: "enrich HTTP destination with Istio cluster",
			listener: &v1alpha1.ListenerSummary{
				Name:    "outbound_0.0.0.0_80",
				Type:    v1alpha1.ListenerType_PORT_OUTBOUND,
				Address: "0.0.0.0",
				Port:    80,
				Rules: []*v1alpha1.ListenerRule{
					{
						Destination: &v1alpha1.ListenerDestination{
							DestinationType: "cluster",
							ClusterName:     "outbound|80|v1|myservice.mynamespace.svc.cluster.local",
						},
					},
				},
			},
			expectedDestCount: 1,
			expectedDest0Type: "outbound",
			expectedDest0FQDN: "myservice.mynamespace.svc.cluster.local",
			expectedDest0Port: 80,
			description:       "Should enrich destination with service FQDN from Istio cluster name",
		},
		{
			name: "enrich TCP destination with passthrough cluster",
			listener: &v1alpha1.ListenerSummary{
				Name:    "virtualOutbound",
				Type:    v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
				Address: "0.0.0.0",
				Port:    15001,
				Rules: []*v1alpha1.ListenerRule{
					{
						Destination: &v1alpha1.ListenerDestination{
							DestinationType: "cluster",
							ClusterName:     "PassthroughCluster",
						},
					},
				},
			},
			expectedDestCount: 1,
			expectedDest0Type: "passthrough",
			expectedDest0FQDN: "",
			expectedDest0Port: 0,
			description:       "Should classify PassthroughCluster as passthrough destination type",
		},
		{
			name: "enrich HTTP route match for inbound listener",
			listener: &v1alpha1.ListenerSummary{
				Name:    "virtualInbound",
				Type:    v1alpha1.ListenerType_VIRTUAL_INBOUND,
				Address: "0.0.0.0",
				Port:    15006,
				Rules: []*v1alpha1.ListenerRule{
					{
						Match: &v1alpha1.ListenerMatch{
							MatchType: &v1alpha1.ListenerMatch_HttpRoute{
								HttpRoute: &v1alpha1.HttpRouteMatch{
									PathMatch: &v1alpha1.PathMatchInfo{
										MatchType: "prefix",
										Path:      "/health",
									},
								},
							},
						},
					},
				},
			},
			expectedMatchType: "istio_health",
			expectedPathMatch: "/health",
			description:       "Should enrich inbound HTTP route match and identify health endpoint",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := enrichFunc(test.listener)
			require.NoError(t, err, test.description)

			// Check destinations if expected
			if test.expectedDestCount > 0 {
				require.Len(t, test.listener.Rules, test.expectedDestCount, "Should have expected number of rules")
				rule0 := test.listener.Rules[0]
				require.NotNil(t, rule0.Destination, "First rule should have a destination")
				dest0 := rule0.Destination
				assert.Equal(t, test.expectedDest0Type, dest0.DestinationType, test.description)
				if test.expectedDest0FQDN != "" {
					assert.Equal(t, test.expectedDest0FQDN, dest0.ServiceFqdn, test.description)
				}
				if test.expectedDest0Port > 0 {
					assert.Equal(t, test.expectedDest0Port, dest0.Port, test.description)
				}
			}

			// Check matches if expected
			if test.expectedMatchType != "" {
				require.Len(t, test.listener.Rules, 1, "Should have exactly one rule")
				rule := test.listener.Rules[0]
				require.NotNil(t, rule.Match, "Rule should have a match")
				match := rule.Match
				httpRoute := match.GetHttpRoute()
				require.NotNil(t, httpRoute, "Should have HTTP route match")
				require.NotNil(t, httpRoute.PathMatch, "Should have path match")
				assert.Equal(t, test.expectedMatchType, httpRoute.PathMatch.MatchType, test.description)
				if test.expectedPathMatch != "" {
					assert.Equal(t, test.expectedPathMatch, httpRoute.PathMatch.Path, test.description)
				}
			}
		})
	}

	t.Run("handles nil listener", func(t *testing.T) {
		err := enrichFunc(nil)
		assert.NoError(t, err)
	})

	t.Run("handles empty listener", func(t *testing.T) {
		listener := &v1alpha1.ListenerSummary{}
		err := enrichFunc(listener)
		assert.NoError(t, err)
	})
}

func TestEnrichDestinationWithIstioInfo(t *testing.T) {
	tests := []struct {
		name         string
		destination  *v1alpha1.ListenerDestination
		expectedType string
		expectedFQDN string
		expectedPort uint32
		description  string
	}{
		{
			name: "outbound service cluster",
			destination: &v1alpha1.ListenerDestination{
				DestinationType: "cluster",
				ClusterName:     "outbound|80|v1|httpbin.default.svc.cluster.local",
			},
			expectedType: "outbound",
			expectedFQDN: "httpbin.default.svc.cluster.local",
			expectedPort: 80,
			description:  "Should parse Istio outbound cluster name",
		},
		{
			name: "inbound service cluster",
			destination: &v1alpha1.ListenerDestination{
				DestinationType: "cluster",
				ClusterName:     "inbound|8080|http|myapp.mynamespace.svc.cluster.local",
			},
			expectedType: "inbound",
			expectedFQDN: "myapp.mynamespace.svc.cluster.local",
			expectedPort: 8080,
			description:  "Should parse Istio inbound cluster name",
		},
		{
			name: "passthrough cluster",
			destination: &v1alpha1.ListenerDestination{
				DestinationType: "cluster",
				ClusterName:     "PassthroughCluster",
			},
			expectedType: "passthrough",
			expectedFQDN: "",
			expectedPort: 0,
			description:  "Should classify PassthroughCluster",
		},
		{
			name: "blackhole cluster",
			destination: &v1alpha1.ListenerDestination{
				DestinationType: "cluster",
				ClusterName:     "BlackHoleCluster",
			},
			expectedType: "blackhole",
			expectedFQDN: "",
			expectedPort: 0,
			description:  "Should classify BlackHoleCluster",
		},
		{
			name: "inbound passthrough cluster",
			destination: &v1alpha1.ListenerDestination{
				DestinationType: "cluster",
				ClusterName:     "InboundPassthroughCluster",
			},
			expectedType: "passthrough",
			expectedFQDN: "",
			expectedPort: 0,
			description:  "Should classify InboundPassthroughCluster",
		},
		{
			name: "outbound service with subset",
			destination: &v1alpha1.ListenerDestination{
				DestinationType: "cluster",
				ClusterName:     "outbound|443|v2|api.production.svc.cluster.local",
			},
			expectedType: "outbound",
			expectedFQDN: "api.production.svc.cluster.local",
			expectedPort: 443,
			description:  "Should parse outbound cluster with subset",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			enrichDestinationWithIstioInfo(test.destination)

			assert.Equal(t, test.expectedType, test.destination.DestinationType, test.description)
			assert.Equal(t, test.expectedFQDN, test.destination.ServiceFqdn, test.description)
			if test.expectedPort > 0 {
				assert.Equal(t, test.expectedPort, test.destination.Port, test.description)
			}
		})
	}

	t.Run("handles nil destination", func(t *testing.T) {
		enrichDestinationWithIstioInfo(nil) // Should not panic
	})

	t.Run("handles empty cluster name", func(t *testing.T) {
		dest := &v1alpha1.ListenerDestination{ClusterName: ""}
		enrichDestinationWithIstioInfo(dest)
		// Should not modify anything
	})
}

func TestEnrichMatchWithIstioInfo(t *testing.T) {
	t.Run("enriches HTTP route match", func(t *testing.T) {
		match := &v1alpha1.ListenerMatch{
			MatchType: &v1alpha1.ListenerMatch_HttpRoute{
				HttpRoute: &v1alpha1.HttpRouteMatch{
					PathMatch: &v1alpha1.PathMatchInfo{
						Path: "/health",
					},
					HeaderMatches: []*v1alpha1.HeaderMatchInfo{
						{
							Name:  ":authority",
							Value: "myservice.mynamespace.svc.cluster.local",
						},
					},
				},
			},
		}
		listener := &v1alpha1.ListenerSummary{
			Type: v1alpha1.ListenerType_VIRTUAL_INBOUND,
		}

		enrichMatchWithIstioInfo(match, listener)

		httpRoute := match.GetHttpRoute()
		require.NotNil(t, httpRoute)
		assert.Equal(t, "istio_health", httpRoute.PathMatch.MatchType)
		assert.Equal(t, "istio_service_host", httpRoute.HeaderMatches[0].Name)
	})

	t.Run("enriches filter chain match", func(t *testing.T) {
		match := &v1alpha1.ListenerMatch{
			MatchType: &v1alpha1.ListenerMatch_FilterChain{
				FilterChain: &v1alpha1.FilterChainMatch{
					ServerNames:       []string{"myservice.mynamespace.svc.cluster.local"},
					TransportProtocol: "tls",
				},
			},
		}
		listener := &v1alpha1.ListenerSummary{}

		enrichMatchWithIstioInfo(match, listener)

		filterChain := match.GetFilterChain()
		require.NotNil(t, filterChain)
		assert.Equal(t, "istio_service_myservice.mynamespace.svc.cluster.local", filterChain.ServerNames[0])
	})

	t.Run("enriches TCP proxy match", func(t *testing.T) {
		match := &v1alpha1.ListenerMatch{
			MatchType: &v1alpha1.ListenerMatch_TcpProxy{
				TcpProxy: &v1alpha1.TcpProxyMatch{
					ClusterName: "outbound|80||myservice.mynamespace.svc.cluster.local",
				},
			},
		}
		listener := &v1alpha1.ListenerSummary{}

		enrichMatchWithIstioInfo(match, listener)

		tcpProxy := match.GetTcpProxy()
		require.NotNil(t, tcpProxy)
		assert.Equal(t, "istio_service_outbound|80||myservice.mynamespace.svc.cluster.local", tcpProxy.ClusterName)
	})

	t.Run("handles nil match", func(t *testing.T) {
		listener := &v1alpha1.ListenerSummary{}
		enrichMatchWithIstioInfo(nil, listener) // Should not panic
	})

	t.Run("handles nil listener", func(t *testing.T) {
		match := &v1alpha1.ListenerMatch{}
		enrichMatchWithIstioInfo(match, nil) // Should not panic
	})

	t.Run("handles empty match", func(t *testing.T) {
		match := &v1alpha1.ListenerMatch{}
		listener := &v1alpha1.ListenerSummary{}
		enrichMatchWithIstioInfo(match, listener) // Should not panic
	})
}
