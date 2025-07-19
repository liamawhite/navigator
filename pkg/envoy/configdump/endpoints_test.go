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

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestParser_summarizeEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint *endpointv3.ClusterLoadAssignment
		expected *v1alpha1.EndpointSummary
	}{
		{
			name: "valid endpoint with single healthy endpoint",
			endpoint: &endpointv3.ClusterLoadAssignment{
				ClusterName: "outbound|80||httpbin.default.svc.cluster.local",
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						Priority:            0,
						LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 100},
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HealthStatus:        corev3.HealthStatus_HEALTHY,
								LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 1},
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: "10.244.0.10",
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: 80,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &v1alpha1.EndpointSummary{
				ClusterName: "outbound|80||httpbin.default.svc.cluster.local",
				Endpoints: []*v1alpha1.EndpointInfo{
					{
						Address:             "10.244.0.10",
						Port:                80,
						HostIdentifier:      "10.244.0.10:80",
						Health:              "HEALTHY",
						LoadBalancingWeight: 1,
						Priority:            0,
						Weight:              100,
					},
				},
			},
		},
		{
			name: "endpoint with hostname",
			endpoint: &endpointv3.ClusterLoadAssignment{
				ClusterName: "outbound|443||api.external.com",
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						Priority: 0,
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HealthStatus: corev3.HealthStatus_HEALTHY,
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Hostname: "api.external.com",
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: "203.0.113.1",
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: 443,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &v1alpha1.EndpointSummary{
				ClusterName: "outbound|443||api.external.com",
				Endpoints: []*v1alpha1.EndpointInfo{
					{
						Address:             "203.0.113.1",
						Port:                443,
						HostIdentifier:      "api.external.com",
						Health:              "HEALTHY",
						LoadBalancingWeight: 0,
						Priority:            0,
						Weight:              0,
					},
				},
			},
		},
		{
			name: "endpoint with multiple localities and health statuses",
			endpoint: &endpointv3.ClusterLoadAssignment{
				ClusterName: "outbound|8080||backend.demo.svc.cluster.local",
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						Priority:            0,
						LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 50},
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HealthStatus:        corev3.HealthStatus_HEALTHY,
								LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 1},
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: "10.244.0.20",
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: 8080,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Priority:            1,
						LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 25},
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HealthStatus:        corev3.HealthStatus_UNHEALTHY,
								LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 1},
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: "10.244.0.21",
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: 8080,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &v1alpha1.EndpointSummary{
				ClusterName: "outbound|8080||backend.demo.svc.cluster.local",
				Endpoints: []*v1alpha1.EndpointInfo{
					{
						Address:             "10.244.0.20",
						Port:                8080,
						HostIdentifier:      "10.244.0.20:8080",
						Health:              "HEALTHY",
						LoadBalancingWeight: 1,
						Priority:            0,
						Weight:              50,
					},
					{
						Address:             "10.244.0.21",
						Port:                8080,
						HostIdentifier:      "10.244.0.21:8080",
						Health:              "UNHEALTHY",
						LoadBalancingWeight: 1,
						Priority:            1,
						Weight:              25,
					},
				},
			},
		},
		{
			name: "endpoint with no address",
			endpoint: &endpointv3.ClusterLoadAssignment{
				ClusterName: "test-cluster",
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						Priority: 0,
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HealthStatus: corev3.HealthStatus_HEALTHY,
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										// No address specified
									},
								},
							},
						},
					},
				},
			},
			expected: &v1alpha1.EndpointSummary{
				ClusterName: "test-cluster",
				Endpoints: []*v1alpha1.EndpointInfo{
					{
						Address:             "",
						Port:                0,
						HostIdentifier:      ":0",
						Health:              "HEALTHY",
						LoadBalancingWeight: 0,
						Priority:            0,
						Weight:              0,
					},
				},
			},
		},
		{
			name: "empty endpoint cluster",
			endpoint: &endpointv3.ClusterLoadAssignment{
				ClusterName: "empty-cluster",
				Endpoints:   []*endpointv3.LocalityLbEndpoints{},
			},
			expected: &v1alpha1.EndpointSummary{
				ClusterName: "empty-cluster",
				Endpoints:   []*v1alpha1.EndpointInfo{},
			},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.summarizeEndpoint(tt.endpoint)

			require.NotNil(t, result, "summary should not be nil")
			assert.Equal(t, tt.expected.ClusterName, result.ClusterName, "cluster name should match")
			assert.Equal(t, len(tt.expected.Endpoints), len(result.Endpoints), "endpoint count should match")

			for i, expectedEndpoint := range tt.expected.Endpoints {
				if i < len(result.Endpoints) {
					actualEndpoint := result.Endpoints[i]
					assert.Equal(t, expectedEndpoint.Address, actualEndpoint.Address, "endpoint %d address should match", i)
					assert.Equal(t, expectedEndpoint.Port, actualEndpoint.Port, "endpoint %d port should match", i)
					assert.Equal(t, expectedEndpoint.HostIdentifier, actualEndpoint.HostIdentifier, "endpoint %d host identifier should match", i)
					assert.Equal(t, expectedEndpoint.Health, actualEndpoint.Health, "endpoint %d health should match", i)
					assert.Equal(t, expectedEndpoint.LoadBalancingWeight, actualEndpoint.LoadBalancingWeight, "endpoint %d load balancing weight should match", i)
					assert.Equal(t, expectedEndpoint.Priority, actualEndpoint.Priority, "endpoint %d priority should match", i)
					assert.Equal(t, expectedEndpoint.Weight, actualEndpoint.Weight, "endpoint %d weight should match", i)
				}
			}
		})
	}
}

