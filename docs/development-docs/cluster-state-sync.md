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

### Data Packaging

The collected Kubernetes resources are packaged into a ClusterState message that includes:

- **Services**: Service definitions including names, namespaces, selectors, and ports
- **EndpointSlices**: Endpoint information including IP addresses, ports, and readiness status
- **Pods**: Pod metadata including names, namespaces, labels, and current phase

### Transmission to Manager

The ClusterState message is sent to the manager over the established streaming connection:

- **Streaming Protocol**: Uses the bidirectional gRPC stream for real-time delivery
- **Message Identification**: Each message includes edge identification and timestamp
- **Acknowledgment**: Manager acknowledges receipt to ensure reliable delivery

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
- **Max Message Size**: gRPC maximum message size limit (default 4MB may need adjustment for large clusters)


