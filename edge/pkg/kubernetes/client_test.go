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
	"errors"
	"testing"

	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istioapi "istio.io/api/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}

func TestClient_GetClusterState(t *testing.T) {
	tests := []struct {
		name           string
		services       []corev1.Service
		endpointSlices []discoveryv1.EndpointSlice
		pods           []corev1.Pod
		wantServices   int
		wantErr        bool
	}{
		{
			name: "single service with endpoints",
			services: []corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service",
						Namespace: "default",
					},
				},
			},
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-abc123",
						Namespace: "default",
						Labels: map[string]string{
							"kubernetes.io/service-name": "test-service",
						},
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
			wantServices: 1,
			wantErr:      false,
		},
		{
			name: "multiple services",
			services: []corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-1",
						Namespace: "default",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-2",
						Namespace: "kube-system",
					},
				},
			},
			endpointSlices: []discoveryv1.EndpointSlice{},
			pods:           []corev1.Pod{},
			wantServices:   2,
			wantErr:        false,
		},
		{
			name:           "no services",
			services:       []corev1.Service{},
			endpointSlices: []discoveryv1.EndpointSlice{},
			pods:           []corev1.Pod{},
			wantServices:   0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset
			clientset := fake.NewSimpleClientset()

			// Add objects to fake clientset
			for _, svc := range tt.services {
				_, _ = clientset.CoreV1().Services(svc.Namespace).Create(context.TODO(), &svc, metav1.CreateOptions{})
			}
			for _, eps := range tt.endpointSlices {
				_, _ = clientset.DiscoveryV1().EndpointSlices(eps.Namespace).Create(context.TODO(), &eps, metav1.CreateOptions{})
			}
			for _, pod := range tt.pods {
				_, _ = clientset.CoreV1().Pods(pod.Namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
			}

			k8sClient := &Client{
				clientset:   clientset,
				istioClient: istiofake.NewSimpleClientset(),
				logger:      logging.For("test"),
			}

			got, err := k8sClient.GetClusterState(context.TODO())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Len(t, got.Services, tt.wantServices)
			}
		})
	}
}

func TestClient_mergeErrors(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	tests := []struct {
		name   string
		errors []error
		want   string
	}{
		{
			name:   "no errors",
			errors: nil,
			want:   "",
		},
		{
			name:   "single error",
			errors: []error{errors.New("first error")},
			want:   "first error",
		},
		{
			name: "multiple errors",
			errors: []error{
				errors.New("first error"),
				errors.New("second error"),
				errors.New("third error"),
			},
			want: "multiple errors occurred (3 total): first error; second error; third error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.mergeErrors(tt.errors)
			if tt.want == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.want, err.Error())
			}
		})
	}
}

func TestClient_GetClusterStateWithIstio(t *testing.T) {
	// Create test Kubernetes resources
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	}

	// Create test Istio resources
	dr := &istionetworkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dr",
			Namespace: "default",
		},
		Spec: istioapi.DestinationRule{
			Host: "test-service",
		},
	}

	// Create fake clients
	k8sClient := fake.NewSimpleClientset(&service)
	istioClient := istiofake.NewSimpleClientset(dr)

	client := &Client{
		clientset:   k8sClient,
		istioClient: istioClient,
		logger:      logging.For("test"),
	}

	result, err := client.GetClusterState(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify Kubernetes resources
	assert.Len(t, result.Services, 1)
	assert.Equal(t, "test-service", result.Services[0].Name)

	// Verify Istio resources
	assert.Len(t, result.DestinationRules, 1)
	assert.Equal(t, "test-dr", result.DestinationRules[0].Name)
	assert.Contains(t, result.DestinationRules[0].RawSpec, "test-service")
}