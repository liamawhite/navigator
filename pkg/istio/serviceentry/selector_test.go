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

package serviceentry

import (
	"testing"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestIsVisibleToNamespace(t *testing.T) {
	tests := []struct {
		name              string
		se                *typesv1alpha1.ServiceEntry
		workloadNamespace string
		expectedVisible   bool
	}{
		{
			name:              "nil service entry should not be visible",
			se:                nil,
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name: "empty exportTo defaults to visible to all namespaces",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "nil exportTo defaults to visible to all namespaces",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  nil,
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with * makes visible to all namespaces",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"*"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . makes visible only to same namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			workloadNamespace: "production",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . not visible to different namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name: "exportTo with specific namespace makes visible to that namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with specific namespace not visible to other namespaces",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			workloadNamespace: "staging",
			expectedVisible:   false,
		},
		{
			name: "exportTo with multiple namespaces including workload namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			workloadNamespace: "staging",
			expectedVisible:   true,
		},
		{
			name: "exportTo with multiple namespaces not including workload namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			workloadNamespace: "development",
			expectedVisible:   false,
		},
		{
			name: "exportTo with . and specific namespace - visible to same namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "production",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . and specific namespace - visible to specific namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name: "exportTo with . and specific namespace - not visible to other namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{".", "default"},
			},
			workloadNamespace: "staging",
			expectedVisible:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVisibleToNamespace(tt.se, tt.workloadNamespace)
			assert.Equal(t, tt.expectedVisible, result, "isVisibleToNamespace result mismatch")
		})
	}
}

func TestMatchesWorkload(t *testing.T) {
	tests := []struct {
		name              string
		se                *typesv1alpha1.ServiceEntry
		instance          *backendv1alpha1.ServiceInstance
		workloadNamespace string
		expectedMatch     bool
	}{
		{
			name:              "nil service entry should not match",
			se:                nil,
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "nil instance should not match",
			se: &typesv1alpha1.ServiceEntry{
				Name:     "test-se",
				ExportTo: []string{"*"},
			},
			instance:          nil,
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "service entry visible to workload namespace - should match",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "httpbin-external",
				Namespace: "production",
				ExportTo:  []string{"*"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "service entry not visible to workload namespace - should not match",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "httpbin-external",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     false,
		},
		{
			name: "service entry in same namespace with dot export - should match",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "httpbin-external",
				Namespace: "default",
				ExportTo:  []string{"."},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "service entry with specific namespace export - should match",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "httpbin-external",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "service entry with specific namespace export - should not match excluded namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "httpbin-external",
				Namespace: "production",
				ExportTo:  []string{"default", "staging"},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "development",
			expectedMatch:     false,
		},
		{
			name: "service entry with empty exportTo defaults to global - should match",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "httpbin-external",
				Namespace: "production",
				ExportTo:  []string{},
			},
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
		{
			name: "service entry applies to gateway workloads too",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "external-api",
				Namespace: "istio-system",
				ExportTo:  []string{"*"},
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
			name: "service entry applies to workloads with no proxy type",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "external-api",
				Namespace: "production",
				ExportTo:  []string{"*"},
			},
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_NONE,
			},
			workloadNamespace: "default",
			expectedMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWorkload(tt.se, tt.instance, tt.workloadNamespace)
			assert.Equal(t, tt.expectedMatch, result, "matchesWorkload result mismatch")
		})
	}
}

