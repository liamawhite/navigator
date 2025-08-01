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

// matchesWorkloadSelector determines if a destination rule's workload selector
// matches the specific workload instance labels
func destinationRuleMatchesWorkloadSelector(dr *typesv1alpha1.DestinationRule, instance *backendv1alpha1.ServiceInstance) bool {
	if dr == nil || instance == nil {
		return false
	}

	// If no workload selector is specified, the destination rule applies to all workloads
	if dr.WorkloadSelector == nil || len(dr.WorkloadSelector.MatchLabels) == 0 {
		return true
	}

	// Check if the workload instance has labels
	if instance.Labels == nil {
		return false
	}

	// All labels in the workload selector must match the workload instance labels
	for key, value := range dr.WorkloadSelector.MatchLabels {
		if instanceValue, exists := instance.Labels[key]; !exists || instanceValue != value {
			return false
		}
	}

	return true
}

// destinationRuleMatchesHost determines if a destination rule applies to services that this workload might communicate with.
// This is a placeholder for future host-based filtering - currently returns true to include all destination rules
// regardless of their host field.
func destinationRuleMatchesHost(dr *typesv1alpha1.DestinationRule, instance *backendv1alpha1.ServiceInstance) bool {
	// TODO: Implement host-based filtering in the future
	// For now, we include all destination rules regardless of their host field
	return true
}

// destinationRuleMatchesWorkload determines if a destination rule applies to a specific workload instance.
// It implements Istio's destination rule visibility and applicability logic by checking:
// 1. Namespace visibility (exportTo field)
// 2. Workload selector matching (workloadSelector field)
// 3. Host matching (placeholder for future implementation)
func destinationRuleMatchesWorkload(dr *typesv1alpha1.DestinationRule, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
	if dr == nil || instance == nil {
		return false
	}

	// Stage 1: Check namespace visibility
	if !isVisibleToNamespace(destinationRuleExporter(dr), workloadNamespace) {
		return false
	}

	// Stage 2: Check workload selector matching
	if !destinationRuleMatchesWorkloadSelector(dr, instance) {
		return false
	}

	// Stage 3: Check host matching (placeholder for future implementation)
	if !destinationRuleMatchesHost(dr, instance) {
		return false
	}

	return true
}

// FilterDestinationRulesForWorkload returns all destination rules that apply to a specific workload instance.
// This filters destination rules based on namespace visibility (exportTo field) and workload selector matching
// to show only those destination rules that are relevant to the workload.
func FilterDestinationRulesForWorkload(destinationRules []*typesv1alpha1.DestinationRule, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) []*typesv1alpha1.DestinationRule {
	var matchingDestinationRules []*typesv1alpha1.DestinationRule

	for _, dr := range destinationRules {
		if destinationRuleMatchesWorkload(dr, instance, workloadNamespace) {
			matchingDestinationRules = append(matchingDestinationRules, dr)
		}
	}

	return matchingDestinationRules
}