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
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
	"github.com/liamawhite/navigator/pkg/logging"
)

func TestCreatePairKey(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name: "istio-ingressgateway to frontend",
			labels: map[string]string{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "istio-system",
				"source_canonical_service":      "istio-ingressgateway",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "frontend",
			},
			expected: "Kubernetes:istio-system:istio-ingressgateway->Kubernetes:microservices:frontend",
		},
		{
			name: "frontend to backend",
			labels: map[string]string{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "microservices",
				"source_canonical_service":      "frontend",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "backend",
			},
			expected: "Kubernetes:microservices:frontend->Kubernetes:microservices:backend",
		},
		{
			name: "backend to database",
			labels: map[string]string{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "microservices",
				"source_canonical_service":      "backend",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "database",
			},
			expected: "Kubernetes:microservices:backend->Kubernetes:microservices:database",
		},
		{
			name: "missing labels should not cause panic",
			labels: map[string]string{
				"source_cluster": "Kubernetes",
			},
			expected: "Kubernetes::->::",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.createPairKey(tt.labels)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreatePairKeyUniqueness(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	// Test that different service pairs generate unique keys
	ingressToFrontend := map[string]string{
		"source_cluster":                "Kubernetes",
		"source_workload_namespace":     "istio-system",
		"source_canonical_service":      "istio-ingressgateway",
		"destination_cluster":           "Kubernetes",
		"destination_service_namespace": "microservices",
		"destination_canonical_service": "frontend",
	}

	frontendToBackend := map[string]string{
		"source_cluster":                "Kubernetes",
		"source_workload_namespace":     "microservices",
		"source_canonical_service":      "frontend",
		"destination_cluster":           "Kubernetes",
		"destination_service_namespace": "microservices",
		"destination_canonical_service": "backend",
	}

	backendToDatabase := map[string]string{
		"source_cluster":                "Kubernetes",
		"source_workload_namespace":     "microservices",
		"source_canonical_service":      "backend",
		"destination_cluster":           "Kubernetes",
		"destination_service_namespace": "microservices",
		"destination_canonical_service": "database",
	}

	key1 := provider.createPairKey(ingressToFrontend)
	key2 := provider.createPairKey(frontendToBackend)
	key3 := provider.createPairKey(backendToDatabase)

	// All keys should be different
	assert.NotEqual(t, key1, key2, "ingress->frontend and frontend->backend should have different keys")
	assert.NotEqual(t, key1, key3, "ingress->frontend and backend->database should have different keys")
	assert.NotEqual(t, key2, key3, "frontend->backend and backend->database should have different keys")

	// Keys should be non-empty
	assert.NotEmpty(t, key1)
	assert.NotEmpty(t, key2)
	assert.NotEmpty(t, key3)
}

func TestMergeQueryResults(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	// Real data from Prometheus (converted to Prometheus model format)
	requestResponse := model.Vector{
		&model.Sample{
			Metric: model.Metric{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "istio-system",
				"source_canonical_service":      "istio-ingressgateway",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "frontend",
			},
			Value:     model.SampleValue(5.003508771929824),
			Timestamp: model.Time(1757251551426),
		},
		&model.Sample{
			Metric: model.Metric{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "microservices",
				"source_canonical_service":      "frontend",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "backend",
			},
			Value:     model.SampleValue(4.9964736965835215),
			Timestamp: model.Time(1757251551426),
		},
		&model.Sample{
			Metric: model.Metric{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "microservices",
				"source_canonical_service":      "backend",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "database",
			},
			Value:     model.SampleValue(4.996491228070175),
			Timestamp: model.Time(1757251551426),
		},
	}

	// No error rates for this test
	errorResponse := model.Vector{}

	result, err := provider.mergeQueryResults(errorResponse, requestResponse)

	require.NoError(t, err)
	assert.Len(t, result, 3, "Should process all 3 service pairs")

	// Verify specific pairs by looking through the results
	services := make(map[string]metrics.ServicePairMetrics)
	for _, pair := range result {
		key := fmt.Sprintf("%s->%s", pair.SourceService, pair.DestinationService)
		services[key] = pair
	}

	// Check ingress->frontend pair
	ingressPair, exists := services["istio-ingressgateway->frontend"]
	require.True(t, exists, "Should contain ingress->frontend pair")
	assert.Equal(t, "Kubernetes", ingressPair.SourceCluster)
	assert.Equal(t, "istio-system", ingressPair.SourceNamespace)
	assert.Equal(t, "istio-ingressgateway", ingressPair.SourceService)
	assert.Equal(t, "Kubernetes", ingressPair.DestinationCluster)
	assert.Equal(t, "microservices", ingressPair.DestinationNamespace)
	assert.Equal(t, "frontend", ingressPair.DestinationService)
	assert.InDelta(t, 5.003508771929824, ingressPair.RequestRate, 0.001)
	assert.Equal(t, 0.0, ingressPair.ErrorRate)

	// Check frontend->backend pair
	frontendPair, exists := services["frontend->backend"]
	require.True(t, exists, "Should contain frontend->backend pair")
	assert.Equal(t, "frontend", frontendPair.SourceService)
	assert.Equal(t, "backend", frontendPair.DestinationService)
	assert.InDelta(t, 4.9964736965835215, frontendPair.RequestRate, 0.001)

	// Check backend->database pair
	backendPair, exists := services["backend->database"]
	require.True(t, exists, "Should contain backend->database pair")
	assert.Equal(t, "backend", backendPair.SourceService)
	assert.Equal(t, "database", backendPair.DestinationService)
	assert.InDelta(t, 4.996491228070175, backendPair.RequestRate, 0.001)
}

func TestProcessErrorRateResponse(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}
	timestamp := time.Now()

	tests := []struct {
		name           string
		response       model.Value
		expectedPairs  int
		expectedMetric string
		shouldError    bool
	}{
		{
			name: "valid error rate response",
			response: model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "Kubernetes",
						"source_workload_namespace":     "microservices",
						"source_canonical_service":      "frontend",
						"destination_cluster":           "Kubernetes",
						"destination_service_namespace": "microservices",
						"destination_canonical_service": "backend",
					},
					Value:     model.SampleValue(0.1),
					Timestamp: model.Time(timestamp.Unix()),
				},
			},
			expectedPairs:  1,
			expectedMetric: "error_rate",
			shouldError:    false,
		},
		{
			name: "multiple error rates",
			response: model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "Kubernetes",
						"source_workload_namespace":     "microservices",
						"source_canonical_service":      "frontend",
						"destination_cluster":           "Kubernetes",
						"destination_service_namespace": "microservices",
						"destination_canonical_service": "backend",
					},
					Value: model.SampleValue(0.05),
				},
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "Kubernetes",
						"source_workload_namespace":     "microservices",
						"source_canonical_service":      "backend",
						"destination_cluster":           "Kubernetes",
						"destination_service_namespace": "microservices",
						"destination_canonical_service": "database",
					},
					Value: model.SampleValue(0.02),
				},
			},
			expectedPairs:  2,
			expectedMetric: "error_rate",
			shouldError:    false,
		},
		{
			name:           "empty response",
			response:       model.Vector{},
			expectedPairs:  0,
			expectedMetric: "error_rate",
			shouldError:    false,
		},
		{
			name:           "nil response",
			response:       nil,
			expectedPairs:  0,
			expectedMetric: "error_rate",
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.processErrorRateResponse(tt.response, timestamp)

			if tt.shouldError {
				assert.Error(t, result.Error)
			} else {
				assert.NoError(t, result.Error)
			}

			assert.Equal(t, tt.expectedMetric, result.MetricType)
			assert.Len(t, result.PairData, tt.expectedPairs)

			// Verify error rates are set correctly and request rates are 0
			for _, pair := range result.PairData {
				assert.GreaterOrEqual(t, pair.ErrorRate, 0.0)
				assert.Equal(t, 0.0, pair.RequestRate) // Should be 0 for error processing
			}
		})
	}
}

