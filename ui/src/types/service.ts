export interface Service {
    id: string;
    name: string;
    namespace: string;
    instances: ServiceInstance[];
}

export interface ServiceInstance {
    ip: string;
    pod: string;
    namespace: string;
    hasProxySidecar: boolean;
}

export interface ServiceListResponse {
    services: Service[];
}
