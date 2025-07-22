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

package clusters

import (
	"os"
	"path/filepath"
	"testing"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// loadTestData reads test data from the testdata directory
func loadTestData(t *testing.T, filename string) string {
	t.Helper()
	// Only allow basic filenames to prevent path traversal
	if filepath.Base(filename) != filename {
		t.Fatalf("invalid filename: %s", filename)
	}
	data, err := os.ReadFile(filepath.Join("testdata", filename)) // #nosec G304 - filename is validated above
	require.NoError(t, err, "failed to read test data file: %s", filename)
	return string(data)
}

func TestParser_ParseJSON(t *testing.T) {
	tests := []struct {
		name         string
		testDataFile string
		expected     []*v1alpha1.EndpointSummary
		wantErr      bool
	}{
		{
			name:         "valid clusters response with endpoints",
			testDataFile: "generic_clusters_with_endpoints.json",
			expected: []*v1alpha1.EndpointSummary{
				{
					ClusterName: "example_service_cluster",
					ClusterType: v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
					Direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
					Port:        0,
					Subset:      "",
					ServiceFqdn: "example_service_cluster", // Uses cluster name as-is
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:        "192.168.1.10",
							Port:           8080,
							HostIdentifier: "192.168.1.10:8080",
							Health:         "HEALTHY",
							Priority:       0,
							Weight:         1,
							Metadata:       map[string]string{},
							AddressType:    v1alpha1.AddressType_SOCKET_ADDRESS,
							Locality:       nil,
						},
						{
							Address:        "192.168.1.11",
							Port:           8080,
							HostIdentifier: "192.168.1.11:8080",
							Health:         "UNHEALTHY",
							Priority:       0,
							Weight:         1,
							Metadata:       map[string]string{},
							AddressType:    v1alpha1.AddressType_SOCKET_ADDRESS,
							Locality:       nil,
						},
					},
				},
				{
					ClusterName: "api_service",
					ClusterType: v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
					Direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
					Port:        0,
					Subset:      "",
					ServiceFqdn: "api_service",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:        "203.0.113.1",
							Port:           443,
							HostIdentifier: "203.0.113.1:443",
							Health:         "HEALTHY",
							Priority:       0,
							Weight:         1,
							Metadata:       map[string]string{},
							AddressType:    v1alpha1.AddressType_SOCKET_ADDRESS,
							Locality:       nil,
						},
					},
				},
			},
		},
		{
			name:         "cluster with unix socket address",
			testDataFile: "unix_socket_cluster.json",
			expected: []*v1alpha1.EndpointSummary{
				{
					ClusterName: "uds_cluster",
					ClusterType: v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
					Direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
					Port:        0,
					Subset:      "",
					ServiceFqdn: "uds_cluster", // Uses cluster name as-is
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:        "unix:///tmp/socket",
							Port:           0,
							HostIdentifier: "unix:///tmp/socket",
							Health:         "HEALTHY",
							Priority:       0,
							Weight:         1,
							Metadata:       map[string]string{},
							AddressType:    v1alpha1.AddressType_PIPE_ADDRESS,
							Locality:       nil,
						},
					},
				},
			},
		},
		{
			name:         "cluster with endpoints having different priorities",
			testDataFile: "cluster_with_priorities.json",
			expected: []*v1alpha1.EndpointSummary{
				{
					ClusterName: "priority_service_cluster",
					ClusterType: v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
					Direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
					Port:        0,
					Subset:      "",
					ServiceFqdn: "priority_service_cluster",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:        "10.244.0.20",
							Port:           80,
							HostIdentifier: "10.244.0.20:80",
							Health:         "HEALTHY",
							Priority:       0,
							Weight:         1,
							Metadata:       map[string]string{},
							AddressType:    v1alpha1.AddressType_SOCKET_ADDRESS,
						},
						{
							Address:        "10.244.0.21",
							Port:           80,
							HostIdentifier: "10.244.0.21:80",
							Health:         "HEALTHY",
							Priority:       1,
							Weight:         1,
							Metadata:       map[string]string{},
							AddressType:    v1alpha1.AddressType_SOCKET_ADDRESS,
						},
						{
							Address:        "10.244.0.22",
							Port:           80,
							HostIdentifier: "10.244.0.22:80",
							Health:         "HEALTHY",
							Priority:       2,
							Weight:         1,
							Metadata:       map[string]string{},
							AddressType:    v1alpha1.AddressType_SOCKET_ADDRESS,
						},
					},
				},
			},
		},
		{
			name:         "empty clusters response",
			testDataFile: "empty_clusters_response.json",
			expected:     []*v1alpha1.EndpointSummary{},
		},
		{
			name:         "cluster with no endpoints",
			testDataFile: "cluster_with_no_endpoints.json",
			expected:     []*v1alpha1.EndpointSummary{},
		},
		{
			name:         "invalid JSON",
			testDataFile: "invalid_json.json",
			wantErr:      true,
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input string
			if tt.testDataFile != "" {
				input = loadTestData(t, tt.testDataFile)
			}

			result, err := parser.ParseJSON(input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result), "number of endpoint summaries should match")

			for i, expected := range tt.expected {
				if i < len(result) {
					actual := result[i]
					assert.Equal(t, expected.ClusterName, actual.ClusterName, "cluster name should match")
					assert.Equal(t, len(expected.Endpoints), len(actual.Endpoints), "number of endpoints should match")

					for j, expectedEndpoint := range expected.Endpoints {
						if j < len(actual.Endpoints) {
							actualEndpoint := actual.Endpoints[j]
							assert.Equal(t, expectedEndpoint.Address, actualEndpoint.Address, "endpoint %d address should match", j)
							assert.Equal(t, expectedEndpoint.Port, actualEndpoint.Port, "endpoint %d port should match", j)
							assert.Equal(t, expectedEndpoint.HostIdentifier, actualEndpoint.HostIdentifier, "endpoint %d host identifier should match", j)
							assert.Equal(t, expectedEndpoint.Health, actualEndpoint.Health, "endpoint %d health should match", j)
							assert.Equal(t, expectedEndpoint.Priority, actualEndpoint.Priority, "endpoint %d priority should match", j)
							assert.Equal(t, expectedEndpoint.Weight, actualEndpoint.Weight, "endpoint %d weight should match", j)
							assert.NotNil(t, actualEndpoint.Metadata, "endpoint %d metadata should not be nil", j)
						}
					}
				}
			}
		})
	}
}

