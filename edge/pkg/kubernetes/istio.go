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

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fetchDestinationRules fetches and converts all destination rules from the cluster
func (k *Client) fetchDestinationRules(ctx context.Context, wg *sync.WaitGroup, result *[]*v1alpha1.DestinationRule, errChan chan<- error) {
	defer wg.Done()
	drList, err := k.istioClient.NetworkingV1beta1().DestinationRules("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list destination rules: %w", err)
		return
	}

	var protoDestinationRules []*v1alpha1.DestinationRule
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
func (k *Client) fetchEnvoyFilters(ctx context.Context, wg *sync.WaitGroup, result *[]*v1alpha1.EnvoyFilter, errChan chan<- error) {
	defer wg.Done()
	efList, err := k.istioClient.NetworkingV1alpha3().EnvoyFilters("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list envoy filters: %w", err)
		return
	}

	var protoEnvoyFilters []*v1alpha1.EnvoyFilter
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

// fetchGateways fetches and converts all gateways from the cluster
func (k *Client) fetchGateways(ctx context.Context, wg *sync.WaitGroup, result *[]*v1alpha1.Gateway, errChan chan<- error) {
	defer wg.Done()
	gwList, err := k.istioClient.NetworkingV1beta1().Gateways("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list gateways: %w", err)
		return
	}

	var protoGateways []*v1alpha1.Gateway
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
func (k *Client) fetchSidecars(ctx context.Context, wg *sync.WaitGroup, result *[]*v1alpha1.Sidecar, errChan chan<- error) {
	defer wg.Done()
	scList, err := k.istioClient.NetworkingV1beta1().Sidecars("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list sidecars: %w", err)
		return
	}

	var protoSidecars []*v1alpha1.Sidecar
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
func (k *Client) fetchVirtualServices(ctx context.Context, wg *sync.WaitGroup, result *[]*v1alpha1.VirtualService, errChan chan<- error) {
	defer wg.Done()
	vsList, err := k.istioClient.NetworkingV1beta1().VirtualServices("").List(ctx, metav1.ListOptions{})
	if err != nil {
		errChan <- fmt.Errorf("failed to list virtual services: %w", err)
		return
	}

	var protoVirtualServices []*v1alpha1.VirtualService
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

// convertDestinationRule converts an Istio DestinationRule to a protobuf DestinationRule
func (k *Client) convertDestinationRule(dr *istionetworkingv1beta1.DestinationRule) (*v1alpha1.DestinationRule, error) {
	specBytes, err := json.Marshal(&dr.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destination rule spec: %w", err)
	}

	return &v1alpha1.DestinationRule{
		Name:      dr.Name,
		Namespace: dr.Namespace,
		RawSpec:   string(specBytes),
	}, nil
}

// convertEnvoyFilter converts an Istio EnvoyFilter to a protobuf EnvoyFilter
func (k *Client) convertEnvoyFilter(ef *istionetworkingv1alpha3.EnvoyFilter) (*v1alpha1.EnvoyFilter, error) {
	specBytes, err := json.Marshal(&ef.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envoy filter spec: %w", err)
	}

	return &v1alpha1.EnvoyFilter{
		Name:      ef.Name,
		Namespace: ef.Namespace,
		RawSpec:   string(specBytes),
	}, nil
}

// convertGateway converts an Istio Gateway to a protobuf Gateway
func (k *Client) convertGateway(gw *istionetworkingv1beta1.Gateway) (*v1alpha1.Gateway, error) {
	specBytes, err := json.Marshal(&gw.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gateway spec: %w", err)
	}

	// Extract selector from gateway spec
	selector := make(map[string]string)
	if gw.Spec.Selector != nil {
		for key, value := range gw.Spec.Selector {
			selector[key] = value
		}
	}

	return &v1alpha1.Gateway{
		Name:      gw.Name,
		Namespace: gw.Namespace,
		RawSpec:   string(specBytes),
		Selector:  selector,
	}, nil
}

// convertSidecar converts an Istio Sidecar to a protobuf Sidecar
func (k *Client) convertSidecar(sc *istionetworkingv1beta1.Sidecar) (*v1alpha1.Sidecar, error) {
	specBytes, err := json.Marshal(&sc.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sidecar spec: %w", err)
	}

	return &v1alpha1.Sidecar{
		Name:      sc.Name,
		Namespace: sc.Namespace,
		RawSpec:   string(specBytes),
	}, nil
}

// convertVirtualService converts an Istio VirtualService to a protobuf VirtualService
func (k *Client) convertVirtualService(vs *istionetworkingv1beta1.VirtualService) (*v1alpha1.VirtualService, error) {
	specBytes, err := json.Marshal(&vs.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal virtual service spec: %w", err)
	}

	return &v1alpha1.VirtualService{
		Name:      vs.Name,
		Namespace: vs.Namespace,
		RawSpec:   string(specBytes),
	}, nil
}

// fetchIstioControlPlaneConfig fetches Istio control plane configuration.
// Supports canary upgrades and revision-based Istio installations by discovering
// all istiod deployments and selecting the active control plane.
func (k *Client) fetchIstioControlPlaneConfig(ctx context.Context, wg *sync.WaitGroup, result **v1alpha1.IstioControlPlaneConfig, errChan chan<- error) {
	defer wg.Done()

	config := &v1alpha1.IstioControlPlaneConfig{
		PilotScopeGatewayToNamespace: false, // default value
	}

	// Find all istiod deployments using label selector
	deployments, err := k.clientset.AppsV1().Deployments("istio-system").List(ctx, metav1.ListOptions{
		LabelSelector: "app=istiod",
	})
	if err != nil {
		k.logger.Debug("failed to list istiod deployments, using default Istio configuration", "error", err)
		*result = config
		return
	}

	if len(deployments.Items) == 0 {
		k.logger.Debug("no istiod deployments found, using default Istio configuration")
		*result = config
		return
	}

	// Select the active control plane deployment
	activeDeployment := k.selectActiveControlPlane(deployments.Items)
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
func (k *Client) extractPilotConfiguration(deployment *appsv1.Deployment, config *v1alpha1.IstioControlPlaneConfig) {
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
