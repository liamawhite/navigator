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

package gateway

import (
	"testing"

	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name              string
		gateway           *typesv1alpha1.Gateway
		workloadLabels    map[string]string
		workloadNamespace string
		scopeToNamespace  bool
		expectedMatch     bool
	}{
		{
			name:              "nil gateway should not match",
			gateway:           nil,
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     false,
		},
		{
			name: "gateway with empty selector matches all workloads",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "istio-system",
				Selector:  map[string]string{},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     true,
		},
		{
			name: "gateway with nil selector matches all workloads",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "istio-system",
				Selector:  nil,
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     true,
		},
		{
			name: "gateway selector matches workload labels exactly",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     true,
		},
		{
			name: "gateway selector matches subset of workload labels",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
			workloadLabels:    map[string]string{"app": "test", "version": "v1", "env": "prod"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     true,
		},
		{
			name: "gateway selector does not match when workload missing required label",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
			workloadLabels:    map[string]string{"version": "v1"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     false,
		},
		{
			name: "gateway selector does not match when label values differ",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
			workloadLabels:    map[string]string{"app": "other"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     false,
		},
		{
			name: "multiple selector labels must all match",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test", "version": "v1"},
			},
			workloadLabels:    map[string]string{"app": "test", "version": "v1", "env": "prod"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     true,
		},
		{
			name: "multiple selector labels fail when one does not match",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test", "version": "v1"},
			},
			workloadLabels:    map[string]string{"app": "test", "version": "v2"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     false,
		},
		{
			name: "cross-namespace matching works when scoping disabled",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "istio-system",
				Selector:  map[string]string{"app": "test"},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedMatch:     true,
		},
		{
			name: "cross-namespace matching fails when scoping enabled",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "istio-system",
				Selector:  map[string]string{"app": "test"},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  true,
			expectedMatch:     false,
		},
		{
			name: "same-namespace matching works when scoping enabled",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{"app": "test"},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  true,
			expectedMatch:     true,
		},
		{
			name: "empty selector matches all workloads even with namespace scoping",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "istio-system",
				Selector:  map[string]string{},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  true,
			expectedMatch:     false, // should fail due to namespace mismatch
		},
		{
			name: "empty selector with same namespace and scoping enabled",
			gateway: &typesv1alpha1.Gateway{
				Name:      "test-gateway",
				Namespace: "default",
				Selector:  map[string]string{},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  true,
			expectedMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesWorkload(tt.gateway, tt.workloadLabels, tt.workloadNamespace, tt.scopeToNamespace)
			assert.Equal(t, tt.expectedMatch, result, "MatchesWorkload result mismatch")
		})
	}
}

func TestFilterGatewaysForWorkload(t *testing.T) {
	gateways := []*typesv1alpha1.Gateway{
		{
			Name:      "all-workloads-gateway",
			Namespace: "istio-system",
			Selector:  map[string]string{}, // Empty selector matches all
		},
		{
			Name:      "app-specific-gateway",
			Namespace: "default",
			Selector:  map[string]string{"app": "test"},
		},
		{
			Name:      "version-specific-gateway",
			Namespace: "default",
			Selector:  map[string]string{"app": "test", "version": "v1"},
		},
		{
			Name:      "other-app-gateway",
			Namespace: "default",
			Selector:  map[string]string{"app": "other"},
		},
		{
			Name:      "cross-namespace-gateway",
			Namespace: "production",
			Selector:  map[string]string{"app": "test"},
		},
	}

	tests := []struct {
		name              string
		workloadLabels    map[string]string
		workloadNamespace string
		scopeToNamespace  bool
		expectedGateways  []string // Gateway names that should match
	}{
		{
			name:              "workload matches multiple gateways without scoping",
			workloadLabels:    map[string]string{"app": "test", "version": "v1"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedGateways: []string{
				"all-workloads-gateway",
				"app-specific-gateway",
				"version-specific-gateway",
				"cross-namespace-gateway",
			},
		},
		{
			name:              "workload matches fewer gateways with namespace scoping",
			workloadLabels:    map[string]string{"app": "test", "version": "v1"},
			workloadNamespace: "default",
			scopeToNamespace:  true,
			expectedGateways: []string{
				"app-specific-gateway",
				"version-specific-gateway",
			},
		},
		{
			name:              "workload with partial labels matches subset",
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedGateways: []string{
				"all-workloads-gateway",
				"app-specific-gateway",
				"cross-namespace-gateway",
			},
		},
		{
			name:              "workload with no matching labels only matches empty selector",
			workloadLabels:    map[string]string{"app": "unrelated"},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedGateways: []string{
				"all-workloads-gateway",
			},
		},
		{
			name:              "workload with no labels only matches empty selector",
			workloadLabels:    map[string]string{},
			workloadNamespace: "default",
			scopeToNamespace:  false,
			expectedGateways: []string{
				"all-workloads-gateway",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterGatewaysForWorkload(gateways, tt.workloadLabels, tt.workloadNamespace, tt.scopeToNamespace)

			// Convert result to gateway names for easier comparison
			var resultNames []string
			for _, gw := range result {
				resultNames = append(resultNames, gw.Name)
			}

			assert.ElementsMatch(t, tt.expectedGateways, resultNames, "Filtered gateways mismatch")
		})
	}
}

func TestFilterGatewaysForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty gateway list returns empty result", func(t *testing.T) {
		result := FilterGatewaysForWorkload([]*typesv1alpha1.Gateway{}, map[string]string{"app": "test"}, "default", false)
		assert.Empty(t, result, "Expected empty result for empty gateway list")
	})

	t.Run("nil gateway list returns empty result", func(t *testing.T) {
		result := FilterGatewaysForWorkload(nil, map[string]string{"app": "test"}, "default", false)
		assert.Empty(t, result, "Expected empty result for nil gateway list")
	})
}
