/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { rpcStatus } from '../models/rpcStatus';
import type { v1alpha1GetProxyConfigResponse } from '../models/v1alpha1GetProxyConfigResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class TroubleshootingServiceService {
    /**
     * GetProxyConfig returns the proxy configuration for a specific service instance.
     * @param serviceId service_id is the unique identifier of the service.
     * Format: namespace:service-name (e.g., "default:nginx-service")
     * @param instanceId instance_id is the unique identifier of the specific service instance.
     * Format: cluster_name:namespace:pod_name (e.g., "prod-west:default:nginx-abc123")
     * @returns v1alpha1GetProxyConfigResponse A successful response.
     * @returns rpcStatus An unexpected error response.
     * @throws ApiError
     */
    public static troubleshootingServiceGetProxyConfig(
        serviceId: string,
        instanceId: string
    ): CancelablePromise<v1alpha1GetProxyConfigResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/troubleshooting/services/{serviceId}/instances/{instanceId}/proxy-config',
            path: {
                serviceId: serviceId,
                instanceId: instanceId,
            },
        });
    }
}
