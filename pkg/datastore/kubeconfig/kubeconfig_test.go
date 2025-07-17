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

package kubeconfig

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

func createTestDatastore(t *testing.T) (*datastore, context.CancelFunc) {
	fakeClient := fake.NewSimpleClientset()

	ds := &datastore{
		client: fakeClient,
		cache: &serviceCache{
			services: make(map[string]*v1alpha1.Service),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	ds.cancel = cancel

	go ds.startWatchers(ctx)

	// Give the watchers a moment to start
	time.Sleep(100 * time.Millisecond)

	return ds, cancel
}

func waitForCacheUpdate(t *testing.T, ds *datastore, expectedCount int, timeout time.Duration) {
	start := time.Now()
	for {
		ds.cache.mu.RLock()
		count := len(ds.cache.services)
		ds.cache.mu.RUnlock()

		if count >= expectedCount {
			break
		}

		if time.Since(start) > timeout {
			t.Fatalf("timeout waiting for cache to update: expected %d services, got %d", expectedCount, count)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func TestListServices(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		services       []corev1.Service
		endpointSlices []discoveryv1.EndpointSlice
		expected       int
		wantErr        bool
	}{
		{
			name:      "happy path - single service with endpoints",
			namespace: "default",
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
							discoveryv1.LabelServiceName: "test-service",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: func() *bool { b := true; return &b }(),
							},
							TargetRef: &corev1.ObjectReference{
								Kind:      "Pod",
								Name:      "test-pod",
								Namespace: "default",
							},
						},
					},
				},
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name:           "empty namespace",
			namespace:      "default",
			services:       []corev1.Service{},
			endpointSlices: []discoveryv1.EndpointSlice{},
			expected:       0,
			wantErr:        false,
		},
		{
			name:      "multiple services",
			namespace: "default",
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
						Namespace: "default",
					},
				},
			},
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-1-abc123",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "service-1",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: func() *bool { b := true; return &b }(),
							},
							TargetRef: &corev1.ObjectReference{
								Kind:      "Pod",
								Name:      "pod-1",
								Namespace: "default",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-2-def456",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "service-2",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.2"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: func() *bool { b := true; return &b }(),
							},
							TargetRef: &corev1.ObjectReference{
								Kind:      "Pod",
								Name:      "pod-2",
								Namespace: "default",
							},
						},
					},
				},
			},
			expected: 2,
			wantErr:  false,
		},
		{
			name:      "all namespaces - empty namespace parameter",
			namespace: "",
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
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-3",
						Namespace: "test-ns",
					},
				},
			},
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-1-abc123",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "service-1",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: func() *bool { b := true; return &b }(),
							},
							TargetRef: &corev1.ObjectReference{
								Kind:      "Pod",
								Name:      "pod-1",
								Namespace: "default",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-2-def456",
						Namespace: "kube-system",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "service-2",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.2"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: func() *bool { b := true; return &b }(),
							},
							TargetRef: &corev1.ObjectReference{
								Kind:      "Pod",
								Name:      "pod-2",
								Namespace: "kube-system",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-3-ghi789",
						Namespace: "test-ns",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "service-3",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{"10.0.0.3"},
							Conditions: discoveryv1.EndpointConditions{
								Ready: func() *bool { b := true; return &b }(),
							},
							TargetRef: &corev1.ObjectReference{
								Kind:      "Pod",
								Name:      "pod-3",
								Namespace: "test-ns",
							},
						},
					},
				},
			},
			expected: 3,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, cancel := createTestDatastore(t)
			defer cancel()

			// Add services to fake client
			for _, service := range tt.services {
				_, err := ds.client.CoreV1().Services(service.Namespace).Create(context.Background(), &service, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Add endpoint slices to fake client
			for _, endpointSlice := range tt.endpointSlices {
				_, err := ds.client.DiscoveryV1().EndpointSlices(endpointSlice.Namespace).Create(context.Background(), &endpointSlice, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Wait for cache to be updated
			if tt.expected > 0 {
				waitForCacheUpdate(t, ds, tt.expected, 2*time.Second)
			} else {
				time.Sleep(200 * time.Millisecond)
			}

			// Test
			result, err := ds.ListServices(context.Background(), tt.namespace)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result, tt.expected)

			// Verify service details if we have results
			if tt.expected > 0 {
				for _, service := range result {
					assert.NotEmpty(t, service.Name)
					assert.NotEmpty(t, service.Namespace)
					assert.NotNil(t, service.Instances)

					if tt.namespace != "" {
						assert.Equal(t, tt.namespace, service.Namespace)
					}
				}
			}
		})
	}
}

