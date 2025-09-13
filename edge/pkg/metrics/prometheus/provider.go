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
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Provider implements the metrics.Provider interface for Prometheus
type Provider struct {
	client      *Client
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
			Health: metrics.ProviderHealth{
				Status:  metrics.HealthStatusUnknown,
				Message: "Not yet checked",
			},
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
func (p *Provider) GetServiceConnections(ctx context.Context, serviceName, namespace string, startTime, endTime *timestamppb.Timestamp) (*typesv1alpha1.ServiceGraphMetrics, error) {
	p.logger.Info("retrieving service connections from Prometheus",
		"service_name", serviceName,
		"namespace", namespace,
		"cluster", p.clusterName)

	// Health check will be performed by the actual query - no need to precheck

	// Convert protobuf timestamps to time.Time
	var start, end time.Time
	if startTime != nil {
		start = startTime.AsTime()
	} else {
		start = time.Now().Add(-5 * time.Minute) // Default to 5 minutes ago
	}

	if endTime != nil {
		end = endTime.AsTime()
	} else {
		end = time.Now() // Default to now
	}

	// Call the client's GetServiceConnections method directly
	return p.client.GetServiceConnections(ctx, serviceName, namespace, start, end)
}

// Close closes the provider and cleans up resources
func (p *Provider) Close() error {
	// Nothing to cleanup for Prometheus provider
	return nil
}
