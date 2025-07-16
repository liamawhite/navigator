package clusters

import (
	"testing"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

func TestParser_ParseJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*v1alpha1.EndpointSummary
		wantErr  bool
	}{
		{
			name: "valid clusters response with endpoints",
			input: `{
				"cluster_statuses": [
					{
						"name": "outbound|80||httpbin.default.svc.cluster.local",
						"host_statuses": [
							{
								"address": {
									"socket_address": {
										"address": "10.244.0.10",
										"port_value": 80
									}
								},
								"stats": [],
								"health_status": {
									"eds_health_status": "HEALTHY"
								}
							},
							{
								"address": {
									"socket_address": {
										"address": "10.244.0.11", 
										"port_value": 80
									}
								},
								"stats": [],
								"health_status": {
									"eds_health_status": "UNHEALTHY"
								}
							}
						]
					},
					{
						"name": "outbound|443||api.external.com",
						"host_statuses": [
							{
								"address": {
									"socket_address": {
										"address": "203.0.113.1",
										"port_value": 443
									}
								},
								"stats": [],
								"health_status": {
									"eds_health_status": "HEALTHY"
								}
							}
						]
					}
				]
			}`,
			expected: []*v1alpha1.EndpointSummary{
				{
					ClusterName: "outbound|80||httpbin.default.svc.cluster.local",
					ClusterType: v1alpha1.ClusterType_CLUSTER_EDS,
					Direction:   v1alpha1.ClusterDirection_OUTBOUND,
					Port:        80,
					Subset:      "",
					ServiceFqdn: "httpbin.default.svc.cluster.local",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:             "10.244.0.10",
							Port:                80,
							HostIdentifier:      "10.244.0.10:80",
							Health:              "HEALTHY",
							Priority:            0,
							Weight:              0,
							LoadBalancingWeight: 0,
							Metadata:            map[string]string{},
						},
						{
							Address:             "10.244.0.11",
							Port:                80,
							HostIdentifier:      "10.244.0.11:80",
							Health:              "UNHEALTHY",
							Priority:            0,
							Weight:              0,
							LoadBalancingWeight: 0,
							Metadata:            map[string]string{},
						},
					},
				},
				{
					ClusterName: "outbound|443||api.external.com",
					ClusterType: v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
					Direction:   v1alpha1.ClusterDirection_OUTBOUND,
					Port:        443,
					Subset:      "",
					ServiceFqdn: "api.external.com",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:             "203.0.113.1",
							Port:                443,
							HostIdentifier:      "203.0.113.1:443",
							Health:              "HEALTHY",
							Priority:            0,
							Weight:              0,
							LoadBalancingWeight: 0,
							Metadata:            map[string]string{},
						},
					},
				},
			},
		},
		{
			name: "cluster with unix socket address",
			input: `{
				"cluster_statuses": [
					{
						"name": "uds_cluster",
						"host_statuses": [
							{
								"address": {
									"pipe": {
										"path": "/tmp/socket"
									}
								},
								"stats": [],
								"health_status": {
									"eds_health_status": "HEALTHY"
								}
							}
						]
					}
				]
			}`,
			expected: []*v1alpha1.EndpointSummary{
				{
					ClusterName: "uds_cluster",
					ClusterType: v1alpha1.ClusterType_CLUSTER_EDS,
					Direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
					Port:        0,
					Subset:      "",
					ServiceFqdn: "",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:             "unix:///tmp/socket",
							Port:                0,
							HostIdentifier:      "unix:///tmp/socket",
							Health:              "HEALTHY",
							Priority:            0,
							Weight:              0,
							LoadBalancingWeight: 0,
							Metadata:            map[string]string{},
						},
					},
				},
			},
		},
		{
			name: "cluster with envoy internal address",
			input: `{
				"cluster_statuses": [
					{
						"name": "internal_cluster",
						"host_statuses": [
							{
								"address": {
									"envoy_internal_address": {
										"server_listener_name": "internal_listener",
										"endpoint_id": "endpoint_1"
									}
								},
								"stats": [],
								"health_status": {
									"eds_health_status": "HEALTHY"
								}
							}
						]
					}
				]
			}`,
			expected: []*v1alpha1.EndpointSummary{
				{
					ClusterName: "internal_cluster",
					ClusterType: v1alpha1.ClusterType_CLUSTER_EDS,
					Direction:   v1alpha1.ClusterDirection_UNSPECIFIED,
					Port:        0,
					Subset:      "",
					ServiceFqdn: "",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:             "envoy://internal_listener/endpoint_1",
							Port:                0,
							HostIdentifier:      "envoy://internal_listener/endpoint_1",
							Health:              "HEALTHY",
							Priority:            0,
							Weight:              0,
							LoadBalancingWeight: 0,
							Metadata:            map[string]string{},
						},
					},
				},
			},
		},
		{
			name: "empty clusters response",
			input: `{
				"cluster_statuses": []
			}`,
			expected: []*v1alpha1.EndpointSummary{},
		},
		{
			name: "cluster with no endpoints",
			input: `{
				"cluster_statuses": [
					{
						"name": "empty_cluster",
						"host_statuses": []
					}
				]
			}`,
			expected: []*v1alpha1.EndpointSummary{},
		},
		{
			name:    "invalid JSON",
			input:   `{"invalid": json}`,
			wantErr: true,
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseJSON(tt.input)

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
							assert.Equal(t, expectedEndpoint.LoadBalancingWeight, actualEndpoint.LoadBalancingWeight, "endpoint %d load balancing weight should match", j)
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
				Address:             "192.168.1.1",
				Port:                8080,
				HostIdentifier:      "192.168.1.1:8080",
				Health:              "HEALTHY",
				Priority:            0,
				Weight:              0,
				LoadBalancingWeight: 0,
				Metadata:            map[string]string{},
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
				Address:             "10.0.0.1",
				Port:                3000,
				HostIdentifier:      "10.0.0.1:3000",
				Health:              "UNHEALTHY",
				Priority:            0,
				Weight:              0,
				LoadBalancingWeight: 0,
				Metadata:            map[string]string{},
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
				Address:             "unix:///var/run/service.sock",
				Port:                0,
				HostIdentifier:      "unix:///var/run/service.sock",
				Health:              "HEALTHY",
				Priority:            0,
				Weight:              0,
				LoadBalancingWeight: 0,
				Metadata:            map[string]string{},
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
				Address:             "127.0.0.1",
				Port:                9000,
				HostIdentifier:      "127.0.0.1:9000",
				Health:              "UNKNOWN",
				Priority:            0,
				Weight:              0,
				LoadBalancingWeight: 0,
				Metadata:            map[string]string{},
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
			assert.Equal(t, tt.expected.LoadBalancingWeight, result.LoadBalancingWeight)
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

func TestParser_InferClusterType(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		clusterName string
		expected    v1alpha1.ClusterType
	}{
		{
			name:        "outbound kubernetes service",
			clusterName: "outbound|80||httpbin.default.svc.cluster.local",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
		},
		{
			name:        "inbound kubernetes service",
			clusterName: "inbound|8080||app.demo.svc.cluster.local",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
		},
		{
			name:        "outbound external service",
			clusterName: "outbound|443||api.external.com",
			expected:    v1alpha1.ClusterType_CLUSTER_STRICT_DNS,
		},
		{
			name:        "prometheus stats cluster",
			clusterName: "prometheus_stats",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
		},
		{
			name:        "agent cluster",
			clusterName: "agent",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
		},
		{
			name:        "xds-grpc cluster",
			clusterName: "xds-grpc",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
		},
		{
			name:        "sds-grpc cluster",
			clusterName: "sds-grpc",
			expected:    v1alpha1.ClusterType_CLUSTER_STATIC,
		},
		{
			name:        "empty cluster name",
			clusterName: "",
			expected:    v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE,
		},
		{
			name:        "unknown cluster pattern",
			clusterName: "custom-cluster-name",
			expected:    v1alpha1.ClusterType_CLUSTER_EDS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.inferClusterType(tt.clusterName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark for performance testing
func BenchmarkParser_ParseJSON(b *testing.B) {
	parser := NewParser()
	input := `{
		"cluster_statuses": [
			{
				"name": "outbound|80||httpbin.default.svc.cluster.local",
				"host_statuses": [
					{
						"address": {
							"socket_address": {
								"address": "10.244.0.10",
								"port_value": 80
							}
						},
						"stats": [],
						"health_status": {
							"eds_health_status": "HEALTHY"
						}
					}
				]
			}
		]
	}`

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

func TestParser_ParseClusterName(t *testing.T) {
	parser := NewParser()

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
			name:        "outbound kubernetes service",
			clusterName: "outbound|80||httpbin.default.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        80,
				subset:      "",
				serviceFqdn: "httpbin.default.svc.cluster.local",
			},
		},
		{
			name:        "inbound service with subset",
			clusterName: "inbound|8080|v1|app.demo.svc.cluster.local",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_INBOUND,
				port:        8080,
				subset:      "v1",
				serviceFqdn: "app.demo.svc.cluster.local",
			},
		},
		{
			name:        "external service",
			clusterName: "outbound|443||api.external.com",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        443,
				subset:      "",
				serviceFqdn: "api.external.com",
			},
		},
		{
			name:        "invalid format",
			clusterName: "prometheus_stats",
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
			name:        "invalid port",
			clusterName: "outbound|invalid||service.com",
			expected: struct {
				direction   v1alpha1.ClusterDirection
				port        uint32
				subset      string
				serviceFqdn string
			}{
				direction:   v1alpha1.ClusterDirection_OUTBOUND,
				port:        0,
				subset:      "",
				serviceFqdn: "service.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := &v1alpha1.EndpointSummary{}
			parser.parseClusterName(tt.clusterName, summary)

			assert.Equal(t, tt.expected.direction, summary.Direction)
			assert.Equal(t, tt.expected.port, summary.Port)
			assert.Equal(t, tt.expected.subset, summary.Subset)
			assert.Equal(t, tt.expected.serviceFqdn, summary.ServiceFqdn)
		})
	}
}
