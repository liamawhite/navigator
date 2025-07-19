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

// Package proxy provides services for retrieving and processing Envoy proxy configuration
// from Istio sidecar containers using the same approach as istioctl.
package proxy

import (
	"context"
	"fmt"
	"log/slog"

	types "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/envoy/admin"
	"github.com/liamawhite/navigator/pkg/envoy/clusters"
	"github.com/liamawhite/navigator/pkg/envoy/configdump"
)

// ProxyService provides access to Envoy proxy configuration via pilot-agent
type ProxyService struct {
	adminClient admin.AdminClient
	parser      *configdump.Parser
	logger      *slog.Logger
}

// NewProxyService creates a new proxy service
func NewProxyService(adminClient admin.AdminClient, logger *slog.Logger) *ProxyService {
	return &ProxyService{
		adminClient: adminClient,
		parser:      configdump.NewParser(),
		logger:      logger,
	}
}

// GetProxyConfig retrieves and parses the complete proxy configuration for a pod
// This method implements the same workflow as istioctl proxy-config:
// 1. Execute pilot-agent request GET config_dump in the istio-proxy container
// 2. Parse the JSON configuration dump into structured protobuf types
// 3. Return the v1alpha1.ProxyConfig message
func (s *ProxyService) GetProxyConfig(ctx context.Context, namespace, podName string) (*types.ProxyConfig, error) {
	s.logger.Debug("retrieving proxy config", "namespace", namespace, "pod", podName)

	// Step 1: Get raw config dump from pilot-agent
	rawConfigDump, err := s.adminClient.GetConfigDump(ctx, namespace, podName)
	if err != nil {
		s.logger.Error("failed to get config dump", "namespace", namespace, "pod", podName, "error", err)
		return nil, fmt.Errorf("failed to get config dump for pod %s/%s: %w", namespace, podName, err)
	}

	// Step 2: Get proxy version
	version, err := s.adminClient.GetProxyVersion(ctx, namespace, podName)
	if err != nil {
		s.logger.Warn("failed to get proxy version", "namespace", namespace, "pod", podName, "error", err)
		version = "unknown"
	}

	// Step 3: Get live cluster endpoint data
	rawClusters, err := s.adminClient.GetClusters(ctx, namespace, podName)
	if err != nil {
		s.logger.Warn("failed to get clusters data", "namespace", namespace, "pod", podName, "error", err)
		// Continue without live cluster data - we'll use static config only
		rawClusters = ""
	}

	// Step 4: Parse the config dump into summary structures
	summary, err := s.parser.ParseJSONToSummary(rawConfigDump)
	if err != nil {
		s.logger.Error("failed to parse config dump", "namespace", namespace, "pod", podName, "error", err)
		return nil, fmt.Errorf("failed to parse config dump for pod %s/%s: %w", namespace, podName, err)
	}

	// Step 5: Parse cluster endpoint data from admin interface (clusters-only approach)
	var endpoints []*types.EndpointSummary
	if rawClusters != "" {
		liveEndpoints, err := clusters.ParseClustersAdminOutput(rawClusters)
		if err != nil {
			s.logger.Warn("failed to parse clusters output", "namespace", namespace, "pod", podName, "error", err)
			endpoints = []*types.EndpointSummary{}
		} else {
			// Use clusters admin interface as the exclusive source for endpoint data
			endpoints = clusters.ConvertToEndpointSummaries(liveEndpoints)
			s.logger.Debug("converted cluster endpoint data", "namespace", namespace, "pod", podName, "endpoint_summaries", len(endpoints))
		}
	} else {
		// No clusters data available, use empty endpoints (don't fall back to config dump)
		endpoints = []*types.EndpointSummary{}
		s.logger.Warn("no clusters data available, endpoints will be empty", "namespace", namespace, "pod", podName)
	}

	// Step 6: Build the ProxyConfig response
	proxyConfig := &types.ProxyConfig{
		Version:       version,
		RawConfigDump: rawConfigDump,
		Bootstrap:     summary.Bootstrap,
		Listeners:     summary.Listeners,
		Clusters:      summary.Clusters,
		Endpoints:     endpoints,
		Routes:        summary.Routes,
		RawClusters:   rawClusters,
	}

	s.logger.Debug("successfully retrieved proxy config",
		"namespace", namespace,
		"pod", podName,
		"version", version,
		"listeners", len(summary.Listeners),
		"clusters", len(summary.Clusters),
		"endpoints", len(endpoints),
		"routes", len(summary.Routes))

	return proxyConfig, nil
}

