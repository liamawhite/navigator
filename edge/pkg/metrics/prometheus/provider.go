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
	"fmt"
	"log/slog"
	"time"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
)

// Provider implements the metrics.Provider interface for Prometheus
type Provider struct {
	client *Client
	config metrics.Config
	info   metrics.ProviderInfo
	logger *slog.Logger
}

// NewProvider creates a new Prometheus metrics provider
func NewProvider(config metrics.Config, logger *slog.Logger) (*Provider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := NewClient(config.Endpoint, time.Duration(config.Timeout)*time.Second, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	provider := &Provider{
		client: client,
		config: config,
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

	return provider, nil
}

// GetProviderInfo returns information about this Prometheus provider
func (p *Provider) GetProviderInfo() metrics.ProviderInfo {
	return p.info
}

// Close closes the provider and cleans up resources
func (p *Provider) Close() error {
	// Nothing to cleanup for Prometheus provider
	return nil
}
