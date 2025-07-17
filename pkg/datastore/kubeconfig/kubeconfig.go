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
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/datastore"
	"github.com/liamawhite/navigator/pkg/logging"
)

// Ensure datastore implements the ServiceDatastore interface
var _ types.ServiceDatastore = (*datastore)(nil)

type datastore struct {
	client      kubernetes.Interface
	cache       *serviceCache
	cancel      context.CancelFunc
	clusterName string
}

type serviceCache struct {
	mu       sync.RWMutex
	services map[string]*v1alpha1.Service // key: namespace:name
}

func New(kubeconfigPath string) (*datastore, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Extract cluster name from kubeconfig
	clusterName, err := extractClusterName(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract cluster name: %w", err)
	}

	ds := &datastore{
		client:      client,
		clusterName: clusterName,
		cache: &serviceCache{
			services: make(map[string]*v1alpha1.Service),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	ds.cancel = cancel

	go ds.startWatchers(ctx)

	return ds, nil
}

// extractClusterName extracts the cluster name from kubeconfig
func extractClusterName(kubeconfigPath string) (string, error) {
	// Load the kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		// Try loading from default locations if explicit path fails
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		if kubeconfigPath != "" {
			loadingRules.ExplicitPath = kubeconfigPath
		}
		config, err = loadingRules.Load()
		if err != nil {
			return "", fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	// Get the current context
	currentContext := config.CurrentContext
	if currentContext == "" {
		return "", fmt.Errorf("no current context set in kubeconfig")
	}

	// Get the context details
	context, exists := config.Contexts[currentContext]
	if !exists {
		return "", fmt.Errorf("current context %s not found in kubeconfig", currentContext)
	}

	// Return the cluster name from the context
	clusterName := context.Cluster
	if clusterName == "" {
		return "", fmt.Errorf("no cluster specified in current context %s", currentContext)
	}

	return clusterName, nil
}

func (d *datastore) ListServices(ctx context.Context, namespace string) ([]*v1alpha1.Service, error) {
	logger := logging.LoggerFromContextOrDefault(ctx, logging.For(logging.ComponentDatastore), logging.ComponentDatastore)

	logger.Debug("listing services from cache", "namespace", namespace)

	d.cache.mu.RLock()
	defer d.cache.mu.RUnlock()

	var result []*v1alpha1.Service
	for _, service := range d.cache.services {
		// Filter by namespace if specified
		if namespace != "" && service.Namespace != namespace {
			continue
		}

		// Create a copy of the service to avoid shared reference issues
		serviceCopy := &v1alpha1.Service{
			Id:        service.Id,
			Name:      service.Name,
			Namespace: service.Namespace,
			Instances: make([]*v1alpha1.ServiceInstance, len(service.Instances)),
		}

		// Copy instances
		for i, instance := range service.Instances {
			serviceCopy.Instances[i] = &v1alpha1.ServiceInstance{
				InstanceId:     instance.InstanceId,
				Ip:             instance.Ip,
				Pod:            instance.Pod,
				Namespace:      instance.Namespace,
				ClusterName:    instance.ClusterName,
				IsEnvoyPresent: instance.IsEnvoyPresent,
			}
		}

		result = append(result, serviceCopy)
	}

	logger.Info("listed services from cache", "count", len(result), "namespace", namespace)
	return result, nil
}

func (d *datastore) GetService(ctx context.Context, id string) (*v1alpha1.Service, error) {
	logger := logging.LoggerFromContextOrDefault(ctx, logging.For(logging.ComponentDatastore), logging.ComponentDatastore)

	// Parse namespace:name from ID
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		logger.Error("invalid service ID format", "id", id)
		return nil, fmt.Errorf("invalid service ID format: %s (expected namespace:name)", id)
	}

	logger.Debug("getting service from cache", "id", id)

	d.cache.mu.RLock()
	service, exists := d.cache.services[id]
	d.cache.mu.RUnlock()

	if !exists {
		logger.Error("service not found in cache", "id", id)
		return nil, fmt.Errorf("service not found: %s", id)
	}

	// Create a copy of the service to avoid shared reference issues
	serviceCopy := &v1alpha1.Service{
		Id:        service.Id,
		Name:      service.Name,
		Namespace: service.Namespace,
		Instances: make([]*v1alpha1.ServiceInstance, len(service.Instances)),
	}

	// Copy instances
	for i, instance := range service.Instances {
		serviceCopy.Instances[i] = &v1alpha1.ServiceInstance{
			InstanceId:     instance.InstanceId,
			Ip:             instance.Ip,
			Pod:            instance.Pod,
			Namespace:      instance.Namespace,
			ClusterName:    instance.ClusterName,
			IsEnvoyPresent: instance.IsEnvoyPresent,
		}
	}

	logger.Info("retrieved service from cache", "id", id, "instances", len(serviceCopy.Instances))
	return serviceCopy, nil
}

func (d *datastore) GetServiceInstance(ctx context.Context, serviceID, instanceID string) (*v1alpha1.ServiceInstanceDetail, error) {
	logger := logging.LoggerFromContextOrDefault(ctx, logging.For(logging.ComponentDatastore), logging.ComponentDatastore)

	// Parse service ID to get namespace and service name
	serviceParts := strings.SplitN(serviceID, ":", 2)
	if len(serviceParts) != 2 {
		logger.Error("invalid service ID format", "id", serviceID)
		return nil, fmt.Errorf("invalid service ID format: %s (expected namespace:name)", serviceID)
	}
	_, serviceName := serviceParts[0], serviceParts[1]

	// Parse instance ID to get cluster, namespace, and pod name
	instanceParts := strings.SplitN(instanceID, ":", 3)
	if len(instanceParts) != 3 {
		logger.Error("invalid instance ID format", "id", instanceID)
		return nil, fmt.Errorf("invalid instance ID format: %s (expected cluster:namespace:pod)", instanceID)
	}
	clusterName, namespace, podName := instanceParts[0], instanceParts[1], instanceParts[2]

	logger.Debug("getting service instance detail", "service_id", serviceID, "instance_id", instanceID, "pod", podName)

	// Get the pod from Kubernetes
	pod, err := d.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		logger.Error("failed to get pod", "namespace", namespace, "pod", podName, "error", err)
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podName, err)
	}

	// Check if pod has proxy sidecar
	hasProxySidecar := d.checkPodForProxySidecar(pod)

	// Convert containers info
	var containers []*v1alpha1.ContainerInfo
	for _, container := range pod.Spec.Containers {
		// Find container status
		var ready bool
		var restartCount int32
		var status string
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name == container.Name {
				ready = containerStatus.Ready
				restartCount = containerStatus.RestartCount
				if containerStatus.State.Running != nil {
					status = "Running"
				} else if containerStatus.State.Waiting != nil {
					status = "Waiting"
				} else if containerStatus.State.Terminated != nil {
					status = "Terminated"
				} else {
					status = "Unknown"
				}
				break
			}
		}

		containers = append(containers, &v1alpha1.ContainerInfo{
			Name:         container.Name,
			Image:        container.Image,
			Ready:        ready,
			RestartCount: restartCount,
			Status:       status,
		})
	}

	// Convert labels and annotations
	labels := make(map[string]string)
	for k, v := range pod.Labels {
		labels[k] = v
	}

	annotations := make(map[string]string)
	for k, v := range pod.Annotations {
		annotations[k] = v
	}

	// Create the detailed instance response
	instanceDetail := &v1alpha1.ServiceInstanceDetail{
		InstanceId:     instanceID,
		Ip:             pod.Status.PodIP,
		Pod:            podName,
		Namespace:      namespace,
		ClusterName:    clusterName,
		IsEnvoyPresent: hasProxySidecar,
		ServiceName:    serviceName,
		PodStatus:      string(pod.Status.Phase),
		CreatedAt:      pod.CreationTimestamp.Format(time.RFC3339),
		Labels:         labels,
		Annotations:    annotations,
		Containers:     containers,
		NodeName:       pod.Spec.NodeName,
	}

	logger.Info("retrieved service instance detail", "service_id", serviceID, "instance_id", instanceID, "pod_status", instanceDetail.PodStatus)
	return instanceDetail, nil
}

