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

package sidecar

import (
	"testing"

	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name              string
		sidecar           *typesv1alpha1.Sidecar
		workloadLabels    map[string]string
		workloadNamespace string
		expectedMatch     bool
	}{
		{
			name:              "nil sidecar should not match",
			sidecar:           nil,
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "sidecar with nil workload selector matches all workloads in same namespace",
			sidecar: &typesv1alpha1.Sidecar{
				Name:             "test-sidecar",
				Namespace:        "default",
				WorkloadSelector: nil,
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "sidecar with empty workload selector matches all workloads in same namespace",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "sidecar with nil match_labels matches all workloads in same namespace",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: nil,
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "sidecar does not match workload in different namespace",
			sidecar: &typesv1alpha1.Sidecar{
				Name:             "test-sidecar",
				Namespace:        "istio-system",
				WorkloadSelector: nil,
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "sidecar selector matches workload labels exactly",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "sidecar selector matches subset of workload labels",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			workloadLabels:    map[string]string{"app": "test", "version": "v1", "env": "prod"},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "sidecar selector does not match when workload missing required label",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			workloadLabels:    map[string]string{"version": "v1"},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "sidecar selector does not match when label values differ",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			workloadLabels:    map[string]string{"app": "other"},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "multiple selector labels must all match",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test", "version": "v1"},
				},
			},
			workloadLabels:    map[string]string{"app": "test", "version": "v1", "env": "prod"},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "multiple selector labels fail when one does not match",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test", "version": "v1"},
				},
			},
			workloadLabels:    map[string]string{"app": "test", "version": "v2"},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "sidecar with workload selector does not match workload in different namespace even with matching labels",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "istio-system",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "workload with no labels matches sidecar with empty selector in same namespace",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{},
				},
			},
			workloadLabels:    map[string]string{},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "workload with no labels does not match sidecar with specific selector",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			workloadLabels:    map[string]string{},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesWorkload(tt.sidecar, tt.workloadLabels, tt.workloadNamespace)
			assert.Equal(t, tt.expectedMatch, result, "MatchesWorkload result mismatch")
		})
	}
}

func TestFilterSidecarsForWorkload(t *testing.T) {
	sidecars := []*typesv1alpha1.Sidecar{
		{
			Name:             "all-workloads-sidecar",
			Namespace:        "default",
			WorkloadSelector: nil, // nil selector matches all workloads in namespace
		},
		{
			Name:      "app-specific-sidecar",
			Namespace: "default",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		{
			Name:      "version-specific-sidecar",
			Namespace: "default",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "test", "version": "v1"},
			},
		},
		{
			Name:      "other-app-sidecar",
			Namespace: "default",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "other"},
			},
		},
		{
			Name:      "different-namespace-sidecar",
			Namespace: "production",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		{
			Name:      "empty-selector-sidecar",
			Namespace: "default",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
		},
	}

	tests := []struct {
		name              string
		workloadLabels    map[string]string
		workloadNamespace string
		expectedSidecars  []string // Sidecar names that should match
	}{
		{
			name:              "workload matches multiple sidecars in same namespace",
			workloadLabels:    map[string]string{"app": "test", "version": "v1"},
			workloadNamespace: "default",
			expectedSidecars: []string{
				"all-workloads-sidecar",
				"app-specific-sidecar",
				"version-specific-sidecar",
				"empty-selector-sidecar",
			},
		},
		{
			name:              "workload with partial labels matches subset",
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			expectedSidecars: []string{
				"all-workloads-sidecar",
				"app-specific-sidecar",
				"empty-selector-sidecar",
			},
		},
		{
			name:              "workload with no matching labels only matches empty/nil selectors",
			workloadLabels:    map[string]string{"app": "unrelated"},
			workloadNamespace: "default",
			expectedSidecars: []string{
				"all-workloads-sidecar",
				"empty-selector-sidecar",
			},
		},
		{
			name:              "workload with no labels only matches empty/nil selectors",
			workloadLabels:    map[string]string{},
			workloadNamespace: "default",
			expectedSidecars: []string{
				"all-workloads-sidecar",
				"empty-selector-sidecar",
			},
		},
		{
			name:              "workload in different namespace does not match any sidecars",
			workloadLabels:    map[string]string{"app": "test", "version": "v1"},
			workloadNamespace: "staging",
			expectedSidecars:  []string{},
		},
		{
			name:              "workload in production namespace matches only sidecars in that namespace",
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "production",
			expectedSidecars: []string{
				"different-namespace-sidecar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterSidecarsForWorkload(sidecars, tt.workloadLabels, tt.workloadNamespace)

			// Convert result to sidecar names for easier comparison
			var resultNames []string
			for _, sc := range result {
				resultNames = append(resultNames, sc.Name)
			}

			assert.ElementsMatch(t, tt.expectedSidecars, resultNames, "Filtered sidecars mismatch")
		})
	}
}

func TestFilterSidecarsForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty sidecar list returns empty result", func(t *testing.T) {
		result := FilterSidecarsForWorkload([]*typesv1alpha1.Sidecar{}, map[string]string{"app": "test"}, "default")
		assert.Empty(t, result, "Expected empty result for empty sidecar list")
	})

	t.Run("nil sidecar list returns empty result", func(t *testing.T) {
		result := FilterSidecarsForWorkload(nil, map[string]string{"app": "test"}, "default")
		assert.Empty(t, result, "Expected empty result for nil sidecar list")
	})
}
