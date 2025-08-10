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
	"sync"
	"testing"

	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extensionsapi "istio.io/api/extensions/v1alpha1"
	istioapi "istio.io/api/networking/v1alpha3"
	securityapi "istio.io/api/security/v1beta1"
	istiotype "istio.io/api/type/v1beta1"
	istioextensionsv1alpha1 "istio.io/client-go/pkg/apis/extensions/v1alpha1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClient_convertDestinationRule(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name                 string
		destinationRule      *istionetworkingv1beta1.DestinationRule
		wantHost             string
		wantSubsets          []*typesv1alpha1.DestinationRuleSubset
		wantExportTo         []string
		wantWorkloadSelector *typesv1alpha1.WorkloadSelector
	}{
		{
			name: "all fields specified",
			destinationRule: &istionetworkingv1beta1.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dr-full",
					Namespace: "default",
				},
				Spec: istioapi.DestinationRule{
					Host:     "reviews.bookinfo.svc.cluster.local",
					ExportTo: []string{".", "production"},
					Subsets: []*istioapi.Subset{
						{
							Name:   "v1",
							Labels: map[string]string{"version": "v1"},
						},
						{
							Name:   "v2",
							Labels: map[string]string{"version": "v2", "app": "reviews"},
						},
					},
					WorkloadSelector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{"app": "reviews", "tier": "backend"},
					},
				},
			},
			wantHost: "reviews.bookinfo.svc.cluster.local",
			wantSubsets: []*typesv1alpha1.DestinationRuleSubset{
				{Name: "v1", Labels: map[string]string{"version": "v1"}},
				{Name: "v2", Labels: map[string]string{"version": "v2", "app": "reviews"}},
			},
			wantExportTo: []string{".", "production"},
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{"app": "reviews", "tier": "backend"},
			},
		},
		{
			name: "host only - defaults for exportTo",
			destinationRule: &istionetworkingv1beta1.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dr-host-only",
					Namespace: "default",
				},
				Spec: istioapi.DestinationRule{
					Host: "productpage.bookinfo.svc.cluster.local",
				},
			},
			wantHost:             "productpage.bookinfo.svc.cluster.local",
			wantSubsets:          []*typesv1alpha1.DestinationRuleSubset{},
			wantExportTo:         []string{"*"}, // default
			wantWorkloadSelector: nil,
		},
		{
			name: "empty exportTo should get default",
			destinationRule: &istionetworkingv1beta1.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dr-empty-export",
					Namespace: "istio-system",
				},
				Spec: istioapi.DestinationRule{
					Host:     "istio-proxy",
					ExportTo: []string{}, // empty slice
				},
			},
			wantHost:             "istio-proxy",
			wantSubsets:          []*typesv1alpha1.DestinationRuleSubset{},
			wantExportTo:         []string{"*"}, // default for empty slice
			wantWorkloadSelector: nil,
		},
		{
			name: "subsets without labels",
			destinationRule: &istionetworkingv1beta1.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dr-empty-labels",
					Namespace: "default",
				},
				Spec: istioapi.DestinationRule{
					Host: "details.bookinfo.svc.cluster.local",
					Subsets: []*istioapi.Subset{
						{
							Name:   "default",
							Labels: nil, // nil labels
						},
						{
							Name:   "empty",
							Labels: map[string]string{}, // empty labels
						},
					},
					ExportTo: []string{"*"},
				},
			},
			wantHost: "details.bookinfo.svc.cluster.local",
			wantSubsets: []*typesv1alpha1.DestinationRuleSubset{
				{Name: "default", Labels: map[string]string{}},
				{Name: "empty", Labels: map[string]string{}},
			},
			wantExportTo:         []string{"*"},
			wantWorkloadSelector: nil,
		},
		{
			name: "workload selector without labels",
			destinationRule: &istionetworkingv1beta1.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dr-empty-selector",
					Namespace: "default",
				},
				Spec: istioapi.DestinationRule{
					Host: "ratings.bookinfo.svc.cluster.local",
					WorkloadSelector: &istiotype.WorkloadSelector{
						MatchLabels: nil, // nil labels
					},
				},
			},
			wantHost:             "ratings.bookinfo.svc.cluster.local",
			wantSubsets:          []*typesv1alpha1.DestinationRuleSubset{},
			wantExportTo:         []string{"*"}, // default
			wantWorkloadSelector: nil,
		},
		{
			name: "empty host should be preserved",
			destinationRule: &istionetworkingv1beta1.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dr-no-host",
					Namespace: "default",
				},
				Spec: istioapi.DestinationRule{
					Host: "", // empty host
					Subsets: []*istioapi.Subset{
						{
							Name:   "canary",
							Labels: map[string]string{"deployment": "canary"},
						},
					},
				},
			},
			wantHost: "",
			wantSubsets: []*typesv1alpha1.DestinationRuleSubset{
				{Name: "canary", Labels: map[string]string{"deployment": "canary"}},
			},
			wantExportTo:         []string{"*"}, // default
			wantWorkloadSelector: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertDestinationRule(tt.destinationRule)

			require.NoError(t, err)
			assert.Equal(t, tt.destinationRule.Name, result.Name)
			assert.Equal(t, tt.destinationRule.Namespace, result.Namespace)
			assert.Equal(t, tt.wantHost, result.Host)
			assert.Equal(t, len(tt.wantSubsets), len(result.Subsets))

			// Check subsets in detail
			for i, expectedSubset := range tt.wantSubsets {
				assert.Equal(t, expectedSubset.Name, result.Subsets[i].Name)
				assert.Equal(t, expectedSubset.Labels, result.Subsets[i].Labels)
			}

			assert.Equal(t, tt.wantExportTo, result.ExportTo)
			assert.Equal(t, tt.wantWorkloadSelector, result.WorkloadSelector)
			assert.NotEmpty(t, result.RawConfig)
		})
	}
}

