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




