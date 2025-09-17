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
	"testing"
	"text/template"
	"time"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetServiceConnections_NoClient(t *testing.T) {
	logger := logging.For("test")

	// Create provider with no client
	provider := &Provider{
		logger: logger,
		client: nil, // No client
	}

	filters := metrics.MeshMetricsFilters{}
	result, err := provider.getServiceConnectionsInternal(context.Background(), "frontend", "default", typesv1alpha1.ProxyMode_SIDECAR, filters)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetServiceConnections_DoubleCountingFix(t *testing.T) {
	logger := logging.For("test")

	// This test verifies that the double-counting bug is FIXED by using
	// getServiceConnectionsInternal with proper reporter field queries
	mockClient := &mockClient{
		responses: map[string]mockResponse{
			// NEW: Inbound query WITH reporter="destination" (FIXED queries)
			`sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="destination", destination_canonical_service="backend", destination_service_namespace="microservices"}[5m])
)`: {
				result: createMockVector(map[string]interface{}{
					"source_cluster":                "Kubernetes",
					"source_workload_namespace":     "microservices",
					"source_canonical_service":      "frontend",
					"destination_cluster":           "Kubernetes",
					"destination_service_namespace": "microservices",
					"destination_canonical_service": "backend",
				}, 15.0),
			},
			// NEW: Outbound query WITH reporter="source" (FIXED queries) - returns different connection
			`sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="source", source_canonical_service="backend", source_workload_namespace="microservices"}[5m])
)`: {
				result: createMockVector(map[string]interface{}{
					"source_cluster":                "Kubernetes",
					"source_workload_namespace":     "microservices",
					"source_canonical_service":      "backend",
					"destination_cluster":           "Kubernetes",
					"destination_service_namespace": "microservices",
					"destination_canonical_service": "database",
				}, 15.0),
			},
			// Error rate queries return empty (no errors)
			`sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="destination", destination_canonical_service="backend", destination_service_namespace="microservices", response_code=~"0|4..|5.."}[5m])
)`: {result: nil},
			`sum by (
  source_cluster, source_workload_namespace, source_canonical_service,
  destination_cluster, destination_service_namespace, destination_canonical_service
)(
  rate(istio_requests_total{reporter="source", source_canonical_service="backend", source_workload_namespace="microservices", response_code=~"0|4..|5.."}[5m])
)`: {result: nil},
		},
	}

	provider := &Provider{
		logger:      logger,
		client:      mockClient, // Now implements ClientInterface
		clusterName: "Kubernetes",
		info: metrics.ProviderInfo{
			Type:     metrics.ProviderTypePrometheus,
			Endpoint: "test-endpoint",
		},
	}

	// Execute the service connections query using the Provider's public method (which now calls the FIXED method)
	now := timestamppb.Now()
	fiveMinutesAgo := timestamppb.New(time.Now().Add(-5 * time.Minute))
	result, err := provider.GetServiceConnections(context.Background(), "backend", "microservices", typesv1alpha1.ProxyMode_SIDECAR, fiveMinutesAgo, now)

	require.NoError(t, err)
	require.NotNil(t, result)

	// FIXED: Now we should have 2 separate connections with no double-counting
	assert.Len(t, result.Pairs, 2, "Should have 2 unique connections")

	// Find the frontend -> backend connection
	var frontendToBackend *typesv1alpha1.ServicePairMetrics
	var backendToDatabase *typesv1alpha1.ServicePairMetrics
	for _, pair := range result.Pairs {
		if pair.SourceService == "frontend" && pair.DestinationService == "backend" {
			frontendToBackend = pair
		} else if pair.SourceService == "backend" && pair.DestinationService == "database" {
			backendToDatabase = pair
		}
	}

	require.NotNil(t, frontendToBackend, "Should have frontend -> backend connection")
	require.NotNil(t, backendToDatabase, "Should have backend -> database connection")

	// FIXED: Each connection should have its correct rate with no double-counting
	assert.Equal(t, 15.0, frontendToBackend.RequestRate, "Frontend -> backend should have 15 RPS")
	assert.Equal(t, 15.0, backendToDatabase.RequestRate, "Backend -> database should have 15 RPS")
}

func TestBuildFilterClause(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	tests := []struct {
		name           string
		filters        metrics.MeshMetricsFilters
		expectedClause string
	}{
		{
			name:           "no filters",
			filters:        metrics.MeshMetricsFilters{},
			expectedClause: "",
		},
		{
			name: "single namespace filter",
			filters: metrics.MeshMetricsFilters{
				Namespaces: []string{"default"},
			},
			expectedClause: `, source_workload_namespace=~""default""`,
		},
		{
			name: "multiple namespace filters",
			filters: metrics.MeshMetricsFilters{
				Namespaces: []string{"default", "production"},
			},
			expectedClause: `, source_workload_namespace=~""default"|"production""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clause := provider.buildFilterClause(tt.filters)
			assert.Equal(t, tt.expectedClause, clause)
		})
	}
}

func TestExecuteTemplate(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	data := serviceConnectionsQueryTemplateData{
		FilterClause:     `, namespace="default"`,
		TimeRange:        "5m",
		ServiceName:      "frontend",
		ServiceNamespace: "default",
	}

	result, err := provider.executeTemplate(inboundRequestRateQueryTemplate, data)

	require.NoError(t, err)
	assert.Contains(t, result, "frontend")
	assert.Contains(t, result, "default")
	assert.Contains(t, result, "5m")
	assert.Contains(t, result, `namespace="default"`)
}

func TestBuildServiceConnectionQuery(t *testing.T) {
	logger := logging.For("test")
	provider := &Provider{logger: logger}

	filters := metrics.MeshMetricsFilters{
		Namespaces: []string{"default"},
	}

	query, err := provider.buildServiceConnectionQuery(
		inboundRequestRateQueryTemplate,
		"frontend",
		"default",
		filters,
		"5m",
	)

	require.NoError(t, err)
	assert.Contains(t, query, "frontend")
	assert.Contains(t, query, "default")
	assert.Contains(t, query, "5m")
	assert.Contains(t, query, `source_workload_namespace=~""default""`)
}

func TestServiceConnectionsQueryTemplates(t *testing.T) {
	// Test that all query templates are valid and can be parsed
	templates := []*template.Template{
		inboundRequestRateQueryTemplate,
		outboundRequestRateQueryTemplate,
		inboundErrorRateQueryTemplate,
		outboundErrorRateQueryTemplate,
	}

	data := serviceConnectionsQueryTemplateData{
		FilterClause:     "",
		TimeRange:        "5m",
		ServiceName:      "test-service",
		ServiceNamespace: "test-namespace",
	}

	for i, tmpl := range templates {
		t.Run(fmt.Sprintf("template_%d", i), func(t *testing.T) {
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, data)
			require.NoError(t, err)

			result := buf.String()
			assert.Contains(t, result, "test-service")
			assert.Contains(t, result, "test-namespace")
			assert.Contains(t, result, "5m")
			assert.Contains(t, result, "istio_requests_total")
		})
	}
}

// Mock client for testing - implements ClientInterface
type mockClient struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	result model.Value
	err    error
}

func (m *mockClient) query(ctx context.Context, query string) (model.Value, error) {
	if resp, exists := m.responses[query]; exists {
		return resp.result, resp.err
	}
	return nil, fmt.Errorf("unexpected query: %s", query)
}

// GetServiceConnections is needed to satisfy ClientInterface but not used since we fixed the Provider
func (m *mockClient) GetServiceConnections(ctx context.Context, serviceName, namespace string, startTime, endTime time.Time) (*typesv1alpha1.ServiceGraphMetrics, error) {
	return nil, fmt.Errorf("GetServiceConnections not implemented in mock - Provider now uses getServiceConnectionsInternal")
}

// Helper to create mock Prometheus vector data
func createMockVector(labels map[string]interface{}, value float64) model.Vector {
	metric := model.Metric{}
	for k, v := range labels {
		metric[model.LabelName(k)] = model.LabelValue(fmt.Sprintf("%v", v))
	}

	return model.Vector{
		&model.Sample{
			Metric: metric,
			Value:  model.SampleValue(value),
		},
	}
}