func TestClient_convertGateway(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name         string
		gateway      *istionetworkingv1beta1.Gateway
		wantName     string
		wantSelector map[string]string
	}{
		{
			name: "gateway with selector",
			gateway: &istionetworkingv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "default",
				},
				Spec: istioapi.Gateway{
					Selector: map[string]string{
						"istio": "ingressgateway",
						"app":   "gateway",
					},
					Servers: []*istioapi.Server{
						{
							Port: &istioapi.Port{
								Number:   80,
								Name:     "http",
								Protocol: "HTTP",
							},
							Hosts: []string{"example.com"},
						},
					},
				},
			},
			wantName: "test-gateway",
			wantSelector: map[string]string{
				"istio": "ingressgateway",
				"app":   "gateway",
			},
		},
		{
			name: "gateway without selector",
			gateway: &istionetworkingv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway-no-selector",
					Namespace: "default",
				},
				Spec: istioapi.Gateway{
					Servers: []*istioapi.Server{
						{
							Port: &istioapi.Port{
								Number:   443,
								Name:     "https",
								Protocol: "HTTPS",
							},
							Hosts: []string{"secure.example.com"},
						},
					},
				},
			},
			wantName:     "test-gateway-no-selector",
			wantSelector: map[string]string{},
		},
		{
			name: "gateway with empty selector",
			gateway: &istionetworkingv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway-empty-selector",
					Namespace: "default",
				},
				Spec: istioapi.Gateway{
					Selector: map[string]string{},
					Servers: []*istioapi.Server{
						{
							Port: &istioapi.Port{
								Number:   8080,
								Name:     "http-alt",
								Protocol: "HTTP",
							},
							Hosts: []string{"alt.example.com"},
						},
					},
				},
			},
			wantName:     "test-gateway-empty-selector",
			wantSelector: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertGateway(tt.gateway)

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, "default", result.Namespace)
			assert.Equal(t, tt.wantSelector, result.Selector)
			assert.NotEmpty(t, result.RawConfig)
		})
	}
}

func TestClient_fetchIstioControlPlaneConfig(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	type testDeployment struct {
		name          string
		envVars       []corev1.EnvVar
		readyReplicas int32
	}

	tests := []struct {
		name                             string
		deployments                      []testDeployment
		wantPilotScopeGatewayToNamespace bool
		expectedSelectedDeployment       string
	}{
		{
			name:                             "no deployments found - default config",
			deployments:                      []testDeployment{},
			wantPilotScopeGatewayToNamespace: false,
			expectedSelectedDeployment:       "",
		},
		{
			name: "single traditional istiod - no env var set",
			deployments: []testDeployment{
				{name: "istiod", envVars: []corev1.EnvVar{}, readyReplicas: 1},
			},
			wantPilotScopeGatewayToNamespace: false,
			expectedSelectedDeployment:       "istiod",
		},
		{
			name: "single traditional istiod - env var set to true",
			deployments: []testDeployment{
				{
					name: "istiod",
					envVars: []corev1.EnvVar{
						{Name: "PILOT_SCOPE_GATEWAY_TO_NAMESPACE", Value: "true"},
					},
					readyReplicas: 1,
				},
			},
			wantPilotScopeGatewayToNamespace: true,
			expectedSelectedDeployment:       "istiod",
		},
		{
			name: "canary upgrade - traditional istiod preferred",
			deployments: []testDeployment{
				{name: "istiod", envVars: []corev1.EnvVar{{Name: "PILOT_SCOPE_GATEWAY_TO_NAMESPACE", Value: "false"}}, readyReplicas: 1},
				{name: "istiod-1-26-0", envVars: []corev1.EnvVar{{Name: "PILOT_SCOPE_GATEWAY_TO_NAMESPACE", Value: "true"}}, readyReplicas: 2},
			},
			wantPilotScopeGatewayToNamespace: false, // Should use traditional istiod
			expectedSelectedDeployment:       "istiod",
		},
		{
			name: "canary upgrade - no traditional istiod, select by ready replicas",
			deployments: []testDeployment{
				{name: "istiod-1-25-0", envVars: []corev1.EnvVar{}, readyReplicas: 1},
				{name: "istiod-1-26-0", envVars: []corev1.EnvVar{{Name: "PILOT_SCOPE_GATEWAY_TO_NAMESPACE", Value: "true"}}, readyReplicas: 3},
				{name: "istiod-canary", envVars: []corev1.EnvVar{}, readyReplicas: 2},
			},
			wantPilotScopeGatewayToNamespace: true, // Should use istiod-1-26-0 (highest replicas)
			expectedSelectedDeployment:       "istiod-1-26-0",
		},
		{
			name: "revision-based install - single deployment",
			deployments: []testDeployment{
				{
					name: "istiod-1-26-0",
					envVars: []corev1.EnvVar{
						{Name: "PILOT_SCOPE_GATEWAY_TO_NAMESPACE", Value: "true"},
					},
					readyReplicas: 2,
				},
			},
			wantPilotScopeGatewayToNamespace: true,
			expectedSelectedDeployment:       "istiod-1-26-0",
		},
		{
			name: "multiple deployments - same ready replicas, use first",
			deployments: []testDeployment{
				{name: "istiod-1-25-0", envVars: []corev1.EnvVar{}, readyReplicas: 2},
				{name: "istiod-1-26-0", envVars: []corev1.EnvVar{{Name: "PILOT_SCOPE_GATEWAY_TO_NAMESPACE", Value: "true"}}, readyReplicas: 2},
			},
			wantPilotScopeGatewayToNamespace: false, // Should use first one (istiod-1-25-0)
			expectedSelectedDeployment:       "istiod-1-25-0",
		},
		{
			name: "deployment with zero ready replicas",
			deployments: []testDeployment{
				{name: "istiod-1-26-0", envVars: []corev1.EnvVar{{Name: "PILOT_SCOPE_GATEWAY_TO_NAMESPACE", Value: "true"}}, readyReplicas: 0},
			},
			wantPilotScopeGatewayToNamespace: true,
			expectedSelectedDeployment:       "istiod-1-26-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake Kubernetes client
			k8sClient := fake.NewSimpleClientset()
			client.clientset = k8sClient

			// Create all specified deployments
			for _, deployment := range tt.deployments {
				dep := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployment.name,
						Namespace: "istio-system",
						Labels: map[string]string{
							"app": "istiod",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "discovery",
										Env:  deployment.envVars,
									},
								},
							},
						},
					},
					Status: appsv1.DeploymentStatus{
						ReadyReplicas: deployment.readyReplicas,
					},
				}
				_, err := k8sClient.AppsV1().Deployments("istio-system").Create(context.TODO(), dep, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			var wg sync.WaitGroup
			var result *typesv1alpha1.IstioControlPlaneConfig
			errChan := make(chan error, 1)
			wg.Add(1)

			client.fetchIstioControlPlaneConfig(context.TODO(), &wg, &result, errChan)

			wg.Wait()
			close(errChan)

			// Check for errors
			var errors []error
			for err := range errChan {
				if err != nil {
					errors = append(errors, err)
				}
			}
			assert.Empty(t, errors, "No errors should occur during config detection")

			// Verify result
			require.NotNil(t, result)
			assert.Equal(t, tt.wantPilotScopeGatewayToNamespace, result.PilotScopeGatewayToNamespace)
		})
	}
}

