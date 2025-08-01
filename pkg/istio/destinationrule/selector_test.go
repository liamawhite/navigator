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

package destinationrule

import (
	"testing"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestIsVisibleToNamespace(t *testing.T) {
	tests := []struct {
		name              string
		dr                *typesv1alpha1.DestinationRule
		workloadNamespace string
		expectedVisible   bool
	}{
		{
			name:              "nil destination rule should not be visible",
			dr:                nil,
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name: "empty exportTo defaults to visible to all namespaces",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "nil exportTo defaults to visible to all namespaces",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  nil,
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with * makes visible to all namespaces",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{"*"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . makes visible only to same namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			workloadNamespace: "production",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . not visible to different namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name: "exportTo with specific namespace makes visible to that namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with specific namespace not visible to other namespaces",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			workloadNamespace: "staging",
			expectedVisible:   false,
		},
		{
			name: "exportTo with multiple namespaces including workload namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			workloadNamespace: "staging",
			expectedVisible:   true,
		},
		{
			name: "exportTo with multiple namespaces not including workload namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			workloadNamespace: "development",
			expectedVisible:   false,
		},
		{
			name: "exportTo with . and specific namespace - visible to same namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "production",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . and specific namespace - visible to specific namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . and specific namespace - not visible to other namespace",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "staging",
			expectedVisible:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVisibleToNamespace(tt.dr, tt.workloadNamespace)
			assert.Equal(t, tt.expectedVisible, result, "isVisibleToNamespace result mismatch")
		})
	}
}

func TestMatchesWorkloadSelector(t *testing.T) {
	tests := []struct {
		name          string
		dr            *typesv1alpha1.DestinationRule
		instance      *backendv1alpha1.ServiceInstance
		expectedMatch bool
	}{
		{
			name:          "nil destination rule should not match",
			dr:            nil,
			instance:      &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			expectedMatch: false,
		},
		{
			name: "nil instance should not match",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      nil,
			expectedMatch: false,
		},
		{
			name: "no workload selector should match any workload",
			dr: &typesv1alpha1.DestinationRule{
				Name:             "test-dr",
				WorkloadSelector: nil,
			},
			instance:      &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			expectedMatch: true,
		},
		{
			name: "empty workload selector should match any workload",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			expectedMatch: true,
		},
		{
			name: "workload selector matches instance labels",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin", "version": "v1"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v1",
					"env":     "production",
				},
			},
			expectedMatch: true,
		},
		{
			name: "workload selector partial match fails",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin", "version": "v2"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v1",
					"env":     "production",
				},
			},
			expectedMatch: false,
		},
		{
			name: "workload selector key missing fails",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin", "team": "platform"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v1",
				},
			},
			expectedMatch: false,
		},
		{
			name: "instance with no labels fails non-empty selector",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels:    nil,
			},
			expectedMatch: false,
		},
		{
			name: "instance with empty labels fails non-empty selector",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels:    map[string]string{},
			},
			expectedMatch: false,
		},
		{
			name: "single label exact match",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app": "httpbin",
				},
			},
			expectedMatch: true,
		},
		{
			name: "gateway workload with matching selector",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"istio": "ingressgateway"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
					"app":   "istio-ingressgateway",
				},
			},
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWorkloadSelector(tt.dr, tt.instance)
			assert.Equal(t, tt.expectedMatch, result, "matchesWorkloadSelector result mismatch")
		})
	}
}

