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

package metrics

import (
	"time"
)

// ProviderType represents the type of metrics provider
type ProviderType string

const (
	// ProviderTypePrometheus indicates a Prometheus metrics provider
	ProviderTypePrometheus ProviderType = "prometheus"
	// ProviderTypeNone indicates no metrics provider
	ProviderTypeNone ProviderType = "none"
)

// ProviderHealth represents the health status of a metrics provider
type ProviderHealth struct {
	// Status indicates the overall health status
	Status HealthStatus `json:"status"`
	// Message provides additional details about the health status
	Message string `json:"message"`
	// LastCheck is the timestamp of the last health check
	LastCheck time.Time `json:"last_check"`
	// Version is the version of the metrics provider (if available)
	Version string `json:"version,omitempty"`
}

// HealthStatus represents the health status values
type HealthStatus string

const (
	// HealthStatusHealthy indicates the provider is healthy and accessible
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusUnhealthy indicates the provider has issues
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusUnavailable indicates the provider is not accessible
	HealthStatusUnavailable HealthStatus = "unavailable"
	// HealthStatusUnknown indicates the health status is unknown
	HealthStatusUnknown HealthStatus = "unknown"
)

// ProviderInfo contains information about a metrics provider
type ProviderInfo struct {
	// Type is the type of metrics provider
	Type ProviderType `json:"type"`
	// Endpoint is the endpoint URL of the metrics provider
	Endpoint string `json:"endpoint"`
	// Health is the current health status of the provider
	Health ProviderHealth `json:"health"`
}

// ServiceMetrics contains metrics data for a service
type ServiceMetrics struct {
	// ServiceName is the name of the service
	ServiceName string `json:"service_name"`
	// Namespace is the namespace of the service
	Namespace string `json:"namespace"`
	// RequestRate is the request rate per second
	RequestRate float64 `json:"request_rate"`
	// ErrorRate is the error rate as a percentage (0-100)
	ErrorRate float64 `json:"error_rate"`
	// Timestamp is when these metrics were collected
	Timestamp time.Time `json:"timestamp"`
}

// MetricsQuery represents a query for metrics data
type MetricsQuery struct {
	// ServiceName is the name of the service to query metrics for
	ServiceName string
	// Namespace is the namespace of the service
	Namespace string
	// TimeRange is the time range for the query
	TimeRange TimeRange
}

// TimeRange represents a time range for metrics queries
type TimeRange struct {
	// Start is the start time for the query
	Start time.Time
	// End is the end time for the query
	End time.Time
}

// TimeSeries represents a time series of metric values
type TimeSeries struct {
	// MetricName is the name of the metric
	MetricName string `json:"metric_name"`
	// Values are the time-series data points
	Values []DataPoint `json:"values"`
}

// DataPoint represents a single data point in a time series
type DataPoint struct {
	// Timestamp is when this data point was recorded
	Timestamp time.Time `json:"timestamp"`
	// Value is the metric value at this timestamp
	Value float64 `json:"value"`
}
