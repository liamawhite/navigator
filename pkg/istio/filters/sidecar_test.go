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

func TestFilterSidecarsForWorkload(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "test"},
	}

	sidecars := []*typesv1alpha1.Sidecar{
		{
			Name:      "match-all",
			Namespace: "default",
		},
		{
			Name:      "match-workload",
			Namespace: "default",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		{
			Name:      "no-match-workload",
			Namespace: "default",
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "other"},
			},
		},
		{
			Name:      "cross-namespace",
			Namespace: "other",
		},
	}

	result := FilterSidecarsForWorkload(sidecars, instance, "default")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "match-all", result[0].Name)
	assert.Equal(t, "match-workload", result[1].Name)
}

func TestSidecarMatchesWorkload(t *testing.T) {
	tests := []struct {
		name          string
		sidecar       *typesv1alpha1.Sidecar
		instance      *backendv1alpha1.ServiceInstance
		namespace     string
		expectedMatch bool
	}{
		{
			name:          "nil sidecar should not match",
			sidecar:       nil,
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: false,
		},
		{
			name: "sidecar in different namespace should not match",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "other",
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: false,
		},
		{
			name: "sidecar with no workload selector matches all in same namespace",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "sidecar with empty workload selector matches all in same namespace",
			sidecar: &typesv1alpha1.Sidecar{
				Name:             "test-sidecar",
				Namespace:        "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{MatchLabels: map[string]string{}},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "sidecar workload selector matches instance labels",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test", "version": "v1"}},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "sidecar workload selector does not match instance labels",
			sidecar: &typesv1alpha1.Sidecar{
				Name:      "test-sidecar",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "other"}},
			namespace:     "default",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sidecarMatchesWorkload(tt.sidecar, tt.instance, tt.namespace)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}
