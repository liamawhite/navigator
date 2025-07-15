/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */

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
     * instances are the backend instances (pods) that serve this service.
     */
    instances?: Array<v1alpha1ServiceInstance>;
};
