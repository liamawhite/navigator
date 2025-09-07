# Metrics System

This document provides detailed technical documentation for Navigator's metrics subsystem, covering the architecture, implementation, and integration patterns for service-to-service communication monitoring.

## System Overview

Navigator's metrics system is designed with a modular, provider-agnostic architecture that supports multiple metrics backends while providing a unified interface for service graph visualization. The system spans all three core Navigator components (edge, manager, UI) with well-defined interfaces and data flow patterns.

### Design Principles

- **Provider Agnostic**: Generic interface supports multiple metrics backends
- **Capability Reporting**: Edges dynamically report their metrics support to the manager
- **Performance Focused**: Metrics collection integrated into existing sync cycles
- **Fault Tolerant**: Graceful degradation when metrics providers are unavailable
- **Multi-Cluster**: Native support for aggregating metrics across clusters

## Provider Interface Design

### MetricsProvider Interface

The core abstraction that enables pluggable metrics backends:

```go
// edge/pkg/interfaces/metrics_provider.go
type MetricsProvider interface {
    // GetServiceGraphMetrics retrieves service-to-service communication metrics
    GetServiceGraphMetrics(ctx context.Context, req ServiceGraphMetricsRequest) (*ServiceGraphMetricsResponse, error)
    
    // HealthCheck verifies the metrics provider is accessible and functional
    HealthCheck(ctx context.Context) error
    
    // Close gracefully shuts down connections and releases resources
    Close() error
}

type ServiceGraphMetricsRequest struct {
    StartTime time.Time
    EndTime   time.Time
    Clusters  []string // Optional cluster filtering
}

type ServiceGraphMetricsResponse struct {
    ServicePairs []ServiceGraphPair
    Timestamp    time.Time
}

type ServiceGraphPair struct {
    SourceService      string
    SourceNamespace    string
    SourceCluster      string
    DestinationService string
    DestinationNamespace string
    DestinationCluster string
    RequestRate        float64 // requests per second
    ErrorRate          float64 // errors per second
    // Future: Latency percentiles (P50, P95, P99)
}
```

### Provider Registration

Providers register themselves during edge service initialization:

```go
// edge/pkg/metrics/registry.go
type ProviderRegistry struct {
    providers map[string]ProviderFactory
    mu        sync.RWMutex
}

type ProviderFactory func(config ProviderConfig) (MetricsProvider, error)

func (r *ProviderRegistry) Register(name string, factory ProviderFactory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers[name] = factory
}

func (r *ProviderRegistry) Create(providerType string, config ProviderConfig) (MetricsProvider, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    factory, exists := r.providers[providerType]
    if !exists {
        return nil, fmt.Errorf("unknown provider type: %s", providerType)
    }
    
    return factory(config)
}
```

## Prometheus Provider Implementation

### Architecture

The Prometheus provider implements the `MetricsProvider` interface with comprehensive support for Istio service mesh metrics:

```go
// edge/pkg/metrics/prometheus/provider.go
type Provider struct {
    client   api.Client
    config   Config
    logger   *slog.Logger
    mutex    sync.RWMutex
    lastSeen map[string]time.Time // Connection health tracking
}

type Config struct {
    Endpoint     string        // Prometheus API endpoint
    Timeout      time.Duration // Query timeout
    MaxRetries   int          // Retry attempts for failed queries
    RetryDelay   time.Duration // Delay between retries
}
```

### Service Graph Query Strategy

The provider uses sophisticated PromQL queries to extract service communication patterns:

#### Request Rate Queries
```promql
# Total requests between service pairs
sum(rate(istio_requests_total{reporter="source"}[5m])) by (
    source_service_name, source_service_namespace, source_cluster,
    destination_service_name, destination_service_namespace, destination_cluster
)
```

#### Error Rate Queries
```promql
# Error requests (4xx, 5xx responses)
sum(rate(istio_requests_total{reporter="source",response_code=~"[45].."}[5m])) by (
    source_service_name, source_service_namespace, source_cluster,
    destination_service_name, destination_service_namespace, destination_cluster
)
```

### Query Optimization

- **Batch Queries**: Multiple metrics retrieved in single API call
- **Time Range Optimization**: Configurable query windows with intelligent defaults
- **Label Filtering**: Cluster-specific filtering to reduce query scope
- **Connection Pooling**: HTTP client reuse for performance

### Error Handling and Resilience

```go
func (p *Provider) GetServiceGraphMetrics(ctx context.Context, req ServiceGraphMetricsRequest) (*ServiceGraphMetricsResponse, error) {
    var lastErr error
    
    for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(p.config.RetryDelay):
            }
        }
        
        result, err := p.executeQuery(ctx, req)
        if err == nil {
            p.updateLastSeen()
            return result, nil
        }
        
        lastErr = err
        p.logger.Warn("Metrics query failed", "attempt", attempt+1, "error", err)
    }
    
    return nil, fmt.Errorf("metrics query failed after %d attempts: %w", p.config.MaxRetries+1, lastErr)
}
```

