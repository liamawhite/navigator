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

package config

import (
	"fmt"
	"log/slog"

	edgeConfig "github.com/liamawhite/navigator/edge/pkg/config"
	"github.com/liamawhite/navigator/edge/pkg/metrics"
	managerConfig "github.com/liamawhite/navigator/manager/pkg/config"
)

// Manager encapsulates configuration management for navctl
type Manager struct {
	config        *Config
	tokenExecutor *TokenExecutor
	logger        *slog.Logger
}

// NewManager creates a new configuration manager
func NewManager(configPath string, logger *slog.Logger) (*Manager, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Perform post-load processing
	config.PostLoad()

	return &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}, nil
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// GetManagerConfig returns a manager configuration
func (m *Manager) GetManagerConfig() *managerConfig.Config {
	return &managerConfig.Config{
		Port:           m.config.Manager.Port,
		LogLevel:       "info", // Will be overridden by CLI flags
		LogFormat:      "text", // Will be overridden by CLI flags
		MaxMessageSize: m.config.Manager.MaxMessageSize,
	}
}

// GetEdgeConfig returns an edge configuration for the specified edge name
func (m *Manager) GetEdgeConfig(edgeName string, globalLogLevel, globalLogFormat string) (*edgeConfig.Config, error) {
	var edge *EdgeConfig
	for i := range m.config.Edges {
		if m.config.Edges[i].Name == edgeName {
			edge = &m.config.Edges[i]
			break
		}
	}

	if edge == nil {
		return nil, fmt.Errorf("edge configuration not found: %s", edgeName)
	}

	// Build metrics config
	metricsConfig := metrics.Config{
		Enabled:       edge.Metrics != nil,
		Type:          metrics.ProviderTypeNone,
		Endpoint:      "",
		QueryInterval: 30,
		Timeout:       10,
		BearerToken:   "",
	}

	if edge.Metrics != nil {
		metricsConfig.Type = metrics.ProviderType(edge.Metrics.Type)
		metricsConfig.Endpoint = edge.Metrics.Endpoint
		metricsConfig.QueryInterval = edge.Metrics.QueryInterval
		metricsConfig.Timeout = edge.Metrics.Timeout

		// Get bearer token
		if edge.Metrics.Auth != nil {
			token, err := m.tokenExecutor.GetBearerToken(edgeName, edge.Metrics.Auth)
			if err != nil {
				return nil, fmt.Errorf("failed to get bearer token for edge %s: %w", edgeName, err)
			}
			metricsConfig.BearerToken = token
		}
	}

	// Use global log settings if provided, otherwise use edge-specific settings
	logLevel := edge.LogLevel
	if globalLogLevel != "" {
		logLevel = globalLogLevel
	}

	logFormat := edge.LogFormat
	if globalLogFormat != "" {
		logFormat = globalLogFormat
	}

	return &edgeConfig.Config{
		ClusterID:       edge.ClusterID,
		ManagerEndpoint: fmt.Sprintf("%s:%d", m.config.Manager.Host, m.config.Manager.Port),
		SyncInterval:    edge.SyncInterval,
		KubeconfigPath:  edge.Kubeconfig,
		LogLevel:        logLevel,
		LogFormat:       logFormat,
		MaxMessageSize:  m.config.Manager.MaxMessageSize,
		MetricsConfig:   metricsConfig,
	}, nil
}

// GetEdgeNames returns a list of all configured edge names
func (m *Manager) GetEdgeNames() []string {
	names := make([]string, len(m.config.Edges))
	for i, edge := range m.config.Edges {
		names[i] = edge.Name
	}
	return names
}

// GetEdgeKubeContext returns the kubeconfig context for the specified edge
func (m *Manager) GetEdgeKubeContext(edgeName string) (string, error) {
	for _, edge := range m.config.Edges {
		if edge.Name == edgeName {
			return edge.Context, nil
		}
	}
	return "", fmt.Errorf("edge not found: %s", edgeName)
}

// GetEdgeKubeconfig returns the kubeconfig path for the specified edge
func (m *Manager) GetEdgeKubeconfig(edgeName string) (string, error) {
	for _, edge := range m.config.Edges {
		if edge.Name == edgeName {
			return edge.Kubeconfig, nil
		}
	}
	return "", fmt.Errorf("edge not found: %s", edgeName)
}

// RefreshTokens forces a refresh of all cached bearer tokens
func (m *Manager) RefreshTokens() error {
	for _, edge := range m.config.Edges {
		if edge.Metrics != nil && edge.Metrics.Auth != nil {
			_, err := m.tokenExecutor.RefreshToken(edge.Name, edge.Metrics.Auth)
			if err != nil {
				return fmt.Errorf("failed to refresh token for edge %s: %w", edge.Name, err)
			}
		}
	}
	return nil
}

// GetUIConfig returns UI configuration
func (m *Manager) GetUIConfig() *UIConfig {
	return m.config.UI
}

// ValidateEdges validates that all edge configurations are valid
func (m *Manager) ValidateEdges() error {
	if len(m.config.Edges) == 0 {
		return fmt.Errorf("no edges configured")
	}

	// Check for duplicate cluster IDs
	clusterIDs := make(map[string]string)
	for _, edge := range m.config.Edges {
		if existingEdge, exists := clusterIDs[edge.ClusterID]; exists {
			return fmt.Errorf("duplicate cluster ID %s used by edges %s and %s", edge.ClusterID, existingEdge, edge.Name)
		}
		clusterIDs[edge.ClusterID] = edge.Name
	}

	return nil
}

// GetTokenCacheStats returns statistics about the token cache
func (m *Manager) GetTokenCacheStats() map[string]interface{} {
	return m.tokenExecutor.GetCacheStats()
}

// ClearTokenCache clears all cached tokens
func (m *Manager) ClearTokenCache() {
	m.tokenExecutor.ClearCache()
}

// HasEdges returns true if there are any edges configured
func (m *Manager) HasEdges() bool {
	return len(m.config.Edges) > 0
}

// IsMultiEdge returns true if there are multiple edges configured
func (m *Manager) IsMultiEdge() bool {
	return len(m.config.Edges) > 1
}

// Close cleans up resources used by the manager
func (m *Manager) Close() {
	if m.tokenExecutor != nil {
		m.tokenExecutor.Close()
	}
}
