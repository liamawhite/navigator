/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { rpcStatus } from '../models/rpcStatus';
import type { v1alpha1GetServiceInstanceResponse } from '../models/v1alpha1GetServiceInstanceResponse';
import type { v1alpha1GetServiceResponse } from '../models/v1alpha1GetServiceResponse';
import type { v1alpha1ListServicesResponse } from '../models/v1alpha1ListServicesResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ServiceRegistryServiceService {
    /**
     * ListServices returns all services in the specified namespace, or all namespaces if not specified.
     * @param namespace namespace is the Kubernetes namespace to list services from.
     * If not specified, services from all namespaces are returned.
     * @returns v1alpha1ListServicesResponse A successful response.
     * @returns rpcStatus An unexpected error response.
     * @throws ApiError
     */
    public static serviceRegistryServiceListServices(
        namespace?: string,
    ): CancelablePromise<v1alpha1ListServicesResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/services',
            query: {
                'namespace': namespace,
            },
        });
    }
    /**
     * GetService returns detailed information about a specific service.
     * @param id id is the unique identifier of the service to retrieve.
     * Format: namespace:service-name (e.g., "default:nginx-service")
     * @returns v1alpha1GetServiceResponse A successful response.
     * @returns rpcStatus An unexpected error response.
     * @throws ApiError
     */
    public static serviceRegistryServiceGetService(
        id: string,
    ): CancelablePromise<v1alpha1GetServiceResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/services/{id}',
            path: {
                'id': id,
            },
        });
    }
    /**
     * GetServiceInstance returns detailed information about a specific service instance.
     * @param serviceId service_id is the unique identifier of the service.
     * Format: namespace:service-name (e.g., "default:nginx-service")
     * @param instanceId instance_id is the unique identifier of the specific service instance.
     * Format: cluster_name:namespace:pod_name (e.g., "prod-west:default:nginx-abc123")
     * @returns v1alpha1GetServiceInstanceResponse A successful response.
     * @returns rpcStatus An unexpected error response.
     * @throws ApiError
     */
    public static serviceRegistryServiceGetServiceInstance(
        serviceId: string,
        instanceId: string,
    ): CancelablePromise<v1alpha1GetServiceInstanceResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/services/{serviceId}/instances/{instanceId}',
            path: {
                'serviceId': serviceId,
                'instanceId': instanceId,
            },
        });
    }
}
