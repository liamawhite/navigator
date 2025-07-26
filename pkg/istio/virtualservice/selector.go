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

package virtualservice

import (
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// isVisibleToNamespace determines if a virtual service is visible to a specific namespace
// based on its exportTo field following Istio's visibility rules:
// - Empty exportTo defaults to ["*"] (visible to all namespaces)
// - "*" means visible to all namespaces
// - "." means visible only to the same namespace as the virtual service
// - Specific namespace names mean visible only to those namespaces
func isVisibleToNamespace(vs *typesv1alpha1.VirtualService, workloadNamespace string) bool {
	if vs == nil {
		return false
	}

	// Empty exportTo defaults to ["*"] (visible to all namespaces)
	if len(vs.ExportTo) == 0 {
		return true
	}

	for _, export := range vs.ExportTo {
		if export == "*" {
			return true // visible to all namespaces
		}
		if export == "." && vs.Namespace == workloadNamespace {
			return true // visible to same namespace
		}
		if export == workloadNamespace {
			return true // visible to specific namespace
		}
	}

	return false
}

// isGatewayWorkload determines if a workload is an Istio gateway based on its proxy type
func isGatewayWorkload(instance *backendv1alpha1.ServiceInstance) bool {
	if instance == nil {
		return false
	}

	return instance.ProxyType == backendv1alpha1.ProxyType_GATEWAY
}

// appliesToWorkloadTraffic determines if a virtual service applies to the specific workload
// based on its gateways field and the workload type:
// - For sidecar workloads: check if "mesh" is in gateways
// - For gateway workloads: check if the gateway name matches any gateways in the virtual service
func appliesToWorkloadTraffic(vs *typesv1alpha1.VirtualService, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
	if vs == nil {
		return false
	}

	// Empty gateways defaults to ["mesh"] (applies to sidecar traffic)
	gateways := vs.Gateways
	if len(gateways) == 0 {
		gateways = []string{"mesh"}
	}

	// Check if this is a gateway workload
	if isGatewayWorkload(instance) {
		// For gateway workloads, check if any of the virtual service's gateways
		// could potentially match this gateway workload
		return appliesToGatewayWorkload(gateways, instance, workloadNamespace)
	}

	// For regular sidecar workloads, check if "mesh" is in the gateways list
	for _, gateway := range gateways {
		if gateway == "mesh" {
			return true
		}
	}

	return false
}

// appliesToGatewayWorkload determines if a virtual service applies to a specific gateway workload
// by checking if any of the gateways in the virtual service could match this gateway
func appliesToGatewayWorkload(gateways []string, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
	if instance == nil || instance.Labels == nil {
		return false
	}

	// Get potential gateway names from the workload labels
	gatewayNames := getGatewayNamesFromWorkload(instance, workloadNamespace)
	if len(gatewayNames) == 0 {
		return false
	}

	// Check if any of the virtual service's gateways match this workload's gateway names
	for _, vsGateway := range gateways {
		if vsGateway == "mesh" {
			continue // gateway workloads don't handle mesh traffic
		}

		for _, workloadGateway := range gatewayNames {
			if vsGateway == workloadGateway {
				return true
			}
		}
	}

	return false
}

// getGatewayNamesFromWorkload extracts potential gateway names that this workload might serve
// based on common Istio gateway naming patterns
func getGatewayNamesFromWorkload(instance *backendv1alpha1.ServiceInstance, workloadNamespace string) []string {
	if instance == nil || instance.Labels == nil {
		return nil
	}

	labels := instance.Labels
	var gatewayNames []string

	// Check for explicit gateway name label
	if gatewayName := labels["istio.io/gateway-name"]; gatewayName != "" {
		// Add both simple name and namespaced name
		gatewayNames = append(gatewayNames, gatewayName)
		gatewayNames = append(gatewayNames, workloadNamespace+"/"+gatewayName)
	}

	// Check for standard gateway patterns based on app label
	if app := labels["app"]; app != "" {
		switch app {
		case "istio-ingressgateway":
			gatewayNames = append(gatewayNames, "istio-ingressgateway")
			gatewayNames = append(gatewayNames, workloadNamespace+"/istio-ingressgateway")
		case "istio-egressgateway":
			gatewayNames = append(gatewayNames, "istio-egressgateway")
			gatewayNames = append(gatewayNames, workloadNamespace+"/istio-egressgateway")
		}
	}

	// Check for istio label
	if istio := labels["istio"]; istio == "ingressgateway" {
		gatewayNames = append(gatewayNames, "istio-ingressgateway")
		gatewayNames = append(gatewayNames, workloadNamespace+"/istio-ingressgateway")
	}

	return gatewayNames
}

// MatchesWorkload determines if a virtual service applies to a specific workload instance.
// It implements Istio's virtual service visibility and applicability logic by checking:
// 1. Namespace visibility (exportTo field)
// 2. Traffic context applicability (gateways field - considers both mesh and gateway traffic)
func matchesWorkload(vs *typesv1alpha1.VirtualService, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
	if vs == nil || instance == nil {
		return false
	}

	// Stage 1: Check namespace visibility
	if !isVisibleToNamespace(vs, workloadNamespace) {
		return false
	}

	// Stage 2: Check if applies to this workload's traffic context
	// This handles both sidecar (mesh) and gateway workloads
	if !appliesToWorkloadTraffic(vs, instance, workloadNamespace) {
		return false
	}

	return true
}

// FilterVirtualServicesForWorkload returns all virtual services that apply to a specific workload instance.
// This filters virtual services based on namespace visibility and traffic context
// to show only those virtual services that are relevant to the workload's traffic patterns.
func FilterVirtualServicesForWorkload(virtualServices []*typesv1alpha1.VirtualService, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) []*typesv1alpha1.VirtualService {
	var matchingVirtualServices []*typesv1alpha1.VirtualService

	for _, vs := range virtualServices {
		if matchesWorkload(vs, instance, workloadNamespace) {
			matchingVirtualServices = append(matchingVirtualServices, vs)
		}
	}

	return matchingVirtualServices
}
