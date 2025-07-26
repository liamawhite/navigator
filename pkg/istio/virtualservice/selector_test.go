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

package virtualservice

import (
	"testing"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestIsVisibleToNamespace(t *testing.T) {
	tests := []struct {
		name              string
		vs                *typesv1alpha1.VirtualService
		workloadNamespace string
		expectedVisible   bool
	}{
		{
			name:              "nil virtual service should not be visible",
			vs:                nil,
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name: "empty exportTo defaults to visible to all namespaces",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "nil exportTo defaults to visible to all namespaces",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  nil,
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with * makes visible to all namespaces",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{"*"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . makes visible only to same namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			workloadNamespace: "production",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . not visible to different namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name: "exportTo with specific namespace makes visible to that namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with specific namespace not visible to other namespaces",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			workloadNamespace: "staging",
			expectedVisible:   false,
		},
		{
			name: "exportTo with multiple namespaces including workload namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			workloadNamespace: "staging",
			expectedVisible:   true,
		},
		{
			name: "exportTo with multiple namespaces not including workload namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			workloadNamespace: "development",
			expectedVisible:   false,
		},
		{
			name: "exportTo with . and specific namespace - visible to same namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "production",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . and specific namespace - visible to specific namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . and specific namespace - not visible to other namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "staging",
			expectedVisible:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVisibleToNamespace(tt.vs, tt.workloadNamespace)
			assert.Equal(t, tt.expectedVisible, result, "isVisibleToNamespace result mismatch")
		})
	}
}

func TestIsGatewayWorkload(t *testing.T) {
	tests := []struct {
		name            string
		instance        *backendv1alpha1.ServiceInstance
		expectedGateway bool
	}{
		{
			name:            "nil instance should not be gateway",
			instance:        nil,
			expectedGateway: false,
		},
		{
			name: "instance with no proxy type should not be gateway",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_UNSPECIFIED,
			},
			expectedGateway: false,
		},
		{
			name: "standard istio ingress gateway",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
			},
			expectedGateway: true,
		},
		{
			name: "istio-proxy with gateway type",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
			},
			expectedGateway: true,
		},
		{
			name: "sidecar proxy type",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
			},
			expectedGateway: false,
		},
		{
			name: "explicit gateway proxy type",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
			},
			expectedGateway: true,
		},
		{
			name: "no proxy type",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_NONE,
			},
			expectedGateway: false,
		},
		{
			name: "regular sidecar workload",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
			},
			expectedGateway: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGatewayWorkload(tt.instance)
			assert.Equal(t, tt.expectedGateway, result, "isGatewayWorkload result mismatch")
		})
	}
}

