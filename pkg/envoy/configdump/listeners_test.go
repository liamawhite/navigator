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
	"encoding/json"
	"os"
	"testing"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_DetermineListenerType(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name           string
		listenerName   string
		address        string
		port           uint32
		useOriginalDst bool
		expected       v1alpha1.ListenerType
	}{
		// Admin ports
		{
			name:         "XDS port - port-based",
			listenerName: "0.0.0.0_15010",
			address:      "0.0.0.0",
			port:         15010,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
		},
		{
			name:         "Webhook admin port on specific IP",
			listenerName: "10.96.245.191_15012",
			address:      "10.96.245.191",
			port:         15012,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
		},
		{
			name:         "Debug port - port-based",
			listenerName: "0.0.0.0_15014",
			address:      "0.0.0.0",
			port:         15014,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
		},
		{
			name:         "Metrics port",
			listenerName: "0.0.0.0_15090",
			address:      "0.0.0.0",
			port:         15090,
			expected:     v1alpha1.ListenerType_PROXY_METRICS,
		},
		{
			name:         "Health check port - 0.0.0.0",
			listenerName: "0.0.0.0_15021",
			address:      "0.0.0.0",
			port:         15021,
			expected:     v1alpha1.ListenerType_PROXY_HEALTHCHECK,
		},
		{
			name:         "Health check port - specific IP",
			listenerName: "10.96.240.89_15021",
			address:      "10.96.240.89",
			port:         15021,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
		},
		// Virtual listeners by name (most reliable)
		{
			name:         "Virtual outbound by name",
			listenerName: "virtualOutbound",
			address:      "0.0.0.0",
			port:         15001,
			expected:     v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
		},
		{
			name:         "Virtual inbound by name",
			listenerName: "virtualInbound",
			address:      "0.0.0.0",
			port:         9000,
			expected:     v1alpha1.ListenerType_VIRTUAL_INBOUND,
		},
		// Other listeners on 0.0.0.0 without literal names fall back to virtual types
		{
			name:           "Port 15001 without literal name",
			listenerName:   "0.0.0.0_15001",
			address:        "0.0.0.0",
			port:           15001,
			useOriginalDst: true,
			expected:       v1alpha1.ListenerType_VIRTUAL_OUTBOUND,
		},
		{
			name:           "Port 9000 without literal name",
			listenerName:   "0.0.0.0_9000",
			address:        "0.0.0.0",
			port:           9000,
			useOriginalDst: false,
			expected:       v1alpha1.ListenerType_PORT_OUTBOUND,
		},
		// Service-specific outbound listeners
		{
			name:         "Service outbound listener - specific IP",
			listenerName: "10.96.173.1_8080",
			address:      "10.96.173.1",
			port:         8080,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
		},
		{
			name:         "Application port - port-based",
			listenerName: "0.0.0.0_80",
			address:      "0.0.0.0",
			port:         80,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
		},
		{
			name:         "Port-based listener - generic port on 0.0.0.0",
			listenerName: "0.0.0.0_3000",
			address:      "0.0.0.0",
			port:         3000,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.determineListenerType(tt.listenerName, tt.address, tt.port, tt.useOriginalDst)
			assert.Equal(t, tt.expected, result,
				"Expected %s for name=%s, address=%s, port=%d, useOriginalDst=%t, got %s",
				tt.expected.String(), tt.listenerName, tt.address, tt.port, tt.useOriginalDst, result.String())
		})
	}
}

