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
	"sync"

	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	istioextensionsv1alpha1 "istio.io/client-go/pkg/apis/extensions/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fetchDestinationRules fetches and converts all destination rules from the cluster
func (k *Client) fetchDestinationRules(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.DestinationRule, errChan chan<- error) {
	defer wg.Done()
	drList, err := k.istioClient.NetworkingV1beta1().DestinationRules("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list destination rules: %w", err)
		return
	}

	var protoDestinationRules []*typesv1alpha1.DestinationRule
	for i := range drList.Items {
		dr := drList.Items[i]
		protoDR, convertErr := k.convertDestinationRule(dr)
		if convertErr != nil {
			k.logger.Warn("failed to convert destination rule", "name", dr.Name, "namespace", dr.Namespace, "error", convertErr)
			continue
		}
		protoDestinationRules = append(protoDestinationRules, protoDR)
	}
	*result = protoDestinationRules
}

// fetchEnvoyFilters fetches and converts all envoy filters from the cluster
func (k *Client) fetchEnvoyFilters(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.EnvoyFilter, errChan chan<- error) {
	defer wg.Done()
	efList, err := k.istioClient.NetworkingV1alpha3().EnvoyFilters("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list envoy filters: %w", err)
		return
	}

	var protoEnvoyFilters []*typesv1alpha1.EnvoyFilter
	for i := range efList.Items {
		ef := efList.Items[i]
		protoEF, convertErr := k.convertEnvoyFilter(ef)
		if convertErr != nil {
			k.logger.Warn("failed to convert envoy filter", "name", ef.Name, "namespace", ef.Namespace, "error", convertErr)
			continue
		}
		protoEnvoyFilters = append(protoEnvoyFilters, protoEF)
	}
	*result = protoEnvoyFilters
}

// fetchRequestAuthentications fetches and converts all request authentications from the cluster
func (k *Client) fetchRequestAuthentications(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.RequestAuthentication, errChan chan<- error) {
	defer wg.Done()
	raList, err := k.istioClient.SecurityV1beta1().RequestAuthentications("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list request authentications: %w", err)
		return
	}

	var protoRequestAuthentications []*typesv1alpha1.RequestAuthentication
	for i := range raList.Items {
		ra := raList.Items[i]
		protoRA, convertErr := k.convertRequestAuthentication(ra)
		if convertErr != nil {
			k.logger.Warn("failed to convert request authentication", "name", ra.Name, "namespace", ra.Namespace, "error", convertErr)
			continue
		}
		protoRequestAuthentications = append(protoRequestAuthentications, protoRA)
	}
	*result = protoRequestAuthentications
}

// fetchPeerAuthentications fetches and converts all peer authentications from the cluster
func (k *Client) fetchPeerAuthentications(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.PeerAuthentication, errChan chan<- error) {
	defer wg.Done()
	paList, err := k.istioClient.SecurityV1beta1().PeerAuthentications("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list peer authentications: %w", err)
		return
	}

	var protoPeerAuthentications []*typesv1alpha1.PeerAuthentication
	for i := range paList.Items {
		pa := paList.Items[i]
		protoPA, convertErr := k.convertPeerAuthentication(pa)
		if convertErr != nil {
			k.logger.Warn("failed to convert peer authentication", "name", pa.Name, "namespace", pa.Namespace, "error", convertErr)
			continue
		}
		protoPeerAuthentications = append(protoPeerAuthentications, protoPA)
	}
	*result = protoPeerAuthentications
}

// fetchAuthorizationPolicies fetches and converts all authorization policies from the cluster
func (k *Client) fetchAuthorizationPolicies(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.AuthorizationPolicy, errChan chan<- error) {
	defer wg.Done()
	apList, err := k.istioClient.SecurityV1beta1().AuthorizationPolicies("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list authorization policies: %w", err)
		return
	}

	var protoAuthorizationPolicies []*typesv1alpha1.AuthorizationPolicy
	for i := range apList.Items {
		ap := apList.Items[i]
		protoAP, convertErr := k.convertAuthorizationPolicy(ap)
		if convertErr != nil {
			k.logger.Warn("failed to convert authorization policy", "name", ap.Name, "namespace", ap.Namespace, "error", convertErr)
			continue
		}
		protoAuthorizationPolicies = append(protoAuthorizationPolicies, protoAP)
	}
	*result = protoAuthorizationPolicies
}

