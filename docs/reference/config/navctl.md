# navctl Configuration Reference

This document describes the configuration file format for navctl.

## Table of Contents

- [Config](#config)
- [ManagerConfig](#managerconfig)
- [EdgeConfig](#edgeconfig)
- [UIConfig](#uiconfig)
- [MetricsConfig](#metricsconfig)
- [MetricsAuth](#metricsauth)
- [ExecConfig](#execconfig)
- [EnvVar](#envvar)

## Config

Config represents the root configuration structure for navctl.

This configuration supports both YAML and JSON formats and allows for
comprehensive edge-specific configuration including metrics, authentication,
and Kubernetes contexts. The configuration file enables declarative
management of Navigator services across multiple clusters.

Example YAML configuration:

apiVersion: navigator.io/v1alpha1
kind: NavctlConfig
manager:
host: localhost
port: 8080
edges:
- name: production
clusterId: prod-cluster-1
context: prod-context
metrics:
type: prometheus
endpoint: https://prometheus.prod.example.com
auth:
bearerTokenExec:
command: kubectl
args: ["get", "secret", "token", "-o", "jsonpath={.data.token}"]
ui:
port: 8082
disabled: false

### Fields

#### `apiVersion`

APIVersion specifies the configuration schema version. Currently supported: "navigator.io/v1alpha1"

#### `kind`

Kind identifies this as a NavctlConfig. Must be: "NavctlConfig"

#### `manager`

Manager contains configuration for the Navigator manager service. The manager coordinates communication between multiple edge services and serves the frontend API.

See [ManagerConfig](#managerconfig) for configuration details.

#### `edges`

Edges contains configuration for each edge service. Each edge connects to a specific Kubernetes cluster and streams cluster state to the manager. Multiple edges enable multi-cluster service discovery and monitoring.

See [EdgeConfig](#edgeconfig) for configuration details.

#### `ui`

UI contains configuration for the web UI server. Optional - if omitted, default UI settings will be used.

See [UIConfig](#uiconfig) for configuration details.

## ManagerConfig

ManagerConfig holds configuration for the Navigator manager service.

The manager service acts as the central coordination point, maintaining
bidirectional streaming gRPC connections with edge services and serving
the frontend API through both gRPC and HTTP gateway endpoints.

Example configuration:

manager:
host: localhost
port: 8080
maxMessageSize: 10

### Fields

#### `host`

Host specifies the hostname or IP address for the manager service. Default: "localhost" The manager will bind to this address for incoming connections.

#### `port`

Port specifies the gRPC port for the manager service. Default: 8080 The HTTP gateway will automatically use port+1 (e.g., 8081).

#### `maxMessageSize`

MaxMessageSize specifies the maximum gRPC message size in megabytes. Default: 10 Increase this value if you have large service discovery payloads.

## EdgeConfig

EdgeConfig holds configuration for a single edge service.

An edge service connects to a specific Kubernetes cluster and streams
cluster state (services, pods, endpoints) to the manager. It also provides
on-demand proxy configuration analysis via the Envoy admin API.

Example configuration:

edges:
- name: production
clusterId: prod-cluster-1
context: prod-context
kubeconfig: /path/to/prod-kubeconfig
syncInterval: 30
logLevel: info
metrics:
type: prometheus
endpoint: https://prometheus.prod.example.com

### Fields

#### `name`

Name is a unique identifier for this edge configuration. Required. Used in logs and UI to distinguish between edges.

#### `clusterId`

ClusterID is a unique identifier for the Kubernetes cluster. Required. Used to identify the cluster in multi-cluster scenarios. Should be unique across all clusters in your Navigator deployment.

#### `context`

Context specifies the kubeconfig context to use for this edge. Optional. If omitted, uses the current context from kubeconfig. Must exist in the specified kubeconfig file.

#### `kubeconfig`

Kubeconfig specifies the path to the kubeconfig file. Optional. If omitted, uses the default kubeconfig location (~/.kube/config). Can be an absolute path or relative to the working directory.

#### `syncInterval`

SyncInterval specifies how often to sync cluster state, in seconds. Default: 30 Lower values provide more real-time updates but increase load.

#### `logLevel`

LogLevel specifies the logging level for this edge service. Default: "info" Valid values: "debug", "info", "warn", "error"

#### `logFormat`

LogFormat specifies the logging format for this edge service. Default: "text" Valid values: "text", "json"

#### `metrics`

Metrics contains configuration for metrics collection from this cluster. Optional. If omitted, metrics collection is disabled for this edge.

See [MetricsConfig](#metricsconfig) for configuration details.

## UIConfig

UIConfig holds configuration for the Navigator web UI server.

The UI server provides a web interface for service discovery, proxy
configuration viewing, and real-time cluster monitoring across all
connected edge services.

Example configuration:

ui:
port: 8082
disabled: false
noBrowser: false

### Fields

#### `port`

Port specifies the port for the web UI server. Default: 8082 The UI will be accessible at http://localhost:<port>

#### `disabled`

Disabled determines whether to start the UI server. Default: false Set to true to run navctl without the web interface.

#### `noBrowser`

NoBrowser determines whether to automatically open a browser. Default: false Set to true to prevent automatic browser launching when starting navctl.

## MetricsConfig

MetricsConfig holds configuration for metrics collection from a cluster.

Navigator supports pluggable metrics providers, with Prometheus being
the primary implementation. Metrics are collected during cluster state
synchronization and include service request rates, error rates, and
latency percentiles.

Example configuration:

metrics:
type: prometheus
endpoint: https://prometheus.prod.example.com
queryInterval: 30
timeout: 10
auth:
bearerTokenExec:
command: kubectl
args: ["get", "secret", "prometheus-token", "-o", "jsonpath={.data.token}"]

### Fields

#### `type`

Type specifies the metrics provider type. Currently supported: "Prometheus" Required when metrics collection is enabled.

#### `endpoint`

Endpoint specifies the URL for the metrics provider. Required. For Prometheus, this should be the base URL (e.g., https://Prometheus.example.com). The endpoint should be accessible from where navctl is running.

#### `queryInterval`

QueryInterval specifies how often to query for metrics, in seconds. Default: 30 Lower values provide more real-time metrics but increase load on the metrics provider.

#### `timeout`

Timeout specifies the timeout for metrics queries, in seconds. Default: 10 Increase this value if your metrics provider has high latency.

#### `auth`

Auth contains authentication configuration for the metrics provider. Optional. If omitted, no authentication is used. Supports static bearer tokens and dynamic token generation via exec commands.

See [MetricsAuth](#metricsauth) for configuration details.

## MetricsAuth

MetricsAuth holds authentication configuration for metrics providers.

Supports both static bearer tokens and dynamic token generation through
command execution, similar to Kubernetes client authentication.

Example configurations:

Static token:

auth:
bearerToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

Dynamic token via kubectl:

auth:
bearerTokenExec:
command: kubectl
args: ["get", "secret", "prometheus-token", "-o", "jsonpath={.data.token}"]
timeout: 30s

### Fields

#### `bearerToken`

BearerToken specifies a static bearer token for authentication. Optional. Mutually exclusive with BearerTokenExec. Use this for long-lived tokens or when you want to manage token rotation externally.

#### `bearerTokenExec`

BearerTokenExec specifies a command to execute to obtain a bearer token. Optional. Mutually exclusive with BearerToken. Tokens are cached for 15 minutes to avoid excessive command execution. Use this for dynamic token generation, similar to Kubernetes exec authentication.

See [ExecConfig](#execconfig) for configuration details.

## ExecConfig

ExecConfig holds configuration for executing commands to get bearer tokens.

This configuration enables dynamic token generation through command execution,
similar to the exec authentication mechanism used in Kubernetes kubeconfig files.
Commands are executed in a secure environment with configurable timeouts and
environment variables.

Example configuration:

bearerTokenExec:
command: kubectl
args: ["get", "secret", "prometheus-token", "-o", "jsonpath={.data.token}"]
timeout: 30s
env:
- name: KUBECONFIG
value: /path/to/kubeconfig

### Fields

#### `command`

Command specifies the executable to run for token generation. Required. Should be an absolute path or command available in PATH. Common examples: "`kubectl`", "gcloud", "aws", "/usr/local/bin/get-token"

#### `args`

Args specifies the command-line arguments to pass to the command. Optional. Use this to specify subcommands and parameters. Example: ["get", "secret", "token", "-o", "jsonpath={.data.token}"]

#### `env`

Env specifies additional environment variables for the command. Optional. Use this to set context-specific environment variables. These are added to the existing environment, not replacing it.

See [EnvVar](#envvar) for configuration details.

#### `timeout`

Timeout specifies the maximum duration to wait for command completion. Default: "30s" Valid formats: "30s", "5m", "1h" (any duration parseable by time.ParseDuration)

## EnvVar

EnvVar represents an environment variable for exec commands.

Environment variables are passed to the command execution environment
and can be used to provide context-specific configuration such as
authentication credentials, API endpoints, or other runtime parameters.

Example usage in exec configuration:

env:
- name: KUBECONFIG
value: /path/to/specific/kubeconfig
- name: AWS_PROFILE
value: production
- name: PROMETHEUS_URL
value: https://prometheus.example.com

### Fields

#### `name`

Name is the environment variable name. Required. Should follow standard environment variable naming conventions. Example: "KUBECONFIG", "AWS_PROFILE", "PROMETHEUS_URL"

#### `value`

Value is the environment variable value. Required. The value to set for the named environment variable. Can contain absolute paths, URLs, or any string value.

