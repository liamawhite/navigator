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
	"time"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
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

// getStringValue safely extracts string values from Prometheus metric labels
func (p *Provider) getStringValue(metric model.Metric, key string) string {
	if value, ok := metric[model.LabelName(key)]; ok {
		return string(value)
	}
	return ""
}