// fetchWasmPlugins fetches and converts all wasm plugins from the cluster
func (k *Client) fetchWasmPlugins(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.WasmPlugin, errChan chan<- error) {
	defer wg.Done()
	wpList, err := k.istioClient.ExtensionsV1alpha1().WasmPlugins("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list wasm plugins: %w", err)
		return
	}

	var protoWasmPlugins []*typesv1alpha1.WasmPlugin
	for i := range wpList.Items {
		wp := wpList.Items[i]
		protoWP, convertErr := k.convertWasmPlugin(wp)
		if convertErr != nil {
			k.logger.Warn("failed to convert wasm plugin", "name", wp.Name, "namespace", wp.Namespace, "error", convertErr)
			continue
		}
		protoWasmPlugins = append(protoWasmPlugins, protoWP)
	}
	*result = protoWasmPlugins
}

// fetchGateways fetches and converts all gateways from the cluster
func (k *Client) fetchGateways(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.Gateway, errChan chan<- error) {
	defer wg.Done()
	gwList, err := k.istioClient.NetworkingV1beta1().Gateways("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list gateways: %w", err)
		return
	}

	var protoGateways []*typesv1alpha1.Gateway
	for i := range gwList.Items {
		gw := gwList.Items[i]
		protoGW, convertErr := k.convertGateway(gw)
		if convertErr != nil {
			k.logger.Warn("failed to convert gateway", "name", gw.Name, "namespace", gw.Namespace, "error", convertErr)
			continue
		}
		protoGateways = append(protoGateways, protoGW)
	}
	*result = protoGateways
}

// fetchSidecars fetches and converts all sidecars from the cluster
func (k *Client) fetchSidecars(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.Sidecar, errChan chan<- error) {
	defer wg.Done()
	scList, err := k.istioClient.NetworkingV1beta1().Sidecars("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list sidecars: %w", err)
		return
	}

	var protoSidecars []*typesv1alpha1.Sidecar
	for i := range scList.Items {
		sc := scList.Items[i]
		protoSC, convertErr := k.convertSidecar(sc)
		if convertErr != nil {
			k.logger.Warn("failed to convert sidecar", "name", sc.Name, "namespace", sc.Namespace, "error", convertErr)
			continue
		}
		protoSidecars = append(protoSidecars, protoSC)
	}
	*result = protoSidecars
}

// fetchVirtualServices fetches and converts all virtual services from the cluster
func (k *Client) fetchVirtualServices(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.VirtualService, errChan chan<- error) {
	defer wg.Done()
	vsList, err := k.istioClient.NetworkingV1beta1().VirtualServices("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list virtual services: %w", err)
		return
	}

	var protoVirtualServices []*typesv1alpha1.VirtualService
	for i := range vsList.Items {
		vs := vsList.Items[i]
		protoVS, convertErr := k.convertVirtualService(vs)
		if convertErr != nil {
			k.logger.Warn("failed to convert virtual service", "name", vs.Name, "namespace", vs.Namespace, "error", convertErr)
			continue
		}
		protoVirtualServices = append(protoVirtualServices, protoVS)
	}
	*result = protoVirtualServices
}

// fetchServiceEntries fetches and converts all service entries from the cluster
func (k *Client) fetchServiceEntries(ctx context.Context, wg *sync.WaitGroup, result *[]*typesv1alpha1.ServiceEntry, errChan chan<- error) {
	defer wg.Done()
	seList, err := k.istioClient.NetworkingV1beta1().ServiceEntries("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list service entries: %w", err)
		return
	}

	var protoServiceEntries []*typesv1alpha1.ServiceEntry
	for i := range seList.Items {
		se := seList.Items[i]
		protoSE, convertErr := k.convertServiceEntry(se)
		if convertErr != nil {
			k.logger.Warn("failed to convert service entry", "name", se.Name, "namespace", se.Namespace, "error", convertErr)
			continue
		}
		protoServiceEntries = append(protoServiceEntries, protoSE)
	}
	*result = protoServiceEntries
}