func TestClient_convertVirtualService(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name           string
		virtualService *istionetworkingv1beta1.VirtualService
		wantHosts      []string
		wantGateways   []string
		wantExportTo   []string
	}{
		{
			name: "all fields specified",
			virtualService: &istionetworkingv1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vs",
					Namespace: "default",
				},
				Spec: istioapi.VirtualService{
					Hosts:    []string{"bookinfo.com", "reviews.bookinfo.com"},
					Gateways: []string{"bookinfo-gateway", "mesh"},
					ExportTo: []string{".", "production"},
				},
			},
			wantHosts:    []string{"bookinfo.com", "reviews.bookinfo.com"},
			wantGateways: []string{"bookinfo-gateway", "mesh"},
			wantExportTo: []string{".", "production"},
		},
		{
			name: "hosts only - defaults for gateways and exportTo",
			virtualService: &istionetworkingv1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vs-hosts-only",
					Namespace: "default",
				},
				Spec: istioapi.VirtualService{
					Hosts: []string{"api.example.com"},
				},
			},
			wantHosts:    []string{"api.example.com"},
			wantGateways: []string{"mesh"}, // default
			wantExportTo: []string{"*"},    // default
		},
		{
			name: "empty slices should get defaults",
			virtualService: &istionetworkingv1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vs-empty",
					Namespace: "default",
				},
				Spec: istioapi.VirtualService{
					Hosts:    []string{"service.local"},
					Gateways: []string{}, // empty slice
					ExportTo: []string{}, // empty slice
				},
			},
			wantHosts:    []string{"service.local"},
			wantGateways: []string{"mesh"}, // default for empty slice
			wantExportTo: []string{"*"},    // default for empty slice
		},
		{
			name: "nil slices should get defaults",
			virtualService: &istionetworkingv1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vs-nil",
					Namespace: "default",
				},
				Spec: istioapi.VirtualService{
					Hosts: []string{"another.service"},
					// Gateways and ExportTo are nil
				},
			},
			wantHosts:    []string{"another.service"},
			wantGateways: []string{"mesh"}, // default for nil
			wantExportTo: []string{"*"},    // default for nil
		},
		{
			name: "custom gateways with default exportTo",
			virtualService: &istionetworkingv1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vs-custom-gw",
					Namespace: "istio-system",
				},
				Spec: istioapi.VirtualService{
					Hosts:    []string{"*.example.com"},
					Gateways: []string{"custom-gateway", "another-gateway"},
					// ExportTo is nil - should get default
				},
			},
			wantHosts:    []string{"*.example.com"},
			wantGateways: []string{"custom-gateway", "another-gateway"},
			wantExportTo: []string{"*"}, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertVirtualService(tt.virtualService)

			require.NoError(t, err)
			assert.Equal(t, tt.virtualService.Name, result.Name)
			assert.Equal(t, tt.virtualService.Namespace, result.Namespace)
			assert.Equal(t, tt.wantHosts, result.Hosts)
			assert.Equal(t, tt.wantGateways, result.Gateways)
			assert.Equal(t, tt.wantExportTo, result.ExportTo)
			assert.NotEmpty(t, result.RawConfig)
		})
	}
}

