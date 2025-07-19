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
	"testing"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestConvertToEndpointSummaries(t *testing.T) {
	tests := []struct {
		name             string
		input            []*ClusterEndpointInfo
		expectedCount    int
		expectedClusters []string
	}{
		{
			name: "convert ambient mode endpoints",
			input: []*ClusterEndpointInfo{
				{
					ClusterName: "outbound|8080||pihole-http.pihole.svc.cluster.local",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:        "10.42.0.26:8080",
							Port:           0, // Internal addresses don't have ports
							HostIdentifier: "envoy://connect_originate/10.42.0.26:8080",
							AddressType:    v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS,
							Health:         "HEALTHY",
						},
					},
				},
				{
					ClusterName: "inbound|8080||",
					Endpoints: []*v1alpha1.EndpointInfo{
						{
							Address:        "192.168.1.100",
							Port:           8080,
							HostIdentifier: "192.168.1.100:8080",
							AddressType:    v1alpha1.AddressType_SOCKET_ADDRESS,
							Health:         "HEALTHY",
						},
					},
				},
			},
			expectedCount: 2,
			expectedClusters: []string{
				"outbound|8080||pihole-http.pihole.svc.cluster.local",
				"inbound|8080||",
			},
		},
		{
			name:          "empty input",
			input:         []*ClusterEndpointInfo{},
			expectedCount: 0,
		},
		{
			name: "cluster with no endpoints",
			input: []*ClusterEndpointInfo{
				{
					ClusterName: "empty-cluster",
					Endpoints:   []*v1alpha1.EndpointInfo{},
				},
			},
			expectedCount: 0, // Should skip empty clusters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToEndpointSummaries(tt.input)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d summaries, got %d", tt.expectedCount, len(result))
			}

			// Check cluster names if we have expected clusters
			if tt.expectedClusters != nil {
				for i, expectedCluster := range tt.expectedClusters {
					if i >= len(result) {
						t.Errorf("Missing expected cluster %s", expectedCluster)
						continue
					}
					if result[i].ClusterName != expectedCluster {
						t.Errorf("Expected cluster name %s, got %s", expectedCluster, result[i].ClusterName)
					}
				}
			}

			// Verify address type parsing for ambient mode test
			if tt.name == "convert ambient mode endpoints" {
				if len(result) >= 1 {
					// Check outbound cluster
					if result[0].Direction != v1alpha1.ClusterDirection_OUTBOUND {
						t.Errorf("Expected OUTBOUND direction, got %v", result[0].Direction)
					}
					if result[0].Port != 8080 {
						t.Errorf("Expected port 8080, got %d", result[0].Port)
					}
					if result[0].ServiceFqdn != "pihole-http.pihole.svc.cluster.local" {
						t.Errorf("Expected service FQDN pihole-http.pihole.svc.cluster.local, got %s", result[0].ServiceFqdn)
					}
				}
				if len(result) >= 2 {
					// Check inbound cluster
					if result[1].Direction != v1alpha1.ClusterDirection_INBOUND {
						t.Errorf("Expected INBOUND direction, got %v", result[1].Direction)
					}
				}
			}
		})
	}
}

func TestParseClusterNameComponents(t *testing.T) {
	tests := []struct {
		name              string
		clusterName       string
		expectedDirection v1alpha1.ClusterDirection
		expectedPort      uint32
		expectedSubset    string
		expectedFqdn      string
	}{
		{
			name:              "outbound service",
			clusterName:       "outbound|8080||pihole-http.pihole.svc.cluster.local",
			expectedDirection: v1alpha1.ClusterDirection_OUTBOUND,
			expectedPort:      8080,
			expectedSubset:    "",
			expectedFqdn:      "pihole-http.pihole.svc.cluster.local",
		},
		{
			name:              "inbound service with subset",
			clusterName:       "inbound|9090|v1|backend.demo.svc.cluster.local",
			expectedDirection: v1alpha1.ClusterDirection_INBOUND,
			expectedPort:      9090,
			expectedSubset:    "v1",
			expectedFqdn:      "backend.demo.svc.cluster.local",
		},
		{
			name:              "malformed cluster name",
			clusterName:       "invalid-format",
			expectedDirection: v1alpha1.ClusterDirection_UNSPECIFIED,
			expectedPort:      0,
			expectedSubset:    "",
			expectedFqdn:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			direction, port, subset, fqdn := parseClusterNameComponents(tt.clusterName)

			if direction != tt.expectedDirection {
				t.Errorf("Expected direction %v, got %v", tt.expectedDirection, direction)
			}
			if port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, port)
			}
			if subset != tt.expectedSubset {
				t.Errorf("Expected subset %s, got %s", tt.expectedSubset, subset)
			}
			if fqdn != tt.expectedFqdn {
				t.Errorf("Expected FQDN %s, got %s", tt.expectedFqdn, fqdn)
			}
		})
	}
}