func TestMatchesHost(t *testing.T) {
	// Test that matchesHost currently returns true for all cases as it's a placeholder
	tests := []struct {
		name     string
		dr       *typesv1alpha1.DestinationRule
		instance *backendv1alpha1.ServiceInstance
	}{
		{
			name:     "nil destination rule returns true",
			dr:       nil,
			instance: &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
		},
		{
			name: "destination rule with any content returns true",
			dr: &typesv1alpha1.DestinationRule{
				Name: "test-dr",
			},
			instance: &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesHost(tt.dr, tt.instance)
			assert.True(t, result, "matchesHost should currently return true as it's a placeholder")
		})
	}
}

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name              string
		dr                *typesv1alpha1.DestinationRule
		instance          *backendv1alpha1.ServiceInstance
		workloadNamespace string
		expectedMatch     bool
	}{
		{
			name:              "nil destination rule should not match",
			dr:                nil,
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "nil instance should not match",
			dr: &typesv1alpha1.DestinationRule{
				Name:     "test-dr",
				ExportTo: []string{"*"},
			},
			instance:          nil,
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "visible destination rule with no workload selector should match",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "httpbin-dr",
				Namespace: "production",
				ExportTo:  []string{"*"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "non-visible destination rule should not match",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "httpbin-dr",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "visible destination rule with matching workload selector should match",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "httpbin-dr",
				Namespace: "default",
				ExportTo:  []string{"."},
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v1",
				},
			},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "visible destination rule with non-matching workload selector should not match",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "httpbin-dr",
				Namespace: "default",
				ExportTo:  []string{"."},
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "nginx"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v1",
				},
			},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "complex matching case with multiple criteria",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "complex-dr",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin", "version": "v1"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v1",
					"env":     "production",
				},
			},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "gateway workload matching",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "gateway-dr",
				Namespace: "istio-system",
				ExportTo:  []string{"*"},
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"istio": "ingressgateway"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
					"app":   "istio-ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWorkload(tt.dr, tt.instance, tt.workloadNamespace)
			assert.Equal(t, tt.expectedMatch, result, "matchesWorkload result mismatch")
		})
	}
}