func (d *datastore) Close() {
	if d.cancel != nil {
		d.cancel()
	}
}

func (d *datastore) startWatchers(ctx context.Context) {
	logger := logging.For(logging.ComponentDatastore)
	logger.Info("starting kubernetes watchers")

	go d.watchServices(ctx)
	go d.watchEndpoints(ctx)
	go d.watchPods(ctx)
}

func (d *datastore) watchServices(ctx context.Context) {
	logger := logging.For(logging.ComponentDatastore)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		watcher, err := d.client.CoreV1().Services(metav1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error("failed to start services watch", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		logger.Info("started services watch")

		for event := range watcher.ResultChan() {
			if event.Type == watch.Error {
				logger.Error("services watch error", "error", event.Object)
				break
			}

			service, ok := event.Object.(*corev1.Service)
			if !ok {
				logger.Warn("unexpected object type in services watch", "type", fmt.Sprintf("%T", event.Object))
				continue
			}

			d.handleServiceEvent(ctx, event.Type, service)
		}

		watcher.Stop()
		logger.Warn("services watch stopped, restarting in 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func (d *datastore) handleServiceEvent(ctx context.Context, eventType watch.EventType, service *corev1.Service) {
	logger := logging.For(logging.ComponentDatastore)
	serviceID := fmt.Sprintf("%s:%s", service.Namespace, service.Name)

	logger.Debug("handling service event", "event", eventType, "service", serviceID)

	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	switch eventType {
	case watch.Added, watch.Modified:
		// Build or update the service object
		if d.cache.services[serviceID] == nil {
			d.cache.services[serviceID] = &v1alpha1.Service{
				Id:        serviceID,
				Name:      service.Name,
				Namespace: service.Namespace,
				Instances: []*v1alpha1.ServiceInstance{},
			}
		} else {
			// Update metadata but preserve instances
			d.cache.services[serviceID].Name = service.Name
			d.cache.services[serviceID].Namespace = service.Namespace
		}

		// Trigger endpoints refresh for this service
		go d.refreshServiceEndpoints(ctx, serviceID)

	case watch.Deleted:
		delete(d.cache.services, serviceID)
	}
}

func (d *datastore) watchEndpoints(ctx context.Context) {
	logger := logging.For(logging.ComponentDatastore)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		watcher, err := d.client.DiscoveryV1().EndpointSlices(metav1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error("failed to start endpoint slices watch", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		logger.Info("started endpoint slices watch")

		for event := range watcher.ResultChan() {
			if event.Type == watch.Error {
				logger.Error("endpoint slices watch error", "error", event.Object)
				break
			}

			endpointSlice, ok := event.Object.(*discoveryv1.EndpointSlice)
			if !ok {
				logger.Warn("unexpected object type in endpoint slices watch", "type", fmt.Sprintf("%T", event.Object))
				continue
			}

			d.handleEndpointSliceEvent(ctx, event.Type, endpointSlice)
		}

		watcher.Stop()
		logger.Warn("endpoint slices watch stopped, restarting in 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func (d *datastore) handleEndpointSliceEvent(ctx context.Context, eventType watch.EventType, endpointSlice *discoveryv1.EndpointSlice) {
	logger := logging.For(logging.ComponentDatastore)

	// Get service name from EndpointSlice labels
	serviceName := endpointSlice.Labels[discoveryv1.LabelServiceName]
	if serviceName == "" {
		logger.Warn("endpoint slice missing service name label", "name", endpointSlice.Name)
		return
	}

	serviceID := fmt.Sprintf("%s:%s", endpointSlice.Namespace, serviceName)

	logger.Debug("handling endpoint slice event", "event", eventType, "service", serviceID)

	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	// Ensure service exists in cache
	if d.cache.services[serviceID] == nil {
		d.cache.services[serviceID] = &v1alpha1.Service{
			Id:        serviceID,
			Name:      serviceName,
			Namespace: endpointSlice.Namespace,
			Instances: []*v1alpha1.ServiceInstance{},
		}
	}

	switch eventType {
	case watch.Added, watch.Modified:
		// For EndpointSlice, we need to aggregate instances from all slices for this service
		// Get all endpoint slices for this service
		d.rebuildServiceInstancesFromEndpointSlicesLocked(ctx, serviceID)

	case watch.Deleted:
		// Rebuild instances from remaining endpoint slices
		d.rebuildServiceInstancesFromEndpointSlicesLocked(ctx, serviceID)
	}
}

func (d *datastore) rebuildServiceInstancesFromEndpointSlices(ctx context.Context, serviceID string) {
	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()
	d.rebuildServiceInstancesFromEndpointSlicesLocked(ctx, serviceID)
}

func (d *datastore) rebuildServiceInstancesFromEndpointSlicesLocked(ctx context.Context, serviceID string) {
	logger := logging.For(logging.ComponentDatastore)
	parts := strings.SplitN(serviceID, ":", 2)
	if len(parts) != 2 {
		return
	}
	namespace, serviceName := parts[0], parts[1]

	// Get all endpoint slices for this service
	endpointSlices, err := d.client.DiscoveryV1().EndpointSlices(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", discoveryv1.LabelServiceName, serviceName),
	})
	if err != nil {
		logger.Error("failed to list endpoint slices", "error", err, "service", serviceID)
		return
	}

	logger.Debug("found endpoint slices for service", "service", serviceID, "count", len(endpointSlices.Items))

	// Aggregate instances from all endpoint slices
	var instances []*v1alpha1.ServiceInstance
	for _, endpointSlice := range endpointSlices.Items {
		logger.Debug("processing endpoint slice", "name", endpointSlice.Name, "endpoints", len(endpointSlice.Endpoints))
		for _, endpoint := range endpointSlice.Endpoints {
			if endpoint.Conditions.Ready == nil || !*endpoint.Conditions.Ready {
				continue // Skip non-ready endpoints
			}

			for _, address := range endpoint.Addresses {
				podName := ""
				podNamespace := namespace
				if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
					podName = endpoint.TargetRef.Name
					if endpoint.TargetRef.Namespace != "" {
						podNamespace = endpoint.TargetRef.Namespace
					}
				}

				instanceID := fmt.Sprintf("%s:%s:%s", d.clusterName, podNamespace, podName)
				instances = append(instances, &v1alpha1.ServiceInstance{
					InstanceId:     instanceID,
					Ip:             address,
					Pod:            podName,
					Namespace:      podNamespace,
					ClusterName:    d.clusterName,
					IsEnvoyPresent: false, // Will be updated by pod watch
				})
			}
		}
	}

	logger.Debug("created service instances from endpoint slices", "service", serviceID, "instances", len(instances))

	// Update cache with new instances (lock already held)
	if d.cache.services[serviceID] != nil {
		d.cache.services[serviceID].Instances = instances
	}

	// Trigger sidecar detection for all pods in this service
	go d.refreshSidecarStatus(ctx, serviceID)
}

func (d *datastore) refreshServiceEndpoints(ctx context.Context, serviceID string) {
	d.rebuildServiceInstancesFromEndpointSlices(ctx, serviceID)
}

func (d *datastore) watchPods(ctx context.Context) {
	logger := logging.For(logging.ComponentDatastore)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		watcher, err := d.client.CoreV1().Pods(metav1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error("failed to start pods watch", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		logger.Info("started pods watch")

		for event := range watcher.ResultChan() {
			if event.Type == watch.Error {
				logger.Error("pods watch error", "error", event.Object)
				break
			}

			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				logger.Warn("unexpected object type in pods watch", "type", fmt.Sprintf("%T", event.Object))
				continue
			}

			d.handlePodEvent(ctx, event.Type, pod)
		}

		watcher.Stop()
		logger.Warn("pods watch stopped, restarting in 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func (d *datastore) handlePodEvent(ctx context.Context, eventType watch.EventType, pod *corev1.Pod) {
	logger := logging.For(logging.ComponentDatastore)

	// Check if pod has proxy sidecar
	hasProxySidecar := d.checkPodForProxySidecar(pod)

	logger.Debug("handling pod event", "event", eventType, "pod", pod.Name, "namespace", pod.Namespace, "hasProxySidecar", hasProxySidecar)

	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	// Update sidecar status in all services that use this pod
	for _, service := range d.cache.services {
		if service.Namespace != pod.Namespace {
			continue
		}

		updated := false
		for _, instance := range service.Instances {
			if instance.Pod == pod.Name && instance.Namespace == pod.Namespace {
				switch eventType {
				case watch.Added, watch.Modified:
					instance.IsEnvoyPresent = hasProxySidecar
				case watch.Deleted:
					instance.IsEnvoyPresent = false
				}
				updated = true
			}
		}

		if updated {
			logger.Debug("updated sidecar status for service", "service", service.Id, "pod", pod.Name, "hasProxySidecar", hasProxySidecar)
		}
	}
}

func (d *datastore) checkPodForProxySidecar(pod *corev1.Pod) bool {
	// Check all containers in the pod for proxy sidecars
	for _, container := range pod.Spec.Containers {
		// Istio proxy patterns
		if container.Name == "istio-proxy" || strings.HasPrefix(container.Image, "istio/proxyv2") {
			return true
		}
		// Generic Envoy patterns
		if strings.Contains(container.Name, "envoy") || strings.Contains(container.Image, "envoy") {
			return true
		}
		// Other common proxy patterns (be more specific to avoid false positives)
		if container.Name == "proxy" || container.Name == "sidecar-proxy" {
			return true
		}
	}

	// Check init containers as well
	for _, container := range pod.Spec.InitContainers {
		// Istio proxy patterns
		if container.Name == "istio-proxy" || strings.HasPrefix(container.Image, "istio/proxyv2") {
			return true
		}
		// Generic Envoy patterns
		if strings.Contains(container.Name, "envoy") || strings.Contains(container.Image, "envoy") {
			return true
		}
		// Other common proxy patterns (be more specific to avoid false positives)
		if container.Name == "proxy" || container.Name == "sidecar-proxy" {
			return true
		}
	}

	return false
}

func (d *datastore) refreshSidecarStatus(ctx context.Context, serviceID string) {
	d.cache.mu.RLock()
	service := d.cache.services[serviceID]
	d.cache.mu.RUnlock()

	if service == nil {
		return
	}

	for _, instance := range service.Instances {
		if instance.Pod == "" {
			continue
		}

		pod, err := d.client.CoreV1().Pods(instance.Namespace).Get(ctx, instance.Pod, metav1.GetOptions{})
		if err != nil {
			continue
		}

		hasProxySidecar := d.checkPodForProxySidecar(pod)

		d.cache.mu.Lock()
		instance.IsEnvoyPresent = hasProxySidecar
		d.cache.mu.Unlock()
	}
}
