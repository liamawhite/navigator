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
	"sync"
	"testing"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istioapi "istio.io/api/networking/v1alpha3"
	istiotype "istio.io/api/type/v1beta1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
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
		wantSubsets          []*v1alpha1.DestinationRuleSubset
		wantExportTo         []string
		wantWorkloadSelector *v1alpha1.WorkloadSelector
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
			wantSubsets: []*v1alpha1.DestinationRuleSubset{
				{Name: "v1", Labels: map[string]string{"version": "v1"}},
				{Name: "v2", Labels: map[string]string{"version": "v2", "app": "reviews"}},
			},
			wantExportTo: []string{".", "production"},
			wantWorkloadSelector: &v1alpha1.WorkloadSelector{
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
			wantSubsets:          []*v1alpha1.DestinationRuleSubset{},
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
			wantSubsets:          []*v1alpha1.DestinationRuleSubset{},
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
			wantSubsets: []*v1alpha1.DestinationRuleSubset{
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
			wantSubsets:          []*v1alpha1.DestinationRuleSubset{},
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
			wantSubsets: []*v1alpha1.DestinationRuleSubset{
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
			assert.NotEmpty(t, result.RawSpec)
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
			assert.NotEmpty(t, result.RawSpec)
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
			var result *v1alpha1.IstioControlPlaneConfig
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
			assert.NotEmpty(t, result.RawSpec)
		})
	}
}
