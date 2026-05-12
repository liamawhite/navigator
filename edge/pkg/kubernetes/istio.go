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

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	istioextensionsv1alpha1 "istio.io/client-go/pkg/apis/extensions/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// listDestinationRules reads DestinationRules from the informer cache and converts them.
func (k *Client) listDestinationRules() []*typesv1alpha1.DestinationRule {
	items, err := k.destinationRulesLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list destination rules from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.DestinationRule
	for _, dr := range items {
		proto, err := k.convertDestinationRule(dr)
		if err != nil {
			k.logger.Warn("failed to convert destination rule", "name", dr.Name, "namespace", dr.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listEnvoyFilters reads EnvoyFilters from the informer cache and converts them.
func (k *Client) listEnvoyFilters() []*typesv1alpha1.EnvoyFilter {
	items, err := k.envoyFiltersLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list envoy filters from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.EnvoyFilter
	for _, ef := range items {
		proto, err := k.convertEnvoyFilter(ef)
		if err != nil {
			k.logger.Warn("failed to convert envoy filter", "name", ef.Name, "namespace", ef.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listRequestAuthentications reads RequestAuthentications from the informer cache and converts them.
func (k *Client) listRequestAuthentications() []*typesv1alpha1.RequestAuthentication {
	items, err := k.requestAuthenticationsLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list request authentications from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.RequestAuthentication
	for _, ra := range items {
		proto, err := k.convertRequestAuthentication(ra)
		if err != nil {
			k.logger.Warn("failed to convert request authentication", "name", ra.Name, "namespace", ra.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listPeerAuthentications reads PeerAuthentications from the informer cache and converts them.
func (k *Client) listPeerAuthentications() []*typesv1alpha1.PeerAuthentication {
	items, err := k.peerAuthenticationsLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list peer authentications from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.PeerAuthentication
	for _, pa := range items {
		proto, err := k.convertPeerAuthentication(pa)
		if err != nil {
			k.logger.Warn("failed to convert peer authentication", "name", pa.Name, "namespace", pa.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listAuthorizationPolicies reads AuthorizationPolicies from the informer cache and converts them.
func (k *Client) listAuthorizationPolicies() []*typesv1alpha1.AuthorizationPolicy {
	items, err := k.authorizationPoliciesLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list authorization policies from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.AuthorizationPolicy
	for _, ap := range items {
		proto, err := k.convertAuthorizationPolicy(ap)
		if err != nil {
			k.logger.Warn("failed to convert authorization policy", "name", ap.Name, "namespace", ap.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listWasmPlugins reads WasmPlugins from the informer cache and converts them.
func (k *Client) listWasmPlugins() []*typesv1alpha1.WasmPlugin {
	items, err := k.wasmPluginsLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list wasm plugins from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.WasmPlugin
	for _, wp := range items {
		proto, err := k.convertWasmPlugin(wp)
		if err != nil {
			k.logger.Warn("failed to convert wasm plugin", "name", wp.Name, "namespace", wp.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listGateways reads Gateways from the informer cache and converts them.
func (k *Client) listGateways() []*typesv1alpha1.Gateway {
	items, err := k.gatewaysLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list gateways from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.Gateway
	for _, gw := range items {
		proto, err := k.convertGateway(gw)
		if err != nil {
			k.logger.Warn("failed to convert gateway", "name", gw.Name, "namespace", gw.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listSidecars reads Sidecars from the informer cache and converts them.
func (k *Client) listSidecars() []*typesv1alpha1.Sidecar {
	items, err := k.sidecarsLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list sidecars from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.Sidecar
	for _, sc := range items {
		proto, err := k.convertSidecar(sc)
		if err != nil {
			k.logger.Warn("failed to convert sidecar", "name", sc.Name, "namespace", sc.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listVirtualServices reads VirtualServices from the informer cache and converts them.
func (k *Client) listVirtualServices() []*typesv1alpha1.VirtualService {
	items, err := k.virtualServicesLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list virtual services from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.VirtualService
	for _, vs := range items {
		proto, err := k.convertVirtualService(vs)
		if err != nil {
			k.logger.Warn("failed to convert virtual service", "name", vs.Name, "namespace", vs.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// listServiceEntries reads ServiceEntries from the informer cache and converts them.
func (k *Client) listServiceEntries() []*typesv1alpha1.ServiceEntry {
	items, err := k.serviceEntriesLister.List(labels.Everything())
	if err != nil {
		k.logger.Error("failed to list service entries from cache", "error", err)
		return nil
	}
	var result []*typesv1alpha1.ServiceEntry
	for _, se := range items {
		proto, err := k.convertServiceEntry(se)
		if err != nil {
			k.logger.Warn("failed to convert service entry", "name", se.Name, "namespace", se.Namespace, "error", err)
			continue
		}
		result = append(result, proto)
	}
	return result
}

// getIstioControlPlaneConfig reads Istio control plane config from the informer cache.
func (k *Client) getIstioControlPlaneConfig() *typesv1alpha1.IstioControlPlaneConfig {
	config := &typesv1alpha1.IstioControlPlaneConfig{
		PilotScopeGatewayToNamespace: false,
		RootNamespace:                "istio-system",
	}

	// List all deployments with label app=istiod from the cache
	selector := labels.Set{"app": "istiod"}.AsSelector()
	allDeps, err := k.deploymentsLister.List(selector)
	if err != nil {
		k.logger.Debug("failed to list deployments from cache, using defaults", "error", err)
		return config
	}

	if len(allDeps) == 0 {
		k.logger.Debug("no istiod deployments found in cache, using default Istio configuration")
		return config
	}

	// Group by namespace to replicate discoverIstioControlPlane logic
	byNamespace := make(map[string][]*appsv1.Deployment)
	for _, d := range allDeps {
		byNamespace[d.Namespace] = append(byNamespace[d.Namespace], d)
	}

	// Prefer istio-system namespace
	if deps, ok := byNamespace["istio-system"]; ok {
		active := k.selectActiveControlPlane(derefDeployments(deps))
		if active != nil {
			config.RootNamespace = "istio-system"
			k.logger.Debug("selected active istiod deployment", "name", active.Name, "namespace", active.Namespace)
			k.extractPilotConfiguration(active, config)
			return config
		}
	}

	// Fall back to namespace with most-ready-replica deployment
	var bestDep *appsv1.Deployment
	var bestNS string
	maxReady := int32(-1)
	for ns, deps := range byNamespace {
		active := k.selectActiveControlPlane(derefDeployments(deps))
		if active != nil && active.Status.ReadyReplicas > maxReady {
			maxReady = active.Status.ReadyReplicas
			bestDep = active
			bestNS = ns
		}
	}

	if bestDep != nil {
		config.RootNamespace = bestNS
		k.logger.Debug("discovered Istio control plane", "namespace", bestNS, "deployment", bestDep.Name)
		k.extractPilotConfiguration(bestDep, config)
	}

	return config
}

// discoverIstioControlPlane discovers the Istio control plane via direct API calls.
// Used by GetClusterName before informers are started.
func (k *Client) discoverIstioControlPlane(ctx context.Context) (string, *appsv1.Deployment) {
	candidateNamespaces := []string{
		"istio-system",
		"istio-control-plane",
		"istiod",
		"istio",
	}

	// Also check all namespaces for istiod deployments (for custom installations)
	allNamespaces, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, ns := range allNamespaces.Items {
			if ns.Name != "istio-system" &&
				ns.Name != "istio-control-plane" &&
				ns.Name != "istiod" &&
				ns.Name != "istio" {
				candidateNamespaces = append(candidateNamespaces, ns.Name)
			}
		}
	}

	var bestDeployment *appsv1.Deployment
	var bestNamespace string
	maxReadyReplicas := int32(-1)

	for _, namespace := range candidateNamespaces {
		deployments, err := k.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app=istiod",
		})
		if err != nil {
			continue
		}

		if len(deployments.Items) == 0 {
			continue
		}

		activeDeployment := k.selectActiveControlPlane(deployments.Items)
		if activeDeployment == nil {
			continue
		}

		if namespace == "istio-system" {
			k.logger.Debug("found istiod in traditional istio-system namespace")
			return namespace, activeDeployment
		}

		if activeDeployment.Status.ReadyReplicas > maxReadyReplicas {
			maxReadyReplicas = activeDeployment.Status.ReadyReplicas
			bestDeployment = activeDeployment
			bestNamespace = namespace
		}
	}

	if bestDeployment != nil {
		k.logger.Debug("discovered Istio control plane",
			"namespace", bestNamespace,
			"deployment", bestDeployment.Name,
			"readyReplicas", maxReadyReplicas)
		return bestNamespace, bestDeployment
	}

	k.logger.Debug("no istiod deployments found in any namespace")
	return "", nil
}

// selectActiveControlPlane selects the active control plane from multiple istiod deployments.
// Priority order:
// 1. Deployment named "istiod" (traditional default)
// 2. Deployment with highest ready replicas
// 3. First deployment (fallback)
func (k *Client) selectActiveControlPlane(deployments []appsv1.Deployment) *appsv1.Deployment {
	if len(deployments) == 0 {
		return nil
	}

	// Priority 1: Look for traditional "istiod" deployment
	for i := range deployments {
		if deployments[i].Name == "istiod" {
			k.logger.Debug("found traditional istiod deployment")
			return &deployments[i]
		}
	}

	// Priority 2: Select deployment with highest ready replicas
	var bestDeployment *appsv1.Deployment
	maxReadyReplicas := int32(-1)

	for i := range deployments {
		deployment := &deployments[i]
		readyReplicas := deployment.Status.ReadyReplicas

		if readyReplicas > maxReadyReplicas {
			maxReadyReplicas = readyReplicas
			bestDeployment = deployment
		}
	}

	if bestDeployment != nil {
		k.logger.Debug("selected deployment with most ready replicas",
			"name", bestDeployment.Name,
			"readyReplicas", maxReadyReplicas)
		return bestDeployment
	}

	// Priority 3: Fallback to first deployment
	k.logger.Debug("using first available deployment as fallback", "name", deployments[0].Name)
	return &deployments[0]
}

// extractPilotConfiguration extracts pilot configuration from an istiod deployment
func (k *Client) extractPilotConfiguration(deployment *appsv1.Deployment, config *typesv1alpha1.IstioControlPlaneConfig) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == "discovery" {
			for _, env := range container.Env {
				if env.Name == "PILOT_SCOPE_GATEWAY_TO_NAMESPACE" {
					if env.Value == "true" {
						config.PilotScopeGatewayToNamespace = true
						k.logger.Debug("found PILOT_SCOPE_GATEWAY_TO_NAMESPACE=true", "deployment", deployment.Name)
					}
					return
				}
			}
			break
		}
	}
}

// convertDestinationRule converts an Istio DestinationRule to a protobuf DestinationRule
func (k *Client) convertDestinationRule(dr *istionetworkingv1beta1.DestinationRule) (*typesv1alpha1.DestinationRule, error) {
	resourceBytes, err := json.Marshal(dr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destination rule resource: %w", err)
	}

	var host string
	if dr.Spec.Host != "" {
		host = dr.Spec.Host
	}

	var subsets []*typesv1alpha1.DestinationRuleSubset
	for _, subset := range dr.Spec.Subsets {
		protoSubset := &typesv1alpha1.DestinationRuleSubset{
			Name:   subset.Name,
			Labels: make(map[string]string),
		}
		if subset.Labels != nil {
			for key, value := range subset.Labels {
				protoSubset.Labels[key] = value
			}
		}
		subsets = append(subsets, protoSubset)
	}

	var exportTo []string
	if len(dr.Spec.ExportTo) > 0 {
		exportTo = dr.Spec.ExportTo
	} else {
		exportTo = []string{"*"}
	}

	var workloadSelector *typesv1alpha1.WorkloadSelector
	if dr.Spec.WorkloadSelector != nil && dr.Spec.WorkloadSelector.MatchLabels != nil {
		matchLabels := make(map[string]string)
		for key, value := range dr.Spec.WorkloadSelector.MatchLabels {
			matchLabels[key] = value
		}
		workloadSelector = &typesv1alpha1.WorkloadSelector{
			MatchLabels: matchLabels,
		}
	}

	return &typesv1alpha1.DestinationRule{
		Name:             dr.Name,
		Namespace:        dr.Namespace,
		RawConfig:        string(resourceBytes),
		Host:             host,
		Subsets:          subsets,
		ExportTo:         exportTo,
		WorkloadSelector: workloadSelector,
	}, nil
}

// convertEnvoyFilter converts an Istio EnvoyFilter to a protobuf EnvoyFilter
func (k *Client) convertEnvoyFilter(ef *istionetworkingv1alpha3.EnvoyFilter) (*typesv1alpha1.EnvoyFilter, error) {
	resourceBytes, err := json.Marshal(ef)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envoy filter resource: %w", err)
	}

	var workloadSelector *typesv1alpha1.WorkloadSelector
	if ef.Spec.WorkloadSelector != nil && ef.Spec.WorkloadSelector.Labels != nil {
		matchLabels := make(map[string]string)
		for key, value := range ef.Spec.WorkloadSelector.Labels {
			matchLabels[key] = value
		}
		workloadSelector = &typesv1alpha1.WorkloadSelector{
			MatchLabels: matchLabels,
		}
	}

	var targetRefs []*typesv1alpha1.PolicyTargetReference
	for _, targetRef := range ef.Spec.TargetRefs {
		if targetRef != nil {
			protoTargetRef := &typesv1alpha1.PolicyTargetReference{
				Group:     targetRef.Group,
				Kind:      targetRef.Kind,
				Name:      targetRef.Name,
				Namespace: targetRef.Namespace,
			}
			targetRefs = append(targetRefs, protoTargetRef)
		}
	}

	return &typesv1alpha1.EnvoyFilter{
		Name:             ef.Name,
		Namespace:        ef.Namespace,
		RawConfig:        string(resourceBytes),
		WorkloadSelector: workloadSelector,
		TargetRefs:       targetRefs,
	}, nil
}

// convertRequestAuthentication converts an Istio RequestAuthentication to a protobuf RequestAuthentication
func (k *Client) convertRequestAuthentication(ra *istiosecurityv1beta1.RequestAuthentication) (*typesv1alpha1.RequestAuthentication, error) {
	resourceBytes, err := json.Marshal(ra)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request authentication resource: %w", err)
	}

	var selector *typesv1alpha1.WorkloadSelector
	if ra.Spec.Selector != nil && ra.Spec.Selector.MatchLabels != nil {
		matchLabels := make(map[string]string)
		for key, value := range ra.Spec.Selector.MatchLabels {
			matchLabels[key] = value
		}
		selector = &typesv1alpha1.WorkloadSelector{
			MatchLabels: matchLabels,
		}
	}

	var targetRefs []*typesv1alpha1.PolicyTargetReference
	for _, targetRef := range ra.Spec.TargetRefs {
		if targetRef != nil {
			protoTargetRef := &typesv1alpha1.PolicyTargetReference{
				Group:     targetRef.Group,
				Kind:      targetRef.Kind,
				Name:      targetRef.Name,
				Namespace: targetRef.Namespace,
			}
			targetRefs = append(targetRefs, protoTargetRef)
		}
	}

	return &typesv1alpha1.RequestAuthentication{
		Name:       ra.Name,
		Namespace:  ra.Namespace,
		RawConfig:  string(resourceBytes),
		Selector:   selector,
		TargetRefs: targetRefs,
	}, nil
}

// convertPeerAuthentication converts an Istio PeerAuthentication to a protobuf PeerAuthentication
func (k *Client) convertPeerAuthentication(pa *istiosecurityv1beta1.PeerAuthentication) (*typesv1alpha1.PeerAuthentication, error) {
	resourceBytes, err := json.Marshal(pa)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal peer authentication resource: %w", err)
	}

	var selector *typesv1alpha1.WorkloadSelector
	if pa.Spec.Selector != nil && pa.Spec.Selector.MatchLabels != nil {
		matchLabels := make(map[string]string)
		for key, value := range pa.Spec.Selector.MatchLabels {
			matchLabels[key] = value
		}
		selector = &typesv1alpha1.WorkloadSelector{
			MatchLabels: matchLabels,
		}
	}

	return &typesv1alpha1.PeerAuthentication{
		Name:      pa.Name,
		Namespace: pa.Namespace,
		RawConfig: string(resourceBytes),
		Selector:  selector,
	}, nil
}

// convertAuthorizationPolicy converts an Istio AuthorizationPolicy to a protobuf AuthorizationPolicy
func (k *Client) convertAuthorizationPolicy(ap *istiosecurityv1beta1.AuthorizationPolicy) (*typesv1alpha1.AuthorizationPolicy, error) {
	resourceBytes, err := json.Marshal(ap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization policy resource: %w", err)
	}

	var selector *typesv1alpha1.WorkloadSelector
	if ap.Spec.Selector != nil && ap.Spec.Selector.MatchLabels != nil {
		matchLabels := make(map[string]string)
		for key, value := range ap.Spec.Selector.MatchLabels {
			matchLabels[key] = value
		}
		selector = &typesv1alpha1.WorkloadSelector{
			MatchLabels: matchLabels,
		}
	}

	var targetRefs []*typesv1alpha1.PolicyTargetReference

	// Handle TargetRef (singular) - older API
	if ap.Spec.TargetRef != nil {
		protoTargetRef := &typesv1alpha1.PolicyTargetReference{
			Group:     ap.Spec.TargetRef.Group,
			Kind:      ap.Spec.TargetRef.Kind,
			Name:      ap.Spec.TargetRef.Name,
			Namespace: ap.Spec.TargetRef.Namespace,
		}
		targetRefs = append(targetRefs, protoTargetRef)
	}

	// Handle TargetRefs (plural) - newer API
	for _, targetRef := range ap.Spec.TargetRefs {
		if targetRef != nil {
			protoTargetRef := &typesv1alpha1.PolicyTargetReference{
				Group:     targetRef.Group,
				Kind:      targetRef.Kind,
				Name:      targetRef.Name,
				Namespace: targetRef.Namespace,
			}
			targetRefs = append(targetRefs, protoTargetRef)
		}
	}

	return &typesv1alpha1.AuthorizationPolicy{
		Name:       ap.Name,
		Namespace:  ap.Namespace,
		RawConfig:  string(resourceBytes),
		Selector:   selector,
		TargetRefs: targetRefs,
	}, nil
}

// convertWasmPlugin converts an Istio WasmPlugin to a protobuf WasmPlugin
func (k *Client) convertWasmPlugin(wp *istioextensionsv1alpha1.WasmPlugin) (*typesv1alpha1.WasmPlugin, error) {
	resourceBytes, err := json.Marshal(wp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wasm plugin resource: %w", err)
	}

	var selector *typesv1alpha1.WorkloadSelector
	if wp.Spec.Selector != nil && wp.Spec.Selector.MatchLabels != nil {
		matchLabels := make(map[string]string)
		for key, value := range wp.Spec.Selector.MatchLabels {
			matchLabels[key] = value
		}
		selector = &typesv1alpha1.WorkloadSelector{
			MatchLabels: matchLabels,
		}
	}

	var targetRefs []*typesv1alpha1.PolicyTargetReference
	for _, targetRef := range wp.Spec.TargetRefs {
		if targetRef != nil {
			protoTargetRef := &typesv1alpha1.PolicyTargetReference{
				Group:     targetRef.Group,
				Kind:      targetRef.Kind,
				Name:      targetRef.Name,
				Namespace: targetRef.Namespace,
			}
			targetRefs = append(targetRefs, protoTargetRef)
		}
	}

	return &typesv1alpha1.WasmPlugin{
		Name:       wp.Name,
		Namespace:  wp.Namespace,
		RawConfig:  string(resourceBytes),
		Selector:   selector,
		TargetRefs: targetRefs,
	}, nil
}

// convertGateway converts an Istio Gateway to a protobuf Gateway
func (k *Client) convertGateway(gw *istionetworkingv1beta1.Gateway) (*typesv1alpha1.Gateway, error) {
	resourceBytes, err := json.Marshal(gw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gateway resource: %w", err)
	}

	selector := make(map[string]string)
	if gw.Spec.Selector != nil {
		for key, value := range gw.Spec.Selector {
			selector[key] = value
		}
	}

	return &typesv1alpha1.Gateway{
		Name:      gw.Name,
		Namespace: gw.Namespace,
		RawConfig: string(resourceBytes),
		Selector:  selector,
	}, nil
}

// convertSidecar converts an Istio Sidecar to a protobuf Sidecar
func (k *Client) convertSidecar(sc *istionetworkingv1beta1.Sidecar) (*typesv1alpha1.Sidecar, error) {
	resourceBytes, err := json.Marshal(sc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sidecar resource: %w", err)
	}

	var workloadSelector *typesv1alpha1.WorkloadSelector
	if sc.Spec.WorkloadSelector != nil && sc.Spec.WorkloadSelector.Labels != nil {
		matchLabels := make(map[string]string)
		for key, value := range sc.Spec.WorkloadSelector.Labels {
			matchLabels[key] = value
		}
		workloadSelector = &typesv1alpha1.WorkloadSelector{
			MatchLabels: matchLabels,
		}
	}

	return &typesv1alpha1.Sidecar{
		Name:             sc.Name,
		Namespace:        sc.Namespace,
		RawConfig:        string(resourceBytes),
		WorkloadSelector: workloadSelector,
	}, nil
}

// convertVirtualService converts an Istio VirtualService to a protobuf VirtualService
func (k *Client) convertVirtualService(vs *istionetworkingv1beta1.VirtualService) (*typesv1alpha1.VirtualService, error) {
	resourceBytes, err := json.Marshal(vs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal virtual service resource: %w", err)
	}

	var hosts []string
	if vs.Spec.Hosts != nil {
		hosts = vs.Spec.Hosts
	}

	var gateways []string
	if len(vs.Spec.Gateways) > 0 {
		gateways = vs.Spec.Gateways
	} else {
		gateways = []string{"mesh"}
	}

	var exportTo []string
	if len(vs.Spec.ExportTo) > 0 {
		exportTo = vs.Spec.ExportTo
	} else {
		exportTo = []string{"*"}
	}

	return &typesv1alpha1.VirtualService{
		Name:      vs.Name,
		Namespace: vs.Namespace,
		RawConfig: string(resourceBytes),
		Hosts:     hosts,
		Gateways:  gateways,
		ExportTo:  exportTo,
	}, nil
}

// convertServiceEntry converts an Istio ServiceEntry to a protobuf ServiceEntry
func (k *Client) convertServiceEntry(se *istionetworkingv1beta1.ServiceEntry) (*typesv1alpha1.ServiceEntry, error) {
	resourceBytes, err := json.Marshal(se)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service entry resource: %w", err)
	}

	var exportTo []string
	if len(se.Spec.ExportTo) > 0 {
		exportTo = se.Spec.ExportTo
	} else {
		exportTo = []string{"*"}
	}

	return &typesv1alpha1.ServiceEntry{
		Name:      se.Name,
		Namespace: se.Namespace,
		RawConfig: string(resourceBytes),
		ExportTo:  exportTo,
	}, nil
}

// derefDeployments converts a slice of Deployment pointers to values
func derefDeployments(ptrs []*appsv1.Deployment) []appsv1.Deployment {
	out := make([]appsv1.Deployment, len(ptrs))
	for i, p := range ptrs {
		out[i] = *p
	}
	return out
}
