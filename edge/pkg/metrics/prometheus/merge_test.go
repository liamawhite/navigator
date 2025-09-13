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
	"testing"
	"time"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessErrorRateResponse(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}
	timestamp := time.Now()

	tests := []struct {
		name          string
		response      model.Value
		expectedPairs int
		expectedType  string
		expectedError bool
		validatePairs func(t *testing.T, pairs map[string]*metrics.ServicePairMetrics)
	}{
		{
			name:          "nil response",
			response:      nil,
			expectedPairs: 0,
			expectedType:  "error_rate",
			expectedError: false,
		},
		{
			name:          "wrong response type",
			response:      model.Matrix{},
			expectedPairs: 0,
			expectedType:  "error_rate",
			expectedError: true,
		},
		{
			name: "valid error rate data",
			response: model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "cluster1",
						"source_workload_namespace":     "default",
						"source_canonical_service":      "frontend",
						"destination_cluster":           "cluster1",
						"destination_service_namespace": "default",
						"destination_canonical_service": "backend",
					},
					Value: model.SampleValue(0.05),
				},
			},
			expectedPairs: 1,
			expectedType:  "error_rate",
			expectedError: false,
			validatePairs: func(t *testing.T, pairs map[string]*metrics.ServicePairMetrics) {
				require.Len(t, pairs, 1)
				for _, pair := range pairs {
					assert.Equal(t, "cluster1", pair.SourceCluster)
					assert.Equal(t, "default", pair.SourceNamespace)
					assert.Equal(t, "frontend", pair.SourceService)
					assert.Equal(t, "cluster1", pair.DestinationCluster)
					assert.Equal(t, "default", pair.DestinationNamespace)
					assert.Equal(t, "backend", pair.DestinationService)
					assert.Equal(t, 0.05, pair.ErrorRate)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.processErrorRateResponse(tt.response, timestamp)

			assert.Equal(t, tt.expectedType, result.MetricType)
			assert.Len(t, result.PairData, tt.expectedPairs)

			if tt.expectedError {
				assert.Error(t, result.Error)
			} else {
				assert.NoError(t, result.Error)
			}

			if tt.validatePairs != nil {
				tt.validatePairs(t, result.PairData)
			}
		})
	}
}

func TestProcessRequestRateResponse(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}
	timestamp := time.Now()

	tests := []struct {
		name          string
		response      model.Value
		expectedPairs int
		expectedType  string
		expectedError bool
		validatePairs func(t *testing.T, pairs map[string]*metrics.ServicePairMetrics)
	}{
		{
			name:          "nil response",
			response:      nil,
			expectedPairs: 0,
			expectedType:  "request_rate",
			expectedError: false,
		},
		{
			name:          "wrong response type",
			response:      model.Matrix{},
			expectedPairs: 0,
			expectedType:  "request_rate",
			expectedError: true,
		},
		{
			name: "valid request rate data",
			response: model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"source_cluster":                "cluster1",
						"source_workload_namespace":     "default",
						"source_canonical_service":      "frontend",
						"destination_cluster":           "cluster1",
						"destination_service_namespace": "default",
						"destination_canonical_service": "backend",
					},
					Value: model.SampleValue(100.0),
				},
			},
			expectedPairs: 1,
			expectedType:  "request_rate",
			expectedError: false,
			validatePairs: func(t *testing.T, pairs map[string]*metrics.ServicePairMetrics) {
				require.Len(t, pairs, 1)
				for _, pair := range pairs {
					assert.Equal(t, "cluster1", pair.SourceCluster)
					assert.Equal(t, "default", pair.SourceNamespace)
					assert.Equal(t, "frontend", pair.SourceService)
					assert.Equal(t, "cluster1", pair.DestinationCluster)
					assert.Equal(t, "default", pair.DestinationNamespace)
					assert.Equal(t, "backend", pair.DestinationService)
					assert.Equal(t, 100.0, pair.RequestRate)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.processRequestRateResponse(tt.response, timestamp)

			assert.Equal(t, tt.expectedType, result.MetricType)
			assert.Len(t, result.PairData, tt.expectedPairs)

			if tt.expectedError {
				assert.Error(t, result.Error)
			} else {
				assert.NoError(t, result.Error)
			}

			if tt.validatePairs != nil {
				tt.validatePairs(t, result.PairData)
			}
		})
	}
}