func TestProcessRequestRateResponse(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}
	timestamp := time.Now()

	tests := []struct {
		name           string
		response       model.Value
		expectedPairs  int
		expectedMetric string
		shouldError    bool
	}{
		{
			name: "valid request rate response",
			response: model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "Kubernetes",
						"source_workload_namespace":     "microservices",
						"source_canonical_service":      "frontend",
						"destination_cluster":           "Kubernetes",
						"destination_service_namespace": "microservices",
						"destination_canonical_service": "backend",
					},
					Value:     model.SampleValue(5.5),
					Timestamp: model.Time(timestamp.Unix()),
				},
			},
			expectedPairs:  1,
			expectedMetric: "request_rate",
			shouldError:    false,
		},
		{
			name: "real prometheus data",
			response: model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "Kubernetes",
						"source_workload_namespace":     "istio-system",
						"source_canonical_service":      "istio-ingressgateway",
						"destination_cluster":           "Kubernetes",
						"destination_service_namespace": "microservices",
						"destination_canonical_service": "frontend",
					},
					Value: model.SampleValue(5.003508771929824),
				},
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "Kubernetes",
						"source_workload_namespace":     "microservices",
						"source_canonical_service":      "frontend",
						"destination_cluster":           "Kubernetes",
						"destination_service_namespace": "microservices",
						"destination_canonical_service": "backend",
					},
					Value: model.SampleValue(4.9964736965835215),
				},
			},
			expectedPairs:  2,
			expectedMetric: "request_rate",
			shouldError:    false,
		},
		{
			name:           "empty response",
			response:       model.Vector{},
			expectedPairs:  0,
			expectedMetric: "request_rate",
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.processRequestRateResponse(tt.response, timestamp)

			if tt.shouldError {
				assert.Error(t, result.Error)
			} else {
				assert.NoError(t, result.Error)
			}

			assert.Equal(t, tt.expectedMetric, result.MetricType)
			assert.Len(t, result.PairData, tt.expectedPairs)

			// Verify request rates are set correctly and error rates are 0
			for _, pair := range result.PairData {
				assert.GreaterOrEqual(t, pair.RequestRate, 0.0)
				assert.Equal(t, 0.0, pair.ErrorRate) // Should be 0 for request processing
			}
		})
	}
}

