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

	"github.com/stretchr/testify/assert"
)

func TestIsVisibleToNamespace(t *testing.T) {
	tests := []struct {
		name              string
		resource          ExporterResource
		workloadNamespace string
		expectedVisible   bool
	}{
		{
			name:              "nil resource should not be visible",
			resource:          nil,
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name:              "empty exportTo defaults to visible to all namespaces",
			resource:          newExportToResource("production", []string{}),
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name:              "nil exportTo defaults to visible to all namespaces",
			resource:          newExportToResource("production", nil),
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name:              "wildcard exportTo visible to all namespaces",
			resource:          newExportToResource("production", []string{"*"}),
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name:              "dot exportTo visible to same namespace only",
			resource:          newExportToResource("production", []string{"."}),
			workloadNamespace: "production",
			expectedVisible:   true,
		},
		{
			name:              "dot exportTo not visible to different namespace",
			resource:          newExportToResource("production", []string{"."}),
			workloadNamespace: "default",
			expectedVisible:   false,
		},
		{
			name:              "specific namespace exportTo visible to that namespace",
			resource:          newExportToResource("production", []string{"default"}),
			workloadNamespace: "default",
			expectedVisible:   true,
		},
		{
			name:              "specific namespace exportTo not visible to other namespace",
			resource:          newExportToResource("production", []string{"default"}),
			workloadNamespace: "staging",
			expectedVisible:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVisibleToNamespace(tt.resource, tt.workloadNamespace)
			assert.Equal(t, tt.expectedVisible, result)
		})
	}
}

func TestMatchesLabelSelector(t *testing.T) {
	tests := []struct {
		name           string
		selectorLabels map[string]string
		workloadLabels map[string]string
		expectedMatch  bool
	}{
		{
			name:           "empty selector matches all",
			selectorLabels: map[string]string{},
			workloadLabels: map[string]string{"app": "test"},
			expectedMatch:  true,
		},
		{
			name:           "nil selector matches all",
			selectorLabels: nil,
			workloadLabels: map[string]string{"app": "test"},
			expectedMatch:  true,
		},
		{
			name:           "selector matches workload labels",
			selectorLabels: map[string]string{"app": "test"},
			workloadLabels: map[string]string{"app": "test", "version": "v1"},
			expectedMatch:  true,
		},
		{
			name:           "selector does not match workload labels",
			selectorLabels: map[string]string{"app": "test"},
			workloadLabels: map[string]string{"app": "other", "version": "v1"},
			expectedMatch:  false,
		},
		{
			name:           "selector does not match when workload missing label",
			selectorLabels: map[string]string{"app": "test"},
			workloadLabels: map[string]string{"version": "v1"},
			expectedMatch:  false,
		},
		{
			name:           "selector does not match nil workload labels",
			selectorLabels: map[string]string{"app": "test"},
			workloadLabels: nil,
			expectedMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesLabelSelector(tt.selectorLabels, tt.workloadLabels)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}
