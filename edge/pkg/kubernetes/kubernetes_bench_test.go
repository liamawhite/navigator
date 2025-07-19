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
	"fmt"
	"testing"

	"github.com/liamawhite/navigator/pkg/logging"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// BenchmarkGetClusterState_SmallCluster benchmarks cluster state retrieval for a small cluster
func BenchmarkGetClusterState_SmallCluster(b *testing.B) {
	benchmarkGetClusterState(b, 10, 20) // 10 services, 20 pods
}

// BenchmarkGetClusterState_MediumCluster benchmarks cluster state retrieval for a medium cluster
func BenchmarkGetClusterState_MediumCluster(b *testing.B) {
	benchmarkGetClusterState(b, 100, 200) // 100 services, 200 pods
}

// BenchmarkGetClusterState_LargeCluster benchmarks cluster state retrieval for a large cluster
func BenchmarkGetClusterState_LargeCluster(b *testing.B) {
	benchmarkGetClusterState(b, 1000, 2000) // 1000 services, 2000 pods
}

// BenchmarkGetClusterState_VeryLargeCluster benchmarks cluster state retrieval for a very large cluster
func BenchmarkGetClusterState_VeryLargeCluster(b *testing.B) {
	benchmarkGetClusterState(b, 5000, 10000) // 5000 services, 10000 pods
}

// benchmarkGetClusterState is a helper function that benchmarks cluster state retrieval
func benchmarkGetClusterState(b *testing.B, numServices, numPods int) {
	// Create fake clientset with test data
	clientset := fake.NewSimpleClientset()

	// Create services
	services := generateServices(numServices)
	for _, svc := range services {
		_, _ = clientset.CoreV1().Services(svc.Namespace).Create(context.TODO(), &svc, metav1.CreateOptions{})
	}

	// Create endpoint slices
	endpointSlices := generateEndpointSlices(numServices, numPods)
	for _, eps := range endpointSlices {
		_, _ = clientset.DiscoveryV1().EndpointSlices(eps.Namespace).Create(context.TODO(), &eps, metav1.CreateOptions{})
	}

	// Create pods
	pods := generatePods(numPods)
	for _, pod := range pods {
		_, _ = clientset.CoreV1().Pods(pod.Namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
	}

	// Create kubernetes client
	k8sClient := &Client{
		clientset: clientset,
		logger:    logging.For("bench"),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := k8sClient.GetClusterState(context.TODO())
		if err != nil {
			b.Fatalf("GetClusterState failed: %v", err)
		}
	}
}

// BenchmarkMapOperations benchmarks the performance of map-based lookups
func BenchmarkMapOperations(b *testing.B) {
	numServices := 1000
	numPods := 2000

	// Generate test data
	endpointSlices := generateEndpointSlices(numServices, numPods)
	pods := generatePods(numPods)

	client := &Client{
		logger: logging.For("bench"),
	}

	b.Run("BuildEndpointSliceMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.buildEndpointSliceMap(endpointSlices)
		}
	})

	b.Run("BuildPodMap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.buildPodMap(pods)
		}
	})

	// Benchmark map lookups
	endpointSliceMap := client.buildEndpointSliceMap(endpointSlices)
	podMap := client.buildPodMap(pods)

	b.Run("MapLookups", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				serviceKey := fmt.Sprintf("default/service-%d", j%numServices)
				podKey := fmt.Sprintf("default/pod-%d", j%numPods)

				_ = endpointSliceMap[serviceKey]
				_ = podMap[podKey]
			}
		}
	})
}

// generateServices creates test services
func generateServices(count int) []corev1.Service {
	services := make([]corev1.Service, count)

	for i := 0; i < count; i++ {
		services[i] = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("service-%d", i),
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": fmt.Sprintf("app-%d", i),
				},
				Ports: []corev1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		}
	}

	return services
}

// generateEndpointSlices creates test endpoint slices
func generateEndpointSlices(numServices, numPods int) []discoveryv1.EndpointSlice {
	endpointSlices := make([]discoveryv1.EndpointSlice, numServices)

	for i := 0; i < numServices; i++ {
		// Create 1-3 endpoints per service
		numEndpoints := 1 + (i % 3)
		endpoints := make([]discoveryv1.Endpoint, numEndpoints)

		for j := 0; j < numEndpoints; j++ {
			podIndex := (i*3 + j) % numPods
			endpoints[j] = discoveryv1.Endpoint{
				Addresses: []string{fmt.Sprintf("10.0.%d.%d", i%256, j%256)},
				Conditions: discoveryv1.EndpointConditions{
					Ready: boolPtr(true),
				},
				TargetRef: &corev1.ObjectReference{
					Kind: "Pod",
					Name: fmt.Sprintf("pod-%d", podIndex),
				},
			}
		}

		endpointSlices[i] = discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("service-%d-abc123", i),
				Namespace: "default",
				Labels: map[string]string{
					"kubernetes.io/service-name": fmt.Sprintf("service-%d", i),
				},
			},
			Endpoints: endpoints,
		}
	}

	return endpointSlices
}

// generatePods creates test pods
func generatePods(count int) []corev1.Pod {
	pods := make([]corev1.Pod, count)

	for i := 0; i < count; i++ {
		containers := []corev1.Container{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		}

		// Every 3rd pod has an Envoy sidecar
		if i%3 == 0 {
			containers = append(containers, corev1.Container{
				Name:  "envoy",
				Image: "envoyproxy/envoy:v1.20.0",
			})
		}

		pods[i] = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("pod-%d", i),
				Namespace: "default",
				Labels: map[string]string{
					"app": fmt.Sprintf("app-%d", i%100),
				},
			},
			Spec: corev1.PodSpec{
				Containers: containers,
			},
		}
	}

	return pods
}