func TestGetService(t *testing.T) {
	tests := []struct {
		name              string
		serviceName       string
		namespace         string
		service           *corev1.Service
		endpointSlice     *discoveryv1.EndpointSlice
		expectedName      string
		expectedInstances int
		wantErr           bool
	}{
		{
			name:        "happy path - service with endpoint slice",
			serviceName: "test-service",
			namespace:   "default",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			},
			endpointSlice: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-abc123",
					Namespace: "default",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "test-service",
					},
				},
				AddressType: discoveryv1.AddressTypeIPv4,
				Endpoints: []discoveryv1.Endpoint{
					{
						Addresses: []string{"10.0.0.1"},
						Conditions: discoveryv1.EndpointConditions{
							Ready: func() *bool { b := true; return &b }(),
						},
						TargetRef: &corev1.ObjectReference{
							Kind:      "Pod",
							Name:      "test-pod",
							Namespace: "default",
						},
					},
				},
			},
			expectedName:      "test-service",
			expectedInstances: 1,
			wantErr:           false,
		},
		{
			name:        "service without endpoint slice",
			serviceName: "headless-service",
			namespace:   "default",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "headless-service",
					Namespace: "default",
				},
			},
			endpointSlice: &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "headless-service-abc123",
					Namespace: "default",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "headless-service",
					},
				},
				AddressType: discoveryv1.AddressTypeIPv4,
				Endpoints:   []discoveryv1.Endpoint{},
			},
			expectedName:      "headless-service",
			expectedInstances: 0,
			wantErr:           false,
		},
		{
			name:          "service not found",
			serviceName:   "nonexistent-service",
			namespace:     "default",
			service:       nil,
			endpointSlice: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, cancel := createTestDatastore(t)
			defer cancel()

			// Add service to fake client
			if tt.service != nil {
				_, err := ds.client.CoreV1().Services(tt.namespace).Create(context.Background(), tt.service, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Add endpoint slice to fake client
			if tt.endpointSlice != nil {
				_, err := ds.client.DiscoveryV1().EndpointSlices(tt.namespace).Create(context.Background(), tt.endpointSlice, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Wait for cache to be updated
			if tt.service != nil {
				waitForCacheUpdate(t, ds, 1, 2*time.Second)
			} else {
				time.Sleep(200 * time.Millisecond)
			}

			// Test
			serviceID := tt.namespace + ":" + tt.serviceName
			result, err := ds.GetService(context.Background(), serviceID)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.namespace, result.Namespace)
			assert.Len(t, result.Instances, tt.expectedInstances)
		})
	}
}

func TestNamespaceIsolation(t *testing.T) {
	tests := []struct {
		name                 string
		servicesPerNamespace map[string][]string
		testNamespace        string
		expectedServices     []string
	}{
		{
			name: "services in different namespaces are isolated",
			servicesPerNamespace: map[string][]string{
				"default":     {"service-1", "service-2"},
				"kube-system": {"kube-dns", "kube-proxy"},
				"test-ns":     {"test-service"},
			},
			testNamespace:    "default",
			expectedServices: []string{"service-1", "service-2"},
		},
		{
			name: "kube-system namespace isolation",
			servicesPerNamespace: map[string][]string{
				"default":     {"service-1", "service-2"},
				"kube-system": {"kube-dns", "kube-proxy"},
				"test-ns":     {"test-service"},
			},
			testNamespace:    "kube-system",
			expectedServices: []string{"kube-dns", "kube-proxy"},
		},
		{
			name: "single service namespace",
			servicesPerNamespace: map[string][]string{
				"default":     {"service-1", "service-2"},
				"kube-system": {"kube-dns", "kube-proxy"},
				"test-ns":     {"test-service"},
			},
			testNamespace:    "test-ns",
			expectedServices: []string{"test-service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, cancel := createTestDatastore(t)
			defer cancel()

			totalServices := 0
			for _, serviceNames := range tt.servicesPerNamespace {
				totalServices += len(serviceNames)
			}

			// Create all services across namespaces
			for namespace, serviceNames := range tt.servicesPerNamespace {
				for i, serviceName := range serviceNames {
					service := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      serviceName,
							Namespace: namespace,
						},
					}

					endpointSlice := &discoveryv1.EndpointSlice{
						ObjectMeta: metav1.ObjectMeta{
							Name:      serviceName + "-abc123",
							Namespace: namespace,
							Labels: map[string]string{
								discoveryv1.LabelServiceName: serviceName,
							},
						},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses: []string{fmt.Sprintf("10.%d.%d.1", len(namespace), i+1)},
								Conditions: discoveryv1.EndpointConditions{
									Ready: func() *bool { b := true; return &b }(),
								},
								TargetRef: &corev1.ObjectReference{
									Kind:      "Pod",
									Name:      fmt.Sprintf("%s-pod-%d", serviceName, i+1),
									Namespace: namespace,
								},
							},
						},
					}

					_, err := ds.client.CoreV1().Services(namespace).Create(context.Background(), service, metav1.CreateOptions{})
					require.NoError(t, err)

					_, err = ds.client.DiscoveryV1().EndpointSlices(namespace).Create(context.Background(), endpointSlice, metav1.CreateOptions{})
					require.NoError(t, err)
				}
			}

			// Wait for cache to be updated with all services
			waitForCacheUpdate(t, ds, totalServices, 5*time.Second)

			// Test that the target namespace only returns its own services
			services, err := ds.ListServices(context.Background(), tt.testNamespace)
			require.NoError(t, err)
			assert.Len(t, services, len(tt.expectedServices))

			// Verify service names match expected
			serviceNames := make([]string, len(services))
			for i, service := range services {
				serviceNames[i] = service.Name
				assert.Equal(t, tt.testNamespace, service.Namespace)
			}

			require.ElementsMatch(t, tt.expectedServices, serviceNames)
		})
	}
}

