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
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// sidecarMatchesWorkload determines if a sidecar applies to a specific workload instance.
// It implements Istio's sidecar workload selector matching logic.
//
// Sidecar selector matching rules (from Istio documentation):
// - If workload selector is nil/empty, the sidecar applies to all workloads in the same namespace
// - If workload selector has match_labels, they must match the workload labels
// - Sidecars are always namespace-scoped (unlike gateways which can be scoped globally)
func sidecarMatchesWorkload(sidecar *typesv1alpha1.Sidecar, instance *backendv1alpha1.ServiceInstance, namespace string) bool {
	if sidecar == nil {
		return false
	}

	// Extract workload labels from the service instance
	workloadLabels := instance.Labels
	if workloadLabels == nil {
		workloadLabels = make(map[string]string)
	}

	// Sidecars are always namespace-scoped - they only apply to workloads in the same namespace
	if sidecar.Namespace != namespace {
		return false
	}

	// If sidecar has no workload selector, it applies to all workloads in the same namespace
	if sidecar.WorkloadSelector == nil || len(sidecar.WorkloadSelector.MatchLabels) == 0 {
		return true
	}

	// Use common label selector matching
	return matchesLabelSelector(sidecar.WorkloadSelector.MatchLabels, workloadLabels)
}

// FilterSidecarsForWorkload returns all sidecars that apply to a specific workload instance.
// This is a convenience function that filters a list of sidecars based on the workload's
// labels and namespace. Sidecars are always namespace-scoped.
func FilterSidecarsForWorkload(sidecars []*typesv1alpha1.Sidecar, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) []*typesv1alpha1.Sidecar {
	var matchingSidecars []*typesv1alpha1.Sidecar

	for _, sidecar := range sidecars {
		if sidecarMatchesWorkload(sidecar, instance, workloadNamespace) {
			matchingSidecars = append(matchingSidecars, sidecar)
		}
	}

	return matchingSidecars
}
