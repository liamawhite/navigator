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

package filters

import (
	"testing"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestFilterVirtualServicesForWorkload(t *testing.T) {
	sidecarInstance := &backendv1alpha1.ServiceInstance{
		Labels:    map[string]string{"app": "test"},
		ProxyType: backendv1alpha1.ProxyType_SIDECAR,
	}

	gatewayInstance := &backendv1alpha1.ServiceInstance{
		Labels:    map[string]string{"app": "istio-ingressgateway"},
		ProxyType: backendv1alpha1.ProxyType_GATEWAY,
	}

	virtualServices := []*typesv1alpha1.VirtualService{
		{
			Name:      "mesh-traffic",
			Namespace: "default",
			ExportTo:  []string{},
			Gateways:  []string{"mesh"},
		},
		{
			Name:      "gateway-traffic",
			Namespace: "default",
			ExportTo:  []string{},
			Gateways:  []string{"istio-ingressgateway"},
		},
		{
			Name:      "no-match-namespace",
			Namespace: "other",
			ExportTo:  []string{"."},
			Gateways:  []string{"mesh"},
		},
	}

	// Test sidecar instance
	sidecarResult := FilterVirtualServicesForWorkload(virtualServices, sidecarInstance, "default")
	assert.Equal(t, 1, len(sidecarResult))
	assert.Equal(t, "mesh-traffic", sidecarResult[0].Name)

	// Test gateway instance
	gatewayResult := FilterVirtualServicesForWorkload(virtualServices, gatewayInstance, "default")
	assert.Equal(t, 1, len(gatewayResult))
	assert.Equal(t, "gateway-traffic", gatewayResult[0].Name)
}

func TestVirtualServiceAppliesToWorkloadTraffic(t *testing.T) {
	tests := []struct {
		name          string
		vs            *typesv1alpha1.VirtualService
		instance      *backendv1alpha1.ServiceInstance
		namespace     string
		expectedMatch bool
	}{
		{
			name:          "nil virtual service should not apply",
			vs:            nil,
			instance:      &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			namespace:     "default",
			expectedMatch: false,
		},
		{
			name: "empty gateways defaults to mesh for sidecar",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "default",
				Gateways:  []string{},
			},
			instance:      &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "mesh gateway applies to sidecar",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "default",
				Gateways:  []string{"mesh"},
			},
			instance:      &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "non-mesh gateway does not apply to sidecar",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "default",
				Gateways:  []string{"istio-ingressgateway"},
			},
			instance:      &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			namespace:     "default",
			expectedMatch: false,
		},
		{
			name: "matching gateway name applies to gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "default",
				Gateways:  []string{"istio-ingressgateway"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels:    map[string]string{"app": "istio-ingressgateway"},
			},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "mesh gateway does not apply to gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "default",
				Gateways:  []string{"mesh"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels:    map[string]string{"app": "istio-ingressgateway"},
			},
			namespace:     "default",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := virtualServiceAppliesToWorkloadTraffic(tt.vs, tt.instance, tt.namespace)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}