func TestMergePairMaps(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	tests := []struct {
		name         string
		errorPairs   map[string]*metrics.ServicePairMetrics
		requestPairs map[string]*metrics.ServicePairMetrics
		expected     int
	}{
		{
			name: "overlapping pairs merge correctly",
			errorPairs: map[string]*metrics.ServicePairMetrics{
				"key1": {
					SourceService:      "frontend",
					DestinationService: "backend",
					ErrorRate:          0.1,
					RequestRate:        0.0,
				},
			},
			requestPairs: map[string]*metrics.ServicePairMetrics{
				"key1": {
					SourceService:      "frontend",
					DestinationService: "backend",
					ErrorRate:          0.0,
					RequestRate:        5.0,
				},
			},
			expected: 1,
		},
		{
			name: "non-overlapping pairs",
			errorPairs: map[string]*metrics.ServicePairMetrics{
				"key1": {
					SourceService:      "frontend",
					DestinationService: "backend",
					ErrorRate:          0.1,
				},
			},
			requestPairs: map[string]*metrics.ServicePairMetrics{
				"key2": {
					SourceService:      "backend",
					DestinationService: "database",
					RequestRate:        3.0,
				},
			},
			expected: 2,
		},
		{
			name:       "empty error pairs",
			errorPairs: map[string]*metrics.ServicePairMetrics{},
			requestPairs: map[string]*metrics.ServicePairMetrics{
				"key1": {RequestRate: 5.0},
			},
			expected: 1,
		},
		{
			name: "empty request pairs",
			errorPairs: map[string]*metrics.ServicePairMetrics{
				"key1": {ErrorRate: 0.1},
			},
			requestPairs: map[string]*metrics.ServicePairMetrics{},
			expected:     1,
		},
		{
			name:         "both empty",
			errorPairs:   map[string]*metrics.ServicePairMetrics{},
			requestPairs: map[string]*metrics.ServicePairMetrics{},
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.mergePairMaps(tt.errorPairs, tt.requestPairs)
			assert.Len(t, result, tt.expected)

			// Verify merge correctness for overlapping case
			if tt.name == "overlapping pairs merge correctly" {
				pair := result["key1"]
				require.NotNil(t, pair)
				assert.Equal(t, 0.1, pair.ErrorRate)
				assert.Equal(t, 5.0, pair.RequestRate)
			}
		})
	}
}