## Edge Integration

### Capability Detection and Reporting

Edges automatically detect metrics capabilities during initialization:

```go
// edge/pkg/service/service.go
func (e *EdgeService) initializeMetrics() error {
    if e.config.MetricsEndpoint == "" {
        e.logger.Info("Metrics endpoint not configured, metrics disabled")
        return nil
    }
    
    provider, err := e.metricsRegistry.Create(e.config.MetricsType, prometheus.Config{
        Endpoint: e.config.MetricsEndpoint,
        Timeout:  e.config.MetricsTimeout,
    })
    if err != nil {
        return fmt.Errorf("failed to create metrics provider: %w", err)
    }
    
    // Test connectivity
    ctx, cancel := context.WithTimeout(context.Background(), e.config.MetricsTimeout)
    defer cancel()
    
    if err := provider.HealthCheck(ctx); err != nil {
        e.logger.Warn("Metrics provider health check failed", "error", err)
        return err
    }
    
    e.metricsProvider = provider
    e.logger.Info("Metrics provider initialized successfully", "type", e.config.MetricsType)
    return nil
}
```

### Capability Reporting Protocol

During cluster identification, edges report their metrics capabilities:

```go
func (e *EdgeService) sendClusterIdentification() error {
    capabilities := &v1alpha1.EdgeCapabilities{
        MetricsEnabled: e.metricsProvider != nil,
    }
    
    identification := &v1alpha1.ClusterIdentification{
        ClusterId:    e.config.ClusterID,
        Capabilities: capabilities,
    }
    
    return e.stream.Send(&v1alpha1.ConnectRequest{
        Payload: &v1alpha1.ConnectRequest_ClusterIdentification{
            ClusterIdentification: identification,
        },
    })
}
```

### Metrics Collection Integration

Metrics collection is integrated into the periodic cluster sync cycle:

```go
func (e *EdgeService) collectMetrics(ctx context.Context) (*v1alpha1.ServiceGraphMetricsResponse, error) {
    if e.metricsProvider == nil {
        return nil, nil // Metrics not enabled
    }
    
    endTime := time.Now()
    startTime := endTime.Add(-5 * time.Minute) // 5-minute window
    
    req := interfaces.ServiceGraphMetricsRequest{
        StartTime: startTime,
        EndTime:   endTime,
        Clusters:  []string{e.config.ClusterID},
    }
    
    response, err := e.metricsProvider.GetServiceGraphMetrics(ctx, req)
    if err != nil {
        e.logger.Error("Failed to collect metrics", "error", err)
        return nil, err
    }
    
    return convertToProtoResponse(response), nil
}
```

## Manager Aggregation

### Connection Management

The manager tracks metrics capabilities for each connected edge:

```go
// manager/pkg/connections/types.go
type Connection struct {
    ID           string
    ClusterID    string
    Stream       v1alpha1.ManagerService_ConnectServer
    LastSeen     time.Time
    Capabilities *v1alpha1.EdgeCapabilities // Includes metrics_enabled
    MetricsData  *ServiceGraphMetrics      // Cached metrics data
    mutex        sync.RWMutex
}

type ConnectionInfo struct {
    ClusterID      string
    LastUpdate     time.Time
    ServiceCount   int32
    SyncStatus     v1alpha1.SyncStatus
    MetricsEnabled bool // Exposed to frontend API
}
```

### Metrics Aggregation Logic

```go
// manager/pkg/frontend/metrics_service.go
func (s *MetricsService) GetServiceGraphMetrics(
    ctx context.Context,
    req *v1alpha1.GetServiceGraphMetricsRequest,
) (*v1alpha1.GetServiceGraphMetricsResponse, error) {
    
    connections := s.connectionManager.GetConnections()
    var allServicePairs []ServicePair
    
    for _, conn := range connections {
        if !conn.Capabilities.MetricsEnabled {
            continue // Skip connections without metrics
        }
        
        // Request fresh metrics from this edge
        metricsReq := &v1alpha1.ServiceGraphMetricsRequest{
            StartTime: req.StartTime,
            EndTime:   req.EndTime,
        }
        
        metricsResp, err := s.requestMetricsFromEdge(ctx, conn, metricsReq)
        if err != nil {
            s.logger.Warn("Failed to get metrics from edge", 
                "cluster", conn.ClusterID, "error", err)
            continue
        }
        
        allServicePairs = append(allServicePairs, metricsResp.ServicePairs...)
    }
    
    return &v1alpha1.GetServiceGraphMetricsResponse{
        ServicePairs: allServicePairs,
        Timestamp:    timestamppb.Now(),
    }, nil
}
```

