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

package peerauthentication

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name               string
		peerAuthentication *typesv1alpha1.PeerAuthentication
		instance           *backendv1alpha1.ServiceInstance
		namespace          string
		expected           bool
	}{
		{
			name: "empty selector matches all workloads in same namespace",
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "test-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "test-peer-auth",
				Namespace: "production",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "root namespace peer auth with empty selector matches all workloads",
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "global-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "app-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "app-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "app-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "app-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "app-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "app-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "app-peer-auth",
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
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "catch-all-peer-auth",
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
			name: "nil selector matches all workloads in same namespace",
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "nil-selector-peer-auth",
				Namespace: "production",
				Selector:  nil,
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "custom root namespace peer auth with empty selector does not match with default root",
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "global-peer-auth",
				Namespace: "custom-istio-root",
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace: "production",
			expected:  false,
		},
		{
			name: "workload selector with single app label matches",
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "single-label-peer-auth",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"tier": "backend"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "api", "tier": "backend", "version": "v2"}},
			namespace: "production",
			expected:  true,
		},
		{
			name: "workload selector with multiple labels requires all to match",
			peerAuthentication: &typesv1alpha1.PeerAuthentication{
				Name:      "multi-label-peer-auth",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "tier": "frontend", "version": "v1"},
				},
			},
			instance:  &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "tier": "frontend"}}, // missing version
			namespace: "production",
			expected:  false,
		},
	}

	// First run tests with default istio-system root namespace
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWorkload(tt.peerAuthentication, tt.instance, tt.namespace, "istio-system")
			assert.Equal(t, tt.expected, result, "matchesWorkload result should match expected value")
		})
	}

	// Test with custom root namespace
	t.Run("custom root namespace peer auth matches with custom root", func(t *testing.T) {
		peerAuthentication := &typesv1alpha1.PeerAuthentication{
			Name:      "global-peer-auth",
			Namespace: "custom-istio-root",
		}
		instance := &backendv1alpha1.ServiceInstance{
			Labels: map[string]string{"app": "test"},
		}
		result := matchesWorkload(peerAuthentication, instance, "production", "custom-istio-root")
		assert.True(t, result, "PeerAuthentication in custom root namespace should match all workloads")
	})
}

func TestFilterPeerAuthenticationsForWorkload(t *testing.T) {
	peerAuthentications := []*typesv1alpha1.PeerAuthentication{
		{
			Name:      "all-workloads-peer-auth",
			Namespace: "production",
			Selector:  nil, // nil selector matches all workloads in namespace
		},
		{
			Name:      "app-specific-peer-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "version-specific-peer-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web", "version": "v1"},
			},
		},
		{
			Name:      "other-app-peer-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "api"},
			},
		},
		{
			Name:      "different-namespace-peer-auth",
			Namespace: "staging",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "root-namespace-peer-auth",
			Namespace: "istio-system",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "empty-selector-peer-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
		},
		{
			Name:      "strict-mode-peer-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"security": "strict"},
			},
		},
		{
			Name:      "tier-based-peer-auth",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"tier": "backend"},
			},
		},
	}

	tests := []struct {
		name                        string
		instance                    *backendv1alpha1.ServiceInstance
		workloadNamespace           string
		rootNamespace               string
		expectedPeerAuthentications []string // PeerAuthentication names that should match
	}{
		{
			name:              "workload matches multiple peer authentications in same namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"app-specific-peer-auth",
				"version-specific-peer-auth",
				"root-namespace-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "workload with partial labels matches subset",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"app-specific-peer-auth",
				"root-namespace-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "workload with no matching labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "unrelated"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "workload with no labels only matches empty/nil selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "workload in different namespace does not match namespace-scoped peer authentications",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1"}},
			workloadNamespace: "staging",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"different-namespace-peer-auth",
				"root-namespace-peer-auth", // root namespace peer authentications match across all namespaces
			},
		},
		{
			name:              "api workload matches only relevant peer authentications",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "api"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"other-app-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "workload with custom root namespace",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web"}},
			workloadNamespace: "production",
			rootNamespace:     "custom-istio-root",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"app-specific-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "workload with security label matches strict mode",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "database", "security": "strict"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"strict-mode-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "backend tier workload matches tier-based peer auth",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "api", "tier": "backend", "version": "v2"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"other-app-peer-auth",
				"tier-based-peer-auth",
				"empty-selector-peer-auth",
			},
		},
		{
			name:              "workload with complex labels matches multiple specific selectors",
			instance:          &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "web", "version": "v1", "tier": "backend", "security": "strict"}},
			workloadNamespace: "production",
			rootNamespace:     "istio-system",
			expectedPeerAuthentications: []string{
				"all-workloads-peer-auth",
				"app-specific-peer-auth",
				"version-specific-peer-auth",
				"root-namespace-peer-auth",
				"strict-mode-peer-auth",
				"tier-based-peer-auth",
				"empty-selector-peer-auth",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterPeerAuthenticationsForWorkload(peerAuthentications, tt.instance, tt.workloadNamespace, tt.rootNamespace)

			// Convert result to peer authentication names for easier comparison
			var resultNames []string
			for _, pa := range result {
				resultNames = append(resultNames, pa.Name)
			}

			assert.ElementsMatch(t, tt.expectedPeerAuthentications, resultNames, "Filtered PeerAuthentications mismatch")
		})
	}
}

func TestFilterPeerAuthenticationsForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty peer authentication list returns empty result", func(t *testing.T) {
		result := FilterPeerAuthenticationsForWorkload([]*typesv1alpha1.PeerAuthentication{}, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for empty peer authentication list")
	})

	t.Run("nil peer authentication list returns empty result", func(t *testing.T) {
		result := FilterPeerAuthenticationsForWorkload(nil, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "default", "istio-system")
		assert.Empty(t, result, "Expected empty result for nil peer authentication list")
	})
}

func TestFilterPeerAuthenticationsForWorkload_EdgeCases(t *testing.T) {
	t.Run("nil service instance labels handled gracefully", func(t *testing.T) {
		peerAuthentications := []*typesv1alpha1.PeerAuthentication{
			{
				Name:      "nil-selector-peer-auth",
				Namespace: "production",
				Selector:  nil,
			},
			{
				Name:      "app-specific-peer-auth",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
		}

		result := FilterPeerAuthenticationsForWorkload(peerAuthentications, &backendv1alpha1.ServiceInstance{Labels: nil}, "production", "istio-system")

		var resultNames []string
		for _, pa := range result {
			resultNames = append(resultNames, pa.Name)
		}

		assert.ElementsMatch(t, []string{"nil-selector-peer-auth"}, resultNames, "Only nil selector should match workload with nil labels")
	})

	t.Run("empty root namespace uses default", func(t *testing.T) {
		peerAuthentications := []*typesv1alpha1.PeerAuthentication{
			{
				Name:      "root-peer-auth",
				Namespace: "istio-system", // default root namespace
				Selector:  nil,
			},
		}

		result := FilterPeerAuthenticationsForWorkload(peerAuthentications, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "production", "")

		assert.Len(t, result, 1, "Should match with default root namespace")
		assert.Equal(t, "root-peer-auth", result[0].Name)
	})

	t.Run("workload selector with nil match labels treated as empty selector", func(t *testing.T) {
		peerAuthentications := []*typesv1alpha1.PeerAuthentication{
			{
				Name:      "nil-match-labels-peer-auth",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: nil,
				},
			},
		}

		result := FilterPeerAuthenticationsForWorkload(peerAuthentications, &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}}, "production", "istio-system")

		assert.Len(t, result, 1, "Nil match labels should be treated as empty selector and match all workloads")
		assert.Equal(t, "nil-match-labels-peer-auth", result[0].Name)
	})
}
