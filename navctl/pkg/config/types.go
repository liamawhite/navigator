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

// Config represents the root configuration structure for navctl.
//
// This configuration supports both YAML and JSON formats and allows for
// comprehensive edge-specific configuration including metrics, authentication,
// and Kubernetes contexts. The configuration file enables declarative
// management of Navigator services across multiple clusters.
//
// Example YAML configuration:
//
//	apiVersion: navigator.io/v1alpha1
//	kind: NavctlConfig
//	manager:
//	  host: localhost
//	  port: 8080
//	edges:
//	  - name: production
//	    clusterId: prod-cluster-1
//	    context: prod-context
//	    metrics:
//	      type: prometheus
//	      endpoint: https://prometheus.prod.example.com
//	      auth:
//	        bearerTokenExec:
//	          command: kubectl
//	          args: ["get", "secret", "token", "-o", "jsonpath={.data.token}"]
//	ui:
//	  port: 8082
//	  disabled: false
type Config struct {
	// APIVersion specifies the configuration schema version.
	// Currently supported: "navigator.io/v1alpha1"
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`

	// Kind identifies this as a NavctlConfig.
	// Must be: "NavctlConfig"
	Kind string `yaml:"kind" json:"kind"`

	// Manager contains configuration for the Navigator manager service.
	// The manager coordinates communication between multiple edge services
	// and serves the frontend API.
	Manager *ManagerConfig `yaml:"manager" json:"manager"`

	// Edges contains configuration for each edge service.
	// Each edge connects to a specific Kubernetes cluster and streams
	// cluster state to the manager. Multiple edges enable multi-cluster
	// service discovery and monitoring.
	Edges []EdgeConfig `yaml:"edges" json:"edges"`

	// UI contains configuration for the web UI server.
	// Optional - if omitted, default UI settings will be used.
	UI *UIConfig `yaml:"ui,omitempty" json:"ui,omitempty"`
}

// ManagerConfig holds configuration for the Navigator manager service.
//
// The manager service acts as the central coordination point, maintaining
// bidirectional streaming gRPC connections with edge services and serving
// the frontend API through both gRPC and HTTP gateway endpoints.
//
// Example configuration:
//
//	manager:
//	  host: localhost
//	  port: 8080
//	  maxMessageSize: 10
type ManagerConfig struct {
	// Host specifies the hostname or IP address for the manager service.
	// Default: "localhost"
	// The manager will bind to this address for incoming connections.
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	// Port specifies the gRPC port for the manager service.
	// Default: 8080
	// The HTTP gateway will automatically use port+1 (e.g., 8081).
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// MaxMessageSize specifies the maximum gRPC message size in megabytes.
	// Default: 10
	// Increase this value if you have large service discovery payloads.
	MaxMessageSize int `yaml:"maxMessageSize,omitempty" json:"maxMessageSize,omitempty"`
}

// EdgeConfig holds configuration for a single edge service.
//
// An edge service connects to a specific Kubernetes cluster and streams
// cluster state (services, pods, endpoints) to the manager. It also provides
// on-demand proxy configuration analysis via the Envoy admin API.
//
// Example configuration:
//
//	edges:
//	  - name: production
//	    clusterId: prod-cluster-1
//	    context: prod-context
//	    kubeconfig: /path/to/prod-kubeconfig
//	    syncInterval: 30
//	    logLevel: info
//	    metrics:
//	      type: prometheus
//	      endpoint: https://prometheus.prod.example.com
type EdgeConfig struct {
	// Name is a unique identifier for this edge configuration.
	// Required. Used in logs and UI to distinguish between edges.
	Name string `yaml:"name" json:"name"`

	// ClusterID is a unique identifier for the Kubernetes cluster.
	// Required. Used to identify the cluster in multi-cluster scenarios.
	// Should be unique across all clusters in your Navigator deployment.
	ClusterID string `yaml:"clusterId" json:"clusterId"`

	// Context specifies the kubeconfig context to use for this edge.
	// Optional. If omitted, uses the current context from kubeconfig.
	// Must exist in the specified kubeconfig file.
	Context string `yaml:"context,omitempty" json:"context,omitempty"`

	// Kubeconfig specifies the path to the kubeconfig file.
	// Optional. If omitted, uses the default kubeconfig location (~/.kube/config).
	// Can be an absolute path or relative to the working directory.
	Kubeconfig string `yaml:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`

	// SyncInterval specifies how often to sync cluster state, in seconds.
	// Default: 30
	// Lower values provide more real-time updates but increase load.
	SyncInterval int `yaml:"syncInterval,omitempty" json:"syncInterval,omitempty"`

	// LogLevel specifies the logging level for this edge service.
	// Default: "info"
	// Valid values: "debug", "info", "warn", "error"
	LogLevel string `yaml:"logLevel,omitempty" json:"logLevel,omitempty"`

	// LogFormat specifies the logging format for this edge service.
	// Default: "text"
	// Valid values: "text", "json"
	LogFormat string `yaml:"logFormat,omitempty" json:"logFormat,omitempty"`

	// Metrics contains configuration for metrics collection from this cluster.
	// Optional. If omitted, metrics collection is disabled for this edge.
	Metrics *MetricsConfig `yaml:"metrics,omitempty" json:"metrics,omitempty"`
}

// UIConfig holds configuration for the Navigator web UI server.
//
// The UI server provides a web interface for service discovery, proxy
// configuration viewing, and real-time cluster monitoring across all
// connected edge services.
//
// Example configuration:
//
//	ui:
//	  port: 8082
//	  disabled: false
//	  noBrowser: false
type UIConfig struct {
	// Port specifies the port for the web UI server.
	// Default: 8082
	// The UI will be accessible at http://localhost:<port>
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Disabled determines whether to start the UI server.
	// Default: false
	// Set to true to run navctl without the web interface.
	Disabled bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`

	// NoBrowser determines whether to automatically open a browser.
	// Default: false
	// Set to true to prevent automatic browser launching when starting navctl.
	NoBrowser bool `yaml:"noBrowser,omitempty" json:"noBrowser,omitempty"`
}

