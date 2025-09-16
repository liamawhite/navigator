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
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
)

// Targeted query templates for specific service connections
var (
	inboundRequestRateQueryTemplate = template.Must(template.New("inboundRequestRate").Parse(`
sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="destination", destination_canonical_service="{{.ServiceName}}", destination_service_namespace="{{.ServiceNamespace}}"{{.FilterClause}}}[{{.TimeRange}}])
)`))

	outboundRequestRateQueryTemplate = template.Must(template.New("outboundRequestRate").Parse(`
sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="source", source_canonical_service="{{.ServiceName}}", source_workload_namespace="{{.ServiceNamespace}}"{{.FilterClause}}}[{{.TimeRange}}])
)`))

	inboundErrorRateQueryTemplate = template.Must(template.New("inboundErrorRate").Parse(`
sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="destination", destination_canonical_service="{{.ServiceName}}", destination_service_namespace="{{.ServiceNamespace}}", response_code=~"0|4..|5.."{{.FilterClause}}}[{{.TimeRange}}])
)`))

	outboundErrorRateQueryTemplate = template.Must(template.New("outboundErrorRate").Parse(`
sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="source", source_canonical_service="{{.ServiceName}}", source_workload_namespace="{{.ServiceNamespace}}", response_code=~"0|4..|5.."{{.FilterClause}}}[{{.TimeRange}}])
)`))

	inboundLatencyP99QueryTemplate = template.Must(template.New("inboundLatencyP99").Parse(`
histogram_quantile(0.99,
  sum by (
    source_cluster, source_workload_namespace, source_canonical_service,
    destination_cluster, destination_service_namespace, destination_canonical_service, le
  )(
    rate(istio_request_duration_milliseconds_bucket{reporter="destination", destination_canonical_service="{{.ServiceName}}", destination_service_namespace="{{.ServiceNamespace}}"{{.FilterClause}}}[{{.TimeRange}}])
  )
)`))

	outboundLatencyP99QueryTemplate = template.Must(template.New("outboundLatencyP99").Parse(`
histogram_quantile(0.99,
  sum by (
    source_cluster, source_workload_namespace, source_canonical_service,
    destination_cluster, destination_service_namespace, destination_canonical_service, le
  )(
    rate(istio_request_duration_milliseconds_bucket{reporter="source", source_canonical_service="{{.ServiceName}}", source_workload_namespace="{{.ServiceNamespace}}"{{.FilterClause}}}[{{.TimeRange}}])
  )
)`))
)

// serviceConnectionsQueryTemplateData holds the data for service-specific query templates
type serviceConnectionsQueryTemplateData struct {
	FilterClause     string
	TimeRange        string
	ServiceName      string
	ServiceNamespace string
}

