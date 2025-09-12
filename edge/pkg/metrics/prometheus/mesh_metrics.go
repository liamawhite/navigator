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
	"github.com/prometheus/common/model"
)

// Query templates for Prometheus
var (
	errorRateQueryTemplate = template.Must(template.New("errorRate").Parse(`
max by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter=~"source|destination", response_code=~"0|4..|5.."{{.FilterClause}}}[{{.TimeRange}}])
)`))

	requestRateQueryTemplate = template.Must(template.New("requestRate").Parse(`
max by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter=~"source|destination"{{.FilterClause}}}[{{.TimeRange}}])
)`))
)

// queryTemplateData holds the data for query templates
type queryTemplateData struct {
	FilterClause string
	TimeRange    string
}

// queryResult holds the result of a Prometheus query execution
type queryResult struct {
	Response  model.Value
	Error     error
	QueryType string
}

// processedMetrics holds the processed metrics from a query response
type processedMetrics struct {
	PairData   map[string]*metrics.ServicePairMetrics
	Error      error
	MetricType string
}

// formatPrometheusTimeRange converts a Go duration to a Prometheus-compatible time range string
func formatPrometheusTimeRange(duration time.Duration) string {
	// Convert to seconds for simplicity
	seconds := int(duration.Seconds())

	// Use minutes if it's a clean multiple of 60 seconds
	if seconds >= 60 && seconds%60 == 0 {
		minutes := seconds / 60
		return fmt.Sprintf("%dm", minutes)
	}

	// Use seconds for everything else, with a minimum of 60s
	if seconds < 60 {
		seconds = 60
	}
	return fmt.Sprintf("%ds", seconds)
}

// GetServiceGraphMetrics retrieves service-to-service metrics across the service graph
func (p *Provider) GetServiceGraphMetrics(ctx context.Context, query metrics.MeshMetricsQuery) (*metrics.ServiceGraphMetrics, error) {
	// Use the start and end times from the query
	startTime := query.StartTime
	endTime := query.EndTime

	// Calculate the time range duration for queries that need it
	duration := endTime.Sub(startTime)
	timeRange := formatPrometheusTimeRange(duration)

	p.logger.Debug("executing parallel mesh metrics queries", "time_range", timeRange, "filters", query.Filters)

	// Execute both queries in parallel
	errorResults, requestResults := p.executeQueriesInParallel(ctx, query.Filters, timeRange)

	// Check for errors
	if errorResults.Error != nil {
		p.logger.Error("error rate query failed", "error", errorResults.Error)
		return nil, fmt.Errorf("failed to query error rates: %w", errorResults.Error)
	}
	if requestResults.Error != nil {
		p.logger.Debug("request rate query failed, continuing without request rates", "error", requestResults.Error)
		// Continue without request rates rather than failing completely
	}

	p.logger.Debug("parallel queries completed", "error_response_type", fmt.Sprintf("%T", errorResults.Response),
		"request_response_type", fmt.Sprintf("%T", requestResults.Response))

	// Merge the results into service pair metrics
	pairs, err := p.mergeQueryResults(errorResults.Response, requestResults.Response)
	if err != nil {
		p.logger.Error("failed to merge query results", "error", err)
		return nil, fmt.Errorf("failed to merge query results: %w", err)
	}

	p.logger.Debug("merged query results", "pairs_count", len(pairs))

	result := &metrics.ServiceGraphMetrics{
		Pairs:     pairs,
		ClusterID: "", // Will be filled in by the caller
		Timestamp: time.Now(),
	}

	return result, nil
}

