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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/homedir"
)

const (
	defaultConfigFilename = "navctl-config.yaml"
)

// LoadConfig loads a configuration from a file path or discovers it automatically
func LoadConfig(configPath string) (*Config, error) {
	var filePath string
	var err error

	if configPath != "" {
		// Use specified config path
		filePath = configPath
	} else {
		// Try to discover config file
		filePath, err = discoverConfigFile()
		if err != nil {
			return nil, fmt.Errorf("failed to discover config file: %w", err)
		}
		if filePath == "" {
			// No config file found, return default config
			return DefaultConfig(), nil
		}
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file does not exist: %s", filePath)
		}
		return nil, fmt.Errorf("cannot access config file: %w", err)
	}

	// Read file contents - Note: This reads user-provided config files
	// File path is validated to exist before reading
	data, err := os.ReadFile(filePath) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Parse the config file
	config, err := parseConfig(data, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filePath, err)
	}

	// Apply defaults and validate
	if err := applyDefaultsAndValidate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// discoverConfigFile looks for config files in standard locations
func discoverConfigFile() (string, error) {
	// Search paths in order of preference
	searchPaths := []string{
		"./navctl-config.yaml",
		"./navctl-config.yml",
		"./navctl-config.json",
	}

	// Add home directory paths
	if home := homedir.HomeDir(); home != "" {
		homePaths := []string{
			filepath.Join(home, ".navigator", "config.yaml"),
			filepath.Join(home, ".navigator", "config.yml"),
			filepath.Join(home, ".navigator", "config.json"),
			filepath.Join(home, ".navigator", defaultConfigFilename),
		}
		searchPaths = append(searchPaths, homePaths...)
	}

	// Check each path
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// No config file found
	return "", nil
}

// parseConfig parses config data based on file extension
func parseConfig(data []byte, filePath string) (*Config, error) {
	config := &Config{}

	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, config); err != nil {
			if jsonErr := json.Unmarshal(data, config); jsonErr != nil {
				return nil, fmt.Errorf("failed to parse as YAML (%v) or JSON (%v)", err, jsonErr)
			}
		}
	}

	return config, nil
}

// applyDefaultsAndValidate applies defaults to config and validates it
func applyDefaultsAndValidate(config *Config) error {
	// Set defaults for missing values
	if config.APIVersion == "" {
		config.APIVersion = "navigator.io/v1alpha1"
	}
	if config.Kind == "" {
		config.Kind = "NavctlConfig"
	}

	// Apply manager defaults
	if config.Manager == nil {
		config.Manager = &ManagerConfig{}
	}
	if config.Manager.Host == "" {
		config.Manager.Host = "localhost"
	}
	if config.Manager.Port == 0 {
		config.Manager.Port = 8080
	}
	if config.Manager.MaxMessageSize == 0 {
		config.Manager.MaxMessageSize = 10
	}

	// Apply UI defaults
	if config.UI == nil {
		config.UI = &UIConfig{}
	}
	if config.UI.Port == 0 {
		config.UI.Port = 8082
	}

	// Apply edge defaults and validate
	for i := range config.Edges {
		edge := &config.Edges[i]

		// Apply defaults
		if edge.SyncInterval == 0 {
			edge.SyncInterval = 30
		}
		if edge.LogLevel == "" {
			edge.LogLevel = "info"
		}
		if edge.LogFormat == "" {
			edge.LogFormat = "text"
		}

		// Apply metrics defaults
		if edge.Metrics != nil {
			if edge.Metrics.Type == "" {
				edge.Metrics.Type = "prometheus"
			}
			if edge.Metrics.QueryInterval == 0 {
				edge.Metrics.QueryInterval = 30
			}
			if edge.Metrics.Timeout == 0 {
				edge.Metrics.Timeout = 10
			}
		}

		// Validate metrics configuration
		if edge.Metrics != nil {
			if edge.Metrics.Endpoint == "" {
				return fmt.Errorf("edge %d: metrics endpoint is required when metrics is configured", i)
			}

			// Validate auth configuration
			if edge.Metrics.Auth != nil {
				if edge.Metrics.Auth.BearerToken != "" && edge.Metrics.Auth.BearerTokenExec != nil {
					return fmt.Errorf("edge %d: cannot specify both bearerToken and bearerTokenExec", i)
				}

				if edge.Metrics.Auth.BearerTokenExec != nil {
					if edge.Metrics.Auth.BearerTokenExec.Command == "" {
						return fmt.Errorf("edge %d: bearerTokenExec command is required", i)
					}
				}
			}
		}

		// Validate log level
		validLogLevels := []string{"debug", "info", "warn", "error"}
		validLevel := slices.Contains(validLogLevels, edge.LogLevel)
		if !validLevel {
			return fmt.Errorf("edge %d: invalid log level %s, must be one of: %v", i, edge.LogLevel, validLogLevels)
		}

		// Validate log format
		validLogFormats := []string{"text", "json"}
		validFormat := slices.Contains(validLogFormats, edge.LogFormat)
		if !validFormat {
			return fmt.Errorf("edge %d: invalid log format %s, must be one of: %v", i, edge.LogFormat, validLogFormats)
		}
	}

	return nil
}

// expandEnvVars expands environment variables in strings using ${VAR} or $VAR syntax
func expandEnvVars(s string) string {
	return os.ExpandEnv(s)
}

// expandConfigEnvVars recursively expands environment variables in config
func (c *Config) expandEnvVars() {
	// Expand manager config
	if c.Manager != nil {
		c.Manager.Host = expandEnvVars(c.Manager.Host)
	}

	// Expand edge configs
	for i := range c.Edges {
		edge := &c.Edges[i]
		edge.Context = expandEnvVars(edge.Context)
		edge.Kubeconfig = expandEnvVars(edge.Kubeconfig)

		if edge.Metrics != nil {
			edge.Metrics.Endpoint = expandEnvVars(edge.Metrics.Endpoint)

			if edge.Metrics.Auth != nil {
				edge.Metrics.Auth.BearerToken = expandEnvVars(edge.Metrics.Auth.BearerToken)

				if edge.Metrics.Auth.BearerTokenExec != nil {
					exec := edge.Metrics.Auth.BearerTokenExec
					exec.Command = expandEnvVars(exec.Command)
					for j := range exec.Args {
						exec.Args[j] = expandEnvVars(exec.Args[j])
					}
					for j := range exec.Env {
						exec.Env[j].Value = expandEnvVars(exec.Env[j].Value)
					}
				}
			}
		}
	}
}

// PostLoad performs post-loading processing like environment variable expansion
func (c *Config) PostLoad() {
	c.expandEnvVars()
}
