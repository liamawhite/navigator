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

package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestParseClustersOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected number of cluster endpoints
		wantErr  bool
	}{
		{
			name:     "empty input",
			input:    "",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			input:    "invalid json",
			expected: 0,
			wantErr:  true,
		},
		{
			name: "simple cluster with one endpoint",
			input: `{
				"cluster_statuses": [
					{
						"name": "test-cluster",
						"host_statuses": [
							{
								"address": {
									"socket_address": {
										"address": "10.244.1.5",
										"port_value": 8080
									}
								},
								"health_status": {
									"eds_health_status": "HEALTHY"
								},
								"weight": 1,
								"locality": {}
							}
						]
					}
				]
			}`,
			expected: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseClustersOutput(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, tt.expected)

			if tt.expected > 0 && len(result) > 0 {
				cluster := result[0]
				assert.NotEmpty(t, cluster.ClusterName)
				assert.NotEmpty(t, cluster.Endpoints)

				if len(cluster.Endpoints) > 0 {
					endpoint := cluster.Endpoints[0]
					assert.NotEmpty(t, endpoint.Address)
					assert.NotZero(t, endpoint.Port)
					assert.NotEmpty(t, endpoint.HostIdentifier)
					assert.NotEmpty(t, endpoint.Health)
				}
			}
		})
	}
}

func TestParseClustersOutputWithRealData(t *testing.T) {
	// Read the real clusters data from testdata
	testDataPath := filepath.Join("..", "..", "..", "pkg", "envoy", "configdump", "testdata", "envoy_clusters.json")
	// #nosec G304 - This is a hardcoded test file path, safe to use
	data, err := os.ReadFile(testDataPath)
	if os.IsNotExist(err) {
		t.Skip("envoy_clusters.json testdata file not found")
		return
	}
	require.NoError(t, err)

	result, err := ParseClustersOutput(string(data))
	require.NoError(t, err)

	// Should have at least one cluster with endpoints
	assert.NotEmpty(t, result)

	foundEndpoints := false
	for _, cluster := range result {
		assert.NotEmpty(t, cluster.ClusterName)

		if len(cluster.Endpoints) > 0 {
			foundEndpoints = true

			for _, endpoint := range cluster.Endpoints {
				// Validate endpoint structure
				assert.NotEmpty(t, endpoint.Address, "endpoint address should not be empty")
				assert.NotEmpty(t, endpoint.HostIdentifier, "host identifier should not be empty")
				assert.NotNil(t, endpoint.Metadata, "metadata should be initialized")

				// Check if this is a Unix domain socket (pipe) or TCP socket
				if strings.HasPrefix(endpoint.HostIdentifier, "unix://") {
					// Unix domain socket - should have zero port
					assert.Equal(t, uint32(0), endpoint.Port, "pipe endpoint should have zero port")
					assert.Contains(t, endpoint.HostIdentifier, "unix://", "pipe host identifier should start with unix://")
				} else {
					// TCP socket - should have non-zero port
					assert.NotZero(t, endpoint.Port, "TCP endpoint port should not be zero")
					assert.Contains(t, endpoint.HostIdentifier, ":", "TCP host identifier should contain port")
				}

				// Health status should be set for live endpoints
				if endpoint.Health != "" {
					assert.Contains(t, []string{"HEALTHY", "UNHEALTHY", "DRAINING", "TIMEOUT", "DEGRADED"},
						endpoint.Health, "health status should be valid")
				}

				// Weight should be reasonable
				if endpoint.Weight > 0 {
					assert.LessOrEqual(t, endpoint.Weight, uint32(1000), "weight should be reasonable")
				}
			}
		}
	}

	assert.True(t, foundEndpoints, "should find at least one cluster with endpoints")
}

func TestMergeClusterEndpointsWithConfig(t *testing.T) {
	// Test merging functionality
	staticEndpoints := []*v1alpha1.EndpointSummary{
		{
			ClusterName: "static-cluster",
			Endpoints: []*v1alpha1.EndpointInfo{
				{
					Address: "1.2.3.4",
					Port:    8080,
					Health:  "UNKNOWN",
				},
			},
		},
	}

	liveEndpoints := []*ClusterEndpointInfo{
		{
			ClusterName: "static-cluster",
			Endpoints: []*v1alpha1.EndpointInfo{
				{
					Address: "1.2.3.4",
					Port:    8080,
					Health:  "HEALTHY",
					Weight:  100,
				},
			},
		},
		{
			ClusterName: "new-cluster",
			Endpoints: []*v1alpha1.EndpointInfo{
				{
					Address: "5.6.7.8",
					Port:    9090,
					Health:  "HEALTHY",
					Weight:  50,
				},
			},
		},
	}

	result := MergeClusterEndpointsWithConfig(staticEndpoints, liveEndpoints)

	// Should have both clusters
	assert.Len(t, result, 2)

	// Find the merged static cluster
	var mergedStatic *v1alpha1.EndpointSummary
	var newCluster *v1alpha1.EndpointSummary

	for _, cluster := range result {
		switch cluster.ClusterName {
		case "static-cluster":
			mergedStatic = cluster
		case "new-cluster":
			newCluster = cluster
		}
	}

	require.NotNil(t, mergedStatic, "merged static cluster should exist")
	require.NotNil(t, newCluster, "new cluster should exist")

	// Static cluster should use live endpoint data
	assert.Equal(t, "HEALTHY", mergedStatic.Endpoints[0].Health)
	assert.Equal(t, uint32(100), mergedStatic.Endpoints[0].Weight)

	// New cluster should be present
	assert.Equal(t, "5.6.7.8", newCluster.Endpoints[0].Address)
	assert.Equal(t, uint32(9090), newCluster.Endpoints[0].Port)
}

func TestParseClusterName(t *testing.T) {
	tests := []struct {
		name              string
		clusterName       string
		expectedDirection v1alpha1.ClusterDirection
		expectedPort      uint32
		expectedSubset    string
		expectedFqdn      string
	}{
		{
			name:              "outbound with service",
			clusterName:       "outbound|8080||backend.demo.svc.cluster.local",
			expectedDirection: v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:      8080,
			expectedSubset:    "",
			expectedFqdn:      "backend.demo.svc.cluster.local",
		},
		{
			name:              "inbound",
			clusterName:       "inbound|8080||",
			expectedDirection: v1alpha1.ClusterDirection_INBOUND,
			expectedPort:      8080,
			expectedSubset:    "",
			expectedFqdn:      "",
		},
		{
			name:              "outbound with subset",
			clusterName:       "outbound|443|v1|api.example.com",
			expectedDirection: v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:      443,
			expectedSubset:    "v1",
			expectedFqdn:      "api.example.com",
		},
		{
			name:              "non-standard cluster",
			clusterName:       "agent",
			expectedDirection: v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:      0,
			expectedSubset:    "",
			expectedFqdn:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			direction, port, subset, fqdn := parseClusterName(tt.clusterName)

			assert.Equal(t, tt.expectedDirection, direction, "direction should match")
			assert.Equal(t, tt.expectedPort, port, "port should match")
			assert.Equal(t, tt.expectedSubset, subset, "subset should match")
			assert.Equal(t, tt.expectedFqdn, fqdn, "FQDN should match")
		})
	}
}
