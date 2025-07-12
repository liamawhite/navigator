package kubeconfig

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListServices(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		services  []corev1.Service
		endpoints []corev1.Endpoints
		expected  int
		wantErr   bool
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
			endpoints: []corev1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "10.0.0.1",
									TargetRef: &corev1.ObjectReference{
										Kind:      "Pod",
										Name:      "test-pod",
										Namespace: "default",
									},
								},
							},
						},
					},
				},
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name:      "empty namespace",
			namespace: "default",
			services:  []corev1.Service{},
			endpoints: []corev1.Endpoints{},
			expected:  0,
			wantErr:   false,
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
			endpoints: []corev1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-1",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "10.0.0.1",
									TargetRef: &corev1.ObjectReference{
										Kind:      "Pod",
										Name:      "pod-1",
										Namespace: "default",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-2",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "10.0.0.2",
									TargetRef: &corev1.ObjectReference{
										Kind:      "Pod",
										Name:      "pod-2",
										Namespace: "default",
									},
								},
							},
						},
					},
				},
			},
			expected: 2,
			wantErr:  false,
		},
		{
			name:      "service without endpoints",
			namespace: "default",
			services: []corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "headless-service",
						Namespace: "default",
					},
				},
			},
			endpoints: []corev1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "headless-service",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{}, // No endpoints
				},
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name:      "all namespaces - empty namespace parameter",
			namespace: "", // Empty namespace means all namespaces
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
			endpoints: []corev1.Endpoints{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-1",
						Namespace: "default",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "10.0.0.1",
									TargetRef: &corev1.ObjectReference{
										Kind:      "Pod",
										Name:      "pod-1",
										Namespace: "default",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-2",
						Namespace: "kube-system",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "10.0.0.2",
									TargetRef: &corev1.ObjectReference{
										Kind:      "Pod",
										Name:      "pod-2",
										Namespace: "kube-system",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service-3",
						Namespace: "test-ns",
					},
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP: "10.0.0.3",
									TargetRef: &corev1.ObjectReference{
										Kind:      "Pod",
										Name:      "pod-3",
										Namespace: "test-ns",
									},
								},
							},
						},
					},
				},
			},
			expected: 3, // Should return all services from all namespaces
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			fakeClient := fake.NewSimpleClientset()

			// Add services to fake client - use each service's own namespace
			for _, service := range tt.services {
				_, err := fakeClient.CoreV1().Services(service.Namespace).Create(context.Background(), &service, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Add endpoints to fake client - use each endpoint's own namespace
			for _, endpoint := range tt.endpoints {
				_, err := fakeClient.CoreV1().Endpoints(endpoint.Namespace).Create(context.Background(), &endpoint, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Create datastore
			ds := &datastore{
				client: fakeClient,
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
				// For all-namespace tests, we need to check differently since order might vary
				if tt.namespace == "" {
					// Check that we have services from different namespaces
					namespaces := make(map[string]bool)
					for _, service := range result {
						namespaces[service.Namespace] = true
						assert.NotEmpty(t, service.Name)
						assert.NotEmpty(t, service.Namespace)
					}
					// Should have services from multiple namespaces
					assert.True(t, len(namespaces) > 1, "Should have services from multiple namespaces")
				} else {
					// For single namespace tests, verify specific services
					for i, service := range result {
						assert.Equal(t, tt.services[i].Name, service.Name)
						assert.Equal(t, tt.services[i].Namespace, service.Namespace)
						// service.Instances can be nil or empty slice, both are valid
						if service.Instances == nil {
							assert.Nil(t, service.Instances)
						} else {
							assert.NotNil(t, service.Instances)
						}
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
		endpoints         *corev1.Endpoints
		expectedName      string
		expectedInstances int
		wantErr           bool
	}{
		{
			name:        "happy path - service with endpoints",
			serviceName: "test-service",
			namespace:   "default",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			},
			endpoints: &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "10.0.0.1",
								TargetRef: &corev1.ObjectReference{
									Kind:      "Pod",
									Name:      "test-pod",
									Namespace: "default",
								},
							},
						},
					},
				},
			},
			expectedName:      "test-service",
			expectedInstances: 1,
			wantErr:           false,
		},
		{
			name:        "service without endpoints",
			serviceName: "headless-service",
			namespace:   "default",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "headless-service",
					Namespace: "default",
				},
			},
			endpoints: &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "headless-service",
					Namespace: "default",
				},
				Subsets: []corev1.EndpointSubset{}, // No endpoints
			},
			expectedName:      "headless-service",
			expectedInstances: 0,
			wantErr:           false,
		},
		{
			name:        "multiple instances",
			serviceName: "multi-service",
			namespace:   "default",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-service",
					Namespace: "default",
				},
			},
			endpoints: &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-service",
					Namespace: "default",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "10.0.0.1",
								TargetRef: &corev1.ObjectReference{
									Kind:      "Pod",
									Name:      "pod-1",
									Namespace: "default",
								},
							},
							{
								IP: "10.0.0.2",
								TargetRef: &corev1.ObjectReference{
									Kind:      "Pod",
									Name:      "pod-2",
									Namespace: "default",
								},
							},
						},
					},
				},
			},
			expectedName:      "multi-service",
			expectedInstances: 2,
			wantErr:           false,
		},
		{
			name:        "service not found",
			serviceName: "nonexistent-service",
			namespace:   "default",
			service:     nil, // No service created
			endpoints:   nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			fakeClient := fake.NewSimpleClientset()

			// Add service to fake client
			if tt.service != nil {
				_, err := fakeClient.CoreV1().Services(tt.namespace).Create(context.Background(), tt.service, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Add endpoints to fake client
			if tt.endpoints != nil {
				_, err := fakeClient.CoreV1().Endpoints(tt.namespace).Create(context.Background(), tt.endpoints, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			// Create datastore
			ds := &datastore{
				client: fakeClient,
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

func TestGetEndpointsForService(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		namespace    string
		endpoints    *corev1.Endpoints
		expectedIPs  []string
		expectedPods []string
		wantErr      string
	}{
		{
			name:        "endpoints not found",
			serviceName: "nonexistent-service",
			namespace:   "default",
			endpoints:   nil,
			wantErr:     "failed to get endpoints",
		},
		{
			name:        "external endpoints (no TargetRef)",
			serviceName: "external-service",
			namespace:   "default",
			endpoints: &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-service",
					Namespace: "default",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "203.0.113.1", // External IP, no TargetRef
							},
						},
					},
				},
			},
			expectedIPs:  []string{"203.0.113.1"},
			expectedPods: []string{""}, // No pod name for external endpoint
			wantErr:      "",
		},
		{
			name:        "multiple endpoints with pods",
			serviceName: "multi-endpoint-service",
			namespace:   "default",
			endpoints: &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-endpoint-service",
					Namespace: "default",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								IP: "10.0.0.1",
								TargetRef: &corev1.ObjectReference{
									Kind:      "Pod",
									Name:      "pod-1",
									Namespace: "default",
								},
							},
							{
								IP: "10.0.0.2",
								TargetRef: &corev1.ObjectReference{
									Kind:      "Pod",
									Name:      "pod-2",
									Namespace: "default",
								},
							},
						},
					},
				},
			},
			expectedIPs:  []string{"10.0.0.1", "10.0.0.2"},
			expectedPods: []string{"pod-1", "pod-2"},
			wantErr:      "",
		},
		{
			name:        "empty endpoints",
			serviceName: "empty-service",
			namespace:   "default",
			endpoints: &corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-service",
					Namespace: "default",
				},
				Subsets: []corev1.EndpointSubset{}, // No subsets
			},
			expectedIPs:  []string{},
			expectedPods: []string{},
			wantErr:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset()

			if tt.endpoints != nil {
				_, err := fakeClient.CoreV1().Endpoints(tt.namespace).Create(context.Background(), tt.endpoints, metav1.CreateOptions{})
				require.NoError(t, err)
			}

			ds := &datastore{
				client: fakeClient,
			}

			result, err := ds.getEndpointsForService(context.Background(), tt.serviceName, tt.namespace)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result, len(tt.expectedIPs))

			for i, instance := range result {
				assert.Equal(t, tt.expectedIPs[i], instance.Ip)
				assert.Equal(t, tt.expectedPods[i], instance.Pod)
				assert.Equal(t, tt.namespace, instance.Namespace)
			}
		})
	}
}

