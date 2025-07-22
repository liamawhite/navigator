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
)

func TestParseDirection(t *testing.T) {
	tests := []struct {
		name        string
		direction   string
		expected    v1alpha1.ClusterDirection
		description string
	}{
		{
			name:        "outbound lowercase",
			direction:   "outbound",
			expected:    v1alpha1.ClusterDirection_OUTBOUND,
			description: "Standard outbound direction",
		},
		{
			name:        "outbound uppercase",
			direction:   "OUTBOUND",
			expected:    v1alpha1.ClusterDirection_OUTBOUND,
			description: "Case insensitive outbound",
		},
		{
			name:        "outbound mixed case",
			direction:   "OutBound",
			expected:    v1alpha1.ClusterDirection_OUTBOUND,
			description: "Mixed case outbound",
		},
		{
			name:        "inbound lowercase",
			direction:   "inbound",
			expected:    v1alpha1.ClusterDirection_INBOUND,
			description: "Standard inbound direction",
		},
		{
			name:        "inbound uppercase",
			direction:   "INBOUND",
			expected:    v1alpha1.ClusterDirection_INBOUND,
			description: "Case insensitive inbound",
		},
		{
			name:        "inbound mixed case",
			direction:   "InBound",
			expected:    v1alpha1.ClusterDirection_INBOUND,
			description: "Mixed case inbound",
		},
		{
			name:        "unknown direction",
			direction:   "unknown",
			expected:    v1alpha1.ClusterDirection_UNSPECIFIED,
			description: "Unknown direction should be unspecified",
		},
		{
			name:        "empty direction",
			direction:   "",
			expected:    v1alpha1.ClusterDirection_UNSPECIFIED,
			description: "Empty direction should be unspecified",
		},
		{
			name:        "invalid direction",
			direction:   "invalid-direction",
			expected:    v1alpha1.ClusterDirection_UNSPECIFIED,
			description: "Invalid direction should be unspecified",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseDirection(test.direction)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestParseClusterComponents(t *testing.T) {
	tests := []struct {
		name                string
		clusterName         string
		expectedDirection   v1alpha1.ClusterDirection
		expectedPort        uint32
		expectedSubset      string
		expectedServiceFqdn string
		description         string
	}{
		{
			name:                "complete outbound cluster",
			clusterName:         "outbound|8080|v1|backend.demo.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "v1",
			expectedServiceFqdn: "backend.demo.svc.cluster.local",
			description:         "Full four-part cluster name",
		},
		{
			name:                "inbound cluster",
			clusterName:         "inbound|8080||",
			expectedDirection:   v1alpha1.ClusterDirection_INBOUND,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Standard inbound cluster",
		},
		{
			name:                "outbound without subset",
			clusterName:         "outbound|443||api.external.com",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        443,
			expectedSubset:      "",
			expectedServiceFqdn: "api.external.com",
			description:         "External service without subset",
		},
		{
			name:                "minimal cluster - direction only",
			clusterName:         "outbound",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Only direction component",
		},
		{
			name:                "direction and port",
			clusterName:         "inbound|9090",
			expectedDirection:   v1alpha1.ClusterDirection_INBOUND,
			expectedPort:        9090,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Direction and port only",
		},
		{
			name:                "direction, port and subset",
			clusterName:         "outbound|8080|canary",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "canary",
			expectedServiceFqdn: "",
			description:         "First three components",
		},
		{
			name:                "invalid port",
			clusterName:         "outbound|invalid||service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Invalid port should default to 0",
		},
		{
			name:                "large port number",
			clusterName:         "outbound|65535||service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        65535,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Maximum valid port",
		},
		{
			name:                "empty cluster name",
			clusterName:         "",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Empty cluster name",
		},
		{
			name:                "no pipe separators",
			clusterName:         "simple-cluster-name",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Non-Istio cluster format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			direction, port, subset, serviceFqdn := parseClusterComponents(test.clusterName)
			assert.Equal(t, test.expectedDirection, direction, test.description)
			assert.Equal(t, test.expectedPort, port, test.description)
			assert.Equal(t, test.expectedSubset, subset, test.description)
			assert.Equal(t, test.expectedServiceFqdn, serviceFqdn, test.description)
		})
	}
}

func TestIsIstioClusterPattern(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		expected    bool
		description string
	}{
		{
			name:        "valid outbound cluster",
			clusterName: "outbound|8080|v1|backend.demo.svc.cluster.local",
			expected:    true,
			description: "Complete outbound cluster should match",
		},
		{
			name:        "valid inbound cluster",
			clusterName: "inbound|8080||",
			expected:    true,
			description: "Complete inbound cluster should match",
		},
		{
			name:        "outbound with empty components",
			clusterName: "outbound|||",
			expected:    true,
			description: "Outbound with empty components should match",
		},
		{
			name:        "inbound with service fqdn",
			clusterName: "inbound|8080||service.ns.svc.cluster.local",
			expected:    true,
			description: "Inbound with FQDN should match",
		},
		{
			name:        "too few parts - 3 components",
			clusterName: "outbound|8080|v1",
			expected:    false,
			description: "Only 3 parts should not match",
		},
		{
			name:        "too few parts - 2 components",
			clusterName: "outbound|8080",
			expected:    false,
			description: "Only 2 parts should not match",
		},
		{
			name:        "too few parts - 1 component",
			clusterName: "outbound",
			expected:    false,
			description: "Only 1 part should not match",
		},
		{
			name:        "too many parts - 5 components",
			clusterName: "outbound|8080|v1|service.ns.svc.cluster.local|extra",
			expected:    false,
			description: "5 parts should not match",
		},
		{
			name:        "no pipe separators",
			clusterName: "simple-cluster-name",
			expected:    false,
			description: "No pipes should not match",
		},
		{
			name:        "invalid direction",
			clusterName: "invalid|8080||service.ns.svc.cluster.local",
			expected:    false,
			description: "Invalid direction should not match",
		},
		{
			name:        "empty cluster name",
			clusterName: "",
			expected:    false,
			description: "Empty name should not match",
		},
		{
			name:        "static cluster name",
			clusterName: "prometheus_stats",
			expected:    false,
			description: "Static cluster should not match",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isIstioClusterPattern(test.clusterName)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestParseFQDN(t *testing.T) {
	tests := []struct {
		name            string
		serviceFqdn     string
		expectedService string
		expectedNs      string
		description     string
	}{
		{
			name:            "standard kubernetes service",
			serviceFqdn:     "backend.demo.svc.cluster.local",
			expectedService: "backend",
			expectedNs:      "demo",
			description:     "Standard Kubernetes service FQDN",
		},
		{
			name:            "service in default namespace",
			serviceFqdn:     "kubernetes.default.svc.cluster.local",
			expectedService: "kubernetes",
			expectedNs:      "default",
			description:     "Kubernetes API server in default namespace",
		},
		{
			name:            "service in kube-system",
			serviceFqdn:     "kube-dns.kube-system.svc.cluster.local",
			expectedService: "kube-dns",
			expectedNs:      "kube-system",
			description:     "DNS service in kube-system namespace",
		},
		{
			name:            "service in istio-system",
			serviceFqdn:     "istiod.istio-system.svc.cluster.local",
			expectedService: "istiod",
			expectedNs:      "istio-system",
			description:     "Istio discovery service",
		},
		{
			name:            "service with hyphenated name",
			serviceFqdn:     "my-service.production.svc.cluster.local",
			expectedService: "my-service",
			expectedNs:      "production",
			description:     "Service with hyphens in name",
		},
		{
			name:            "service with hyphenated namespace",
			serviceFqdn:     "api.istio-system.svc.cluster.local",
			expectedService: "api",
			expectedNs:      "istio-system",
			description:     "Service in namespace with hyphens",
		},
		{
			name:            "external service",
			serviceFqdn:     "api.external.com",
			expectedService: "api.external.com",
			expectedNs:      "",
			description:     "External service should return full name as service",
		},
		{
			name:            "external service with subdomain",
			serviceFqdn:     "api.v1.external.company.com",
			expectedService: "api.v1.external.company.com",
			expectedNs:      "",
			description:     "Complex external domain should return full name",
		},
		{
			name:            "simple hostname",
			serviceFqdn:     "localhost",
			expectedService: "localhost",
			expectedNs:      "",
			description:     "Simple hostname should return as service",
		},
		{
			name:            "empty service fqdn",
			serviceFqdn:     "",
			expectedService: "",
			expectedNs:      "",
			description:     "Empty FQDN should return empty values",
		},
		{
			name:            "minimal kubernetes format",
			serviceFqdn:     "a.b.svc.cluster.local",
			expectedService: "a",
			expectedNs:      "b",
			description:     "Minimal valid Kubernetes FQDN",
		},
		{
			name:            "malformed kubernetes format - no namespace",
			serviceFqdn:     "service.svc.cluster.local",
			expectedService: "service",
			expectedNs:      "svc",
			description:     "Malformed FQDN treats svc as namespace",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serviceName, namespace := parseFQDN(test.serviceFqdn)
			assert.Equal(t, test.expectedService, serviceName, test.description)
			assert.Equal(t, test.expectedNs, namespace, test.description)
		})
	}
}

func TestInferClusterType(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		expected    v1alpha1.ClusterType
		description string
	}{
		{
			name:        "kubernetes service EDS",
			clusterName: "outbound|8080||backend.demo.svc.cluster.local",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Kubernetes service should use EDS",
		},
		{
			name:        "inbound cluster EDS",
			clusterName: "inbound|8080||",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Inbound cluster should use EDS",
		},
		{
			name:        "external service strict DNS",
			clusterName: "outbound|443||api.external.com",
			expected:    v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description: "External service should use strict DNS",
		},
		{
			name:        "complex external domain",
			clusterName: "outbound|443||api.v1.company.external.com",
			expected:    v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description: "Complex external domain should use strict DNS",
		},
		{
			name:        "prometheus static cluster",
			clusterName: "prometheus_stats",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "Prometheus stats should be static",
		},
		{
			name:        "agent static cluster",
			clusterName: "agent",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "Agent cluster should be static",
		},
		{
			name:        "sds-grpc static cluster",
			clusterName: "sds-grpc",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "SDS gRPC should be static",
		},
		{
			name:        "xds-grpc static cluster",
			clusterName: "xds-grpc",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "XDS gRPC should be static",
		},
		{
			name:        "zipkin static cluster",
			clusterName: "zipkin",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "Zipkin should be static",
		},
		{
			name:        "jaeger static cluster",
			clusterName: "jaeger",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "Jaeger should be static",
		},
		{
			name:        "envoy access log service",
			clusterName: "envoy_accesslog_service",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "Envoy access log should be static",
		},
		{
			name:        "IP address DNS",
			clusterName: "192.168.1.100",
			expected:    v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description: "IP address should use strict DNS",
		},
		{
			name:        "localhost DNS",
			clusterName: "127.0.0.1",
			expected:    v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description: "Localhost IP should use strict DNS",
		},
		{
			name:        "simple domain DNS",
			clusterName: "api.company.io",
			expected:    v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description: "Simple domain should use strict DNS",
		},
		{
			name:        "unknown cluster defaults to EDS",
			clusterName: "unknown-cluster",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Unknown format should default to EDS",
		},
		{
			name:        "empty cluster name",
			clusterName: "",
			expected:    v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
			description: "Empty name should be unknown",
		},
		{
			name:        "kubernetes API server",
			clusterName: "outbound|443||kubernetes.default.svc.cluster.local",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Kubernetes API should use EDS",
		},
		{
			name:        "istio control plane",
			clusterName: "outbound|15010||istiod.istio-system.svc.cluster.local",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Istio control plane should use EDS",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := InferClusterType(test.clusterName)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestInferClusterTypeFromName(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		expected    v1alpha1.ClusterType
		description string
	}{
		{
			name:        "outbound in cluster name",
			clusterName: "outbound|8080||backend.demo.svc.cluster.local",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Outbound clusters should use EDS",
		},
		{
			name:        "inbound in cluster name",
			clusterName: "inbound|8080||",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Inbound clusters should use EDS",
		},
		{
			name:        "outbound uppercase",
			clusterName: "OUTBOUND|443||api.service.com",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Case insensitive outbound detection",
		},
		{
			name:        "inbound mixed case",
			clusterName: "InBound|8080||",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
			description: "Case insensitive inbound detection",
		},
		{
			name:        "static cluster name",
			clusterName: "static-cluster-name",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "Static keyword should be detected",
		},
		{
			name:        "static uppercase",
			clusterName: "STATIC-CLUSTER",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
			description: "Case insensitive static detection",
		},
		{
			name:        "unknown cluster type",
			clusterName: "prometheus_stats",
			expected:    v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
			description: "Non-matching pattern should be unknown",
		},
		{
			name:        "empty cluster name",
			clusterName: "",
			expected:    v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
			description: "Empty name should be unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := InferClusterTypeFromName(test.clusterName)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestIsIstioCluster(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		expected    bool
		description string
	}{
		{
			name:        "valid istio outbound cluster",
			clusterName: "outbound|8080|v1|backend.demo.svc.cluster.local",
			expected:    true,
			description: "Standard Istio outbound should be detected",
		},
		{
			name:        "valid istio inbound cluster",
			clusterName: "inbound|8080||",
			expected:    true,
			description: "Standard Istio inbound should be detected",
		},
		{
			name:        "prometheus static cluster",
			clusterName: "prometheus_stats",
			expected:    true,
			description: "Prometheus stats is Istio static cluster",
		},
		{
			name:        "agent static cluster",
			clusterName: "agent",
			expected:    true,
			description: "Agent is Istio static cluster",
		},
		{
			name:        "sds-grpc static cluster",
			clusterName: "sds-grpc",
			expected:    true,
			description: "SDS gRPC is Istio static cluster",
		},
		{
			name:        "xds-grpc static cluster",
			clusterName: "xds-grpc",
			expected:    true,
			description: "XDS gRPC is Istio static cluster",
		},
		{
			name:        "zipkin static cluster",
			clusterName: "zipkin",
			expected:    true,
			description: "Zipkin is Istio static cluster",
		},
		{
			name:        "jaeger static cluster",
			clusterName: "jaeger",
			expected:    true,
			description: "Jaeger is Istio static cluster",
		},
		{
			name:        "envoy access log service",
			clusterName: "envoy_accesslog_service",
			expected:    true,
			description: "Envoy access log is Istio static cluster",
		},
		{
			name:        "non-istio cluster",
			clusterName: "random-cluster-name",
			expected:    false,
			description: "Random cluster should not be Istio",
		},
		{
			name:        "empty cluster name",
			clusterName: "",
			expected:    false,
			description: "Empty name should not be Istio",
		},
		{
			name:        "malformed istio pattern",
			clusterName: "outbound|8080",
			expected:    false,
			description: "Incomplete Istio pattern should not match",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsIstioCluster(test.clusterName)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		name        string
		serviceFqdn string
		expected    string
		description string
	}{
		{
			name:        "kubernetes service",
			serviceFqdn: "backend.demo.svc.cluster.local",
			expected:    "backend",
			description: "Should extract service name from Kubernetes FQDN",
		},
		{
			name:        "external service",
			serviceFqdn: "api.external.com",
			expected:    "api.external.com",
			description: "Should return full external domain as service name",
		},
		{
			name:        "empty fqdn",
			serviceFqdn: "",
			expected:    "",
			description: "Should return empty for empty FQDN",
		},
		{
			name:        "simple hostname",
			serviceFqdn: "localhost",
			expected:    "localhost",
			description: "Should return hostname as service name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractServiceName(test.serviceFqdn)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestExtractNamespace(t *testing.T) {
	tests := []struct {
		name        string
		serviceFqdn string
		expected    string
		description string
	}{
		{
			name:        "kubernetes service",
			serviceFqdn: "backend.demo.svc.cluster.local",
			expected:    "demo",
			description: "Should extract namespace from Kubernetes FQDN",
		},
		{
			name:        "external service",
			serviceFqdn: "api.external.com",
			expected:    "",
			description: "Should return empty namespace for external domain",
		},
		{
			name:        "empty fqdn",
			serviceFqdn: "",
			expected:    "",
			description: "Should return empty for empty FQDN",
		},
		{
			name:        "simple hostname",
			serviceFqdn: "localhost",
			expected:    "",
			description: "Should return empty namespace for hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractNamespace(test.serviceFqdn)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}
