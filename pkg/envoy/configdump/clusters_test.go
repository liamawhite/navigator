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

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

func TestParser_parseClusterName(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		expected    struct {
			direction   v1alpha1.ClusterDirection
			port        uint32
			subset      string
			serviceFqdn string
		}
	}{
		{
			name:        "valid outbound cluster with service and port",
			clusterName: "outbound|80||productpage.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        80,
				subset:      "",
				serviceFqdn: "productpage.default.svc.cluster.local",
			},
		},
		{
			name:        "valid inbound cluster with service FQDN",
			clusterName: "inbound|9080||details.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_INBOUND,
				port:        9080,
				subset:      "",
				serviceFqdn: "details.default.svc.cluster.local",
			},
		},
		{
			name:        "valid inbound cluster without service FQDN",
			clusterName: "inbound|8080||",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_INBOUND,
				port:        8080,
				subset:      "",
				serviceFqdn: "",
			},
		},
		{
			name:        "outbound cluster with subset",
			clusterName: "outbound|9080|v1|reviews.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        9080,
				subset:      "v1",
				serviceFqdn: "reviews.default.svc.cluster.local",
			},
		},
		{
			name:        "outbound cluster with complex subset",
			clusterName: "outbound|443|tls|istio-egressgateway.istio-system.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        443,
				subset:      "tls",
				serviceFqdn: "istio-egressgateway.istio-system.svc.cluster.local",
			},
		},
		{
			name:        "outbound cluster with external service",
			clusterName: "outbound|443||httpbin.org",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        443,
				subset:      "",
				serviceFqdn: "httpbin.org",
			},
		},
		{
			name:        "cluster name with different case direction",
			clusterName: "OUTBOUND|8080||nginx.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        8080,
				subset:      "",
				serviceFqdn: "nginx.default.svc.cluster.local",
			},
		},
		{
			name:        "invalid direction",
			clusterName: "unknown|80||service.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        80,
				subset:      "",
				serviceFqdn: "service.default.svc.cluster.local",
			},
		},
		{
			name:        "invalid port number",
			clusterName: "outbound|notaport||service.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        0, // Should default to 0 when parsing fails
				subset:      "",
				serviceFqdn: "service.default.svc.cluster.local",
			},
		},
		{
			name:        "empty port",
			clusterName: "outbound|||service.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        0,
				subset:      "",
				serviceFqdn: "service.default.svc.cluster.local",
			},
		},
		{
			name:        "cluster name with too few parts",
			clusterName: "outbound|80",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        0,
				subset:      "",
				serviceFqdn: "",
			},
		},
		{
			name:        "cluster name with too many parts",
			clusterName: "outbound|80||service.default.svc.cluster.local|extra",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        0,
				subset:      "",
				serviceFqdn: "",
			},
		},
		{
			name:        "empty cluster name",
			clusterName: "",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        0,
				subset:      "",
				serviceFqdn: "",
			},
		},
		{
			name:        "cluster name without pipes",
			clusterName: "BlackHoleCluster",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        0,
				subset:      "",
				serviceFqdn: "",
			},
		},
		{
			name:        "cluster name with large port number",
			clusterName: "outbound|65535||service.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        65535,
				subset:      "",
				serviceFqdn: "service.default.svc.cluster.local",
			},
		},
		{
			name:        "cluster name with port number too large for uint32",
			clusterName: "outbound|4294967296||service.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        0, // Should default to 0 when parsing fails
				subset:      "",
				serviceFqdn: "service.default.svc.cluster.local",
			},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := &v1alpha1.ClusterSummary{}
			parser.parseClusterName(tt.clusterName, summary)

			assert.Equal(t, tt.expected.direction, summary.Direction, "direction should match")
			assert.Equal(t, tt.expected.port, summary.Port, "port should match")
			assert.Equal(t, tt.expected.subset, summary.Subset, "subset should match")
			assert.Equal(t, tt.expected.serviceFqdn, summary.ServiceFqdn, "service FQDN should match")
		})
	}
}

