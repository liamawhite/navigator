// Copyright (c) 2025 Navigator Authors
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

package wasmplugin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name       string
		wasmPlugin *typesv1alpha1.WasmPlugin
		instance   *backendv1alpha1.ServiceInstance
		namespace  string
		expected   bool
	}{
		{
			name: "empty selector matches all workloads in same namespace",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "test-wasm",
				Namespace: "production",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "production",
			expected:  true,
		},
		{
			name: "empty selector does not match workloads in different namespace",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "test-wasm",
				Namespace: "production",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "root namespace wasm plugin with empty selector matches all workloads",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "global-wasm",
				Namespace: "istio-system",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "production",
			expected:  true,
		},
		{
			name: "selector matches with exact labels",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "app-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "version": "v1"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "selector matches with subset of labels",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "app-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1", "tier": "frontend"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "selector does not match with missing labels",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "app-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "version": "v2"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "selector does not match with different label values",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "app-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "api"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "selector does not match across namespaces",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "app-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "empty workload labels do not match non-empty selector",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "app-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "nil workload labels do not match non-empty selector",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "app-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: nil},
			namespace: "production",
			expected:  false,
		},
		{
			name: "empty match labels in selector matches all workloads",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "catch-all-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "wasm plugin with targetRefs delegates to targetRefs matching (no context = no match)",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "gateway-wasm",
				Namespace: "production",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "my-gateway",
					},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "wasm plugin with targetRefs and selector uses targetRefs (ignores selector)",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "mixed-wasm",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"}, // would match
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "my-gateway", // won't match without context
					},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "production",
			expected:  false, // targetRefs takes precedence
		},
		{
			name: "custom root namespace wasm plugin with empty selector does not match with default root",
			wasmPlugin: &typesv1alpha1.WasmPlugin{
				Name:      "global-wasm",
				Namespace: "custom-istio-root",
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace: "production",
			expected:  false,
		},
	}

	// First run tests with default istio-system root namespace
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWorkload(tt.wasmPlugin, tt.instance, tt.namespace, "istio-system")
			assert.Equal(t, tt.expected, result, "matchesWorkload result should match expected value")
		})
	}

	// Test with custom root namespace
	t.Run("custom root namespace wasm plugin matches with custom root", func(t *testing.T) {
		wasmPlugin := &typesv1alpha1.WasmPlugin{
			Name:      "global-wasm",
			Namespace: "custom-istio-root",
		}
		instance := &backendv1alpha1.ServiceInstance{
			Labels: map[string]string{"app": "test"},
		}
		result := matchesWorkload(wasmPlugin, instance, "production", "custom-istio-root")
		assert.True(t, result, "WasmPlugin in custom root namespace should match all workloads")
	})
}

// Note: TestMatchesWorkloadWithTargetRefs was removed because the API now uses ServiceInstance objects
// which don't contain gateway/service context information. The targetRefs functionality is tested
// indirectly through TestMatchesWorkload, though with limited context (empty services/gateways arrays).

func TestFilterWasmPluginsForWorkload(t *testing.T) {
	wasmPlugins := []*typesv1alpha1.WasmPlugin{
		{
			Name:      "all-workloads-wasm",
			Namespace: "production",
			Selector:  nil, // nil selector matches all workloads in namespace
		},
		{
			Name:      "app-specific-wasm",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "version-specific-wasm",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web", "version": "v1"},
			},
		},
		{
			Name:      "other-app-wasm",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "api"},
			},
		},
		{
			Name:      "different-namespace-wasm",
			Namespace: "staging",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "root-namespace-wasm",
			Namespace: "istio-system",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "empty-selector-wasm",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
		},
		{
			Name:      "targetrefs-wasm",
			Namespace: "production",
			TargetRefs: []*typesv1alpha1.PolicyTargetReference{
				{
					Group: "gateway.networking.k8s.io",
					Kind:  "Gateway",
					Name:  "my-gateway",
				},
			},
		},
	}

	tests := []struct {
		name                string
		instance            *backendv1alpha1.ServiceInstance
		workloadNamespace   string
		rootNamespace       string
		expectedWasmPlugins []string // WasmPlugin names that should match
	}{
		{
			name:              "workload matches multiple wasm plugins in same namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedWasmPlugins: []string{
				"all-workloads-wasm",
				"app-specific-wasm",
				"version-specific-wasm",
				"root-namespace-wasm",
				"empty-selector-wasm",
			},
		},
		{
			name:              "workload with partial labels matches subset",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedWasmPlugins: []string{
				"all-workloads-wasm",
				"app-specific-wasm",
				"root-namespace-wasm",
				"empty-selector-wasm",
			},
		},
		{
			name:              "workload with no matching labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "unrelated"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedWasmPlugins: []string{
				"all-workloads-wasm",
				"empty-selector-wasm",
			},
		},
		{
			name:              "workload with no labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedWasmPlugins: []string{
				"all-workloads-wasm",
				"empty-selector-wasm",
			},
		},
		{
			name:              "workload in different namespace does not match namespace-scoped wasm plugins",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "staging",
			rootNamespace:     "istio-system",
			expectedWasmPlugins: []string{
				"different-namespace-wasm",
				"root-namespace-wasm", // root namespace wasm plugins match across all namespaces
			},
		},
		{
			name:              "api workload matches only relevant wasm plugins",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "api"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedWasmPlugins: []string{
				"all-workloads-wasm",
				"other-app-wasm",
				"empty-selector-wasm",
			},
		},
		{
			name:              "workload with custom root namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "custom-istio-root",
			expectedWasmPlugins: []string{
				"all-workloads-wasm",
				"app-specific-wasm",
				"empty-selector-wasm",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterWasmPluginsForWorkload(wasmPlugins, tt.instance, tt.workloadNamespace, tt.rootNamespace)

			// Convert result to wasm plugin names for easier comparison
			var resultNames []string
			for _, wp := range result {
				resultNames = append(resultNames, wp.Name)
			}

			assert.ElementsMatch(t, tt.expectedWasmPlugins, resultNames, "Filtered WasmPlugins mismatch")
		})
	}
}

func TestFilterWasmPluginsForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty wasm plugin list returns empty result", func(t *testing.T) {
		result := FilterWasmPluginsForWorkload([]*typesv1alpha1.WasmPlugin{}, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for empty wasm plugin list")
	})

	t.Run("nil wasm plugin list returns empty result", func(t *testing.T) {
		result := FilterWasmPluginsForWorkload(nil, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for nil wasm plugin list")
	})
}
