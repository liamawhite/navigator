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

package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/istio/filters"
)

// ClusterStateProvider defines the interface for accessing cluster state
type ClusterStateProvider interface {
	GetClusterState(clusterID string) (*backendv1alpha1.ClusterState, error)
}

// IstioService handles Istio resource requests and filtering
type IstioService struct {
	clusterProvider ClusterStateProvider
	logger          *slog.Logger
}

// NewIstioService creates a new Istio service
func NewIstioService(clusterProvider ClusterStateProvider, logger *slog.Logger) *IstioService {
	return &IstioService{
		clusterProvider: clusterProvider,
		logger:          logger,
	}
}

// GetIstioResourcesForWorkload retrieves and filters Istio resources for a specific workload
func (i *IstioService) GetIstioResourcesForWorkload(ctx context.Context, clusterID, namespace string, instance *backendv1alpha1.ServiceInstance) (*frontendv1alpha1.GetIstioResourcesResponse, error) {
	i.logger.Debug("getting istio resources for workload",
		"cluster_id", clusterID,
		"namespace", namespace,
		"labels", instance.Labels)

	// Get cluster state
	clusterState, err := i.clusterProvider.GetClusterState(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster state for cluster %s: %w", clusterID, err)
	}

	// Determine namespace scoping from control plane config
	scopeToNamespace := false
	if clusterState.IstioControlPlaneConfig != nil {
		scopeToNamespace = clusterState.IstioControlPlaneConfig.PilotScopeGatewayToNamespace
	}

	// Determine root namespace for EnvoyFilter filtering
	rootNamespace := "istio-system" // default fallback
	if clusterState.IstioControlPlaneConfig != nil && clusterState.IstioControlPlaneConfig.RootNamespace != "" {
		rootNamespace = clusterState.IstioControlPlaneConfig.RootNamespace
	}

	// Parallelize filtering operations for better performance
	var wg sync.WaitGroup
	var matchingGateways []*typesv1alpha1.Gateway
	var matchingSidecars []*typesv1alpha1.Sidecar
	var matchingEnvoyFilters []*typesv1alpha1.EnvoyFilter
	var matchingRequestAuthentications []*typesv1alpha1.RequestAuthentication
	var matchingPeerAuthentications []*typesv1alpha1.PeerAuthentication
	var matchingAuthorizationPolicies []*typesv1alpha1.AuthorizationPolicy
	var matchingWasmPlugins []*typesv1alpha1.WasmPlugin
	var matchingVirtualServices []*typesv1alpha1.VirtualService
	var matchingServiceEntries []*typesv1alpha1.ServiceEntry
	var matchingDestinationRules []*typesv1alpha1.DestinationRule

	wg.Add(10)

	// Filter gateways concurrently
	go func() {
		defer wg.Done()
		matchingGateways = filters.FilterGatewaysForWorkload(clusterState.Gateways, instance, namespace, scopeToNamespace)
	}()

	// Filter sidecars concurrently
	go func() {
		defer wg.Done()
		matchingSidecars = filters.FilterSidecarsForWorkload(clusterState.Sidecars, instance, namespace)
	}()

	// Filter envoy filters concurrently
	go func() {
		defer wg.Done()
		matchingEnvoyFilters = filters.FilterEnvoyFiltersForWorkload(clusterState.EnvoyFilters, instance, namespace, rootNamespace)
	}()

	// Filter request authentications concurrently
	go func() {
		defer wg.Done()
		matchingRequestAuthentications = filters.FilterRequestAuthenticationsForWorkload(clusterState.RequestAuthentications, instance, namespace, rootNamespace)
	}()

	// Filter peer authentications concurrently
	go func() {
		defer wg.Done()
		matchingPeerAuthentications = filters.FilterPeerAuthenticationsForWorkload(clusterState.PeerAuthentications, instance, namespace, rootNamespace)
	}()

	// Filter authorization policies concurrently
	go func() {
		defer wg.Done()
		matchingAuthorizationPolicies = filters.FilterAuthorizationPoliciesForWorkload(clusterState.AuthorizationPolicies, instance, namespace, rootNamespace)
	}()

	// Filter wasm plugins concurrently
	go func() {
		defer wg.Done()
		matchingWasmPlugins = filters.FilterWasmPluginsForWorkload(clusterState.WasmPlugins, instance, namespace, rootNamespace)
	}()

	// Filter virtual services concurrently
	go func() {
		defer wg.Done()
		matchingVirtualServices = filters.FilterVirtualServicesForWorkload(clusterState.VirtualServices, instance, namespace)
	}()

	// Filter service entries concurrently
	go func() {
		defer wg.Done()
		matchingServiceEntries = filters.FilterServiceEntriesForWorkload(clusterState.ServiceEntries, instance, namespace)
	}()

	// Filter destination rules concurrently
	go func() {
		defer wg.Done()
		matchingDestinationRules = filters.FilterDestinationRulesForWorkload(clusterState.DestinationRules, instance, namespace)
	}()

	// Wait for all filtering operations to complete
	wg.Wait()

	// Additional filtering for gateway workloads: if we found matching gateways,
	// also find VirtualServices that reference those gateways
	if len(matchingGateways) > 0 {
		gatewayVirtualServices := filters.FilterVirtualServicesForMatchingGateways(clusterState.VirtualServices, matchingGateways, namespace)
		if len(gatewayVirtualServices) > 0 {
			originalCount := len(matchingVirtualServices)
			matchingVirtualServices = mergeUniqueVirtualServices(matchingVirtualServices, gatewayVirtualServices)
			i.logger.Debug("merged gateway-based virtual services",
				"original_count", originalCount,
				"gateway_vs_count", len(gatewayVirtualServices),
				"merged_count", len(matchingVirtualServices))
		}
	}

	i.logger.Debug("filtered istio resources",
		"cluster_id", clusterID,
		"total_gateways", len(clusterState.Gateways),
		"matching_gateways", len(matchingGateways),
		"total_sidecars", len(clusterState.Sidecars),
		"matching_sidecars", len(matchingSidecars),
		"total_envoyfilters", len(clusterState.EnvoyFilters),
		"matching_envoyfilters", len(matchingEnvoyFilters),
		"total_request_authentications", len(clusterState.RequestAuthentications),
		"matching_request_authentications", len(matchingRequestAuthentications),
		"total_peer_authentications", len(clusterState.PeerAuthentications),
		"matching_peer_authentications", len(matchingPeerAuthentications),
		"total_authorization_policies", len(clusterState.AuthorizationPolicies),
		"matching_authorization_policies", len(matchingAuthorizationPolicies),
		"total_wasm_plugins", len(clusterState.WasmPlugins),
		"matching_wasm_plugins", len(matchingWasmPlugins),
		"total_virtual_services", len(clusterState.VirtualServices),
		"matching_virtual_services", len(matchingVirtualServices),
		"total_service_entries", len(clusterState.ServiceEntries),
		"matching_service_entries", len(matchingServiceEntries),
		"total_destination_rules", len(clusterState.DestinationRules),
		"matching_destination_rules", len(matchingDestinationRules),
		"scope_to_namespace", scopeToNamespace)

	return &frontendv1alpha1.GetIstioResourcesResponse{
		VirtualServices:        matchingVirtualServices,
		DestinationRules:       matchingDestinationRules,
		Gateways:               matchingGateways,
		Sidecars:               matchingSidecars,
		EnvoyFilters:           matchingEnvoyFilters,
		RequestAuthentications: matchingRequestAuthentications,
		PeerAuthentications:    matchingPeerAuthentications,
		AuthorizationPolicies:  matchingAuthorizationPolicies,
		WasmPlugins:            matchingWasmPlugins,
		ServiceEntries:         matchingServiceEntries,
	}, nil
}

// mergeUniqueVirtualServices combines two slices of VirtualServices, removing duplicates based on name and namespace.
// This is used to merge VirtualServices found by different filtering approaches (workload-based and gateway-based).
func mergeUniqueVirtualServices(vs1, vs2 []*typesv1alpha1.VirtualService) []*typesv1alpha1.VirtualService {
	if len(vs2) == 0 {
		return vs1
	}
	if len(vs1) == 0 {
		return vs2
	}

	// Create a map to track existing VirtualServices by name+namespace
	existing := make(map[string]bool)
	for _, vs := range vs1 {
		key := vs.Namespace + "/" + vs.Name
		existing[key] = true
	}

	// Start with the first slice
	result := make([]*typesv1alpha1.VirtualService, len(vs1))
	copy(result, vs1)

	// Add unique items from the second slice
	for _, vs := range vs2 {
		key := vs.Namespace + "/" + vs.Name
		if !existing[key] {
			result = append(result, vs)
			existing[key] = true
		}
	}

	return result
}