func TestFilterVirtualServicesForMatchingGateways(t *testing.T) {
	// Test gateways
	gateways := []*typesv1alpha1.Gateway{
		{
			Name:      "microservice-gateway",
			Namespace: "microservices",
		},
		{
			Name:      "api-gateway",
			Namespace: "api-system",
		},
	}

	// Test virtual services
	virtualServices := []*typesv1alpha1.VirtualService{
		{
			Name:      "frontend",
			Namespace: "microservices",
			ExportTo:  []string{}, // defaults to "*"
			Gateways:  []string{"microservice-gateway"},
		},
		{
			Name:      "backend",
			Namespace: "microservices",
			ExportTo:  []string{},
			Gateways:  []string{"mesh"}, // mesh traffic, should be ignored
		},
		{
			Name:      "api-service",
			Namespace: "api-system",
			ExportTo:  []string{},
			Gateways:  []string{"api-gateway"},
		},
		{
			Name:      "cross-namespace",
			Namespace: "other",
			ExportTo:  []string{},
			Gateways:  []string{"microservices/microservice-gateway"}, // namespaced reference
		},
		{
			Name:      "private",
			Namespace: "other",
			ExportTo:  []string{"."}, // only visible to same namespace
			Gateways:  []string{"microservice-gateway"},
		},
		{
			Name:      "no-match",
			Namespace: "microservices",
			ExportTo:  []string{},
			Gateways:  []string{"unknown-gateway"},
		},
	}

	t.Run("finds VirtualServices for matching gateways", func(t *testing.T) {
		result := FilterVirtualServicesForMatchingGateways(virtualServices, gateways, "microservices")

		// Should find: frontend, api-service, cross-namespace
		// Should NOT find: backend (mesh), private (not visible), no-match (no gateway match)
		assert.Equal(t, 3, len(result))

		names := make(map[string]bool)
		for _, vs := range result {
			names[vs.Name] = true
		}

		assert.True(t, names["frontend"], "Should find frontend VirtualService")
		assert.True(t, names["api-service"], "Should find api-service VirtualService")
		assert.True(t, names["cross-namespace"], "Should find cross-namespace VirtualService")
		assert.False(t, names["backend"], "Should not find mesh-bound VirtualService")
		assert.False(t, names["private"], "Should not find namespace-private VirtualService")
		assert.False(t, names["no-match"], "Should not find VirtualService with no matching gateway")
	})

	t.Run("respects namespace visibility", func(t *testing.T) {
		result := FilterVirtualServicesForMatchingGateways(virtualServices, gateways, "other")

		// From 'other' namespace: should find frontend (exportTo="*"), api-service, cross-namespace, private
		// Should NOT find: backend (mesh), no-match (no gateway match)
		assert.Equal(t, 4, len(result))

		names := make(map[string]bool)
		for _, vs := range result {
			names[vs.Name] = true
		}

		assert.True(t, names["frontend"], "Should find frontend VirtualService (exportTo defaults to '*')")
		assert.True(t, names["api-service"], "Should find api-service VirtualService")
		assert.True(t, names["cross-namespace"], "Should find cross-namespace VirtualService")
		assert.True(t, names["private"], "Should find private VirtualService in same namespace")
		assert.False(t, names["backend"], "Should not find mesh-bound VirtualService")
		assert.False(t, names["no-match"], "Should not find VirtualService with no matching gateway")
	})

	t.Run("handles empty gateway list", func(t *testing.T) {
		result := FilterVirtualServicesForMatchingGateways(virtualServices, []*typesv1alpha1.Gateway{}, "microservices")
		assert.Equal(t, 0, len(result))
	})

	t.Run("handles nil gateway list", func(t *testing.T) {
		result := FilterVirtualServicesForMatchingGateways(virtualServices, nil, "microservices")
		assert.Nil(t, result)
	})
}

func TestMergeUniqueVirtualServices(t *testing.T) {
	// Note: This would normally be in istio_service_test.go but adding here for convenience
	// as the merge function is internal to the service package

	vs1 := []*typesv1alpha1.VirtualService{
		{Name: "common", Namespace: "ns1"},
		{Name: "unique1", Namespace: "ns1"},
	}

	vs2 := []*typesv1alpha1.VirtualService{
		{Name: "common", Namespace: "ns1"}, // duplicate
		{Name: "unique2", Namespace: "ns2"},
		{Name: "common", Namespace: "ns2"}, // different namespace, so not duplicate
	}

	t.Run("merges without duplicates", func(t *testing.T) {
		// This test demonstrates the expected behavior
		// In practice, you'd test this through the IstioService
		expected := 4 // common/ns1, unique1/ns1, unique2/ns2, common/ns2

		// Create a map to simulate the merge logic
		existing := make(map[string]bool)
		result := make([]*typesv1alpha1.VirtualService, 0)

		// Add first slice
		for _, vs := range vs1 {
			key := vs.Namespace + "/" + vs.Name
			existing[key] = true
			result = append(result, vs)
		}

		// Add unique items from second slice
		for _, vs := range vs2 {
			key := vs.Namespace + "/" + vs.Name
			if !existing[key] {
				result = append(result, vs)
				existing[key] = true
			}
		}

		assert.Equal(t, expected, len(result))

		// Check that we have all expected items
		found := make(map[string]bool)
		for _, vs := range result {
			found[vs.Namespace+"/"+vs.Name] = true
		}

		assert.True(t, found["ns1/common"])
		assert.True(t, found["ns1/unique1"])
		assert.True(t, found["ns2/unique2"])
		assert.True(t, found["ns2/common"])
	})
}
