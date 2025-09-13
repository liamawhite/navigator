/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { rpcStatus } from '../models/rpcStatus';
import type { v1alpha1GetServiceConnectionsResponse } from '../models/v1alpha1GetServiceConnectionsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class MetricsServiceService {
    /**
     * GetServiceConnections returns inbound and outbound connections for a specific service.
     * @param serviceName service_name is the name of the service to get connections for.
     * @param namespace namespace is the Kubernetes namespace of the service.
     * @param startTime start_time specifies the start time for the metrics query (required).
     * Must be in the past (before current time).
     * @param endTime end_time specifies the end time for the metrics query (required).
     * Must be in the past (before current time) and after start_time.
     * @returns v1alpha1GetServiceConnectionsResponse A successful response.
     * @returns rpcStatus An unexpected error response.
     * @throws ApiError
     */
    public static metricsServiceGetServiceConnections(
        serviceName: string,
        namespace?: string,
        startTime?: string,
        endTime?: string,
    ): CancelablePromise<v1alpha1GetServiceConnectionsResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/metrics/service/{serviceName}/connections',
            path: {
                'serviceName': serviceName,
            },
            query: {
                'namespace': namespace,
                'startTime': startTime,
                'endTime': endTime,
            },
        });
    }
}