func TestClient_convertSidecar(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name                 string
		sidecar              *istionetworkingv1beta1.Sidecar
		wantName             string
		wantNamespace        string
		wantWorkloadSelector *typesv1alpha1.WorkloadSelector
	}{
		{
			name: "sidecar with workload selector",
			sidecar: &istionetworkingv1beta1.Sidecar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sidecar",
					Namespace: "production",
				},
				Spec: istioapi.Sidecar{
					WorkloadSelector: &istioapi.WorkloadSelector{
						Labels: map[string]string{
							"app":     "reviews",
							"version": "v2",
							"tier":    "backend",
						},
					},
					Ingress: []*istioapi.IstioIngressListener{
						{
							Port: &istioapi.SidecarPort{
								Number:   9080,
								Protocol: "HTTP",
								Name:     "http",
							},
							DefaultEndpoint: "127.0.0.1:8080",
						},
					},
				},
			},
			wantName:      "test-sidecar",
			wantNamespace: "production",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app":     "reviews",
					"version": "v2",
					"tier":    "backend",
				},
			},
		},
		{
			name: "sidecar without workload selector",
			sidecar: &istionetworkingv1beta1.Sidecar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sidecar-no-selector",
					Namespace: "default",
				},
				Spec: istioapi.Sidecar{
					Egress: []*istioapi.IstioEgressListener{
						{
							Hosts: []string{"./productpage.default.svc.cluster.local"},
						},
					},
				},
			},
			wantName:             "test-sidecar-no-selector",
			wantNamespace:        "default",
			wantWorkloadSelector: nil,
		},
		{
			name: "sidecar with nil workload selector",
			sidecar: &istionetworkingv1beta1.Sidecar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sidecar-nil-selector",
					Namespace: "istio-system",
				},
				Spec: istioapi.Sidecar{
					WorkloadSelector: nil,
					Egress: []*istioapi.IstioEgressListener{
						{
							Hosts: []string{"./ratings.default.svc.cluster.local"},
						},
					},
				},
			},
			wantName:             "test-sidecar-nil-selector",
			wantNamespace:        "istio-system",
			wantWorkloadSelector: nil,
		},
		{
			name: "sidecar with empty workload selector labels",
			sidecar: &istionetworkingv1beta1.Sidecar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sidecar-empty-labels",
					Namespace: "test",
				},
				Spec: istioapi.Sidecar{
					WorkloadSelector: &istioapi.WorkloadSelector{
						Labels: map[string]string{}, // empty labels
					},
					Ingress: []*istioapi.IstioIngressListener{
						{
							Port: &istioapi.SidecarPort{
								Number:   8080,
								Protocol: "HTTP",
								Name:     "http",
							},
						},
					},
				},
			},
			wantName:      "test-sidecar-empty-labels",
			wantNamespace: "test",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
		},
		{
			name: "sidecar with nil workload selector labels",
			sidecar: &istionetworkingv1beta1.Sidecar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sidecar-nil-labels",
					Namespace: "default",
				},
				Spec: istioapi.Sidecar{
					WorkloadSelector: &istioapi.WorkloadSelector{
						Labels: nil, // nil labels
					},
					Egress: []*istioapi.IstioEgressListener{
						{
							Hosts: []string{"./details.default.svc.cluster.local"},
						},
					},
				},
			},
			wantName:             "test-sidecar-nil-labels",
			wantNamespace:        "default",
			wantWorkloadSelector: nil,
		},
		{
			name: "sidecar with single label",
			sidecar: &istionetworkingv1beta1.Sidecar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sidecar-single-label",
					Namespace: "bookinfo",
				},
				Spec: istioapi.Sidecar{
					WorkloadSelector: &istioapi.WorkloadSelector{
						Labels: map[string]string{
							"app": "ratings",
						},
					},
				},
			},
			wantName:      "test-sidecar-single-label",
			wantNamespace: "bookinfo",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app": "ratings",
				},
			},
		},
		{
			name: "minimal sidecar configuration",
			sidecar: &istionetworkingv1beta1.Sidecar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "minimal-sidecar",
					Namespace: "minimal",
				},
				Spec: istioapi.Sidecar{
					// Minimal spec with no additional configuration
				},
			},
			wantName:             "minimal-sidecar",
			wantNamespace:        "minimal",
			wantWorkloadSelector: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertSidecar(tt.sidecar)

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.wantNamespace, result.Namespace)
			assert.Equal(t, tt.wantWorkloadSelector, result.WorkloadSelector)
			assert.NotEmpty(t, result.RawConfig)

			// Verify RawConfig contains valid JSON
			var spec map[string]interface{}
			err = json.Unmarshal([]byte(result.RawConfig), &spec)
			assert.NoError(t, err, "RawConfig should be valid JSON")
		})
	}
}