// MetricsConfig holds configuration for metrics collection from a cluster.
//
// Navigator supports pluggable metrics providers, with Prometheus being
// the primary implementation. Metrics are collected during cluster state
// synchronization and include service request rates, error rates, and
// latency percentiles.
//
// Example configuration:
//
//	metrics:
//	  type: prometheus
//	  endpoint: https://prometheus.prod.example.com
//	  queryInterval: 30
//	  timeout: 10
//	  auth:
//	    bearerTokenExec:
//	      command: kubectl
//	      args: ["get", "secret", "prometheus-token", "-o", "jsonpath={.data.token}"]
type MetricsConfig struct {
	// Type specifies the metrics provider type.
	// Currently supported: "prometheus"
	// Required when metrics collection is enabled.
	Type string `yaml:"type" json:"type"`

	// Endpoint specifies the URL for the metrics provider.
	// Required. For Prometheus, this should be the base URL (e.g., https://prometheus.example.com).
	// The endpoint should be accessible from where navctl is running.
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// QueryInterval specifies how often to query for metrics, in seconds.
	// Default: 30
	// Lower values provide more real-time metrics but increase load on the metrics provider.
	QueryInterval int `yaml:"queryInterval,omitempty" json:"queryInterval,omitempty"`

	// Timeout specifies the timeout for metrics queries, in seconds.
	// Default: 10
	// Increase this value if your metrics provider has high latency.
	Timeout int `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Auth contains authentication configuration for the metrics provider.
	// Optional. If omitted, no authentication is used.
	// Supports static bearer tokens and dynamic token generation via exec commands.
	Auth *MetricsAuth `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// MetricsAuth holds authentication configuration for metrics providers.
//
// Supports both static bearer tokens and dynamic token generation through
// command execution, similar to Kubernetes client authentication.
//
// Example configurations:
//
// Static token:
//
//	auth:
//	  bearerToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
//
// Dynamic token via kubectl:
//
//	auth:
//	  bearerTokenExec:
//	    command: kubectl
//	    args: ["get", "secret", "prometheus-token", "-o", "jsonpath={.data.token}"]
//	    timeout: 30s
type MetricsAuth struct {
	// BearerToken specifies a static bearer token for authentication.
	// Optional. Mutually exclusive with BearerTokenExec.
	// Use this for long-lived tokens or when you want to manage token rotation externally.
	BearerToken string `yaml:"bearerToken,omitempty" json:"bearerToken,omitempty"`

	// BearerTokenExec specifies a command to execute to obtain a bearer token.
	// Optional. Mutually exclusive with BearerToken.
	// Tokens are cached for 15 minutes to avoid excessive command execution.
	// Use this for dynamic token generation, similar to Kubernetes exec authentication.
	BearerTokenExec *ExecConfig `yaml:"bearerTokenExec,omitempty" json:"bearerTokenExec,omitempty"`
}

// ExecConfig holds configuration for executing commands to get bearer tokens.
//
// This configuration enables dynamic token generation through command execution,
// similar to the exec authentication mechanism used in Kubernetes kubeconfig files.
// Commands are executed in a secure environment with configurable timeouts and
// environment variables.
//
// Example configuration:
//
//	bearerTokenExec:
//	  command: kubectl
//	  args: ["get", "secret", "prometheus-token", "-o", "jsonpath={.data.token}"]
//	  timeout: 30s
//	  env:
//	    - name: KUBECONFIG
//	      value: /path/to/kubeconfig
type ExecConfig struct {
	// Command specifies the executable to run for token generation.
	// Required. Should be an absolute path or command available in PATH.
	// Common examples: "kubectl", "gcloud", "aws", "/usr/local/bin/get-token"
	Command string `yaml:"command" json:"command"`

	// Args specifies the command-line arguments to pass to the command.
	// Optional. Use this to specify subcommands and parameters.
	// Example: ["get", "secret", "token", "-o", "jsonpath={.data.token}"]
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`

	// Env specifies additional environment variables for the command.
	// Optional. Use this to set context-specific environment variables.
	// These are added to the existing environment, not replacing it.
	Env []EnvVar `yaml:"env,omitempty" json:"env,omitempty"`

	// Timeout specifies the maximum duration to wait for command completion.
	// Default: "30s"
	// Valid formats: "30s", "5m", "1h" (any duration parseable by time.ParseDuration)
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// EnvVar represents an environment variable for exec commands.
//
// Environment variables are passed to the command execution environment
// and can be used to provide context-specific configuration such as
// authentication credentials, API endpoints, or other runtime parameters.
//
// Example usage in exec configuration:
//
//	env:
//	  - name: KUBECONFIG
//	    value: /path/to/specific/kubeconfig
//	  - name: AWS_PROFILE
//	    value: production
//	  - name: PROMETHEUS_URL
//	    value: https://prometheus.example.com
type EnvVar struct {
	// Name is the environment variable name.
	// Required. Should follow standard environment variable naming conventions.
	// Example: "KUBECONFIG", "AWS_PROFILE", "PROMETHEUS_URL"
	Name string `yaml:"name" json:"name"`

	// Value is the environment variable value.
	// Required. The value to set for the named environment variable.
	// Can contain absolute paths, URLs, or any string value.
	Value string `yaml:"value" json:"value"`
}

// TokenCache represents a cached bearer token with expiration.
//
// Token caching is used to avoid excessive command execution when using
// dynamic token generation via ExecConfig. Tokens are cached for a configurable
// duration (default 15 minutes) to balance security and performance.
//
// The cache is automatically managed by the TokenExecutor and provides:
// - Automatic expiration checking
// - Thread-safe access
// - Transparent token refresh
type TokenCache struct {
	// Token is the cached bearer token value.
	// This is the raw token string returned by the exec command.
	Token string

	// ExpiresAt is the time when this cached token expires.
	// After this time, a new token will be generated via command execution.
	ExpiresAt time.Time
}

// IsExpired returns true if the cached token has expired.
//
// This method is used internally by the TokenExecutor to determine
// whether a cached token needs to be refreshed. The comparison is
// based on the current time versus the stored expiration time.
func (tc *TokenCache) IsExpired() bool {
	return time.Now().After(tc.ExpiresAt)
}

// DefaultConfig returns a new Config with sensible defaults.
//
// This function creates a minimal configuration suitable for local development
// with Navigator. The returned configuration includes:
// - Standard API version and kind
// - Manager service on localhost:8080 with 10MB message limit
// - UI server on port 8082 with browser auto-open enabled
// - Empty edges slice (must be configured separately)
//
// Use this as a starting point for programmatic configuration creation
// or when config file loading fails and fallback defaults are needed.
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

// DefaultEdgeConfig returns an EdgeConfig with sensible defaults.
//
// This function provides default values for optional edge configuration
// fields, suitable for most deployment scenarios. The defaults include:
// - 30-second sync interval for cluster state updates
// - Info-level logging for operational visibility
// - Text log format for human readability
//
// Required fields (Name, ClusterID) must still be set by the caller.
// Use this function when creating edge configurations programmatically
// or when applying defaults to partially-specified configurations.
func DefaultEdgeConfig() EdgeConfig {
	return EdgeConfig{
		SyncInterval: 30,
		LogLevel:     "info",
		LogFormat:    "text",
	}
}

// DefaultMetricsConfig returns a MetricsConfig with sensible defaults.
//
// This function provides default values for optional metrics configuration
// fields, optimized for typical Prometheus deployments. The defaults include:
// - Prometheus as the metrics provider type
// - 30-second query interval for real-time metrics without overwhelming the provider
// - 10-second timeout for metrics queries to handle network latency
//
// The Endpoint field must still be set by the caller as it's environment-specific.
// Use this function when creating metrics configurations programmatically
// or when applying defaults to partially-specified configurations.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Type:          "prometheus",
		QueryInterval: 30,
		Timeout:       10,
	}
}
