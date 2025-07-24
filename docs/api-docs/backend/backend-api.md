# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [backend/v1alpha1/clusterstate.proto](#backend_v1alpha1_clusterstate-proto)
    - [ClusterState](#navigator-backend-v1alpha1-ClusterState)
    - [Container](#navigator-backend-v1alpha1-Container)
    - [DestinationRule](#navigator-backend-v1alpha1-DestinationRule)
    - [DestinationRuleSubset](#navigator-backend-v1alpha1-DestinationRuleSubset)
    - [DestinationRuleSubset.LabelsEntry](#navigator-backend-v1alpha1-DestinationRuleSubset-LabelsEntry)
    - [EnvoyFilter](#navigator-backend-v1alpha1-EnvoyFilter)
    - [Gateway](#navigator-backend-v1alpha1-Gateway)
    - [Gateway.SelectorEntry](#navigator-backend-v1alpha1-Gateway-SelectorEntry)
    - [IstioControlPlaneConfig](#navigator-backend-v1alpha1-IstioControlPlaneConfig)
    - [PolicyTargetReference](#navigator-backend-v1alpha1-PolicyTargetReference)
    - [Service](#navigator-backend-v1alpha1-Service)
    - [ServiceInstance](#navigator-backend-v1alpha1-ServiceInstance)
    - [ServiceInstance.AnnotationsEntry](#navigator-backend-v1alpha1-ServiceInstance-AnnotationsEntry)
    - [ServiceInstance.LabelsEntry](#navigator-backend-v1alpha1-ServiceInstance-LabelsEntry)
    - [Sidecar](#navigator-backend-v1alpha1-Sidecar)
    - [VirtualService](#navigator-backend-v1alpha1-VirtualService)
    - [WorkloadSelector](#navigator-backend-v1alpha1-WorkloadSelector)
    - [WorkloadSelector.MatchLabelsEntry](#navigator-backend-v1alpha1-WorkloadSelector-MatchLabelsEntry)
  
- [backend/v1alpha1/manager_service.proto](#backend_v1alpha1_manager_service-proto)
    - [ClusterIdentification](#navigator-backend-v1alpha1-ClusterIdentification)
    - [ConnectRequest](#navigator-backend-v1alpha1-ConnectRequest)
    - [ConnectResponse](#navigator-backend-v1alpha1-ConnectResponse)
    - [ConnectionAck](#navigator-backend-v1alpha1-ConnectionAck)
    - [ErrorMessage](#navigator-backend-v1alpha1-ErrorMessage)
    - [ProxyConfigRequest](#navigator-backend-v1alpha1-ProxyConfigRequest)
    - [ProxyConfigResponse](#navigator-backend-v1alpha1-ProxyConfigResponse)
  
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
| destination_rules | [DestinationRule](#navigator-backend-v1alpha1-DestinationRule) | repeated | destination_rules is the list of all destination rules in the cluster. |
| envoy_filters | [EnvoyFilter](#navigator-backend-v1alpha1-EnvoyFilter) | repeated | envoy_filters is the list of all envoy filters in the cluster. |
| gateways | [Gateway](#navigator-backend-v1alpha1-Gateway) | repeated | gateways is the list of all gateways in the cluster. |
| sidecars | [Sidecar](#navigator-backend-v1alpha1-Sidecar) | repeated | sidecars is the list of all sidecars in the cluster. |
| virtual_services | [VirtualService](#navigator-backend-v1alpha1-VirtualService) | repeated | virtual_services is the list of all virtual services in the cluster. |
| istio_control_plane_config | [IstioControlPlaneConfig](#navigator-backend-v1alpha1-IstioControlPlaneConfig) |  | istio_control_plane_config contains Istio control plane configuration. |






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






<a name="navigator-backend-v1alpha1-DestinationRule"></a>

### DestinationRule
DestinationRule represents an Istio DestinationRule resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the destination rule. |
| namespace | [string](#string) |  | namespace is the namespace of the destination rule. |
| raw_spec | [string](#string) |  | raw_spec is the destination rule spec as a JSON string. |
| host | [string](#string) |  | host is the name of a service from the service registry. |
| subsets | [DestinationRuleSubset](#navigator-backend-v1alpha1-DestinationRuleSubset) | repeated | subsets is the list of named subsets for traffic routing. |
| export_to | [string](#string) | repeated | export_to controls the visibility of this destination rule to other namespaces. |
| workload_selector | [WorkloadSelector](#navigator-backend-v1alpha1-WorkloadSelector) |  | workload_selector is the criteria used to select the specific set of pods/VMs. |






<a name="navigator-backend-v1alpha1-DestinationRuleSubset"></a>

### DestinationRuleSubset
DestinationRuleSubset represents a named subset for destination rule traffic routing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the subset. |
| labels | [DestinationRuleSubset.LabelsEntry](#navigator-backend-v1alpha1-DestinationRuleSubset-LabelsEntry) | repeated | labels are the key-value pairs that define the subset. |






<a name="navigator-backend-v1alpha1-DestinationRuleSubset-LabelsEntry"></a>

### DestinationRuleSubset.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-backend-v1alpha1-EnvoyFilter"></a>

### EnvoyFilter
EnvoyFilter represents an Istio EnvoyFilter resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the envoy filter. |
| namespace | [string](#string) |  | namespace is the namespace of the envoy filter. |
| raw_spec | [string](#string) |  | raw_spec is the envoy filter spec as a JSON string. |
| workload_selector | [WorkloadSelector](#navigator-backend-v1alpha1-WorkloadSelector) |  | workload_selector is the criteria used to select the specific set of pods/VMs. |
| target_refs | [PolicyTargetReference](#navigator-backend-v1alpha1-PolicyTargetReference) | repeated | target_refs is the list of resources that this envoy filter applies to. |






<a name="navigator-backend-v1alpha1-Gateway"></a>

### Gateway
Gateway represents an Istio Gateway resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the gateway. |
| namespace | [string](#string) |  | namespace is the namespace of the gateway. |
| raw_spec | [string](#string) |  | raw_spec is the gateway spec as a JSON string. |
| selector | [Gateway.SelectorEntry](#navigator-backend-v1alpha1-Gateway-SelectorEntry) | repeated | selector is the workload selector for the gateway. |






<a name="navigator-backend-v1alpha1-Gateway-SelectorEntry"></a>

### Gateway.SelectorEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-backend-v1alpha1-IstioControlPlaneConfig"></a>

### IstioControlPlaneConfig
IstioControlPlaneConfig represents configuration from the Istio control plane.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pilot_scope_gateway_to_namespace | [bool](#bool) |  | pilot_scope_gateway_to_namespace indicates whether gateway selector scope is restricted to namespace. When true, gateway selectors only match workloads in the same namespace as the gateway. When false (default), gateway selectors match workloads across all namespaces. |






<a name="navigator-backend-v1alpha1-PolicyTargetReference"></a>

### PolicyTargetReference
PolicyTargetReference represents a reference to a specific resource based on Istio&#39;s PolicyTargetReference.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) |  | group specifies the group of the target resource. |
| kind | [string](#string) |  | kind indicates the kind of target resource (required). |
| name | [string](#string) |  | name provides the name of the target resource (required). |
| namespace | [string](#string) |  | namespace defines the namespace of the referenced resource. When unspecified, the local namespace is inferred. |






<a name="navigator-backend-v1alpha1-Service"></a>

### Service
Service represents a Kubernetes Service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the service. |
| namespace | [string](#string) |  | namespace is the namespace of the service. |
| instances | [ServiceInstance](#navigator-backend-v1alpha1-ServiceInstance) | repeated | instances is the list of service instances backing this service. |






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






<a name="navigator-backend-v1alpha1-Sidecar"></a>

### Sidecar
Sidecar represents an Istio Sidecar resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the sidecar. |
| namespace | [string](#string) |  | namespace is the namespace of the sidecar. |
| raw_spec | [string](#string) |  | raw_spec is the sidecar spec as a JSON string. |






<a name="navigator-backend-v1alpha1-VirtualService"></a>

### VirtualService
VirtualService represents an Istio VirtualService resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the virtual service. |
| namespace | [string](#string) |  | namespace is the namespace of the virtual service. |
| raw_spec | [string](#string) |  | raw_spec is the virtual service spec as a JSON string. |
| hosts | [string](#string) | repeated | hosts is the list of destination hosts that these routing rules apply to. |
| gateways | [string](#string) | repeated | gateways is the list of gateway names that should apply these routes. |
| export_to | [string](#string) | repeated | export_to controls the visibility of this virtual service to other namespaces. |






<a name="navigator-backend-v1alpha1-WorkloadSelector"></a>

### WorkloadSelector
WorkloadSelector represents the workload selector criteria used across Istio resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match_labels | [WorkloadSelector.MatchLabelsEntry](#navigator-backend-v1alpha1-WorkloadSelector-MatchLabelsEntry) | repeated | match_labels are the labels used to select pods/VMs. |






<a name="navigator-backend-v1alpha1-WorkloadSelector-MatchLabelsEntry"></a>

### WorkloadSelector.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 

 



<a name="backend_v1alpha1_manager_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## backend/v1alpha1/manager_service.proto



<a name="navigator-backend-v1alpha1-ClusterIdentification"></a>

### ClusterIdentification
ClusterIdentification is sent by the edge process to identify which cluster it manages.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_id | [string](#string) |  | cluster_id is a unique identifier for the cluster this edge manages. |






<a name="navigator-backend-v1alpha1-ConnectRequest"></a>

### ConnectRequest
ConnectRequest represents messages sent from the edge process to the manager.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_identification | [ClusterIdentification](#navigator-backend-v1alpha1-ClusterIdentification) |  | cluster_identification is sent when the edge process connects to identify which cluster it manages. |
| cluster_state | [ClusterState](#navigator-backend-v1alpha1-ClusterState) |  | cluster_state contains the current state of the cluster. |
| proxy_config_response | [ProxyConfigResponse](#navigator-backend-v1alpha1-ProxyConfigResponse) |  | proxy_config_response is sent in response to a proxy config request from the manager. |






<a name="navigator-backend-v1alpha1-ConnectResponse"></a>

### ConnectResponse
ConnectResponse represents messages sent from the manager to the edge process.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connection_ack | [ConnectionAck](#navigator-backend-v1alpha1-ConnectionAck) |  | connection_ack acknowledges the cluster identification and indicates if the connection is accepted. once received, the edge process can start sending cluster state updates. |
| error | [ErrorMessage](#navigator-backend-v1alpha1-ErrorMessage) |  | error indicates an error condition. |
| proxy_config_request | [ProxyConfigRequest](#navigator-backend-v1alpha1-ProxyConfigRequest) |  | proxy_config_request asks the edge process to provide proxy config for a specific pod. |






<a name="navigator-backend-v1alpha1-ConnectionAck"></a>

### ConnectionAck
ConnectionAck acknowledges cluster identification and indicates connection status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| accepted | [bool](#bool) |  | accepted indicates whether the connection was accepted. |






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

