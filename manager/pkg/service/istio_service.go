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
	"github.com/liamawhite/navigator/pkg/istio/envoyfilter"
	"github.com/liamawhite/navigator/pkg/istio/gateway"
	"github.com/liamawhite/navigator/pkg/istio/peerauthentication"
	"github.com/liamawhite/navigator/pkg/istio/sidecar"
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
	var matchingPeerAuthentications []*typesv1alpha1.PeerAuthentication

	wg.Add(4)

	// Filter gateways concurrently
	go func() {
		defer wg.Done()
		matchingGateways = gateway.FilterGatewaysForWorkload(clusterState.Gateways, instance, namespace, scopeToNamespace)
	}()

	// Filter sidecars concurrently
	go func() {
		defer wg.Done()
		matchingSidecars = sidecar.FilterSidecarsForWorkload(clusterState.Sidecars, instance, namespace)
	}()

	// Filter envoy filters concurrently
	go func() {
		defer wg.Done()
		matchingEnvoyFilters = envoyfilter.FilterEnvoyFiltersForWorkload(clusterState.EnvoyFilters, instance, namespace, rootNamespace)
	}()

	// Filter peer authentications concurrently
	go func() {
		defer wg.Done()
		matchingPeerAuthentications = peerauthentication.FilterPeerAuthenticationsForWorkload(clusterState.PeerAuthentications, instance, namespace, rootNamespace)
	}()

	// Wait for all filtering operations to complete
	wg.Wait()

	i.logger.Debug("filtered istio resources",
		"cluster_id", clusterID,
		"total_gateways", len(clusterState.Gateways),
		"matching_gateways", len(matchingGateways),
		"total_sidecars", len(clusterState.Sidecars),
		"matching_sidecars", len(matchingSidecars),
		"total_envoyfilters", len(clusterState.EnvoyFilters),
		"matching_envoyfilters", len(matchingEnvoyFilters),
		"total_peer_authentications", len(clusterState.PeerAuthentications),
		"matching_peer_authentications", len(matchingPeerAuthentications),
		"scope_to_namespace", scopeToNamespace)

	// For now, return all other resources - in a more sophisticated implementation,
	// we would filter these based on relevance to the workload
	return &frontendv1alpha1.GetIstioResourcesResponse{
		VirtualServices:     clusterState.VirtualServices,
		DestinationRules:    clusterState.DestinationRules,
		Gateways:            matchingGateways,
		Sidecars:            matchingSidecars,
		EnvoyFilters:        matchingEnvoyFilters,
		PeerAuthentications: matchingPeerAuthentications,
	}, nil
}
