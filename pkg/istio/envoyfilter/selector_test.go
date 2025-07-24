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

package envoyfilter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name        string
		envoyFilter *typesv1alpha1.EnvoyFilter
		instance    *backendv1alpha1.ServiceInstance
		namespace   string
		expected    bool
	}{
		{
			name: "empty selector matches all workloads in same namespace",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "test-filter",
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
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "test-filter",
				Namespace: "production",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "root namespace filter with empty selector matches all workloads",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "global-filter",
				Namespace: "istio-system",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "production",
			expected:  true,
		},
		{
			name: "workload selector matches with exact labels",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "app-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "version": "v1"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "workload selector matches with subset of labels",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "app-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1", "tier": "frontend"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "workload selector does not match with missing labels",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "app-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "version": "v2"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "workload selector does not match with different label values",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "app-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "api"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "workload selector does not match across namespaces",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "app-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "empty workload labels do not match non-empty selector",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "app-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "nil workload labels do not match non-empty selector",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "app-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: nil},
			namespace: "production",
			expected:  false,
		},
		{
			name: "empty match labels in workload selector matches all workloads",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "catch-all-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "envoy filter with targetRefs delegates to targetRefs matching (no context = no match)",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "gateway-filter",
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
			name: "envoy filter with targetRefs and workloadSelector uses targetRefs (ignores workloadSelector)",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "mixed-filter",
				Namespace: "production",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
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
			name: "custom root namespace filter with empty selector does not match with default root",
			envoyFilter: &typesv1alpha1.EnvoyFilter{
				Name:      "global-filter",
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
			result := matchesWorkload(tt.envoyFilter, tt.instance, tt.namespace, "istio-system")
			assert.Equal(t, tt.expected, result, "matchesWorkload result should match expected value")
		})
	}

	// Test with custom root namespace
	t.Run("custom root namespace filter matches with custom root", func(t *testing.T) {
		envoyFilter := &typesv1alpha1.EnvoyFilter{
			Name:      "global-filter",
			Namespace: "custom-istio-root",
		}
		instance := &backendv1alpha1.ServiceInstance{
			Labels: map[string]string{"app": "test"},
		}
		result := matchesWorkload(envoyFilter, instance, "production", "custom-istio-root")
		assert.True(t, result, "EnvoyFilter in custom root namespace should match all workloads")
	})
}

// Note: TestMatchesWorkloadWithTargetRefs was removed because the API now uses ServiceInstance objects
// which don't contain gateway/service context information. The targetRefs functionality is tested
// indirectly through TestMatchesWorkload, though with limited context (empty services/gateways arrays).

func TestFilterEnvoyFiltersForWorkload(t *testing.T) {
	envoyFilters := []*typesv1alpha1.EnvoyFilter{
		{
			Name:             "all-workloads-filter",
			Namespace:        "production",
			WorkloadSelector: nil, // nil selector matches all workloads in namespace
		},
		{
			Name:      "app-specific-filter",
			Namespace: "production",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "version-specific-filter",
			Namespace: "production",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web", "version": "v1"},
			},
		},
		{
			Name:      "other-app-filter",
			Namespace: "production",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "api"},
			},
		},
		{
			Name:      "different-namespace-filter",
			Namespace: "staging",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "root-namespace-filter",
			Namespace: "istio-system",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "empty-selector-filter",
			Namespace: "production",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
		},
		{
			Name:      "targetrefs-filter",
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
		name                 string
		instance             *backendv1alpha1.ServiceInstance
		workloadNamespace    string
		rootNamespace        string
		expectedEnvoyFilters []string // EnvoyFilter names that should match
	}{
		{
			name:              "workload matches multiple envoyfilters in same namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedEnvoyFilters: []string{
				"all-workloads-filter",
				"app-specific-filter",
				"version-specific-filter",
				"root-namespace-filter",
				"empty-selector-filter",
			},
		},
		{
			name:              "workload with partial labels matches subset",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedEnvoyFilters: []string{
				"all-workloads-filter",
				"app-specific-filter",
				"root-namespace-filter",
				"empty-selector-filter",
			},
		},
		{
			name:              "workload with no matching labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "unrelated"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedEnvoyFilters: []string{
				"all-workloads-filter",
				"empty-selector-filter",
			},
		},
		{
			name:              "workload with no labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedEnvoyFilters: []string{
				"all-workloads-filter",
				"empty-selector-filter",
			},
		},
		{
			name:              "workload in different namespace does not match namespace-scoped filters",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "staging",
			rootNamespace:     "istio-system",
			expectedEnvoyFilters: []string{
				"different-namespace-filter",
				"root-namespace-filter", // root namespace filters match across all namespaces
			},
		},
		{
			name:              "api workload matches only relevant filters",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "api"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedEnvoyFilters: []string{
				"all-workloads-filter",
				"other-app-filter",
				"empty-selector-filter",
			},
		},
		{
			name:              "workload with custom root namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "custom-istio-root",
			expectedEnvoyFilters: []string{
				"all-workloads-filter",
				"app-specific-filter",
				"empty-selector-filter",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterEnvoyFiltersForWorkload(envoyFilters, tt.instance, tt.workloadNamespace, tt.rootNamespace)

			// Convert result to envoyfilter names for easier comparison
			var resultNames []string
			for _, ef := range result {
				resultNames = append(resultNames, ef.Name)
			}

			assert.ElementsMatch(t, tt.expectedEnvoyFilters, resultNames, "Filtered EnvoyFilters mismatch")
		})
	}
}

func TestFilterEnvoyFiltersForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty envoyfilter list returns empty result", func(t *testing.T) {
		result := FilterEnvoyFiltersForWorkload([]*typesv1alpha1.EnvoyFilter{}, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for empty envoyfilter list")
	})

	t.Run("nil envoyfilter list returns empty result", func(t *testing.T) {
		result := FilterEnvoyFiltersForWorkload(nil, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for nil envoyfilter list")
	})
}
