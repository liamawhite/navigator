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
	"k8s.io/apimachinery/pkg/labels"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// MatchesWorkload determines if a PeerAuthentication applies to a specific workload based on its
// selector configuration.
//
// PeerAuthentication selection rules:
// 1. If selector is specified, it must match the workload's labels
// 2. If selector is not specified, the PeerAuthentication applies to all workloads in the same namespace
// 3. If PeerAuthentication is in root namespace and has no selector, it applies to all workloads
// 4. PeerAuthentications are namespace-scoped unless in root namespace
//
// This is simpler than RequestAuthentication as PeerAuthentication only supports WorkloadSelector,
// not targetRefs.
func matchesWorkload(peerAuthentication *typesv1alpha1.PeerAuthentication, instance *backendv1alpha1.ServiceInstance, namespace, rootNamespace string) bool {
	// Use default root namespace if not provided
	if rootNamespace == "" {
		rootNamespace = "istio-system"
	}

	// Extract workload labels from the service instance
	workloadLabels := instance.Labels
	if workloadLabels == nil {
		workloadLabels = make(map[string]string)
	}

	// If PeerAuthentication is in root namespace with no selector, it applies to all workloads
	if peerAuthentication.Namespace == rootNamespace &&
		(peerAuthentication.Selector == nil || len(peerAuthentication.Selector.MatchLabels) == 0) {
		return true
	}

	// PeerAuthentications are namespace-scoped unless in root namespace
	if peerAuthentication.Namespace != rootNamespace && peerAuthentication.Namespace != namespace {
		return false
	}

	// If selector is not specified, match all workloads in same namespace
	if peerAuthentication.Selector == nil || len(peerAuthentication.Selector.MatchLabels) == 0 {
		return true
	}

	// Use Kubernetes label selector matching for selector
	peerAuthenticationSelector := labels.Set(peerAuthentication.Selector.MatchLabels).AsSelector()
	workloadLabelSet := labels.Set(workloadLabels)
	return peerAuthenticationSelector.Matches(workloadLabelSet)
}

// FilterPeerAuthenticationsForWorkload returns all PeerAuthentications that apply to a specific workload instance.
// This is a convenience function that filters a list of PeerAuthentications based on the workload's
// labels and namespace, respecting the selector matching logic.
func FilterPeerAuthenticationsForWorkload(peerAuthentications []*typesv1alpha1.PeerAuthentication, instance *backendv1alpha1.ServiceInstance, workloadNamespace string, rootNamespace string) []*typesv1alpha1.PeerAuthentication {
	var matchingPeerAuthentications []*typesv1alpha1.PeerAuthentication

	for _, peerAuthentication := range peerAuthentications {
		if matchesWorkload(peerAuthentication, instance, workloadNamespace, rootNamespace) {
			matchingPeerAuthentications = append(matchingPeerAuthentications, peerAuthentication)
		}
	}

	return matchingPeerAuthentications
}
