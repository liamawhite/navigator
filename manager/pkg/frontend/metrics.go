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

package frontend

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/liamawhite/navigator/manager/pkg/providers"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// MetricsService implements the frontend MetricsService
type MetricsService struct {
	frontendv1alpha1.UnimplementedMetricsServiceServer
	connectionManager   providers.ReadOptimizedConnectionManager
	meshMetricsProvider providers.MeshMetricsProvider
	logger              *slog.Logger
}

// NewMetricsService creates a new metrics service
func NewMetricsService(connectionManager providers.ReadOptimizedConnectionManager, meshMetricsProvider providers.MeshMetricsProvider, logger *slog.Logger) *MetricsService {
	return &MetricsService{
		connectionManager:   connectionManager,
		meshMetricsProvider: meshMetricsProvider,
		logger:              logger,
	}
}

// GetServiceConnections returns inbound and outbound connections for a specific service
func (m *MetricsService) GetServiceConnections(ctx context.Context, req *frontendv1alpha1.GetServiceConnectionsRequest) (*frontendv1alpha1.GetServiceConnectionsResponse, error) {
	m.logger.Debug("getting service connections", "service_name", req.ServiceName, "namespace", req.Namespace)

	// Get all connected clusters
	connectionInfos := m.connectionManager.GetConnectionInfo()
	var clustersQueried []string
	var allPairs []*typesv1alpha1.ServicePairMetrics

	// Collect all connected cluster IDs
	var healthyClusters []string
	for clusterID := range connectionInfos {
		healthyClusters = append(healthyClusters, clusterID)
	}

	// Query clusters in parallel using the targeted service connections method
	type clusterResult struct {
		clusterID string
		pairs     []*typesv1alpha1.ServicePairMetrics
		err       error
	}

	results := make(chan clusterResult, len(healthyClusters))
	var wg sync.WaitGroup

	for _, clusterID := range healthyClusters {
		wg.Add(1)
		go func(cID string) {
			defer wg.Done()

			// Request targeted service connections from this cluster
			serviceConnectionsMetrics, err := m.meshMetricsProvider.GetServiceConnections(ctx, cID, req)
			if err != nil {
				m.logger.Error("failed to get service connections from cluster", "cluster_id", cID, "error", err)
				results <- clusterResult{clusterID: cID, err: err}
				return
			}

			if serviceConnectionsMetrics != nil && len(serviceConnectionsMetrics.Pairs) > 0 {
				// Convert from metrics.ServiceGraphMetrics to []*typesv1alpha1.ServicePairMetrics
				var pairs []*typesv1alpha1.ServicePairMetrics
				for _, pair := range serviceConnectionsMetrics.Pairs {
					pairs = append(pairs, &typesv1alpha1.ServicePairMetrics{
						SourceCluster:        pair.SourceCluster,
						SourceNamespace:      pair.SourceNamespace,
						SourceService:        pair.SourceService,
						DestinationCluster:   pair.DestinationCluster,
						DestinationNamespace: pair.DestinationNamespace,
						DestinationService:   pair.DestinationService,
						ErrorRate:            pair.ErrorRate,
						RequestRate:          pair.RequestRate,
					})
				}
				results <- clusterResult{clusterID: cID, pairs: pairs}
			} else {
				results <- clusterResult{clusterID: cID, pairs: nil}
			}
		}(clusterID)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Collect results
	for result := range results {
		if result.err != nil {
			continue // Error already logged above
		}
		if len(result.pairs) > 0 {
			allPairs = append(allPairs, result.pairs...)
			clustersQueried = append(clustersQueried, result.clusterID)
		}
	}

	// Separate inbound and outbound connections
	var inbound []*typesv1alpha1.ServicePairMetrics
	var outbound []*typesv1alpha1.ServicePairMetrics

	for _, pair := range allPairs {
		// Inbound: services calling this service
		if pair.DestinationService == req.ServiceName && pair.DestinationNamespace == req.Namespace {
			inbound = append(inbound, pair)
		}
		// Outbound: services this service calls
		if pair.SourceService == req.ServiceName && pair.SourceNamespace == req.Namespace {
			outbound = append(outbound, pair)
		}
	}

	m.logger.Debug("retrieved targeted service connections",
		"service_name", req.ServiceName,
		"namespace", req.Namespace,
		"clusters_queried", len(clustersQueried),
		"total_pairs", len(allPairs),
		"inbound_count", len(inbound),
		"outbound_count", len(outbound))

	return &frontendv1alpha1.GetServiceConnectionsResponse{
		Inbound:         inbound,
		Outbound:        outbound,
		Timestamp:       time.Now().Format("2006-01-02T15:04:05Z07:00"),
		ClustersQueried: clustersQueried,
	}, nil
}