// executeQueriesInParallel executes error rate and request rate queries concurrently
func (p *Provider) executeQueriesInParallel(ctx context.Context, filters metrics.MeshMetricsFilters, timeRange string) (queryResult, queryResult) {
	var wg sync.WaitGroup
	var errorResults, requestResults queryResult

	// Execute error rate query
	wg.Add(1)
	go func() {
		defer wg.Done()

		query, err := p.buildQueryFromTemplate(errorRateQueryTemplate, filters, timeRange)
		if err != nil {
			errorResults = queryResult{Error: fmt.Errorf("failed to build error rate query: %w", err), QueryType: "error_rate"}
			return
		}

		p.logger.Debug("executing error rate query", "query", query)
		resp, err := p.client.query(ctx, query)
		errorResults = queryResult{Response: resp, Error: err, QueryType: "error_rate"}

		if err != nil {
			p.logger.Error("error rate query failed", "query", query, "error", err)
		}
	}()

	// Execute request rate query
	wg.Add(1)
	go func() {
		defer wg.Done()

		query, err := p.buildQueryFromTemplate(requestRateQueryTemplate, filters, timeRange)
		if err != nil {
			requestResults = queryResult{Error: fmt.Errorf("failed to build request rate query: %w", err), QueryType: "request_rate"}
			return
		}

		p.logger.Debug("executing request rate query", "query", query)
		resp, err := p.client.query(ctx, query)
		requestResults = queryResult{Response: resp, Error: err, QueryType: "request_rate"}

		if err != nil {
			p.logger.Error("request rate query failed", "query", query, "error", err)
		}
	}()

	wg.Wait()
	return errorResults, requestResults
}

// mergeQueryResults combines error rate and request rate query responses into service pair metrics using parallel processing
func (p *Provider) mergeQueryResults(errorResponse, requestResponse model.Value) ([]metrics.ServicePairMetrics, error) {
	timestamp := time.Now()

	// Process both response types in parallel
	errorMetrics, requestMetrics := p.processResponsesInParallel(errorResponse, requestResponse, timestamp)

	// Check for processing errors
	if errorMetrics.Error != nil {
		return nil, fmt.Errorf("failed to process error rates: %w", errorMetrics.Error)
	}
	if requestMetrics.Error != nil {
		p.logger.Debug("failed to process request rates, continuing with error data only", "error", requestMetrics.Error)
		// Continue with just error data
	}

	// Merge the processed data
	finalMap := p.mergePairMaps(errorMetrics.PairData, requestMetrics.PairData)

	// Convert map to slice
	pairs := make([]metrics.ServicePairMetrics, 0, len(finalMap))
	for _, pair := range finalMap {
		pairs = append(pairs, *pair)
	}

	return pairs, nil
}

// processResponsesInParallel processes error rate and request rate responses concurrently
func (p *Provider) processResponsesInParallel(errorResponse, requestResponse model.Value, timestamp time.Time) (processedMetrics, processedMetrics) {
	var wg sync.WaitGroup
	var errorMetrics, requestMetrics processedMetrics

	p.logger.Debug("starting parallel response processing")

	// Process error rate response
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorMetrics = p.processErrorRateResponse(errorResponse, timestamp)
	}()

	// Process request rate response
	wg.Add(1)
	go func() {
		defer wg.Done()
		requestMetrics = p.processRequestRateResponse(requestResponse, timestamp)
	}()

	wg.Wait()
	p.logger.Debug("parallel response processing completed",
		"error_pairs", len(errorMetrics.PairData),
		"request_pairs", len(requestMetrics.PairData))

	return errorMetrics, requestMetrics
}

// processErrorRateResponse processes error rate response data
func (p *Provider) processErrorRateResponse(response model.Value, timestamp time.Time) processedMetrics {
	pairMap := make(map[string]*metrics.ServicePairMetrics)

	if response == nil {
		return processedMetrics{PairData: pairMap, MetricType: "error_rate"}
	}

	errorVector, ok := response.(model.Vector)
	if !ok {
		return processedMetrics{
			Error:      fmt.Errorf("expected Vector result for error rates, got %T", response),
			MetricType: "error_rate",
		}
	}

	for _, sample := range errorVector {
		errorRate := float64(sample.Value)

		// Convert metric labels to string map
		labels := make(map[string]string)
		for key, value := range sample.Metric {
			labels[string(key)] = string(value)
		}

		pairKey := p.createPairKey(labels)
		pair := &metrics.ServicePairMetrics{
			SourceCluster:        getStringValue(labels, "source_cluster"),
			SourceNamespace:      getStringValue(labels, "source_workload_namespace"),
			SourceService:        getStringValue(labels, "source_canonical_service"),
			DestinationCluster:   getStringValue(labels, "destination_cluster"),
			DestinationNamespace: getStringValue(labels, "destination_service_namespace"),
			DestinationService:   getStringValue(labels, "destination_canonical_service"),
			ErrorRate:            errorRate,
			RequestRate:          0.0, // Will be set during merge
			Timestamp:            timestamp,
		}

		pairMap[pairKey] = pair
	}

	return processedMetrics{PairData: pairMap, MetricType: "error_rate"}
}