func TestSidecarDetection(t *testing.T) {
	tests := []struct {
		name               string
		pod                *corev1.Pod
		expectedHasSidecar bool
	}{
		{
			name: "pod with istio-proxy sidecar by name",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.18.0",
						},
					},
				},
			},
			expectedHasSidecar: true,
		},
		{
			name: "pod with istio sidecar by image",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
						{
							Name:  "sidecar",
							Image: "istio/proxyv2:1.20.0",
						},
					},
				},
			},
			expectedHasSidecar: true,
		},
		{
			name: "pod without istio proxy sidecar",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:latest",
						},
						{
							Name:  "logger",
							Image: "fluent/fluent-bit:latest",
						},
					},
				},
			},
			expectedHasSidecar: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, cancel := createTestDatastore(t)
			defer cancel()

			// Create service and endpoints that reference the pod
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			}

			endpointSlice := &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-abc123",
					Namespace: "default",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "test-service",
					},
				},
				AddressType: discoveryv1.AddressTypeIPv4,
				Endpoints: []discoveryv1.Endpoint{
					{
						Addresses: []string{"10.0.0.1"},
						Conditions: discoveryv1.EndpointConditions{
							Ready: func() *bool { b := true; return &b }(),
						},
						TargetRef: &corev1.ObjectReference{
							Kind:      "Pod",
							Name:      tt.pod.Name,
							Namespace: tt.pod.Namespace,
						},
					},
				},
			}

			// Create resources in the fake client
			_, err := ds.client.CoreV1().Services("default").Create(context.Background(), service, metav1.CreateOptions{})
			require.NoError(t, err)

			_, err = ds.client.DiscoveryV1().EndpointSlices("default").Create(context.Background(), endpointSlice, metav1.CreateOptions{})
			require.NoError(t, err)

			_, err = ds.client.CoreV1().Pods("default").Create(context.Background(), tt.pod, metav1.CreateOptions{})
			require.NoError(t, err)

			// Wait for cache to be updated
			waitForCacheUpdate(t, ds, 1, 2*time.Second)

			// Give pod watcher time to update sidecar status
			time.Sleep(500 * time.Millisecond)

			// Get the service and check sidecar status
			result, err := ds.GetService(context.Background(), "default:test-service")
			require.NoError(t, err)
			require.Len(t, result.Instances, 1)

			instance := result.Instances[0]
			assert.Equal(t, "10.0.0.1", instance.Ip)
			assert.Equal(t, tt.pod.Name, instance.Pod)
			assert.Equal(t, "default", instance.Namespace)
			assert.Equal(t, tt.expectedHasSidecar, instance.IsEnvoyPresent)
		})
	}
}

func TestCacheConsistency(t *testing.T) {
	ds, cancel := createTestDatastore(t)
	defer cancel()

	// Create a service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	}

	_, err := ds.client.CoreV1().Services("default").Create(context.Background(), service, metav1.CreateOptions{})
	require.NoError(t, err)

	// Wait for service to appear in cache
	waitForCacheUpdate(t, ds, 1, 2*time.Second)

	// Verify GetService and ListServices return consistent data
	serviceFromGet, err := ds.GetService(context.Background(), "default:test-service")
	require.NoError(t, err)

	servicesFromList, err := ds.ListServices(context.Background(), "default")
	require.NoError(t, err)
	require.Len(t, servicesFromList, 1)

	serviceFromList := servicesFromList[0]

	// Both should return equivalent data
	assert.Equal(t, serviceFromGet.Id, serviceFromList.Id)
	assert.Equal(t, serviceFromGet.Name, serviceFromList.Name)
	assert.Equal(t, serviceFromGet.Namespace, serviceFromList.Namespace)
	assert.Equal(t, len(serviceFromGet.Instances), len(serviceFromList.Instances))
}

