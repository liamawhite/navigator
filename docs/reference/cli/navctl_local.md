## navctl local

Run manager and edge services locally

### Synopsis

Run both Navigator manager and edge services locally.
This command starts the manager service first, then connects edge services to it.
By default, it uses the current context from your kubeconfig. You can specify
multiple contexts using the --contexts flag to monitor multiple clusters simultaneously.

The --contexts flag supports both exact context names and glob patterns:
  * - matches any sequence of characters
  ? - matches any single character
  [abc] - matches any character in brackets
  [a-z] - matches any character in range

Examples:
  # Use current context
  navctl local

  # Use specific contexts
  navctl local --contexts context1,context2,context3

  # Use glob patterns to select multiple contexts
  navctl local --contexts "*-prod"
  navctl local --contexts "team-*"
  navctl local --contexts "*-prod,*-staging"
  navctl local --contexts "dev-*,test-?"

  # Mix exact names and patterns
  navctl local --contexts "production,*-staging"

  # Use custom kubeconfig with patterns
  navctl local --kube-config ~/.kube/config --contexts "*-prod"

Available contexts will be shown from your kubeconfig file.

```
navctl local [flags]
```

### Options

```
      --contexts strings          Comma-separated list of kubeconfig contexts to use (uses current context if not specified)
      --disable-ui                Disable UI server
  -h, --help                      help for local
  -k, --kube-config string        Path to kubeconfig file (default "~/.kube/config")
      --manager-host string       Host for manager service (default "localhost")
      --manager-port int          Port for manager service (default 8080)
      --max-message-size int      Maximum gRPC message size in MB (default 10)
      --metrics-endpoint string   Metrics provider endpoint accessible from this machine (e.g., http://prometheus:9090). Enables metrics if provided.
      --metrics-timeout int       Metrics query timeout in seconds (default 10)
      --metrics-type string       Metrics provider type (prometheus) (default "prometheus")
      --no-browser                Don't open browser automatically
      --ui-port int               Port for UI server (default 8082)
```

### Options inherited from parent commands

```
      --log-format string   Log format (text, json) (default "text")
      --log-level string    Log level (debug, info, warn, error) (default "info")
```

## Examples

### Basic Usage

Start Navigator with current kubeconfig context:
```bash
navctl local
```

Start with specific contexts:
```bash
navctl local --contexts context1,context2,context3
```

### Metrics Integration Examples

Enable metrics with Prometheus endpoint:
```bash
# Basic Prometheus integration
navctl local --metrics-endpoint http://localhost:9090

# With custom timeout
navctl local --metrics-endpoint http://prometheus:9090 --metrics-timeout 15
```

#### Istio with Prometheus Addon

Port-forward Prometheus from Istio and start Navigator:
```bash
# Terminal 1: Port-forward Prometheus
kubectl port-forward -n istio-system service/prometheus 9090:9090

# Terminal 2: Start Navigator with metrics
navctl local --metrics-endpoint http://localhost:9090
```

#### External Prometheus Setup

Connect to external Prometheus instance:
```bash
navctl local \
  --metrics-endpoint http://prometheus.monitoring.svc.cluster.local:9090 \
  --metrics-timeout 30
```

### Multi-cluster with Metrics

Monitor multiple production clusters with metrics:
```bash
navctl local \
  --contexts "*-prod" \
  --metrics-endpoint http://prometheus.global.monitoring:9090 \
  --metrics-timeout 20
```

### Development Workflow

Development setup with debug logging and metrics:
```bash
navctl local \
  --metrics-endpoint http://localhost:9090 \
  --log-level debug \
  --log-format json \
  --no-browser
```

### Custom Port Configuration

Use custom ports to avoid conflicts:
```bash
navctl local \
  --manager-port 9090 \
  --ui-port 3001 \
  --metrics-endpoint http://localhost:15090
```

### Production Deployment

Production setup with multiple clusters and metrics:
```bash
navctl local \
  --contexts "prod-us-east,prod-us-west,prod-eu-central" \
  --metrics-endpoint http://prometheus.prod.monitoring:9090 \
  --metrics-timeout 30 \
  --max-message-size 20 \
  --manager-port 8080 \
  --ui-port 8082 \
  --no-browser
```

### Troubleshooting Examples

Debug metrics connectivity issues:
```bash
# Test metrics provider connectivity first
curl -f http://localhost:9090/api/v1/query?query=up

# Start Navigator with debug logging
navctl local \
  --metrics-endpoint http://localhost:9090 \
  --log-level debug \
  --log-format json
```

Verify Istio metrics availability:
```bash
# Check for Istio request metrics
curl 'http://localhost:9090/api/v1/query?query=istio_requests_total' | jq '.data.result | length'

# Start Navigator if metrics are available
navctl local --metrics-endpoint http://localhost:9090
```

### SEE ALSO

* [navctl](navctl.md)	 - Navigator control plane CLI
* [Metrics Guide](../../user-guide/metrics.md) - Detailed metrics setup and configuration