// processRequestRateResponse processes request rate response data
func (p *Provider) processRequestRateResponse(response model.Value, timestamp time.Time) processedMetrics {
	pairMap := make(map[string]*metrics.ServicePairMetrics)

	if response == nil {
		return processedMetrics{PairData: pairMap, MetricType: "request_rate"}
	}

	requestVector, ok := response.(model.Vector)
	if !ok {
		return processedMetrics{
			Error:      fmt.Errorf("expected Vector result for request rates, got %T", response),
			MetricType: "request_rate",
		}
	}

	for _, sample := range requestVector {
		rate := float64(sample.Value)

		// Convert metric labels to string map
		labels := make(map[string]string)
		for key, value := range sample.Metric {
			labels[string(key)] = string(value)
		}

		pairKey := p.createPairKey(labels)
		pair := &metrics.ServicePairMetrics{
			SourceCluster:        getStringValue(labels, "source_cluster"),
			SourceNamespace:      getStringValue(labels, "source_workload_namespace"),
			SourceService:        getStringValue(labels, "source_canonical_service"),
			DestinationCluster:   getStringValue(labels, "destination_cluster"),
			DestinationNamespace: getStringValue(labels, "destination_service_namespace"),
			DestinationService:   getStringValue(labels, "destination_canonical_service"),
			ErrorRate:            0.0, // Will be set during merge
			RequestRate:          rate,
			Timestamp:            timestamp,
		}

		pairMap[pairKey] = pair
	}

	return processedMetrics{PairData: pairMap, MetricType: "request_rate"}
}

// mergePairMaps efficiently merges error rate and request rate pair maps
func (p *Provider) mergePairMaps(errorPairs, requestPairs map[string]*metrics.ServicePairMetrics) map[string]*metrics.ServicePairMetrics {
	// Start with error pairs as base
	finalMap := make(map[string]*metrics.ServicePairMetrics, len(errorPairs)+len(requestPairs))

	// Copy all error pairs
	for key, pair := range errorPairs {
		finalMap[key] = pair
	}

	// Merge request rate data
	for key, requestPair := range requestPairs {
		if existingPair, exists := finalMap[key]; exists {
			// Update existing pair with request rate
			existingPair.RequestRate = requestPair.RequestRate
		} else {
			// Add new pair with only request rate data
			finalMap[key] = requestPair
		}
	}

	return finalMap
}

// buildQueryFromTemplate builds a Prometheus query using a template
func (p *Provider) buildQueryFromTemplate(tmpl *template.Template, filters metrics.MeshMetricsFilters, timeRange string) (string, error) {
	data := queryTemplateData{
		FilterClause: p.buildFilterClause(filters),
		TimeRange:    timeRange,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute query template: %w", err)
	}

	// Clean up whitespace for more readable logs
	query := strings.TrimSpace(buf.String())
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", "")

	// Collapse multiple spaces into single spaces
	for strings.Contains(query, "  ") {
		query = strings.ReplaceAll(query, "  ", " ")
	}

	return query, nil
}

// buildFilterClause builds Prometheus filter clauses based on the provided filters
func (p *Provider) buildFilterClause(filters metrics.MeshMetricsFilters) string {
	var clauses []string

	// For service connections, we want mesh-wide visibility
	// Remove cluster filtering to see cross-cluster connections
	
	if len(filters.Namespaces) > 0 {
		namespaces := strings.Join(filters.Namespaces, "|")
		clauses = append(clauses, fmt.Sprintf(`destination_namespace=~"%s"`, namespaces))
	}

	if len(clauses) > 0 {
		return ", " + strings.Join(clauses, ", ")
	}
	return ""
}

// createPairKey creates a unique key for a service pair from Prometheus labels
func (p *Provider) createPairKey(metric map[string]string) string {
	return fmt.Sprintf("%s:%s:%s->%s:%s:%s",
		getStringValue(metric, "source_cluster"),
		getStringValue(metric, "source_workload_namespace"),
		getStringValue(metric, "source_canonical_service"),
		getStringValue(metric, "destination_cluster"),
		getStringValue(metric, "destination_service_namespace"),
		getStringValue(metric, "destination_canonical_service"),
	)
}

// getStringValue safely gets a string value from a map with empty string as default
func getStringValue(metric map[string]string, key string) string {
	if value, exists := metric[key]; exists {
		return value
	}
	return ""
}