### Cross-Cluster Aggregation

Service pairs from multiple clusters are consolidated with cluster context preserved:

```go
func aggregateServicePairs(pairs []ServicePair) []ServicePair {
    pairMap := make(map[string]*ServicePair)
    
    for _, pair := range pairs {
        key := fmt.Sprintf("%s:%s->%s:%s", 
            pair.SourceService, pair.SourceCluster,
            pair.DestinationService, pair.DestinationCluster)
        
        if existing, exists := pairMap[key]; exists {
            // Aggregate metrics for same service pair
            existing.RequestRate += pair.RequestRate
            existing.ErrorRate += pair.ErrorRate
        } else {
            pairCopy := pair
            pairMap[key] = &pairCopy
        }
    }
    
    result := make([]ServicePair, 0, len(pairMap))
    for _, pair := range pairMap {
        result = append(result, *pair)
    }
    
    return result
}
```

## Frontend API

### HTTP Gateway Integration

The metrics API is exposed through the same HTTP gateway as other Navigator APIs:

```go
// Generated gRPC-Gateway registration
func RegisterMetricsServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
    // Auto-generated gateway registration for MetricsService
}
```

### RESTful Endpoints

The metrics service provides RESTful endpoints via gRPC-Gateway:

- `GET /api/v1alpha1/metrics/service-graph` - Retrieve service graph metrics
- Query parameters: `start_time`, `end_time` (ISO 8601 format)

Example request:
```bash
curl "http://localhost:8081/api/v1alpha1/metrics/service-graph?start_time=2024-01-15T10:00:00Z&end_time=2024-01-15T10:05:00Z"
```

### Response Format

```json
{
  "servicePairs": [
    {
      "sourceService": "productpage",
      "sourceNamespace": "default",
      "sourceCluster": "prod-us-east",
      "destinationService": "reviews",
      "destinationNamespace": "default", 
      "destinationCluster": "prod-us-east",
      "requestRate": 12.5,
      "errorRate": 0.2
    }
  ],
  "timestamp": "2024-01-15T10:05:00Z"
}
```

## UI Implementation

### React Hook for Metrics

Custom hook manages metrics data fetching and state:

```typescript
// ui/src/hooks/useServiceGraphMetrics.ts
export function useServiceGraphMetrics(
    request: ServiceGraphMetricsRequest,
    refetchInterval: number | false = false
) {
    return useQuery({
        queryKey: ['service-graph-metrics', request],
        queryFn: async () => {
            const api = new MetricsServiceService({
                baseUrl: '/api/v1alpha1'
            });
            return api.getServiceGraphMetrics(request);
        },
        refetchInterval,
        staleTime: 30000, // Consider data stale after 30 seconds
        retry: (failureCount, error) => {
            // Retry up to 2 times for network errors
            return failureCount < 2 && isNetworkError(error);
        }
    });
}
```

### Capability Detection

UI components check for metrics capabilities before rendering topology features:

```typescript
// ui/src/components/Navbar.tsx
const [hasMetricsCapability, setHasMetricsCapability] = useState<boolean>(false);

const checkMetricsCapability = useCallback(async () => {
    try {
        const clusterData = await serviceApi.listClusters();
        const hasAnyMetrics = clusterData.some(
            (cluster: v1alpha1ClusterSyncInfo) => cluster.metricsEnabled
        );
        setHasMetricsCapability(hasAnyMetrics);
    } catch (error) {
        console.error('Failed to check metrics capability:', error);
        setHasMetricsCapability(false);
    }
}, []);

// Conditional rendering based on capabilities
{hasMetricsCapability && (
    <NavigationItem href="/topology" icon={Waypoints}>
        Topology
    </NavigationItem>
)}
```

### State Management Patterns

The UI uses React patterns for managing metrics state:

- **Loading States**: Progressive loading with skeleton UI
- **Error Boundaries**: Graceful error handling for failed metrics queries
- **Real-time Updates**: Configurable refresh intervals with manual refresh options
- **Capability Caching**: Cluster capability state cached and periodically refreshed

### Mixed Capability Handling

Special handling for mixed environments where some clusters have metrics:

```typescript
// ui/src/components/ClusterSyncStatus.tsx
const hasMixedMetricsCapability = (clusters: v1alpha1ClusterSyncInfo[]): boolean => {
    if (clusters.length === 0) return false;
    
    const metricsEnabled = clusters.map(c => c.metricsEnabled);
    const hasMetrics = metricsEnabled.some(enabled => enabled);
    const hasNoMetrics = metricsEnabled.some(enabled => !enabled);
    
    return hasMetrics && hasNoMetrics;
};

// Visual warning for mixed capabilities
{hasMixedMetricsCapability(clusters) && (
    <Tooltip content="Some clusters have metrics enabled while others don't">
        <AlertTriangle className="h-4 w-4 text-orange-600" />
    </Tooltip>
)}
```

