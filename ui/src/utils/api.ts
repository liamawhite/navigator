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

import axios from 'axios';
import type { v1alpha1Service } from '../types/generated/openapi';
import type { v1alpha1ListServicesResponse } from '../types/generated/openapi';
import type {
    v1alpha1GetProxyConfigResponse,
    v1alpha1ServiceInstanceDetail,
    v1alpha1ListClustersResponse,
    v1alpha1ClusterSyncInfo,
    v1alpha1GetIstioResourcesResponse,
} from '../types/generated/openapi-service_registry';
import type {
    v1alpha1GetServiceGraphMetricsResponse,
    v1alpha1ServicePairMetrics,
} from '../types/generated/openapi-metrics_service';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

export const serviceApi = {
    listServices: async (): Promise<v1alpha1Service[]> => {
        const response = await api.get<v1alpha1ListServicesResponse>(
            '/api/v1alpha1/services'
        );
        return response.data.services || [];
    },

    getService: async (id: string): Promise<v1alpha1Service> => {
        const response = await api.get<{ service: v1alpha1Service }>(
            `/api/v1alpha1/services/${id}`
        );
        return response.data.service;
    },

    getServiceInstance: async (
        serviceId: string,
        instanceId: string
    ): Promise<v1alpha1ServiceInstanceDetail> => {
        const response = await api.get<{
            instance: v1alpha1ServiceInstanceDetail;
        }>(`/api/v1alpha1/services/${serviceId}/instances/${instanceId}`);
        return response.data.instance;
    },

    getProxyConfig: async (
        serviceId: string,
        instanceId: string
    ): Promise<v1alpha1GetProxyConfigResponse> => {
        const response = await api.get<v1alpha1GetProxyConfigResponse>(
            `/api/v1alpha1/services/${serviceId}/instances/${instanceId}/proxy-config`
        );
        return response.data;
    },

    listClusters: async (): Promise<v1alpha1ClusterSyncInfo[]> => {
        const response = await api.get<v1alpha1ListClustersResponse>(
            '/api/v1alpha1/clusters'
        );
        return response.data.clusters || [];
    },

    getIstioResources: async (
        serviceId: string,
        instanceId: string
    ): Promise<v1alpha1GetIstioResourcesResponse> => {
        const response = await api.get<v1alpha1GetIstioResourcesResponse>(
            `/api/v1alpha1/services/${serviceId}/instances/${instanceId}/istio-resources`
        );
        return response.data;
    },
};

export const metricsApi = {
    getServiceGraphMetrics: async (params?: {
        namespaces?: string[];
        clusters?: string[];
        startTime?: string;
        endTime?: string;
    }): Promise<v1alpha1ServicePairMetrics[]> => {
        const response = await api.get<v1alpha1GetServiceGraphMetricsResponse>(
            '/api/v1alpha1/metrics/graph',
            {
                params: {
                    namespaces: params?.namespaces,
                    clusters: params?.clusters,
                    startTime: params?.startTime,
                    endTime: params?.endTime,
                },
            }
        );
        return response.data.pairs || [];
    },
};

export default api;
