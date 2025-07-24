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

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/istio/gateway"
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
func (i *IstioService) GetIstioResourcesForWorkload(ctx context.Context, clusterID, namespace string, labels map[string]string) (*frontendv1alpha1.GetIstioResourcesResponse, error) {
	i.logger.Debug("getting istio resources for workload",
		"cluster_id", clusterID,
		"namespace", namespace,
		"labels", labels)

	// Get cluster state
	clusterState, err := i.clusterProvider.GetClusterState(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster state for cluster %s: %w", clusterID, err)
	}

	// Determine namespace scoping from control plane config
	var scopeToNamespace bool
	if clusterState.IstioControlPlaneConfig != nil {
		scopeToNamespace = clusterState.IstioControlPlaneConfig.PilotScopeGatewayToNamespace
	}

	// Filter gateways based on workload selector matching
	matchingGateways := gateway.FilterGatewaysForWorkload(clusterState.Gateways, labels, namespace, scopeToNamespace)

	// Filter sidecars based on workload selector matching
	matchingSidecars := sidecar.FilterSidecarsForWorkload(clusterState.Sidecars, labels, namespace)

	i.logger.Debug("filtered istio resources",
		"cluster_id", clusterID,
		"total_gateways", len(clusterState.Gateways),
		"matching_gateways", len(matchingGateways),
		"total_sidecars", len(clusterState.Sidecars),
		"matching_sidecars", len(matchingSidecars),
		"scope_to_namespace", scopeToNamespace)

	// For now, return all other resources - in a more sophisticated implementation,
	// we would filter these based on relevance to the workload
	return &frontendv1alpha1.GetIstioResourcesResponse{
		VirtualServices:  clusterState.VirtualServices,
		DestinationRules: clusterState.DestinationRules,
		Gateways:         matchingGateways,
		Sidecars:         matchingSidecars,
		EnvoyFilters:     clusterState.EnvoyFilters,
	}, nil
}
