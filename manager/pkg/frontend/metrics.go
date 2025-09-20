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
	"fmt"
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

	// Validate service name and namespace are provided
	if req.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if req.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	// Validate that the service exists before querying metrics
	// Use the same service ID format as the rest of the system: namespace:serviceName
	serviceID := fmt.Sprintf("%s:%s", req.Namespace, req.ServiceName)
	aggregatedService, serviceExists := m.connectionManager.GetAggregatedService(serviceID)
	if !serviceExists {
		m.logger.Debug("service not found", "service_name", req.ServiceName, "namespace", req.Namespace, "service_id", serviceID)
		return &frontendv1alpha1.GetServiceConnectionsResponse{
			Inbound:         []*typesv1alpha1.ServicePairMetrics{},
			Outbound:        []*typesv1alpha1.ServicePairMetrics{},
			Timestamp:       time.Now().Format("2006-01-02T15:04:05Z07:00"),
			ClustersQueried: []string{},
		}, nil
	}

	// Determine the ProxyMode from service instances
	// All instances of a service should have the same ProxyMode
	proxyMode := typesv1alpha1.ProxyMode_SIDECAR // Default to SIDECAR
	if len(aggregatedService.Instances) > 0 {
		proxyMode = aggregatedService.Instances[0].ProxyMode
	}

	var clustersQueried []string
	var allPairs []*typesv1alpha1.ServicePairMetrics

	// Get all connected clusters for metrics querying
	connectionInfos := m.connectionManager.GetConnectionInfo()

	// Collect all connected cluster IDs
	var healthyClusters []string
	for clusterID := range connectionInfos {
		healthyClusters = append(healthyClusters, clusterID)
	}

	// Query clusters in parallel using the targeted service connections method with proper cancellation
	type clusterResult struct {
		clusterID string
		pairs     []*typesv1alpha1.ServicePairMetrics
		err       error
	}

	results := make(chan clusterResult, len(healthyClusters))
	var wg sync.WaitGroup

	// Create cancellable context for cluster queries
	clusterCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, clusterID := range healthyClusters {
		wg.Add(1)
		go func(cID string) {
			defer wg.Done()

			// Check for cancellation before starting work
			select {
			case <-clusterCtx.Done():
				results <- clusterResult{clusterID: cID, err: clusterCtx.Err()}
				return
			default:
			}

			// Request targeted service connections from this cluster
			serviceConnectionsMetrics, err := m.meshMetricsProvider.GetServiceConnections(clusterCtx, cID, req, proxyMode)
			if err != nil {
				m.logger.Error("failed to get service connections from cluster", "cluster_id", cID, "error", err)
				results <- clusterResult{clusterID: cID, err: err}
				return
			}

			if serviceConnectionsMetrics != nil && len(serviceConnectionsMetrics.Pairs) > 0 {
				// Convert from metrics.ServiceGraphMetrics to []*typesv1alpha1.ServicePairMetrics
				// P99 is now calculated by the edge service, so we just pass it through
				var pairs []*typesv1alpha1.ServicePairMetrics
				for _, pair := range serviceConnectionsMetrics.Pairs {
					var distBuckets int
					if pair.LatencyDistribution != nil {
						distBuckets = len(pair.LatencyDistribution.Buckets)
					}
					m.logger.Debug("manager received pair from edge", "cluster", cID, "source", pair.SourceService, "dest", pair.DestinationService, "has_distribution", pair.LatencyDistribution != nil, "dist_buckets", distBuckets, "p99", pair.LatencyP99)
					pairs = append(pairs, &typesv1alpha1.ServicePairMetrics{
						SourceCluster:        pair.SourceCluster,
						SourceNamespace:      pair.SourceNamespace,
						SourceService:        pair.SourceService,
						DestinationCluster:   pair.DestinationCluster,
						DestinationNamespace: pair.DestinationNamespace,
						DestinationService:   pair.DestinationService,
						ErrorRate:            pair.ErrorRate,
						RequestRate:          pair.RequestRate,
						LatencyP99:           pair.LatencyP99, // Calculated by edge
						LatencyDistribution:  pair.LatencyDistribution,
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
