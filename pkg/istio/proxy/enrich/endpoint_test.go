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

func TestEnrichEndpointClusterName(t *testing.T) {
	enrichFunc := enrichEndpointClusterName()

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
			name:                "outbound service endpoint",
			clusterName:         "outbound|8080|v1|backend.demo.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "v1",
			expectedServiceFqdn: "backend.demo.svc.cluster.local",
			description:         "Standard outbound service endpoint with subset",
		},
		{
			name:                "inbound service endpoint",
			clusterName:         "inbound|8080||",
			expectedDirection:   v1alpha1.ClusterDirection_INBOUND,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Standard inbound service endpoint",
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
			name:                "kubernetes service endpoint",
			clusterName:         "outbound|443||kubernetes.default.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        443,
			expectedSubset:      "",
			expectedServiceFqdn: "kubernetes.default.svc.cluster.local",
			description:         "Kubernetes API server endpoint",
		},
		{
			name:                "dns service endpoint",
			clusterName:         "outbound|53||kube-dns.kube-system.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        53,
			expectedSubset:      "",
			expectedServiceFqdn: "kube-dns.kube-system.svc.cluster.local",
			description:         "DNS service endpoint",
		},
		{
			name:                "istio control plane endpoint",
			clusterName:         "outbound|15010||istiod.istio-system.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        15010,
			expectedSubset:      "",
			expectedServiceFqdn: "istiod.istio-system.svc.cluster.local",
			description:         "Istio discovery service endpoint",
		},
		{
			name:                "production service with canary subset",
			clusterName:         "outbound|8080|canary|payment-service.production.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "canary",
			expectedServiceFqdn: "payment-service.production.svc.cluster.local",
			description:         "Production service with canary deployment",
		},
		{
			name:                "high port number endpoint",
			clusterName:         "outbound|32000||monitoring.observability.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        32000,
			expectedSubset:      "",
			expectedServiceFqdn: "monitoring.observability.svc.cluster.local",
			description:         "Service with high port number",
		},
		{
			name:                "non-istio cluster name",
			clusterName:         "prometheus_stats",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Static cluster name should not be parsed",
		},
		{
			name:                "malformed cluster name",
			clusterName:         "outbound|invalid-port||service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Invalid port should default to 0",
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
			name:                "cluster with only direction",
			clusterName:         "outbound",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Non-4-part cluster should use defaults",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			endpoint := &v1alpha1.EndpointSummary{
				ClusterName: test.clusterName,
			}

			err := enrichFunc(endpoint)
			require.NoError(t, err, test.description)
			assert.Equal(t, test.expectedDirection, endpoint.Direction, test.description)
			assert.Equal(t, test.expectedPort, endpoint.Port, test.description)
			assert.Equal(t, test.expectedSubset, endpoint.Subset, test.description)
			assert.Equal(t, test.expectedServiceFqdn, endpoint.ServiceFqdn, test.description)
		})
	}

	t.Run("handles nil endpoint", func(t *testing.T) {
		err := enrichFunc(nil)
		assert.NoError(t, err)
	})
}

