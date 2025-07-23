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
	"log/slog"
	"strings"
	"sync"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client and provides service discovery functionality
type Client struct {
	clientset  kubernetes.Interface
	restConfig *rest.Config
	logger     *slog.Logger
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string, logger *slog.Logger) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		// Use kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	} else {
		// Use in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset:  clientset,
		restConfig: config,
		logger:     logger,
	}, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (k *Client) GetClientset() kubernetes.Interface {
	return k.clientset
}

// GetRestConfig returns the underlying Kubernetes REST config
func (k *Client) GetRestConfig() *rest.Config {
	return k.restConfig
}

// GetClusterState discovers all services in the cluster and returns the cluster state
func (k *Client) GetClusterState(ctx context.Context) (*v1alpha1.ClusterState, error) {
	// Parallelize API calls and map building in single goroutines
	var wg sync.WaitGroup
	var servicesResult *corev1.ServiceList
	var endpointSlicesByService map[string][]discoveryv1.EndpointSlice
	var podsByName map[string]*corev1.Pod
	var servicesErr, endpointSlicesErr, podsErr error

	wg.Add(3)

	// Fetch services concurrently
	go func() {
		defer wg.Done()
		servicesResult, servicesErr = k.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	}()

	// Fetch endpoint slices and build map concurrently
	go func() {
		defer wg.Done()
		endpointSlicesResult, endpointSlicesErr := k.clientset.DiscoveryV1().EndpointSlices("").List(ctx, metav1.ListOptions{})
		if endpointSlicesErr == nil {
			endpointSlicesByService = k.buildEndpointSliceMap(endpointSlicesResult.Items)
		}
	}()

	// Fetch pods and build map concurrently
	go func() {
		defer wg.Done()
		podsResult, podsErr := k.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if podsErr == nil {
			podsByName = k.buildPodMap(podsResult.Items)
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for errors
	if servicesErr != nil {
		return nil, fmt.Errorf("failed to list services: %w", servicesErr)
	}
	if endpointSlicesErr != nil {
		return nil, fmt.Errorf("failed to list endpoint slices: %w", endpointSlicesErr)
	}
	if podsErr != nil {
		return nil, fmt.Errorf("failed to list pods: %w", podsErr)
	}

	var protoServices []*v1alpha1.Service

	for _, svc := range servicesResult.Items {
		protoService := k.convertServiceWithMaps(&svc, endpointSlicesByService, podsByName)
		protoServices = append(protoServices, protoService)
	}

	return &v1alpha1.ClusterState{
		Services: protoServices,
	}, nil
}

// buildEndpointSliceMap creates a map of service name to endpoint slices for efficient lookup
func (k *Client) buildEndpointSliceMap(endpointSlices []discoveryv1.EndpointSlice) map[string][]discoveryv1.EndpointSlice {
	endpointSlicesByService := make(map[string][]discoveryv1.EndpointSlice)

	for _, slice := range endpointSlices {
		serviceName := slice.Labels["kubernetes.io/service-name"]
		if serviceName != "" {
			key := slice.Namespace + "/" + serviceName
			endpointSlicesByService[key] = append(endpointSlicesByService[key], slice)
		}
	}

	return endpointSlicesByService
}

// buildPodMap creates a map of namespace/podname to pod for efficient lookup
func (k *Client) buildPodMap(pods []corev1.Pod) map[string]*corev1.Pod {
	podsByName := make(map[string]*corev1.Pod)

	for i, pod := range pods {
		key := pod.Namespace + "/" + pod.Name
		podsByName[key] = &pods[i]
	}

	return podsByName
}

// convertServiceWithMaps converts a Kubernetes Service to a protobuf Service using prebuilt maps
func (k *Client) convertServiceWithMaps(
	svc *corev1.Service,
	endpointSlicesByService map[string][]discoveryv1.EndpointSlice,
	podsByName map[string]*corev1.Pod,
) *v1alpha1.Service {
	protoService := &v1alpha1.Service{
		Name:      svc.Name,
		Namespace: svc.Namespace,
	}

	// Get endpoint slices for this service
	serviceKey := svc.Namespace + "/" + svc.Name
	endpointSlices, exists := endpointSlicesByService[serviceKey]
	if !exists {
		// Service has no endpoints
		return protoService
	}

	// Convert endpoint slices to service instances
	instances := k.convertEndpointSlicesToInstancesWithMaps(endpointSlices, podsByName)
	protoService.Instances = instances

	return protoService
}

// convertEndpointSlicesToInstancesWithMaps converts EndpointSlices to ServiceInstances using prebuilt maps
func (k *Client) convertEndpointSlicesToInstancesWithMaps(
	endpointSlices []discoveryv1.EndpointSlice,
	podsByName map[string]*corev1.Pod,
) []*v1alpha1.ServiceInstance {
	var instances []*v1alpha1.ServiceInstance

	for _, slice := range endpointSlices {
		for _, endpoint := range slice.Endpoints {
			// Only process ready endpoints
			if endpoint.Conditions.Ready != nil && !*endpoint.Conditions.Ready {
				continue
			}

			// Get the pod name from the endpoint
			podName := ""
			if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
				podName = endpoint.TargetRef.Name
			}

			// Check for Envoy sidecar and extract additional pod info if we have a pod name
			envoyPresent := false
			var containers []*v1alpha1.Container
			podStatus := ""
			nodeName := ""
			createdAt := ""
			labels := make(map[string]string)
			annotations := make(map[string]string)

			if podName != "" {
				podKey := slice.Namespace + "/" + podName
				if pod, exists := podsByName[podKey]; exists {
					envoyPresent = k.hasEnvoySidecarInPod(pod)

					// Extract container information
					containers = k.extractContainerInfo(pod)

					// Extract pod metadata
					podStatus = string(pod.Status.Phase)
					nodeName = pod.Spec.NodeName
					if !pod.CreationTimestamp.IsZero() {
						createdAt = pod.CreationTimestamp.Format("2006-01-02T15:04:05Z")
					}

					// Copy labels and annotations (avoid nil maps)
					if pod.Labels != nil {
						for k, v := range pod.Labels {
							labels[k] = v
						}
					}
					if pod.Annotations != nil {
						for k, v := range pod.Annotations {
							annotations[k] = v
						}
					}
				}
			}

			// Create service instance for each IP address
			for _, address := range endpoint.Addresses {
				instance := &v1alpha1.ServiceInstance{
					Ip:           address,
					PodName:      podName,
					EnvoyPresent: envoyPresent,
					Containers:   containers,
					PodStatus:    podStatus,
					NodeName:     nodeName,
					CreatedAt:    createdAt,
					Labels:       labels,
					Annotations:  annotations,
				}
				instances = append(instances, instance)
			}
		}
	}

	return instances
}

// hasEnvoySidecarInPod checks if a pod has an Envoy sidecar container (no API call)
func (k *Client) hasEnvoySidecarInPod(pod *corev1.Pod) bool {
	// Check all containers for Envoy indicators
	for _, container := range pod.Spec.Containers {
		if k.isEnvoyContainer(container) {
			return true
		}
	}

	// Check init containers as well
	for _, container := range pod.Spec.InitContainers {
		if k.isEnvoyContainer(container) {
			return true
		}
	}

	return false
}

// isEnvoyContainer checks if a container is an Envoy proxy
func (k *Client) isEnvoyContainer(container corev1.Container) bool {
	// Check container name
	if strings.Contains(strings.ToLower(container.Name), "envoy") ||
		strings.Contains(strings.ToLower(container.Name), "proxy") ||
		strings.Contains(strings.ToLower(container.Name), "sidecar") {
		return true
	}

	// Check container image
	if strings.Contains(strings.ToLower(container.Image), "envoy") ||
		strings.Contains(strings.ToLower(container.Image), "istio/proxyv2") ||
		strings.Contains(strings.ToLower(container.Image), "istio-proxy") {
		return true
	}

	return false
}

// extractContainerInfo extracts container information from a pod
func (k *Client) extractContainerInfo(pod *corev1.Pod) []*v1alpha1.Container {
	var containers []*v1alpha1.Container

	// Extract information from all containers
	for _, container := range pod.Spec.Containers {
		ready := false
		status := "Unknown"
		restartCount := int32(0)

		// Find matching container status
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == container.Name {
				ready = cs.Ready
				restartCount = cs.RestartCount

				// Determine status
				if cs.State.Running != nil {
					status = "Running"
				} else if cs.State.Waiting != nil {
					status = "Waiting"
					if cs.State.Waiting.Reason != "" {
						status = cs.State.Waiting.Reason
					}
				} else if cs.State.Terminated != nil {
					status = "Terminated"
					if cs.State.Terminated.Reason != "" {
						status = cs.State.Terminated.Reason
					}
				}
				break
			}
		}

		containers = append(containers, &v1alpha1.Container{
			Name:         container.Name,
			Image:        container.Image,
			Status:       status,
			Ready:        ready,
			RestartCount: restartCount,
		})
	}

	return containers
}
