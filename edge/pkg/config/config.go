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
	"flag"
	"fmt"

	"github.com/liamawhite/navigator/edge/pkg/metrics"
)

// Config holds the configuration for the edge service
type Config struct {
	ManagerEndpoint string
	SyncInterval    int
	KubeconfigPath  string
	LogLevel        string
	LogFormat       string
	MaxMessageSize  int // Maximum gRPC message size in MB
	MetricsConfig   metrics.Config
}

// ParseFlags parses command line flags and returns a Config
func ParseFlags() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.ManagerEndpoint, "manager-endpoint", "", "gRPC endpoint of the manager service (required)")
	flag.IntVar(&config.SyncInterval, "sync-interval", 30, "Interval between cluster state sync operations (in seconds)")
	flag.StringVar(&config.KubeconfigPath, "kubeconfig", "", "Path to kubeconfig file (uses in-cluster config if empty)")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&config.LogFormat, "log-format", "text", "Log format (text, json)")
	flag.IntVar(&config.MaxMessageSize, "max-message-size", 10, "Maximum gRPC message size in MB")

	// Metrics configuration
	flag.BoolVar(&config.MetricsConfig.Enabled, "metrics-enabled", false, "Enable metrics collection")
	flag.StringVar(&config.MetricsConfig.Endpoint, "metrics-endpoint", "", "Metrics provider endpoint URL")
	flag.StringVar((*string)(&config.MetricsConfig.Type), "metrics-type", "none", "Metrics provider type (none, prometheus)")
	flag.IntVar(&config.MetricsConfig.QueryInterval, "metrics-query-interval", 30, "Metrics query interval in seconds")
	flag.IntVar(&config.MetricsConfig.Timeout, "metrics-timeout", 10, "Metrics query timeout in seconds")
	flag.StringVar(&config.MetricsConfig.BearerToken, "metrics-auth-bearer", "", "Bearer token for metrics provider authentication")

	flag.Parse()

	return config, config.Validate()
}

// Validate checks that required configuration is provided
func (c *Config) Validate() error {
	if c.ManagerEndpoint == "" {
		return fmt.Errorf("manager-endpoint is required")
	}

	if c.SyncInterval <= 0 {
		return fmt.Errorf("sync-interval must be positive")
	}

	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return fmt.Errorf("log-level must be one of: debug, info, warn, error")
	}

	if c.LogFormat != "text" && c.LogFormat != "json" {
		return fmt.Errorf("log-format must be one of: text, json")
	}

	if c.MaxMessageSize <= 0 {
		return fmt.Errorf("max-message-size must be greater than 0")
	}

	// Validate metrics configuration
	if err := c.MetricsConfig.Validate(); err != nil {
		return fmt.Errorf("metrics configuration error: %w", err)
	}

	return nil
}

// GetManagerEndpoint returns the manager endpoint
func (c *Config) GetManagerEndpoint() string {
	return c.ManagerEndpoint
}

// GetSyncInterval returns the sync interval in seconds
func (c *Config) GetSyncInterval() int {
	return c.SyncInterval
}

// GetMaxMessageSize returns the maximum gRPC message size in bytes
func (c *Config) GetMaxMessageSize() int {
	return c.MaxMessageSize * 1024 * 1024 // Convert MB to bytes
}

// GetMetricsConfig returns the metrics configuration
func (c *Config) GetMetricsConfig() metrics.Config {
	return c.MetricsConfig
}
