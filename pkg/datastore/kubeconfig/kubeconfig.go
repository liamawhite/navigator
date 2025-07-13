package kubeconfig

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
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
	client kubernetes.Interface
	cache  *serviceCache
	cancel context.CancelFunc
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

	ds := &datastore{
		client: client,
		cache: &serviceCache{
			services: make(map[string]*v1alpha1.Service),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	ds.cancel = cancel

	go ds.startWatchers(ctx)

	return ds, nil
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
				Ip:              instance.Ip,
				Pod:             instance.Pod,
				Namespace:       instance.Namespace,
				HasProxySidecar: instance.HasProxySidecar,
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
			Ip:              instance.Ip,
			Pod:             instance.Pod,
			Namespace:       instance.Namespace,
			HasProxySidecar: instance.HasProxySidecar,
		}
	}

	logger.Info("retrieved service from cache", "id", id, "instances", len(serviceCopy.Instances))
	return serviceCopy, nil
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

		watcher, err := d.client.CoreV1().Endpoints(metav1.NamespaceAll).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error("failed to start endpoints watch", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		logger.Info("started endpoints watch")

		for event := range watcher.ResultChan() {
			if event.Type == watch.Error {
				logger.Error("endpoints watch error", "error", event.Object)
				break
			}

			endpoints, ok := event.Object.(*corev1.Endpoints)
			if !ok {
				logger.Warn("unexpected object type in endpoints watch", "type", fmt.Sprintf("%T", event.Object))
				continue
			}

			d.handleEndpointsEvent(ctx, event.Type, endpoints)
		}

		watcher.Stop()
		logger.Warn("endpoints watch stopped, restarting in 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func (d *datastore) handleEndpointsEvent(ctx context.Context, eventType watch.EventType, endpoints *corev1.Endpoints) {
	logger := logging.For(logging.ComponentDatastore)
	serviceID := fmt.Sprintf("%s:%s", endpoints.Namespace, endpoints.Name)

	logger.Debug("handling endpoints event", "event", eventType, "service", serviceID)

	d.cache.mu.Lock()
	defer d.cache.mu.Unlock()

	// Ensure service exists in cache
	if d.cache.services[serviceID] == nil {
		d.cache.services[serviceID] = &v1alpha1.Service{
			Id:        serviceID,
			Name:      endpoints.Name,
			Namespace: endpoints.Namespace,
			Instances: []*v1alpha1.ServiceInstance{},
		}
	}

	switch eventType {
	case watch.Added, watch.Modified:
		// Rebuild instances from endpoints
		var instances []*v1alpha1.ServiceInstance
		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				podName := ""
				podNamespace := endpoints.Namespace
				if address.TargetRef != nil && address.TargetRef.Kind == "Pod" {
					podName = address.TargetRef.Name
					if address.TargetRef.Namespace != "" {
						podNamespace = address.TargetRef.Namespace
					}
				}

				instances = append(instances, &v1alpha1.ServiceInstance{
					Ip:              address.IP,
					Pod:             podName,
					Namespace:       podNamespace,
					HasProxySidecar: false, // Will be updated by pod watch
				})
			}
		}

		d.cache.services[serviceID].Instances = instances

		// Trigger sidecar detection for all pods in this service
		go d.refreshSidecarStatus(ctx, serviceID)

	case watch.Deleted:
		// Clear instances but keep service if it exists
		if d.cache.services[serviceID] != nil {
			d.cache.services[serviceID].Instances = []*v1alpha1.ServiceInstance{}
		}
	}
}

func (d *datastore) refreshServiceEndpoints(ctx context.Context, serviceID string) {
	parts := strings.SplitN(serviceID, ":", 2)
	if len(parts) != 2 {
		return
	}
	namespace, name := parts[0], parts[1]

	endpoints, err := d.client.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// Endpoints might not exist yet, that's okay
		return
	}

	d.handleEndpointsEvent(ctx, watch.Modified, endpoints)
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
					instance.HasProxySidecar = hasProxySidecar
				case watch.Deleted:
					instance.HasProxySidecar = false
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
	// Check all containers in the pod for Istio proxy
	for _, container := range pod.Spec.Containers {
		if container.Name == "istio-proxy" || strings.HasPrefix(container.Image, "istio/proxyv2") {
			return true
		}
	}

	// Check init containers as well
	for _, container := range pod.Spec.InitContainers {
		if container.Name == "istio-proxy" || strings.HasPrefix(container.Image, "istio/proxyv2") {
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
		instance.HasProxySidecar = hasProxySidecar
		d.cache.mu.Unlock()
	}
}