func TestClient_convertPeerAuthentication(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name                 string
		peerAuthentication   *istiosecurityv1beta1.PeerAuthentication
		wantName             string
		wantNamespace        string
		wantWorkloadSelector *typesv1alpha1.WorkloadSelector
	}{
		{
			name: "peer authentication with workload selector",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-peer-auth",
					Namespace: "production",
				},
				Spec: securityapi.PeerAuthentication{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"app":     "web",
							"version": "v1",
							"tier":    "frontend",
						},
					},
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_STRICT,
					},
				},
			},
			wantName:      "test-peer-auth",
			wantNamespace: "production",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app":     "web",
					"version": "v1",
					"tier":    "frontend",
				},
			},
		},
		{
			name: "peer authentication without workload selector",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-peer-auth",
					Namespace: "default",
				},
				Spec: securityapi.PeerAuthentication{
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_PERMISSIVE,
					},
				},
			},
			wantName:             "default-peer-auth",
			wantNamespace:        "default",
			wantWorkloadSelector: nil,
		},
		{
			name: "peer authentication with nil workload selector",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nil-selector-peer-auth",
					Namespace: "istio-system",
				},
				Spec: securityapi.PeerAuthentication{
					Selector: nil,
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_STRICT,
					},
				},
			},
			wantName:             "nil-selector-peer-auth",
			wantNamespace:        "istio-system",
			wantWorkloadSelector: nil,
		},
		{
			name: "peer authentication with empty workload selector labels",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-labels-peer-auth",
					Namespace: "test",
				},
				Spec: securityapi.PeerAuthentication{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{}, // empty labels
					},
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_DISABLE,
					},
				},
			},
			wantName:      "empty-labels-peer-auth",
			wantNamespace: "test",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
		},
		{
			name: "peer authentication with nil workload selector labels",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nil-labels-peer-auth",
					Namespace: "default",
				},
				Spec: securityapi.PeerAuthentication{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: nil, // nil labels
					},
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_STRICT,
					},
				},
			},
			wantName:             "nil-labels-peer-auth",
			wantNamespace:        "default",
			wantWorkloadSelector: nil,
		},
		{
			name: "peer authentication with single label",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "single-label-peer-auth",
					Namespace: "bookinfo",
				},
				Spec: securityapi.PeerAuthentication{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"app": "ratings",
						},
					},
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_PERMISSIVE,
					},
				},
			},
			wantName:      "single-label-peer-auth",
			wantNamespace: "bookinfo",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app": "ratings",
				},
			},
		},
		{
			name: "minimal peer authentication configuration",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "minimal-peer-auth",
					Namespace: "minimal",
				},
				Spec: securityapi.PeerAuthentication{
					// Minimal spec with no additional configuration
				},
			},
			wantName:             "minimal-peer-auth",
			wantNamespace:        "minimal",
			wantWorkloadSelector: nil,
		},
		{
			name: "peer authentication with port-specific mTLS",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "port-specific-peer-auth",
					Namespace: "secure",
				},
				Spec: securityapi.PeerAuthentication{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"app":      "secure-service",
							"security": "strict",
						},
					},
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_STRICT,
					},
					PortLevelMtls: map[uint32]*securityapi.PeerAuthentication_MutualTLS{
						8080: {Mode: securityapi.PeerAuthentication_MutualTLS_PERMISSIVE},
						9090: {Mode: securityapi.PeerAuthentication_MutualTLS_DISABLE},
					},
				},
			},
			wantName:      "port-specific-peer-auth",
			wantNamespace: "secure",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app":      "secure-service",
					"security": "strict",
				},
			},
		},
		{
			name: "peer authentication with complex selector",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complex-selector-peer-auth",
					Namespace: "enterprise",
				},
				Spec: securityapi.PeerAuthentication{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"app":         "payment-service",
							"version":     "v2",
							"tier":        "backend",
							"environment": "production",
							"security":    "pci-compliant",
						},
					},
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_STRICT,
					},
				},
			},
			wantName:      "complex-selector-peer-auth",
			wantNamespace: "enterprise",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app":         "payment-service",
					"version":     "v2",
					"tier":        "backend",
					"environment": "production",
					"security":    "pci-compliant",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertPeerAuthentication(tt.peerAuthentication)

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.wantNamespace, result.Namespace)
			assert.Equal(t, tt.wantWorkloadSelector, result.Selector)
			assert.NotEmpty(t, result.RawConfig)

			// Verify RawConfig contains valid JSON
			var spec map[string]interface{}
			err = json.Unmarshal([]byte(result.RawConfig), &spec)
			assert.NoError(t, err, "RawConfig should be valid JSON")
		})
	}
}

