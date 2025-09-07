/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { rpcStatus } from '../models/rpcStatus';
import type { v1alpha1GetServiceGraphMetricsResponse } from '../models/v1alpha1GetServiceGraphMetricsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class MetricsServiceService {
    /**
     * GetServiceGraphMetrics returns service-to-service graph metrics across the mesh.
     * @param namespaces namespaces filters metrics to only include these namespaces.
     * @param clusters clusters filters metrics to only include these clusters.
     * @param startTime start_time specifies the start time for the metrics query (required).
     * Must be in the past (before current time).
     * @param endTime end_time specifies the end time for the metrics query (required).
     * Must be in the past (before current time) and after start_time.
     * @returns v1alpha1GetServiceGraphMetricsResponse A successful response.
     * @returns rpcStatus An unexpected error response.
     * @throws ApiError
     */
    public static metricsServiceGetServiceGraphMetrics(
        namespaces?: Array<string>,
        clusters?: Array<string>,
        startTime?: string,
        endTime?: string,
    ): CancelablePromise<v1alpha1GetServiceGraphMetricsResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/metrics/graph',
            query: {
                'namespaces': namespaces,
                'clusters': clusters,
                'startTime': startTime,
                'endTime': endTime,
            },
        });
    }
}
