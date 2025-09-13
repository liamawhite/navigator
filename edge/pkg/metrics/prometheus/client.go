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
	"sync"
	"time"

	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
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

// GetServiceConnections retrieves service connection metrics for a specific service
func (c *Client) GetServiceConnections(ctx context.Context, serviceName, namespace string, startTime, endTime time.Time) (*typesv1alpha1.ServiceGraphMetrics, error) {
	c.logger.Info("querying service connections from Prometheus",
		"service", serviceName,
		"namespace", namespace,
		"start", startTime,
		"end", endTime)

	// Build base queries for inbound and outbound connections
	// Inbound: traffic coming TO this service
	inboundRequestQuery := fmt.Sprintf(`
		sum(rate(istio_requests_total{destination_canonical_service="%s",destination_service_namespace="%s"}[5m])) by (
			source_cluster,
			source_workload_namespace,
			source_canonical_service,
			destination_cluster,
			destination_service_namespace,
			destination_canonical_service
		)`, serviceName, namespace)

	// Outbound: traffic going FROM this service
	outboundRequestQuery := fmt.Sprintf(`
		sum(rate(istio_requests_total{source_canonical_service="%s",source_workload_namespace="%s"}[5m])) by (
			source_cluster,
			source_workload_namespace,
			source_canonical_service,
			destination_cluster,
			destination_service_namespace,
			destination_canonical_service
		)`, serviceName, namespace)

	// Inbound error rate
	inboundErrorQuery := fmt.Sprintf(`
		sum(rate(istio_requests_total{destination_canonical_service="%s",destination_service_namespace="%s",response_code=~"5.*"}[5m])) by (
			source_cluster,
			source_workload_namespace,
			source_canonical_service,
			destination_cluster,
			destination_service_namespace,
			destination_canonical_service
		) / sum(rate(istio_requests_total{destination_canonical_service="%s",destination_service_namespace="%s"}[5m])) by (
			source_cluster,
			source_workload_namespace,
			source_canonical_service,
			destination_cluster,
			destination_service_namespace,
			destination_canonical_service
		)`, serviceName, namespace, serviceName, namespace)

	// Outbound error rate
	outboundErrorQuery := fmt.Sprintf(`
		sum(rate(istio_requests_total{source_canonical_service="%s",source_workload_namespace="%s",response_code=~"5.*"}[5m])) by (
			source_cluster,
			source_workload_namespace,
			source_canonical_service,
			destination_cluster,
			destination_service_namespace,
			destination_canonical_service
		) / sum(rate(istio_requests_total{source_canonical_service="%s",source_workload_namespace="%s"}[5m])) by (
			source_cluster,
			source_workload_namespace,
			source_canonical_service,
			destination_cluster,
			destination_service_namespace,
			destination_canonical_service
		)`, serviceName, namespace, serviceName, namespace)

	// Execute queries in parallel
	type queryResult struct {
		data model.Value
		err  error
		name string
	}

	queries := []struct {
		name  string
		query string
	}{
		{"inbound_requests", inboundRequestQuery},
		{"outbound_requests", outboundRequestQuery},
		{"inbound_errors", inboundErrorQuery},
		{"outbound_errors", outboundErrorQuery},
	}

	results := make(chan queryResult, len(queries))
	var wg sync.WaitGroup

	for _, q := range queries {
		wg.Add(1)
		go func(name, query string) {
			defer wg.Done()
			data, err := c.query(ctx, query)
			results <- queryResult{data: data, err: err, name: name}
		}(q.name, q.query)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	queryData := make(map[string]model.Value)
	for result := range results {
		if result.err != nil {
			c.logger.Error("failed to execute query", "query", result.name, "error", result.err)
			// Don't fail completely, just log the error
			continue
		}
		queryData[result.name] = result.data
	}

	// Process results into service pairs
	pairs := make(map[string]*typesv1alpha1.ServicePairMetrics)

	// Process inbound requests
	if data, ok := queryData["inbound_requests"]; ok {
		c.processRequestData(data, pairs, true)
	}

	// Process outbound requests
	if data, ok := queryData["outbound_requests"]; ok {
		c.processRequestData(data, pairs, false)
	}

	// Process inbound errors
	if data, ok := queryData["inbound_errors"]; ok {
		c.processErrorData(data, pairs, true)
	}

	// Process outbound errors
	if data, ok := queryData["outbound_errors"]; ok {
		c.processErrorData(data, pairs, false)
	}

	// Convert pairs map to slice
	var servicePairs []*typesv1alpha1.ServicePairMetrics
	for _, pair := range pairs {
		servicePairs = append(servicePairs, pair)
	}

	return &typesv1alpha1.ServiceGraphMetrics{
		Pairs:     servicePairs,
		ClusterId: "", // We don't have cluster context here
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// processRequestData processes Prometheus query results for request rate metrics
func (c *Client) processRequestData(data model.Value, pairs map[string]*typesv1alpha1.ServicePairMetrics, isInbound bool) {
	vector, ok := data.(model.Vector)
	if !ok {
		c.logger.Warn("expected vector result for request data")
		return
	}

	for _, sample := range vector {
		key := c.createPairKey(sample.Metric)
		if key == "" {
			continue
		}

		if pairs[key] == nil {
			pairs[key] = &typesv1alpha1.ServicePairMetrics{
				SourceCluster:        string(sample.Metric["source_cluster"]),
				SourceNamespace:      string(sample.Metric["source_workload_namespace"]),
				SourceService:        string(sample.Metric["source_canonical_service"]),
				DestinationCluster:   string(sample.Metric["destination_cluster"]),
				DestinationNamespace: string(sample.Metric["destination_service_namespace"]),
				DestinationService:   string(sample.Metric["destination_canonical_service"]),
			}
		}

		pairs[key].RequestRate = float64(sample.Value)
	}
}

// processErrorData processes Prometheus query results for error rate metrics
func (c *Client) processErrorData(data model.Value, pairs map[string]*typesv1alpha1.ServicePairMetrics, isInbound bool) {
	vector, ok := data.(model.Vector)
	if !ok {
		c.logger.Warn("expected vector result for error data")
		return
	}

	for _, sample := range vector {
		key := c.createPairKey(sample.Metric)
		if key == "" {
			continue
		}

		if pairs[key] == nil {
			pairs[key] = &typesv1alpha1.ServicePairMetrics{
				SourceCluster:        string(sample.Metric["source_cluster"]),
				SourceNamespace:      string(sample.Metric["source_workload_namespace"]),
				SourceService:        string(sample.Metric["source_canonical_service"]),
				DestinationCluster:   string(sample.Metric["destination_cluster"]),
				DestinationNamespace: string(sample.Metric["destination_service_namespace"]),
				DestinationService:   string(sample.Metric["destination_canonical_service"]),
			}
		}

		pairs[key].ErrorRate = float64(sample.Value)
	}
}

// createPairKey creates a unique key for a service pair
func (c *Client) createPairKey(metric model.Metric) string {
	sourceCluster := string(metric["source_cluster"])
	sourceNamespace := string(metric["source_workload_namespace"])
	sourceService := string(metric["source_canonical_service"])
	destCluster := string(metric["destination_cluster"])
	destNamespace := string(metric["destination_service_namespace"])
	destService := string(metric["destination_canonical_service"])

	if sourceService == "" || destService == "" {
		return ""
	}

	return fmt.Sprintf("%s:%s:%s->%s:%s:%s",
		sourceCluster, sourceNamespace, sourceService,
		destCluster, destNamespace, destService)
}