func TestClient_convertPeerAuthentication_ErrorCases(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name               string
		peerAuthentication *istiosecurityv1beta1.PeerAuthentication
		expectError        bool
	}{
		{
			name: "valid peer authentication should not error",
			peerAuthentication: &istiosecurityv1beta1.PeerAuthentication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-peer-auth",
					Namespace: "default",
				},
				Spec: securityapi.PeerAuthentication{
					Mtls: &securityapi.PeerAuthentication_MutualTLS{
						Mode: securityapi.PeerAuthentication_MutualTLS_STRICT,
					},
				},
			},
			expectError: false,
		},
		// Note: It's hard to create a scenario where JSON marshaling fails for PeerAuthentication
		// since the Istio types are well-defined, but we include this test structure for completeness
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertPeerAuthentication(tt.peerAuthentication)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestClient_convertWasmPlugin(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name                 string
		wasmPlugin           *istioextensionsv1alpha1.WasmPlugin
		wantName             string
		wantNamespace        string
		wantWorkloadSelector *typesv1alpha1.WorkloadSelector
		wantTargetRefs       []*typesv1alpha1.PolicyTargetReference
	}{
		{
			name: "wasm plugin with workload selector",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-wasm-plugin",
					Namespace: "production",
				},
				Spec: extensionsapi.WasmPlugin{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"app":     "web",
							"version": "v1",
							"tier":    "frontend",
						},
					},
					Url: "oci://docker.io/istio/wasm-plugin:v1.0.0",
				},
			},
			wantName:      "test-wasm-plugin",
			wantNamespace: "production",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app":     "web",
					"version": "v1",
					"tier":    "frontend",
				},
			},
			wantTargetRefs: nil,
		},
		{
			name: "wasm plugin without workload selector",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-wasm-plugin",
					Namespace: "default",
				},
				Spec: extensionsapi.WasmPlugin{
					Url: "oci://docker.io/istio/auth-plugin:latest",
				},
			},
			wantName:             "default-wasm-plugin",
			wantNamespace:        "default",
			wantWorkloadSelector: nil,
			wantTargetRefs:       nil,
		},
		{
			name: "wasm plugin with nil workload selector",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nil-selector-wasm-plugin",
					Namespace: "istio-system",
				},
				Spec: extensionsapi.WasmPlugin{
					Selector: nil,
					Url:      "oci://docker.io/istio/logging-plugin:v2.0.0",
				},
			},
			wantName:             "nil-selector-wasm-plugin",
			wantNamespace:        "istio-system",
			wantWorkloadSelector: nil,
			wantTargetRefs:       nil,
		},
		{
			name: "wasm plugin with empty workload selector labels",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-labels-wasm-plugin",
					Namespace: "test",
				},
				Spec: extensionsapi.WasmPlugin{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{}, // empty labels
					},
					Url: "oci://docker.io/istio/metrics-plugin:v1.5.0",
				},
			},
			wantName:      "empty-labels-wasm-plugin",
			wantNamespace: "test",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{},
			},
			wantTargetRefs: nil,
		},
		{
			name: "wasm plugin with nil workload selector labels",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nil-labels-wasm-plugin",
					Namespace: "staging",
				},
				Spec: extensionsapi.WasmPlugin{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: nil,
					},
					Url: "oci://docker.io/istio/security-plugin:latest",
				},
			},
			wantName:             "nil-labels-wasm-plugin",
			wantNamespace:        "staging",
			wantWorkloadSelector: nil,
			wantTargetRefs:       nil,
		},
		{
			name: "wasm plugin with single label",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "single-label-wasm-plugin",
					Namespace: "development",
				},
				Spec: extensionsapi.WasmPlugin{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"environment": "dev",
						},
					},
					Url: "oci://docker.io/istio/debug-plugin:dev",
				},
			},
			wantName:      "single-label-wasm-plugin",
			wantNamespace: "development",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"environment": "dev",
				},
			},
			wantTargetRefs: nil,
		},
		{
			name: "wasm plugin with target refs",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway-wasm-plugin",
					Namespace: "istio-ingress",
				},
				Spec: extensionsapi.WasmPlugin{
					TargetRefs: []*istiotype.PolicyTargetReference{
						{
							Group: "gateway.networking.k8s.io",
							Kind:  "Gateway",
							Name:  "istio-gateway",
						},
						{
							Group:     "",
							Kind:      "Service",
							Name:      "web-service",
							Namespace: "production",
						},
					},
					Url: "oci://docker.io/istio/gateway-plugin:v1.0.0",
				},
			},
			wantName:             "gateway-wasm-plugin",
			wantNamespace:        "istio-ingress",
			wantWorkloadSelector: nil,
			wantTargetRefs: []*typesv1alpha1.PolicyTargetReference{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "istio-gateway",
					Namespace: "",
				},
				{
					Group:     "",
					Kind:      "Service",
					Name:      "web-service",
					Namespace: "production",
				},
			},
		},
		{
			name: "minimal wasm plugin configuration",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "minimal-wasm-plugin",
					Namespace: "default",
				},
				Spec: extensionsapi.WasmPlugin{
					Url: "oci://docker.io/istio/minimal-plugin:latest",
				},
			},
			wantName:             "minimal-wasm-plugin",
			wantNamespace:        "default",
			wantWorkloadSelector: nil,
			wantTargetRefs:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertWasmPlugin(tt.wasmPlugin)

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.wantNamespace, result.Namespace)

			// Verify that RawConfig is valid JSON
			var spec map[string]interface{}
			err = json.Unmarshal([]byte(result.RawConfig), &spec)
			assert.NoError(t, err, "RawConfig should be valid JSON")

			if tt.wantWorkloadSelector == nil {
				assert.Nil(t, result.Selector)
			} else {
				require.NotNil(t, result.Selector)
				assert.Equal(t, tt.wantWorkloadSelector.MatchLabels, result.Selector.MatchLabels)
			}

			if tt.wantTargetRefs == nil {
				assert.Nil(t, result.TargetRefs)
			} else {
				require.Equal(t, len(tt.wantTargetRefs), len(result.TargetRefs))
				for i, expectedRef := range tt.wantTargetRefs {
					actualRef := result.TargetRefs[i]
					assert.Equal(t, expectedRef.Group, actualRef.Group)
					assert.Equal(t, expectedRef.Kind, actualRef.Kind)
					assert.Equal(t, expectedRef.Name, actualRef.Name)
					assert.Equal(t, expectedRef.Namespace, actualRef.Namespace)
				}
			}
		})
	}
}