func TestGetServiceInstance(t *testing.T) {
	tests := []struct {
		name         string
		serviceID    string
		instanceID   string
		pod          *corev1.Pod
		expectError  bool
		expectedPod  string
		expectedIP   string
		expectedNode string
	}{
		{
			name:       "valid instance with envoy",
			serviceID:  "demo:frontend",
			instanceID: "kind-demo:demo:frontend-abc123",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "frontend-abc123",
					Namespace: "demo",
					Labels: map[string]string{
						"app": "frontend",
					},
					Annotations: map[string]string{
						"deployment.kubernetes.io/revision": "1",
					},
					CreationTimestamp: metav1.Time{Time: time.Now()},
				},
				Spec: corev1.PodSpec{
					NodeName: "node-1",
					Containers: []corev1.Container{
						{
							Name:  "frontend",
							Image: "nginx:latest",
						},
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.20.0",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "10.244.1.4",
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "frontend",
							Ready:        true,
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
						{
							Name:         "istio-proxy",
							Ready:        true,
							RestartCount: 0,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			expectError:  false,
			expectedPod:  "frontend-abc123",
			expectedIP:   "10.244.1.4",
			expectedNode: "node-1",
		},
		{
			name:       "valid instance without envoy",
			serviceID:  "demo:backend",
			instanceID: "kind-demo:demo:backend-def456",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "backend-def456",
					Namespace: "demo",
					Labels: map[string]string{
						"app": "backend",
					},
					CreationTimestamp: metav1.Time{Time: time.Now()},
				},
				Spec: corev1.PodSpec{
					NodeName: "node-2",
					Containers: []corev1.Container{
						{
							Name:  "backend",
							Image: "backend:v1.0",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					PodIP: "10.244.1.5",
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:         "backend",
							Ready:        true,
							RestartCount: 2,
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			expectError:  false,
			expectedPod:  "backend-def456",
			expectedIP:   "10.244.1.5",
			expectedNode: "node-2",
		},
		{
			name:        "invalid service ID format",
			serviceID:   "invalid-service-id",
			instanceID:  "cluster:namespace:pod",
			expectError: true,
		},
		{
			name:        "invalid instance ID format",
			serviceID:   "demo:frontend",
			instanceID:  "invalid-instance-id",
			expectError: true,
		},
		{
			name:        "pod not found",
			serviceID:   "demo:frontend",
			instanceID:  "kind-demo:demo:nonexistent-pod",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake Kubernetes client
			clientset := fake.NewSimpleClientset()

			// Add the pod to the fake client if provided
			if tt.pod != nil {
				_, err := clientset.CoreV1().Pods(tt.pod.Namespace).Create(
					context.Background(), tt.pod, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Create datastore with fake client
			ds := &datastore{
				client:      clientset,
				clusterName: "kind-demo",
				cache: &serviceCache{
					services: make(map[string]*v1alpha1.Service),
				},
			}

			// Call GetServiceInstance
			ctx := context.Background()
			result, err := ds.GetServiceInstance(ctx, tt.serviceID, tt.instanceID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify the result
			assert.Equal(t, tt.instanceID, result.InstanceId)
			assert.Equal(t, tt.expectedPod, result.Pod)
			assert.Equal(t, tt.expectedIP, result.Ip)
			assert.Equal(t, tt.expectedNode, result.NodeName)
			assert.Equal(t, "kind-demo", result.ClusterName)

			// Verify envoy detection
			if tt.pod != nil {
				hasEnvoy := false
				for _, container := range tt.pod.Spec.Containers {
					if container.Name == "istio-proxy" {
						hasEnvoy = true
						break
					}
				}
				assert.Equal(t, hasEnvoy, result.IsEnvoyPresent)
			}

			// Verify containers are populated
			if tt.pod != nil {
				assert.Len(t, result.Containers, len(tt.pod.Spec.Containers))
				for i, container := range result.Containers {
					expectedContainer := tt.pod.Spec.Containers[i]
					assert.Equal(t, expectedContainer.Name, container.Name)
					assert.Equal(t, expectedContainer.Image, container.Image)
				}
			}
		})
	}
}
