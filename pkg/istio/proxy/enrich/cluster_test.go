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

func TestEnrichClusterNameComponents(t *testing.T) {
	enrichFunc := enrichClusterNameComponents()

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
			name:                "outbound cluster with subset",
			clusterName:         "outbound|8080|v1|backend.demo.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "v1",
			expectedServiceFqdn: "backend.demo.svc.cluster.local",
			description:         "Standard outbound cluster with version subset",
		},
		{
			name:                "outbound cluster without subset",
			clusterName:         "outbound|8080||backend.demo.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "backend.demo.svc.cluster.local",
			description:         "Outbound cluster with empty subset",
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
			name:                "inbound cluster with non-standard port",
			clusterName:         "inbound|9090||",
			expectedDirection:   v1alpha1.ClusterDirection_INBOUND,
			expectedPort:        9090,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Inbound cluster with custom port",
		},
		{
			name:                "outbound HTTPS cluster",
			clusterName:         "outbound|443||api.example.com",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        443,
			expectedSubset:      "",
			expectedServiceFqdn: "api.example.com",
			description:         "External HTTPS service cluster",
		},
		{
			name:                "outbound DNS cluster",
			clusterName:         "outbound|53||kube-dns.kube-system.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        53,
			expectedSubset:      "",
			expectedServiceFqdn: "kube-dns.kube-system.svc.cluster.local",
			description:         "DNS service cluster",
		},
		{
			name:                "outbound with complex subset",
			clusterName:         "outbound|8080|canary-v2|my-service.production.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "canary-v2",
			expectedServiceFqdn: "my-service.production.svc.cluster.local",
			description:         "Complex subset naming with canary deployment",
		},
		{
			name:                "outbound with numeric subset",
			clusterName:         "outbound|8080|v123|service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "v123",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Numeric version subset",
		},
		{
			name:                "outbound with high port number",
			clusterName:         "outbound|65535||test-service.test.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        65535,
			expectedSubset:      "",
			expectedServiceFqdn: "test-service.test.svc.cluster.local",
			description:         "Maximum valid port number",
		},
		{
			name:                "outbound external service",
			clusterName:         "outbound|443|v1|external-api.company.com",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        443,
			expectedSubset:      "v1",
			expectedServiceFqdn: "external-api.company.com",
			description:         "External service with subset",
		},
		{
			name:                "case insensitive direction - uppercase",
			clusterName:         "OUTBOUND|8080||service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Case insensitive direction parsing",
		},
		{
			name:                "case insensitive direction - mixed case",
			clusterName:         "InBound|8080||",
			expectedDirection:   v1alpha1.ClusterDirection_INBOUND,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Mixed case direction parsing",
		},
		{
			name:                "non-istio cluster name",
			clusterName:         "prometheus_stats",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Static cluster should not be parsed",
		},
		{
			name:                "cluster with two parts",
			clusterName:         "outbound|8080",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Two-part cluster should parse direction and port",
		},
		{
			name:                "cluster with extra parts",
			clusterName:         "outbound|8080||service.ns.svc.cluster.local|extra",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Extra parts should be ignored",
		},
		{
			name:                "invalid port number",
			clusterName:         "outbound|invalid||service.ns.svc.cluster.local",
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
			name:                "unknown direction",
			clusterName:         "unknown|8080||service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        8080,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Unknown direction should be unspecified",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cluster := &v1alpha1.ClusterSummary{
				Name: test.clusterName,
			}

			err := enrichFunc(cluster)
			require.NoError(t, err, test.description)
			assert.Equal(t, test.expectedDirection, cluster.Direction, test.description)
			assert.Equal(t, test.expectedPort, cluster.Port, test.description)
			assert.Equal(t, test.expectedSubset, cluster.Subset, test.description)
			assert.Equal(t, test.expectedServiceFqdn, cluster.ServiceFqdn, test.description)
		})
	}

	t.Run("handles nil cluster", func(t *testing.T) {
		err := enrichFunc(nil)
		assert.NoError(t, err)
	})
}

func TestParseClusterNameComponents(t *testing.T) {
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
			clusterName:         "outbound|8080|v2|backend.production.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "v2",
			expectedServiceFqdn: "backend.production.svc.cluster.local",
			description:         "All components present",
		},
		{
			name:                "minimal inbound cluster",
			clusterName:         "inbound",
			expectedDirection:   v1alpha1.ClusterDirection_INBOUND,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Only direction component",
		},
		{
			name:                "direction and port only",
			clusterName:         "outbound|9090",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        9090,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Direction and port components only",
		},
		{
			name:                "direction, port, and subset",
			clusterName:         "outbound|8080|canary",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "canary",
			expectedServiceFqdn: "",
			description:         "First three components only",
		},
		{
			name:                "port with zero value",
			clusterName:         "outbound|0||service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Port zero should be preserved",
		},
		{
			name:                "large port number",
			clusterName:         "outbound|32768||service.ns.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        32768,
			expectedSubset:      "",
			expectedServiceFqdn: "service.ns.svc.cluster.local",
			description:         "Large valid port number",
		},
		{
			name:                "subset with special characters",
			clusterName:         "outbound|8080|v1-beta.2|service-name.namespace.svc.cluster.local",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        8080,
			expectedSubset:      "v1-beta.2",
			expectedServiceFqdn: "service-name.namespace.svc.cluster.local",
			description:         "Subset with hyphens and dots",
		},
		{
			name:                "fqdn with multiple subdomains",
			clusterName:         "outbound|443||api.v1.backend.production.example.com",
			expectedDirection:   v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:        443,
			expectedSubset:      "",
			expectedServiceFqdn: "api.v1.backend.production.example.com",
			description:         "Complex FQDN with multiple levels",
		},
		{
			name:                "empty string",
			clusterName:         "",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Empty cluster name",
		},
		{
			name:                "only pipe separators",
			clusterName:         "|||",
			expectedDirection:   v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:        0,
			expectedSubset:      "",
			expectedServiceFqdn: "",
			description:         "Only separators without content",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			direction, port, subset, serviceFqdn := ParseClusterNameComponents(test.clusterName)
			assert.Equal(t, test.expectedDirection, direction, test.description)
			assert.Equal(t, test.expectedPort, port, test.description)
			assert.Equal(t, test.expectedSubset, subset, test.description)
			assert.Equal(t, test.expectedServiceFqdn, serviceFqdn, test.description)
		})
	}
}
