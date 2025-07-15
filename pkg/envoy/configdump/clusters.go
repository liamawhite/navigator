package configdump

import (
	"fmt"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// parseClustersFromAny extracts cluster configurations from protobuf Any
func (p *Parser) parseClustersFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	clusterDump := &admin.ClustersConfigDump{}
	if err := configAny.UnmarshalTo(clusterDump); err != nil {
		return fmt.Errorf("failed to unmarshal clusters config dump: %w", err)
	}

	// Extract dynamic clusters (like istioctl)
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
func (p *Parser) summarizeCluster(cluster *clusterv3.Cluster) *v1alpha1.ClusterSummary {
	if cluster == nil {
		return nil
	}

	summary := &v1alpha1.ClusterSummary{
		Name:                cluster.Name,
		Type:                "UNKNOWN", // Will be determined from cluster type enum
		LoadBalancingPolicy: cluster.LbPolicy.String(),
		AltStatName:         cluster.AltStatName,
	}

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

	// Extract load assignment
	if cluster.LoadAssignment != nil {
		summary.LoadAssignment = &v1alpha1.EndpointConfigInfo{
			ClusterName: cluster.LoadAssignment.ClusterName,
		}

		// Extract endpoints
		for _, endpoint := range cluster.LoadAssignment.Endpoints {
			localityInfo := &v1alpha1.LocalityLbEndpointsInfo{
				LoadBalancingWeight: endpoint.LoadBalancingWeight.GetValue(),
				Priority:            endpoint.Priority,
				Proximity:           endpoint.Proximity.GetValue(),
			}

			if endpoint.Locality != nil {
				localityInfo.Locality = &v1alpha1.LocalityInfo{
					Region: endpoint.Locality.Region,
					Zone:   endpoint.Locality.Zone,
				}
			}

			// Extract lb endpoints
			for _, lbEndpoint := range endpoint.LbEndpoints {
				lbInfo := &v1alpha1.LbEndpointInfo{
					HealthStatus:        lbEndpoint.HealthStatus.String(),
					LoadBalancingWeight: lbEndpoint.LoadBalancingWeight.GetValue(),
				}

				if lbEndpoint.GetEndpoint() != nil && lbEndpoint.GetEndpoint().Address != nil {
					if sockAddr := lbEndpoint.GetEndpoint().Address.GetSocketAddress(); sockAddr != nil {
						lbInfo.Endpoint = &v1alpha1.EndpointDetailsInfo{
							Address: sockAddr.Address,
							Port:    sockAddr.GetPortValue(),
						}
					}
				}

				localityInfo.LbEndpoints = append(localityInfo.LbEndpoints, lbInfo)
			}

			summary.LoadAssignment.Endpoints = append(summary.LoadAssignment.Endpoints, localityInfo)
		}
	}

	// Extract health checks (simplified)
	for _, hc := range cluster.HealthChecks {
		hcInfo := &v1alpha1.HealthCheckInfo{
			UnhealthyThreshold: hc.UnhealthyThreshold.GetValue(),
			HealthyThreshold:   hc.HealthyThreshold.GetValue(),
		}

		if hc.Timeout != nil {
			hcInfo.Timeout = hc.Timeout.String()
		}
		if hc.Interval != nil {
			hcInfo.Interval = hc.Interval.String()
		}

		summary.HealthChecks = append(summary.HealthChecks, hcInfo)
	}

	// Extract circuit breakers (simplified)
	if cluster.CircuitBreakers != nil {
		summary.CircuitBreakers = &v1alpha1.CircuitBreakersInfo{}
		for _, threshold := range cluster.CircuitBreakers.Thresholds {
			thInfo := &v1alpha1.ThresholdInfo{
				Priority:           threshold.Priority.String(),
				MaxConnections:     threshold.MaxConnections.GetValue(),
				MaxPendingRequests: threshold.MaxPendingRequests.GetValue(),
				MaxRequests:        threshold.MaxRequests.GetValue(),
				MaxRetries:         threshold.MaxRetries.GetValue(),
				MaxConnectionPools: threshold.MaxConnectionPools.GetValue(),
			}
			summary.CircuitBreakers.Thresholds = append(summary.CircuitBreakers.Thresholds, thInfo)
		}
	}

	// Extract EDS config (simplified)
	if cluster.EdsClusterConfig != nil {
		summary.EdsClusterConfig = &v1alpha1.EdsClusterConfigInfo{
			ServiceName: cluster.EdsClusterConfig.ServiceName,
		}
	}

	return summary
}