// IsProxyReady checks if the Envoy proxy in the specified pod is ready for configuration requests
func (s *ProxyService) IsProxyReady(ctx context.Context, namespace, podName string) (bool, error) {
	s.logger.Debug("checking proxy readiness", "namespace", namespace, "pod", podName)

	ready, err := s.adminClient.IsIstioProxyReady(ctx, namespace, podName)
	if err != nil {
		s.logger.Error("failed to check proxy readiness", "namespace", namespace, "pod", podName, "error", err)
		return false, fmt.Errorf("failed to check proxy readiness for pod %s/%s: %w", namespace, podName, err)
	}

	s.logger.Debug("proxy readiness check completed", "namespace", namespace, "pod", podName, "ready", ready)
	return ready, nil
}

// GetProxyVersion retrieves the Envoy proxy version for a pod
func (s *ProxyService) GetProxyVersion(ctx context.Context, namespace, podName string) (string, error) {
	s.logger.Debug("retrieving proxy version", "namespace", namespace, "pod", podName)

	version, err := s.adminClient.GetProxyVersion(ctx, namespace, podName)
	if err != nil {
		s.logger.Error("failed to get proxy version", "namespace", namespace, "pod", podName, "error", err)
		return "", fmt.Errorf("failed to get proxy version for pod %s/%s: %w", namespace, podName, err)
	}

	s.logger.Debug("successfully retrieved proxy version", "namespace", namespace, "pod", podName, "version", version)
	return version, nil
}

// ValidateProxyAccess validates that the specified pod has an accessible Envoy proxy
// This method performs basic validation before attempting configuration retrieval
func (s *ProxyService) ValidateProxyAccess(ctx context.Context, namespace, podName string) error {
	s.logger.Debug("validating proxy access", "namespace", namespace, "pod", podName)

	// Check if the proxy is ready (this also validates pod existence and istio-proxy container)
	ready, err := s.adminClient.IsIstioProxyReady(ctx, namespace, podName)
	if err != nil {
		return fmt.Errorf("proxy validation failed for pod %s/%s: %w", namespace, podName, err)
	}

	if !ready {
		return fmt.Errorf("proxy is not ready for pod %s/%s", namespace, podName)
	}

	s.logger.Debug("proxy access validation successful", "namespace", namespace, "pod", podName)
	return nil
}

// MockProxyService provides a mock implementation for testing
type MockProxyService struct {
	GetProxyConfigFunc      func(ctx context.Context, namespace, podName string) (*types.ProxyConfig, error)
	IsProxyReadyFunc        func(ctx context.Context, namespace, podName string) (bool, error)
	GetProxyVersionFunc     func(ctx context.Context, namespace, podName string) (string, error)
	ValidateProxyAccessFunc func(ctx context.Context, namespace, podName string) error
}

// GetProxyConfig mock implementation
func (m *MockProxyService) GetProxyConfig(ctx context.Context, namespace, podName string) (*types.ProxyConfig, error) {
	if m.GetProxyConfigFunc != nil {
		return m.GetProxyConfigFunc(ctx, namespace, podName)
	}
	return &types.ProxyConfig{Version: "mock"}, nil
}

// IsProxyReady mock implementation
func (m *MockProxyService) IsProxyReady(ctx context.Context, namespace, podName string) (bool, error) {
	if m.IsProxyReadyFunc != nil {
		return m.IsProxyReadyFunc(ctx, namespace, podName)
	}
	return true, nil
}

// GetProxyVersion mock implementation
func (m *MockProxyService) GetProxyVersion(ctx context.Context, namespace, podName string) (string, error) {
	if m.GetProxyVersionFunc != nil {
		return m.GetProxyVersionFunc(ctx, namespace, podName)
	}
	return "mock-version", nil
}

// ValidateProxyAccess mock implementation
func (m *MockProxyService) ValidateProxyAccess(ctx context.Context, namespace, podName string) error {
	if m.ValidateProxyAccessFunc != nil {
		return m.ValidateProxyAccessFunc(ctx, namespace, podName)
	}
	return nil
}