func TestClient_convertWasmPlugin_ErrorCases(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name        string
		wasmPlugin  *istioextensionsv1alpha1.WasmPlugin
		expectError bool
	}{
		{
			name: "valid wasm plugin should not error",
			wasmPlugin: &istioextensionsv1alpha1.WasmPlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-wasm-plugin",
					Namespace: "default",
				},
				Spec: extensionsapi.WasmPlugin{
					Url: "oci://docker.io/istio/test-plugin:latest",
				},
			},
			expectError: false,
		},
		// Note: It's hard to create a scenario where JSON marshaling fails for WasmPlugin
		// since the Istio types are well-defined, but we include this test structure for completeness
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertWasmPlugin(tt.wasmPlugin)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestClient_convertServiceEntry(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name         string
		serviceEntry *istionetworkingv1beta1.ServiceEntry
		wantName     string
		wantExportTo []string
	}{
		{
			name: "service entry with all fields",
			serviceEntry: &istionetworkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httpbin",
					Namespace: "default",
				},
				Spec: istioapi.ServiceEntry{
					Hosts: []string{"httpbin.example.com"},
					Ports: []*istioapi.ServicePort{
						{
							Number:   80,
							Name:     "http",
							Protocol: "HTTP",
						},
					},
					Location:   istioapi.ServiceEntry_MESH_EXTERNAL,
					Resolution: istioapi.ServiceEntry_DNS,
					ExportTo:   []string{".", "production"},
				},
			},
			wantName:     "httpbin",
			wantExportTo: []string{".", "production"},
		},
		{
			name: "service entry with empty exportTo defaults to global",
			serviceEntry: &istionetworkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "database",
					Namespace: "production",
				},
				Spec: istioapi.ServiceEntry{
					Hosts:      []string{"database.internal"},
					Location:   istioapi.ServiceEntry_MESH_INTERNAL,
					Resolution: istioapi.ServiceEntry_DNS,
					ExportTo:   []string{},
				},
			},
			wantName:     "database",
			wantExportTo: []string{"*"},
		},
		{
			name: "service entry with nil exportTo defaults to global",
			serviceEntry: &istionetworkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-gateway",
					Namespace: "istio-system",
				},
				Spec: istioapi.ServiceEntry{
					Hosts:      []string{"api.example.com"},
					Location:   istioapi.ServiceEntry_MESH_EXTERNAL,
					Resolution: istioapi.ServiceEntry_DNS,
					ExportTo:   nil,
				},
			},
			wantName:     "api-gateway",
			wantExportTo: []string{"*"},
		},
		{
			name: "service entry with wildcard export",
			serviceEntry: &istionetworkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-service",
					Namespace: "services",
				},
				Spec: istioapi.ServiceEntry{
					Hosts:      []string{"external.service.com"},
					Location:   istioapi.ServiceEntry_MESH_EXTERNAL,
					Resolution: istioapi.ServiceEntry_DNS,
					ExportTo:   []string{"*"},
				},
			},
			wantName:     "external-service",
			wantExportTo: []string{"*"},
		},
		{
			name: "service entry with dot export (same namespace only)",
			serviceEntry: &istionetworkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "local-service",
					Namespace: "team-a",
				},
				Spec: istioapi.ServiceEntry{
					Hosts:      []string{"local.team-a.internal"},
					Location:   istioapi.ServiceEntry_MESH_INTERNAL,
					Resolution: istioapi.ServiceEntry_DNS,
					ExportTo:   []string{"."},
				},
			},
			wantName:     "local-service",
			wantExportTo: []string{"."},
		},
		{
			name: "service entry with multiple specific namespaces",
			serviceEntry: &istionetworkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shared-cache",
					Namespace: "infrastructure",
				},
				Spec: istioapi.ServiceEntry{
					Hosts:      []string{"cache.infrastructure.svc.cluster.local"},
					Location:   istioapi.ServiceEntry_MESH_INTERNAL,
					Resolution: istioapi.ServiceEntry_DNS,
					ExportTo:   []string{"frontend", "backend", "api"},
				},
			},
			wantName:     "shared-cache",
			wantExportTo: []string{"frontend", "backend", "api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertServiceEntry(tt.serviceEntry)

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.serviceEntry.Namespace, result.Namespace)
			assert.Equal(t, tt.wantExportTo, result.ExportTo)
			assert.NotEmpty(t, result.RawConfig)

			// Verify RawConfig contains valid JSON
			var spec map[string]interface{}
			err = json.Unmarshal([]byte(result.RawConfig), &spec)
			assert.NoError(t, err, "RawConfig should be valid JSON")
		})
	}
}

