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
)

// Provider represents a generic metrics provider interface
type Provider interface {
	// GetProviderInfo returns information about this metrics provider
	GetProviderInfo() ProviderInfo

	// GetServiceGraphMetrics retrieves service-to-service metrics across the service graph
	GetServiceGraphMetrics(ctx context.Context, query MeshMetricsQuery) (*ServiceGraphMetrics, error)

	// Close closes the provider and cleans up resources
	Close() error
}

// Config represents the configuration for a metrics provider
type Config struct {
	// Type is the type of metrics provider
	Type ProviderType `json:"type" yaml:"type"`
	// Endpoint is the endpoint URL for the metrics provider
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	// Enabled indicates whether metrics collection is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`
	// QueryInterval is how often to query for metrics (in seconds)
	QueryInterval int `json:"query_interval" yaml:"query_interval"`
	// Timeout is the timeout for metrics queries (in seconds)
	Timeout int `json:"timeout" yaml:"timeout"`
	// BearerToken for bearer token authentication
	BearerToken string `json:"bearer_token,omitempty" yaml:"bearer_token,omitempty"`
}

// Validate validates the metrics configuration
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Type == "" {
		c.Type = ProviderTypeNone
	}

	if c.Type != ProviderTypeNone && c.Endpoint == "" {
		return ErrMissingEndpoint
	}

	if c.QueryInterval <= 0 {
		c.QueryInterval = 30 // Default to 30 seconds
	}

	if c.Timeout <= 0 {
		c.Timeout = 10 // Default to 10 seconds
	}

	return nil
}
