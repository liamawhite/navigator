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

// MeshMetricsFilters represents filters for service mesh metrics queries
type MeshMetricsFilters struct {
	Namespaces []string
	Clusters   []string
}

// MeshMetricsQuery represents a query for service mesh metrics
type MeshMetricsQuery struct {
	Filters   MeshMetricsFilters
	StartTime time.Time
	EndTime   time.Time
}

// ServicePairMetrics represents metrics between a source and destination service
type ServicePairMetrics struct {
	SourceCluster        string    `json:"source_cluster"`
	SourceNamespace      string    `json:"source_namespace"`
	SourceService        string    `json:"source_service"`
	DestinationCluster   string    `json:"destination_cluster"`
	DestinationNamespace string    `json:"destination_namespace"`
	DestinationService   string    `json:"destination_service"`
	ErrorRate            float64   `json:"error_rate"`   // requests per second
	RequestRate          float64   `json:"request_rate"` // requests per second
	Timestamp            time.Time `json:"timestamp"`
}

// ServiceGraphMetrics contains service graph metrics for a cluster
type ServiceGraphMetrics struct {
	Pairs     []ServicePairMetrics `json:"pairs"`
	ClusterID string               `json:"cluster_id"`
	Timestamp time.Time            `json:"timestamp"`
}
