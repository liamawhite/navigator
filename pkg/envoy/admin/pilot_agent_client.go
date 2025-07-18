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

package admin

import (
	"context"
	"fmt"
	"strings"
)

// Istio sidecar container name
const IstioProxyContainer = "istio-proxy"

// PilotAgentClient implements Envoy admin interface access via pilot-agent
type PilotAgentClient struct {
	kubectlExec KubectlExecInterface
}

// NewPilotAgentClient creates a new pilot-agent admin client
func NewPilotAgentClient(kubectlExec KubectlExecInterface) *PilotAgentClient {
	return &PilotAgentClient{
		kubectlExec: kubectlExec,
	}
}

// GetConfigDump retrieves the Envoy configuration dump using pilot-agent
// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET config_dump
func (c *PilotAgentClient) GetConfigDump(ctx context.Context, namespace, podName string) (string, error) {
	// Validate the pod has istio-proxy container
	if err := c.validateIstioProxy(ctx, namespace, podName); err != nil {
		return "", err
	}

	// Execute pilot-agent request GET config_dump
	command := []string{"pilot-agent", "request", "GET", "config_dump"}
	output, err := c.kubectlExec.ExecInContainer(ctx, namespace, podName, IstioProxyContainer, command)
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

// GetServerInfo retrieves Envoy server information using pilot-agent
// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET server_info
func (c *PilotAgentClient) GetServerInfo(ctx context.Context, namespace, podName string) (string, error) {
	// Validate the pod has istio-proxy container
	if err := c.validateIstioProxy(ctx, namespace, podName); err != nil {
		return "", err
	}

	// Execute pilot-agent request GET server_info
	command := []string{"pilot-agent", "request", "GET", "server_info"}
	output, err := c.kubectlExec.ExecInContainer(ctx, namespace, podName, IstioProxyContainer, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute pilot-agent server_info: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// GetClusters retrieves live cluster status with endpoint health information using pilot-agent
// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET clusters
func (c *PilotAgentClient) GetClusters(ctx context.Context, namespace, podName string) (string, error) {
	// Validate the pod has istio-proxy container
	if err := c.validateIstioProxy(ctx, namespace, podName); err != nil {
		return "", err
	}

	// Execute pilot-agent request GET clusters?format=json
	command := []string{"pilot-agent", "request", "GET", "clusters?format=json"}
	output, err := c.kubectlExec.ExecInContainer(ctx, namespace, podName, IstioProxyContainer, command)
	if err != nil {
		return "", fmt.Errorf("failed to execute pilot-agent clusters: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// GetProxyVersion extracts the Envoy version from server info
func (c *PilotAgentClient) GetProxyVersion(ctx context.Context, namespace, podName string) (string, error) {
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
func (c *PilotAgentClient) IsIstioProxyReady(ctx context.Context, namespace, podName string) (bool, error) {
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
func (c *PilotAgentClient) validateIstioProxy(ctx context.Context, namespace, podName string) error {
	// Check if pod exists
	exists, err := c.kubectlExec.PodExists(ctx, namespace, podName)
	if err != nil {
		return fmt.Errorf("failed to check pod existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("pod %s/%s does not exist", namespace, podName)
	}

	// Check if pod has istio-proxy container
	hasContainer, err := c.kubectlExec.HasContainer(ctx, namespace, podName, IstioProxyContainer)
	if err != nil {
		return fmt.Errorf("failed to check istio-proxy container: %w", err)
	}
	if !hasContainer {
		return fmt.Errorf("pod %s/%s does not have istio-proxy sidecar container", namespace, podName)
	}

	return nil
}

// MockPilotAgentClient provides a mock implementation for testing
type MockPilotAgentClient struct {
	GetConfigDumpFunc     func(ctx context.Context, namespace, podName string) (string, error)
	GetServerInfoFunc     func(ctx context.Context, namespace, podName string) (string, error)
	GetClustersFunc       func(ctx context.Context, namespace, podName string) (string, error)
	GetProxyVersionFunc   func(ctx context.Context, namespace, podName string) (string, error)
	IsIstioProxyReadyFunc func(ctx context.Context, namespace, podName string) (bool, error)
}

// GetConfigDump mock implementation
func (m *MockPilotAgentClient) GetConfigDump(ctx context.Context, namespace, podName string) (string, error) {
	if m.GetConfigDumpFunc != nil {
		return m.GetConfigDumpFunc(ctx, namespace, podName)
	}
	return "{}", nil
}

// GetServerInfo mock implementation
func (m *MockPilotAgentClient) GetServerInfo(ctx context.Context, namespace, podName string) (string, error) {
	if m.GetServerInfoFunc != nil {
		return m.GetServerInfoFunc(ctx, namespace, podName)
	}
	return "envoy 1.28.0", nil
}

// GetClusters mock implementation
func (m *MockPilotAgentClient) GetClusters(ctx context.Context, namespace, podName string) (string, error) {
	if m.GetClustersFunc != nil {
		return m.GetClustersFunc(ctx, namespace, podName)
	}
	return "", nil
}

// GetProxyVersion mock implementation
func (m *MockPilotAgentClient) GetProxyVersion(ctx context.Context, namespace, podName string) (string, error) {
	if m.GetProxyVersionFunc != nil {
		return m.GetProxyVersionFunc(ctx, namespace, podName)
	}
	return "1.28.0", nil
}

// IsIstioProxyReady mock implementation
func (m *MockPilotAgentClient) IsIstioProxyReady(ctx context.Context, namespace, podName string) (bool, error) {
	if m.IsIstioProxyReadyFunc != nil {
		return m.IsIstioProxyReadyFunc(ctx, namespace, podName)
	}
	return true, nil
}
