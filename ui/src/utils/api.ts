import axios from 'axios';
import type {
    Service,
    ServiceListResponse,
    ServiceInstanceDetail,
    GetProxyConfigResponse,
} from '../types/service';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

export const serviceApi = {
    listServices: async (): Promise<Service[]> => {
        const response = await api.get<ServiceListResponse>(
            '/api/v1alpha1/services'
        );
        return response.data.services || [];
    },

    getService: async (id: string): Promise<Service> => {
        const response = await api.get<{ service: Service }>(
            `/api/v1alpha1/services/${id}`
        );
        return response.data.service;
    },

    getServiceInstance: async (
        serviceId: string,
        instanceId: string
    ): Promise<ServiceInstanceDetail> => {
        const response = await api.get<{ instance: ServiceInstanceDetail }>(
            `/api/v1alpha1/services/${encodeURIComponent(serviceId)}/instances/${encodeURIComponent(instanceId)}`
        );
        return response.data.instance;
    },

    getProxyConfig: async (
        serviceId: string,
        instanceId: string
    ): Promise<GetProxyConfigResponse> => {
        const response = await api.get<GetProxyConfigResponse>(
            `/api/v1alpha1/troubleshooting/services/${serviceId}/instances/${instanceId}/proxy-config`
        );
        return response.data;
    },
};

export default api;
