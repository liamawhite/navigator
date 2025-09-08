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
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client is a Prometheus HTTP API client
type Client struct {
	api    v1.API
	logger *slog.Logger
}

// ClientOption is a functional option for configuring the Prometheus client
type ClientOption func(*clientConfig)

// clientConfig holds the configuration for the Prometheus client
type clientConfig struct {
	bearerToken string
	timeout     time.Duration
}

// WithBearerToken configures bearer token authentication
func WithBearerToken(token string) ClientOption {
	return func(c *clientConfig) {
		c.bearerToken = token
	}
}

// WithTimeout configures the timeout for Prometheus requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// BearerTokenRoundTripper adds bearer token authentication to HTTP requests
type BearerTokenRoundTripper struct {
	Token string
	Next  http.RoundTripper
}

// RoundTrip implements http.RoundTripper
func (rt *BearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.Token != "" {
		req.Header.Set("Authorization", "Bearer "+rt.Token)
	}

	next := rt.Next
	if next == nil {
		next = http.DefaultTransport
	}

	return next.RoundTrip(req)
}

// NewClient creates a new Prometheus client with optional configuration
func NewClient(endpoint string, logger *slog.Logger, opts ...ClientOption) (*Client, error) {
	// Apply functional options with defaults
	cfg := &clientConfig{
		timeout: 5 * time.Second, // Default timeout
	}
	for _, opt := range opts {
		opt(cfg)
	}

	config := api.Config{
		Address: endpoint,
	}

	// Configure bearer token authentication if provided
	if cfg.bearerToken != "" {
		config.RoundTripper = &BearerTokenRoundTripper{
			Token: cfg.bearerToken,
		}
		logger.Debug("configured bearer token authentication for Prometheus client")
	}

	// Create Prometheus API client
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	promAPI := v1.NewAPI(client)

	return &Client{
		api:    promAPI,
		logger: logger,
	}, nil
}

// query executes a Prometheus query and returns native Prometheus types
func (c *Client) query(ctx context.Context, query string) (model.Value, error) {
	result, warnings, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}

	if len(warnings) > 0 {
		c.logger.Warn("Prometheus query returned warnings", "warnings", warnings)
	}

	return result, nil
}
