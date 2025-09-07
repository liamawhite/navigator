# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [backend/v1alpha1/clusterstate.proto](#backend_v1alpha1_clusterstate-proto)
    - [ClusterState](#navigator-backend-v1alpha1-ClusterState)
    - [Container](#navigator-backend-v1alpha1-Container)
    - [Service](#navigator-backend-v1alpha1-Service)
    - [ServiceInstance](#navigator-backend-v1alpha1-ServiceInstance)
    - [ServiceInstance.AnnotationsEntry](#navigator-backend-v1alpha1-ServiceInstance-AnnotationsEntry)
    - [ServiceInstance.LabelsEntry](#navigator-backend-v1alpha1-ServiceInstance-LabelsEntry)
  
    - [ProxyType](#navigator-backend-v1alpha1-ProxyType)
  
- [backend/v1alpha1/manager_service.proto](#backend_v1alpha1_manager_service-proto)
    - [ClusterIdentification](#navigator-backend-v1alpha1-ClusterIdentification)
    - [ConnectRequest](#navigator-backend-v1alpha1-ConnectRequest)
    - [ConnectResponse](#navigator-backend-v1alpha1-ConnectResponse)
    - [ConnectionAck](#navigator-backend-v1alpha1-ConnectionAck)
    - [EdgeCapabilities](#navigator-backend-v1alpha1-EdgeCapabilities)
    - [ErrorMessage](#navigator-backend-v1alpha1-ErrorMessage)
    - [ProxyConfigRequest](#navigator-backend-v1alpha1-ProxyConfigRequest)
    - [ProxyConfigResponse](#navigator-backend-v1alpha1-ProxyConfigResponse)
    - [ServiceGraphMetricsRequest](#navigator-backend-v1alpha1-ServiceGraphMetricsRequest)
    - [ServiceGraphMetricsResponse](#navigator-backend-v1alpha1-ServiceGraphMetricsResponse)
  
    - [ManagerService](#navigator-backend-v1alpha1-ManagerService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="backend_v1alpha1_clusterstate-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## backend/v1alpha1/clusterstate.proto



<a name="navigator-backend-v1alpha1-ClusterState"></a>

### ClusterState
ClusterState contains the current state of a cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| services | [Service](#navigator-backend-v1alpha1-Service) | repeated | services is the list of all services in the cluster. |
| destination_rules | [navigator.types.v1alpha1.DestinationRule](#navigator-types-v1alpha1-DestinationRule) | repeated | destination_rules is the list of all destination rules in the cluster. |
| envoy_filters | [navigator.types.v1alpha1.EnvoyFilter](#navigator-types-v1alpha1-EnvoyFilter) | repeated | envoy_filters is the list of all envoy filters in the cluster. |
| request_authentications | [navigator.types.v1alpha1.RequestAuthentication](#navigator-types-v1alpha1-RequestAuthentication) | repeated | request_authentications is the list of all request authentications in the cluster. |
| gateways | [navigator.types.v1alpha1.Gateway](#navigator-types-v1alpha1-Gateway) | repeated | gateways is the list of all gateways in the cluster. |
| sidecars | [navigator.types.v1alpha1.Sidecar](#navigator-types-v1alpha1-Sidecar) | repeated | sidecars is the list of all sidecars in the cluster. |
| virtual_services | [navigator.types.v1alpha1.VirtualService](#navigator-types-v1alpha1-VirtualService) | repeated | virtual_services is the list of all virtual services in the cluster. |
| istio_control_plane_config | [navigator.types.v1alpha1.IstioControlPlaneConfig](#navigator-types-v1alpha1-IstioControlPlaneConfig) |  | istio_control_plane_config contains Istio control plane configuration. |
| peer_authentications | [navigator.types.v1alpha1.PeerAuthentication](#navigator-types-v1alpha1-PeerAuthentication) | repeated | peer_authentications is the list of all peer authentications in the cluster. |
| authorization_policies | [navigator.types.v1alpha1.AuthorizationPolicy](#navigator-types-v1alpha1-AuthorizationPolicy) | repeated | authorization_policies is the list of all authorization policies in the cluster. |
| wasm_plugins | [navigator.types.v1alpha1.WasmPlugin](#navigator-types-v1alpha1-WasmPlugin) | repeated | wasm_plugins is the list of all wasm plugins in the cluster. |
| service_entries | [navigator.types.v1alpha1.ServiceEntry](#navigator-types-v1alpha1-ServiceEntry) | repeated | service_entries is the list of all service entries in the cluster. |






<a name="navigator-backend-v1alpha1-Container"></a>

### Container
Container represents a container running in a pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the container. |
| image | [string](#string) |  | image is the container image. |
| status | [string](#string) |  | status is the current status of the container (e.g., &#34;Running&#34;, &#34;Waiting&#34;, &#34;Terminated&#34;). |
| ready | [bool](#bool) |  | ready indicates whether the container is ready to serve requests. |
| restart_count | [int32](#int32) |  | restart_count is the number of times the container has been restarted. |






<a name="navigator-backend-v1alpha1-Service"></a>

### Service
Service represents a Kubernetes Service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the service. |
| namespace | [string](#string) |  | namespace is the namespace of the service. |
| instances | [ServiceInstance](#navigator-backend-v1alpha1-ServiceInstance) | repeated | instances is the list of service instances backing this service. |
| service_type | [navigator.types.v1alpha1.ServiceType](#navigator-types-v1alpha1-ServiceType) |  | service_type is the type of the service. |
| cluster_ip | [string](#string) |  | cluster_ip is the cluster IP address assigned to the service. |
| external_ip | [string](#string) |  | external_ip is the external IP address (for LoadBalancer services or manually assigned external IPs). |






<a name="navigator-backend-v1alpha1-ServiceInstance"></a>

### ServiceInstance
ServiceInstance represents a single instance of a service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ip | [string](#string) |  | ip is the IP address of the service instance. |
| pod_name | [string](#string) |  | pod_name is the name of the pod backing this service instance. |
| envoy_present | [bool](#bool) |  | envoy_present indicates whether an Envoy proxy is present in this instance. |
| containers | [Container](#navigator-backend-v1alpha1-Container) | repeated | containers is the list of containers running in this pod. |
| pod_status | [string](#string) |  | pod_status is the current status of the pod (e.g., &#34;Running&#34;, &#34;Pending&#34;). |
| node_name | [string](#string) |  | node_name is the name of the Kubernetes node hosting this pod. |
| created_at | [string](#string) |  | created_at is the timestamp when the pod was created. |
| labels | [ServiceInstance.LabelsEntry](#navigator-backend-v1alpha1-ServiceInstance-LabelsEntry) | repeated | labels are the Kubernetes labels assigned to the pod. |
| annotations | [ServiceInstance.AnnotationsEntry](#navigator-backend-v1alpha1-ServiceInstance-AnnotationsEntry) | repeated | annotations are the Kubernetes annotations assigned to the pod. |
| proxy_type | [ProxyType](#navigator-backend-v1alpha1-ProxyType) |  | proxy_type indicates the type of Istio proxy running in this instance. |






<a name="navigator-backend-v1alpha1-ServiceInstance-AnnotationsEntry"></a>

### ServiceInstance.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-backend-v1alpha1-ServiceInstance-LabelsEntry"></a>

### ServiceInstance.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="navigator-backend-v1alpha1-ProxyType"></a>

### ProxyType
ProxyType indicates the type of Istio proxy running in a service instance.

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNSPECIFIED | 0 | UNSPECIFIED indicates the proxy type is not specified or unknown. |
| NONE | 1 | NONE indicates no Istio proxy is present. |
| SIDECAR | 2 | SIDECAR indicates an Istio sidecar proxy is present. |
| GATEWAY | 3 | GATEWAY indicates an Istio gateway proxy is present. |


 

 

 



<a name="backend_v1alpha1_manager_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## backend/v1alpha1/manager_service.proto



<a name="navigator-backend-v1alpha1-ClusterIdentification"></a>

### ClusterIdentification
ClusterIdentification is sent by the edge process to identify which cluster it manages.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_id | [string](#string) |  | cluster_id is a unique identifier for the cluster this edge manages. |
| capabilities | [EdgeCapabilities](#navigator-backend-v1alpha1-EdgeCapabilities) |  | capabilities describe what features this edge process supports. |






<a name="navigator-backend-v1alpha1-ConnectRequest"></a>

### ConnectRequest
ConnectRequest represents messages sent from the edge process to the manager.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_identification | [ClusterIdentification](#navigator-backend-v1alpha1-ClusterIdentification) |  | cluster_identification is sent when the edge process connects to identify which cluster it manages. |
| cluster_state | [ClusterState](#navigator-backend-v1alpha1-ClusterState) |  | cluster_state contains the current state of the cluster. |
| proxy_config_response | [ProxyConfigResponse](#navigator-backend-v1alpha1-ProxyConfigResponse) |  | proxy_config_response is sent in response to a proxy config request from the manager. |
| service_graph_metrics_response | [ServiceGraphMetricsResponse](#navigator-backend-v1alpha1-ServiceGraphMetricsResponse) |  | service_graph_metrics_response is sent in response to a service graph metrics request from the manager. |






<a name="navigator-backend-v1alpha1-ConnectResponse"></a>

### ConnectResponse
ConnectResponse represents messages sent from the manager to the edge process.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connection_ack | [ConnectionAck](#navigator-backend-v1alpha1-ConnectionAck) |  | connection_ack acknowledges the cluster identification and indicates if the connection is accepted. once received, the edge process can start sending cluster state updates. |
| error | [ErrorMessage](#navigator-backend-v1alpha1-ErrorMessage) |  | error indicates an error condition. |
| proxy_config_request | [ProxyConfigRequest](#navigator-backend-v1alpha1-ProxyConfigRequest) |  | proxy_config_request asks the edge process to provide proxy config for a specific pod. |
| service_graph_metrics_request | [ServiceGraphMetricsRequest](#navigator-backend-v1alpha1-ServiceGraphMetricsRequest) |  | service_graph_metrics_request asks the edge process to provide service graph metrics. |






<a name="navigator-backend-v1alpha1-ConnectionAck"></a>

### ConnectionAck
ConnectionAck acknowledges cluster identification and indicates connection status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accepted | [bool](#bool) |  | accepted indicates whether the connection was accepted. |






<a name="navigator-backend-v1alpha1-EdgeCapabilities"></a>

### EdgeCapabilities
EdgeCapabilities describes what features an edge process supports.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics_enabled | [bool](#bool) |  | metrics_enabled indicates whether this edge process supports metrics collection. |






<a name="navigator-backend-v1alpha1-ErrorMessage"></a>

### ErrorMessage
ErrorMessage indicates an error condition.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| error_code | [string](#string) |  | error_code provides a machine-readable error identifier. |
| error_message | [string](#string) |  | error_message provides a human-readable error description. |






<a name="navigator-backend-v1alpha1-ProxyConfigRequest"></a>

### ProxyConfigRequest
ProxyConfigRequest is sent by the manager to request proxy configuration for a specific pod.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | request_id is a unique identifier for this request, used for correlating the response. |
| pod_namespace | [string](#string) |  | pod_namespace is the Kubernetes namespace of the pod. |
| pod_name | [string](#string) |  | pod_name is the Kubernetes name of the pod. |






<a name="navigator-backend-v1alpha1-ProxyConfigResponse"></a>

### ProxyConfigResponse
ProxyConfigResponse is sent by the edge process in response to a proxy config request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | request_id matches the request_id from the corresponding ProxyConfigRequest. |
| proxy_config | [navigator.types.v1alpha1.ProxyConfig](#navigator-types-v1alpha1-ProxyConfig) |  | proxy_config contains the proxy configuration for the requested pod. |
| error_message | [string](#string) |  | error_message indicates that the proxy config could not be retrieved. |






<a name="navigator-backend-v1alpha1-ServiceGraphMetricsRequest"></a>

### ServiceGraphMetricsRequest
ServiceGraphMetricsRequest is sent by the manager to request service graph metrics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | request_id is a unique identifier for this request, used for correlating the response. |
| filters | [navigator.types.v1alpha1.GraphMetricsFilters](#navigator-types-v1alpha1-GraphMetricsFilters) |  | filters specify which metrics to include based on source and destination attributes. |
| start_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | start_time specifies the start time for the metrics query (required). |
| end_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | end_time specifies the end time for the metrics query (required). |






<a name="navigator-backend-v1alpha1-ServiceGraphMetricsResponse"></a>

### ServiceGraphMetricsResponse
ServiceGraphMetricsResponse is sent by the edge process in response to a service graph metrics request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request_id | [string](#string) |  | request_id matches the request_id from the corresponding ServiceGraphMetricsRequest. |
| service_graph_metrics | [navigator.types.v1alpha1.ServiceGraphMetrics](#navigator-types-v1alpha1-ServiceGraphMetrics) |  | service_graph_metrics contains the service-to-service metrics data. |
| error_message | [string](#string) |  | error_message indicates that the service graph metrics could not be retrieved. |





 

 

 


<a name="navigator-backend-v1alpha1-ManagerService"></a>

### ManagerService
ManagerService provides bidirectional streaming communication between edge processes and the central manager.
This enables edge processes to sync cluster state information to the manager.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Connect | [ConnectRequest](#navigator-backend-v1alpha1-ConnectRequest) stream | [ConnectResponse](#navigator-backend-v1alpha1-ConnectResponse) stream | Connect establishes a bidirectional streaming connection between an edge process and the manager. The edge process identifies its cluster and sends periodic cluster state updates. |

 



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