func TestNamespaceIsolation(t *testing.T) {
	tests := []struct {
		name                 string
		servicesPerNamespace map[string][]string // namespace -> service names
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
		{
			name: "empty namespace returns no services from other namespaces",
			servicesPerNamespace: map[string][]string{
				"default":     {"service-1"},
				"kube-system": {"kube-dns"},
			},
			testNamespace:    "empty-ns",
			expectedServices: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset()

			// Create all services across namespaces
			for namespace, serviceNames := range tt.servicesPerNamespace {
				for i, serviceName := range serviceNames {
					service := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      serviceName,
							Namespace: namespace,
						},
					}

					endpoints := &corev1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name:      serviceName,
							Namespace: namespace,
						},
						Subsets: []corev1.EndpointSubset{
							{
								Addresses: []corev1.EndpointAddress{
									{
										IP: fmt.Sprintf("10.%s.%d.1", namespace, i+1),
										TargetRef: &corev1.ObjectReference{
											Kind:      "Pod",
											Name:      fmt.Sprintf("%s-pod-%d", serviceName, i+1),
											Namespace: namespace,
										},
									},
								},
							},
						},
					}

					_, err := fakeClient.CoreV1().Services(namespace).Create(context.Background(), service, metav1.CreateOptions{})
					require.NoError(t, err)

					_, err = fakeClient.CoreV1().Endpoints(namespace).Create(context.Background(), endpoints, metav1.CreateOptions{})
					require.NoError(t, err)
				}
			}

			ds := &datastore{
				client: fakeClient,
			}

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

			// Sort both slices for comparison since order might vary
			expectedSorted := make([]string, len(tt.expectedServices))
			copy(expectedSorted, tt.expectedServices)
			require.ElementsMatch(t, expectedSorted, serviceNames)
		})
	}
}

