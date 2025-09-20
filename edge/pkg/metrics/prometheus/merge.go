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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	sharedmetrics "github.com/liamawhite/navigator/pkg/metrics"
	"github.com/prometheus/common/model"
)

// processedMetrics represents the result of processing a Prometheus response
type processedMetrics struct {
	PairData   map[string]*metrics.ServicePairMetrics
	MetricType string
	Error      error
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
		key := p.createPairKey(sample.Metric)
		if key == "" {
			continue
		}

		pair := &metrics.ServicePairMetrics{
			SourceCluster:        p.getStringValue(sample.Metric, "source_cluster"),
			SourceNamespace:      p.getStringValue(sample.Metric, "source_workload_namespace"),
			SourceService:        p.getStringValue(sample.Metric, "source_canonical_service"),
			DestinationCluster:   p.getStringValue(sample.Metric, "destination_cluster"),
			DestinationNamespace: p.getStringValue(sample.Metric, "destination_service_namespace"),
			DestinationService:   p.getStringValue(sample.Metric, "destination_canonical_service"),
			ErrorRate:            float64(sample.Value),
		}

		pairMap[key] = pair
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
		key := p.createPairKey(sample.Metric)
		if key == "" {
			continue
		}

		pair := &metrics.ServicePairMetrics{
			SourceCluster:        p.getStringValue(sample.Metric, "source_cluster"),
			SourceNamespace:      p.getStringValue(sample.Metric, "source_workload_namespace"),
			SourceService:        p.getStringValue(sample.Metric, "source_canonical_service"),
			DestinationCluster:   p.getStringValue(sample.Metric, "destination_cluster"),
			DestinationNamespace: p.getStringValue(sample.Metric, "destination_service_namespace"),
			DestinationService:   p.getStringValue(sample.Metric, "destination_canonical_service"),
			RequestRate:          float64(sample.Value),
		}

		pairMap[key] = pair
	}

	return processedMetrics{PairData: pairMap, MetricType: "request_rate"}
}

// processLatencyResponse processes latency P99 response data
func (p *Provider) processLatencyResponse(response model.Value, timestamp time.Time) processedMetrics {
	pairMap := make(map[string]*metrics.ServicePairMetrics)

	if response == nil {
		return processedMetrics{PairData: pairMap, MetricType: "latency_p99"}
	}

	latencyVector, ok := response.(model.Vector)
	if !ok {
		return processedMetrics{
			Error:      fmt.Errorf("expected Vector result for latency P99, got %T", response),
			MetricType: "latency_p99",
		}
	}

	for _, sample := range latencyVector {
		key := p.createPairKey(sample.Metric)
		if key == "" {
			continue
		}

		// Handle NaN values and keep as-is (check if already in milliseconds)
		latencyMs := float64(sample.Value)
		if latencyMs != latencyMs { // Check for NaN
			latencyMs = 0.0
		}

		pair := &metrics.ServicePairMetrics{
			SourceCluster:        p.getStringValue(sample.Metric, "source_cluster"),
			SourceNamespace:      p.getStringValue(sample.Metric, "source_workload_namespace"),
			SourceService:        p.getStringValue(sample.Metric, "source_canonical_service"),
			DestinationCluster:   p.getStringValue(sample.Metric, "destination_cluster"),
			DestinationNamespace: p.getStringValue(sample.Metric, "destination_service_namespace"),
			DestinationService:   p.getStringValue(sample.Metric, "destination_canonical_service"),
			LatencyP99:           latencyMs,
		}

		pairMap[key] = pair
	}

	return processedMetrics{PairData: pairMap, MetricType: "latency_p99"}
}

