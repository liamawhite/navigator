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

// Use globals from Jest environment
import axios from 'axios';

// Mock axios before importing the API module
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

// Create a mock axios instance
const mockAxiosInstance = {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
} as any; // eslint-disable-line @typescript-eslint/no-explicit-any

// Set up the mock before importing the API
mockedAxios.create = jest.fn(() => mockAxiosInstance);

// Import the API module after setting up mocks
import { serviceApi } from './api';

describe('API utilities', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockAxiosInstance.get.mockReset();
        mockAxiosInstance.post.mockReset();
        mockAxiosInstance.put.mockReset();
        mockAxiosInstance.delete.mockReset();
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('listServices', () => {
        it('should fetch services successfully', async () => {
            const mockResponse = {
                data: {
                    services: [{ name: 'test-service', namespace: 'default' }],
                },
            };
            mockAxiosInstance.get.mockResolvedValue(mockResponse);

            const result = await serviceApi.listServices();

            expect(mockAxiosInstance.get).toHaveBeenCalledWith(
                '/api/v1alpha1/services'
            );
            expect(result).toEqual(mockResponse.data.services);
        });

        it('should handle fetch services error', async () => {
            const mockError = new Error('Network error');
            mockAxiosInstance.get.mockRejectedValue(mockError);

            await expect(serviceApi.listServices()).rejects.toThrow(
                'Network error'
            );
            expect(mockAxiosInstance.get).toHaveBeenCalledWith(
                '/api/v1alpha1/services'
            );
        });
    });

    describe('getService', () => {
        it('should fetch service successfully', async () => {
            const mockResponse = {
                data: {
                    service: {
                        name: 'test-service',
                        namespace: 'default',
                    },
                },
            };
            mockAxiosInstance.get.mockResolvedValue(mockResponse);

            const result = await serviceApi.getService('test-service-id');

            expect(mockAxiosInstance.get).toHaveBeenCalledWith(
                '/api/v1alpha1/services/test-service-id'
            );
            expect(result).toEqual(mockResponse.data.service);
        });

        it('should handle fetch service error', async () => {
            const mockError = new Error('Not found');
            mockAxiosInstance.get.mockRejectedValue(mockError);

            await expect(
                serviceApi.getService('test-service-id')
            ).rejects.toThrow('Not found');
        });
    });

    describe('getServiceInstance', () => {
        it('should fetch service instance successfully', async () => {
            const mockResponse = {
                data: {
                    instance: {
                        instanceId: 'test-instance',
                        envoyPresent: true,
                    },
                },
            };
            mockAxiosInstance.get.mockResolvedValue(mockResponse);

            const result = await serviceApi.getServiceInstance(
                'test-service-id',
                'test-instance-id'
            );

            expect(mockAxiosInstance.get).toHaveBeenCalledWith(
                '/api/v1alpha1/services/test-service-id/instances/test-instance-id'
            );
            expect(result).toEqual(mockResponse.data.instance);
        });

        it('should handle fetch service instance error', async () => {
            const mockError = new Error('Not found');
            mockAxiosInstance.get.mockRejectedValue(mockError);

            await expect(
                serviceApi.getServiceInstance(
                    'test-service-id',
                    'test-instance-id'
                )
            ).rejects.toThrow('Not found');
        });
    });

    describe('getProxyConfig', () => {
        it('should fetch proxy config successfully', async () => {
            const mockResponse = {
                data: {
                    config: { bootstrap: {}, clusters: [] },
                },
            };
            mockAxiosInstance.get.mockResolvedValue(mockResponse);

            const result = await serviceApi.getProxyConfig(
                'test-service-id',
                'test-instance-id'
            );

            expect(mockAxiosInstance.get).toHaveBeenCalledWith(
                '/api/v1alpha1/services/test-service-id/instances/test-instance-id/proxy-config'
            );
            expect(result).toEqual(mockResponse.data);
        });

        it('should handle fetch proxy config error', async () => {
            const mockError = new Error('Config not available');
            mockAxiosInstance.get.mockRejectedValue(mockError);

            await expect(
                serviceApi.getProxyConfig('test-service-id', 'test-instance-id')
            ).rejects.toThrow('Config not available');
        });
    });

    describe('listClusters', () => {
        it('should fetch clusters successfully', async () => {
            const mockResponse = {
                data: {
                    clusters: [{ name: 'cluster1' }, { name: 'cluster2' }],
                },
            };
            mockAxiosInstance.get.mockResolvedValue(mockResponse);

            const result = await serviceApi.listClusters();

            expect(mockAxiosInstance.get).toHaveBeenCalledWith(
                '/api/v1alpha1/clusters'
            );
            expect(result).toEqual(mockResponse.data.clusters);
        });

        it('should handle fetch clusters error', async () => {
            const mockError = new Error('Clusters unavailable');
            mockAxiosInstance.get.mockRejectedValue(mockError);

            await expect(serviceApi.listClusters()).rejects.toThrow(
                'Clusters unavailable'
            );
        });
    });

    describe('getIstioResources', () => {
        it('should fetch Istio resources successfully', async () => {
            const mockResponse = {
                data: {
                    gateways: [],
                    virtualServices: [],
                    destinationRules: [],
                },
            };
            mockAxiosInstance.get.mockResolvedValue(mockResponse);

            const result = await serviceApi.getIstioResources(
                'test-service-id',
                'test-instance-id'
            );

            expect(mockAxiosInstance.get).toHaveBeenCalledWith(
                '/api/v1alpha1/services/test-service-id/instances/test-instance-id/istio-resources'
            );
            expect(result).toEqual(mockResponse.data);
        });

        it('should handle fetch Istio resources error', async () => {
            const mockError = new Error('Istio resources unavailable');
            mockAxiosInstance.get.mockRejectedValue(mockError);

            await expect(
                serviceApi.getIstioResources(
                    'test-service-id',
                    'test-instance-id'
                )
            ).rejects.toThrow('Istio resources unavailable');
        });
    });
});