func TestParser_convertHostToEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		host     *admin.HostStatus
		expected *v1alpha1.EndpointInfo
	}{
		{
			name: "healthy socket address endpoint",
			host: &admin.HostStatus{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "192.168.1.1",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 8080,
							},
						},
					},
				},
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_HEALTHY,
				},
			},
			expected: &v1alpha1.EndpointInfo{
				Address:        "192.168.1.1",
				Port:           8080,
				HostIdentifier: "192.168.1.1:8080",
				Health:         "HEALTHY",
				Priority:       0,
				Weight:         1,
				Metadata:       map[string]string{},
				Locality:       nil,
			},
		},
		{
			name: "unhealthy endpoint with outlier check",
			host: &admin.HostStatus{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "10.0.0.1",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 3000,
							},
						},
					},
				},
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_UNHEALTHY,
				},
			},
			expected: &v1alpha1.EndpointInfo{
				Address:        "10.0.0.1",
				Port:           3000,
				HostIdentifier: "10.0.0.1:3000",
				Health:         "UNHEALTHY",
				Priority:       0,
				Weight:         1,
				Metadata:       map[string]string{},
				Locality:       nil,
			},
		},
		{
			name: "pipe address endpoint",
			host: &admin.HostStatus{
				Address: &core.Address{
					Address: &core.Address_Pipe{
						Pipe: &core.Pipe{
							Path: "/var/run/service.sock",
						},
					},
				},
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_HEALTHY,
				},
			},
			expected: &v1alpha1.EndpointInfo{
				Address:        "unix:///var/run/service.sock",
				Port:           0,
				HostIdentifier: "unix:///var/run/service.sock",
				Health:         "HEALTHY",
				Priority:       0,
				Weight:         1,
				Metadata:       map[string]string{},
				Locality:       nil,
			},
		},
		{
			name:     "nil host",
			host:     nil,
			expected: nil,
		},
		{
			name: "endpoint with no health status",
			host: &admin.HostStatus{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "127.0.0.1",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 9000,
							},
						},
					},
				},
				HealthStatus: nil,
			},
			expected: &v1alpha1.EndpointInfo{
				Address:        "127.0.0.1",
				Port:           9000,
				HostIdentifier: "127.0.0.1:9000",
				Health:         "UNKNOWN",
				Priority:       0,
				Weight:         1,
				Metadata:       map[string]string{},
				Locality:       nil,
			},
		},
		{
			name: "endpoint with weight, priority, and locality information",
			host: &admin.HostStatus{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "10.1.2.3",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 8080,
							},
						},
					},
				},
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_HEALTHY,
				},
				Weight:   100,
				Priority: 1,
				Locality: &core.Locality{
					Region: "us-west-2",
					Zone:   "us-west-2a",
				},
			},
			expected: &v1alpha1.EndpointInfo{
				Address:        "10.1.2.3",
				Port:           8080,
				HostIdentifier: "10.1.2.3:8080",
				Health:         "HEALTHY",
				Priority:       1,
				Weight:         100,
				Metadata:       map[string]string{},
				Locality: &v1alpha1.LocalityInfo{
					Region: "us-west-2",
					Zone:   "us-west-2a",
				},
			},
		},
		{
			name: "endpoint with partial locality information",
			host: &admin.HostStatus{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "192.168.1.100",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 3000,
							},
						},
					},
				},
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_HEALTHY,
				},
				Weight: 75,
				Locality: &core.Locality{
					Region: "eu-central-1",
					Zone:   "", // Empty zone
				},
			},
			expected: &v1alpha1.EndpointInfo{
				Address:        "192.168.1.100",
				Port:           3000,
				HostIdentifier: "192.168.1.100:3000",
				Health:         "HEALTHY",
				Priority:       0,
				Weight:         75,
				Metadata:       map[string]string{},
				Locality: &v1alpha1.LocalityInfo{
					Region: "eu-central-1",
					Zone:   "",
				},
			},
		},
		{
			name: "endpoint with priority 2",
			host: &admin.HostStatus{
				Address: &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "10.0.0.100",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 8080,
							},
						},
					},
				},
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_HEALTHY,
				},
				Priority: 2,
			},
			expected: &v1alpha1.EndpointInfo{
				Address:        "10.0.0.100",
				Port:           8080,
				HostIdentifier: "10.0.0.100:8080",
				Health:         "HEALTHY",
				Priority:       2,
				Weight:         1,
				Metadata:       map[string]string{},
				Locality:       nil,
			},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.convertHostToEndpoint(tt.host)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expected.Address, result.Address)
			assert.Equal(t, tt.expected.Port, result.Port)
			assert.Equal(t, tt.expected.HostIdentifier, result.HostIdentifier)
			assert.Equal(t, tt.expected.Health, result.Health)
			assert.Equal(t, tt.expected.Priority, result.Priority)
			assert.Equal(t, tt.expected.Weight, result.Weight)
			assert.Equal(t, tt.expected.Locality, result.Locality)
			assert.NotNil(t, result.Metadata)
		})
	}
}