func TestAppliesToWorkloadTraffic(t *testing.T) {
	tests := []struct {
		name              string
		vs                *typesv1alpha1.VirtualService
		instance          *backendv1alpha1.ServiceInstance
		workloadNamespace string
		expectedApplies   bool
	}{
		{
			name:              "nil virtual service should not apply",
			vs:                nil,
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedApplies:   false,
		},
		{
			name: "empty gateways defaults to mesh traffic for sidecar",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedApplies:   true,
		},
		{
			name: "nil gateways defaults to mesh traffic for sidecar",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: nil,
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedApplies:   true,
		},
		{
			name: "gateways with mesh applies to sidecar workload",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{"mesh"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedApplies:   true,
		},
		{
			name: "gateways with specific gateway name does not apply to sidecar workload",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{"my-gateway"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedApplies:   false,
		},
		{
			name: "gateways with mesh and specific gateway applies to sidecar workload",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{"mesh", "my-gateway"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedApplies:   true,
		},
		{
			name: "gateways with specific gateway applies to matching gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{"istio-ingressgateway"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedApplies:   true,
		},
		{
			name: "gateways with namespaced gateway applies to matching gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{"istio-system/istio-ingressgateway"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedApplies:   true,
		},
		{
			name: "gateways with different gateway does not apply to gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{"other-gateway"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedApplies:   false,
		},
		{
			name: "gateways with mesh does not apply to gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Gateways: []string{"mesh"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedApplies:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appliesToWorkloadTraffic(tt.vs, tt.instance, tt.workloadNamespace)
			assert.Equal(t, tt.expectedApplies, result, "appliesToWorkloadTraffic result mismatch")
		})
	}
}

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name              string
		vs                *typesv1alpha1.VirtualService
		instance          *backendv1alpha1.ServiceInstance
		workloadNamespace string
		expectedMatch     bool
	}{
		{
			name:              "nil virtual service should not match",
			vs:                nil,
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "nil instance should not match",
			vs: &typesv1alpha1.VirtualService{
				Name:     "test-vs",
				Hosts:    []string{"productpage"},
				ExportTo: []string{"*"},
				Gateways: []string{"mesh"},
			},
			instance:          nil,
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "complete match - visible and mesh traffic",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{"*"},
				Gateways:  []string{"mesh"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "not visible to namespace",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{"."},
				Gateways:  []string{"mesh"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "does not apply to sidecar workload (gateway-only virtual service)",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{"*"},
				Gateways:  []string{"my-gateway"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "applies to matching gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "istio-system",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{"*"},
				Gateways:  []string{"istio-ingressgateway"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedMatch:     true,
		},
		{
			name: "does not apply to non-matching gateway workload",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "istio-system",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{"*"},
				Gateways:  []string{"other-gateway"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedMatch:     false,
		},
		{
			name: "defaults work - empty exportTo and gateways",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{}, // defaults to ["*"]
				Gateways:  []string{}, // defaults to ["mesh"]
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "same namespace with dot export",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "default",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{"."},
				Gateways:  []string{"mesh"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "mixed gateways including mesh",
			vs: &typesv1alpha1.VirtualService{
				Name:      "test-vs",
				Namespace: "production",
				Hosts:     []string{"productpage"},
				ExportTo:  []string{"*"},
				Gateways:  []string{"my-gateway", "mesh"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWorkload(tt.vs, tt.instance, tt.workloadNamespace)
			assert.Equal(t, tt.expectedMatch, result, "matchesWorkload result mismatch")
		})
	}
}

func TestFilterVirtualServicesForWorkload(t *testing.T) {
	virtualServices := []*typesv1alpha1.VirtualService{
		{
			Name:      "global-productpage-vs",
			Namespace: "istio-system",
			Hosts:     []string{"productpage"},
			ExportTo:  []string{"*"},
			Gateways:  []string{"mesh"},
		},
		{
			Name:      "local-productpage-vs",
			Namespace: "default",
			Hosts:     []string{"productpage"},
			ExportTo:  []string{"."},
			Gateways:  []string{"mesh"},
		},
		{
			Name:      "reviews-vs",
			Namespace: "default",
			Hosts:     []string{"reviews"},
			ExportTo:  []string{"*"},
			Gateways:  []string{"mesh"},
		},
		{
			Name:      "gateway-only-vs",
			Namespace: "default",
			Hosts:     []string{"productpage"},
			ExportTo:  []string{"*"},
			Gateways:  []string{"my-gateway"},
		},
		{
			Name:      "ingress-gateway-vs",
			Namespace: "istio-system",
			Hosts:     []string{"productpage"},
			ExportTo:  []string{"*"},
			Gateways:  []string{"istio-ingressgateway"},
		},
		{
			Name:      "specific-namespace-vs",
			Namespace: "production",
			Hosts:     []string{"productpage"},
			ExportTo:  []string{"default", "staging"},
			Gateways:  []string{"mesh"},
		},
		{
			Name:      "wildcard-host-vs",
			Namespace: "default",
			Hosts:     []string{"*.example.com"},
			ExportTo:  []string{"*"},
			Gateways:  []string{"mesh"},
		},
	}

	tests := []struct {
		name                    string
		instance                *backendv1alpha1.ServiceInstance
		workloadNamespace       string
		expectedVirtualServices []string // VS names that should match
	}{
		{
			name:              "workload in default namespace",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedVirtualServices: []string{
				"global-productpage-vs",
				"local-productpage-vs",
				"reviews-vs",
				"specific-namespace-vs",
				"wildcard-host-vs",
				// gateway-only-vs excluded because it doesn't apply to mesh traffic
			},
		},
		{
			name:              "workload in staging namespace",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "staging",
			expectedVirtualServices: []string{
				"global-productpage-vs",
				"specific-namespace-vs",
				"reviews-vs",
				"wildcard-host-vs",
				// local-productpage-vs excluded because it's only exported to same namespace
				// gateway-only-vs excluded because it doesn't apply to mesh traffic
			},
		},
		{
			name:              "workload in production namespace (no access to local-only)",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "production",
			expectedVirtualServices: []string{
				"global-productpage-vs",
				"reviews-vs",
				"wildcard-host-vs",
				// local-productpage-vs excluded (local to default namespace)
				// specific-namespace-vs excluded (not exported to production)
				// gateway-only-vs excluded (doesn't apply to mesh traffic)
				// ingress-gateway-vs excluded (doesn't apply to mesh traffic)
			},
		},
		{
			name: "istio ingress gateway workload",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedVirtualServices: []string{
				"ingress-gateway-vs",
				// Only virtual services that reference istio-ingressgateway in gateways
				// All mesh-only virtual services are excluded
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterVirtualServicesForWorkload(virtualServices, tt.instance, tt.workloadNamespace)

			// Convert result to VS names for easier comparison
			var resultNames []string
			for _, vs := range result {
				resultNames = append(resultNames, vs.Name)
			}

			assert.ElementsMatch(t, tt.expectedVirtualServices, resultNames, "Filtered virtual services mismatch")
		})
	}
}

func TestFilterVirtualServicesForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty virtual service list returns empty result", func(t *testing.T) {
		result := FilterVirtualServicesForWorkload([]*typesv1alpha1.VirtualService{}, &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR}, "default")
		assert.Empty(t, result, "Expected empty result for empty virtual service list")
	})

	t.Run("nil virtual service list returns empty result", func(t *testing.T) {
		result := FilterVirtualServicesForWorkload(nil, &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR}, "default")
		assert.Empty(t, result, "Expected empty result for nil virtual service list")
	})
}
