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

import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
    useServices,
    useService,
    useServiceInstance,
    useProxyConfig,
    useIstioResources,
} from './useServices';
import { serviceApi } from '../utils/api';

// Mock the API
jest.mock('../utils/api', () => ({
    serviceApi: {
        listServices: jest.fn(),
        getService: jest.fn(),
        getServiceInstance: jest.fn(),
        getProxyConfig: jest.fn(),
        getIstioResources: jest.fn(),
    },
}));

const mockedServiceApi = serviceApi as jest.Mocked<typeof serviceApi>;

describe('Services hooks', () => {
    let queryClient: QueryClient;
    let wrapper: React.ComponentType<{ children: React.ReactNode }>;

    beforeEach(() => {
        queryClient = new QueryClient({
            defaultOptions: {
                queries: {
                    retry: false,
                },
            },
        });
        wrapper = ({ children }: { children: React.ReactNode }) => (
            <QueryClientProvider client={queryClient}>
                {children}
            </QueryClientProvider>
        );
        jest.clearAllMocks();
    });

    describe('useServices', () => {
        it('should fetch services successfully', async () => {
            const mockServices = [
                { name: 'test-service', namespace: 'default' },
            ];
            mockedServiceApi.listServices.mockResolvedValue(mockServices);

            const { result } = renderHook(() => useServices(), { wrapper });

            await waitFor(() => {
                expect(result.current.isSuccess).toBe(true);
            });

            expect(mockedServiceApi.listServices).toHaveBeenCalled();
            expect(result.current.data).toEqual(mockServices);
        });

        it('should handle services fetch error', async () => {
            mockedServiceApi.listServices.mockRejectedValue(
                new Error('Network error')
            );

            const { result } = renderHook(() => useServices(), { wrapper });

            await waitFor(() => {
                expect(result.current.isError).toBe(true);
            });

            expect(result.current.error).toEqual(new Error('Network error'));
        });
    });

    describe('useService', () => {
        it('should fetch service when id is provided', async () => {
            const mockService = {
                name: 'test-service',
                namespace: 'default',
            };
            mockedServiceApi.getService.mockResolvedValue(mockService);

            const { result } = renderHook(() => useService('test-service'), {
                wrapper,
            });

            await waitFor(() => {
                expect(result.current.isSuccess).toBe(true);
            });

            expect(mockedServiceApi.getService).toHaveBeenCalledWith(
                'test-service'
            );
            expect(result.current.data).toEqual(mockService);
        });

        it('should not fetch when id is empty', () => {
            const { result } = renderHook(() => useService(''), { wrapper });

            expect(result.current.isLoading).toBe(false);
            expect(mockedServiceApi.getService).not.toHaveBeenCalled();
        });
    });

    describe('useServiceInstance', () => {
        it('should fetch service instance when both ids are provided', async () => {
            const mockInstance = {
                instanceId: 'test-instance',
                envoyPresent: true,
            };
            mockedServiceApi.getServiceInstance.mockResolvedValue(mockInstance);

            const { result } = renderHook(
                () => useServiceInstance('test-service', 'test-instance'),
                { wrapper }
            );

            await waitFor(() => {
                expect(result.current.isSuccess).toBe(true);
            });

            expect(mockedServiceApi.getServiceInstance).toHaveBeenCalledWith(
                'test-service',
                'test-instance'
            );
            expect(result.current.data).toEqual(mockInstance);
        });

        it('should not fetch when serviceId is empty', () => {
            const { result } = renderHook(
                () => useServiceInstance('', 'test-instance'),
                { wrapper }
            );

            expect(result.current.isLoading).toBe(false);
            expect(mockedServiceApi.getServiceInstance).not.toHaveBeenCalled();
        });

        it('should not fetch when instanceId is empty', () => {
            const { result } = renderHook(
                () => useServiceInstance('test-service', ''),
                { wrapper }
            );

            expect(result.current.isLoading).toBe(false);
            expect(mockedServiceApi.getServiceInstance).not.toHaveBeenCalled();
        });
    });

    describe('useProxyConfig', () => {
        it('should fetch proxy config when both ids are provided', async () => {
            const mockConfig = {
                bootstrap: {},
                clusters: [],
            } as any; // eslint-disable-line @typescript-eslint/no-explicit-any
            mockedServiceApi.getProxyConfig.mockResolvedValue(mockConfig);

            const { result } = renderHook(
                () => useProxyConfig('test-service', 'test-instance'),
                { wrapper }
            );

            await waitFor(() => {
                expect(result.current.isSuccess).toBe(true);
            });

            expect(mockedServiceApi.getProxyConfig).toHaveBeenCalledWith(
                'test-service',
                'test-instance'
            );
            expect(result.current.data).toEqual(mockConfig);
        });

        it('should not fetch when serviceId is empty', () => {
            const { result } = renderHook(
                () => useProxyConfig('', 'test-instance'),
                { wrapper }
            );

            expect(result.current.isLoading).toBe(false);
            expect(mockedServiceApi.getProxyConfig).not.toHaveBeenCalled();
        });
    });

    describe('useIstioResources', () => {
        it('should fetch Istio resources when both ids are provided', async () => {
            const mockResources = {
                gateways: [],
                virtualServices: [],
                destinationRules: [],
            };
            mockedServiceApi.getIstioResources.mockResolvedValue(mockResources);

            const { result } = renderHook(
                () => useIstioResources('test-service', 'test-instance'),
                { wrapper }
            );

            await waitFor(() => {
                expect(result.current.isSuccess).toBe(true);
            });

            expect(mockedServiceApi.getIstioResources).toHaveBeenCalledWith(
                'test-service',
                'test-instance'
            );
            expect(result.current.data).toEqual(mockResources);
        });

        it('should not fetch when instanceId is empty', () => {
            const { result } = renderHook(
                () => useIstioResources('test-service', ''),
                { wrapper }
            );

            expect(result.current.isLoading).toBe(false);
            expect(mockedServiceApi.getIstioResources).not.toHaveBeenCalled();
        });
    });
});