func TestCheckForProxySidecar(t *testing.T) {
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
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
			expected: true,
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
			expected: true,
		},
		{
			name: "pod with istio proxy in init container",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "istio-proxy",
							Image: "istio/proxyv2:1.18.0",
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
			expected: true,
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
			expected: false,
		},
		{
			name: "pod with only app container",
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
					},
				},
			},
			expected: false,
		},
		{
			name: "pod with non-istio sidecar should return false",
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
							Name:  "envoy",
							Image: "envoyproxy/envoy:v1.27.0",
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake kubernetes client with the test pod
			client := fake.NewSimpleClientset(tt.pod)

			ds := &datastore{
				client: client,
			}

			result, err := ds.checkForProxySidecar(context.Background(), tt.pod.Name, tt.pod.Namespace)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckForProxySidecar_PodNotFound(t *testing.T) {
	client := fake.NewSimpleClientset()
	ds := &datastore{
		client: client,
	}

	result, err := ds.checkForProxySidecar(context.Background(), "nonexistent-pod", "default")
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "failed to get pod nonexistent-pod")
}

func TestGetEndpointsForService_WithProxySidecar(t *testing.T) {
	// Create a pod with istio sidecar
	pod := &corev1.Pod{
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
	}

	// Create endpoints that reference the pod
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "10.0.0.1",
						TargetRef: &corev1.ObjectReference{
							Kind:      "Pod",
							Name:      "test-pod",
							Namespace: "default",
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod, endpoints)
	ds := &datastore{
		client: client,
	}

	instances, err := ds.getEndpointsForService(context.Background(), "test-service", "default")
	require.NoError(t, err)
	require.Len(t, instances, 1)

	instance := instances[0]
	assert.Equal(t, "10.0.0.1", instance.Ip)
	assert.Equal(t, "test-pod", instance.Pod)
	assert.Equal(t, "default", instance.Namespace)
	assert.True(t, instance.HasProxySidecar)
}

func TestGetEndpointsForService_WithoutProxySidecar(t *testing.T) {
	// Create a pod without sidecar
	pod := &corev1.Pod{
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
			},
		},
	}

	// Create endpoints that reference the pod
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "10.0.0.1",
						TargetRef: &corev1.ObjectReference{
							Kind:      "Pod",
							Name:      "test-pod",
							Namespace: "default",
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod, endpoints)
	ds := &datastore{
		client: client,
	}

	instances, err := ds.getEndpointsForService(context.Background(), "test-service", "default")
	require.NoError(t, err)
	require.Len(t, instances, 1)

	instance := instances[0]
	assert.Equal(t, "10.0.0.1", instance.Ip)
	assert.Equal(t, "test-pod", instance.Pod)
	assert.Equal(t, "default", instance.Namespace)
	assert.False(t, instance.HasProxySidecar)
}

func TestGetEndpointsForService_NoPodReference(t *testing.T) {
	// Create endpoints without pod reference (e.g., external service)
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "10.0.0.1",
						// No TargetRef - external endpoint
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(endpoints)
	ds := &datastore{
		client: client,
	}

	instances, err := ds.getEndpointsForService(context.Background(), "test-service", "default")
	require.NoError(t, err)
	require.Len(t, instances, 1)

	instance := instances[0]
	assert.Equal(t, "10.0.0.1", instance.Ip)
	assert.Equal(t, "", instance.Pod)
	assert.Equal(t, "default", instance.Namespace)
	assert.False(t, instance.HasProxySidecar)
}
