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

func TestFilterDestinationRulesForWorkload(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "test"},
	}

	destinationRules := []*typesv1alpha1.DestinationRule{
		{
			Name:      "match-all",
			Namespace: "default",
			ExportTo:  []string{},
		},
		{
			Name:      "match-workload",
			Namespace: "default",
			ExportTo:  []string{},
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
		{
			Name:      "no-match-workload",
			Namespace: "default",
			ExportTo:  []string{},
			WorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "other"},
			},
		},
		{
			Name:      "no-match-namespace",
			Namespace: "other",
			ExportTo:  []string{"."},
		},
	}

	result := FilterDestinationRulesForWorkload(destinationRules, instance, "default")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "match-all", result[0].Name)
	assert.Equal(t, "match-workload", result[1].Name)
}

func TestDestinationRuleMatchesWorkloadSelector(t *testing.T) {
	tests := []struct {
		name            string
		dr              *typesv1alpha1.DestinationRule
		instance        *backendv1alpha1.ServiceInstance
		expectedMatch   bool
	}{
		{
			name:          "nil destination rule should not match",
			dr:            nil,
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			expectedMatch: false,
		},
		{
			name:          "nil instance should not match",
			dr:            &typesv1alpha1.DestinationRule{Name: "test"},
			instance:      nil,
			expectedMatch: false,
		},
		{
			name: "no workload selector matches all workloads",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "default",
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			expectedMatch: true,
		},
		{
			name: "empty workload selector matches all workloads",
			dr: &typesv1alpha1.DestinationRule{
				Name:             "test-dr",
				Namespace:        "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{MatchLabels: map[string]string{}},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			expectedMatch: true,
		},
		{
			name: "workload selector matches instance labels",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test", "version": "v1"}},
			expectedMatch: true,
		},
		{
			name: "workload selector does not match instance labels",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "other", "version": "v1"}},
			expectedMatch: false,
		},
		{
			name: "workload selector does not match when instance has nil labels",
			dr: &typesv1alpha1.DestinationRule{
				Name:      "test-dr",
				Namespace: "default",
				WorkloadSelector: &typesv1alpha1.WorkloadSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: nil},
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := destinationRuleMatchesWorkloadSelector(tt.dr, tt.instance)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}