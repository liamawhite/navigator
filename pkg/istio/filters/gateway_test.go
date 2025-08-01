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

func TestFilterGatewaysForWorkload(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "istio-ingressgateway"},
	}

	gateways := []*typesv1alpha1.Gateway{
		{
			Name:      "match-all",
			Namespace: "default",
			Selector:  map[string]string{},
		},
		{
			Name:      "match-workload",
			Namespace: "default",
			Selector:  map[string]string{"app": "istio-ingressgateway"},
		},
		{
			Name:      "no-match-workload",
			Namespace: "default",
			Selector:  map[string]string{"app": "other"},
		},
		{
			Name:      "cross-namespace",
			Namespace: "other",
			Selector:  map[string]string{"app": "istio-ingressgateway"},
		},
	}

	// Test without namespace scoping
	result := FilterGatewaysForWorkload(gateways, instance, "default", false)
	assert.Equal(t, 3, len(result))

	// Test with namespace scoping
	result = FilterGatewaysForWorkload(gateways, instance, "default", true)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "match-all", result[0].Name)
	assert.Equal(t, "match-workload", result[1].Name)
}

func TestGatewayMatchesWorkload(t *testing.T) {
	tests := []struct {
		name             string
		gateway          *typesv1alpha1.Gateway
		instance         *backendv1alpha1.ServiceInstance
		namespace        string
		scopeToNamespace bool
		expectedMatch    bool
	}{
		{
			name:             "nil gateway should not match",
			gateway:          nil,
			instance:         &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:        "default",
			scopeToNamespace: false,
			expectedMatch:    false,
		},
		{
			name: "empty selector matches all workloads",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{},
			},
			instance:         &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:        "default",
			scopeToNamespace: false,
			expectedMatch:    true,
		},
		{
			name: "nil selector matches all workloads",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  nil,
			},
			instance:         &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:        "default",
			scopeToNamespace: false,
			expectedMatch:    true,
		},
		{
			name: "selector matches workload labels",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
			instance:         &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test", "version": "v1"}},
			namespace:        "default",
			scopeToNamespace: false,
			expectedMatch:    true,
		},
		{
			name: "selector does not match workload labels",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
			instance:         &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "other"}},
			namespace:        "default",
			scopeToNamespace: false,
			expectedMatch:    false,
		},
		{
			name: "cross-namespace allowed when not scoped",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "other",
				Selector:  map[string]string{"app": "test"},
			},
			instance:         &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:        "default",
			scopeToNamespace: false,
			expectedMatch:    true,
		},
		{
			name: "cross-namespace not allowed when scoped",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "other",
				Selector:  map[string]string{"app": "test"},
			},
			instance:         &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:        "default",
			scopeToNamespace: true,
			expectedMatch:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gatewayMatchesWorkload(tt.gateway, tt.instance, tt.namespace, tt.scopeToNamespace)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}
