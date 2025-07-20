# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [types/v1alpha1/proxy_types.proto](#types_v1alpha1_proxy_types-proto)
    - [BootstrapSummary](#navigator-types-v1alpha1-BootstrapSummary)
    - [ClusterManagerInfo](#navigator-types-v1alpha1-ClusterManagerInfo)
    - [ClusterSummary](#navigator-types-v1alpha1-ClusterSummary)
    - [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo)
    - [DynamicConfigInfo](#navigator-types-v1alpha1-DynamicConfigInfo)
    - [EndpointInfo](#navigator-types-v1alpha1-EndpointInfo)
    - [EndpointInfo.MetadataEntry](#navigator-types-v1alpha1-EndpointInfo-MetadataEntry)
    - [EndpointSummary](#navigator-types-v1alpha1-EndpointSummary)
    - [ListenerSummary](#navigator-types-v1alpha1-ListenerSummary)
    - [LocalityInfo](#navigator-types-v1alpha1-LocalityInfo)
    - [NodeSummary](#navigator-types-v1alpha1-NodeSummary)
    - [NodeSummary.MetadataEntry](#navigator-types-v1alpha1-NodeSummary-MetadataEntry)
    - [ProxyConfig](#navigator-types-v1alpha1-ProxyConfig)
    - [RouteActionInfo](#navigator-types-v1alpha1-RouteActionInfo)
    - [RouteConfigSummary](#navigator-types-v1alpha1-RouteConfigSummary)
    - [RouteInfo](#navigator-types-v1alpha1-RouteInfo)
    - [RouteMatchInfo](#navigator-types-v1alpha1-RouteMatchInfo)
    - [VirtualHostInfo](#navigator-types-v1alpha1-VirtualHostInfo)
    - [WeightedClusterInfo](#navigator-types-v1alpha1-WeightedClusterInfo)
    - [WeightedClusterInfo.MetadataMatchEntry](#navigator-types-v1alpha1-WeightedClusterInfo-MetadataMatchEntry)
  
    - [AddressType](#navigator-types-v1alpha1-AddressType)
    - [ClusterDirection](#navigator-types-v1alpha1-ClusterDirection)
    - [ClusterType](#navigator-types-v1alpha1-ClusterType)
    - [ListenerType](#navigator-types-v1alpha1-ListenerType)
    - [ProxyMode](#navigator-types-v1alpha1-ProxyMode)
    - [RouteType](#navigator-types-v1alpha1-RouteType)
  
- [Scalar Value Types](#scalar-value-types)



<a name="types_v1alpha1_proxy_types-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## types/v1alpha1/proxy_types.proto



<a name="navigator-types-v1alpha1-BootstrapSummary"></a>

### BootstrapSummary
BootstrapSummary contains essential bootstrap configuration information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node | [NodeSummary](#navigator-types-v1alpha1-NodeSummary) |  |  |
| static_resources_version | [string](#string) |  |  |
| dynamic_resources_config | [DynamicConfigInfo](#navigator-types-v1alpha1-DynamicConfigInfo) |  |  |
| admin_port | [uint32](#uint32) |  |  |
| admin_address | [string](#string) |  |  |
| cluster_manager | [ClusterManagerInfo](#navigator-types-v1alpha1-ClusterManagerInfo) |  |  |






<a name="navigator-types-v1alpha1-ClusterManagerInfo"></a>

### ClusterManagerInfo
ClusterManagerInfo contains cluster manager configuration


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| local_cluster_name | [string](#string) |  |  |
| outlier_detection | [bool](#bool) |  |  |
| upstream_bind_config | [bool](#bool) |  |  |
| load_stats_config | [bool](#bool) |  |  |
| connect_timeout | [string](#string) |  |  |
| per_connection_buffer_limit_bytes | [uint32](#uint32) |  |  |






<a name="navigator-types-v1alpha1-ClusterSummary"></a>

### ClusterSummary
ClusterSummary contains essential cluster configuration information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| type | [string](#string) |  |  |
| connect_timeout | [string](#string) |  |  |
| load_balancing_policy | [string](#string) |  |  |
| alt_stat_name | [string](#string) |  |  |
| direction | [ClusterDirection](#navigator-types-v1alpha1-ClusterDirection) |  |  |
| port | [uint32](#uint32) |  |  |
| subset | [string](#string) |  |  |
| service_fqdn | [string](#string) |  |  |
| raw_config | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-ConfigSourceInfo"></a>

### ConfigSourceInfo
ConfigSourceInfo contains information about a configuration source


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| config_source_specifier | [string](#string) |  |  |
| transport_api_version | [string](#string) |  |  |
| api_type | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-DynamicConfigInfo"></a>

### DynamicConfigInfo
DynamicConfigInfo contains information about dynamic resource configuration


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ads_config | [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo) |  |  |
| lds_config | [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo) |  |  |
| cds_config | [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo) |  |  |
| eds_config | [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo) |  |  |
| rds_config | [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo) |  |  |
| sds_config | [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo) |  |  |
| initial_fetch_timeout | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-EndpointInfo"></a>

### EndpointInfo
EndpointInfo contains individual endpoint information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| address | [string](#string) |  |  |
| port | [uint32](#uint32) |  |  |
| health | [string](#string) |  |  |
| weight | [uint32](#uint32) |  |  |
| priority | [uint32](#uint32) |  |  |
| host_identifier | [string](#string) |  |  |
| metadata | [EndpointInfo.MetadataEntry](#navigator-types-v1alpha1-EndpointInfo-MetadataEntry) | repeated |  |
| address_type | [AddressType](#navigator-types-v1alpha1-AddressType) |  |  |
| locality | [LocalityInfo](#navigator-types-v1alpha1-LocalityInfo) |  |  |






<a name="navigator-types-v1alpha1-EndpointInfo-MetadataEntry"></a>

### EndpointInfo.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-EndpointSummary"></a>

### EndpointSummary
EndpointSummary contains endpoint configuration information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_name | [string](#string) |  |  |
| endpoints | [EndpointInfo](#navigator-types-v1alpha1-EndpointInfo) | repeated |  |
| cluster_type | [ClusterType](#navigator-types-v1alpha1-ClusterType) |  |  |
| direction | [ClusterDirection](#navigator-types-v1alpha1-ClusterDirection) |  |  |
| port | [uint32](#uint32) |  |  |
| subset | [string](#string) |  |  |
| service_fqdn | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-ListenerSummary"></a>

### ListenerSummary
ListenerSummary contains essential listener configuration information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| address | [string](#string) |  |  |
| port | [uint32](#uint32) |  |  |
| type | [ListenerType](#navigator-types-v1alpha1-ListenerType) |  |  |
| use_original_dst | [bool](#bool) |  |  |
| raw_config | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-LocalityInfo"></a>

### LocalityInfo
LocalityInfo contains locality information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| region | [string](#string) |  |  |
| zone | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-NodeSummary"></a>

### NodeSummary
NodeSummary contains information about the Envoy node


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| cluster | [string](#string) |  |  |
| metadata | [NodeSummary.MetadataEntry](#navigator-types-v1alpha1-NodeSummary-MetadataEntry) | repeated |  |
| locality | [LocalityInfo](#navigator-types-v1alpha1-LocalityInfo) |  |  |
| proxy_mode | [ProxyMode](#navigator-types-v1alpha1-ProxyMode) |  |  |






<a name="navigator-types-v1alpha1-NodeSummary-MetadataEntry"></a>

### NodeSummary.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-ProxyConfig"></a>

### ProxyConfig
ProxyConfig represents the configuration of a proxy sidecar (e.g., Envoy).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [string](#string) |  | version is the version of the proxy software. |
| raw_config_dump | [string](#string) |  | raw_config_dump is the original raw configuration dump for debugging. |
| bootstrap | [BootstrapSummary](#navigator-types-v1alpha1-BootstrapSummary) |  | bootstrap contains the bootstrap configuration summary. |
| listeners | [ListenerSummary](#navigator-types-v1alpha1-ListenerSummary) | repeated | listeners contains the listener configuration summaries. |
| clusters | [ClusterSummary](#navigator-types-v1alpha1-ClusterSummary) | repeated | clusters contains the cluster configuration summaries. |
| endpoints | [EndpointSummary](#navigator-types-v1alpha1-EndpointSummary) | repeated | endpoints contains the endpoint configuration summaries. |
| routes | [RouteConfigSummary](#navigator-types-v1alpha1-RouteConfigSummary) | repeated | routes contains the route configuration summaries. |
| raw_clusters | [string](#string) |  | raw_clusters is the original raw clusters output from /clusters?format=json endpoint. |






<a name="navigator-types-v1alpha1-RouteActionInfo"></a>

### RouteActionInfo
RouteActionInfo contains route action information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action_type | [string](#string) |  |  |
| cluster | [string](#string) |  |  |
| weighted_clusters | [WeightedClusterInfo](#navigator-types-v1alpha1-WeightedClusterInfo) | repeated |  |
| timeout | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-RouteConfigSummary"></a>

### RouteConfigSummary
RouteConfigSummary contains route configuration summary


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| virtual_hosts | [VirtualHostInfo](#navigator-types-v1alpha1-VirtualHostInfo) | repeated |  |
| internal_only_headers | [string](#string) | repeated |  |
| validate_clusters | [bool](#bool) |  |  |
| raw_config | [string](#string) |  |  |
| type | [RouteType](#navigator-types-v1alpha1-RouteType) |  |  |






<a name="navigator-types-v1alpha1-RouteInfo"></a>

### RouteInfo
RouteInfo contains route information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| match | [RouteMatchInfo](#navigator-types-v1alpha1-RouteMatchInfo) |  |  |
| action | [RouteActionInfo](#navigator-types-v1alpha1-RouteActionInfo) |  |  |






<a name="navigator-types-v1alpha1-RouteMatchInfo"></a>

### RouteMatchInfo
RouteMatchInfo contains route matching information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path_specifier | [string](#string) |  |  |
| path | [string](#string) |  |  |
| case_sensitive | [bool](#bool) |  |  |






<a name="navigator-types-v1alpha1-VirtualHostInfo"></a>

### VirtualHostInfo
VirtualHostInfo contains virtual host information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| domains | [string](#string) | repeated |  |
| routes | [RouteInfo](#navigator-types-v1alpha1-RouteInfo) | repeated |  |






<a name="navigator-types-v1alpha1-WeightedClusterInfo"></a>

### WeightedClusterInfo
WeightedClusterInfo contains weighted cluster information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| weight | [uint32](#uint32) |  |  |
| metadata_match | [WeightedClusterInfo.MetadataMatchEntry](#navigator-types-v1alpha1-WeightedClusterInfo-MetadataMatchEntry) | repeated |  |






<a name="navigator-types-v1alpha1-WeightedClusterInfo-MetadataMatchEntry"></a>

### WeightedClusterInfo.MetadataMatchEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="navigator-types-v1alpha1-AddressType"></a>

### AddressType
AddressType represents the type of endpoint address

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_ADDRESS_TYPE | 0 | UNKNOWN_ADDRESS_TYPE indicates an unknown or unspecified address type |
| SOCKET_ADDRESS | 1 | SOCKET_ADDRESS indicates a standard network socket address (IP:port) |
| ENVOY_INTERNAL_ADDRESS | 2 | ENVOY_INTERNAL_ADDRESS indicates an internal Envoy address for listener routing |
| PIPE_ADDRESS | 3 | PIPE_ADDRESS indicates a Unix domain socket address |



<a name="navigator-types-v1alpha1-ClusterDirection"></a>

### ClusterDirection
ClusterDirection represents the traffic direction for a cluster

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNSPECIFIED | 0 | UNSPECIFIED indicates the direction is not specified or unknown |
| INBOUND | 1 | INBOUND indicates traffic flowing into the service |
| OUTBOUND | 2 | OUTBOUND indicates traffic flowing out of the service |



<a name="navigator-types-v1alpha1-ClusterType"></a>

### ClusterType
ClusterType represents the discovery type of a cluster

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_CLUSTER_TYPE | 0 | UNKNOWN_CLUSTER_TYPE indicates an unknown or unspecified cluster type |
| CLUSTER_EDS | 1 | CLUSTER_EDS indicates Endpoint Discovery Service clusters (dynamic service discovery) |
| CLUSTER_STATIC | 2 | CLUSTER_STATIC indicates static clusters with predefined endpoints |
| CLUSTER_STRICT_DNS | 3 | CLUSTER_STRICT_DNS indicates clusters using strict DNS resolution |
| CLUSTER_LOGICAL_DNS | 4 | CLUSTER_LOGICAL_DNS indicates clusters using logical DNS resolution |
| CLUSTER_ORIGINAL_DST | 5 | CLUSTER_ORIGINAL_DST indicates clusters using original destination routing |



<a name="navigator-types-v1alpha1-ListenerType"></a>

### ListenerType
ListenerType indicates the type/direction of a listener

| Name | Number | Description |
| ---- | ------ | ----------- |
| VIRTUAL_INBOUND | 0 | VIRTUAL_INBOUND listeners are virtual inbound listeners (typically 0.0.0.0 without use_original_dst) |
| VIRTUAL_OUTBOUND | 1 | VIRTUAL_OUTBOUND listeners are virtual outbound listeners (typically 0.0.0.0 with use_original_dst) |
| SERVICE_OUTBOUND | 2 | SERVICE_OUTBOUND listeners for specific upstream services (service.namespace.svc.cluster.local:port) |
| PORT_OUTBOUND | 3 | PORT_OUTBOUND listeners for generic port traffic outbound (e.g., &#34;80&#34;, &#34;443&#34;) |
| PROXY_METRICS | 4 | PROXY_METRICS listeners serve Prometheus metrics (typically on port 15090) |
| PROXY_HEALTHCHECK | 5 | PROXY_HEALTHCHECK listeners serve health check endpoints (typically on port 15021) |
| ADMIN_XDS | 6 | ADMIN_XDS listeners serve Envoy xDS configuration (typically on port 15010) |
| ADMIN_WEBHOOK | 7 | ADMIN_WEBHOOK listeners serve Istio webhook endpoints (typically on port 15012) |
| ADMIN_DEBUG | 8 | ADMIN_DEBUG listeners serve Envoy debug/admin interface (typically on port 15014) |



<a name="navigator-types-v1alpha1-ProxyMode"></a>

### ProxyMode
ProxyMode indicates the type of proxy (extracted from node ID)

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_PROXY_MODE | 0 | UNKNOWN_PROXY_MODE indicates an unknown or unspecified proxy mode |
| SIDECAR | 1 | SIDECAR indicates a sidecar proxy (most common in Istio) |
| GATEWAY | 2 | GATEWAY indicates a gateway proxy (ingress/egress gateways) |
| ROUTER | 3 | ROUTER indicates a router proxy |



<a name="navigator-types-v1alpha1-RouteType"></a>

### RouteType
RouteType indicates the type/category of a route configuration

| Name | Number | Description |
| ---- | ------ | ----------- |
| PORT_BASED | 0 | PORT_BASED routes are routes with just port numbers (e.g., &#34;80&#34;, &#34;443&#34;, &#34;15010&#34;) |
| SERVICE_SPECIFIC | 1 | SERVICE_SPECIFIC routes are routes with service hostnames and ports (e.g., &#34;backend.demo.svc.cluster.local:8080&#34;, external domains from ServiceEntries) |
| STATIC | 2 | STATIC routes are Istio/Envoy internal routing patterns (e.g., &#34;InboundPassthroughCluster&#34;, &#34;inbound|8080||&#34;) |


 

 

 



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

