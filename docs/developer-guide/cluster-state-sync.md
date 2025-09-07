# Cluster State Synchronization

This document describes how Navigator performs cluster state synchronization between edge processes and the central manager.

## Syncing Cluster State

### Overview

Edge processes continuously monitor their local Kubernetes clusters and synchronize high-level state information with the central manager. This synchronization provides the manager with a consolidated view of all clusters while keeping detailed data distributed at the edges.

### Periodic Collection Process

Edge processes operate on configurable intervals (typically every 30 seconds) to collect cluster state:

1. **Service Discovery**: Query the Kubernetes API server for all Services across all namespaces
2. **Endpoint Collection**: Query for all EndpointSlices to understand service endpoints
3. **Pod Enumeration**: Query for all Pods to track workload state
4. **Istio Resource Discovery**: Query for Istio Custom Resource Definitions (CRDs) including VirtualServices, DestinationRules, Gateways, ServiceEntries, Sidecars, EnvoyFilters, authentication policies, and WebAssembly plugins across all namespaces
5. **Metrics Collection**: Query configured metrics providers for service-to-service communication data (when metrics capabilities are enabled)

### Data Packaging

The collected Kubernetes and Istio resources are packaged into a ClusterState message that includes:

#### Core Kubernetes Resources
- **Services**: Service definitions including names, namespaces, selectors, and ports
- **EndpointSlices**: Endpoint information including IP addresses, ports, and readiness status
- **Pods**: Pod metadata including names, namespaces, labels, and current phase

#### Istio Service Mesh Resources
- **VirtualService**: Traffic routing and splitting rules for service-to-service communication
- **DestinationRule**: Load balancing policies, connection pooling, and circuit breaker configurations
- **Gateway**: Ingress and egress gateway configurations for external traffic management
- **ServiceEntry**: External service registrations to extend the service mesh
- **Sidecar**: Sidecar proxy configurations controlling traffic flow and resource consumption
- **EnvoyFilter**: Custom Envoy proxy filters for advanced traffic manipulation
- **PeerAuthentication**: Mutual TLS (mTLS) authentication policies between services
- **RequestAuthentication**: JWT token authentication policies for incoming requests
- **WasmPlugin**: WebAssembly plugin configurations for extending proxy functionality
- **IstioControlPlaneConfig**: Istio control plane metadata and configuration settings

#### Metrics Data (Optional)
When metrics capabilities are enabled on the edge, additional metrics data is collected and included:
- **Service Graph Metrics**: Service-to-service communication patterns with request rates, error rates, and latency data
- **Performance Metrics**: Golden signal metrics (request rate, error rate, duration) aggregated over configurable time windows
- **Multi-cluster Context**: Metrics tagged with source and destination cluster information for federated environments

### Transmission to Manager

The ClusterState message is sent to the manager over the established streaming connection:

- **Streaming Protocol**: Uses the bidirectional gRPC stream for real-time delivery
- **Message Identification**: Each message includes edge identification and timestamp
- **Acknowledgment**: Manager acknowledges receipt to ensure reliable delivery

### Metrics Collection Details

When an edge service has metrics capabilities enabled, it performs additional data collection during each sync cycle:

#### Metrics Provider Interaction
1. **Provider Health Check**: Verify metrics provider connectivity and availability
2. **Time Window Calculation**: Define query time range (typically last 5 minutes)  
3. **Service Graph Query**: Execute provider-specific queries for service-to-service metrics
4. **Data Transformation**: Convert provider-specific response format to Navigator's standard metrics model
5. **Error Handling**: Graceful degradation when metrics queries fail or timeout

#### Query Strategy
- **Batched Queries**: Multiple metrics retrieved in single operations for efficiency
- **Cluster Filtering**: Queries scoped to current cluster to avoid cross-cluster data leakage
- **Time Range Optimization**: Configurable query windows balance data freshness with performance
- **Retry Logic**: Failed queries automatically retried with exponential backoff

