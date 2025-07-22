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

// Package client provides Istio-specific Envoy admin interface access.
// This package defines interfaces and factory functions for accessing
// Envoy configuration through pilot-agent commands.
package client

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/liamawhite/navigator/pkg/istio/proxy/client/pilotagent"
)

// AdminClient is the interface used by the proxy service
// This maintains backward compatibility with existing code
type AdminClient interface {
	GetConfigDump(ctx context.Context, namespace, podName string) (string, error)
	GetServerInfo(ctx context.Context, namespace, podName string) (string, error)
	GetClusters(ctx context.Context, namespace, podName string) (string, error)
	GetProxyVersion(ctx context.Context, namespace, podName string) (string, error)
	IsIstioProxyReady(ctx context.Context, namespace, podName string) (bool, error)
}

// NewAdminClient creates a new Istio AdminClient
// This maintains backward compatibility with existing code
func NewAdminClient(clientset kubernetes.Interface, restConfig *rest.Config) AdminClient {
	return pilotagent.NewClient(clientset, restConfig)
}

// KubectlExecInterface defines the interface for executing commands in Kubernetes pods
// This is kept for backward compatibility with existing test mocks
type KubectlExecInterface interface {
	ExecInContainer(ctx context.Context, namespace, podName, container string, command []string) (string, error)
	HasContainer(ctx context.Context, namespace, podName, container string) (bool, error)
	PodExists(ctx context.Context, namespace, podName string) (bool, error)
}

// NewAdminClientWithKubectl creates an Istio AdminClient with custom kubectl exec
// This maintains backward compatibility with existing code that provides custom exec implementations
func NewAdminClientWithKubectl(kubectlExec KubectlExecInterface) AdminClient {
	// For backward compatibility, we create a wrapper that adapts the kubectl interface
	return &kubectlExecAdapter{exec: kubectlExec}
}

// kubectlExecAdapter adapts the old KubectlExecInterface to the new AdminClient interface
type kubectlExecAdapter struct {
	exec KubectlExecInterface
}

// GetConfigDump implementation for backward compatibility
func (a *kubectlExecAdapter) GetConfigDump(ctx context.Context, namespace, podName string) (string, error) {
	// This is a simple adapter - for production use, prefer NewAdminClient
	command := []string{"pilot-agent", "request", "GET", "config_dump"}
	return a.exec.ExecInContainer(ctx, namespace, podName, "istio-proxy", command)
}

// GetServerInfo implementation for backward compatibility
func (a *kubectlExecAdapter) GetServerInfo(ctx context.Context, namespace, podName string) (string, error) {
	command := []string{"pilot-agent", "request", "GET", "server_info"}
	return a.exec.ExecInContainer(ctx, namespace, podName, "istio-proxy", command)
}

// GetClusters implementation for backward compatibility
func (a *kubectlExecAdapter) GetClusters(ctx context.Context, namespace, podName string) (string, error) {
	command := []string{"pilot-agent", "request", "GET", "clusters?format=json"}
	return a.exec.ExecInContainer(ctx, namespace, podName, "istio-proxy", command)
}

// GetProxyVersion implementation for backward compatibility
func (a *kubectlExecAdapter) GetProxyVersion(ctx context.Context, namespace, podName string) (string, error) {
	// Simple implementation - just return "unknown" for adapter
	return "unknown", nil
}

// IsIstioProxyReady implementation for backward compatibility
func (a *kubectlExecAdapter) IsIstioProxyReady(ctx context.Context, namespace, podName string) (bool, error) {
	exists, err := a.exec.PodExists(ctx, namespace, podName)
	if err != nil || !exists {
		return false, err
	}
	return a.exec.HasContainer(ctx, namespace, podName, "istio-proxy")
}