// getServiceConnectionsInternal returns targeted metrics for a specific service's connections
func (p *Provider) getServiceConnectionsInternal(ctx context.Context, serviceName, serviceNamespace string, filters metrics.MeshMetricsFilters) (*metrics.ServiceGraphMetrics, error) {
	// Check if client is available
	if p.client == nil {
		return nil, fmt.Errorf("prometheus client not available")
	}

	// Default to 5-minute time range if not specified
	timeRange := "5m"
	timestamp := time.Now()

	// Execute targeted queries in parallel with proper cancellation support
	type connectionQueryResult struct {
		ProcessedMetrics processedMetrics
		QueryType        string
		Error            error
	}

	results := make(chan connectionQueryResult, 6)
	var wg sync.WaitGroup

	// Create cancellable context for goroutines
	queryCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Inbound request rate query
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check for cancellation before starting work
		select {
		case <-queryCtx.Done():
			results <- connectionQueryResult{Error: queryCtx.Err(), QueryType: "inbound_request_rate"}
			return
		default:
		}

		query, err := p.buildServiceConnectionQuery(inboundRequestRateQueryTemplate, serviceName, serviceNamespace, filters, timeRange)
		if err != nil {
			results <- connectionQueryResult{Error: fmt.Errorf("failed to build inbound request rate query: %w", err), QueryType: "inbound_request_rate"}
			return
		}

		p.logger.Debug("executing inbound request rate query", "query", query, "service", serviceName, "namespace", serviceNamespace)
		resp, err := p.client.query(queryCtx, query)
		if err != nil {
			results <- connectionQueryResult{Error: err, QueryType: "inbound_request_rate"}
			return
		}

		processedMetrics := p.processRequestRateResponse(resp, timestamp)
		results <- connectionQueryResult{ProcessedMetrics: processedMetrics, QueryType: "inbound_request_rate"}
	}()

	// Outbound request rate query
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check for cancellation before starting work
		select {
		case <-queryCtx.Done():
			results <- connectionQueryResult{Error: queryCtx.Err(), QueryType: "outbound_request_rate"}
			return
		default:
		}

		query, err := p.buildServiceConnectionQuery(outboundRequestRateQueryTemplate, serviceName, serviceNamespace, filters, timeRange)
		if err != nil {
			results <- connectionQueryResult{Error: fmt.Errorf("failed to build outbound request rate query: %w", err), QueryType: "outbound_request_rate"}
			return
		}

		p.logger.Debug("executing outbound request rate query", "query", query, "service", serviceName, "namespace", serviceNamespace)
		resp, err := p.client.query(queryCtx, query)
		if err != nil {
			results <- connectionQueryResult{Error: err, QueryType: "outbound_request_rate"}
			return
		}

		processedMetrics := p.processRequestRateResponse(resp, timestamp)
		results <- connectionQueryResult{ProcessedMetrics: processedMetrics, QueryType: "outbound_request_rate"}
	}()

	// Inbound error rate query
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check for cancellation before starting work
		select {
		case <-queryCtx.Done():
			results <- connectionQueryResult{Error: queryCtx.Err(), QueryType: "inbound_error_rate"}
			return
		default:
		}

		query, err := p.buildServiceConnectionQuery(inboundErrorRateQueryTemplate, serviceName, serviceNamespace, filters, timeRange)
		if err != nil {
			results <- connectionQueryResult{Error: fmt.Errorf("failed to build inbound error rate query: %w", err), QueryType: "inbound_error_rate"}
			return
		}

		p.logger.Debug("executing inbound error rate query", "query", query, "service", serviceName, "namespace", serviceNamespace)
		resp, err := p.client.query(queryCtx, query)
		if err != nil {
			results <- connectionQueryResult{Error: err, QueryType: "inbound_error_rate"}
			return
		}

		processedMetrics := p.processErrorRateResponse(resp, timestamp)
		results <- connectionQueryResult{ProcessedMetrics: processedMetrics, QueryType: "inbound_error_rate"}
	}()

	// Outbound error rate query
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check for cancellation before starting work
		select {
		case <-queryCtx.Done():
			results <- connectionQueryResult{Error: queryCtx.Err(), QueryType: "outbound_error_rate"}
			return
		default:
		}

		query, err := p.buildServiceConnectionQuery(outboundErrorRateQueryTemplate, serviceName, serviceNamespace, filters, timeRange)
		if err != nil {
			results <- connectionQueryResult{Error: fmt.Errorf("failed to build outbound error rate query: %w", err), QueryType: "outbound_error_rate"}
			return
		}

		p.logger.Debug("executing outbound error rate query", "query", query, "service", serviceName, "namespace", serviceNamespace)
		resp, err := p.client.query(queryCtx, query)
		if err != nil {
			results <- connectionQueryResult{Error: err, QueryType: "outbound_error_rate"}
			return
		}

		processedMetrics := p.processErrorRateResponse(resp, timestamp)
		results <- connectionQueryResult{ProcessedMetrics: processedMetrics, QueryType: "outbound_error_rate"}
	}()

	// Inbound latency P99 query
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check for cancellation before starting work
		select {
		case <-queryCtx.Done():
			results <- connectionQueryResult{Error: queryCtx.Err(), QueryType: "inbound_latency_p99"}
			return
		default:
		}

		query, err := p.buildServiceConnectionQuery(inboundLatencyP99QueryTemplate, serviceName, serviceNamespace, filters, timeRange)
		if err != nil {
			results <- connectionQueryResult{Error: fmt.Errorf("failed to build inbound latency P99 query: %w", err), QueryType: "inbound_latency_p99"}
			return
		}

		p.logger.Debug("executing inbound latency P99 query", "query", query, "service", serviceName, "namespace", serviceNamespace)
		resp, err := p.client.query(queryCtx, query)
		if err != nil {
			results <- connectionQueryResult{Error: err, QueryType: "inbound_latency_p99"}
			return
		}

		processedMetrics := p.processLatencyResponse(resp, timestamp)
		results <- connectionQueryResult{ProcessedMetrics: processedMetrics, QueryType: "inbound_latency_p99"}
	}()

	// Outbound latency P99 query
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Check for cancellation before starting work
		select {
		case <-queryCtx.Done():
			results <- connectionQueryResult{Error: queryCtx.Err(), QueryType: "outbound_latency_p99"}
			return
		default:
		}

		query, err := p.buildServiceConnectionQuery(outboundLatencyP99QueryTemplate, serviceName, serviceNamespace, filters, timeRange)
		if err != nil {
			results <- connectionQueryResult{Error: fmt.Errorf("failed to build outbound latency P99 query: %w", err), QueryType: "outbound_latency_p99"}
			return
		}

		p.logger.Debug("executing outbound latency P99 query", "query", query, "service", serviceName, "namespace", serviceNamespace)
		resp, err := p.client.query(queryCtx, query)
		if err != nil {
			results <- connectionQueryResult{Error: err, QueryType: "outbound_latency_p99"}
			return
		}

		processedMetrics := p.processLatencyResponse(resp, timestamp)
		results <- connectionQueryResult{ProcessedMetrics: processedMetrics, QueryType: "outbound_latency_p99"}
	}()

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Collect and merge results
	allRequestPairs := make(map[string]*metrics.ServicePairMetrics)
	allErrorPairs := make(map[string]*metrics.ServicePairMetrics)
	allLatencyPairs := make(map[string]*metrics.ServicePairMetrics)

	for result := range results {
		if result.Error != nil {
			p.logger.Error("query failed", "query_type", result.QueryType, "error", result.Error, "service", serviceName, "namespace", serviceNamespace)
			continue
		}

		if result.ProcessedMetrics.Error != nil {
			p.logger.Error("query processing failed", "query_type", result.QueryType, "error", result.ProcessedMetrics.Error, "service", serviceName, "namespace", serviceNamespace)
			continue
		}

		// Merge the processed metrics based on type
		switch result.ProcessedMetrics.MetricType {
		case "request_rate":
			for key, pair := range result.ProcessedMetrics.PairData {
				allRequestPairs[key] = pair
			}
		case "error_rate":
			for key, pair := range result.ProcessedMetrics.PairData {
				allErrorPairs[key] = pair
			}
		case "latency_p99":
			for key, pair := range result.ProcessedMetrics.PairData {
				allLatencyPairs[key] = pair
			}
		}
	}

	// Merge request, error, and latency data
	mergedPairs := p.mergePairMaps(allRequestPairs, allErrorPairs, allLatencyPairs)

	// Convert to slice
	var pairs []metrics.ServicePairMetrics
	for _, pair := range mergedPairs {
		pairs = append(pairs, *pair)
	}

	p.logger.Debug("completed service connections query",
		"service", serviceName,
		"namespace", serviceNamespace,
		"total_pairs", len(pairs),
		"request_pairs", len(allRequestPairs),
		"error_pairs", len(allErrorPairs),
		"latency_pairs", len(allLatencyPairs))

	return &metrics.ServiceGraphMetrics{
		Pairs: pairs,
	}, nil
}

