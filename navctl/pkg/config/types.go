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
	"time"
)

// Config represents the navctl configuration file structure
type Config struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       string         `yaml:"kind" json:"kind"`
	Manager    *ManagerConfig `yaml:"manager" json:"manager"`
	Edges      []EdgeConfig   `yaml:"edges" json:"edges"`
	UI         *UIConfig      `yaml:"ui,omitempty" json:"ui,omitempty"`
}

// ManagerConfig holds manager service configuration
type ManagerConfig struct {
	Host           string `yaml:"host,omitempty" json:"host,omitempty"`
	Port           int    `yaml:"port,omitempty" json:"port,omitempty"`
	MaxMessageSize int    `yaml:"maxMessageSize,omitempty" json:"maxMessageSize,omitempty"`
}

// EdgeConfig holds configuration for a single edge service
type EdgeConfig struct {
	Name         string         `yaml:"name" json:"name"`
	ClusterID    string         `yaml:"clusterId" json:"clusterId"`
	Context      string         `yaml:"context,omitempty" json:"context,omitempty"`
	Kubeconfig   string         `yaml:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	SyncInterval int            `yaml:"syncInterval,omitempty" json:"syncInterval,omitempty"`
	LogLevel     string         `yaml:"logLevel,omitempty" json:"logLevel,omitempty"`
	LogFormat    string         `yaml:"logFormat,omitempty" json:"logFormat,omitempty"`
	Metrics      *MetricsConfig `yaml:"metrics,omitempty" json:"metrics,omitempty"`
}

// UIConfig holds UI server configuration
type UIConfig struct {
	Port      int  `yaml:"port,omitempty" json:"port,omitempty"`
	Disabled  bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	NoBrowser bool `yaml:"noBrowser,omitempty" json:"noBrowser,omitempty"`
}

// MetricsConfig holds metrics provider configuration
type MetricsConfig struct {
	Type          string       `yaml:"type" json:"type"`
	Endpoint      string       `yaml:"endpoint" json:"endpoint"`
	QueryInterval int          `yaml:"queryInterval,omitempty" json:"queryInterval,omitempty"`
	Timeout       int          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Auth          *MetricsAuth `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// MetricsAuth holds metrics authentication configuration
type MetricsAuth struct {
	BearerToken     string      `yaml:"bearerToken,omitempty" json:"bearerToken,omitempty"`
	BearerTokenExec *ExecConfig `yaml:"bearerTokenExec,omitempty" json:"bearerTokenExec,omitempty"`
}

// ExecConfig holds configuration for executing commands to get bearer tokens
type ExecConfig struct {
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
	Env     []EnvVar `yaml:"env,omitempty" json:"env,omitempty"`
	Timeout string   `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// EnvVar represents an environment variable for exec commands
type EnvVar struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

// TokenCache represents a cached bearer token with expiration
type TokenCache struct {
	Token     string
	ExpiresAt time.Time
}

// IsExpired returns true if the cached token has expired
func (tc *TokenCache) IsExpired() bool {
	return time.Now().After(tc.ExpiresAt)
}

// DefaultConfig returns a new Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		APIVersion: "navigator.io/v1alpha1",
		Kind:       "NavctlConfig",
		Manager: &ManagerConfig{
			Host:           "localhost",
			Port:           8080,
			MaxMessageSize: 10,
		},
		UI: &UIConfig{
			Port:      8082,
			Disabled:  false,
			NoBrowser: false,
		},
	}
}

// DefaultEdgeConfig returns an EdgeConfig with sensible defaults
func DefaultEdgeConfig() EdgeConfig {
	return EdgeConfig{
		SyncInterval: 30,
		LogLevel:     "info",
		LogFormat:    "text",
	}
}

// DefaultMetricsConfig returns a MetricsConfig with sensible defaults
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Type:          "prometheus",
		QueryInterval: 30,
		Timeout:       10,
	}
}
