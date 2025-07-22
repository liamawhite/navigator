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

package enrich

import (
	"testing"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichBootstrapProxyMode(t *testing.T) {
	enrichFunc := enrichBootstrapProxyMode()

	tests := []struct {
		name        string
		nodeID      string
		expected    v1alpha1.ProxyMode
		description string
	}{
		{
			name:        "sidecar proxy with full Istio node ID",
			nodeID:      "sidecar~10.244.0.1~pod.namespace~cluster.local",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Standard Istio sidecar proxy node ID format",
		},
		{
			name:        "sidecar proxy with different IP",
			nodeID:      "sidecar~192.168.1.100~my-app.production~my-cluster.local",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Sidecar with production environment details",
		},
		{
			name:        "sidecar proxy uppercase",
			nodeID:      "SIDECAR~10.244.0.1~pod.namespace~cluster.local",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Case insensitive matching for sidecar",
		},
		{
			name:        "sidecar proxy mixed case",
			nodeID:      "SideCar~10.244.0.1~pod.namespace~cluster.local",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Mixed case should still match",
		},
		{
			name:        "router gateway proxy",
			nodeID:      "router~10.244.0.2~gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Istio ingress gateway with router prefix",
		},
		{
			name:        "gateway proxy explicit",
			nodeID:      "gateway~10.244.0.3~istio-gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Gateway proxy with explicit gateway prefix",
		},
		{
			name:        "router uppercase",
			nodeID:      "ROUTER~10.244.0.2~gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Case insensitive matching for router",
		},
		{
			name:        "gateway uppercase",
			nodeID:      "GATEWAY~10.244.0.3~gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Case insensitive matching for gateway",
		},
		{
			name:        "unknown proxy format",
			nodeID:      "unknown-format",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Non-Istio proxy format should be unknown",
		},
		{
			name:        "malformed node ID with tildes",
			nodeID:      "invalid~format~but~not~istio",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Malformed Istio-like format should be unknown",
		},
		{
			name:        "empty node ID",
			nodeID:      "",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Empty node ID should be unknown",
		},
		{
			name:        "whitespace only node ID",
			nodeID:      "   ",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Whitespace-only node ID should be unknown",
		},
		{
			name:        "partial sidecar match",
			nodeID:      "sidecar-but-not-tilde-format",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Partial match without tilde should be unknown",
		},
		{
			name:        "sidecar with minimal format",
			nodeID:      "sidecar~",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Minimal sidecar format should still match",
		},
		{
			name:        "router with minimal format",
			nodeID:      "router~",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Minimal router format should still match",
		},
		{
			name:        "gateway with minimal format",
			nodeID:      "gateway~",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Minimal gateway format should still match",
		},
		{
			name:        "envoy proxy generic format",
			nodeID:      "envoy-proxy-id-12345",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Generic Envoy proxy ID should be unknown",
		},
		{
			name:        "sidecar in middle of string",
			nodeID:      "not-sidecar~but-contains-sidecar",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Sidecar not at beginning should be unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bootstrap := &v1alpha1.BootstrapSummary{
				Node: &v1alpha1.NodeSummary{Id: test.nodeID},
			}

			err := enrichFunc(bootstrap)
			require.NoError(t, err, test.description)
			assert.Equal(t, test.expected, bootstrap.Node.ProxyMode, test.description)
		})
	}

	t.Run("handles nil bootstrap", func(t *testing.T) {
		err := enrichFunc(nil)
		assert.NoError(t, err)
	})

	t.Run("handles bootstrap with nil node", func(t *testing.T) {
		bootstrap := &v1alpha1.BootstrapSummary{Node: nil}
		err := enrichFunc(bootstrap)
		assert.NoError(t, err)
	})
}

func TestInferProxyMode(t *testing.T) {
	tests := []struct {
		name        string
		nodeID      string
		expected    v1alpha1.ProxyMode
		description string
	}{
		{
			name:        "sidecar proxy standard format",
			nodeID:      "sidecar~10.244.0.1~pod.namespace~cluster.local",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Standard Istio sidecar node ID",
		},
		{
			name:        "sidecar proxy uppercase",
			nodeID:      "SIDECAR~10.244.0.1~pod.namespace~cluster.local",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Case insensitive sidecar detection",
		},
		{
			name:        "sidecar proxy mixed case",
			nodeID:      "SideCar~10.244.0.1~pod.namespace~cluster.local",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Mixed case sidecar detection",
		},
		{
			name:        "router proxy standard format",
			nodeID:      "router~10.244.0.2~gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Standard Istio router/gateway node ID",
		},
		{
			name:        "gateway proxy standard format",
			nodeID:      "gateway~10.244.0.3~gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Standard Istio gateway node ID",
		},
		{
			name:        "router uppercase",
			nodeID:      "ROUTER~10.244.0.2~gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Case insensitive router detection",
		},
		{
			name:        "gateway uppercase",
			nodeID:      "GATEWAY~10.244.0.3~gateway.istio-system~cluster.local",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Case insensitive gateway detection",
		},
		{
			name:        "unknown format without tildes",
			nodeID:      "unknown-format",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Non-Istio format should be unknown",
		},
		{
			name:        "empty node ID",
			nodeID:      "",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Empty string should be unknown",
		},
		{
			name:        "random UUID format",
			nodeID:      "550e8400-e29b-41d4-a716-446655440000",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "UUID-like format should be unknown",
		},
		{
			name:        "consul connect format",
			nodeID:      "connect-proxy:web-sidecar-proxy",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Consul Connect format should be unknown",
		},
		{
			name:        "generic envoy format",
			nodeID:      "envoy-node-12345",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Generic Envoy node ID should be unknown",
		},
		{
			name:        "malformed istio-like format",
			nodeID:      "sidecar-but-no-tilde",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Istio-like but malformed should be unknown",
		},
		{
			name:        "sidecar with only tilde",
			nodeID:      "sidecar~",
			expected:    v1alpha1.ProxyMode_SIDECAR,
			description: "Minimal valid sidecar format",
		},
		{
			name:        "gateway with only tilde",
			nodeID:      "gateway~",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Minimal valid gateway format",
		},
		{
			name:        "router with only tilde",
			nodeID:      "router~",
			expected:    v1alpha1.ProxyMode_GATEWAY,
			description: "Minimal valid router format",
		},
		{
			name:        "whitespace node ID",
			nodeID:      "   \t\n  ",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Whitespace-only should be unknown",
		},
		{
			name:        "kubernetes pod name format",
			nodeID:      "my-app-deployment-7d8f9c8b6d-abc12",
			expected:    v1alpha1.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Kubernetes pod name should be unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := inferProxyMode(test.nodeID)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}
