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
	"testing"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClient_isEnvoyContainer(t *testing.T) {
	tests := []struct {
		name      string
		container corev1.Container
		want      bool
	}{
		{
			name: "envoy container by name",
			container: corev1.Container{
				Name:  "envoy",
				Image: "nginx:latest",
			},
			want: true,
		},
		{
			name: "proxy container by name",
			container: corev1.Container{
				Name:  "istio-proxy",
				Image: "nginx:latest",
			},
			want: true,
		},
		{
			name: "sidecar container by name",
			container: corev1.Container{
				Name:  "sidecar",
				Image: "nginx:latest",
			},
			want: true,
		},
		{
			name: "envoy container by image",
			container: corev1.Container{
				Name:  "app",
				Image: "envoyproxy/envoy:v1.20.0",
			},
			want: true,
		},
		{
			name: "istio proxy container by image",
			container: corev1.Container{
				Name:  "app",
				Image: "docker.io/istio/proxyv2:1.10.0",
			},
			want: true,
		},
		{
			name: "istio-proxy container by image",
			container: corev1.Container{
				Name:  "app",
				Image: "gcr.io/istio-release/istio-proxy:1.10.0",
			},
			want: true,
		},
		{
			name: "regular container",
			container: corev1.Container{
				Name:  "app",
				Image: "nginx:latest",
			},
			want: false,
		},
		{
			name: "case insensitive name match",
			container: corev1.Container{
				Name:  "ENVOY-PROXY",
				Image: "nginx:latest",
			},
			want: true,
		},
		{
			name: "case insensitive image match",
			container: corev1.Container{
				Name:  "app",
				Image: "ENVOYPROXY/ENVOY:latest",
			},
			want: true,
		},
	}

	k8sClient := &Client{
		logger: logging.For("test"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := k8sClient.isEnvoyContainer(tt.container)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_convertEndpointSlicesToInstancesWithMaps(t *testing.T) {
	tests := []struct {
		name           string
		endpointSlices []discoveryv1.EndpointSlice
		pods           []corev1.Pod
		want           []*v1alpha1.ServiceInstance
		wantErr        bool
	}{
		{
			name: "single endpoint with envoy sidecar",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-abc123",
						Namespace: "default",
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: boolPtr(true),
							},
							TargetRef: &corev1.ObjectReference{
								Kind: "Pod",
								Name: "test-pod-1",
							},
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app", Image: "nginx:latest"},
							{Name: "envoy", Image: "envoyproxy/envoy:v1.20.0"},
						},
					},
				},
			},
			want: []*v1alpha1.ServiceInstance{
				{
					Ip:           "10.0.0.1",
					PodName:      "test-pod-1",
					EnvoyPresent: true,
				},
			},
			wantErr: false,
		},
		{
			name: "single endpoint without envoy sidecar",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-abc123",
						Namespace: "default",
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: boolPtr(true),
							},
							TargetRef: &corev1.ObjectReference{
								Kind: "Pod",
								Name: "test-pod-1",
							},
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app", Image: "nginx:latest"},
						},
					},
				},
			},
			want: []*v1alpha1.ServiceInstance{
				{
					Ip:           "10.0.0.1",
					PodName:      "test-pod-1",
					EnvoyPresent: false,
				},
			},
			wantErr: false,
		},
		{
			name: "multiple endpoints with mixed envoy presence",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-abc123",
						Namespace: "default",
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: boolPtr(true),
							},
							TargetRef: &corev1.ObjectReference{
								Kind: "Pod",
								Name: "test-pod-1",
							},
						},
						{
							Addresses: []string{"10.0.0.2"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: boolPtr(true),
							},
							TargetRef: &corev1.ObjectReference{
								Kind: "Pod",
								Name: "test-pod-2",
							},
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app", Image: "nginx:latest"},
							{Name: "envoy", Image: "envoyproxy/envoy:v1.20.0"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-2",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app", Image: "nginx:latest"},
						},
					},
				},
			},
			want: []*v1alpha1.ServiceInstance{
				{
					Ip:           "10.0.0.1",
					PodName:      "test-pod-1",
					EnvoyPresent: true,
				},
				{
					Ip:           "10.0.0.2",
					PodName:      "test-pod-2",
					EnvoyPresent: false,
				},
			},
			wantErr: false,
		},
		{
			name: "not ready endpoint filtered out",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-abc123",
						Namespace: "default",
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: boolPtr(false),
							},
							TargetRef: &corev1.ObjectReference{
								Kind: "Pod",
								Name: "test-pod-1",
							},
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app", Image: "nginx:latest"},
						},
					},
				},
			},
			want:    []*v1alpha1.ServiceInstance{},
			wantErr: false,
		},
		{
			name: "endpoint without target ref",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-abc123",
						Namespace: "default",
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: boolPtr(true),
							},
							// No TargetRef
						},
					},
				},
			},
			pods: []corev1.Pod{},
			want: []*v1alpha1.ServiceInstance{
				{
					Ip:           "10.0.0.1",
					PodName:      "",
					EnvoyPresent: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset
			clientset := fake.NewSimpleClientset()

			// Add pods to fake clientset
			for _, pod := range tt.pods {
				_, _ = clientset.CoreV1().Pods(pod.Namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
			}

			k8sClient := &Client{
				clientset: clientset,
				logger:    logging.For("test"),
			}

			// Build pod map from test data
			podMap := k8sClient.buildPodMap(tt.pods)

			got := k8sClient.convertEndpointSlicesToInstancesWithMaps(tt.endpointSlices, podMap)

			// No error expected since we removed the context parameter
			if tt.wantErr {
				t.Errorf("Test expects error but method no longer returns error")
				return
			}

			assert.Len(t, got, len(tt.want))

			for i, instance := range got {
				assert.Equal(t, tt.want[i].Ip, instance.Ip)
				assert.Equal(t, tt.want[i].PodName, instance.PodName)
				assert.Equal(t, tt.want[i].EnvoyPresent, instance.EnvoyPresent)
			}
		})
	}
}

