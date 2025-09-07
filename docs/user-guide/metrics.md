# Metrics Guide

Navigator provides powerful metrics integration to visualize and analyze service-to-service communication patterns in your Kubernetes clusters. This guide covers everything you need to know about setting up and using metrics in Navigator.

## Overview

Navigator's metrics system enables you to:

- **Monitor service communication**: Track request flows between services
- **Analyze performance**: View request rates, error rates, and latency metrics
- **Visualize topology**: See your service mesh communication patterns
- **Multi-cluster insights**: Aggregate metrics across multiple Kubernetes clusters
- **Real-time updates**: Monitor live traffic patterns and performance

## Prerequisites

### Metrics Provider

Navigator requires a compatible metrics provider to collect and query metrics data:

#### Prometheus (Recommended)

The most common setup, especially with Istio service mesh:

1. **Install Prometheus** (if not already installed):
   ```bash
   # For Istio users - install Prometheus addon
   kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.20/samples/addons/prometheus.yaml
   
   # Or install standalone Prometheus via Helm
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm install prometheus prometheus-community/prometheus
   ```

2. **Verify Prometheus is collecting metrics**:
   ```bash
   kubectl port-forward -n istio-system service/prometheus 9090:9090
   # Visit http://localhost:9090 to access Prometheus UI
   ```

#### Other Metrics Providers

Navigator is designed with a generic provider interface to support additional metrics backends in the future. Currently supported:
- **Prometheus**: Full support with service graph metrics
- **Future providers**: Support for Grafana, DataDog, and other observability platforms planned

### Service Mesh (Optional but Recommended)

While not strictly required, a service mesh like Istio provides the richest metrics data:

- **Request-level metrics**: Detailed service-to-service communication data
- **Golden signals**: Request rate, error rate, and latency percentiles
- **Service topology**: Automatic discovery of service dependencies
- **Multi-cluster support**: Metrics across federated service meshes

## Configuration

### Basic Setup

Enable metrics by providing a metrics endpoint when starting Navigator:

```bash
# Basic Prometheus setup
navctl local --metrics-endpoint http://localhost:9090
```

### Advanced Configuration

```bash
# Full configuration with custom settings
navctl local \
  --metrics-endpoint http://prometheus.monitoring:9090 \
  --metrics-type prometheus \
  --metrics-timeout 15 \
  --contexts "*-prod,*-staging"
```

### Configuration Options

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--metrics-endpoint` | Metrics provider endpoint URL | None (disabled) | `http://localhost:9090` |
| `--metrics-type` | Metrics provider type | `prometheus` | `prometheus` |
| `--metrics-timeout` | Query timeout in seconds | `10` | `15` |

### Environment-Specific Examples

#### Development with Istio

```bash
# Port-forward Prometheus from Istio
kubectl port-forward -n istio-system service/prometheus 9090:9090

# Start Navigator with metrics
navctl local --metrics-endpoint http://localhost:9090
```

#### Production Setup

```bash
# Connect to production Prometheus instance
navctl local \
  --metrics-endpoint http://prometheus.monitoring.svc.cluster.local:9090 \
  --metrics-timeout 30 \
  --contexts "prod-*"
```

#### Multi-Cluster Environment

```bash
# Monitor multiple clusters with shared Prometheus
navctl local \
  --contexts "prod-us-east,prod-us-west,prod-eu" \
  --metrics-endpoint http://prometheus.global.monitoring:9090
```

## Using the Topology View

### Accessing the View

When metrics are enabled, the **Topology** navigation item appears in Navigator's UI:

