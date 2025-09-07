/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { rpcStatus } from '../models/rpcStatus';
import type { v1alpha1ListClustersResponse } from '../models/v1alpha1ListClustersResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ClusterRegistryServiceService {
    /**
     * ListClusters returns sync state information for all connected clusters.
     * @returns v1alpha1ListClustersResponse A successful response.
     * @returns rpcStatus An unexpected error response.
     * @throws ApiError
     */
    public static clusterRegistryServiceListClusters(): CancelablePromise<v1alpha1ListClustersResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/clusters',
        });
    }
}