func TestClient_convertServiceType(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name        string
		serviceType corev1.ServiceType
		expected    types.ServiceType
	}{
		{
			name:        "ClusterIP service type",
			serviceType: corev1.ServiceTypeClusterIP,
			expected:    types.ServiceType_CLUSTER_IP,
		},
		{
			name:        "NodePort service type",
			serviceType: corev1.ServiceTypeNodePort,
			expected:    types.ServiceType_NODE_PORT,
		},
		{
			name:        "LoadBalancer service type",
			serviceType: corev1.ServiceTypeLoadBalancer,
			expected:    types.ServiceType_LOAD_BALANCER,
		},
		{
			name:        "ExternalName service type",
			serviceType: corev1.ServiceTypeExternalName,
			expected:    types.ServiceType_EXTERNAL_NAME,
		},
		{
			name:        "Unknown service type",
			serviceType: corev1.ServiceType("Unknown"),
			expected:    types.ServiceType_SERVICE_TYPE_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.convertServiceType(tt.serviceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_extractExternalIP(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		service  *corev1.Service
		expected string
	}{
		{
			name: "LoadBalancer with ingress IP",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "203.0.113.1"},
						},
					},
				},
			},
			expected: "203.0.113.1",
		},
		{
			name: "LoadBalancer with hostname (ignored)",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{Hostname: "example.com"},
						},
					},
				},
			},
			expected: "",
		},
		{
			name: "Service with external IPs",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					ExternalIPs: []string{"198.51.100.1", "198.51.100.2"},
				},
			},
			expected: "198.51.100.1", // Returns first external IP
		},
		{
			name: "LoadBalancer with both ingress IP and external IPs (ingress priority)",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:        corev1.ServiceTypeLoadBalancer,
					ExternalIPs: []string{"198.51.100.1"},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "203.0.113.1"},
						},
					},
				},
			},
			expected: "203.0.113.1", // LoadBalancer ingress takes priority
		},
		{
			name: "ClusterIP service with no external access",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "10.96.0.1",
				},
			},
			expected: "",
		},
		{
			name: "NodePort service with no external IPs",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeNodePort,
				},
			},
			expected: "",
		},
		{
			name: "Service with empty external IPs slice",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					ExternalIPs: []string{},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.extractExternalIP(tt.service)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_determineProxyMode(t *testing.T) {
	client := &Client{
		logger: logging.For("test"),
	}

	tests := []struct {
		name        string
		pod         *corev1.Pod
		expected    types.ProxyMode
		description string
	}{
		{
			name:        "nil pod",
			pod:         nil,
			expected:    types.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Nil pod should return unknown mode",
		},
		{
			name: "pod with nil labels",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: nil,
				},
			},
			expected:    types.ProxyMode_UNKNOWN_PROXY_MODE,
			description: "Pod with nil labels should return unknown mode",
		},
		{
			name: "waypoint proxy with istio.io/waypoint-for label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio.io/waypoint-for": "namespace",
					},
				},
			},
			expected:    types.ProxyMode_SIDECAR,
			description: "Waypoint proxy should be classified as sidecar to avoid false gateway detection",
		},
		{
			name: "gateway with dataplane-mode=none (should not be waypoint)",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio.io/dataplane-mode": "none",
						"app":                     "istio-ingressgateway",
					},
				},
			},
			expected:    types.ProxyMode_ROUTER,
			description: "Gateway with dataplane-mode=none should be router, not sidecar",
		},
		{
			name: "gateway with istio.io/gateway-name label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio.io/gateway-name": "my-gateway",
					},
				},
			},
			expected:    types.ProxyMode_ROUTER,
			description: "Istio gateway label should indicate router mode",
		},
		{
			name: "gateway with kubernetes gateway.networking.k8s.io/gateway-name label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"gateway.networking.k8s.io/gateway-name": "k8s-gateway",
					},
				},
			},
			expected:    types.ProxyMode_ROUTER,
			description: "Kubernetes Gateway API label should indicate router mode",
		},
		{
			name: "istio ingress gateway with app label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "istio-ingressgateway",
					},
				},
			},
			expected:    types.ProxyMode_ROUTER,
			description: "Istio ingress gateway app label should indicate router mode",
		},
		{
			name: "istio egress gateway with app label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "istio-egressgateway",
					},
				},
			},
			expected:    types.ProxyMode_ROUTER,
			description: "Istio egress gateway app label should indicate router mode",
		},
		{
			name: "legacy istio ingress gateway label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio": "ingressgateway",
					},
				},
			},
			expected:    types.ProxyMode_ROUTER,
			description: "Legacy istio label should indicate router mode",
		},
		{
			name: "envoy container with router arg",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.20.0",
							Args:  []string{"proxy", "router", "--domain", "cluster.local"},
						},
					},
				},
			},
			expected:    types.ProxyMode_ROUTER,
			description: "Envoy container with router argument should indicate router mode",
		},
		{
			name: "envoy container with sidecar arg",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.20.0",
							Args:  []string{"proxy", "sidecar", "--domain", "cluster.local"},
						},
					},
				},
			},
			expected:    types.ProxyMode_SIDECAR,
			description: "Envoy container with sidecar argument should indicate sidecar mode",
		},
		{
			name: "envoy container with insufficient args",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.20.0",
							Args:  []string{"proxy"},
						},
					},
				},
			},
			expected:    types.ProxyMode_SIDECAR,
			description: "Envoy container with insufficient args should fall back to general Envoy detection",
		},
		{
			name: "non-envoy container with router arg",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
							Args:  []string{"nginx", "router"},
						},
					},
				},
			},
			expected:    types.ProxyMode_NONE,
			description: "Non-Envoy container should not affect proxy mode detection",
		},
		{
			name: "envoy container with unknown arg",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.20.0",
							Args:  []string{"proxy", "unknown-mode"},
						},
					},
				},
			},
			expected:    types.ProxyMode_SIDECAR,
			description: "Envoy container with unknown args should fall back to general Envoy detection",
		},
		{
			name: "pod without envoy containers",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
					},
				},
			},
			expected:    types.ProxyMode_NONE,
			description: "Pod without Envoy containers should return none mode",
		},
		{
			name: "multiple containers with envoy sidecar",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.20.0",
						},
					},
				},
			},
			expected:    types.ProxyMode_SIDECAR,
			description: "Pod with Envoy sidecar should return sidecar mode",
		},
		{
			name: "waypoint takes precedence over gateway label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"istio.io/gateway-name": "my-gateway",
						"istio.io/waypoint-for": "namespace",
					},
				},
			},
			expected:    types.ProxyMode_SIDECAR,
			description: "Waypoint detection should take precedence to avoid false gateway detection",
		},
		{
			name: "empty labels map",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
					},
				},
			},
			expected:    types.ProxyMode_NONE,
			description: "Pod with empty labels and no Envoy should return none mode",
		},
		{
			name: "container args with router but not envoy",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
							Args:  []string{"nginx", "router"},
						},
					},
				},
			},
			expected:    types.ProxyMode_NONE,
			description: "Non-Envoy container with router args should not affect detection",
		},
		{
			name: "envoy container in init containers",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "istio-init",
							Image: "istio/proxyv2:1.20.0",
							Args:  []string{"proxy", "router"},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
					},
				},
			},
			expected:    types.ProxyMode_SIDECAR,
			description: "Envoy in init containers is detected by hasEnvoySidecarInPod",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := client.determineProxyMode(test.pod)
			assert.Equal(t, test.expected, result, test.description)
		})
	}
}

