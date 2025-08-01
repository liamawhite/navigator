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

package filters

import (
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// virtualServiceAppliesToWorkloadTraffic determines if a virtual service applies to the specific workload
// based on its gateways field and the workload type:
// - For sidecar workloads: check if "mesh" is in gateways
// - For gateway workloads: check if the gateway name matches any gateways in the virtual service
func virtualServiceAppliesToWorkloadTraffic(vs *typesv1alpha1.VirtualService, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
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
		return virtualServiceAppliesToGatewayWorkload(gateways, instance, workloadNamespace)
	}

	// For regular sidecar workloads, check if "mesh" is in the gateways list
	for _, gateway := range gateways {
		if gateway == "mesh" {
			return true
		}
	}

	return false
}

// virtualServiceAppliesToGatewayWorkload determines if a virtual service applies to a specific gateway workload
// by checking if any of the gateways in the virtual service could match this gateway
func virtualServiceAppliesToGatewayWorkload(gateways []string, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
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

// virtualServiceMatchesWorkload determines if a virtual service applies to a specific workload instance.
// It implements Istio's virtual service visibility and applicability logic by checking:
// 1. Namespace visibility (exportTo field)
// 2. Traffic context applicability (gateways field - considers both mesh and gateway traffic)
func virtualServiceMatchesWorkload(vs *typesv1alpha1.VirtualService, instance *backendv1alpha1.ServiceInstance, workloadNamespace string) bool {
	if vs == nil || instance == nil {
		return false
	}

	// Stage 1: Check namespace visibility
	if !isVisibleToNamespace(virtualServiceExporter(vs), workloadNamespace) {
		return false
	}

	// Stage 2: Check if applies to this workload's traffic context
	// This handles both sidecar (mesh) and gateway workloads
	if !virtualServiceAppliesToWorkloadTraffic(vs, instance, workloadNamespace) {
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
		if virtualServiceMatchesWorkload(vs, instance, workloadNamespace) {
			matchingVirtualServices = append(matchingVirtualServices, vs)
		}
	}

	return matchingVirtualServices
}
