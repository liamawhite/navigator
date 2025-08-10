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

func TestFilterAuthorizationPoliciesForWorkload(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "test"},
	}

	authorizationPolicies := []*typesv1alpha1.AuthorizationPolicy{
		{
			Name:      "root-namespace-no-selector",
			Namespace: "istio-system",
		},
		{
			Name:      "same-namespace-no-selector",
			Namespace: "default",
		},
		{
			Name:      "match-workload",
			Namespace: "default",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		{
			Name:      "no-match-workload",
			Namespace: "default",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "other"},
			},
		},
	}

	result := FilterAuthorizationPoliciesForWorkload(authorizationPolicies, instance, "default", "istio-system")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "root-namespace-no-selector", result[0].Name)
	assert.Equal(t, "same-namespace-no-selector", result[1].Name)
	assert.Equal(t, "match-workload", result[2].Name)
}

func TestAuthorizationPolicyMatchesWorkload(t *testing.T) {
	tests := []struct {
		name          string
		ap            *typesv1alpha1.AuthorizationPolicy
		instance      *backendv1alpha1.ServiceInstance
		namespace     string
		rootNamespace string
		expectedMatch bool
	}{
		{
			name: "root namespace with no selectors applies to all workloads",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:       "root-ap",
				Namespace:  "istio-system",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "same namespace with no selectors applies to all workloads in namespace",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:       "namespace-ap",
				Namespace:  "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "different namespace (not root) does not apply",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:       "other-ap",
				Namespace:  "other",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "selector matches workload labels",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "selector-ap",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test", "version": "v1"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "selector does not match workload labels",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "selector-ap",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "other"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "selector with multiple labels matches workload with superset of labels",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "multi-selector-ap",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test", "version": "v1"},
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test", "version": "v1", "tier": "backend"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "selector with multiple labels does not match when one label is missing",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "multi-selector-ap",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test", "version": "v1"},
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test", "version": "v2"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "targetRefs with non-matching gateway does not apply",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "targetref-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "non-matching-gateway",
					},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "targetRefs with non-matching service does not apply",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "targetref-service-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "",
						Kind:  "Service",
						Name:  "non-matching-service",
					},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "uses default root namespace when not provided",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:       "default-root-ap",
				Namespace:  "istio-system",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "",
			expectedMatch: true,
		},
		{
			name: "instance with nil labels matches policy with no selector",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:       "nil-labels-ap",
				Namespace:  "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: nil},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "instance with nil labels does not match policy with selector",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "nil-labels-selector-ap",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: nil},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "root namespace policy with selector matches across namespaces",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "root-selector-ap",
				Namespace: "istio-system",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "production",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authorizationPolicyMatchesWorkload(tt.ap, tt.instance, tt.namespace, tt.rootNamespace)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}

func TestAuthorizationPolicyMatchesWorkloadWithTargetRefs(t *testing.T) {
	tests := []struct {
		name              string
		ap                *typesv1alpha1.AuthorizationPolicy
		workloadLabels    map[string]string
		workloadNamespace string
		rootNamespace     string
		workloadServices  []string
		workloadGateways  []string
		expectedMatch     bool
	}{
		{
			name: "no targetRefs falls back to selector matching - no selector matches all",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:       "no-targetrefs-ap",
				Namespace:  "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{},
			workloadGateways:  []string{},
			expectedMatch:     true,
		},
		{
			name: "no targetRefs falls back to selector matching - selector matches",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "no-targetrefs-selector-ap",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{},
			workloadGateways:  []string{},
			expectedMatch:     true,
		},
		{
			name: "gateway targetRef matches when workload has matching gateway",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "gateway-targetref-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "test-gateway",
					},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{},
			workloadGateways:  []string{"test-gateway", "other-gateway"},
			expectedMatch:     true,
		},
		{
			name: "gateway targetRef does not match when workload has no matching gateway",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "gateway-targetref-no-match-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "test-gateway",
					},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{},
			workloadGateways:  []string{"other-gateway"},
			expectedMatch:     false,
		},
		{
			name: "service targetRef matches when workload has matching service",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "service-targetref-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "",
						Kind:  "Service",
						Name:  "test-service",
					},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{"test-service", "other-service"},
			workloadGateways:  []string{},
			expectedMatch:     true,
		},
		{
			name: "service targetRef does not match when workload has no matching service",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "service-targetref-no-match-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "",
						Kind:  "Service",
						Name:  "test-service",
					},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{"other-service"},
			workloadGateways:  []string{},
			expectedMatch:     false,
		},
		{
			name: "multiple targetRefs - matches when any one matches",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "multi-targetref-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					{
						Group: "",
						Kind:  "Service",
						Name:  "non-matching-service",
					},
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "test-gateway",
					},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{"other-service"},
			workloadGateways:  []string{"test-gateway"},
			expectedMatch:     true,
		},
		{
			name: "nil targetRef is ignored",
			ap: &typesv1alpha1.AuthorizationPolicy{
				Name:      "nil-targetref-ap",
				Namespace: "default",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{
					nil,
					{
						Group: "",
						Kind:  "Service",
						Name:  "test-service",
					},
				},
			},
			workloadLabels:    map[string]string{"app": "test"},
			workloadNamespace: "default",
			rootNamespace:     "istio-system",
			workloadServices:  []string{"test-service"},
			workloadGateways:  []string{},
			expectedMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authorizationPolicyMatchesWorkloadWithTargetRefs(
				tt.ap,
				tt.workloadLabels,
				tt.workloadNamespace,
				tt.rootNamespace,
				tt.workloadServices,
				tt.workloadGateways,
			)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}