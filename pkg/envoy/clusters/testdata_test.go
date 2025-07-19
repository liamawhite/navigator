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

// TestRealClusterDataParsing tests the parser with real cluster data from testdata
func TestRealClusterDataParsing(t *testing.T) {
	// Load the real cluster data from testdata
	data, err := os.ReadFile("testdata/sample_clusters.json")
	require.NoError(t, err, "Failed to read sample clusters data")

	// Create parser and parse raw JSON directly
	parser := NewParser()
	summaries, err := parser.ParseJSON(string(data))
	require.NoError(t, err, "Failed to parse clusters")

	// We should have parsed some clusters
	assert.Greater(t, len(summaries), 0, "Should have parsed some clusters")

	// Test specific address type detection
	addressTypeTests := []struct {
		clusterName         string
		expectedAddressType v1alpha1.AddressType
		expectedAddress     string
	}{
		{
			clusterName:         "sds-grpc",
			expectedAddressType: v1alpha1.AddressType_PIPE_ADDRESS,
			expectedAddress:     "unix://./var/run/secrets/workload-spiffe-uds/socket",
		},
		{
			clusterName:         "xds-grpc",
			expectedAddressType: v1alpha1.AddressType_PIPE_ADDRESS,
			expectedAddress:     "unix://./etc/istio/proxy/XDS",
		},
		{
			clusterName:         "prometheus_stats",
			expectedAddressType: v1alpha1.AddressType_SOCKET_ADDRESS,
			expectedAddress:     "127.0.0.1",
		},
		{
			clusterName:         "agent",
			expectedAddressType: v1alpha1.AddressType_SOCKET_ADDRESS,
			expectedAddress:     "127.0.0.1",
		},
	}

	// Find and test each expected cluster
	for _, test := range addressTypeTests {
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
				"Cluster %s should have address type %s", test.clusterName, test.expectedAddressType.String())
			assert.Equal(t, test.expectedAddress, endpoint.Address,
				"Cluster %s should have address %s", test.clusterName, test.expectedAddress)
		})
	}

	// Test that we have both socket and pipe addresses
	socketAddressCount := 0
	pipeAddressCount := 0
	unknownAddressCount := 0

	for _, summary := range summaries {
		for _, endpoint := range summary.Endpoints {
			switch endpoint.AddressType {
			case v1alpha1.AddressType_SOCKET_ADDRESS:
				socketAddressCount++
			case v1alpha1.AddressType_PIPE_ADDRESS:
				pipeAddressCount++
			case v1alpha1.AddressType_UNKNOWN_ADDRESS_TYPE:
				unknownAddressCount++
			}
		}
	}

	assert.Greater(t, socketAddressCount, 0, "Should have some socket addresses")
	assert.Greater(t, pipeAddressCount, 0, "Should have some pipe addresses")
	assert.Equal(t, 0, unknownAddressCount, "Should not have any unknown address types")

	t.Logf("Parsed %d clusters with %d socket addresses and %d pipe addresses",
		len(summaries), socketAddressCount, pipeAddressCount)
}

// TestAddressTypeDetection tests address type detection from real test data
func TestAddressTypeDetection(t *testing.T) {
	// Load the real cluster data from testdata
	data, err := os.ReadFile("testdata/sample_clusters.json")
	require.NoError(t, err, "Failed to read sample clusters data")

	// Create parser and parse raw JSON directly
	parser := NewParser()
	summaries, err := parser.ParseJSON(string(data))
	require.NoError(t, err, "Failed to parse clusters")

	// Check that we detected different address types correctly
	expectedTests := []struct {
		clusterName  string
		expectedType v1alpha1.AddressType
	}{
		{"xds-grpc", v1alpha1.AddressType_PIPE_ADDRESS},
		{"sds-grpc", v1alpha1.AddressType_PIPE_ADDRESS},
		{"prometheus_stats", v1alpha1.AddressType_SOCKET_ADDRESS},
		{"agent", v1alpha1.AddressType_SOCKET_ADDRESS},
		{"outbound|8080||pihole-http.pihole.svc.cluster.local", v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS},
		{"outbound|53||pihole-dns.pihole.svc.cluster.local", v1alpha1.AddressType_ENVOY_INTERNAL_ADDRESS},
	}

	for _, test := range expectedTests {
		t.Run(test.clusterName, func(t *testing.T) {
			var found *v1alpha1.EndpointSummary
			for _, summary := range summaries {
				if summary.ClusterName == test.clusterName {
					found = summary
					break
				}
			}

			require.NotNil(t, found, "Should have found cluster %s", test.clusterName)
			require.Greater(t, len(found.Endpoints), 0, "Cluster should have endpoints")

			endpoint := found.Endpoints[0]
			assert.Equal(t, test.expectedType, endpoint.AddressType,
				"Cluster %s should have address type %s but got %s",
				test.clusterName, test.expectedType.String(), endpoint.AddressType.String())
		})
	}
}