// processLatencyDistributionResponse processes raw histogram distribution response data.
//
// Key design decisions:
// 1. Collects raw histogram buckets from Prometheus instead of using histogram_quantile()
// 2. Reduces Prometheus query load by fetching raw distributions once
// 3. Enables histogram aggregation at multiple levels before percentile calculation
// 4. Calculates P99 at edge to distribute computational load
// 5. Preserves raw histogram data for potential manager-side cross-cluster aggregation
func (p *Provider) processLatencyDistributionResponse(response model.Value, timestamp time.Time) processedMetrics {
	pairMap := make(map[string]*metrics.ServicePairMetrics)

	if response == nil {
		return processedMetrics{PairData: pairMap, MetricType: "latency_distribution"}
	}

	distributionVector, ok := response.(model.Vector)
	if !ok {
		return processedMetrics{
			Error:      fmt.Errorf("expected Vector result for latency distribution, got %T", response),
			MetricType: "latency_distribution",
		}
	}

	// Build lookup map for O(1) source pair metadata access
	sampleLookup := make(map[string]*model.Sample)
	for _, sample := range distributionVector {
		key := p.createPairKey(sample.Metric)
		if key != "" && sampleLookup[key] == nil {
			sampleLookup[key] = sample
		}
	}

	// Group buckets by service pair
	pairBuckets := make(map[string]map[float64]float64) // pairKey -> (le -> count)
	pairTotalCount := make(map[string]float64)          // pairKey -> total count

	for _, sample := range distributionVector {
		key := p.createPairKey(sample.Metric)
		if key == "" {
			continue
		}

		// Get le (less-than-or-equal) value from the bucket
		leStr := p.getStringValue(sample.Metric, "le")
		if leStr == "" || leStr == "+Inf" {
			// Skip +Inf bucket for now, we'll calculate total count separately
			continue
		}

		le, err := strconv.ParseFloat(leStr, 64)
		if err != nil {
			p.logger.Warn("failed to parse le value", "le", leStr, "error", err)
			continue
		}

		if pairBuckets[key] == nil {
			pairBuckets[key] = make(map[float64]float64)
		}

		count := float64(sample.Value)
		pairBuckets[key][le] = count

		// Keep track of highest bucket count as total count approximation
		if count > pairTotalCount[key] {
			pairTotalCount[key] = count
		}
	}

	// Convert to ServicePairMetrics with distributions
	for pairKey, buckets := range pairBuckets {
		// Create sorted buckets
		var histogramBuckets []*typesv1alpha1.HistogramBucket
		var les []float64
		for le := range buckets {
			les = append(les, le)
		}
		sort.Float64s(les)

		for _, le := range les {
			histogramBuckets = append(histogramBuckets, &typesv1alpha1.HistogramBucket{
				Le:    le,
				Count: buckets[le],
			})
		}

		// Extract service pair info using O(1) lookup
		sourceSample := sampleLookup[pairKey]
		if sourceSample == nil {
			continue
		}

		sourcePair := &metrics.ServicePairMetrics{
			SourceCluster:        p.getStringValue(sourceSample.Metric, "source_cluster"),
			SourceNamespace:      p.getStringValue(sourceSample.Metric, "source_workload_namespace"),
			SourceService:        p.getStringValue(sourceSample.Metric, "source_canonical_service"),
			DestinationCluster:   p.getStringValue(sourceSample.Metric, "destination_cluster"),
			DestinationNamespace: p.getStringValue(sourceSample.Metric, "destination_service_namespace"),
			DestinationService:   p.getStringValue(sourceSample.Metric, "destination_canonical_service"),
		}

		// Calculate sum as approximation (we don't have exact sum from rate queries)
		// This is an approximation based on bucket midpoints
		var sum float64
		for i, bucket := range histogramBuckets {
			var lowerBound float64
			if i > 0 {
				lowerBound = histogramBuckets[i-1].Le
			}
			midpoint := (lowerBound + bucket.Le) / 2
			if i == 0 {
				sum += midpoint * bucket.Count
			} else {
				// Calculate count for this bucket (difference from cumulative)
				bucketCount := bucket.Count - histogramBuckets[i-1].Count
				sum += midpoint * bucketCount
			}
		}

		// Create the distribution
		distribution := &typesv1alpha1.LatencyDistribution{
			Buckets:    histogramBuckets,
			TotalCount: pairTotalCount[pairKey],
			Sum:        sum,
		}

		// Calculate P99 from distribution using Prometheus's quantile algorithm.
		// This is done at the edge to distribute computational load while preserving
		// the raw histogram for potential cross-cluster aggregation at the manager.
		var p99Ms float64
		if calculated, err := sharedmetrics.CalculateP99(distribution); err == nil {
			p99Ms = calculated
		}

		// Store both calculated P99 and raw distribution for flexibility:
		// - P99 for immediate use and API responses
		// - Raw distribution for manager-side histogram merging across clusters
		sourcePair.LatencyP99 = p99Ms
		sourcePair.LatencyDistribution = distribution

		pairMap[pairKey] = sourcePair
	}

	return processedMetrics{PairData: pairMap, MetricType: "latency_distribution"}
}

