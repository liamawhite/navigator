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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig_YAML(t *testing.T) {
	yamlContent := `
apiVersion: navigator.io/v1alpha1
kind: NavctlConfig
manager:
  host: testhost
  port: 9090
edges:
  - name: test-edge
    clusterId: test-cluster
    context: test-context
    metrics:
      type: prometheus
      endpoint: http://prometheus:9090
      auth:
        bearerToken: test-token
`

	config, err := parseConfig([]byte(yamlContent), "test.yaml")
	require.NoError(t, err)

	assert.Equal(t, "navigator.io/v1alpha1", config.APIVersion)
	assert.Equal(t, "NavctlConfig", config.Kind)
	assert.Equal(t, "testhost", config.Manager.Host)
	assert.Equal(t, 9090, config.Manager.Port)
	assert.Len(t, config.Edges, 1)
	assert.Equal(t, "test-edge", config.Edges[0].Name)
	assert.Equal(t, "test-cluster", config.Edges[0].ClusterID)
	assert.Equal(t, "test-context", config.Edges[0].Context)
	assert.Equal(t, "prometheus", config.Edges[0].Metrics.Type)
	assert.Equal(t, "http://prometheus:9090", config.Edges[0].Metrics.Endpoint)
	assert.Equal(t, "test-token", config.Edges[0].Metrics.Auth.BearerToken)
}

func TestParseConfig_JSON(t *testing.T) {
	jsonContent := `{
  "apiVersion": "navigator.io/v1alpha1",
  "kind": "NavctlConfig",
  "manager": {
    "host": "testhost",
    "port": 9090
  },
  "edges": [
    {
      "name": "test-edge",
      "clusterId": "test-cluster",
      "context": "test-context",
      "metrics": {
        "type": "prometheus",
        "endpoint": "http://prometheus:9090",
        "auth": {
          "bearerToken": "test-token"
        }
      }
    }
  ]
}`

	config, err := parseConfig([]byte(jsonContent), "test.json")
	require.NoError(t, err)

	assert.Equal(t, "navigator.io/v1alpha1", config.APIVersion)
	assert.Equal(t, "NavctlConfig", config.Kind)
	assert.Equal(t, "testhost", config.Manager.Host)
	assert.Equal(t, 9090, config.Manager.Port)
	assert.Len(t, config.Edges, 1)
	assert.Equal(t, "test-edge", config.Edges[0].Name)
}

func TestApplyDefaultsAndValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name:      "test-edge",
						ClusterID: "test-cluster",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing edge name",
			config: &Config{
				Edges: []EdgeConfig{
					{
						ClusterID: "test-cluster",
					},
				},
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "missing cluster ID",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name: "test-edge",
					},
				},
			},
			wantErr:     true,
			errContains: "clusterId is required",
		},
		{
			name: "duplicate edge names",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name:      "duplicate",
						ClusterID: "cluster1",
					},
					{
						Name:      "duplicate",
						ClusterID: "cluster2",
					},
				},
			},
			wantErr:     true,
			errContains: "duplicate edge name",
		},
		{
			name: "invalid log level",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name:      "test-edge",
						ClusterID: "test-cluster",
						LogLevel:  "invalid",
					},
				},
			},
			wantErr:     true,
			errContains: "invalid log level",
		},
		{
			name: "invalid log format",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name:      "test-edge",
						ClusterID: "test-cluster",
						LogFormat: "invalid",
					},
				},
			},
			wantErr:     true,
			errContains: "invalid log format",
		},
		{
			name: "metrics without endpoint",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name:      "test-edge",
						ClusterID: "test-cluster",
						Metrics: &MetricsConfig{
							Type: "prometheus",
						},
					},
				},
			},
			wantErr:     true,
			errContains: "metrics endpoint is required",
		},
		{
			name: "both bearer token and exec",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name:      "test-edge",
						ClusterID: "test-cluster",
						Metrics: &MetricsConfig{
							Type:     "prometheus",
							Endpoint: "http://prometheus:9090",
							Auth: &MetricsAuth{
								BearerToken: "token",
								BearerTokenExec: &ExecConfig{
									Command: "kubectl",
								},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "cannot specify both bearerToken and bearerTokenExec",
		},
		{
			name: "exec without command",
			config: &Config{
				Edges: []EdgeConfig{
					{
						Name:      "test-edge",
						ClusterID: "test-cluster",
						Metrics: &MetricsConfig{
							Type:     "prometheus",
							Endpoint: "http://prometheus:9090",
							Auth: &MetricsAuth{
								BearerTokenExec: &ExecConfig{},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "bearerTokenExec command is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyDefaultsAndValidate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				
				// Check defaults were applied
				assert.Equal(t, "navigator.io/v1alpha1", tt.config.APIVersion)
				assert.Equal(t, "NavctlConfig", tt.config.Kind)
				assert.Equal(t, "localhost", tt.config.Manager.Host)
				assert.Equal(t, 8080, tt.config.Manager.Port)
				assert.Equal(t, 10, tt.config.Manager.MaxMessageSize)
				assert.Equal(t, 8082, tt.config.UI.Port)
				
				for _, edge := range tt.config.Edges {
					assert.Equal(t, 30, edge.SyncInterval)
					assert.Equal(t, "info", edge.LogLevel)
					assert.Equal(t, "text", edge.LogFormat)
					
					if edge.Metrics != nil {
						assert.Equal(t, "prometheus", edge.Metrics.Type)
						assert.Equal(t, 30, edge.Metrics.QueryInterval)
						assert.Equal(t, 10, edge.Metrics.Timeout)
					}
				}
			}
		})
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	_, err := LoadConfig("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestLoadConfig_WithFile(t *testing.T) {
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
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	require.NoError(t, err)

	assert.Equal(t, "testhost", config.Manager.Host)
	assert.Equal(t, 9090, config.Manager.Port)
	assert.Len(t, config.Edges, 1)
	assert.Equal(t, "test-edge", config.Edges[0].Name)
}

func TestLoadConfig_EmptyPath_ReturnsDefault(t *testing.T) {
	// Save current directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Create temporary directory without config files
	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	config, err := LoadConfig("")
	require.NoError(t, err)

	// Should return default config
	assert.Equal(t, "navigator.io/v1alpha1", config.APIVersion)
	assert.Equal(t, "NavctlConfig", config.Kind)
	assert.Equal(t, "localhost", config.Manager.Host)
	assert.Equal(t, 8080, config.Manager.Port)
}

func TestExpandEnvVars(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_HOST", "envhost")
	defer os.Unsetenv("TEST_HOST")

	config := &Config{
		Manager: &ManagerConfig{
			Host: "${TEST_HOST}",
		},
		Edges: []EdgeConfig{
			{
				ClusterID: "$TEST_HOST-cluster",
				Metrics: &MetricsConfig{
					Endpoint: "http://${TEST_HOST}:9090",
					Auth: &MetricsAuth{
						BearerToken: "${TEST_HOST}-token",
					},
				},
			},
		},
	}

	config.expandEnvVars()

	assert.Equal(t, "envhost", config.Manager.Host)
	assert.Equal(t, "envhost-cluster", config.Edges[0].ClusterID)
	assert.Equal(t, "http://envhost:9090", config.Edges[0].Metrics.Endpoint)
	assert.Equal(t, "envhost-token", config.Edges[0].Metrics.Auth.BearerToken)
}