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

func TestFilterServiceEntriesForWorkload(t *testing.T) {
	instance := &backendv1alpha1.ServiceInstance{
		Labels: map[string]string{"app": "test"},
	}

	serviceEntries := []*typesv1alpha1.ServiceEntry{
		{
			Name:      "visible-all",
			Namespace: "default",
			ExportTo:  []string{},
		},
		{
			Name:      "visible-same-namespace",
			Namespace: "default",
			ExportTo:  []string{"."},
		},
		{
			Name:      "not-visible",
			Namespace: "other",
			ExportTo:  []string{"."},
		},
	}

	result := FilterServiceEntriesForWorkload(serviceEntries, instance, "default")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "visible-all", result[0].Name)
	assert.Equal(t, "visible-same-namespace", result[1].Name)
}

func TestServiceEntryMatchesWorkload(t *testing.T) {
	tests := []struct {
		name          string
		se            *typesv1alpha1.ServiceEntry
		instance      *backendv1alpha1.ServiceInstance
		namespace     string
		expectedMatch bool
	}{
		{
			name:          "nil service entry should not match",
			se:            nil,
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: false,
		},
		{
			name:          "nil instance should not match",
			se:            &typesv1alpha1.ServiceEntry{Name: "test-se", Namespace: "default"},
			instance:      nil,
			namespace:     "default",
			expectedMatch: false,
		},
		{
			name: "service entry with empty exportTo is visible to all namespaces",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "service entry with wildcard exportTo is visible to all namespaces",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"*"},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "service entry with dot exportTo is visible to same namespace only",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "default",
				ExportTo:  []string{"."},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "service entry with dot exportTo is not visible to different namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"."},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: false,
		},
		{
			name: "service entry with specific namespace exportTo is visible to that namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "default",
			expectedMatch: true,
		},
		{
			name: "service entry with specific namespace exportTo is not visible to other namespace",
			se: &typesv1alpha1.ServiceEntry{
				Name:      "test-se",
				Namespace: "production",
				ExportTo:  []string{"default"},
			},
			instance:      &backendv1alpha1.ServiceInstance{Labels: map[string]string{"app": "test"}},
			namespace:     "staging",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serviceEntryMatchesWorkload(tt.se, tt.instance, tt.namespace)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}