// processDownstreamLatencyDistributionResponse processes downstream latency distribution for gateways
func (p *Provider) processDownstreamLatencyDistributionResponse(response model.Value, timestamp time.Time, serviceName string) processedMetrics {
	pairMap := make(map[string]*metrics.ServicePairMetrics)

	if response == nil {
		p.logger.Debug("downstream latency distribution response is nil", "service_name", serviceName)
		return processedMetrics{PairData: pairMap, MetricType: "gateway_downstream_latency_distribution"}
	}

	distributionVector, ok := response.(model.Vector)
	if !ok {
		p.logger.Debug("downstream latency distribution response wrong type", "service_name", serviceName, "type", fmt.Sprintf("%T", response))
		return processedMetrics{
			Error:      fmt.Errorf("expected Vector result for downstream latency distribution, got %T", response),
			MetricType: "gateway_downstream_latency_distribution",
		}
	}

	p.logger.Debug("processing downstream latency distribution", "service_name", serviceName, "samples_count", len(distributionVector))

	// Group buckets by pod/namespace combination
	podBuckets := make(map[string]map[float64]float64) // podKey -> (le -> count)
	podTotalCount := make(map[string]float64)          // podKey -> total count

	for _, sample := range distributionVector {
		pod := p.getStringValue(sample.Metric, "pod")
		namespace := p.getStringValue(sample.Metric, "namespace")

		if pod == "" || namespace == "" {
			continue
		}

		// Get le (less-than-or-equal) value from the bucket
		leStr := p.getStringValue(sample.Metric, "le")
		if leStr == "" || leStr == "+Inf" {
			// Skip +Inf bucket for now
			continue
		}

		le, err := strconv.ParseFloat(leStr, 64)
		if err != nil {
			p.logger.Warn("failed to parse le value", "le", leStr, "error", err)
			continue
		}

		podKey := fmt.Sprintf("%s:%s", namespace, pod)
		if podBuckets[podKey] == nil {
			podBuckets[podKey] = make(map[float64]float64)
		}

		count := float64(sample.Value)
		podBuckets[podKey][le] = count

		// Keep track of highest bucket count as total count approximation
		if count > podTotalCount[podKey] {
			podTotalCount[podKey] = count
		}
	}

	// Convert to ServicePairMetrics with distributions
	for podKey, buckets := range podBuckets {
		// Create sorted buckets
		var histogramBuckets []*typesv1alpha1.HistogramBucket
		var les []float64
		for le := range buckets {
			les = append(les, le)
		}
		sort.Float64s(les)

		for _, le := range les {
			histogramBuckets = append(histogramBuckets, &typesv1alpha1.HistogramBucket{
				Le:    le,
				Count: buckets[le],
			})
		}

		// Calculate sum as approximation
		var sum float64
		for i, bucket := range histogramBuckets {
			var lowerBound float64
			if i > 0 {
				lowerBound = histogramBuckets[i-1].Le
			}
			midpoint := (lowerBound + bucket.Le) / 2
			if i == 0 {
				sum += midpoint * bucket.Count
			} else {
				bucketCount := bucket.Count - histogramBuckets[i-1].Count
				sum += midpoint * bucketCount
			}
		}

		// Create the distribution
		distribution := &typesv1alpha1.LatencyDistribution{
			Buckets:    histogramBuckets,
			TotalCount: podTotalCount[podKey],
			Sum:        sum,
		}

		// Calculate P99 from distribution
		var p99Ms float64
		if calculated, err := sharedmetrics.CalculateP99(distribution); err == nil {
			p99Ms = calculated
		}

		// Create a special pair for gateway downstream metrics
		// Use same key format as downstream request rate for proper merging
		namespace := podKey[:strings.Index(podKey, ":")]
		pairKey := fmt.Sprintf("unknown:->%s:%s:%s", p.clusterName, namespace, serviceName)

		pair := &metrics.ServicePairMetrics{
			SourceCluster:        "unknown",
			SourceNamespace:      "unknown",
			SourceService:        "unknown",
			DestinationCluster:   p.clusterName,
			DestinationNamespace: namespace,
			DestinationService:   serviceName,
			LatencyP99:           p99Ms, // Calculated from distribution
			LatencyDistribution:  distribution,
		}

		pairMap[pairKey] = pair
	}

	p.logger.Debug("completed downstream latency distribution processing", "service_name", serviceName, "pairs_created", len(pairMap))
	return processedMetrics{PairData: pairMap, MetricType: "gateway_downstream_latency_distribution"}
}