func TestClient_convertServiceWithMaps_WithIPs(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		service  *corev1.Service
		expected struct {
			serviceType types.ServiceType
			clusterIP   string
			externalIP  string
		}
	}{
		{
			name: "ClusterIP service with cluster IP",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "10.96.0.1",
				},
			},
			expected: struct {
				serviceType types.ServiceType
				clusterIP   string
				externalIP  string
			}{
				serviceType: types.ServiceType_CLUSTER_IP,
				clusterIP:   "10.96.0.1",
				externalIP:  "",
			},
		},
		{
			name: "LoadBalancer service with both cluster and external IPs",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb-service",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeLoadBalancer,
					ClusterIP: "10.96.0.2",
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "203.0.113.1"},
						},
					},
				},
			},
			expected: struct {
				serviceType types.ServiceType
				clusterIP   string
				externalIP  string
			}{
				serviceType: types.ServiceType_LOAD_BALANCER,
				clusterIP:   "10.96.0.2",
				externalIP:  "203.0.113.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Empty maps for endpoint slices and pods since we're only testing service IP extraction
			endpointSlicesByService := make(map[string][]discoveryv1.EndpointSlice)
			podsByName := make(map[string]*corev1.Pod)

			result := client.convertServiceWithMaps(tt.service, endpointSlicesByService, podsByName)

			assert.Equal(t, tt.service.Name, result.Name)
			assert.Equal(t, tt.service.Namespace, result.Namespace)
			assert.Equal(t, tt.expected.serviceType, result.ServiceType)
			assert.Equal(t, tt.expected.clusterIP, result.ClusterIp)
			assert.Equal(t, tt.expected.externalIP, result.ExternalIp)
		})
	}
}
