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
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	yamlContent := `
apiVersion: navigator.io/v1alpha1
kind: NavctlConfig
manager:
  host: testhost
  port: 9090
edges:
  - name: test-edge
    clusterId: test-cluster
    metrics:
      type: prometheus
      endpoint: http://prometheus:9090
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0600)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager, err := NewManager(configFile, logger)
	require.NoError(t, err)

	assert.NotNil(t, manager)
	assert.Equal(t, "testhost", manager.config.Manager.Host)
	assert.Equal(t, 9090, manager.config.Manager.Port)
}

func TestManager_GetManagerConfig(t *testing.T) {
	config := &Config{
		Manager: &ManagerConfig{
			Host:           "testhost",
			Port:           9090,
			MaxMessageSize: 20,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}

	managerCfg := manager.GetManagerConfig()
	assert.Equal(t, 9090, managerCfg.Port)
	assert.Equal(t, 20, managerCfg.MaxMessageSize)
	assert.Equal(t, "info", managerCfg.LogLevel)  // Default value
	assert.Equal(t, "text", managerCfg.LogFormat) // Default value
}

func TestManager_GetEdgeConfig(t *testing.T) {
	config := &Config{
		Manager: &ManagerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Edges: []EdgeConfig{
			{
				Name:         "test-edge",
				ClusterID:    "test-cluster",
				Context:      "test-context",
				SyncInterval: 45,
				LogLevel:     "debug",
				LogFormat:    "json",
				Metrics: &MetricsConfig{
					Type:          "prometheus",
					Endpoint:      "http://prometheus:9090",
					QueryInterval: 60,
					Timeout:       15,
					Auth: &MetricsAuth{
						BearerToken: "test-token",
					},
				},
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}

	edgeCfg, err := manager.GetEdgeConfig("test-edge", "", "")
	require.NoError(t, err)

	assert.Equal(t, "test-cluster", edgeCfg.ClusterID)
	assert.Equal(t, "localhost:8080", edgeCfg.ManagerEndpoint)
	assert.Equal(t, 45, edgeCfg.SyncInterval)
	assert.Equal(t, "debug", edgeCfg.LogLevel)
	assert.Equal(t, "json", edgeCfg.LogFormat)
	assert.True(t, edgeCfg.MetricsConfig.Enabled)
	assert.Equal(t, "prometheus", string(edgeCfg.MetricsConfig.Type))
	assert.Equal(t, "http://prometheus:9090", edgeCfg.MetricsConfig.Endpoint)
	assert.Equal(t, "test-token", edgeCfg.MetricsConfig.BearerToken)
}

func TestManager_GetEdgeConfig_GlobalOverrides(t *testing.T) {
	config := &Config{
		Manager: &ManagerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Edges: []EdgeConfig{
			{
				Name:      "test-edge",
				ClusterID: "test-cluster",
				LogLevel:  "info",
				LogFormat: "text",
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}

	edgeCfg, err := manager.GetEdgeConfig("test-edge", "debug", "json")
	require.NoError(t, err)

	// Global overrides should take precedence
	assert.Equal(t, "debug", edgeCfg.LogLevel)
	assert.Equal(t, "json", edgeCfg.LogFormat)
}

func TestManager_GetEdgeConfig_NotFound(t *testing.T) {
	config := &Config{
		Edges: []EdgeConfig{
			{
				Name:      "existing-edge",
				ClusterID: "test-cluster",
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}

	_, err := manager.GetEdgeConfig("non-existent-edge", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "edge configuration not found")
}

func TestManager_GetEdgeNames(t *testing.T) {
	config := &Config{
		Edges: []EdgeConfig{
			{Name: "edge1", ClusterID: "cluster1"},
			{Name: "edge2", ClusterID: "cluster2"},
			{Name: "edge3", ClusterID: "cluster3"},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}

	names := manager.GetEdgeNames()
	assert.Equal(t, []string{"edge1", "edge2", "edge3"}, names)
}

func TestManager_GetEdgeKubeContext(t *testing.T) {
	config := &Config{
		Edges: []EdgeConfig{
			{Name: "edge1", ClusterID: "cluster1", Context: "context1"},
			{Name: "edge2", ClusterID: "cluster2", Context: "context2"},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}

	context, err := manager.GetEdgeKubeContext("edge1")
	require.NoError(t, err)
	assert.Equal(t, "context1", context)

	_, err = manager.GetEdgeKubeContext("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "edge not found")
}

func TestManager_GetEdgeKubeconfig(t *testing.T) {
	config := &Config{
		Edges: []EdgeConfig{
			{Name: "edge1", ClusterID: "cluster1", Kubeconfig: "/path/to/config1"},
			{Name: "edge2", ClusterID: "cluster2", Kubeconfig: "/path/to/config2"},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	manager := &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}

	kubeconfig, err := manager.GetEdgeKubeconfig("edge1")
	require.NoError(t, err)
	assert.Equal(t, "/path/to/config1", kubeconfig)

	_, err = manager.GetEdgeKubeconfig("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "edge not found")
}

func TestManager_ValidateEdges(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid edges",
			config: &Config{
				Edges: []EdgeConfig{
					{Name: "edge1", ClusterID: "cluster1"},
					{Name: "edge2", ClusterID: "cluster2"},
				},
			},
			wantErr: false,
		},
		{
			name: "no edges",
			config: &Config{
				Edges: []EdgeConfig{},
			},
			wantErr:     true,
			errContains: "no edges configured",
		},
		{
			name: "duplicate cluster IDs",
			config: &Config{
				Edges: []EdgeConfig{
					{Name: "edge1", ClusterID: "duplicate"},
					{Name: "edge2", ClusterID: "duplicate"},
				},
			},
			wantErr:     true,
			errContains: "duplicate cluster ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			manager := &Manager{
				config:        tt.config,
				tokenExecutor: NewTokenExecutor(logger),
				logger:        logger,
			}

			err := manager.ValidateEdges()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_HasEdges(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Config with edges
	configWithEdges := &Config{
		Edges: []EdgeConfig{
			{Name: "edge1", ClusterID: "cluster1"},
		},
	}
	manager := &Manager{
		config:        configWithEdges,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}
	assert.True(t, manager.HasEdges())

	// Config without edges
	configWithoutEdges := &Config{
		Edges: []EdgeConfig{},
	}
	manager = &Manager{
		config:        configWithoutEdges,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}
	assert.False(t, manager.HasEdges())
}

func TestManager_IsMultiEdge(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Config with multiple edges
	configMultiEdge := &Config{
		Edges: []EdgeConfig{
			{Name: "edge1", ClusterID: "cluster1"},
			{Name: "edge2", ClusterID: "cluster2"},
		},
	}
	manager := &Manager{
		config:        configMultiEdge,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}
	assert.True(t, manager.IsMultiEdge())

	// Config with single edge
	configSingleEdge := &Config{
		Edges: []EdgeConfig{
			{Name: "edge1", ClusterID: "cluster1"},
		},
	}
	manager = &Manager{
		config:        configSingleEdge,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}
	assert.False(t, manager.IsMultiEdge())
}