func TestParser_summarizeCluster_WithNameParsing(t *testing.T) {
	tests := []struct {
		name           string
		clusterName    string
		clusterType    clusterv3.Cluster_DiscoveryType
		lbPolicy       clusterv3.Cluster_LbPolicy
		expectedFields struct {
			name        string
			direction   v1alpha1.ClusterDirection
			port        uint32
			subset      string
			serviceFqdn string
			clusterType string
		}
	}{
		{
			name:        "EDS cluster with outbound direction",
			clusterName: "outbound|9080|v1|reviews.default.svc.cluster.local",
			clusterType: clusterv3.Cluster_EDS,
			lbPolicy:    clusterv3.Cluster_ROUND_ROBIN,
			expectedFields: struct {
				name        string
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
				clusterType string
			}{
				name:        "outbound|9080|v1|reviews.default.svc.cluster.local",
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        9080,
				subset:      "v1",
				serviceFqdn: "reviews.default.svc.cluster.local",
				clusterType: "EDS",
			},
		},
		{
			name:        "STATIC cluster with inbound direction and service FQDN",
			clusterName: "inbound|8080||app.default.svc.cluster.local",
			clusterType: clusterv3.Cluster_STATIC,
			lbPolicy:    clusterv3.Cluster_LEAST_REQUEST,
			expectedFields: struct {
				name        string
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
				clusterType string
			}{
				name:        "inbound|8080||app.default.svc.cluster.local",
				direction:   v1alpha1.ClusterDirection_INBOUND,
				port:        8080,
				subset:      "",
				serviceFqdn: "app.default.svc.cluster.local",
				clusterType: "STATIC",
			},
		},
		{
			name:        "EDS cluster with inbound direction without service FQDN",
			clusterName: "inbound|8080||",
			clusterType: clusterv3.Cluster_EDS,
			lbPolicy:    clusterv3.Cluster_ROUND_ROBIN,
			expectedFields: struct {
				name        string
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
				clusterType string
			}{
				name:        "inbound|8080||",
				direction:   v1alpha1.ClusterDirection_INBOUND,
				port:        8080,
				subset:      "",
				serviceFqdn: "",
				clusterType: "EDS",
			},
		},
		{
			name:        "non-parseable cluster name",
			clusterName: "BlackHoleCluster",
			clusterType: clusterv3.Cluster_STRICT_DNS,
			lbPolicy:    clusterv3.Cluster_RANDOM,
			expectedFields: struct {
				name        string
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
				clusterType string
			}{
				name:        "BlackHoleCluster",
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        0,
				subset:      "",
				serviceFqdn: "",
				clusterType: "STRICT_DNS",
			},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock cluster
			cluster := &clusterv3.Cluster{
				Name:                 tt.clusterName,
				ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: tt.clusterType},
				LbPolicy:             tt.lbPolicy,
			}

			// Create mock parsed config with empty RawClusters map
			parsed := &ParsedConfig{
				RawClusters: make(map[string]string),
			}

			summary := parser.summarizeCluster(cluster, parsed)

			require.NotNil(t, summary, "summary should not be nil")

			// Test all fields
			assert.Equal(t, tt.expectedFields.name, summary.Name, "name should match")
			assert.Equal(t, tt.expectedFields.direction, summary.Direction, "direction should match")
			assert.Equal(t, tt.expectedFields.port, summary.Port, "port should match")
			assert.Equal(t, tt.expectedFields.subset, summary.Subset, "subset should match")
			assert.Equal(t, tt.expectedFields.serviceFqdn, summary.ServiceFqdn, "service FQDN should match")
			assert.Equal(t, tt.expectedFields.clusterType, summary.Type, "cluster type should match")
		})
	}
}

func TestParser_summarizeCluster_NilCluster(t *testing.T) {
	parser := NewParser()
	parsed := &ParsedConfig{
		RawClusters: make(map[string]string),
	}

	summary := parser.summarizeCluster(nil, parsed)
	assert.Nil(t, summary, "summary should be nil for nil cluster")
}

func TestParser_summarizeCluster_WithRawConfig(t *testing.T) {
	parser := NewParser()
	clusterName := "outbound|80||httpbin.default.svc.cluster.local"
	rawConfig := `{"name": "outbound|80||httpbin.default.svc.cluster.local", "type": "EDS"}`

	cluster := &clusterv3.Cluster{
		Name:                 clusterName,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_EDS},
		LbPolicy:             clusterv3.Cluster_ROUND_ROBIN,
	}

	parsed := &ParsedConfig{
		RawClusters: map[string]string{
			clusterName: rawConfig,
		},
	}

	summary := parser.summarizeCluster(cluster, parsed)

	require.NotNil(t, summary)
	assert.Equal(t, rawConfig, summary.RawConfig, "raw config should be populated")
	assert.Equal(t, v1alpha1.ClusterDirection_OUTBOUND, summary.Direction)
	assert.Equal(t, uint32(80), summary.Port)
	assert.Equal(t, "httpbin.default.svc.cluster.local", summary.ServiceFqdn)
}

// Benchmark for performance testing
func BenchmarkParser_parseClusterName(b *testing.B) {
	parser := NewParser()
	clusterName := "outbound|9080|v1|reviews.default.svc.cluster.local"
	summary := &v1alpha1.ClusterSummary{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.parseClusterName(clusterName, summary)
	}
}

func BenchmarkParser_summarizeCluster(b *testing.B) {
	parser := NewParser()
	cluster := &clusterv3.Cluster{
		Name:                 "outbound|9080|v1|reviews.default.svc.cluster.local",
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_EDS},
		LbPolicy:             clusterv3.Cluster_ROUND_ROBIN,
	}
	parsed := &ParsedConfig{
		RawClusters: make(map[string]string),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.summarizeCluster(cluster, parsed)
	}
}