func TestClient_convertServiceEntry_ErrorCases(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name         string
		serviceEntry *istionetworkingv1beta1.ServiceEntry
		expectError  bool
	}{
		{
			name: "valid service entry should not error",
			serviceEntry: &istionetworkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-service-entry",
					Namespace: "default",
				},
				Spec: istioapi.ServiceEntry{
					Hosts:      []string{"example.com"},
					Location:   istioapi.ServiceEntry_MESH_EXTERNAL,
					Resolution: istioapi.ServiceEntry_DNS,
					ExportTo:   []string{"*"},
				},
			},
			expectError: false,
		},
		// Note: It's hard to create a scenario where JSON marshaling fails for ServiceEntry
		// since the Istio types are well-defined, but we include this test structure for completeness
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertServiceEntry(tt.serviceEntry)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestClient_convertAuthorizationPolicy(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name                 string
		authorizationPolicy  *istiosecurityv1beta1.AuthorizationPolicy
		wantName             string
		wantNamespace        string
		wantWorkloadSelector *typesv1alpha1.WorkloadSelector
		wantTargetRefs       []*typesv1alpha1.PolicyTargetReference
	}{
		{
			name: "authorization policy with no selector",
			authorizationPolicy: &istiosecurityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-authz-policy",
					Namespace: "default",
				},
				Spec: securityapi.AuthorizationPolicy{
					Rules: []*securityapi.Rule{
						{
							From: []*securityapi.Rule_From{
								{
									Source: &securityapi.Source{
										Principals: []string{"cluster.local/ns/default/sa/test"},
									},
								},
							},
						},
					},
				},
			},
			wantName:             "test-authz-policy",
			wantNamespace:        "default",
			wantWorkloadSelector: nil,
			wantTargetRefs:       nil,
		},
		{
			name: "authorization policy with workload selector",
			authorizationPolicy: &istiosecurityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "selector-authz-policy",
					Namespace: "production",
				},
				Spec: securityapi.AuthorizationPolicy{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"app":     "backend",
							"version": "v2",
						},
					},
					Action: securityapi.AuthorizationPolicy_ALLOW,
					Rules: []*securityapi.Rule{
						{
							To: []*securityapi.Rule_To{
								{
									Operation: &securityapi.Operation{
										Methods: []string{"GET", "POST"},
									},
								},
							},
						},
					},
				},
			},
			wantName:      "selector-authz-policy",
			wantNamespace: "production",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app":     "backend",
					"version": "v2",
				},
			},
			wantTargetRefs: nil,
		},
		{
			name: "authorization policy with target references",
			authorizationPolicy: &istiosecurityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "targetref-authz-policy",
					Namespace: "test-ns",
				},
				Spec: securityapi.AuthorizationPolicy{
					TargetRef: &istiotype.PolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "test-gateway",
					},
					Action: securityapi.AuthorizationPolicy_DENY,
					Rules: []*securityapi.Rule{
						{
							From: []*securityapi.Rule_From{
								{
									Source: &securityapi.Source{
										IpBlocks: []string{"192.168.1.0/24"},
									},
								},
							},
						},
					},
				},
			},
			wantName:             "targetref-authz-policy",
			wantNamespace:        "test-ns",
			wantWorkloadSelector: nil,
			wantTargetRefs: []*typesv1alpha1.PolicyTargetReference{
				{
					Group: "gateway.networking.k8s.io",
					Kind:  "Gateway",
					Name:  "test-gateway",
				},
			},
		},
		{
			name: "authorization policy with multiple target references",
			authorizationPolicy: &istiosecurityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-targetref-authz-policy",
					Namespace: "multi-ns",
				},
				Spec: securityapi.AuthorizationPolicy{
					TargetRefs: []*istiotype.PolicyTargetReference{
						{
							Group: "",
							Kind:  "Service",
							Name:  "backend-service",
						},
						{
							Group:     "gateway.networking.k8s.io",
							Kind:      "Gateway",
							Name:      "api-gateway",
							Namespace: "gateway-ns",
						},
					},
					Rules: []*securityapi.Rule{
						{
							When: []*securityapi.Condition{
								{
									Key:    "source.ip",
									Values: []string{"10.0.0.0/8"},
								},
							},
						},
					},
				},
			},
			wantName:             "multi-targetref-authz-policy",
			wantNamespace:        "multi-ns",
			wantWorkloadSelector: nil,
			wantTargetRefs: []*typesv1alpha1.PolicyTargetReference{
				{
					Group: "",
					Kind:  "Service",
					Name:  "backend-service",
				},
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "api-gateway",
					Namespace: "gateway-ns",
				},
			},
		},
		{
			name: "authorization policy with both selector and target refs (single targetRef)",
			authorizationPolicy: &istiosecurityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "both-selector-targetref",
					Namespace: "combined-ns",
				},
				Spec: securityapi.AuthorizationPolicy{
					Selector: &istiotype.WorkloadSelector{
						MatchLabels: map[string]string{
							"component": "frontend",
						},
					},
					TargetRef: &istiotype.PolicyTargetReference{
						Group: "networking.istio.io",
						Kind:  "ServiceEntry",
						Name:  "external-service",
					},
					Action: securityapi.AuthorizationPolicy_CUSTOM,
					Rules: []*securityapi.Rule{
						{
							To: []*securityapi.Rule_To{
								{
									Operation: &securityapi.Operation{
										Paths: []string{"/api/*"},
									},
								},
							},
						},
					},
				},
			},
			wantName:      "both-selector-targetref",
			wantNamespace: "combined-ns",
			wantWorkloadSelector: &typesv1alpha1.WorkloadSelector{
				MatchLabels: map[string]string{
					"component": "frontend",
				},
			},
			wantTargetRefs: []*typesv1alpha1.PolicyTargetReference{
				{
					Group: "networking.istio.io",
					Kind:  "ServiceEntry",
					Name:  "external-service",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertAuthorizationPolicy(tt.authorizationPolicy)

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.wantNamespace, result.Namespace)
			assert.Equal(t, tt.wantWorkloadSelector, result.Selector)
			assert.Equal(t, tt.wantTargetRefs, result.TargetRefs)

			// Verify RawConfig is valid JSON
			var jsonData interface{}
			err = json.Unmarshal([]byte(result.RawConfig), &jsonData)
			assert.NoError(t, err, "RawConfig should be valid JSON")
		})
	}
}

func TestClient_convertAuthorizationPolicy_ErrorCases(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name                string
		authorizationPolicy *istiosecurityv1beta1.AuthorizationPolicy
		expectError         bool
	}{
		{
			name: "valid authorization policy should not error",
			authorizationPolicy: &istiosecurityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-authz-policy",
					Namespace: "default",
				},
				Spec: securityapi.AuthorizationPolicy{
					Action: securityapi.AuthorizationPolicy_ALLOW,
				},
			},
			expectError: false,
		},
		// Note: It's hard to create a scenario where JSON marshaling fails for AuthorizationPolicy
		// since the Istio types are well-defined, but we include this test structure for completeness
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.convertAuthorizationPolicy(tt.authorizationPolicy)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
