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

// Package filters provides unified Istio resource filtering functionality.
// This package consolidates all Istio resource type filtering logic that was
// previously spread across individual packages, reducing code duplication
// while maintaining identical behavior.
package filters

import (
	"k8s.io/apimachinery/pkg/labels"

	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// ExporterResource represents an Istio resource that supports exportTo field
type ExporterResource interface {
	GetNamespace() string
	GetExportTo() []string
}

// isVisibleToNamespace determines if an Istio resource is visible to a specific namespace
// based on its exportTo field following Istio's visibility rules:
// - Empty exportTo defaults to ["*"] (visible to all namespaces)
// - "*" means visible to all namespaces
// - "." means visible only to the same namespace as the resource
// - Specific namespace names mean visible only to those namespaces
func isVisibleToNamespace(resource ExporterResource, workloadNamespace string) bool {
	if resource == nil {
		return false
	}

	exportTo := resource.GetExportTo()
	
	// Empty exportTo defaults to ["*"] (visible to all namespaces)
	if len(exportTo) == 0 {
		return true
	}

	for _, export := range exportTo {
		if export == "*" {
			return true // visible to all namespaces
		}
		if export == "." && resource.GetNamespace() == workloadNamespace {
			return true // visible to same namespace
		}
		if export == workloadNamespace {
			return true // visible to specific namespace
		}
	}

	return false
}

// matchesLabelSelector checks if a Kubernetes label selector matches workload labels
func matchesLabelSelector(selectorLabels map[string]string, workloadLabels map[string]string) bool {
	if len(selectorLabels) == 0 {
		return true // empty selector matches all
	}

	if workloadLabels == nil {
		return false
	}

	// Convert to Kubernetes label selector for consistent matching
	selector := labels.Set(selectorLabels).AsSelector()
	workloadLabelSet := labels.Set(workloadLabels)
	
	return selector.Matches(workloadLabelSet)
}

// isGatewayWorkload determines if a workload is an Istio gateway based on its proxy type
func isGatewayWorkload(instance *backendv1alpha1.ServiceInstance) bool {
	if instance == nil {
		return false
	}
	return instance.ProxyType == backendv1alpha1.ProxyType_GATEWAY
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

// exportToResource provides common implementation for resources with exportTo field
type exportToResource struct {
	namespace string
	exportTo  []string
}

func (r *exportToResource) GetNamespace() string {
	return r.namespace
}

func (r *exportToResource) GetExportTo() []string {
	return r.exportTo
}

// newExportToResource creates a wrapper for resources with exportTo field
func newExportToResource(namespace string, exportTo []string) ExporterResource {
	return &exportToResource{
		namespace: namespace,
		exportTo:  exportTo,
	}
}

// Helper functions for specific resource types to create ExporterResource wrappers

func destinationRuleExporter(dr *typesv1alpha1.DestinationRule) ExporterResource {
	if dr == nil {
		return nil
	}
	return newExportToResource(dr.Namespace, dr.ExportTo)
}

func virtualServiceExporter(vs *typesv1alpha1.VirtualService) ExporterResource {
	if vs == nil {
		return nil
	}
	return newExportToResource(vs.Namespace, vs.ExportTo)
}

func serviceEntryExporter(se *typesv1alpha1.ServiceEntry) ExporterResource {
	if se == nil {
		return nil
	}
	return newExportToResource(se.Namespace, se.ExportTo)
}