1. Start Navigator with metrics enabled
2. Open the web interface (usually http://localhost:8082)
3. Click **Topology** in the navigation bar
4. View real-time service communication graphs

### Understanding the Interface

#### Service Graph Table

The topology view displays service-to-service communication in a tabular format:

- **Source**: The service initiating requests
- **Destination**: The service receiving requests
- **Request Rate**: Requests per second between the services
- **Error Rate**: Failed requests per second and error percentage
- **Service Information**: Namespace and cluster details for each service

#### Summary Cards

At the top of the view, summary cards show:

- **Service Pairs**: Total number of communicating service pairs
- **Total Request Rate**: Aggregated requests per second across all services
- **Total Error Rate**: Aggregated error rate across all communications

#### Refresh Controls

- **Manual Refresh**: Click the refresh button to update data on-demand
- **Automatic Refresh**: Configure intervals from 5 seconds to 5 minutes
- **Time Range**: Data covers the last 5 minutes of service communication

### Navigation Behavior

- **Automatic Display**: Topology appears when any connected cluster has metrics capabilities
- **Mixed Capabilities**: Warning icon shows when some clusters have metrics and others don't
- **No Metrics**: Helpful guidance screen appears when no clusters have metrics enabled

## Metrics Data

### Service Graph Metrics

Navigator collects and displays the following metrics for service-to-service communication:

#### Request Rate
- **Metric**: Requests per second between service pairs
- **Source**: Prometheus histogram buckets and counters
- **Aggregation**: Sum across all instances and time windows
- **Display**: Formatted as "X.XX req/s"

#### Error Rate  
- **Metric**: Failed requests per second and error percentage
- **Source**: HTTP response codes 4xx and 5xx from Prometheus
- **Calculation**: Error count / total count over time window
- **Display**: Formatted as "X.XX err/s" with color-coded badges

#### Latency Percentiles (Future)
- **Metrics**: P50, P95, P99 response times
- **Source**: Prometheus histogram quantiles
- **Display**: Millisecond response times

### Data Freshness

- **Collection Interval**: Metrics are queried every 30 seconds during cluster sync
- **Time Window**: Displays data from the last 5 minutes
- **Real-time Updates**: UI refreshes based on configured interval
- **Cross-Cluster**: Metrics aggregated across all connected clusters with capabilities

## Cluster Capabilities

### Edge Reporting

Each edge service reports its metrics capabilities to the manager:

- **Capability Detection**: Edges automatically detect configured metrics providers
- **Status Reporting**: Capabilities included in cluster identification messages
- **Health Monitoring**: Connection status and provider availability tracked

### Mixed Environments

Navigator handles mixed environments where some clusters have metrics and others don't:

- **Selective Display**: Only clusters with metrics contribute to topology view
- **Visual Indicators**: Warning icons show when metrics capabilities are mixed
- **Graceful Degradation**: Non-metrics clusters still provide service registry data

## Troubleshooting

### Common Issues

#### Topology View Not Showing

**Symptoms**: No "Topology" button in navigation
**Cause**: No connected clusters have metrics capabilities enabled
**Solution**: 
```bash
# Verify metrics endpoint is accessible
curl http://localhost:9090/api/v1/query?query=up

# Restart Navigator with metrics enabled
navctl local --metrics-endpoint http://localhost:9090
```

#### No Service Communication Data

**Symptoms**: Topology view shows "No service communication detected"
**Cause**: Service mesh not generating metrics or queries not finding data
**Solution**:
1. Verify Istio is installed and sidecars are injected
2. Generate some traffic between services
3. Check Prometheus has Istio metrics:
   ```bash
   # Query for Istio request metrics
   curl 'http://localhost:9090/api/v1/query?query=istio_requests_total'
   ```

#### Metrics Provider Connection Errors

**Symptoms**: Error messages about failed metrics queries
**Cause**: Prometheus endpoint unreachable or authentication issues
**Solutions**:
- Verify endpoint URL is correct and accessible
- Check network connectivity between Navigator and Prometheus
- Increase timeout with `--metrics-timeout` flag
- Verify Prometheus is running and healthy

#### Mixed Metrics Warning

**Symptoms**: Warning icon in cluster status showing mixed metrics capabilities
**Cause**: Some clusters have metrics enabled, others don't
**Solution**: This is informational - either:
- Enable metrics for all clusters: `--metrics-endpoint` affects all edges
- Accept mixed state if different clusters have different monitoring needs

### Debugging Commands

```bash
# Test Prometheus connectivity
curl -f http://localhost:9090/api/v1/query?query=up

# Check Istio metrics availability
curl 'http://localhost:9090/api/v1/query?query=istio_requests_total' | jq '.data.result | length'

# Verify Navigator can reach Prometheus
navctl local --metrics-endpoint http://localhost:9090 --log-level debug
```

### Log Analysis

Enable debug logging to troubleshoot metrics issues:

```bash
navctl local \
  --metrics-endpoint http://localhost:9090 \
  --log-level debug \
  --log-format json
```

Look for log entries related to:
- Metrics provider initialization
- Query execution and results
- Connection errors or timeouts
- Capability reporting from edges

## Integration Examples

### With Istio Service Mesh

```bash
# 1. Install Istio with Prometheus
istioctl install --set values.telemetry.v2.prometheus.configOverride.inboundSidecar.disable_host_header_fallback=true

# 2. Install Prometheus addon
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.20/samples/addons/prometheus.yaml

# 3. Deploy sample applications
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.20/samples/bookinfo/platform/kube/bookinfo.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.20/samples/bookinfo/networking/bookinfo-gateway.yaml

# 4. Port-forward Prometheus
kubectl port-forward -n istio-system service/prometheus 9090:9090 &

# 5. Start Navigator with metrics
navctl local --metrics-endpoint http://localhost:9090

# 6. Generate traffic to see metrics
curl http://localhost:80/productpage
```

### With External Prometheus

```bash
# Connect to external Prometheus instance
navctl local \
  --metrics-endpoint http://prometheus.external.com:9090 \
  --metrics-timeout 30
```

### CI/CD Integration

```bash
#!/bin/bash
# Deploy and monitor with Navigator

# Deploy application
kubectl apply -f app.yaml

# Wait for rollout
kubectl rollout status deployment/my-app

# Start Navigator with metrics for monitoring
navctl local \
  --metrics-endpoint http://prometheus:9090 \
  --no-browser \
  --contexts "staging" &

NAVCTL_PID=$!

# Run tests while monitoring
run-integration-tests.sh

# Cleanup
kill $NAVCTL_PID
```

## Next Steps

- Learn about [UI Navigation](ui-navigation.md) for detailed interface guidance
- Explore the [Developer Guide](../developer-guide/) for metrics architecture details
- Review the [CLI Reference](../reference/cli/navctl_local.md) for all available options
- Check out [Architecture Documentation](../developer-guide/architecture.md) for system design details