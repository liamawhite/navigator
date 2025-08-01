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
