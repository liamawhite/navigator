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

package metrics

import (
	"context"
	"log/slog"
	"sync"
)

// ProviderFactory is a function that creates a new metrics provider instance
type ProviderFactory func(config Config, logger *slog.Logger, clusterName string) (Provider, error)

// Registry manages metrics provider factories and instances
type Registry struct {
	factories map[ProviderType]ProviderFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[ProviderType]ProviderFactory),
	}
}

// Register registers a provider factory for a specific provider type
func (r *Registry) Register(providerType ProviderType, factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[providerType] = factory
}

// Create creates a new provider instance based on the configuration
func (r *Registry) Create(config Config, logger *slog.Logger) (Provider, error) {
	return r.CreateWithClusterName(config, logger, "")
}

// CreateWithClusterName creates a new provider instance with cluster name for filtering
func (r *Registry) CreateWithClusterName(config Config, logger *slog.Logger, clusterName string) (Provider, error) {
	if !config.Enabled || config.Type == ProviderTypeNone {
		return NewNullProvider(config), nil
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	factory, exists := r.factories[config.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, ErrProviderNotSupported
	}

	return factory(config, logger, clusterName)
}

// GetSupportedTypes returns a list of supported provider types
func (r *Registry) GetSupportedTypes() []ProviderType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]ProviderType, 0, len(r.factories))
	for providerType := range r.factories {
		types = append(types, providerType)
	}
	return types
}

// nullProvider is a no-op provider for when metrics are disabled
type nullProvider struct {
	info ProviderInfo
}

// NewNullProvider creates a new null provider (no-op implementation)
func NewNullProvider(config Config) Provider {
	return &nullProvider{
		info: ProviderInfo{
			Type:     ProviderTypeNone,
			Endpoint: "",
			Health: ProviderHealth{
				Status:  HealthStatusHealthy,
				Message: "Metrics collection is disabled",
			},
		},
	}
}

func (n *nullProvider) GetProviderInfo() ProviderInfo {
	return n.info
}

func (n *nullProvider) GetServiceGraphMetrics(ctx context.Context, query MeshMetricsQuery) (*ServiceGraphMetrics, error) {
	return nil, ErrNoData
}

func (n *nullProvider) Close() error {
	return nil
}
