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

package configdump

import (
	"fmt"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// parseClustersFromAny extracts cluster configurations from protobuf Any
func (p *Parser) parseClustersFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	clusterDump := &admin.ClustersConfigDump{}
	if err := configAny.UnmarshalTo(clusterDump); err != nil {
		return fmt.Errorf("failed to unmarshal clusters config dump: %w", err)
	}

	// Extract dynamic clusters
	for _, c := range clusterDump.DynamicActiveClusters {
		if c.Cluster != nil {
			var cluster clusterv3.Cluster
			if err := c.Cluster.UnmarshalTo(&cluster); err == nil {
				parsed.Clusters = append(parsed.Clusters, &cluster)
			}
		}
	}

	// Extract static clusters
	for _, c := range clusterDump.StaticClusters {
		if c.Cluster != nil {
			var cluster clusterv3.Cluster
			if err := c.Cluster.UnmarshalTo(&cluster); err == nil {
				parsed.Clusters = append(parsed.Clusters, &cluster)
			}
		}
	}

	return nil
}

// summarizeCluster converts a Cluster config to a ClusterSummary
func (p *Parser) summarizeCluster(cluster *clusterv3.Cluster, parsed *ParsedConfig) *v1alpha1.ClusterSummary {
	if cluster == nil {
		return nil
	}

	summary := &v1alpha1.ClusterSummary{
		Name:                cluster.Name,
		Type:                "UNKNOWN", // Will be determined from cluster type enum
		LoadBalancingPolicy: cluster.LbPolicy.String(),
		AltStatName:         cluster.AltStatName,
	}

	// Parse cluster name components (format: direction|port|subset|servicefqdn)
	p.parseClusterName(cluster.Name, summary)

	// Set cluster type based on the enum value
	switch cluster.GetType() {
	case clusterv3.Cluster_STATIC:
		summary.Type = "STATIC"
	case clusterv3.Cluster_STRICT_DNS:
		summary.Type = "STRICT_DNS"
	case clusterv3.Cluster_LOGICAL_DNS:
		summary.Type = "LOGICAL_DNS"
	case clusterv3.Cluster_EDS:
		summary.Type = "EDS"
	case clusterv3.Cluster_ORIGINAL_DST:
		summary.Type = "ORIGINAL_DST"
	}

	// Extract connect timeout
	if cluster.ConnectTimeout != nil {
		summary.ConnectTimeout = cluster.ConnectTimeout.String()
	}

	// Load assignment details are processed separately in endpoints.go
	// This keeps cluster and endpoint concerns properly separated

	// Store raw config for debugging
	if cluster != nil {
		summary.RawConfig = cluster.String()
	}

	// Use the raw JSON config that was extracted directly from the original config dump
	if rawJSON, exists := parsed.RawClusters[cluster.Name]; exists {
		summary.RawConfig = rawJSON
	}

	return summary
}

// parseClusterName provides basic cluster name parsing for generic Envoy deployments
// This function only extracts basic information without service mesh assumptions
func (p *Parser) parseClusterName(clusterName string, summary *v1alpha1.ClusterSummary) {
	// Initialize with default values - no assumptions about service mesh patterns
	summary.Direction = v1alpha1.ClusterDirection_UNSPECIFIED
	summary.Port = 0
	summary.Subset = ""
	summary.ServiceFqdn = clusterName // Use cluster name as default FQDN

	// For generic Envoy deployments, we don't parse specific naming patterns
	// Service mesh specific parsing should be done by dedicated enrichment layers
}
