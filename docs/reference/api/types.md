# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [types/v1alpha1/istio_resources.proto](#types_v1alpha1_istio_resources-proto)
    - [AuthorizationPolicy](#navigator-types-v1alpha1-AuthorizationPolicy)
    - [DestinationRule](#navigator-types-v1alpha1-DestinationRule)
    - [DestinationRuleSubset](#navigator-types-v1alpha1-DestinationRuleSubset)
    - [DestinationRuleSubset.LabelsEntry](#navigator-types-v1alpha1-DestinationRuleSubset-LabelsEntry)
    - [EnvoyFilter](#navigator-types-v1alpha1-EnvoyFilter)
    - [Gateway](#navigator-types-v1alpha1-Gateway)
    - [Gateway.SelectorEntry](#navigator-types-v1alpha1-Gateway-SelectorEntry)
    - [IstioControlPlaneConfig](#navigator-types-v1alpha1-IstioControlPlaneConfig)
    - [PeerAuthentication](#navigator-types-v1alpha1-PeerAuthentication)
    - [PolicyTargetReference](#navigator-types-v1alpha1-PolicyTargetReference)
    - [RequestAuthentication](#navigator-types-v1alpha1-RequestAuthentication)
    - [ServiceEntry](#navigator-types-v1alpha1-ServiceEntry)
    - [Sidecar](#navigator-types-v1alpha1-Sidecar)
    - [VirtualService](#navigator-types-v1alpha1-VirtualService)
    - [WasmPlugin](#navigator-types-v1alpha1-WasmPlugin)
    - [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector)
    - [WorkloadSelector.MatchLabelsEntry](#navigator-types-v1alpha1-WorkloadSelector-MatchLabelsEntry)
  
- [types/v1alpha1/kubernetes_types.proto](#types_v1alpha1_kubernetes_types-proto)
    - [ServiceType](#navigator-types-v1alpha1-ServiceType)
  
- [types/v1alpha1/metrics_types.proto](#types_v1alpha1_metrics_types-proto)
    - [GraphMetricsFilters](#navigator-types-v1alpha1-GraphMetricsFilters)
    - [HistogramBucket](#navigator-types-v1alpha1-HistogramBucket)
    - [LatencyDistribution](#navigator-types-v1alpha1-LatencyDistribution)
    - [ServiceGraphMetrics](#navigator-types-v1alpha1-ServiceGraphMetrics)
    - [ServicePairMetrics](#navigator-types-v1alpha1-ServicePairMetrics)
  
- [types/v1alpha1/proxy_types.proto](#types_v1alpha1_proxy_types-proto)
    - [BootstrapSummary](#navigator-types-v1alpha1-BootstrapSummary)
    - [ClusterManagerInfo](#navigator-types-v1alpha1-ClusterManagerInfo)
    - [ClusterSummary](#navigator-types-v1alpha1-ClusterSummary)
    - [ConfigSourceInfo](#navigator-types-v1alpha1-ConfigSourceInfo)
    - [DynamicConfigInfo](#navigator-types-v1alpha1-DynamicConfigInfo)
    - [EndpointInfo](#navigator-types-v1alpha1-EndpointInfo)
    - [EndpointInfo.MetadataEntry](#navigator-types-v1alpha1-EndpointInfo-MetadataEntry)
    - [EndpointSummary](#navigator-types-v1alpha1-EndpointSummary)
    - [FilterChainMatch](#navigator-types-v1alpha1-FilterChainMatch)
    - [FilterChainSummary](#navigator-types-v1alpha1-FilterChainSummary)
    - [FilterInfo](#navigator-types-v1alpha1-FilterInfo)
    - [HeaderMatchInfo](#navigator-types-v1alpha1-HeaderMatchInfo)
    - [HttpRouteMatch](#navigator-types-v1alpha1-HttpRouteMatch)
    - [ListenerDestination](#navigator-types-v1alpha1-ListenerDestination)
    - [ListenerMatch](#navigator-types-v1alpha1-ListenerMatch)
    - [ListenerRule](#navigator-types-v1alpha1-ListenerRule)
    - [ListenerSummary](#navigator-types-v1alpha1-ListenerSummary)
    - [LocalityInfo](#navigator-types-v1alpha1-LocalityInfo)
    - [NodeSummary](#navigator-types-v1alpha1-NodeSummary)
    - [NodeSummary.MetadataEntry](#navigator-types-v1alpha1-NodeSummary-MetadataEntry)
    - [PathMatchInfo](#navigator-types-v1alpha1-PathMatchInfo)
    - [ProxyConfig](#navigator-types-v1alpha1-ProxyConfig)
    - [RouteActionInfo](#navigator-types-v1alpha1-RouteActionInfo)
    - [RouteConfigSummary](#navigator-types-v1alpha1-RouteConfigSummary)
    - [RouteInfo](#navigator-types-v1alpha1-RouteInfo)
    - [RouteMatchInfo](#navigator-types-v1alpha1-RouteMatchInfo)
    - [TcpProxyMatch](#navigator-types-v1alpha1-TcpProxyMatch)
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



<a name="types_v1alpha1_istio_resources-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## types/v1alpha1/istio_resources.proto



<a name="navigator-types-v1alpha1-AuthorizationPolicy"></a>

### AuthorizationPolicy
AuthorizationPolicy represents an Istio AuthorizationPolicy resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the authorization policy. |
| namespace | [string](#string) |  | namespace is the namespace of the authorization policy. |
| raw_config | [string](#string) |  | raw_config is the complete authorization policy resource as a JSON string. |
| selector | [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector) |  | selector is the criteria used to select the specific set of pods/VMs. |
| target_refs | [PolicyTargetReference](#navigator-types-v1alpha1-PolicyTargetReference) | repeated | target_refs is the list of resources that this authorization policy applies to. |






<a name="navigator-types-v1alpha1-DestinationRule"></a>

### DestinationRule
DestinationRule represents an Istio DestinationRule resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the destination rule. |
| namespace | [string](#string) |  | namespace is the namespace of the destination rule. |
| raw_config | [string](#string) |  | raw_config is the complete destination rule resource as a JSON string. |
| host | [string](#string) |  | host is the name of a service from the service registry. |
| subsets | [DestinationRuleSubset](#navigator-types-v1alpha1-DestinationRuleSubset) | repeated | subsets is the list of named subsets for traffic routing. |
| export_to | [string](#string) | repeated | export_to controls the visibility of this destination rule to other namespaces. |
| workload_selector | [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector) |  | workload_selector is the criteria used to select the specific set of pods/VMs. |






<a name="navigator-types-v1alpha1-DestinationRuleSubset"></a>

### DestinationRuleSubset
DestinationRuleSubset represents a named subset for destination rule traffic routing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the subset. |
| labels | [DestinationRuleSubset.LabelsEntry](#navigator-types-v1alpha1-DestinationRuleSubset-LabelsEntry) | repeated | labels are the key-value pairs that define the subset. |






<a name="navigator-types-v1alpha1-DestinationRuleSubset-LabelsEntry"></a>

### DestinationRuleSubset.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-EnvoyFilter"></a>

### EnvoyFilter
EnvoyFilter represents an Istio EnvoyFilter resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the envoy filter. |
| namespace | [string](#string) |  | namespace is the namespace of the envoy filter. |
| raw_config | [string](#string) |  | raw_config is the complete envoy filter resource as a JSON string. |
| workload_selector | [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector) |  | workload_selector is the criteria used to select the specific set of pods/VMs. |
| target_refs | [PolicyTargetReference](#navigator-types-v1alpha1-PolicyTargetReference) | repeated | target_refs is the list of resources that this envoy filter applies to. |






<a name="navigator-types-v1alpha1-Gateway"></a>

### Gateway
Gateway represents an Istio Gateway resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the gateway. |
| namespace | [string](#string) |  | namespace is the namespace of the gateway. |
| raw_config | [string](#string) |  | raw_config is the complete gateway resource as a JSON string. |
| selector | [Gateway.SelectorEntry](#navigator-types-v1alpha1-Gateway-SelectorEntry) | repeated | selector is the workload selector for the gateway. |






<a name="navigator-types-v1alpha1-Gateway-SelectorEntry"></a>

### Gateway.SelectorEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="navigator-types-v1alpha1-IstioControlPlaneConfig"></a>

### IstioControlPlaneConfig
IstioControlPlaneConfig represents configuration from the Istio control plane.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pilot_scope_gateway_to_namespace | [bool](#bool) |  | pilot_scope_gateway_to_namespace indicates whether gateway selector scope is restricted to namespace. When true, gateway selectors only match workloads in the same namespace as the gateway. When false (default), gateway selectors match workloads across all namespaces. |
| root_namespace | [string](#string) |  | root_namespace is the namespace where the Istio control plane is installed. This is typically &#34;istio-system&#34; but can be customized in multi-cluster or external control plane deployments. Resources in the root namespace have special behavior (e.g., EnvoyFilters apply globally). |






<a name="navigator-types-v1alpha1-PeerAuthentication"></a>

### PeerAuthentication
PeerAuthentication represents an Istio PeerAuthentication resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the peer authentication. |
| namespace | [string](#string) |  | namespace is the namespace of the peer authentication. |
| raw_config | [string](#string) |  | raw_config is the complete peer authentication resource as a JSON string. |
| selector | [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector) |  | selector is the criteria used to select the specific set of pods/VMs. |






<a name="navigator-types-v1alpha1-PolicyTargetReference"></a>

### PolicyTargetReference
PolicyTargetReference represents a reference to a specific resource based on Istio&#39;s PolicyTargetReference.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) |  | group specifies the group of the target resource. |
| kind | [string](#string) |  | kind indicates the kind of target resource (required). |
| name | [string](#string) |  | name provides the name of the target resource (required). |
| namespace | [string](#string) |  | namespace defines the namespace of the referenced resource. When unspecified, the local namespace is inferred. |






<a name="navigator-types-v1alpha1-RequestAuthentication"></a>

### RequestAuthentication
RequestAuthentication represents an Istio RequestAuthentication resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the request authentication. |
| namespace | [string](#string) |  | namespace is the namespace of the request authentication. |
| raw_config | [string](#string) |  | raw_config is the complete request authentication resource as a JSON string. |
| selector | [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector) |  | selector is the criteria used to select the specific set of pods/VMs. |
| target_refs | [PolicyTargetReference](#navigator-types-v1alpha1-PolicyTargetReference) | repeated | target_refs is the list of resources that this request authentication applies to. |






<a name="navigator-types-v1alpha1-ServiceEntry"></a>

### ServiceEntry
ServiceEntry represents an Istio ServiceEntry resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the service entry. |
| namespace | [string](#string) |  | namespace is the namespace of the service entry. |
| raw_config | [string](#string) |  | raw_config is the complete service entry resource as a JSON string. |
| export_to | [string](#string) | repeated | export_to controls the visibility of this service entry to other namespaces. |






<a name="navigator-types-v1alpha1-Sidecar"></a>

### Sidecar
Sidecar represents an Istio Sidecar resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the sidecar. |
| namespace | [string](#string) |  | namespace is the namespace of the sidecar. |
| raw_config | [string](#string) |  | raw_config is the complete sidecar resource as a JSON string. |
| workload_selector | [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector) |  | workload_selector is the criteria used to select the specific set of pods/VMs. |






<a name="navigator-types-v1alpha1-VirtualService"></a>

### VirtualService
VirtualService represents an Istio VirtualService resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the virtual service. |
| namespace | [string](#string) |  | namespace is the namespace of the virtual service. |
| raw_config | [string](#string) |  | raw_config is the complete virtual service resource as a JSON string. |
| hosts | [string](#string) | repeated | hosts is the list of destination hosts that these routing rules apply to. |
| gateways | [string](#string) | repeated | gateways is the list of gateway names that should apply these routes. |
| export_to | [string](#string) | repeated | export_to controls the visibility of this virtual service to other namespaces. |






<a name="navigator-types-v1alpha1-WasmPlugin"></a>

### WasmPlugin
WasmPlugin represents an Istio WasmPlugin resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the name of the wasm plugin. |
| namespace | [string](#string) |  | namespace is the namespace of the wasm plugin. |
| raw_config | [string](#string) |  | raw_config is the complete wasm plugin resource as a JSON string. |
| selector | [WorkloadSelector](#navigator-types-v1alpha1-WorkloadSelector) |  | selector is the criteria used to select the specific set of pods/VMs. |
| target_refs | [PolicyTargetReference](#navigator-types-v1alpha1-PolicyTargetReference) | repeated | target_refs is the list of resources that this wasm plugin applies to. |






<a name="navigator-types-v1alpha1-WorkloadSelector"></a>

### WorkloadSelector
WorkloadSelector represents the workload selector criteria used across Istio resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match_labels | [WorkloadSelector.MatchLabelsEntry](#navigator-types-v1alpha1-WorkloadSelector-MatchLabelsEntry) | repeated | match_labels are the labels used to select pods/VMs. |






<a name="navigator-types-v1alpha1-WorkloadSelector-MatchLabelsEntry"></a>

### WorkloadSelector.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 

 



<a name="types_v1alpha1_kubernetes_types-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## types/v1alpha1/kubernetes_types.proto


 


<a name="navigator-types-v1alpha1-ServiceType"></a>

### ServiceType
ServiceType indicates the type of Kubernetes service.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SERVICE_TYPE_UNSPECIFIED | 0 | SERVICE_TYPE_UNSPECIFIED indicates the service type is not specified or unknown. |
| CLUSTER_IP | 1 | CLUSTER_IP exposes the service on a cluster-internal IP. |
| NODE_PORT | 2 | NODE_PORT exposes the service on each node&#39;s IP at a static port. |
| LOAD_BALANCER | 3 | LOAD_BALANCER exposes the service externally using a cloud provider&#39;s load balancer. |
| EXTERNAL_NAME | 4 | EXTERNAL_NAME maps the service to the contents of the externalName field. |


 

 

 



<a name="types_v1alpha1_metrics_types-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## types/v1alpha1/metrics_types.proto



<a name="navigator-types-v1alpha1-GraphMetricsFilters"></a>

### GraphMetricsFilters
GraphMetricsFilters specify filters for service graph metrics queries.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | [string](#string) | repeated | namespaces filters metrics to only include these namespaces. |
| clusters | [string](#string) | repeated | clusters filters metrics to only include these clusters. |






<a name="navigator-types-v1alpha1-HistogramBucket"></a>

### HistogramBucket
HistogramBucket represents a single bucket in a histogram distribution.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| le | [double](#double) |  | le is the upper bound of the bucket (less-than-or-equal-to). |
| count | [double](#double) |  | count is the cumulative count of observations in this bucket. |






<a name="navigator-types-v1alpha1-LatencyDistribution"></a>

### LatencyDistribution
LatencyDistribution represents a histogram distribution of latency measurements.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| buckets | [HistogramBucket](#navigator-types-v1alpha1-HistogramBucket) | repeated | buckets contains the histogram buckets sorted by upper bound. |
| total_count | [double](#double) |  | total_count is the total number of observations across all buckets. |
| sum | [double](#double) |  | sum is the sum of all observed values. |






<a name="navigator-types-v1alpha1-ServiceGraphMetrics"></a>

### ServiceGraphMetrics
ServiceGraphMetrics contains service-to-service metrics for a cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pairs | [ServicePairMetrics](#navigator-types-v1alpha1-ServicePairMetrics) | repeated | pairs contains the service-to-service metrics. |
| cluster_id | [string](#string) |  | cluster_id is the ID of the cluster these metrics came from. |
| timestamp | [string](#string) |  | timestamp is when these metrics were collected (RFC3339 format). |






<a name="navigator-types-v1alpha1-ServicePairMetrics"></a>

### ServicePairMetrics
ServicePairMetrics represents metrics between a source and destination service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| source_cluster | [string](#string) |  | source_cluster is the cluster name of the source service. |
| source_namespace | [string](#string) |  | source_namespace is the namespace of the source service. |
| source_service | [string](#string) |  | source_service is the service name of the source service. |
| destination_cluster | [string](#string) |  | destination_cluster is the cluster name of the destination service. |
| destination_namespace | [string](#string) |  | destination_namespace is the namespace of the destination service. |
| destination_service | [string](#string) |  | destination_service is the service name of the destination service. |
| error_rate | [double](#double) |  | error_rate is the error rate in requests per second. |
| request_rate | [double](#double) |  | request_rate is the request rate in requests per second. |
| latency_p99 | [google.protobuf.Duration](#google-protobuf-Duration) |  | latency_p99 is the 99th percentile latency. |
| latency_distribution | [LatencyDistribution](#navigator-types-v1alpha1-LatencyDistribution) |  | latency_distribution contains the raw histogram distribution for latency. This enables aggregation and percentile calculation at different levels. |





 

 

 

 



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






<a name="navigator-types-v1alpha1-FilterChainMatch"></a>

### FilterChainMatch
FilterChainMatch represents filter chain matching criteria (TLS/SNI/ALPN)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| server_names | [string](#string) | repeated | server_names contains SNI/TLS server name matching patterns |
| application_protocols | [string](#string) | repeated | application_protocols contains ALPN application protocol matches |
| transport_protocol | [string](#string) |  | transport_protocol contains the transport protocol (raw_buffer, tls, etc.) |






<a name="navigator-types-v1alpha1-FilterChainSummary"></a>

### FilterChainSummary
FilterChainSummary contains filter chain analysis


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| total_chains | [uint32](#uint32) |  | total_chains is the number of filter chains |
| http_filters | [FilterInfo](#navigator-types-v1alpha1-FilterInfo) | repeated | http_filters contains HTTP filter information |
| network_filters | [FilterInfo](#navigator-types-v1alpha1-FilterInfo) | repeated | network_filters contains network filter information |
| tls_context | [bool](#bool) |  | tls_context indicates if TLS is configured |






<a name="navigator-types-v1alpha1-FilterInfo"></a>

### FilterInfo
FilterInfo contains filter information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the filter name |
| type | [string](#string) |  | type is the filter type |
| config_summary | [string](#string) |  | config_summary is a summary of the filter configuration |






<a name="navigator-types-v1alpha1-HeaderMatchInfo"></a>

### HeaderMatchInfo
HeaderMatchInfo contains HTTP header matching information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name is the header name to match |
| match_type | [string](#string) |  | match_type indicates exact, prefix, regex, etc. |
| value | [string](#string) |  | value is the header value pattern to match |
| invert_match | [bool](#bool) |  | invert_match indicates if the match should be inverted |






<a name="navigator-types-v1alpha1-HttpRouteMatch"></a>

### HttpRouteMatch
HttpRouteMatch represents HTTP route matching criteria (from HTTP connection manager)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path_match | [PathMatchInfo](#navigator-types-v1alpha1-PathMatchInfo) |  | path_match contains HTTP path matching patterns |
| header_matches | [HeaderMatchInfo](#navigator-types-v1alpha1-HeaderMatchInfo) | repeated | header_matches contains HTTP header matching patterns |
| methods | [string](#string) | repeated | methods contains HTTP method matching patterns |






<a name="navigator-types-v1alpha1-ListenerDestination"></a>

### ListenerDestination
ListenerDestination contains listener destination information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destination_type | [string](#string) |  | destination_type indicates cluster, static IP, original_dst, etc. |
| cluster_name | [string](#string) |  | cluster_name is the destination cluster name |
| address | [string](#string) |  | address is the destination IP address (for static destinations) |
| port | [uint32](#uint32) |  | port is the destination port |
| weight | [uint32](#uint32) |  | weight is the traffic weight (for weighted destinations) |
| service_fqdn | [string](#string) |  | service_fqdn is the Istio service FQDN (enriched field) |






<a name="navigator-types-v1alpha1-ListenerMatch"></a>

### ListenerMatch
ListenerMatch contains listener matching criteria using discriminated union


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| http_route | [HttpRouteMatch](#navigator-types-v1alpha1-HttpRouteMatch) |  |  |
| filter_chain | [FilterChainMatch](#navigator-types-v1alpha1-FilterChainMatch) |  |  |
| tcp_proxy | [TcpProxyMatch](#navigator-types-v1alpha1-TcpProxyMatch) |  |  |






<a name="navigator-types-v1alpha1-ListenerRule"></a>

### ListenerRule
ListenerRule pairs a match condition with its corresponding destination


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match | [ListenerMatch](#navigator-types-v1alpha1-ListenerMatch) |  | match contains the matching criteria (HTTP route, filter chain, TCP proxy) |
| destination | [ListenerDestination](#navigator-types-v1alpha1-ListenerDestination) |  | destination contains the routing destination for this match |






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
| rules | [ListenerRule](#navigator-types-v1alpha1-ListenerRule) | repeated |  |
| filter_chains | [FilterChainSummary](#navigator-types-v1alpha1-FilterChainSummary) |  |  |






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






<a name="navigator-types-v1alpha1-PathMatchInfo"></a>

### PathMatchInfo
PathMatchInfo contains HTTP path matching information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| match_type | [string](#string) |  | match_type indicates exact, prefix, regex, etc. |
| path | [string](#string) |  | path is the path pattern to match |
| case_sensitive | [bool](#bool) |  | case_sensitive indicates if matching is case sensitive |






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






<a name="navigator-types-v1alpha1-TcpProxyMatch"></a>

### TcpProxyMatch
TcpProxyMatch represents TCP proxy matching criteria


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_name | [string](#string) |  | cluster_name is the destination cluster for TCP proxy |






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
| PIPE_ADDRESS | 2 | PIPE_ADDRESS indicates a Unix domain socket address |



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
| UNKNOWN_LISTENER_TYPE | 0 | UNKNOWN_LISTENER_TYPE indicates an unknown or unspecified listener type |
| VIRTUAL_INBOUND | 1 | VIRTUAL_INBOUND listeners are virtual inbound listeners (typically 0.0.0.0 without use_original_dst) |
| VIRTUAL_OUTBOUND | 2 | VIRTUAL_OUTBOUND listeners are virtual outbound listeners (typically 0.0.0.0 with use_original_dst) |
| SERVICE_OUTBOUND | 3 | SERVICE_OUTBOUND listeners for specific upstream services (service.namespace.svc.cluster.local:port) |
| PORT_OUTBOUND | 4 | PORT_OUTBOUND listeners for generic port traffic outbound (e.g., &#34;80&#34;, &#34;443&#34;) |
| PROXY_METRICS | 5 | PROXY_METRICS listeners serve Prometheus metrics (typically on port 15090) |
| PROXY_HEALTHCHECK | 6 | PROXY_HEALTHCHECK listeners serve health check endpoints (typically on port 15021) |
| ADMIN_XDS | 7 | ADMIN_XDS listeners serve Envoy xDS configuration (typically on port 15010) |
| ADMIN_WEBHOOK | 8 | ADMIN_WEBHOOK listeners serve Istio webhook endpoints (typically on port 15012) |
| ADMIN_DEBUG | 9 | ADMIN_DEBUG listeners serve Envoy debug/admin interface (typically on port 15014) |
| GATEWAY_INBOUND | 10 | GATEWAY_INBOUND listeners accept external traffic into gateway proxies (typically 0.0.0.0 without use_original_dst) |



<a name="navigator-types-v1alpha1-ProxyMode"></a>

### ProxyMode
ProxyMode indicates the type of proxy (extracted from node ID)

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN_PROXY_MODE | 0 | UNKNOWN_PROXY_MODE indicates an unknown or unspecified proxy mode |
| NONE | 1 | NONE indicates no proxy is present |
| SIDECAR | 2 | SIDECAR indicates a sidecar proxy (most common in Istio) |
| ROUTER | 3 | ROUTER indicates a router proxy (used for ingress/egress gateways) |



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

