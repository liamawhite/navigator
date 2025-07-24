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
		name      string
		peerAuth  *typesv1alpha1.PeerAuthentication
		instance  *backendv1alpha1.ServiceInstance
		namespace string
		rootNS    string
		expected  bool
	}{
		{
			name: "empty selector matches all workloads in same namespace",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "test-pa",
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
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "test-pa",
				Namespace: "production",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "test"},
			},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "selector matches workload with correct labels",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "app-pa",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web", "version": "v1"},
			},
			namespace: "production",
			expected:  true,
		},
		{
			name: "selector matches workload with exact labels",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "app-version-pa",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "version": "v1"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web", "version": "v1"},
			},
			namespace: "production",
			expected:  true,
		},
		{
			name: "selector does not match workload with missing labels",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "app-version-pa",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web", "version": "v2"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web", "version": "v1"},
			},
			namespace: "production",
			expected:  false,
		},
		{
			name: "selector does not match workload with different label values",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "app-pa",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "api"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web"},
			},
			namespace: "production",
			expected:  false,
		},
		{
			name: "selector does not match across namespaces",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "app-pa",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web"},
			},
			namespace: "staging",
			expected:  false,
		},
		{
			name: "empty workload labels do not match non-empty selector",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "app-pa",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{},
			},
			namespace: "production",
			expected:  false,
		},
		{
			name: "nil workload labels do not match non-empty selector",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "app-pa",
				Namespace: "production",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: nil,
			},
			namespace: "production",
			expected:  false,
		},
		{
			name: "root namespace policy with no selector applies globally",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "global-pa",
				Namespace: "istio-system",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web"},
			},
			namespace: "production",
			rootNS:    "istio-system",
			expected:  true,
		},
		{
			name: "root namespace policy with empty selector applies globally",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "global-pa",
				Namespace: "istio-system",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web"},
			},
			namespace: "production",
			rootNS:    "istio-system",
			expected:  true,
		},
		{
			name: "root namespace policy with selector applies with matching labels",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "global-app-pa",
				Namespace: "istio-system",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "web"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web", "version": "v1"},
			},
			namespace: "production",
			rootNS:    "istio-system",
			expected:  true,
		},
		{
			name: "root namespace policy with selector does not apply with non-matching labels",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "global-app-pa",
				Namespace: "istio-system",
				Selector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "api"},
				},
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web"},
			},
			namespace: "production",
			rootNS:    "istio-system",
			expected:  false,
		},
		{
			name: "custom root namespace policy applies globally",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "global-pa",
				Namespace: "custom-root",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web"},
			},
			namespace: "production",
			rootNS:    "custom-root",
			expected:  true,
		},
		{
			name: "regular namespace policy does not apply from custom root namespace",
			peerAuth: &typesv1alpha1.PeerAuthentication{
				Name:      "regular-pa",
				Namespace: "istio-system",
			},
			instance: &backendv1alpha1.ServiceInstance{
				Labels: map[string]string{"app": "web"},
			},
			namespace: "production",
			rootNS:    "custom-root",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootNS := tt.rootNS
			if rootNS == "" {
				rootNS = "istio-system"
			}
			result := matchesWorkload(tt.peerAuth, tt.instance, tt.namespace, rootNS)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterPeerAuthenticationsForWorkload(t *testing.T) {
	peerAuths := []*typesv1alpha1.PeerAuthentication{
		{
			Name:      "global-pa",
			Namespace: "istio-system",
		},
		{
			Name:      "namespace-pa",
			Namespace: "production",
		},
		{
			Name:      "app-pa",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "web"},
			},
		},
		{
			Name:      "version-pa",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"version": "v1"},
			},
		},
		{
			Name:      "different-ns-pa",
			Namespace: "staging",
		},
	}

	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "web", "version": "v1"},
	}

	result := FilterPeerAuthenticationsForWorkload(peerAuths, instance, "production", "istio-system")

	// Should match: global-pa, namespace-pa, app-pa, version-pa
	// Should NOT match: different-ns-pa
	expectedNames := []string{"global-pa", "namespace-pa", "app-pa", "version-pa"}
	assert.Len(t, result, len(expectedNames))

	actualNames := make([]string, len(result))
	for i, pa := range result {
		actualNames[i] = pa.Name
	}

	for _, expectedName := range expectedNames {
		assert.Contains(t, actualNames, expectedName)
	}
}

func TestFilterPeerAuthenticationsForWorkload_EmptyList(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "web"},
	}

	result := FilterPeerAuthenticationsForWorkload([]*typesv1alpha1.PeerAuthentication{}, instance, "production", "istio-system")

	assert.Empty(t, result)
}

func TestFilterPeerAuthenticationsForWorkload_NoMatches(t *testing.T) {
	peerAuths := []*typesv1alpha1.PeerAuthentication{
		{
			Name:      "app-pa",
			Namespace: "production",
			Selector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "api"},
			},
		},
		{
			Name:      "different-ns-pa",
			Namespace: "staging",
		},
	}

	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "web"},
	}

	result := FilterPeerAuthenticationsForWorkload(peerAuths, instance, "production", "istio-system")

	assert.Empty(t, result)
}
