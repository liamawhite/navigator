package kubeconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/envoy/configdump"
	"github.com/liamawhite/navigator/pkg/logging"
	types "github.com/liamawhite/navigator/pkg/troubleshooting"
)

// Ensure datastore implements the ProxyDatastore interface
var _ types.ProxyDatastore = (*datastore)(nil)

type datastore struct {
	client kubernetes.Interface
	config *rest.Config
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

	return &datastore{
		client: client,
		config: config,
	}, nil
}

// GetProxyConfig retrieves the proxy configuration for a specific service instance.
func (d *datastore) GetProxyConfig(ctx context.Context, serviceID, instanceID string) (*v1alpha1.ProxyConfig, error) {
	logger := logging.LoggerFromContextOrDefault(ctx, logging.For(logging.ComponentDatastore), logging.ComponentDatastore)

	// Parse instance ID to get cluster, namespace, and pod name
	parts := strings.SplitN(instanceID, ":", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid instance ID format: %s (expected cluster:namespace:pod)", instanceID)
	}
	clusterName, namespace, podName := parts[0], parts[1], parts[2]

	logger.Debug("getting proxy config for instance", "service_id", serviceID, "instance_id", instanceID, "cluster", clusterName, "namespace", namespace, "pod", podName)

	// Check if the pod has a proxy sidecar
	pod, err := d.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podName, err)
	}

	hasProxy := d.checkPodForProxySidecar(pod)
	if !hasProxy {
		return nil, fmt.Errorf("pod %s/%s does not have a proxy sidecar", namespace, podName)
	}

	// Get proxy configuration using port-forward to Envoy admin interface
	configDump, proxyType, version, err := d.getEnvoyConfig(ctx, namespace, podName)
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy config for pod %s/%s: %w", namespace, podName, err)
	}

	// Parse the config dump into Envoy protobuf structures
	proxyConfig, err := d.parseEnvoyConfigDump(configDump, proxyType, version)
	if err != nil {
		logger.Warn("failed to parse config dump, returning raw data only", "error", err)
		// If parsing fails, return basic information with raw config dump
		return &v1alpha1.ProxyConfig{
			ProxyType:     proxyType,
			Version:       version,
			AdminPort:     15000,
			RawConfigDump: configDump,
		}, nil
	}

	logger.Info("retrieved and parsed proxy config", "service_id", serviceID, "instance_id", instanceID, "pod", podName, "proxy_type", proxyType)
	return proxyConfig, nil
}

// checkPodForProxySidecar checks if a pod has a proxy sidecar container
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

