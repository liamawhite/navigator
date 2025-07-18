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

// Package admin provides Envoy admin interface access via pilot-agent commands.
// This package implements the same approach used by istioctl to retrieve Envoy
// configuration from Istio sidecar containers.
package admin

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// AdminClient defines the interface for accessing Envoy admin endpoints
// This interface provides access to Envoy configuration and status information
// through the pilot-agent running in istio-proxy sidecar containers.
type AdminClient interface {
	// GetConfigDump retrieves the complete Envoy configuration dump
	// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET config_dump
	GetConfigDump(ctx context.Context, namespace, podName string) (string, error)

	// GetServerInfo retrieves Envoy server information including version
	// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET server_info
	GetServerInfo(ctx context.Context, namespace, podName string) (string, error)

	// GetClusters retrieves live cluster status with endpoint health information
	// Equivalent to: kubectl exec POD -c istio-proxy -- pilot-agent request GET clusters
	GetClusters(ctx context.Context, namespace, podName string) (string, error)

	// GetProxyVersion extracts and returns the Envoy proxy version
	GetProxyVersion(ctx context.Context, namespace, podName string) (string, error)

	// IsIstioProxyReady checks if the istio-proxy container is ready for admin commands
	IsIstioProxyReady(ctx context.Context, namespace, podName string) (bool, error)
}

// NewAdminClient creates a new AdminClient using the pilot-agent approach
// This is the recommended way to create an AdminClient as it uses the same
// method as istioctl for maximum compatibility and reliability.
func NewAdminClient(clientset kubernetes.Interface, restConfig *rest.Config) AdminClient {
	kubectlExec := NewKubectlExecClient(clientset, restConfig)
	return NewPilotAgentClient(kubectlExec)
}

// NewAdminClientWithKubectl creates an AdminClient with a custom kubectl exec implementation
// This is useful for testing or when you need to customize the kubectl execution behavior.
func NewAdminClientWithKubectl(kubectlExec KubectlExecInterface) AdminClient {
	return NewPilotAgentClient(kubectlExec)
}

// AdminClientConfig holds configuration for admin client creation
type AdminClientConfig struct {
	// Kubernetes client configuration
	Clientset  kubernetes.Interface
	RestConfig *rest.Config
}

// NewAdminClientFromConfig creates an AdminClient from configuration
func NewAdminClientFromConfig(config *AdminClientConfig) AdminClient {
	return NewAdminClient(config.Clientset, config.RestConfig)
}