func TestParser_summarizeEndpoint_NilEndpoint(t *testing.T) {
	parser := NewParser()
	result := parser.summarizeEndpoint(nil)
	assert.Nil(t, result, "summary should be nil for nil endpoint")
}

func TestParser_summarizeEndpoint_HealthStatuses(t *testing.T) {
	healthTests := []struct {
		status   corev3.HealthStatus
		expected string
	}{
		{corev3.HealthStatus_UNKNOWN, "UNKNOWN"},
		{corev3.HealthStatus_HEALTHY, "HEALTHY"},
		{corev3.HealthStatus_UNHEALTHY, "UNHEALTHY"},
		{corev3.HealthStatus_DRAINING, "DRAINING"},
		{corev3.HealthStatus_TIMEOUT, "TIMEOUT"},
		{corev3.HealthStatus_DEGRADED, "DEGRADED"},
	}

	parser := NewParser()

	for _, tt := range healthTests {
		t.Run(tt.expected, func(t *testing.T) {
			endpoint := &endpointv3.ClusterLoadAssignment{
				ClusterName: "test-cluster",
				Endpoints: []*endpointv3.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpointv3.LbEndpoint{
							{
								HealthStatus: tt.status,
								HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
									Endpoint: &endpointv3.Endpoint{
										Address: &corev3.Address{
											Address: &corev3.Address_SocketAddress{
												SocketAddress: &corev3.SocketAddress{
													Address: "127.0.0.1",
													PortSpecifier: &corev3.SocketAddress_PortValue{
														PortValue: 8080,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			result := parser.summarizeEndpoint(endpoint)
			require.NotNil(t, result)
			require.Len(t, result.Endpoints, 1)
			assert.Equal(t, tt.expected, result.Endpoints[0].Health)
		})
	}
}

// Benchmark for performance testing
func BenchmarkParser_summarizeEndpoint(b *testing.B) {
	parser := NewParser()
	endpoint := &endpointv3.ClusterLoadAssignment{
		ClusterName: "outbound|80||httpbin.default.svc.cluster.local",
		Endpoints: []*endpointv3.LocalityLbEndpoints{
			{
				Priority:            0,
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 100},
				LbEndpoints: []*endpointv3.LbEndpoint{
					{
						HealthStatus:        corev3.HealthStatus_HEALTHY,
						LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 1},
						HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
							Endpoint: &endpointv3.Endpoint{
								Address: &corev3.Address{
									Address: &corev3.Address_SocketAddress{
										SocketAddress: &corev3.SocketAddress{
											Address: "10.244.0.10",
											PortSpecifier: &corev3.SocketAddress_PortValue{
												PortValue: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.summarizeEndpoint(endpoint)
	}
}
