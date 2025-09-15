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
import { useClusters } from './useClusters';
import { serviceApi } from '../utils/api';

// Mock the API
jest.mock('../utils/api', () => ({
    serviceApi: {
        listClusters: jest.fn(),
    },
}));

const mockedServiceApi = serviceApi as jest.Mocked<typeof serviceApi>;

describe('useClusters', () => {
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

    it('should fetch clusters successfully', async () => {
        const mockClusters = [
            {
                name: 'cluster-1',
                syncStatus: 'SYNCED',
                lastUpdated: '2024-01-01T12:00:00Z',
            },
            {
                name: 'cluster-2',
                syncStatus: 'SYNCING',
                lastUpdated: '2024-01-01T12:01:00Z',
            },
        ] as any; // eslint-disable-line @typescript-eslint/no-explicit-any
        mockedServiceApi.listClusters.mockResolvedValue(mockClusters);

        const { result } = renderHook(() => useClusters(), { wrapper });

        await waitFor(() => {
            expect(result.current.isSuccess).toBe(true);
        });

        expect(mockedServiceApi.listClusters).toHaveBeenCalled();
        expect(result.current.data).toEqual(mockClusters);
    });

    it('should handle clusters fetch error', async () => {
        mockedServiceApi.listClusters.mockRejectedValue(
            new Error('Clusters unavailable')
        );

        const { result } = renderHook(() => useClusters(), { wrapper });

        await waitFor(() => {
            expect(result.current.isError).toBe(true);
        });

        expect(result.current.error).toEqual(new Error('Clusters unavailable'));
    });

    it('should use correct query key for caching', () => {
        renderHook(() => useClusters(), { wrapper });

        const queries = queryClient.getQueryCache().getAll();
        expect(queries).toHaveLength(1);
        expect(queries[0].queryKey).toEqual(['clusters']);
    });

    it('should have refetch interval configured', () => {
        renderHook(() => useClusters(), { wrapper });

        // The query should have refetchInterval configured
        const query = queryClient.getQueryCache().getAll()[0];
        expect((query.options as any).refetchInterval).toBe(30000); // eslint-disable-line @typescript-eslint/no-explicit-any
    });
});