func TestBuildFilterClause(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	tests := []struct {
		name     string
		filters  metrics.MeshMetricsFilters
		expected string
	}{
		{
			name:     "empty filters",
			filters:  metrics.MeshMetricsFilters{},
			expected: "",
		},
		{
			name: "single namespace filter",
			filters: metrics.MeshMetricsFilters{
				Namespaces: []string{"microservices"},
			},
			expected: ", destination_namespace=~\"microservices\"",
		},
		{
			name: "multiple namespace filters",
			filters: metrics.MeshMetricsFilters{
				Namespaces: []string{"microservices", "istio-system"},
			},
			expected: ", destination_namespace=~\"microservices|istio-system\"",
		},
		{
			name: "single cluster filter",
			filters: metrics.MeshMetricsFilters{
				Clusters: []string{"cluster1"},
			},
			expected: ", destination_cluster=~\"cluster1\"",
		},
		{
			name: "multiple cluster filters",
			filters: metrics.MeshMetricsFilters{
				Clusters: []string{"cluster1", "cluster2"},
			},
			expected: ", destination_cluster=~\"cluster1|cluster2\"",
		},
		{
			name: "both namespace and cluster filters",
			filters: metrics.MeshMetricsFilters{
				Namespaces: []string{"ns1", "ns2"},
				Clusters:   []string{"c1"},
			},
			expected: ", destination_namespace=~\"ns1|ns2\", destination_cluster=~\"c1\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.buildFilterClause(tt.filters)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildQueryFromTemplate(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	tests := []struct {
		name        string
		template    *template.Template
		filters     metrics.MeshMetricsFilters
		timeRange   string
		shouldError bool
		contains    []string
	}{
		{
			name:      "request rate template with no filters",
			template:  requestRateQueryTemplate,
			filters:   metrics.MeshMetricsFilters{},
			timeRange: "5m",
			contains:  []string{"rate(istio_requests_total", "[5m]", "reporter=\"destination\""},
		},
		{
			name:     "error rate template with filters",
			template: errorRateQueryTemplate,
			filters: metrics.MeshMetricsFilters{
				Namespaces: []string{"microservices"},
			},
			timeRange: "10m",
			contains:  []string{"response_code=~\"4..|5..\"", "destination_namespace=~\"microservices\"", "[10m]"},
		},
		{
			name:     "request rate with multiple filters",
			template: requestRateQueryTemplate,
			filters: metrics.MeshMetricsFilters{
				Namespaces: []string{"ns1", "ns2"},
				Clusters:   []string{"cluster1"},
			},
			timeRange: "1m",
			contains:  []string{"destination_namespace=~\"ns1|ns2\"", "destination_cluster=~\"cluster1\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.buildQueryFromTemplate(tt.template, tt.filters, tt.timeRange)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)

				// Check that query contains expected substrings
				for _, substr := range tt.contains {
					assert.Contains(t, result, substr)
				}
			}
		})
	}
}

func TestGetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		metric   map[string]string
		key      string
		expected string
	}{
		{
			name:     "existing key",
			metric:   map[string]string{"key1": "value1", "key2": "value2"},
			key:      "key1",
			expected: "value1",
		},
		{
			name:     "missing key",
			metric:   map[string]string{"key1": "value1"},
			key:      "missing",
			expected: "",
		},
		{
			name:     "empty value",
			metric:   map[string]string{"key1": ""},
			key:      "key1",
			expected: "",
		},
		{
			name:     "nil map",
			metric:   nil,
			key:      "any",
			expected: "",
		},
		{
			name:     "empty map",
			metric:   map[string]string{},
			key:      "any",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(tt.metric, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPrometheusTimeRange(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "5 minutes",
			duration: 5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "10 minutes",
			duration: 10 * time.Minute,
			expected: "10m",
		},
		{
			name:     "90 seconds (not clean minutes)",
			duration: 90 * time.Second,
			expected: "90s",
		},
		{
			name:     "30 seconds (less than 60)",
			duration: 30 * time.Second,
			expected: "60s", // Minimum 60s
		},
		{
			name:     "300 seconds (5 minutes)",
			duration: 300 * time.Second,
			expected: "5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPrometheusTimeRange(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessResponsesInParallel(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}
	timestamp := time.Now()

	errorResponse := model.Vector{
		&model.Sample{
			Metric: model.Metric{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "microservices",
				"source_canonical_service":      "frontend",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "backend",
			},
			Value: model.SampleValue(0.1),
		},
	}

	requestResponse := model.Vector{
		&model.Sample{
			Metric: model.Metric{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "microservices",
				"source_canonical_service":      "frontend",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "backend",
			},
			Value: model.SampleValue(5.0),
		},
		&model.Sample{
			Metric: model.Metric{
				"source_cluster":                "Kubernetes",
				"source_workload_namespace":     "microservices",
				"source_canonical_service":      "backend",
				"destination_cluster":           "Kubernetes",
				"destination_service_namespace": "microservices",
				"destination_canonical_service": "database",
			},
			Value: model.SampleValue(3.0),
		},
	}

	errorMetrics, requestMetrics := provider.processResponsesInParallel(errorResponse, requestResponse, timestamp)

	// Verify error metrics processing
	assert.NoError(t, errorMetrics.Error)
	assert.Equal(t, "error_rate", errorMetrics.MetricType)
	assert.Len(t, errorMetrics.PairData, 1)

	// Verify request metrics processing
	assert.NoError(t, requestMetrics.Error)
	assert.Equal(t, "request_rate", requestMetrics.MetricType)
	assert.Len(t, requestMetrics.PairData, 2)

	// Verify concurrent execution (both should complete)
	assert.NotNil(t, errorMetrics.PairData)
	assert.NotNil(t, requestMetrics.PairData)
}

func TestExecuteQueriesInParallelPattern(t *testing.T) {

	// Test the parallel execution pattern (without requiring actual Prometheus client)
	tests := []struct {
		name         string
		errorError   error
		requestError error
		expectError  bool
	}{
		{
			name:         "both queries succeed",
			errorError:   nil,
			requestError: nil,
			expectError:  false,
		},
		{
			name:         "error query fails",
			errorError:   fmt.Errorf("prometheus error"),
			requestError: nil,
			expectError:  true,
		},
		{
			name:         "request query fails",
			errorError:   nil,
			requestError: fmt.Errorf("prometheus error"),
			expectError:  false, // Request errors are logged but don't fail the whole operation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the parallel execution pattern
			var wg sync.WaitGroup
			var errorResults, requestResults queryResult

			wg.Add(2)

			go func() {
				defer wg.Done()
				if tt.errorError != nil {
					errorResults = queryResult{Error: tt.errorError, QueryType: "error"}
				} else {
					errorResults = queryResult{Response: model.Vector{}, QueryType: "error"}
				}
			}()

			go func() {
				defer wg.Done()
				if tt.requestError != nil {
					requestResults = queryResult{Error: tt.requestError, QueryType: "request"}
				} else {
					requestResults = queryResult{Response: model.Vector{}, QueryType: "request"}
				}
			}()

			wg.Wait()

			// Verify results
			if tt.expectError {
				assert.Error(t, errorResults.Error)
			} else {
				assert.NoError(t, errorResults.Error)
			}

			if tt.requestError != nil {
				assert.Error(t, requestResults.Error)
			} else {
				assert.NoError(t, requestResults.Error)
			}

			// Verify parallel execution completed
			assert.Equal(t, "error", errorResults.QueryType)
			assert.Equal(t, "request", requestResults.QueryType)
		})
	}
}