func TestEnrichEndpointClusterType(t *testing.T) {
	enrichFunc := enrichEndpointClusterType()

	tests := []struct {
		name         string
		clusterName  string
		expectedType v1alpha1.ClusterType
		description  string
	}{
		{
			name:         "kubernetes service endpoint EDS",
			clusterName:  "outbound|8080||backend.demo.svc.cluster.local",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Kubernetes service should use EDS discovery",
		},
		{
			name:         "inbound endpoint EDS",
			clusterName:  "inbound|8080||",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Inbound clusters should use EDS discovery",
		},
		{
			name:         "external service strict DNS",
			clusterName:  "outbound|443||api.external.com",
			expectedType: v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description:  "External services should use strict DNS",
		},
		{
			name:         "external service with multiple domains",
			clusterName:  "outbound|443||api.v1.external.company.com",
			expectedType: v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description:  "Complex external domain should use strict DNS",
		},
		{
			name:         "kubernetes API server",
			clusterName:  "outbound|443||kubernetes.default.svc.cluster.local",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Kubernetes API should use EDS",
		},
		{
			name:         "istio control plane",
			clusterName:  "outbound|15010||istiod.istio-system.svc.cluster.local",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Istio control plane should use EDS",
		},
		{
			name:         "prometheus static cluster",
			clusterName:  "prometheus_stats",
			expectedType: v1alpha1.ClusterType_CLUSTER_STATIC,
			description:  "Prometheus stats should be static",
		},
		{
			name:         "agent static cluster",
			clusterName:  "agent",
			expectedType: v1alpha1.ClusterType_CLUSTER_STATIC,
			description:  "Agent cluster should be static",
		},
		{
			name:         "sds-grpc static cluster",
			clusterName:  "sds-grpc",
			expectedType: v1alpha1.ClusterType_CLUSTER_STATIC,
			description:  "SDS gRPC should be static",
		},
		{
			name:         "xds-grpc static cluster",
			clusterName:  "xds-grpc",
			expectedType: v1alpha1.ClusterType_CLUSTER_STATIC,
			description:  "XDS gRPC should be static",
		},
		{
			name:         "zipkin static cluster",
			clusterName:  "zipkin",
			expectedType: v1alpha1.ClusterType_CLUSTER_STATIC,
			description:  "Zipkin tracing should be static",
		},
		{
			name:         "jaeger static cluster",
			clusterName:  "jaeger",
			expectedType: v1alpha1.ClusterType_CLUSTER_STATIC,
			description:  "Jaeger tracing should be static",
		},
		{
			name:         "envoy access log service static",
			clusterName:  "envoy_accesslog_service",
			expectedType: v1alpha1.ClusterType_CLUSTER_STATIC,
			description:  "Envoy access log service should be static",
		},
		{
			name:         "IP address based cluster DNS",
			clusterName:  "192.168.1.100",
			expectedType: v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description:  "IP-based cluster should use strict DNS",
		},
		{
			name:         "localhost cluster DNS",
			clusterName:  "127.0.0.1:8080",
			expectedType: v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description:  "Localhost cluster should use strict DNS",
		},
		{
			name:         "simple domain DNS",
			clusterName:  "api.company.io",
			expectedType: v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
			description:  "Simple external domain should use DNS",
		},
		{
			name:         "unknown cluster defaults to EDS",
			clusterName:  "unknown-cluster-format",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Unknown format should default to EDS",
		},
		{
			name:         "empty cluster name unknown",
			clusterName:  "",
			expectedType: v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
			description:  "Empty cluster name should be unknown",
		},
		{
			name:         "service with version subset",
			clusterName:  "outbound|8080|v2|backend.production.svc.cluster.local",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Versioned service should use EDS",
		},
		{
			name:         "service in different namespace",
			clusterName:  "outbound|9090||metrics.monitoring.svc.cluster.local",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Cross-namespace service should use EDS",
		},
		{
			name:         "headless service",
			clusterName:  "outbound|5432||postgres.database.svc.cluster.local",
			expectedType: v1alpha1.ClusterType_CLUSTER_EDS,
			description:  "Headless service should use EDS",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			endpoint := &v1alpha1.EndpointSummary{
				ClusterName: test.clusterName,
			}

			err := enrichFunc(endpoint)
			require.NoError(t, err, test.description)
			assert.Equal(t, test.expectedType, endpoint.ClusterType, test.description)
		})
	}

	t.Run("handles nil endpoint", func(t *testing.T) {
		err := enrichFunc(nil)
		assert.NoError(t, err)
	})
}

func TestParseClusterName(t *testing.T) {
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
			description:         "Full Istio cluster format",
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
			name:                "external service",
			clusterName:         "outbound|443||api.external.com",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        443,
			expectedSubset:      "",
			expectedServiceFqdn: "api.external.com",
			description:         "External service cluster",
		},
		{
			name:                "static cluster not parsed",
			clusterName:         "prometheus_stats",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Static clusters should not be parsed",
		},
		{
			name:                "malformed cluster",
			clusterName:         "outbound|8080",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Insufficient parts should not be parsed",
		},
		{
			name:                "too many parts",
			clusterName:         "outbound|8080||service.ns.svc.cluster.local|extra",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Too many parts should not be parsed",
		},
		{
			name:                "empty cluster name",
			clusterName:         "",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Empty cluster name should not be parsed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			endpoint := &v1alpha1.EndpointSummary{
				ClusterName: test.clusterName,
			}

			ParseClusterName(test.clusterName, endpoint)
			assert.Equal(t, test.expectedDirection, endpoint.Direction, test.description)
			assert.Equal(t, test.expectedPort, endpoint.Port, test.description)
			assert.Equal(t, test.expectedSubset, endpoint.Subset, test.description)
			assert.Equal(t, test.expectedServiceFqdn, endpoint.ServiceFqdn, test.description)
		})
	}
}
