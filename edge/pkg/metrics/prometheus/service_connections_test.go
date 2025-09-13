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

	"github.com/liamawhite/navigator/edge/pkg/metrics"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetServiceConnections_UnhealthyProvider(t *testing.T) {
	logger := logging.For("test")

	// Create provider with unhealthy status
	provider := &Provider{
		logger: logger,
		client: nil, // No client to make it unhealthy
	}

	filters := metrics.MeshMetricsFilters{}
	result, err := provider.getServiceConnectionsInternal(context.Background(), "frontend", "default", filters)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "prometheus provider is not healthy")
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
