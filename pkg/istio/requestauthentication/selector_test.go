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

package requestauthentication

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name                  string
		requestAuthentication *typesv1alpha1.RequestAuthentication
		instance              *backendv1alpha1.ServiceInstance
		namespace             string
		expected              bool
	}{
		{
			name: "empty selector matches all workloads in same namespace",
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "test-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "test-auth",
				Namespace: "production",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "root namespace auth with empty selector matches all workloads",
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "global-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "app-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "app-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "app-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "app-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "app-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "app-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "app-auth",
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
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "catch-all-auth",
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
			name: "request authentication with targetRefs delegates to targetRefs matching (no context = no match)",
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "gateway-auth",
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
			name: "request authentication with targetRefs and selector uses targetRefs (ignores selector)",
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "mixed-auth",
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
			name: "custom root namespace auth with empty selector does not match with default root",
			requestAuthentication: &typesv1alpha1.RequestAuthentication{
				Name:      "global-auth",
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
			result := matchesWorkload(tt.requestAuthentication, tt.instance, tt.namespace, "istio-system")
			assert.Equal(t, tt.expected, result, "matchesWorkload result should match expected value")
		})
	}

	// Test with custom root namespace
	t.Run("custom root namespace auth matches with custom root", func(t *testing.T) {
		requestAuthentication := &typesv1alpha1.RequestAuthentication{
			Name:      "global-auth",
			Namespace: "custom-istio-root",
		}
		instance := &backendv1alpha1.ServiceInstance{
			Labels: map[string]string{"app": "test"},
		}
		result := matchesWorkload(requestAuthentication, instance, "production", "custom-istio-root")
		assert.True(t, result, "RequestAuthentication in custom root namespace should match all workloads")
	})
}

// Note: TestMatchesWorkloadWithTargetRefs was removed because the API now uses ServiceInstance objects
// which don't contain gateway/service context information. The targetRefs functionality is tested
// indirectly through TestMatchesWorkload, though with limited context (empty services/gateways arrays).

func TestFilterRequestAuthenticationsForWorkload(t *testing.T) {
	requestAuthentications := []*typesv1alpha1.RequestAuthentication{
		{
			Name:      "all-workloads-auth",
			Namespace: "production",
			Selector:  nil, // nil selector matches all workloads in namespace
		},
		{
			Name:      "app-specific-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "version-specific-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web", "version": "v1"},
			},
		},
		{
			Name:      "other-app-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "api"},
			},
		},
		{
			Name:      "different-namespace-auth",
			Namespace: "staging",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "root-namespace-auth",
			Namespace: "istio-system",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "empty-selector-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
		},
		{
			Name:      "targetrefs-auth",
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
		name                           string
		instance                       *backendv1alpha1.ServiceInstance
		workloadNamespace              string
		rootNamespace                  string
		expectedRequestAuthentications []string // RequestAuthentication names that should match
	}{
		{
			name:              "workload matches multiple request authentications in same namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedRequestAuthentications: []string{
				"all-workloads-auth",
				"app-specific-auth",
				"version-specific-auth",
				"root-namespace-auth",
				"empty-selector-auth",
			},
		},
		{
			name:              "workload with partial labels matches subset",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedRequestAuthentications: []string{
				"all-workloads-auth",
				"app-specific-auth",
				"root-namespace-auth",
				"empty-selector-auth",
			},
		},
		{
			name:              "workload with no matching labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "unrelated"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedRequestAuthentications: []string{
				"all-workloads-auth",
				"empty-selector-auth",
			},
		},
		{
			name:              "workload with no labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedRequestAuthentications: []string{
				"all-workloads-auth",
				"empty-selector-auth",
			},
		},
		{
			name:              "workload in different namespace does not match namespace-scoped authentications",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "staging",
			rootNamespace:     "istio-system",
			expectedRequestAuthentications: []string{
				"different-namespace-auth",
				"root-namespace-auth", // root namespace authentications match across all namespaces
			},
		},
		{
			name:              "api workload matches only relevant authentications",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "api"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedRequestAuthentications: []string{
				"all-workloads-auth",
				"other-app-auth",
				"empty-selector-auth",
			},
		},
		{
			name:              "workload with custom root namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "custom-istio-root",
			expectedRequestAuthentications: []string{
				"all-workloads-auth",
				"app-specific-auth",
				"empty-selector-auth",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterRequestAuthenticationsForWorkload(requestAuthentications, tt.instance, tt.workloadNamespace, tt.rootNamespace)

			// Convert result to request authentication names for easier comparison
			var resultNames []string
			for _, ra := range result {
				resultNames = append(resultNames, ra.Name)
			}

			assert.ElementsMatch(t, tt.expectedRequestAuthentications, resultNames, "Filtered RequestAuthentications mismatch")
		})
	}
}

func TestFilterRequestAuthenticationsForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty request authentication list returns empty result", func(t *testing.T) {
		result := FilterRequestAuthenticationsForWorkload([]*typesv1alpha1.RequestAuthentication{}, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for empty request authentication list")
	})

	t.Run("nil request authentication list returns empty result", func(t *testing.T) {
		result := FilterRequestAuthenticationsForWorkload(nil, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for nil request authentication list")
	})
}
