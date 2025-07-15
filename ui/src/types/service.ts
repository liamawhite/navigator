export interface Service {
    id: string;
    name: string;
    namespace: string;
    instances: ServiceInstance[];
}

export interface ServiceInstance {
    instanceId: string;
    ip: string;
    pod: string;
    namespace: string;
    clusterName: string;
    isEnvoyPresent: boolean;
}

export interface ContainerInfo {
    name: string;
    image: string;
    ready: boolean;
    restartCount: number;
    status: string;
}

export interface ServiceInstanceDetail {
    instanceId: string;
    ip: string;
    pod: string;
    namespace: string;
    clusterName: string;
    isEnvoyPresent: boolean;
    serviceName: string;
    podStatus: string;
    createdAt: string;
    labels: Record<string, string>;
    annotations: Record<string, string>;
    containers: ContainerInfo[];
    nodeName: string;
}

export interface ServiceListResponse {
    services: Service[];
}

export interface ProxyConfig {
    proxyType: string;
    version: string;
    adminPort: number;
    bootstrap?: any; // Envoy Bootstrap configuration
    listeners?: any[]; // Envoy Listener configurations
    clusters?: any[]; // Envoy Cluster configurations
    endpoints?: any[]; // Envoy Endpoint configurations
    routes?: any[]; // Envoy Route configurations
    configDump?: any; // Full Envoy ConfigDump structure
    rawConfigDump?: string; // Raw JSON configuration dump
}

export interface GetProxyConfigResponse {
    proxyConfig: ProxyConfig;
}
