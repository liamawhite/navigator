import axios from 'axios';
import type { v1alpha1Service } from '../types/generated/openapi';
import type { v1alpha1ListServicesResponse } from '../types/generated/openapi';
import type { v1alpha1ServiceInstanceDetail } from '../types/generated/openapi';
import type { v1alpha1GetProxyConfigResponse } from '../types/generated/openapi-troubleshooting';

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
        }>(
            `/api/v1alpha1/services/${encodeURIComponent(serviceId)}/instances/${encodeURIComponent(instanceId)}`
        );
        return response.data.instance;
    },

    getProxyConfig: async (
        serviceId: string,
        instanceId: string
    ): Promise<v1alpha1GetProxyConfigResponse> => {
        const response = await api.get<v1alpha1GetProxyConfigResponse>(
            `/api/v1alpha1/troubleshooting/services/${serviceId}/instances/${instanceId}/proxy-config`
        );
        return response.data;
    },
};

export default api;
