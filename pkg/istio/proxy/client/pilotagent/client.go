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

// Package pilotagent provides Istio pilot-agent based admin client implementation.
package pilotagent

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Istio sidecar container name
const IstioProxyContainer = "istio-proxy"

// Client provides direct access to Istio sidecar admin interface via pilot-agent
type Client struct {
	clientset  kubernetes.Interface
	restConfig *rest.Config
}

// NewClient creates a new Istio pilot-agent client
func NewClient(clientset kubernetes.Interface, restConfig *rest.Config) *Client {
	return &Client{
		clientset:  clientset,
		restConfig: restConfig,
	}
}

// GetConfigDump retrieves the Envoy configuration dump from istio-proxy container
// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET config_dump
func (c *Client) GetConfigDump(ctx context.Context, namespace, podName string) (string, error) {
	// Validate the pod has istio-proxy container
	if err := c.validateIstioProxy(ctx, namespace, podName); err != nil {
		return "", err
	}

	// Execute pilot-agent request GET config_dump
	command := []string{"pilot-agent", "request", "GET", "config_dump"}
	output, err := c.execInContainer(ctx, namespace, podName, IstioProxyContainer, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute pilot-agent config_dump: %w", err)
	}

	// Validate the output is valid JSON (basic check)
	output = strings.TrimSpace(output)
	if !strings.HasPrefix(output, "{") || !strings.HasSuffix(output, "}") {
		return "", fmt.Errorf("invalid config dump output: expected JSON object")
	}

	return output, nil
}

// GetServerInfo retrieves Envoy server information from istio-proxy container
// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET server_info
func (c *Client) GetServerInfo(ctx context.Context, namespace, podName string) (string, error) {
	// Validate the pod has istio-proxy container
	if err := c.validateIstioProxy(ctx, namespace, podName); err != nil {
		return "", err
	}

	// Execute pilot-agent request GET server_info
	command := []string{"pilot-agent", "request", "GET", "server_info"}
	output, err := c.execInContainer(ctx, namespace, podName, IstioProxyContainer, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute pilot-agent server_info: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// GetClusters retrieves live cluster status from istio-proxy container
// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET clusters?format=json
func (c *Client) GetClusters(ctx context.Context, namespace, podName string) (string, error) {
	// Validate the pod has istio-proxy container
	if err := c.validateIstioProxy(ctx, namespace, podName); err != nil {
		return "", err
	}

	// Execute pilot-agent request GET clusters?format=json
	command := []string{"pilot-agent", "request", "GET", "clusters?format=json"}
	output, err := c.execInContainer(ctx, namespace, podName, IstioProxyContainer, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute pilot-agent clusters: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// GetProxyVersion extracts the Envoy version from istio-proxy container
func (c *Client) GetProxyVersion(ctx context.Context, namespace, podName string) (string, error) {
	serverInfo, err := c.GetServerInfo(ctx, namespace, podName)
	if err != nil {
		return "", fmt.Errorf("failed to get server info: %w", err)
	}

	// Parse the server info to extract version
	// Server info format example: "envoy  1.28.0/Clean/RELEASE/BoringSSL"
	lines := strings.Split(serverInfo, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "version") || strings.Contains(line, "envoy") {
			// Extract version information
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.Contains(part, ".") && len(part) > 3 {
					// Likely a version string
					return part, nil
				}
			}
		}
	}

	return "unknown", nil
}

// IsIstioProxyReady checks if the istio-proxy container is ready for admin commands
func (c *Client) IsIstioProxyReady(ctx context.Context, namespace, podName string) (bool, error) {
	// Check if pod exists and has istio-proxy container
	if err := c.validateIstioProxy(ctx, namespace, podName); err != nil {
		return false, err
	}

	// Try to get server info as a health check
	_, err := c.GetServerInfo(ctx, namespace, podName)
	if err != nil {
		return false, nil // Not ready, but not an error
	}

	return true, nil
}

// validateIstioProxy validates that the pod exists and has an istio-proxy container
func (c *Client) validateIstioProxy(ctx context.Context, namespace, podName string) error {
	// Check if pod exists
	exists, err := c.podExists(ctx, namespace, podName)
	if err != nil {
		return fmt.Errorf("failed to check pod existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("pod %s/%s does not exist", namespace, podName)
	}

	// Check if pod has istio-proxy container
	hasContainer, err := c.hasContainer(ctx, namespace, podName, IstioProxyContainer)
	if err != nil {
		return fmt.Errorf("failed to check istio-proxy container: %w", err)
	}
	if !hasContainer {
		return fmt.Errorf("pod %s/%s does not have istio-proxy sidecar container", namespace, podName)
	}

	return nil
}

// execInContainer executes a command in a specific container within a pod
func (c *Client) execInContainer(ctx context.Context, namespace, podName, container string, command []string) (string, error) {
	// Create the exec request
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	// Set up the exec options
	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	// Create the executor
	exec, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create SPDY executor: %w", err)
	}

	// Prepare buffers for stdout and stderr
	var stdout, stderr bytes.Buffer

	// Execute the command
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return "", fmt.Errorf("command execution failed: %w, stderr: %s", err, stderrStr)
		}
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return stdout.String(), nil
}

// hasContainer checks if a pod has a specific container
func (c *Client) hasContainer(ctx context.Context, namespace, podName, container string) (bool, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get pod: %w", err)
	}

	// Check if the container exists in the pod spec
	for _, cont := range pod.Spec.Containers {
		if cont.Name == container {
			return true, nil
		}
	}

	// Also check init containers
	for _, cont := range pod.Spec.InitContainers {
		if cont.Name == container {
			return true, nil
		}
	}

	return false, nil
}

// podExists checks if a pod exists in the given namespace
func (c *Client) podExists(ctx context.Context, namespace, podName string) (bool, error) {
	_, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		// Check if it's a "not found" error
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check pod existence: %w", err)
	}

	return true, nil
}
