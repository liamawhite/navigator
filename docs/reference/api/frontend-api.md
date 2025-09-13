# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [frontend/v1alpha1/cluster_registry.proto](#frontend_v1alpha1_cluster_registry-proto)
    - [ClusterSyncInfo](#navigator-frontend-v1alpha1-ClusterSyncInfo)
    - [ListClustersRequest](#navigator-frontend-v1alpha1-ListClustersRequest)
    - [ListClustersResponse](#navigator-frontend-v1alpha1-ListClustersResponse)
  
    - [SyncStatus](#navigator-frontend-v1alpha1-SyncStatus)
  
    - [ClusterRegistryService](#navigator-frontend-v1alpha1-ClusterRegistryService)
  
- [frontend/v1alpha1/metrics_service.proto](#frontend_v1alpha1_metrics_service-proto)
    - [GetServiceConnectionsRequest](#navigator-frontend-v1alpha1-GetServiceConnectionsRequest)
    - [GetServiceConnectionsResponse](#navigator-frontend-v1alpha1-GetServiceConnectionsResponse)
  
    - [MetricsService](#navigator-frontend-v1alpha1-MetricsService)
  
- [frontend/v1alpha1/service_registry.proto](#frontend_v1alpha1_service_registry-proto)
    - [Container](#navigator-frontend-v1alpha1-Container)
    - [GetIstioResourcesRequest](#navigator-frontend-v1alpha1-GetIstioResourcesRequest)
    - [GetIstioResourcesResponse](#navigator-frontend-v1alpha1-GetIstioResourcesResponse)
    - [GetProxyConfigRequest](#navigator-frontend-v1alpha1-GetProxyConfigRequest)
    - [GetProxyConfigResponse](#navigator-frontend-v1alpha1-GetProxyConfigResponse)
    - [GetServiceInstanceRequest](#navigator-frontend-v1alpha1-GetServiceInstanceRequest)
    - [GetServiceInstanceResponse](#navigator-frontend-v1alpha1-GetServiceInstanceResponse)
    - [GetServiceRequest](#navigator-frontend-v1alpha1-GetServiceRequest)
    - [GetServiceResponse](#navigator-frontend-v1alpha1-GetServiceResponse)
    - [ListServicesRequest](#navigator-frontend-v1alpha1-ListServicesRequest)
    - [ListServicesResponse](#navigator-frontend-v1alpha1-ListServicesResponse)
    - [Service](#navigator-frontend-v1alpha1-Service)
    - [Service.ClusterIpsEntry](#navigator-frontend-v1alpha1-Service-ClusterIpsEntry)
    - [Service.ExternalIpsEntry](#navigator-frontend-v1alpha1-Service-ExternalIpsEntry)
    - [ServiceInstance](#navigator-frontend-v1alpha1-ServiceInstance)
    - [ServiceInstanceDetail](#navigator-frontend-v1alpha1-ServiceInstanceDetail)
    - [ServiceInstanceDetail.AnnotationsEntry](#navigator-frontend-v1alpha1-ServiceInstanceDetail-AnnotationsEntry)
    - [ServiceInstanceDetail.LabelsEntry](#navigator-frontend-v1alpha1-ServiceInstanceDetail-LabelsEntry)
  
    - [ServiceRegistryService](#navigator-frontend-v1alpha1-ServiceRegistryService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="frontend_v1alpha1_cluster_registry-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## frontend/v1alpha1/cluster_registry.proto



<a name="navigator-frontend-v1alpha1-ClusterSyncInfo"></a>

### ClusterSyncInfo
ClusterSyncInfo contains synchronization status and metadata for a connected cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_id | [string](#string) |  | cluster_id uniquely identifies this cluster. |
| connected_at | [string](#string) |  | connected_at is when this cluster initially connected to the manager (RFC3339 format). |
| last_update | [string](#string) |  | last_update is when the last sync occurred (RFC3339 format). |
| service_count | [int32](#int32) |  | service_count is the number of services currently synced from this cluster. |
| sync_status | [SyncStatus](#navigator-frontend-v1alpha1-SyncStatus) |  | sync_status indicates the health of the sync based on last_update timing. |
| metrics_enabled | [bool](#bool) |  | metrics_enabled indicates whether this cluster&#39;s edge supports metrics collection. |






<a name="navigator-frontend-v1alpha1-ListClustersRequest"></a>

### ListClustersRequest
ListClustersRequest for retrieving cluster sync information.

Currently no filters needed, but structured for future extensibility.






<a name="navigator-frontend-v1alpha1-ListClustersResponse"></a>

### ListClustersResponse
ListClustersResponse contains the list of all connected clusters and their sync status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusters | [ClusterSyncInfo](#navigator-frontend-v1alpha1-ClusterSyncInfo) | repeated | clusters contains information about each connected cluster&#39;s sync status. |





 


<a name="navigator-frontend-v1alpha1-SyncStatus"></a>

### SyncStatus
SyncStatus represents the health of cluster synchronization.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SYNC_STATUS_UNSPECIFIED | 0 |  |
| SYNC_STATUS_INITIALIZING | 1 | Connected but hasn&#39;t received full state yet |
| SYNC_STATUS_HEALTHY | 2 | Recent updates within expected timeframe |
| SYNC_STATUS_STALE | 3 | No recent updates, potentially problematic |
| SYNC_STATUS_DISCONNECTED | 4 | Connection lost |


 

 


<a name="navigator-frontend-v1alpha1-ClusterRegistryService"></a>

### ClusterRegistryService
ClusterRegistryService provides APIs for cluster management and monitoring.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListClusters | [ListClustersRequest](#navigator-frontend-v1alpha1-ListClustersRequest) | [ListClustersResponse](#navigator-frontend-v1alpha1-ListClustersResponse) | ListClusters returns sync state information for all connected clusters. |

 



<a name="frontend_v1alpha1_metrics_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## frontend/v1alpha1/metrics_service.proto



<a name="navigator-frontend-v1alpha1-GetServiceConnectionsRequest"></a>

### GetServiceConnectionsRequest
GetServiceConnectionsRequest specifies a service for connection metrics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_name | [string](#string) |  | service_name is the name of the service to get connections for. |
| namespace | [string](#string) |  | namespace is the Kubernetes namespace of the service. |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | start_time specifies the start time for the metrics query (required). Must be in the past (before current time). |
| end_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | end_time specifies the end time for the metrics query (required). Must be in the past (before current time) and after start_time. |






<a name="navigator-frontend-v1alpha1-GetServiceConnectionsResponse"></a>

### GetServiceConnectionsResponse
GetServiceConnectionsResponse contains inbound and outbound service connections.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inbound | [navigator.types.v1alpha1.ServicePairMetrics](#navigator-types-v1alpha1-ServicePairMetrics) | repeated | inbound contains services that call this service. |
| outbound | [navigator.types.v1alpha1.ServicePairMetrics](#navigator-types-v1alpha1-ServicePairMetrics) | repeated | outbound contains services that this service calls. |
| timestamp | [string](#string) |  | timestamp is when these metrics were collected (RFC3339 format). |
| clusters_queried | [string](#string) | repeated | clusters_queried lists the clusters that were queried for these metrics. |





 

 

 


<a name="navigator-frontend-v1alpha1-MetricsService"></a>

### MetricsService
MetricsService provides APIs for service mesh metrics and observability.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetServiceConnections | [GetServiceConnectionsRequest](#navigator-frontend-v1alpha1-GetServiceConnectionsRequest) | [GetServiceConnectionsResponse](#navigator-frontend-v1alpha1-GetServiceConnectionsResponse) | GetServiceConnections returns inbound and outbound connections for a specific service. |

 



<a name="frontend_v1alpha1_service_registry-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## frontend/v1alpha1/service_registry.proto



<a name="navigator-frontend-v1alpha1-Container"></a>

### Container
Container represents a container running in a pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the container. |
| image | [string](#string) |  | image is the container image. |
| status | [string](#string) |  | status is the current status of the container (e.g., &#34;Running&#34;, &#34;Waiting&#34;, &#34;Terminated&#34;). |
| ready | [bool](#bool) |  | ready indicates whether the container is ready to serve requests. |
| restart_count | [int32](#int32) |  | restart_count is the number of times the container has been restarted. |






<a name="navigator-frontend-v1alpha1-GetIstioResourcesRequest"></a>

### GetIstioResourcesRequest
GetIstioResourcesRequest specifies which service instance&#39;s Istio resources to retrieve.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_id | [string](#string) |  | service_id is the unique identifier of the service. Format: namespace:service-name (e.g., &#34;default:nginx-service&#34;) |
| instance_id | [string](#string) |  | instance_id is the unique identifier of the service instance. Format: cluster_id:namespace:pod_name (e.g., &#34;cluster1:default:nginx-pod-123&#34;) |






<a name="navigator-frontend-v1alpha1-GetIstioResourcesResponse"></a>

### GetIstioResourcesResponse
GetIstioResourcesResponse contains the Istio resources for the requested service instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| virtual_services | [navigator.types.v1alpha1.VirtualService](#navigator-types-v1alpha1-VirtualService) | repeated | virtual_services are VirtualService resources affecting this instance. |
| destination_rules | [navigator.types.v1alpha1.DestinationRule](#navigator-types-v1alpha1-DestinationRule) | repeated | destination_rules are DestinationRule resources affecting this instance. |
| gateways | [navigator.types.v1alpha1.Gateway](#navigator-types-v1alpha1-Gateway) | repeated | gateways are Gateway resources affecting this instance. |
| sidecars | [navigator.types.v1alpha1.Sidecar](#navigator-types-v1alpha1-Sidecar) | repeated | sidecars are Sidecar resources affecting this instance. |
| envoy_filters | [navigator.types.v1alpha1.EnvoyFilter](#navigator-types-v1alpha1-EnvoyFilter) | repeated | envoy_filters are EnvoyFilter resources affecting this instance. |
| request_authentications | [navigator.types.v1alpha1.RequestAuthentication](#navigator-types-v1alpha1-RequestAuthentication) | repeated | request_authentications are RequestAuthentication resources affecting this instance. |
| peer_authentications | [navigator.types.v1alpha1.PeerAuthentication](#navigator-types-v1alpha1-PeerAuthentication) | repeated | peer_authentications are PeerAuthentication resources affecting this instance. |
| authorization_policies | [navigator.types.v1alpha1.AuthorizationPolicy](#navigator-types-v1alpha1-AuthorizationPolicy) | repeated | authorization_policies are AuthorizationPolicy resources affecting this instance. |
| wasm_plugins | [navigator.types.v1alpha1.WasmPlugin](#navigator-types-v1alpha1-WasmPlugin) | repeated | wasm_plugins are WasmPlugin resources affecting this instance. |
| service_entries | [navigator.types.v1alpha1.ServiceEntry](#navigator-types-v1alpha1-ServiceEntry) | repeated | service_entries are ServiceEntry resources affecting this instance. |






<a name="navigator-frontend-v1alpha1-GetProxyConfigRequest"></a>

### GetProxyConfigRequest
GetProxyConfigRequest specifies which service instance&#39;s proxy configuration to retrieve.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_id | [string](#string) |  | service_id is the unique identifier of the service. Format: namespace:service-name (e.g., &#34;default:nginx-service&#34;) |
| instance_id | [string](#string) |  | instance_id is the unique identifier of the service instance. Format: cluster_id:namespace:pod_name (e.g., &#34;cluster1:default:nginx-pod-123&#34;) |






<a name="navigator-frontend-v1alpha1-GetProxyConfigResponse"></a>

### GetProxyConfigResponse
GetProxyConfigResponse contains the proxy configuration for the requested pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proxy_config | [navigator.types.v1alpha1.ProxyConfig](#navigator-types-v1alpha1-ProxyConfig) |  | proxy_config contains the complete Envoy proxy configuration. |






<a name="navigator-frontend-v1alpha1-GetServiceInstanceRequest"></a>

### GetServiceInstanceRequest
GetServiceInstanceRequest specifies which service instance to retrieve.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_id | [string](#string) |  | service_id is the unique identifier of the service. Format: namespace:service-name (e.g., &#34;default:nginx-service&#34;) |
| instance_id | [string](#string) |  | instance_id is the unique identifier of the specific service instance. Format: cluster_name:namespace:pod_name (e.g., &#34;prod-west:default:nginx-abc123&#34;) |






<a name="navigator-frontend-v1alpha1-GetServiceInstanceResponse"></a>

### GetServiceInstanceResponse
GetServiceInstanceResponse contains the requested service instance details.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [ServiceInstanceDetail](#navigator-frontend-v1alpha1-ServiceInstanceDetail) |  | instance contains the detailed service instance information. |






<a name="navigator-frontend-v1alpha1-GetServiceRequest"></a>

### GetServiceRequest
GetServiceRequest specifies which service to retrieve.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | id is the unique identifier of the service to retrieve. Format: namespace:service-name (e.g., &#34;default:nginx-service&#34;) |






<a name="navigator-frontend-v1alpha1-GetServiceResponse"></a>

### GetServiceResponse
GetServiceResponse contains the requested service details.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service | [Service](#navigator-frontend-v1alpha1-Service) |  | service contains the detailed service information. |






<a name="navigator-frontend-v1alpha1-ListServicesRequest"></a>

### ListServicesRequest
ListServicesRequest specifies which namespace to list services from.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | [string](#string) | optional | namespace is the Kubernetes namespace to list services from. If not specified, services from all namespaces are returned. |
| cluster_id | [string](#string) | optional | cluster_id filters services to only those from the specified cluster. If not specified, services from all connected clusters are returned. |






<a name="navigator-frontend-v1alpha1-ListServicesResponse"></a>

### ListServicesResponse
ListServicesResponse contains the list of services in the requested namespace(s).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [Service](#navigator-frontend-v1alpha1-Service) | repeated | services is the list of services found in the namespace(s). |






<a name="navigator-frontend-v1alpha1-Service"></a>

### Service
Service represents a Kubernetes service with its backing instances.
Services in different clusters that share the same name and namespace are considered the same service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | id is a unique identifier for the service in format namespace:service-name (e.g., &#34;default:nginx-service&#34;). |
| name | [string](#string) |  | name is the service name. |
| namespace | [string](#string) |  | namespace is the Kubernetes namespace containing the service. |
| instances | [ServiceInstance](#navigator-frontend-v1alpha1-ServiceInstance) | repeated | instances are the backend instances (pods) that serve this service across all clusters. |
| cluster_ips | [Service.ClusterIpsEntry](#navigator-frontend-v1alpha1-Service-ClusterIpsEntry) | repeated | cluster_ips maps cluster names to their cluster IP addresses for this service. |
| external_ips | [Service.ExternalIpsEntry](#navigator-frontend-v1alpha1-Service-ExternalIpsEntry) | repeated | external_ips maps cluster names to their external IP addresses for this service. |






<a name="navigator-frontend-v1alpha1-Service-ClusterIpsEntry"></a>

### Service.ClusterIpsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-frontend-v1alpha1-Service-ExternalIpsEntry"></a>

### Service.ExternalIpsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-frontend-v1alpha1-ServiceInstance"></a>

### ServiceInstance
ServiceInstance represents a single backend instance serving a service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance_id | [string](#string) |  | instance_id is a unique identifier for this instance across all clusters. Format: cluster_name:namespace:pod_name (e.g., &#34;prod-west:default:nginx-abc123&#34;) |
| ip | [string](#string) |  | ip is the IP address of the instance. |
| pod_name | [string](#string) |  | pod_name is the name of the Kubernetes pod backing this instance. |
| namespace | [string](#string) |  | namespace is the Kubernetes namespace containing the pod. |
| cluster_name | [string](#string) |  | cluster_name is the name of the Kubernetes cluster this instance belongs to. |
| envoy_present | [bool](#bool) |  | envoy_present indicates whether this instance has an Envoy proxy sidecar. |






<a name="navigator-frontend-v1alpha1-ServiceInstanceDetail"></a>

### ServiceInstanceDetail
ServiceInstanceDetail represents detailed information about a specific service instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance_id | [string](#string) |  | instance_id is a unique identifier for this instance across all clusters. Format: cluster_name:namespace:pod_name (e.g., &#34;prod-west:default:nginx-abc123&#34;) |
| ip | [string](#string) |  | ip is the IP address of the instance. |
| pod_name | [string](#string) |  | pod_name is the name of the Kubernetes pod backing this instance. |
| namespace | [string](#string) |  | namespace is the Kubernetes namespace containing the pod. |
| cluster_name | [string](#string) |  | cluster_name is the name of the Kubernetes cluster this instance belongs to. |
| envoy_present | [bool](#bool) |  | envoy_present indicates whether this instance has an Envoy proxy sidecar. |
| service_name | [string](#string) |  | service_name is the name of the service this instance belongs to. |
| containers | [Container](#navigator-frontend-v1alpha1-Container) | repeated | containers is the list of containers running in this pod. |
| pod_status | [string](#string) |  | pod_status is the current status of the pod (e.g., &#34;Running&#34;, &#34;Pending&#34;). |
| node_name | [string](#string) |  | node_name is the name of the Kubernetes node hosting this pod. |
| created_at | [string](#string) |  | created_at is the timestamp when the pod was created. |
| labels | [ServiceInstanceDetail.LabelsEntry](#navigator-frontend-v1alpha1-ServiceInstanceDetail-LabelsEntry) | repeated | labels are the Kubernetes labels assigned to the pod. |
| annotations | [ServiceInstanceDetail.AnnotationsEntry](#navigator-frontend-v1alpha1-ServiceInstanceDetail-AnnotationsEntry) | repeated | annotations are the Kubernetes annotations assigned to the pod. |
| is_envoy_present | [bool](#bool) |  | is_envoy_present indicates whether this instance has an Envoy proxy sidecar. |






<a name="navigator-frontend-v1alpha1-ServiceInstanceDetail-AnnotationsEntry"></a>

### ServiceInstanceDetail.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-frontend-v1alpha1-ServiceInstanceDetail-LabelsEntry"></a>

### ServiceInstanceDetail.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 


<a name="navigator-frontend-v1alpha1-ServiceRegistryService"></a>

### ServiceRegistryService
ServiceRegistryService provides APIs for service discovery and management.
It enables listing and retrieving services from multiple Kubernetes clusters via the manager&#39;s aggregated view.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListServices | [ListServicesRequest](#navigator-frontend-v1alpha1-ListServicesRequest) | [ListServicesResponse](#navigator-frontend-v1alpha1-ListServicesResponse) | ListServices returns all services in the specified namespace, or all namespaces if not specified. Services are aggregated across all connected clusters. |
| GetService | [GetServiceRequest](#navigator-frontend-v1alpha1-GetServiceRequest) | [GetServiceResponse](#navigator-frontend-v1alpha1-GetServiceResponse) | GetService returns detailed information about a specific service. The service may have instances across multiple clusters. |
| GetServiceInstance | [GetServiceInstanceRequest](#navigator-frontend-v1alpha1-GetServiceInstanceRequest) | [GetServiceInstanceResponse](#navigator-frontend-v1alpha1-GetServiceInstanceResponse) | GetServiceInstance returns detailed information about a specific service instance. |
| GetProxyConfig | [GetProxyConfigRequest](#navigator-frontend-v1alpha1-GetProxyConfigRequest) | [GetProxyConfigResponse](#navigator-frontend-v1alpha1-GetProxyConfigResponse) | GetProxyConfig retrieves the Envoy proxy configuration for a specific service instance. |
| GetIstioResources | [GetIstioResourcesRequest](#navigator-frontend-v1alpha1-GetIstioResourcesRequest) | [GetIstioResourcesResponse](#navigator-frontend-v1alpha1-GetIstioResourcesResponse) | GetIstioResources retrieves the Istio configuration resources for a specific service instance. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