func TestMergePairMaps(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	requestPairs := map[string]*metrics.ServicePairMetrics{
		"cluster1:default:frontend->cluster1:default:backend": {
			SourceCluster:        "cluster1",
			SourceNamespace:      "default",
			SourceService:        "frontend",
			DestinationCluster:   "cluster1",
			DestinationNamespace: "default",
			DestinationService:   "backend",
			RequestRate:          100.0,
		},
	}

	errorPairs := map[string]*metrics.ServicePairMetrics{
		"cluster1:default:frontend->cluster1:default:backend": {
			SourceCluster:        "cluster1",
			SourceNamespace:      "default",
			SourceService:        "frontend",
			DestinationCluster:   "cluster1",
			DestinationNamespace: "default",
			DestinationService:   "backend",
			ErrorRate:            0.05,
		},
		"cluster1:default:backend->cluster1:default:database": {
			SourceCluster:        "cluster1",
			SourceNamespace:      "default",
			SourceService:        "backend",
			DestinationCluster:   "cluster1",
			DestinationNamespace: "default",
			DestinationService:   "database",
			ErrorRate:            0.02,
		},
	}

	merged := provider.mergePairMaps(requestPairs, errorPairs)

	// Should have 2 pairs total
	assert.Len(t, merged, 2)

	// Check merged pair (has both request and error rate)
	frontendBackend := merged["cluster1:default:frontend->cluster1:default:backend"]
	require.NotNil(t, frontendBackend)
	assert.Equal(t, "frontend", frontendBackend.SourceService)
	assert.Equal(t, "backend", frontendBackend.DestinationService)
	assert.Equal(t, 100.0, frontendBackend.RequestRate)
	assert.Equal(t, 0.05, frontendBackend.ErrorRate)

	// Check error-only pair
	backendDatabase := merged["cluster1:default:backend->cluster1:default:database"]
	require.NotNil(t, backendDatabase)
	assert.Equal(t, "backend", backendDatabase.SourceService)
	assert.Equal(t, "database", backendDatabase.DestinationService)
	assert.Equal(t, 0.0, backendDatabase.RequestRate) // Should default to 0
	assert.Equal(t, 0.02, backendDatabase.ErrorRate)
}

func TestCreatePairKey(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	tests := []struct {
		name        string
		metric      model.Metric
		expectedKey string
	}{
		{
			name: "complete metric",
			metric: model.Metric{
				"source_cluster":                "cluster1",
				"source_workload_namespace":     "default",
				"source_canonical_service":      "frontend",
				"destination_cluster":           "cluster1",
				"destination_service_namespace": "default",
				"destination_canonical_service": "backend",
			},
			expectedKey: "cluster1:default:frontend->cluster1:default:backend",
		},
		{
			name: "missing source service",
			metric: model.Metric{
				"source_cluster":                "cluster1",
				"source_workload_namespace":     "default",
				"destination_cluster":           "cluster1",
				"destination_service_namespace": "default",
				"destination_canonical_service": "backend",
			},
			expectedKey: "", // Should return empty string
		},
		{
			name: "missing destination service",
			metric: model.Metric{
				"source_cluster":            "cluster1",
				"source_workload_namespace": "default",
				"source_canonical_service":  "frontend",
				"destination_cluster":       "cluster1",
			},
			expectedKey: "", // Should return empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := provider.createPairKey(tt.metric)
			assert.Equal(t, tt.expectedKey, key)
		})
	}
}

func TestGetStringValue(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	metric := model.Metric{
		"source_service": "frontend",
		"dest_service":   "backend",
	}

	tests := []struct {
		name          string
		key           string
		expectedValue string
	}{
		{
			name:          "existing key",
			key:           "source_service",
			expectedValue: "frontend",
		},
		{
			name:          "another existing key",
			key:           "dest_service",
			expectedValue: "backend",
		},
		{
			name:          "non-existing key",
			key:           "missing_key",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := provider.getStringValue(metric, tt.key)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}