#### Data Integration
Metrics data is integrated with the regular cluster state synchronization:
- **Atomic Updates**: Metrics included in the same message as Kubernetes resource state
- **Consistency**: Metrics timestamp aligned with cluster state collection timestamp  
- **Failure Isolation**: Metrics collection failures don't prevent Kubernetes state sync
- **Capability Reporting**: Edge continuously reports its metrics capabilities to the manager

### Manager Processing

When the manager receives a ClusterState message:

1. **Validation**: Ensures message integrity and edge authentication
2. **State Update**: Updates the consolidated view of the cluster
3. **Change Detection**: Identifies what has changed since the last sync
4. **Persistence**: Stores the updated state for query processing

## Connection Lifecycle

### Initial Connection

Edge processes establish streaming connections to the manager:

1. **Connection Setup**: Create bidirectional gRPC stream
2. **Cluster Identification**: Send cluster identification to claim responsibility for a specific cluster
3. **Acknowledgment**: Receive confirmation from manager
4. **Sync Initialization**: Begin periodic cluster state collection

### Cluster Identification

When an edge process connects, it must identify which Kubernetes cluster it will be responsible for:

- **Cluster ID**: Each edge declares a unique cluster identifier (e.g., "production-east", "staging-west")
- **Cluster Metadata**: Additional information about the cluster (region, environment, version)
- **Responsibility Claim**: The edge claims exclusive responsibility for syncing this cluster's state

### Connection Rejection Logic

The manager enforces a one-edge-per-cluster policy:

- **Duplicate Detection**: Manager tracks which clusters already have active edge connections
- **Rejection Response**: If a cluster already has an active edge, new connections are rejected

#### Rejection Scenarios

1. **Active Connection Exists**: Another edge is already syncing the same cluster
2. **Recent Disconnection**: Grace period after edge disconnection to prevent race conditions


## Error Scenarios

### Kubernetes API Failures

When local Kubernetes API queries fail:

- **Retry Logic**: Exponential backoff with configurable retry limits
- **Partial Data**: Send partial state if some queries succeed
- **Error Reporting**: Include error details in messages to manager
- **Fallback Behavior**: Continue with cached data if available

### Connection Failures

When the streaming connection to the manager fails:

- **Reconnection Attempts**: Automatic reconnection with backoff
- **Full Resync**: Complete cluster state sync once reconnected
- **Timeout Handling**: Configurable timeouts for connection attempts
- **Cluster Reclaim**: Re-identify cluster during reconnection to reclaim responsibility

### Manager-Side Connection Management

The manager maintains connection state for each cluster:

- **Active Connections**: Track which clusters have active edge connections
- **Connection Registry**: Map cluster IDs to active edge connections
- **Disconnection Cleanup**: Remove cluster registrations when edges disconnect
- **Grace Period**: Temporary hold on cluster assignments after disconnection

#### Failover Scenarios

When an edge connection fails:

1. **Immediate Cleanup**: Mark cluster as available for new connections
2. **Grace Period**: Short delay to prevent connection race conditions
3. **New Edge Registration**: Allow new edge to claim responsibility for the cluster
4. **State Continuity**: Maintain last known cluster state until new edge connects


## Configuration Options

### Sync Intervals

- **Default Interval**: 30 seconds between full cluster scans
- **Adaptive Timing**: Faster sync during high-change periods
- **Minimum Interval**: Prevent excessive API load
- **Maximum Interval**: Ensure timely updates

### Connection Parameters

- **Timeout Settings**: Connection and request timeouts
- **Retry Configuration**: Retry counts and backoff strategies
- **Buffer Sizes**: Message queuing limits
- **Keep-Alive Settings**: Heartbeat intervals
- **Max Message Size**: gRPC maximum message size limit (default 4MB may need adjustment for large clusters or clusters with extensive Istio configurations)

### Istio Resource Considerations

- **Payload Size Impact**: Istio resources can significantly increase sync message sizes, especially in clusters with complex service mesh configurations
- **Resource Filtering**: Edge processes collect all Istio resources across all namespaces; consider namespace-based filtering for very large clusters
- **Sync Performance**: Large numbers of Istio resources may require increased sync intervals or buffer sizes to prevent resource exhaustion
- **Control Plane Detection**: Istio control plane metadata is automatically detected and included in cluster state for proper resource interpretation