func TestParser_DetermineListenerType_RealEnvoyScenarios(t *testing.T) {
	parser := NewParser()

	// Test cases based on actual listener names from real Envoy config:
	// [10.96.0.10_53 10.96.0.10_9153 0.0.0.0_15010 10.96.240.89_15021 0.0.0.0_80
	//  10.96.173.1_8080 10.96.36.70_8080 10.96.86.22_8080 10.96.0.1_443
	//  10.96.245.191_15012 10.96.245.191_443 0.0.0.0_15014 10.96.240.89_443
	//  virtualOutbound virtualInbound 0.0.0.0_15090 0.0.0.0_15021]

	realScenarios := []struct {
		name           string
		listenerName   string
		address        string
		port           uint32
		useOriginalDst bool
		expected       v1alpha1.ListenerType
		description    string
	}{
		{
			name:         "DNS service 53",
			listenerName: "10.96.0.10_53",
			address:      "10.96.0.10",
			port:         53,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:  "Outbound to kube-dns service",
		},
		{
			name:         "DNS service 9153",
			listenerName: "10.96.0.10_9153",
			address:      "10.96.0.10",
			port:         9153,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:  "Outbound to kube-dns metrics",
		},
		{
			name:         "XDS port - 0.0.0.0",
			listenerName: "0.0.0.0_15010",
			address:      "0.0.0.0",
			port:         15010,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
			description:  "Generic traffic on xDS port",
		},
		{
			name:         "Health check - specific IP",
			listenerName: "10.96.240.89_15021",
			address:      "10.96.240.89",
			port:         15021,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:  "Outbound to Istio gateway health check",
		},
		{
			name:         "Application port - port-based",
			listenerName: "0.0.0.0_80",
			address:      "0.0.0.0",
			port:         80,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
			description:  "Generic HTTP port traffic",
		},
		{
			name:         "Backend service",
			listenerName: "10.96.173.1_8080",
			address:      "10.96.173.1",
			port:         8080,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:  "Outbound to backend service",
		},
		{
			name:         "Webhook admin",
			listenerName: "10.96.245.191_15012",
			address:      "10.96.245.191",
			port:         15012,
			expected:     v1alpha1.ListenerType_SERVICE_OUTBOUND,
			description:  "Outbound to Istio webhook service",
		},
		{
			name:         "Debug port - 0.0.0.0",
			listenerName: "0.0.0.0_15014",
			address:      "0.0.0.0",
			port:         15014,
			expected:     v1alpha1.ListenerType_PORT_OUTBOUND,
			description:  "Generic traffic on debug port",
		},
		{
			name:         "Metrics - 0.0.0.0",
			listenerName: "0.0.0.0_15090",
			address:      "0.0.0.0",
			port:         15090,
			expected:     v1alpha1.ListenerType_PROXY_METRICS,
			description:  "Prometheus metrics",
		},
		{
			name:         "Health check - 0.0.0.0",
			listenerName: "0.0.0.0_15021",
			address:      "0.0.0.0",
			port:         15021,
			expected:     v1alpha1.ListenerType_PROXY_HEALTHCHECK,
			description:  "Health check endpoint",
		},
	}

	for _, tt := range realScenarios {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.determineListenerType(tt.listenerName, tt.address, tt.port, tt.useOriginalDst)
			assert.Equal(t, tt.expected, result,
				"Scenario: %s\nExpected %s for name=%s, address=%s, port=%d, got %s",
				tt.description, tt.expected.String(), tt.listenerName, tt.address, tt.port, result.String())
		})
	}
}

