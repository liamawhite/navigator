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

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
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
			name:        "generic parsing uses cluster name as FQDN",
			clusterName: "outbound|80||productpage.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        0,
				subset:      "",
				serviceFqdn: "outbound|80||productpage.default.svc.cluster.local",
			},
		},
		{
			name:        "simple cluster name",
			clusterName: "backend-service",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
				port:        0,
				subset:      "",
				serviceFqdn: "backend-service",
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
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := &v1alpha1.ClusterSummary{
				Name: tt.clusterName,
			}

			parser.parseClusterName(tt.clusterName, summary)

			assert.Equal(t, tt.expected.direction, summary.Direction, "direction should match")
			assert.Equal(t, tt.expected.port, summary.Port, "port should match")
			assert.Equal(t, tt.expected.subset, summary.Subset, "subset should match")
			assert.Equal(t, tt.expected.serviceFqdn, summary.ServiceFqdn, "service FQDN should match")
		})
	}
}

func TestParser_summarizeCluster(t *testing.T) {
	parser := NewParser()

	// Create a test cluster
	cluster := &clusterv3.Cluster{
		Name:                 "test-cluster",
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_EDS},
		LbPolicy:             clusterv3.Cluster_ROUND_ROBIN,
		ConnectTimeout:       nil, // Will be nil in test
	}

	parsed := &ParsedConfig{
		RawClusters: map[string]string{
			"test-cluster": `{"name": "test-cluster", "type": "EDS"}`,
		},
	}

	summary := parser.summarizeCluster(cluster, parsed)

	assert.NotNil(t, summary)
	assert.Equal(t, "test-cluster", summary.Name)
	assert.Equal(t, "EDS", summary.Type)
	assert.Equal(t, "ROUND_ROBIN", summary.LoadBalancingPolicy)
	assert.Equal(t, v1alpha1.ClusterDirection_UNSPECIFIED, summary.Direction)
	assert.Equal(t, uint32(0), summary.Port)
	assert.Equal(t, "", summary.Subset)
	assert.Equal(t, "test-cluster", summary.ServiceFqdn)    // Generic behavior: cluster name as FQDN
	assert.Contains(t, summary.RawConfig, `"test-cluster"`) // Should contain raw JSON
}