// buildServiceConnectionQuery builds a Prometheus query from a template for service connections
func (p *Provider) buildServiceConnectionQuery(tmpl *template.Template, serviceName, serviceNamespace string, filters metrics.MeshMetricsFilters, timeRange string) (string, error) {
	data := serviceConnectionsQueryTemplateData{
		FilterClause:     p.buildFilterClause(filters),
		TimeRange:        timeRange,
		ServiceName:      serviceName,
		ServiceNamespace: serviceNamespace,
	}

	return p.executeTemplate(tmpl, data)
}

// buildFilterClause builds Prometheus filter clause from metrics filters
func (p *Provider) buildFilterClause(filters metrics.MeshMetricsFilters) string {
	var clauses []string

	// Add namespace filter if specified
	if len(filters.Namespaces) > 0 {
		namespaces := make([]string, len(filters.Namespaces))
		for i, ns := range filters.Namespaces {
			namespaces[i] = fmt.Sprintf(`"%s"`, ns)
		}
		clauses = append(clauses, fmt.Sprintf(`source_workload_namespace=~"%s"`, strings.Join(namespaces, "|")))
	}

	if len(clauses) == 0 {
		return ""
	}

	return ", " + strings.Join(clauses, ", ")
}

// executeTemplate executes a template with the given data
func (p *Provider) executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}