// convertDestinationRule converts an Istio DestinationRule to a protobuf DestinationRule
func (k *Client) convertDestinationRule(dr *istionetworkingv1beta1.DestinationRule) (*typesv1alpha1.DestinationRule, error) {
	resourceBytes, err := json.Marshal(dr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destination rule resource: %w", err)
	}

	// Extract host from the spec
	var host string
	if dr.Spec.Host != "" {
		host = dr.Spec.Host
	}

	// Extract subsets from the spec
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

	// Default for exportTo is ["*"] if not specified or empty
	var exportTo []string
	if len(dr.Spec.ExportTo) > 0 {
		exportTo = dr.Spec.ExportTo
	} else {
		exportTo = []string{"*"}
	}

	// Extract workload selector from the spec
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

	// Extract workload selector from the spec
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

	// Extract target refs from the spec
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

	// Extract selector from the spec
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

	// Extract target refs from the spec
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

	// Extract selector from the spec
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

	// Extract selector from the spec
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

	// Extract target refs from the spec
	var targetRefs []*typesv1alpha1.PolicyTargetReference
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

	// Extract selector from the spec
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

	// Extract target refs from the spec
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

	// Extract selector from gateway spec
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

	// Extract workload selector from the spec
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

	// Extract hosts, gateways, and exportTo from the spec
	var hosts []string
	if vs.Spec.Hosts != nil {
		hosts = vs.Spec.Hosts
	}

	// Default for gateways is ["mesh"] if not specified or empty
	var gateways []string
	if len(vs.Spec.Gateways) > 0 {
		gateways = vs.Spec.Gateways
	} else {
		gateways = []string{"mesh"}
	}

	// Default for exportTo is ["*"] if not specified or empty
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

// fetchIstioControlPlaneConfig fetches Istio control plane configuration.
// Supports canary upgrades and revision-based Istio installations by discovering
// all istiod deployments and selecting the active control plane.
func (k *Client) fetchIstioControlPlaneConfig(ctx context.Context, wg *sync.WaitGroup, result **typesv1alpha1.IstioControlPlaneConfig, errChan chan<- error) {
	defer wg.Done()

	config := &typesv1alpha1.IstioControlPlaneConfig{
		PilotScopeGatewayToNamespace: false,          // default value
		RootNamespace:                "istio-system", // default fallback
	}

	// First, try to find the root namespace by searching across multiple common namespaces
	rootNamespace, activeDeployment := k.discoverIstioControlPlane(ctx)
	if rootNamespace != "" {
		config.RootNamespace = rootNamespace
		k.logger.Debug("discovered Istio root namespace", "namespace", rootNamespace)
	}

	if activeDeployment == nil {
		k.logger.Debug("no active istiod deployment found, using default Istio configuration")
		*result = config
		return
	}

	k.logger.Debug("selected active istiod deployment", "name", activeDeployment.Name, "namespace", activeDeployment.Namespace)

	// Extract configuration from the active deployment
	k.extractPilotConfiguration(activeDeployment, config)

	*result = config
}

// discoverIstioControlPlane discovers the Istio control plane by searching across multiple
// potential namespaces and returns the root namespace and active deployment.
func (k *Client) discoverIstioControlPlane(ctx context.Context) (string, *appsv1.Deployment) {
	// Common namespaces where Istio control plane might be installed
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
			// Add any namespace that looks like it could contain Istio control plane
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

	// Search each namespace for istiod deployments
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

		// Select the best deployment from this namespace
		activeDeployment := k.selectActiveControlPlane(deployments.Items)
		if activeDeployment == nil {
			continue
		}

		// Prioritize traditional "istio-system" namespace
		if namespace == "istio-system" {
			k.logger.Debug("found istiod in traditional istio-system namespace")
			return namespace, activeDeployment
		}

		// Otherwise, prefer deployment with most ready replicas
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
	// Check for PILOT_SCOPE_GATEWAY_TO_NAMESPACE environment variable in istiod deployment
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

// convertServiceEntry converts an Istio ServiceEntry to a protobuf ServiceEntry
func (k *Client) convertServiceEntry(se *istionetworkingv1beta1.ServiceEntry) (*typesv1alpha1.ServiceEntry, error) {
	resourceBytes, err := json.Marshal(se)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service entry resource: %w", err)
	}

	// Default for exportTo is ["*"] if not specified or empty
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
