// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
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
        instanceId: string,
    ): CancelablePromise<v1alpha1GetProxyConfigResponse | rpcStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/api/v1alpha1/troubleshooting/services/{serviceId}/instances/{instanceId}/proxy-config',
            path: {
                'serviceId': serviceId,
                'instanceId': instanceId,
            },
        });
    }
}