// getEnvoyConfig retrieves the configuration dump from Envoy admin interface
func (d *datastore) getEnvoyConfig(ctx context.Context, namespace, podName string) (configDump, proxyType, version string, err error) {
	logger := logging.LoggerFromContextOrDefault(ctx, logging.For(logging.ComponentDatastore), logging.ComponentDatastore)

	// Create a port-forward to the Envoy admin interface (port 15000)
	req := d.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(d.config)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create SPDY round tripper: %w", err)
	}

	// Create the dialer
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	// Create a context with timeout for the port-forward operation
	portForwardCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create channels for port-forward communication
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)
	errorChan := make(chan error, 1)

	// Set up port forwarding from local port 0 (auto-assign) to pod port 15000
	forwarder, err := portforward.New(dialer, []string{"0:15000"}, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create port forwarder: %w", err)
	}

	// Start port forwarding in a goroutine
	go func() {
		err := forwarder.ForwardPorts()
		if err != nil {
			errorChan <- fmt.Errorf("port forwarding failed: %w", err)
		}
	}()

	// Wait for port forwarding to be ready or timeout
	select {
	case <-readyChan:
		// Port forwarding is ready
	case err := <-errorChan:
		return "", "", "", err
	case <-portForwardCtx.Done():
		close(stopChan)
		return "", "", "", fmt.Errorf("timeout waiting for port forwarding to be ready")
	}

	// Get the forwarded ports
	ports, err := forwarder.GetPorts()
	if err != nil || len(ports) == 0 {
		close(stopChan)
		return "", "", "", fmt.Errorf("failed to get forwarded ports: %w", err)
	}

	localPort := ports[0].Local

	// Make HTTP request to Envoy admin interface
	client := &http.Client{Timeout: 10 * time.Second}

	// Get config dump
	configURL := fmt.Sprintf("http://localhost:%d/config_dump", localPort)
	configResp, err := client.Get(configURL)
	if err != nil {
		close(stopChan)
		return "", "", "", fmt.Errorf("failed to get config dump: %w", err)
	}
	defer func() {
		if err := configResp.Body.Close(); err != nil {
			logger.Warn("failed to close config response body", "error", err)
		}
	}()

	configData, err := io.ReadAll(configResp.Body)
	if err != nil {
		close(stopChan)
		return "", "", "", fmt.Errorf("failed to read config dump response: %w", err)
	}

	// Get server info for proxy type and version
	infoURL := fmt.Sprintf("http://localhost:%d/server_info", localPort)
	infoResp, err := client.Get(infoURL)
	if err != nil {
		logger.Warn("failed to get server info, using defaults", "error", err)
		proxyType = "envoy"
		version = "unknown"
	} else {
		defer func() {
			if err := infoResp.Body.Close(); err != nil {
				logger.Warn("failed to close info response body", "error", err)
			}
		}()
		infoData, err := io.ReadAll(infoResp.Body)
		if err != nil {
			logger.Warn("failed to read server info response", "error", err)
			proxyType = "envoy"
			version = "unknown"
		} else {
			var serverInfo map[string]interface{}
			if err := json.Unmarshal(infoData, &serverInfo); err != nil {
				logger.Warn("failed to parse server info JSON", "error", err)
				proxyType = "envoy"
				version = "unknown"
			} else {
				if v, ok := serverInfo["version"].(string); ok {
					version = v
				} else {
					version = "unknown"
				}
				// Check if this is Istio proxy
				if cmdLine, ok := serverInfo["command_line_options"].(map[string]interface{}); ok {
					if serviceCluster, ok := cmdLine["service-cluster"].(string); ok && strings.Contains(serviceCluster, "istio") {
						proxyType = "istio-proxy"
					} else {
						proxyType = "envoy"
					}
				} else {
					proxyType = "envoy"
				}
			}
		}
	}

	// Clean up port forwarding
	close(stopChan)

	logger.Debug("retrieved envoy config", "pod", podName, "config_size", len(configData), "proxy_type", proxyType, "version", version)

	return string(configData), proxyType, version, nil
}

// parseEnvoyConfigDump parses the raw Envoy config dump JSON into structured protobuf messages
func (d *datastore) parseEnvoyConfigDump(rawConfigDump, proxyType, version string) (*v1alpha1.ProxyConfig, error) {
	logger := logging.For(logging.ComponentDatastore)

	// Use the dedicated configdump parser to get summary structs
	parser := configdump.NewParser()
	parsed, err := parser.ParseJSONToSummary(rawConfigDump)
	if err != nil {
		logger.Warn("failed to parse config dump, returning raw data only", "error", err)
		// If parsing fails, return basic information with raw config dump
		return &v1alpha1.ProxyConfig{
			ProxyType:     proxyType,
			Version:       version,
			AdminPort:     15000,
			RawConfigDump: rawConfigDump,
		}, nil
	}

	logger.Debug("parsed envoy config", "bootstrap", parsed.Bootstrap != nil, "listeners", len(parsed.Listeners), "clusters", len(parsed.Clusters), "endpoints", len(parsed.Endpoints), "routes", len(parsed.Routes))

	return &v1alpha1.ProxyConfig{
		ProxyType:     proxyType,
		Version:       version,
		AdminPort:     15000,
		Bootstrap:     parsed.Bootstrap,
		Listeners:     parsed.Listeners,
		Clusters:      parsed.Clusters,
		Endpoints:     parsed.Endpoints,
		Routes:        parsed.Routes,
		RawConfigDump: rawConfigDump,
	}, nil
}