// mergePairMaps merges request rate, error rate, and latency data
func (p *Provider) mergePairMaps(requestPairs, errorPairs, latencyPairs map[string]*metrics.ServicePairMetrics) map[string]*metrics.ServicePairMetrics {
	merged := make(map[string]*metrics.ServicePairMetrics)

	// Start with request rate data
	for key, pair := range requestPairs {
		merged[key] = &metrics.ServicePairMetrics{
			SourceCluster:        pair.SourceCluster,
			SourceNamespace:      pair.SourceNamespace,
			SourceService:        pair.SourceService,
			DestinationCluster:   pair.DestinationCluster,
			DestinationNamespace: pair.DestinationNamespace,
			DestinationService:   pair.DestinationService,
			RequestRate:          pair.RequestRate,
			ErrorRate:            0.0, // Default to 0
			LatencyP99:           0.0, // Default to 0
		}
	}

	// Add error rate data
	for key, errorPair := range errorPairs {
		if existing, exists := merged[key]; exists {
			existing.ErrorRate = errorPair.ErrorRate
		} else {
			// Create new pair with just error rate
			merged[key] = &metrics.ServicePairMetrics{
				SourceCluster:        errorPair.SourceCluster,
				SourceNamespace:      errorPair.SourceNamespace,
				SourceService:        errorPair.SourceService,
				DestinationCluster:   errorPair.DestinationCluster,
				DestinationNamespace: errorPair.DestinationNamespace,
				DestinationService:   errorPair.DestinationService,
				RequestRate:          0.0, // Default to 0
				ErrorRate:            errorPair.ErrorRate,
				LatencyP99:           0.0, // Default to 0
			}
		}
	}

	// Add latency data
	for key, latencyPair := range latencyPairs {
		if existing, exists := merged[key]; exists {
			existing.LatencyP99 = latencyPair.LatencyP99
			existing.LatencyDistribution = latencyPair.LatencyDistribution
		} else {
			// Create new pair with just latency
			merged[key] = &metrics.ServicePairMetrics{
				SourceCluster:        latencyPair.SourceCluster,
				SourceNamespace:      latencyPair.SourceNamespace,
				SourceService:        latencyPair.SourceService,
				DestinationCluster:   latencyPair.DestinationCluster,
				DestinationNamespace: latencyPair.DestinationNamespace,
				DestinationService:   latencyPair.DestinationService,
				RequestRate:          0.0, // Default to 0
				ErrorRate:            0.0, // Default to 0
				LatencyP99:           latencyPair.LatencyP99,
				LatencyDistribution:  latencyPair.LatencyDistribution,
			}
		}
	}

	return merged
}

// createPairKey creates a unique key for service pair metrics
func (p *Provider) createPairKey(metric model.Metric) string {
	source := p.getStringValue(metric, "source_canonical_service")
	sourceNs := p.getStringValue(metric, "source_workload_namespace")
	dest := p.getStringValue(metric, "destination_canonical_service")
	destNs := p.getStringValue(metric, "destination_service_namespace")
	srcCluster := p.getStringValue(metric, "source_cluster")
	destCluster := p.getStringValue(metric, "destination_cluster")

	if source == "" || dest == "" {
		return ""
	}

	return fmt.Sprintf("%s:%s:%s->%s:%s:%s", srcCluster, sourceNs, source, destCluster, destNs, dest)
}

// processDownstreamRequestRateResponse processes downstream request rate for gateways
func (p *Provider) processDownstreamRequestRateResponse(response model.Value, timestamp time.Time, serviceName string) processedMetrics {
	pairMap := make(map[string]*metrics.ServicePairMetrics)

	if response == nil {
		return processedMetrics{PairData: pairMap, MetricType: "gateway_downstream_request_rate"}
	}

	requestVector, ok := response.(model.Vector)
	if !ok {
		return processedMetrics{
			Error:      fmt.Errorf("expected Vector result for downstream request rates, got %T", response),
			MetricType: "gateway_downstream_request_rate",
		}
	}

	for _, sample := range requestVector {
		pod := p.getStringValue(sample.Metric, "pod")
		namespace := p.getStringValue(sample.Metric, "namespace")

		if pod == "" || namespace == "" {
			continue
		}

		// Create a special pair for gateway downstream metrics
		// Use "unknown" as source since this represents inbound traffic from outside
		key := fmt.Sprintf("unknown:->%s:%s:%s", p.clusterName, namespace, serviceName)

		pair := &metrics.ServicePairMetrics{
			SourceCluster:        "unknown",
			SourceNamespace:      "unknown",
			SourceService:        "unknown",
			DestinationCluster:   p.clusterName,
			DestinationNamespace: namespace,
			DestinationService:   serviceName, // Use service name instead of pod name
			RequestRate:          float64(sample.Value),
		}

		pairMap[key] = pair
	}

	return processedMetrics{PairData: pairMap, MetricType: "gateway_downstream_request_rate"}
}

