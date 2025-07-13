import axios from 'axios';
import type { Service, ServiceListResponse } from '../types/service';

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
};

export default api;
