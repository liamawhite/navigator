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
	"testing"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAmbientModeAddressTypes tests address type detection with real cluster data
// that includes Istio ambient mode envoy_internal_address configurations.
// This test validates that the parser correctly identifies different address types
// commonly found in ambient mesh deployments.
func TestAmbientModeAddressTypes(t *testing.T) {
	// Load the real cluster data from testdata (contains ambient mode examples)
	data, err := os.ReadFile("testdata/sample_clusters.json")
	require.NoError(t, err, "Failed to read sample clusters data")

	// Create parser and parse raw JSON directly
	parser := NewParser()
	summaries, err := parser.ParseJSON(string(data))
	require.NoError(t, err, "Failed to parse clusters")

	// We should have parsed some clusters
	assert.Greater(t, len(summaries), 0, "Should have parsed some clusters")

	// Test specific ambient mode address type detection scenarios
	ambientTests := []struct {
		clusterName         string
		expectedAddressType v1alpha1.AddressType
		description         string
	}{
		{
			clusterName:         "outbound|8080||pihole-http.pihole.svc.cluster.local",
			expectedAddressType: v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS,
			description:         "Ambient mode outbound cluster with connect_originate listener",
		},
		{
			clusterName:         "outbound|53||pihole-dns.pihole.svc.cluster.local",
			expectedAddressType: v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS,
			description:         "Ambient mode DNS cluster with envoy internal address",
		},
		{
			clusterName:         "sds-grpc",
			expectedAddressType: v1alpha1.AddressType_PIPE_ADDRESS,
			description:         "SDS gRPC communication via Unix domain socket",
		},
		{
			clusterName:         "xds-grpc",
			expectedAddressType: v1alpha1.AddressType_PIPE_ADDRESS,
			description:         "XDS gRPC communication via Unix domain socket",
		},
		{
			clusterName:         "outbound|15010||istiod.istio-system.svc.cluster.local",
			expectedAddressType: v1alpha1.AddressType_SOCKET_ADDRESS,
			description:         "Standard socket address for control plane communication",
		},
	}

	// Find and test each expected cluster
	for _, test := range ambientTests {
		t.Run(test.clusterName, func(t *testing.T) {
			var found *v1alpha1.EndpointSummary
			for _, summary := range summaries {
				if summary.ClusterName == test.clusterName {
					found = summary
					break
				}
			}

			require.NotNil(t, found, "Should have found cluster %s", test.clusterName)
			require.Greater(t, len(found.Endpoints), 0, "Cluster %s should have endpoints", test.clusterName)

			endpoint := found.Endpoints[0]
			assert.Equal(t, test.expectedAddressType, endpoint.AddressType,
				"%s: Expected address type %s but got %s",
				test.description, test.expectedAddressType.String(), endpoint.AddressType.String())

			// Additional validation for envoy internal addresses
			if test.expectedAddressType == v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS {
				assert.Contains(t, endpoint.Address, "envoy://",
					"Envoy internal address should be prefixed with envoy://")
				assert.NotEmpty(t, endpoint.HostIdentifier,
					"Envoy internal address should have host identifier")
			}

			// Additional validation for pipe addresses
			if test.expectedAddressType == v1alpha1.AddressType_PIPE_ADDRESS {
				assert.Contains(t, endpoint.Address, "unix://",
					"Pipe address should be prefixed with unix://")
			}
		})
	}

	// Validate that we correctly identify all three address types
	addressTypeCounts := map[v1alpha1.AddressType]int{
		v1alpha1.AddressType_SOCKET_ADDRESS:         0,
		v1alpha1.AddressType_PIPE_ADDRESS:           0,
		v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS: 0,
		v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE:   0,
	}

	for _, summary := range summaries {
		for _, endpoint := range summary.Endpoints {
			addressTypeCounts[endpoint.AddressType]++
		}
	}

	// We should have examples of all address types in ambient mode
	assert.Greater(t, addressTypeCounts[v1alpha1.AddressType_SOCKET_ADDRESS], 0,
		"Should have socket addresses for regular service communication")
	assert.Greater(t, addressTypeCounts[v1alpha1.AddressType_PIPE_ADDRESS], 0,
		"Should have pipe addresses for local gRPC communication")
	assert.Greater(t, addressTypeCounts[v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS], 0,
		"Should have envoy internal addresses for ambient mode routing")
	assert.Equal(t, 0, addressTypeCounts[v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE],
		"Should not have any unknown address types - all should be correctly classified")

	t.Logf("Address type distribution in ambient mode: Socket=%d, Pipe=%d, EnvoyInternal=%d, Unknown=%d",
		addressTypeCounts[v1alpha1.AddressType_SOCKET_ADDRESS],
		addressTypeCounts[v1alpha1.AddressType_PIPE_ADDRESS],
		addressTypeCounts[v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS],
		addressTypeCounts[v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE])
}

// TestEnvoyInternalAddressFormat tests that envoy internal addresses are correctly
// formatted with the connect_originate pattern common in ambient mode
func TestEnvoyInternalAddressFormat(t *testing.T) {
	data, err := os.ReadFile("testdata/sample_clusters.json")
	require.NoError(t, err)

	parser := NewParser()
	summaries, err := parser.ParseJSON(string(data))
	require.NoError(t, err)

	// Find clusters with envoy internal addresses
	envoyInternalClusters := []string{
		"outbound|8080||pihole-http.pihole.svc.cluster.local",
		"outbound|53||pihole-dns.pihole.svc.cluster.local",
	}

	for _, clusterName := range envoyInternalClusters {
		t.Run(clusterName, func(t *testing.T) {
			var found *v1alpha1.EndpointSummary
			for _, summary := range summaries {
				if summary.ClusterName == clusterName {
					found = summary
					break
				}
			}

			require.NotNil(t, found)
			require.Greater(t, len(found.Endpoints), 0)

			endpoint := found.Endpoints[0]
			assert.Equal(t, v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS, endpoint.AddressType)

			// Check the format: should be envoy://connect_originate/endpoint_id
			assert.Contains(t, endpoint.Address, "envoy://connect_originate/",
				"Ambient mode envoy internal addresses should use connect_originate listener")

			// Check that endpoint ID is preserved
			assert.Contains(t, endpoint.Address, "10.42.0.26:",
				"Should contain the original endpoint ID with IP and port")
		})
	}
}