// mergePairMapsWithDistributions merges request rate, error rate, and latency distribution data
func (p *Provider) mergePairMapsWithDistributions(requestPairs, errorPairs, distributionPairs map[string]*metrics.ServicePairMetrics) map[string]*metrics.ServicePairMetrics {
	merged := make(map[string]*metrics.ServicePairMetrics)

	// Start with request rate data
	for key, pair := range requestPairs {
		merged[key] = &metrics.ServicePairMetrics{
			SourceCluster:        pair.SourceCluster,
			SourceNamespace:      pair.SourceNamespace,
			SourceService:        pair.SourceService,
			DestinationCluster:   pair.DestinationCluster,
			DestinationNamespace: pair.DestinationNamespace,
			DestinationService:   pair.DestinationService,
			RequestRate:          pair.RequestRate,
			ErrorRate:            0.0, // Default to 0
			LatencyP99:           0.0, // Default to 0 (will be calculated by manager)
		}
	}

	// Add error rate data
	for key, errorPair := range errorPairs {
		if existing, exists := merged[key]; exists {
			existing.ErrorRate = errorPair.ErrorRate
		} else {
			// Create new pair with just error rate
			merged[key] = &metrics.ServicePairMetrics{
				SourceCluster:        errorPair.SourceCluster,
				SourceNamespace:      errorPair.SourceNamespace,
				SourceService:        errorPair.SourceService,
				DestinationCluster:   errorPair.DestinationCluster,
				DestinationNamespace: errorPair.DestinationNamespace,
				DestinationService:   errorPair.DestinationService,
				RequestRate:          0.0, // Default to 0
				ErrorRate:            errorPair.ErrorRate,
				LatencyP99:           0.0, // Default to 0 (will be calculated by manager)
			}
		}
	}

	// Add latency distribution data and calculate P99
	for key, distributionPair := range distributionPairs {
		// Calculate P99 from distribution if available
		var p99Ms float64
		if distributionPair.LatencyDistribution != nil {
			if calculated, err := sharedmetrics.CalculateP99(distributionPair.LatencyDistribution); err == nil {
				p99Ms = calculated
			}
		}

		p.logger.Debug("processing distribution pair", "key", key, "source", distributionPair.SourceService, "dest", distributionPair.DestinationService, "has_distribution", distributionPair.LatencyDistribution != nil, "p99", p99Ms)

		if existing, exists := merged[key]; exists {
			p.logger.Debug("merging distribution into existing pair", "key", key)
			existing.LatencyDistribution = distributionPair.LatencyDistribution
			existing.LatencyP99 = p99Ms
		} else {
			p.logger.Debug("creating new pair with distribution only", "key", key)
			// Create new pair with latency distribution and calculated P99
			merged[key] = &metrics.ServicePairMetrics{
				SourceCluster:        distributionPair.SourceCluster,
				SourceNamespace:      distributionPair.SourceNamespace,
				SourceService:        distributionPair.SourceService,
				DestinationCluster:   distributionPair.DestinationCluster,
				DestinationNamespace: distributionPair.DestinationNamespace,
				DestinationService:   distributionPair.DestinationService,
				RequestRate:          0.0,   // Default to 0
				ErrorRate:            0.0,   // Default to 0
				LatencyP99:           p99Ms, // Calculated from distribution
				LatencyDistribution:  distributionPair.LatencyDistribution,
			}
		}
	}

	return merged
}

// getStringValue safely extracts string values from Prometheus metric labels
func (p *Provider) getStringValue(metric model.Metric, key string) string {
	if value, ok := metric[model.LabelName(key)]; ok {
		return string(value)
	}
	return ""
}