func TestParser_getHealthStatus(t *testing.T) {
	tests := []struct {
		name     string
		host     *admin.HostStatus
		expected string
	}{
		{
			name: "healthy status",
			host: &admin.HostStatus{
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_HEALTHY,
				},
			},
			expected: "HEALTHY",
		},
		{
			name: "unhealthy status",
			host: &admin.HostStatus{
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_UNHEALTHY,
				},
			},
			expected: "UNHEALTHY",
		},
		{
			name: "draining status",
			host: &admin.HostStatus{
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_DRAINING,
				},
			},
			expected: "DRAINING",
		},
		{
			name: "unknown status",
			host: &admin.HostStatus{
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_UNKNOWN,
				},
			},
			expected: "UNKNOWN",
		},
		{
			name: "nil health status",
			host: &admin.HostStatus{
				HealthStatus: nil,
			},
			expected: "UNKNOWN",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.getHealthStatus(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_GetOutlierCheckStatus(t *testing.T) {
	// Note: We can't easily test the FailedOutlierCheck field in unit tests without
	// the actual protobuf structure being properly initialized from real Envoy data.
	// This functionality would be better tested with integration tests.
	tests := []struct {
		name     string
		host     *admin.HostStatus
		expected bool
	}{
		{
			name: "nil health status",
			host: &admin.HostStatus{
				HealthStatus: nil,
			},
			expected: false,
		},
		{
			name: "valid health status",
			host: &admin.HostStatus{
				HealthStatus: &admin.HostHealthStatus{
					EdsHealthStatus: core.HealthStatus_HEALTHY,
				},
			},
			expected: false, // Default case when no outlier check is failed
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetOutlierCheckStatus(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark for performance testing
func BenchmarkParser_ParseJSON(b *testing.B) {
	parser := NewParser()
	data, err := os.ReadFile(filepath.Join("testdata", "benchmark_clusters.json")) // #nosec G304 - static path is safe
	if err != nil {
		b.Fatalf("failed to read benchmark test data: %v", err)
	}
	input := string(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseJSON(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParser_convertHostToEndpoint(b *testing.B) {
	parser := NewParser()
	host := &admin.HostStatus{
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Address: "10.244.0.10",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: 80,
					},
				},
			},
		},
		HealthStatus: &admin.HostHealthStatus{
			EdsHealthStatus: core.HealthStatus_HEALTHY,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.convertHostToEndpoint(host)
	}
}