func TestFilterServiceEntriesForWorkload(t *testing.T) {
	serviceEntries := []*typesv1alpha1.ServiceEntry{
		{
			Name:      "global-httpbin",
			Namespace: "istio-system",
			ExportTo:  []string{"*"},
		},
		{
			Name:      "local-database",
			Namespace: "default",
			ExportTo:  []string{"."},
		},
		{
			Name:      "shared-service",
			Namespace: "default",
			ExportTo:  []string{"*"},
		},
		{
			Name:      "team-specific-api",
			Namespace: "production",
			ExportTo:  []string{"default", "staging"},
		},
		{
			Name:      "external-payment-gateway",
			Namespace: "production",
			ExportTo:  []string{"*"},
		},
		{
			Name:      "internal-logging-service",
			Namespace: "monitoring",
			ExportTo:  []string{"monitoring"},
		},
		{
			Name:      "default-export-service",
			Namespace: "services",
			ExportTo:  []string{}, // defaults to ["*"]
		},
	}

	tests := []struct {
		name                   string
		instance               *backendv1alpha1.ServiceInstance
		workloadNamespace      string
		expectedServiceEntries []string // SE names that should match
	}{
		{
			name:              "workload in default namespace",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "default",
			expectedServiceEntries: []string{
				"global-httpbin",
				"local-database",
				"shared-service",
				"team-specific-api",
				"external-payment-gateway",
				"default-export-service",
				// internal-logging-service excluded (only exported to monitoring)
			},
		},
		{
			name:              "workload in staging namespace",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "staging",
			expectedServiceEntries: []string{
				"global-httpbin",
				"shared-service",
				"team-specific-api",
				"external-payment-gateway",
				"default-export-service",
				// local-database excluded (only exported to same namespace)
				// internal-logging-service excluded (only exported to monitoring)
			},
		},
		{
			name:              "workload in production namespace",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "production",
			expectedServiceEntries: []string{
				"global-httpbin",
				"shared-service",
				"external-payment-gateway",
				"default-export-service",
				// local-database excluded (only exported to default namespace)
				// team-specific-api excluded (not exported to production)
				// internal-logging-service excluded (only exported to monitoring)
			},
		},
		{
			name:              "workload in monitoring namespace",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "monitoring",
			expectedServiceEntries: []string{
				"global-httpbin",
				"shared-service",
				"external-payment-gateway",
				"internal-logging-service",
				"default-export-service",
				// local-database excluded (only exported to default namespace)
				// team-specific-api excluded (not exported to monitoring)
			},
		},
		{
			name:              "workload in isolated namespace",
			instance:          &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR},
			workloadNamespace: "isolated",
			expectedServiceEntries: []string{
				"global-httpbin",
				"shared-service",
				"external-payment-gateway",
				"default-export-service",
				// All namespace-specific exports excluded
			},
		},
		{
			name: "gateway workload has same visibility rules",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_GATEWAY,
				Labels: map[string]string{
					"istio": "ingressgateway",
				},
			},
			workloadNamespace: "istio-system",
			expectedServiceEntries: []string{
				"global-httpbin",
				"shared-service",
				"external-payment-gateway",
				"default-export-service",
				// All namespace-specific exports excluded from istio-system
			},
		},
		{
			name: "workload with no proxy type still gets service entries",
			instance: &backendv1alpha1.ServiceInstance{
				ProxyType: backendv1alpha1.ProxyType_NONE,
			},
			workloadNamespace: "default",
			expectedServiceEntries: []string{
				"global-httpbin",
				"local-database",
				"shared-service",
				"team-specific-api",
				"external-payment-gateway",
				"default-export-service",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterServiceEntriesForWorkload(serviceEntries, tt.instance, tt.workloadNamespace)

			// Convert result to SE names for easier comparison
			var resultNames []string
			for _, se := range result {
				resultNames = append(resultNames, se.Name)
			}

			assert.ElementsMatch(t, tt.expectedServiceEntries, resultNames, "Filtered service entries mismatch")
		})
	}
}

func TestFilterServiceEntriesForWorkload_EmptyInput(t *testing.T) {
	t.Run("empty service entry list returns empty result", func(t *testing.T) {
		result := FilterServiceEntriesForWorkload([]*typesv1alpha1.ServiceEntry{}, &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR}, "default")
		assert.Empty(t, result, "Expected empty result for empty service entry list")
	})

	t.Run("nil service entry list returns empty result", func(t *testing.T) {
		result := FilterServiceEntriesForWorkload(nil, &backendv1alpha1.ServiceInstance{ProxyType: backendv1alpha1.ProxyType_SIDECAR}, "default")
		assert.Empty(t, result, "Expected empty result for nil service entry list")
	})

	t.Run("nil instance returns empty result", func(t *testing.T) {
		serviceEntries := []*typesv1alpha1.ServiceEntry{
			{
				Name:     "test-se",
				ExportTo: []string{"*"},
			},
		}
		result := FilterServiceEntriesForWorkload(serviceEntries, nil, "default")
		assert.Empty(t, result, "Expected empty result for nil instance")
	})
}