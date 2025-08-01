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

func TestFilterPeerAuthenticationsForWorkload(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "test"},
	}

	peerAuthentications := []*typesv1alpha1.PeerAuthentication{
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

	result := FilterPeerAuthenticationsForWorkload(peerAuthentications, instance, "default", "istio-system")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "root-namespace-no-selector", result[0].Name)
	assert.Equal(t, "same-namespace-no-selector", result[1].Name)
	assert.Equal(t, "match-workload", result[2].Name)
}

func TestPeerAuthenticationMatchesWorkload(t *testing.T) {
	tests := []struct {
		name          string
		pa            *typesv1alpha1.PeerAuthentication
		instance      *backendv1alpha1.ServiceInstance
		namespace     string
		rootNamespace string
		expectedMatch bool
	}{
		{
			name: "root namespace with no selector applies to all workloads",
			pa: &typesv1alpha1.PeerAuthentication{
				Name:      "root-pa",
				Namespace: "istio-system",
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "same namespace with no selector applies to all workloads in namespace",
			pa: &typesv1alpha1.PeerAuthentication{
				Name:      "namespace-pa",
				Namespace: "default",
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "different namespace (not root) does not apply",
			pa: &typesv1alpha1.PeerAuthentication{
				Name:      "other-pa",
				Namespace: "other",
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "selector matches workload labels",
			pa: &typesv1alpha1.PeerAuthentication{
				Name:      "selector-pa",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test", "version": "v1"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: true,
		},
		{
			name: "selector does not match workload labels",
			pa: &typesv1alpha1.PeerAuthentication{
				Name:      "selector-pa",
				Namespace: "default",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "other"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "root namespace with selector only applies if selector matches",
			pa: &typesv1alpha1.PeerAuthentication{
				Name:      "root-selector-pa",
				Namespace: "istio-system",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "other"}},
			namespace:     "default",
			rootNamespace: "istio-system",
			expectedMatch: false,
		},
		{
			name: "uses default root namespace when not provided",
			pa: &typesv1alpha1.PeerAuthentication{
				Name:      "default-root-pa",
				Namespace: "istio-system",
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			rootNamespace: "",
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := peerAuthenticationMatchesWorkload(tt.pa, tt.instance, tt.namespace, tt.rootNamespace)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}