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

package prometheus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/prometheus/common/model"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ClientInterface defines the interface for Prometheus client operations
type ClientInterface interface {
	query(ctx context.Context, query string) (model.Value, error)
	GetServiceConnections(ctx context.Context, serviceName, namespace string, startTime, endTime time.Time) (*typesv1alpha1.ServiceGraphMetrics, error)
}

// Provider implements the metrics.Provider interface for Prometheus
type Provider struct {
	client      ClientInterface
	config      metrics.Config
	info        metrics.ProviderInfo
	logger      *slog.Logger
	clusterName string
}

// NewProvider creates a new Prometheus metrics provider with cluster name for filtering
func NewProvider(config metrics.Config, logger *slog.Logger, clusterName string) (*Provider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Build client options
	var clientOpts []ClientOption
	if config.BearerToken != "" {
		clientOpts = append(clientOpts, WithBearerToken(config.BearerToken))
	}
	if config.Timeout > 0 {
		clientOpts = append(clientOpts, WithTimeout(time.Duration(config.Timeout)*time.Second))
	}

	client, err := NewClient(config.Endpoint, logger, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	provider := &Provider{
		client:      client,
		config:      config,
		clusterName: clusterName,
		info: metrics.ProviderInfo{
			Type:     metrics.ProviderTypePrometheus,
			Endpoint: config.Endpoint,
		},
		logger: logger,
	}

	if clusterName != "" {
		logger.Debug("created Prometheus provider with cluster filtering", "cluster_name", clusterName)
	}

	return provider, nil
}

// GetProviderInfo returns information about this Prometheus provider
func (p *Provider) GetProviderInfo() metrics.ProviderInfo {
	return p.info
}

// GetClusterName returns the current cluster name
func (p *Provider) GetClusterName() string {
	return p.clusterName
}

// GetServiceConnections (new interface) retrieves service connection metrics for a specific service - implements interfaces.MetricsProvider
func (p *Provider) GetServiceConnections(ctx context.Context, serviceName, namespace string, proxyMode typesv1alpha1.ProxyMode, startTime, endTime *timestamppb.Timestamp) (*typesv1alpha1.ServiceGraphMetrics, error) {
	p.logger.Info("retrieving service connections from Prometheus",
		"service_name", serviceName,
		"namespace", namespace,
		"proxy_mode", proxyMode.String(),
		"cluster", p.clusterName)

	// Health check will be performed by the actual query - no need to precheck

	// Use the fixed getServiceConnectionsInternal method instead of the buggy client method
	// Note: startTime and endTime are currently ignored since getServiceConnectionsInternal uses a fixed 5m window
	result, err := p.getServiceConnectionsInternal(ctx, serviceName, namespace, proxyMode, metrics.MeshMetricsFilters{})
	if err != nil {
		return nil, err
	}

	// Convert from internal metrics format to API format
	var apiPairs []*typesv1alpha1.ServicePairMetrics
	for _, pair := range result.Pairs {
		apiPairs = append(apiPairs, &typesv1alpha1.ServicePairMetrics{
			SourceCluster:        pair.SourceCluster,
			SourceNamespace:      pair.SourceNamespace,
			SourceService:        pair.SourceService,
			DestinationCluster:   pair.DestinationCluster,
			DestinationNamespace: pair.DestinationNamespace,
			DestinationService:   pair.DestinationService,
			RequestRate:          pair.RequestRate,
			ErrorRate:            pair.ErrorRate,
			LatencyP99:           durationpb.New(time.Duration(pair.LatencyP99 * float64(time.Millisecond))),
			LatencyDistribution:  pair.LatencyDistribution,
		})
	}

	return &typesv1alpha1.ServiceGraphMetrics{
		Pairs:     apiPairs,
		ClusterId: p.clusterName,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// Close closes the provider and cleans up resources
func (p *Provider) Close() error {
	// Nothing to cleanup for Prometheus provider
	return nil
}
