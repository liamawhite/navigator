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

func TestFilterRequestAuthenticationsForWorkload(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "test"},
	}

	requestAuthentications := []*typesv1alpha1.RequestAuthentication{
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

	result := FilterRequestAuthenticationsForWorkload(requestAuthentications, instance, "default", "istio-system")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "root-namespace-no-selector", result[0].Name)
	assert.Equal(t, "same-namespace-no-selector", result[1].Name)
	assert.Equal(t, "match-workload", result[2].Name)
}

func TestRequestAuthenticationMatchesWorkload(t *testing.T) {
	tests := []struct {
		name          string
		ra            *typesv1alpha1.RequestAuthentication
		instance      *backendv1alpha1.ServiceInstance
		namespace     string
		rootNamespace string
		expectedMatch bool
	}{
		{
			name: "root namespace with no selectors applies to all workloads",
			ra: &typesv1alpha1.RequestAuthentication{
				Name:       "root-ra",
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
			ra: &typesv1alpha1.RequestAuthentication{
				Name:       "namespace-ra",
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
			ra: &typesv1alpha1.RequestAuthentication{
				Name:       "other-ra",
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
			ra: &typesv1alpha1.RequestAuthentication{
				Name:      "selector-ra",
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
			ra: &typesv1alpha1.RequestAuthentication{
				Name:      "selector-ra",
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
			name: "targetRefs with non-matching gateway does not apply",
			ra: &typesv1alpha1.RequestAuthentication{
				Name:      "targetref-ra",
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
			name: "uses default root namespace when not provided",
			ra: &typesv1alpha1.RequestAuthentication{
				Name:       "default-root-ra",
				Namespace:  "istio-system",
				TargetRefs: []*typesv1alpha1.PolicyTargetReference{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "",
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := requestAuthenticationMatchesWorkload(tt.ra, tt.instance, tt.namespace, tt.rootNamespace)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}