func TestFilterDestinationRulesForWorkload(t *testing.T) {
	destinationRules := []*typesv1alpha1.DestinationRule{
		{
			Name:      "global-httpbin-dr",
			Namespace: "istio-system",
			ExportTo:  []string{"*"},
		},
		{
			Name:      "local-database-dr",
			Namespace: "default",
			ExportTo:  []string{"."},
		},
		{
			Name:      "shared-service-dr",
			Namespace: "default",
			ExportTo:  []string{"*"},
		},
		{
			Name:      "team-specific-dr",
			Namespace: "production",
			ExportTo:  []string{"default", "staging"},
		},
		{
			Name:      "workload-specific-dr",
			Namespace: "default",
			ExportTo:  []string{"."},
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "httpbin"},
			},
		},
		{
			Name:      "version-specific-dr",
			Namespace: "production",
			ExportTo:  []string{"*"},
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "httpbin", "version": "v2"},
			},
		},
		{
			Name:      "internal-monitoring-dr",
			Namespace: "monitoring",
			ExportTo:  []string{"monitoring"},
		},
		{
			Name:      "default-export-dr",
			Namespace: "services",
			ExportTo:  []string{}, // defaults to ["*"]
		},
		{
			Name:      "gateway-specific-dr",
			Namespace: "istio-system",
			ExportTo:  []string{"*"},
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"istio": "ingressgateway"},
			},
		},
	}

	tests := []struct {
		name                         string
		instance                     *backendv1alpha1.ServiceInstance
		workloadNamespace            string
		expectedDestinationRuleNames []string // DR names that should match
	}{
		{
			name: "httpbin workload in default namespace",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v1",
				},
			},
			workloadNamespace: "default",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"local-database-dr",
				"shared-service-dr",
				"team-specific-dr",
				"workload-specific-dr", // matches app=httpbin selector
				"default-export-dr",
				// version-specific-dr excluded (requires version=v2)
				// internal-monitoring-dr excluded (only exported to monitoring)
				// gateway-specific-dr excluded (requires gateway labels)
			},
		},
		{
			name: "httpbin v2 workload in staging namespace",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "httpbin",
					"version": "v2",
				},
			},
			workloadNamespace: "staging",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"shared-service-dr",
				"team-specific-dr",
				"version-specific-dr", // matches app=httpbin,version=v2 selector
				"default-export-dr",
				// local-database-dr excluded (only exported to default)
				// workload-specific-dr excluded (only exported to default)
				// internal-monitoring-dr excluded (only exported to monitoring)
				// gateway-specific-dr excluded (requires gateway labels)
			},
		},
		{
			name: "non-httpbin workload in default namespace",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app":     "nginx",
					"version": "latest",
				},
			},
			workloadNamespace: "default",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"local-database-dr",
				"shared-service-dr",
				"team-specific-dr",
				"default-export-dr",
				// workload-specific-dr excluded (requires app=httpbin)
				// version-specific-dr excluded (requires app=httpbin)
				// internal-monitoring-dr excluded (only exported to monitoring)
				// gateway-specific-dr excluded (requires gateway labels)
			},
		},
		{
			name: "gateway workload in istio-system",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
					"app":   "istio-ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"shared-service-dr",
				"default-export-dr",
				"gateway-specific-dr", // matches gateway selector
				// local-database-dr excluded (only exported to default)
				// team-specific-dr excluded (not exported to istio-system)
				// workload-specific-dr excluded (only exported to default)
				// version-specific-dr excluded (requires app=httpbin)
				// internal-monitoring-dr excluded (only exported to monitoring)
			},
		},
		{
			name: "workload in monitoring namespace",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app": "prometheus",
				},
			},
			workloadNamespace: "monitoring",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"shared-service-dr",
				"internal-monitoring-dr", // only visible to monitoring namespace
				"default-export-dr",
				// All other namespace-specific rules excluded
			},
		},
		{
			name: "workload in isolated namespace",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels: map[string]string{
					"app": "isolated-service",
				},
			},
			workloadNamespace: "isolated",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"shared-service-dr",
				"default-export-dr",
				// All namespace-specific and workload-specific rules excluded
			},
		},
		{
			name: "workload with no labels in default namespace",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_SIDECAR,
				Labels:    nil,
			},
			workloadNamespace: "default",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"local-database-dr",
				"shared-service-dr",
				"team-specific-dr",
				"default-export-dr",
				// All workload-selector rules excluded (no labels to match)
			},
		},
		{
			name: "workload with no proxy type",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_NONE,
				Labels: map[string]string{
					"app": "database",
				},
			},
			workloadNamespace: "default",
			expectedDestinationRuleNames: []string{
				"global-httpbin-dr",
				"local-database-dr",
				"shared-service-dr",
				"team-specific-dr",
				"default-export-dr",
				// workload-specific rules excluded (no matching labels)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterDestinationRulesForWorkload(destinationRules, tt.instance, tt.workloadNamespace)

			// Convert result to DR names for easier comparison
			var resultNames []string
			for _, dr := range result {
				resultNames = append(resultNames, dr.Name)
			}

			assert.ElementsMatch(t, tt.expectedDestinationRuleNames, resultNames, "Filtered destination rules mismatch")
		})
	}
}

func TestFilterDestinationRulesForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty destination rule list returns empty result", func(t *testing.T) {
		result := FilterDestinationRulesForWorkload([]*typesv1alpha1.DestinationRule{}, &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR}, "default")
		assert.Empty(t, result, "Expected empty result for empty destination rule list")
	})

	t.Run("nil destination rule list returns empty result", func(t *testing.T) {
		result := FilterDestinationRulesForWorkload(nil, &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR}, "default")
		assert.Empty(t, result, "Expected empty result for nil destination rule list")
	})

	t.Run("nil instance returns empty result", func(t *testing.T) {
		destinationRules := []*typesv1alpha1.DestinationRule{
			{
				Name:     "test-dr",
				ExportTo: []string{"*"},
			},
		}
		result := FilterDestinationRulesForWorkload(destinationRules, nil, "default")
		assert.Empty(t, result, "Expected empty result for nil instance")
	})
}
