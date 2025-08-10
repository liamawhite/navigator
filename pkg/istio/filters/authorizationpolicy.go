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

// authorizationPolicyMatchesWorkload determines if an AuthorizationPolicy applies to a specific workload based on its
// selector and targetRefs configuration.
//
// AuthorizationPolicy selection rules:
// 1. If selector is specified, it must match the workload's labels
// 2. If targetRefs is specified, the AuthorizationPolicy applies based on the referenced resources
// 3. If neither is specified, the AuthorizationPolicy applies to all workloads in the same namespace
// 4. If AuthorizationPolicy is in root namespace and has no selectors, it applies to all workloads
// 5. At most one of selector and targetRefs can be set
//
// This function handles both selector and targetRefs matching internally.
func authorizationPolicyMatchesWorkload(authorizationPolicy *typesv1alpha1.AuthorizationPolicy, instance *backendv1alpha1.ServiceInstance, namespace, rootNamespace string) bool {
	// Use default root namespace if not provided
	if rootNamespace == "" {
		rootNamespace = "istio-system"
	}

	// Extract workload labels from the service instance
	workloadLabels := instance.Labels
	if workloadLabels == nil {
		workloadLabels = make(map[string]string)
	}

	// If AuthorizationPolicy is in root namespace with no selectors, it applies to all workloads
	if authorizationPolicy.Namespace == rootNamespace &&
		(authorizationPolicy.Selector == nil || len(authorizationPolicy.Selector.MatchLabels) == 0) &&
		len(authorizationPolicy.TargetRefs) == 0 {
		return true
	}

	// AuthorizationPolicies are namespace-scoped unless in root namespace
	if authorizationPolicy.Namespace != rootNamespace && authorizationPolicy.Namespace != namespace {
		return false
	}

	// If targetRefs is specified, delegate to the targetRefs matching function
	// Note: Service and gateway context would need to be provided by the caller
	// For now, we use empty context as these are not available in ServiceInstance
	if len(authorizationPolicy.TargetRefs) > 0 {
		return authorizationPolicyMatchesWorkloadWithTargetRefs(
			authorizationPolicy,
			workloadLabels,
			namespace,
			rootNamespace,
			[]string{}, // empty services - would need service context from caller
			[]string{}, // empty gateways - would need gateway context from caller
		)
	}

	// If selector is not specified, match all workloads in same namespace
	if authorizationPolicy.Selector == nil || len(authorizationPolicy.Selector.MatchLabels) == 0 {
		return true
	}

	// Use common label selector matching
	return matchesLabelSelector(authorizationPolicy.Selector.MatchLabels, workloadLabels)
}

// authorizationPolicyMatchesWorkloadWithTargetRefs determines if an AuthorizationPolicy applies to a workload based on
// targetRefs configuration. This requires additional context about services, gateways, etc.
//
// Supported targetRefs types:
// - Gateway (gateway.networking.k8s.io) in same namespace
// - GatewayClass (gateway.networking.k8s.io) in root namespace
// - Service ("") in same namespace (waypoints only)
// - ServiceEntry (networking.istio.io) in same namespace
func authorizationPolicyMatchesWorkloadWithTargetRefs(
	authorizationPolicy *typesv1alpha1.AuthorizationPolicy,
	workloadLabels map[string]string,
	workloadNamespace string,
	rootNamespace string,
	workloadServices []string,
	workloadGateways []string,
) bool {
	// Use default root namespace if not provided
	if rootNamespace == "" {
		rootNamespace = "istio-system"
	}

	// If no targetRefs, fall back to selector matching
	if len(authorizationPolicy.TargetRefs) == 0 {
		// Create a temporary ServiceInstance for the recursive call
		tempInstance := &backendv1alpha1.ServiceInstance{Labels: workloadLabels}
		return authorizationPolicyMatchesWorkload(authorizationPolicy, tempInstance, workloadNamespace, rootNamespace)
	}

	// Check each targetRef to see if it applies to this workload
	for _, targetRef := range authorizationPolicy.TargetRefs {
		if targetRef == nil {
			continue
		}

		switch {
		case targetRef.Kind == "Gateway" && targetRef.Group == "gateway.networking.k8s.io":
			// Gateway must be in same namespace
			targetNamespace := targetRef.Namespace
			if targetNamespace == "" {
				targetNamespace = authorizationPolicy.Namespace
			}
			if targetNamespace == workloadNamespace {
				// Check if workload is associated with this gateway
				for _, gateway := range workloadGateways {
					if gateway == targetRef.Name {
						return true
					}
				}
			}

		case targetRef.Kind == "GatewayClass" && targetRef.Group == "gateway.networking.k8s.io":
			// GatewayClass must be in root namespace
			targetNamespace := targetRef.Namespace
			if targetNamespace == "" {
				targetNamespace = authorizationPolicy.Namespace
			}
			if targetNamespace == rootNamespace {
				// For GatewayClass, we'd need to check if any of the workload's gateways
				// use this GatewayClass. This requires additional context not available here.
				// For now, we conservatively return false.
				continue
			}

		case targetRef.Kind == "Service" && targetRef.Group == "":
			// Service must be in same namespace (waypoints only)
			targetNamespace := targetRef.Namespace
			if targetNamespace == "" {
				targetNamespace = authorizationPolicy.Namespace
			}
			if targetNamespace == workloadNamespace {
				// Check if workload is associated with this service
				for _, service := range workloadServices {
					if service == targetRef.Name {
						return true
					}
				}
			}

		case targetRef.Kind == "ServiceEntry" && targetRef.Group == "networking.istio.io":
			// ServiceEntry must be in same namespace
			targetNamespace := targetRef.Namespace
			if targetNamespace == "" {
				targetNamespace = authorizationPolicy.Namespace
			}
			if targetNamespace == workloadNamespace {
				// ServiceEntry matching requires checking if the workload communicates
				// with the external service defined by the ServiceEntry.
				// This requires additional context not available here.
				// For now, we conservatively return false.
				continue
			}
		}
	}

	// No targetRefs matched
	return false
}

// FilterAuthorizationPoliciesForWorkload returns all AuthorizationPolicies that apply to a specific workload instance.
// This is a convenience function that filters a list of AuthorizationPolicies based on the workload's
// labels and namespace, respecting both selector and targetRefs matching logic.
func FilterAuthorizationPoliciesForWorkload(authorizationPolicies []*typesv1alpha1.AuthorizationPolicy, instance *backendv1alpha1.ServiceInstance, workloadNamespace string, rootNamespace string) []*typesv1alpha1.AuthorizationPolicy {
	var matchingAuthorizationPolicies []*typesv1alpha1.AuthorizationPolicy

	for _, authorizationPolicy := range authorizationPolicies {
		if authorizationPolicyMatchesWorkload(authorizationPolicy, instance, workloadNamespace, rootNamespace) {
			matchingAuthorizationPolicies = append(matchingAuthorizationPolicies, authorizationPolicy)
		}
	}

	return matchingAuthorizationPolicies
}
