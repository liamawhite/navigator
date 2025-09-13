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

// GetServiceGraphMetrics returns service-to-service graph metrics across the mesh
func (m *MetricsService) GetServiceGraphMetrics(ctx context.Context, req *frontendv1alpha1.GetServiceGraphMetricsRequest) (*frontendv1alpha1.GetServiceGraphMetricsResponse, error) {
	m.logger.Debug("getting service mesh metrics", "filters", req)

	// Get all connected clusters
	connectionInfos := m.connectionManager.GetConnectionInfo()
	var clustersQueried []string
	var allPairs []*typesv1alpha1.ServicePairMetrics

	// Collect all connected cluster IDs
	var healthyClusters []string
	for clusterID := range connectionInfos {
		healthyClusters = append(healthyClusters, clusterID)
	}

	// Query clusters in parallel
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

			// Request mesh metrics from this cluster using the dedicated service
			meshMetrics, err := m.meshMetricsProvider.GetServiceGraphMetrics(ctx, cID, req)
			if err != nil {
				m.logger.Error("failed to get mesh metrics from cluster", "cluster_id", cID, "error", err)
				results <- clusterResult{clusterID: cID, err: err}
				return
			}

			if meshMetrics != nil && len(meshMetrics.Pairs) > 0 {
				// Use the shared types directly - no conversion needed
				results <- clusterResult{clusterID: cID, pairs: meshMetrics.Pairs}
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

	m.logger.Debug("retrieved mesh metrics", "clusters_queried", len(clustersQueried), "total_pairs", len(allPairs))

	return &frontendv1alpha1.GetServiceGraphMetricsResponse{
		Pairs:           allPairs,
		Timestamp:       time.Now().Format("2006-01-02T15:04:05Z07:00"),
		ClustersQueried: clustersQueried,
	}, nil
}

// GetServiceConnections returns inbound and outbound connections for a specific service
func (m *MetricsService) GetServiceConnections(ctx context.Context, req *frontendv1alpha1.GetServiceConnectionsRequest) (*frontendv1alpha1.GetServiceConnectionsResponse, error) {
	m.logger.Debug("getting service connections", "service_name", req.ServiceName, "namespace", req.Namespace)

	// Convert to graph metrics request to reuse existing infrastructure
	graphReq := &frontendv1alpha1.GetServiceGraphMetricsRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		// No namespace or cluster filters - we want all connections across the mesh
	}

	// Get all service mesh metrics
	graphResp, err := m.GetServiceGraphMetrics(ctx, graphReq)
	if err != nil {
		return nil, err
	}

	// Filter and categorize connections for this specific service
	var inbound []*typesv1alpha1.ServicePairMetrics
	var outbound []*typesv1alpha1.ServicePairMetrics

	for _, pair := range graphResp.Pairs {
		// Inbound: services calling this service
		if pair.DestinationService == req.ServiceName && pair.DestinationNamespace == req.Namespace {
			inbound = append(inbound, pair)
		}
		// Outbound: services this service calls
		if pair.SourceService == req.ServiceName && pair.SourceNamespace == req.Namespace {
			outbound = append(outbound, pair)
		}
	}

	m.logger.Debug("filtered service connections",
		"service_name", req.ServiceName,
		"namespace", req.Namespace,
		"inbound_count", len(inbound),
		"outbound_count", len(outbound))

	return &frontendv1alpha1.GetServiceConnectionsResponse{
		Inbound:         inbound,
		Outbound:        outbound,
		Timestamp:       graphResp.Timestamp,
		ClustersQueried: graphResp.ClustersQueried,
	}, nil
}
