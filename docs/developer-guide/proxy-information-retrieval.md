# Proxy Information Retrieval

This document describes Navigator's capability to retrieve and analyze proxy configuration from service mesh sidecars, primarily focusing on Envoy proxy integration within Istio service meshes.

## Overview

Navigator provides comprehensive proxy configuration retrieval and analysis through a multi-layered architecture:

1. **Manager-to-Edge Communication**: Manager service requests proxy configurations from edge services via bidirectional gRPC streaming
2. **Direct Proxy Access**: Edge services connect to Envoy admin endpoints to retrieve configuration dumps
3. **Configuration Analysis**: Raw Envoy configurations are parsed and summarized for easier consumption
4. **Frontend Integration**: Proxy information is exposed through REST APIs for UI consumption

## Architecture Components

### Manager Service Proxy Requests

The manager service can request proxy configuration from any connected edge service using the bidirectional gRPC streaming connection. The request flow involves:

1. **Request Initiation**: Manager sends a proxy configuration request with a unique correlation ID, target pod namespace, and pod name
2. **Edge Processing**: Edge service connects to the pod's Envoy admin interface to retrieve configuration
3. **Response Delivery**: Edge service responds with either the parsed proxy configuration or an error message

### Proxy Configuration Structure

Navigator defines a comprehensive proxy configuration model that summarizes complex Envoy configurations into structured, analyzable data. The configuration includes:

- **Version Information**: Proxy software version and build details
- **Raw Configuration**: Complete original configuration dump for debugging
- **Bootstrap Summary**: Essential startup configuration and node identification
- **Listener Summary**: Network listeners with type classification and filter chains
- **Cluster Summary**: Upstream service clusters with endpoint and health information
- **Route Summary**: HTTP routing rules with virtual hosts and traffic policies



### Frontend API Integration

Proxy configuration is accessible through the Frontend API for UI consumption. The API provides RESTful endpoints that allow web interfaces to:

- **Retrieve Instance Configuration**: Access proxy configuration for specific service instances
- **Browse Proxy Details**: Navigate through listeners, clusters, routes, and endpoints
- **Download Raw Configuration**: Access complete configuration dumps for debugging
- **Monitor Proxy Health**: Check proxy status and connectivity

### Multi-Cluster Coordination

In multi-cluster deployments:

1. **Request Routing**: Manager determines which edge service manages the target pod and sends the proxy configuration request
2. **Async Retrieval**: Edge services fetch configuration asynchronously from local Envoy proxies and respond via the stream
3. **Caching**: Configurations may be cached with appropriate TTL for performance
4. **Error Propagation**: Network and configuration errors are properly surfaced back to the manager and ultimately to clients



## Error Handling

Common error scenarios and handling:

1. **Proxy Not Found**: Pod doesn't have an Envoy sidecar
2. **Admin Interface Unreachable**: Network or firewall issues
3. **Configuration Parse Errors**: Malformed or unexpected configuration format
4. **Timeout Errors**: Slow proxy response or network latency
5. **Permission Errors**: Insufficient access to admin endpoints

Each error is properly categorized and includes actionable troubleshooting information for operators.