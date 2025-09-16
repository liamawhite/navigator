/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ProxyMode } from './v1alpha1ProxyMode';
import type { v1alpha1ServiceInstance } from './v1alpha1ServiceInstance';
/**
 * Service represents a Kubernetes service with its backing instances.
 * Services in different clusters that share the same name and namespace are considered the same service.
 */
export type v1alpha1Service = {
    /**
     * id is a unique identifier for the service in format namespace:service-name (e.g., "default:nginx-service").
     */
    id?: string;
    /**
     * name is the service name.
     */
    name?: string;
    /**
     * namespace is the Kubernetes namespace containing the service.
     */
    namespace?: string;
    /**
     * instances are the backend instances (pods) that serve this service across all clusters.
     */
    instances?: Array<v1alpha1ServiceInstance>;
    /**
     * cluster_ips maps cluster names to their cluster IP addresses for this service.
     */
    clusterIps?: Record<string, string>;
    /**
     * external_ips maps cluster names to their external IP addresses for this service.
     */
    externalIps?: Record<string, string>;
    /**
     * proxy_mode indicates the Istio proxy mode for this service (determined from instances).
     * Services with instances that have ProxyMode_ROUTER are gateway services.
     */
    proxyMode?: v1alpha1ProxyMode;
};

