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
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// isVisibleToNamespace determines if a service entry is visible to a specific namespace
// based on its exportTo field following Istio's visibility rules:
// - Empty exportTo defaults to ["*"] (visible to all namespaces)
// - "*" means visible to all namespaces
// - "." means visible only to the same namespace as the service entry
// - Specific namespace names mean visible only to those namespaces
func isVisibleToNamespace(se *typesv1alpha1.ServiceEntry, workloadNamespace string) bool {
	if se == nil {
		return false
	}

	// Empty exportTo defaults to ["*"] (visible to all namespaces)
	if len(se.ExportTo) == 0 {
		return true
	}

	for _, export := range se.ExportTo {
		if export == "*" {
			return true // visible to all namespaces
		}
		if export == "." && se.Namespace == workloadNamespace {
			return true // visible to same namespace
		}
		if export == workloadNamespace {
			return true // visible to specific namespace
		}
	}

	return false
}

// matchesWorkload determines if a service entry applies to a specific workload instance.
// It implements Istio's service entry visibility logic by checking namespace visibility (exportTo field).
// For now, we only filter based on exportTo - all visible service entries are considered applicable.
func matchesWorkload(se *typesv1alpha1.ServiceEntry, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
	if se == nil || instance == nil {
		return false
	}

	// Check namespace visibility based on exportTo field
	return isVisibleToNamespace(se, workloadNamespace)
}

// FilterServiceEntriesForWorkload returns all service entries that apply to a specific workload instance.
// This filters service entries based on namespace visibility (exportTo field) to show only those
// service entries that are visible to the workload's namespace.
func FilterServiceEntriesForWorkload(serviceEntries []*typesv1alpha1.ServiceEntry, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) []*typesv1alpha1.ServiceEntry {
	var matchingServiceEntries []*typesv1alpha1.ServiceEntry

	for _, se := range serviceEntries {
		if matchesWorkload(se, instance, workloadNamespace) {
			matchingServiceEntries = append(matchingServiceEntries, se)
		}
	}

	return matchingServiceEntries
}