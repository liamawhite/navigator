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

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client is a Prometheus HTTP API client
type Client struct {
	api    v1.API
	logger *slog.Logger
}

// NewClient creates a new Prometheus client
func NewClient(endpoint string, timeout time.Duration, logger *slog.Logger) (*Client, error) {
	config := api.Config{
		Address: endpoint,
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
