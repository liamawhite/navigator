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

// gatewayMatchesWorkload determines if a gateway applies to a specific workload instance.
// It implements Istio's gateway selector matching logic, respecting namespace scoping
// based on the PILOT_SCOPE_GATEWAY_TO_NAMESPACE control plane configuration.
//
// Gateway selector matching rules:
// - If gateway selector is nil/empty, the gateway applies to all workloads
// - If gateway selector has labels, they must match the workload labels
// - If scopeToNamespace is true, gateway and workload must be in the same namespace
// - If scopeToNamespace is false (default), gateway can match workloads across namespaces
func gatewayMatchesWorkload(gateway *typesv1alpha1.Gateway, instance *backendv1alpha1.ServiceInstance, namespace string, scopeToNamespace bool) bool {
	if gateway == nil {
		return false
	}

	// Extract workload labels from the service instance
	workloadLabels := instance.Labels
	if workloadLabels == nil {
		workloadLabels = make(map[string]string)
	}

	// Check namespace scoping constraint
	if scopeToNamespace && gateway.Namespace != namespace {
		return false
	}

	// If gateway has no selector labels, it applies to all workloads
	if len(gateway.Selector) == 0 {
		return true
	}

	// Use common label selector matching
	return matchesLabelSelector(gateway.Selector, workloadLabels)
}

// FilterGatewaysForWorkload returns all gateways that apply to a specific workload instance.
// This is a convenience function that filters a list of gateways based on the workload's
// labels and namespace, respecting the PILOT_SCOPE_GATEWAY_TO_NAMESPACE setting.
func FilterGatewaysForWorkload(gateways []*typesv1alpha1.Gateway, instance *backendv1alpha1.ServiceInstance, workloadNamespace string, scopeToNamespace bool) []*typesv1alpha1.Gateway {
	var matchingGateways []*typesv1alpha1.Gateway

	for _, gateway := range gateways {
		if gatewayMatchesWorkload(gateway, instance, workloadNamespace, scopeToNamespace) {
			matchingGateways = append(matchingGateways, gateway)
		}
	}

	return matchingGateways
}
