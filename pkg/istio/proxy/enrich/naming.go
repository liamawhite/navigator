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

package enrich

import (
	"strconv"
	"strings"

	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// Common Istio static cluster names
var istioStaticClusters = []string{
	"prometheus_stats",
	"agent",
	"sds-grpc",
	"xds-grpc",
	"zipkin",
	"jaeger",
	"envoy_accesslog_service",
}

// parseDirection parses direction string into ClusterDirection enum
func parseDirection(directionStr string) v1alpha1.ClusterDirection {
	switch strings.ToLower(directionStr) {
	case "inbound":
		return v1alpha1.ClusterDirection_INBOUND
	case "outbound":
		return v1alpha1.ClusterDirection_OUTBOUND
	default:
		return v1alpha1.ClusterDirection_UNSPECIFIED
	}
}

// parseClusterComponents parses Istio cluster name into components
// Returns: direction, port, subset, serviceFqdn
func parseClusterComponents(clusterName string) (v1alpha1.ClusterDirection, uint32, string, string) {
	parts := strings.Split(clusterName, "|")

	// Default values
	direction := v1alpha1.ClusterDirection_UNSPECIFIED
	var port uint32 = 0
	subset := ""
	serviceFqdn := ""

	if len(parts) >= 1 {
		direction = parseDirection(parts[0])
	}

	if len(parts) >= 2 {
		if portValue, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
			port = uint32(portValue)
		}
	}

	if len(parts) >= 3 {
		subset = parts[2]
	}

	if len(parts) >= 4 {
		serviceFqdn = parts[3]
	}

	return direction, port, subset, serviceFqdn
}

// isIstioClusterPattern checks if cluster name follows Istio patterns
func isIstioClusterPattern(clusterName string) bool {
	if strings.HasPrefix(clusterName, "outbound|") || strings.HasPrefix(clusterName, "inbound|") {
		parts := strings.Split(clusterName, "|")
		return len(parts) == 4
	}
	return false
}

// parseFQDN extracts service name and namespace from Kubernetes service FQDN
func parseFQDN(serviceFqdn string) (serviceName, namespace string) {
	if strings.Contains(serviceFqdn, ".svc.cluster.local") {
		parts := strings.Split(serviceFqdn, ".")
		if len(parts) >= 1 {
			serviceName = parts[0]
		}
		if len(parts) >= 2 {
			namespace = parts[1]
		}
		return serviceName, namespace
	}
	// For external services
	return serviceFqdn, ""
}

// ParseClusterName parses Istio cluster names in the format: direction|port|subset|servicefqdn
// This function updates the provided EndpointSummary with parsed information
func ParseClusterName(clusterName string, summary *v1alpha1.EndpointSummary) {
	// Use shared parsing logic
	direction, port, subset, serviceFqdn := parseClusterComponents(clusterName)

	// Only update if we have valid Istio format (4 parts)
	parts := strings.Split(clusterName, "|")
	if len(parts) == 4 {
		summary.Direction = direction
		summary.Port = port
		summary.Subset = subset
		summary.ServiceFqdn = serviceFqdn
	} else {
		// Not in expected Istio format, set defaults
		summary.Direction = v1alpha1.ClusterDirection_UNSPECIFIED
		summary.Port = 0
		summary.Subset = ""
		summary.ServiceFqdn = ""
	}
}

// ParseClusterNameComponents parses Istio cluster names and returns individual components
// Examples:
//   - "outbound|8080||backend.demo.svc.cluster.local"
//   - "inbound|8080||"
//   - "outbound|443|v1|api.example.com"
func ParseClusterNameComponents(clusterName string) (direction v1alpha1.ClusterDirection, port uint32, subset string, serviceFqdn string) {
	return parseClusterComponents(clusterName)
}

// InferClusterType infers the cluster type from Istio cluster name patterns
// This follows similar logic to how istioctl and Envoy classify clusters
func InferClusterType(clusterName string) v1alpha1.ClusterType {
	if clusterName == "" {
		return v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE
	}

	// Check for Istio sidecar patterns (most common)
	if isIstioClusterPattern(clusterName) {
		// These are typically EDS clusters for service discovery
		_, _, _, serviceFqdn := parseClusterComponents(clusterName)
		// If it looks like a Kubernetes service FQDN, it's EDS
		if strings.Contains(serviceFqdn, ".svc.cluster.local") {
			return v1alpha1.ClusterType_CLUSTER_EDS
		}
		// External services might be DNS-based
		if strings.Contains(serviceFqdn, ".") && !strings.Contains(serviceFqdn, ".svc.cluster.local") {
			return v1alpha1.ClusterType_CLUSTER_STRICT_DNS
		}
		return v1alpha1.ClusterType_CLUSTER_EDS // Default for outbound/inbound patterns
	}

	// Check for static clusters
	for _, staticCluster := range istioStaticClusters {
		if clusterName == staticCluster {
			return v1alpha1.ClusterType_CLUSTER_STATIC
		}
	}

	// If cluster name contains an IP address, it's likely static
	if strings.Contains(clusterName, ".") && !strings.Contains(clusterName, "svc.cluster.local") {
		// Simple heuristic: if it looks like an external domain, it's DNS-based
		if strings.Count(clusterName, ".") >= 2 {
			return v1alpha1.ClusterType_CLUSTER_STRICT_DNS
		}
		return v1alpha1.ClusterType_CLUSTER_STATIC
	}

	// Default fallback
	return v1alpha1.ClusterType_CLUSTER_EDS
}

// InferClusterTypeFromName is a simplified version for basic cluster type inference
func InferClusterTypeFromName(clusterName string) v1alpha1.ClusterType {
	clusterName = strings.ToLower(clusterName)

	if strings.Contains(clusterName, "outbound") || strings.Contains(clusterName, "inbound") {
		return v1alpha1.ClusterType_CLUSTER_EDS
	}

	if strings.Contains(clusterName, "static") {
		return v1alpha1.ClusterType_CLUSTER_STATIC
	}

	return v1alpha1.ClusterType_UNKNOWN_CLUSTER_TYPE
}

// IsIstioCluster checks if a cluster name follows Istio naming patterns
func IsIstioCluster(clusterName string) bool {
	// Check for Istio cluster patterns
	if isIstioClusterPattern(clusterName) {
		return true
	}

	// Check for common Istio static clusters
	for _, staticCluster := range istioStaticClusters {
		if clusterName == staticCluster {
			return true
		}
	}

	return false
}

// ExtractServiceName extracts the Kubernetes service name from a service FQDN
func ExtractServiceName(serviceFqdn string) string {
	serviceName, _ := parseFQDN(serviceFqdn)
	return serviceName
}

// ExtractNamespace extracts the Kubernetes namespace from a service FQDN
func ExtractNamespace(serviceFqdn string) string {
	_, namespace := parseFQDN(serviceFqdn)
	return namespace
}