## Testing Patterns

### Unit Tests

#### Provider Interface Testing

```go
func TestPrometheusProvider_GetServiceGraphMetrics(t *testing.T) {
    // Mock Prometheus server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        response := mockPrometheusResponse()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    provider, err := prometheus.New(prometheus.Config{
        Endpoint: server.URL,
        Timeout:  5 * time.Second,
    })
    require.NoError(t, err)
    
    req := interfaces.ServiceGraphMetricsRequest{
        StartTime: time.Now().Add(-5 * time.Minute),
        EndTime:   time.Now(),
    }
    
    response, err := provider.GetServiceGraphMetrics(context.Background(), req)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.ServicePairs)
}
```

#### Edge Service Testing

```go
func TestEdgeService_ReportsMetricsCapability(t *testing.T) {
    // Mock manager connection
    mockManager := &MockManagerConnection{}
    
    // Edge with metrics enabled
    edge := &EdgeService{
        metricsProvider: &MockMetricsProvider{},
        config: EdgeConfig{
            ClusterID: "test-cluster",
        },
    }
    
    err := edge.sendClusterIdentification()
    require.NoError(t, err)
    
    // Verify capability was reported
    assert.True(t, mockManager.ReceivedCapabilities.MetricsEnabled)
}
```

### Integration Tests

#### End-to-End Metrics Flow

```go
func TestMetricsE2E(t *testing.T) {
    // Start mock Prometheus server
    prometheusServer := startMockPrometheus(t)
    defer prometheusServer.Close()
    
    // Start manager
    manager := startTestManager(t)
    defer manager.Stop()
    
    // Start edge with metrics
    edge := startTestEdge(t, EdgeConfig{
        ManagerEndpoint: manager.Address(),
        MetricsEndpoint: prometheusServer.URL,
        MetricsType:     "prometheus",
    })
    defer edge.Stop()
    
    // Wait for connection establishment
    require.Eventually(t, func() bool {
        connections := manager.GetConnections()
        return len(connections) == 1 && connections[0].Capabilities.MetricsEnabled
    }, 10*time.Second, 100*time.Millisecond)
    
    // Query metrics via frontend API
    client := manager.GetHTTPClient()
    resp, err := client.Get("/api/v1alpha1/metrics/service-graph")
    require.NoError(t, err)
    
    var metricsResponse v1alpha1.GetServiceGraphMetricsResponse
    err = json.NewDecoder(resp.Body).Decode(&metricsResponse)
    require.NoError(t, err)
    
    assert.NotEmpty(t, metricsResponse.ServicePairs)
}
```

## Performance Considerations

### Query Optimization

- **Batch Processing**: Multiple metrics retrieved in single operations
- **Connection Pooling**: HTTP client reuse across requests
- **Query Caching**: Results cached with TTL to reduce provider load
- **Time Range Limits**: Configurable windows prevent expensive long-range queries

### Memory Management

- **Bounded Caches**: LRU eviction for metrics data caches
- **Connection Limits**: Maximum concurrent connections to metrics providers
- **Garbage Collection**: Regular cleanup of stale metrics data

### Scalability Patterns

- **Horizontal Scaling**: Manager can aggregate metrics from hundreds of edges
- **Provider Federation**: Support for federated Prometheus setups
- **Query Distribution**: Metrics queries distributed across multiple provider instances

## Security Considerations

### Authentication and Authorization

- **Provider Authentication**: Support for basic auth, token auth for metrics providers
- **TLS Security**: Encrypted connections to metrics providers when configured
- **Network Isolation**: Metrics providers accessed through configured network policies

### Data Privacy

- **No PII Storage**: Metrics contain only service names and performance data
- **Configurable Retention**: Metrics data TTL configurable per deployment
- **Audit Logging**: All metrics queries logged for security auditing

## Future Enhancements

### Additional Providers

- **Grafana Integration**: Native Grafana data source support
- **DataDog Provider**: Integration with DataDog APM metrics
- **Custom Metrics**: User-defined metrics queries and aggregations

### Advanced Visualizations

- **Graph Topology**: Visual node-edge graphs for service relationships
- **Latency Heatmaps**: Time-series visualization of service performance
- **Anomaly Detection**: ML-based detection of unusual service patterns

### Performance Improvements

- **Stream Processing**: Real-time metrics streaming instead of polling
- **Metrics Push**: Edge-initiated metrics push to reduce query overhead
- **Distributed Caching**: Redis-backed caching for multi-instance deployments