func TestParser_RealEnvoyConfigListenerTypes(t *testing.T) {
	// Load real Envoy config dump from test data
	testDataPath := "testdata/envoy_config_dump.json"
	_, err := os.Stat(testDataPath)
	if os.IsNotExist(err) {
		t.Skip("Real Envoy config dump test data not found")
		return
	}

	configData, err := os.ReadFile(testDataPath)
	require.NoError(t, err, "Failed to read test data")

	parser := NewParser()
	parsed, err := parser.ParseJSONToSummary(string(configData))
	require.NoError(t, err, "Failed to parse real Envoy config")

	// Expected listener types based on the real config output:
	// [10.96.0.10_53 10.96.0.10_9153 0.0.0.0_15010 10.96.240.89_15021 0.0.0.0_80
	//  10.96.173.1_8080 10.96.36.70_8080 10.96.86.22_8080 10.96.0.1_443
	//  10.96.245.191_15012 10.96.245.191_443 0.0.0.0_15014 10.96.240.89_443
	//  virtualOutbound virtualInbound 0.0.0.0_15090 0.0.0.0_15021]

	expectedTypes := map[string]v1alpha1.ListenerType{
		"10.96.0.10_53":       v1alpha1.ListenerType_SERVICE_OUTBOUND,  // DNS service outbound
		"10.96.0.10_9153":     v1alpha1.ListenerType_SERVICE_OUTBOUND,  // DNS metrics service outbound
		"0.0.0.0_15010":       v1alpha1.ListenerType_PORT_OUTBOUND,     // Generic xDS port traffic
		"10.96.240.89_15021":  v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Gateway health check service outbound
		"0.0.0.0_80":          v1alpha1.ListenerType_PORT_OUTBOUND,     // Generic HTTP port traffic
		"10.96.173.1_8080":    v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Backend service outbound
		"10.96.36.70_8080":    v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Database service outbound
		"10.96.86.22_8080":    v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Frontend service outbound
		"10.96.0.1_443":       v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Kubernetes API service outbound
		"10.96.245.191_15012": v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Istio webhook service outbound
		"10.96.245.191_443":   v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Istio webhook HTTPS service outbound
		"0.0.0.0_15014":       v1alpha1.ListenerType_PORT_OUTBOUND,     // Generic debug port traffic
		"10.96.240.89_443":    v1alpha1.ListenerType_SERVICE_OUTBOUND,  // Gateway HTTPS service outbound
		"virtualOutbound":     v1alpha1.ListenerType_VIRTUAL_OUTBOUND,  // Virtual outbound
		"virtualInbound":      v1alpha1.ListenerType_VIRTUAL_INBOUND,   // Virtual inbound
		"0.0.0.0_15090":       v1alpha1.ListenerType_PROXY_METRICS,     // Prometheus metrics
		"0.0.0.0_15021":       v1alpha1.ListenerType_PROXY_HEALTHCHECK, // Health check
	}

	require.Equal(t, len(expectedTypes), len(parsed.Listeners),
		"Expected %d listeners, got %d", len(expectedTypes), len(parsed.Listeners))

	// Check each listener type
	for _, listener := range parsed.Listeners {
		expectedType, exists := expectedTypes[listener.Name]
		require.True(t, exists, "Unexpected listener found: %s", listener.Name)

		assert.Equal(t, expectedType, listener.Type,
			"Listener %s: expected type %s, got %s (address=%s, port=%d)",
			listener.Name, expectedType.String(), listener.Type.String(),
			listener.Address, listener.Port)
	}

	// Count listener types
	typeCounts := make(map[v1alpha1.ListenerType]int)
	for _, listener := range parsed.Listeners {
		typeCounts[listener.Type]++
	}

	t.Logf("Listener type distribution:")
	for listenerType, count := range typeCounts {
		t.Logf("  %s: %d", listenerType.String(), count)
	}

	// Verify we have all expected types with correct counts
	assert.Greater(t, typeCounts[v1alpha1.ListenerType_SERVICE_OUTBOUND], 0, "Should have service outbound listeners")
	assert.Greater(t, typeCounts[v1alpha1.ListenerType_PORT_OUTBOUND], 0, "Should have port outbound listeners")
	assert.Greater(t, typeCounts[v1alpha1.ListenerType_VIRTUAL_INBOUND], 0, "Should have virtual inbound listeners")
	assert.Greater(t, typeCounts[v1alpha1.ListenerType_VIRTUAL_OUTBOUND], 0, "Should have virtual outbound listeners")
	assert.Greater(t, typeCounts[v1alpha1.ListenerType_PROXY_METRICS], 0, "Should have proxy metrics listeners")
	assert.Greater(t, typeCounts[v1alpha1.ListenerType_PROXY_HEALTHCHECK], 0, "Should have proxy health check listeners")
}

func TestParser_RawConfigPopulation(t *testing.T) {
	// Load real Envoy config dump from test data
	testDataPath := "testdata/envoy_config_dump.json"
	_, err := os.Stat(testDataPath)
	if os.IsNotExist(err) {
		t.Skip("Real Envoy config dump test data not found")
		return
	}

	configData, err := os.ReadFile(testDataPath)
	require.NoError(t, err, "Failed to read test data")

	parser := NewParser()
	parsed, err := parser.ParseJSONToSummary(string(configData))
	require.NoError(t, err, "Failed to parse real Envoy config")

	// Check that we have listeners
	require.Greater(t, len(parsed.Listeners), 0, "Should have listeners")

	// Check that all listeners have raw config populated
	for _, listener := range parsed.Listeners {
		assert.NotEmpty(t, listener.RawConfig, "Listener %s should have raw config", listener.Name)

		// Verify it's valid JSON
		var jsonData interface{}
		err := json.Unmarshal([]byte(listener.RawConfig), &jsonData)
		assert.NoError(t, err, "Raw config for listener %s should be valid JSON", listener.Name)

		// Check that it contains expected fields
		assert.Contains(t, listener.RawConfig, "name", "Raw config should contain 'name' field")
		assert.Contains(t, listener.RawConfig, listener.Name, "Raw config should contain the listener name")

		t.Logf("Listener %s raw config length: %d characters", listener.Name, len(listener.RawConfig))
	}
}
