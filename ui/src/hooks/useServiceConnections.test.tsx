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
import { useServiceConnections } from './useServiceConnections';
import { MetricsServiceService } from '../types/generated/openapi-metrics_service';
import { useMetricsContext } from '../contexts/MetricsContext';

// Mock the MetricsServiceService
jest.mock('../types/generated/openapi-metrics_service', () => ({
    MetricsServiceService: {
        metricsServiceGetServiceConnections: jest.fn(),
    },
}));

// Mock the MetricsContext
jest.mock('../contexts/MetricsContext', () => ({
    useMetricsContext: jest.fn(),
}));

const mockedMetricsService = MetricsServiceService as jest.Mocked<
    typeof MetricsServiceService
>;
const mockedUseMetricsContext = useMetricsContext as jest.MockedFunction<
    typeof useMetricsContext
>;

describe('useServiceConnections', () => {
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

        // Default mock for metrics context
        mockedUseMetricsContext.mockReturnValue({
            timeRange: { label: 'Last hour', value: '1h', minutes: 60 },
            startTime: new Date('2024-01-01T00:00:00Z'),
            endTime: new Date('2024-01-01T01:00:00Z'),
            lastUpdated: null,
            isRefreshing: false,
            refreshTrigger: 1,
            setTimeRange: jest.fn(),
            triggerRefresh: jest.fn(),
            setRefreshing: jest.fn(),
            updateLastUpdated: jest.fn(),
        });
    });

    it('should fetch service connections successfully', async () => {
        const mockResponse = {
            servicePairs: [
                {
                    source: { name: 'service-a', namespace: 'default' },
                    destination: { name: 'service-b', namespace: 'default' },
                    metrics: { requestRate: 10, errorRate: 0.1 },
                },
            ],
        } as any; // eslint-disable-line @typescript-eslint/no-explicit-any
        mockedMetricsService.metricsServiceGetServiceConnections.mockResolvedValue(
            mockResponse
        );

        const { result } = renderHook(
            () => useServiceConnections('test-service', 'default'),
            { wrapper }
        );

        await waitFor(() => {
            expect(result.current.isSuccess).toBe(true);
        });

        expect(
            mockedMetricsService.metricsServiceGetServiceConnections
        ).toHaveBeenCalledWith(
            'test-service',
            'default',
            '2024-01-01T00:00:00.000Z',
            '2024-01-01T01:00:00.000Z'
        );
        expect(result.current.data).toEqual(mockResponse);
    });

    it('should handle error response from service', async () => {
        mockedUseMetricsContext.mockReturnValue({
            timeRange: { label: 'Last hour', value: '1h', minutes: 60 },
            startTime: new Date('2024-01-01T00:00:00Z'),
            endTime: new Date('2024-01-01T01:00:00Z'),
            lastUpdated: null,
            isRefreshing: false,
            refreshTrigger: 1,
            setTimeRange: jest.fn(),
            triggerRefresh: jest.fn(),
            setRefreshing: jest.fn(),
            updateLastUpdated: jest.fn(),
        });

        mockedMetricsService.metricsServiceGetServiceConnections.mockRejectedValue(
            new Error('Service not found')
        );

        const { result } = renderHook(
            () => useServiceConnections('test-service', 'default'),
            { wrapper }
        );

        await waitFor(
            () => {
                expect(result.current.isLoading).toBe(false);
            },
            { timeout: 5000 }
        );

        expect(result.current.isError).toBe(true);

        expect(result.current.error).toEqual(new Error('Service not found'));
    });

    it('should not fetch when refreshTrigger is 0', () => {
        mockedUseMetricsContext.mockReturnValue({
            timeRange: { label: 'Last hour', value: '1h', minutes: 60 },
            startTime: new Date('2024-01-01T00:00:00Z'),
            endTime: new Date('2024-01-01T01:00:00Z'),
            lastUpdated: null,
            isRefreshing: false,
            refreshTrigger: 0,
            setTimeRange: jest.fn(),
            triggerRefresh: jest.fn(),
            setRefreshing: jest.fn(),
            updateLastUpdated: jest.fn(),
        });

        const { result } = renderHook(
            () => useServiceConnections('test-service', 'default'),
            { wrapper }
        );

        expect(result.current.isLoading).toBe(false);
        expect(
            mockedMetricsService.metricsServiceGetServiceConnections
        ).not.toHaveBeenCalled();
    });

    it('should refetch when refreshTrigger changes', async () => {
        const mockResponse = { servicePairs: [] } as any; // eslint-disable-line @typescript-eslint/no-explicit-any
        mockedMetricsService.metricsServiceGetServiceConnections.mockResolvedValue(
            mockResponse
        );

        // Initial render with refreshTrigger: 1
        const { rerender } = renderHook(
            () => useServiceConnections('test-service', 'default'),
            { wrapper }
        );

        await waitFor(() => {
            expect(
                mockedMetricsService.metricsServiceGetServiceConnections
            ).toHaveBeenCalledTimes(1);
        });

        // Update refreshTrigger to 2
        mockedUseMetricsContext.mockReturnValue({
            timeRange: { label: 'Last hour', value: '1h', minutes: 60 },
            startTime: new Date('2024-01-01T00:00:00Z'),
            endTime: new Date('2024-01-01T01:00:00Z'),
            lastUpdated: null,
            isRefreshing: false,
            refreshTrigger: 2,
            setTimeRange: jest.fn(),
            triggerRefresh: jest.fn(),
            setRefreshing: jest.fn(),
            updateLastUpdated: jest.fn(),
        });

        rerender();

        await waitFor(() => {
            expect(
                mockedMetricsService.metricsServiceGetServiceConnections
            ).toHaveBeenCalledTimes(2);
        });
    });

    it('should handle network errors', async () => {
        mockedUseMetricsContext.mockReturnValue({
            timeRange: { label: 'Last hour', value: '1h', minutes: 60 },
            startTime: new Date('2024-01-01T00:00:00Z'),
            endTime: new Date('2024-01-01T01:00:00Z'),
            lastUpdated: null,
            isRefreshing: false,
            refreshTrigger: 1,
            setTimeRange: jest.fn(),
            triggerRefresh: jest.fn(),
            setRefreshing: jest.fn(),
            updateLastUpdated: jest.fn(),
        });

        mockedMetricsService.metricsServiceGetServiceConnections.mockRejectedValue(
            new Error('Network error')
        );

        const { result } = renderHook(
            () => useServiceConnections('test-service', 'default'),
            { wrapper }
        );

        await waitFor(
            () => {
                expect(result.current.isLoading).toBe(false);
            },
            { timeout: 5000 }
        );

        expect(result.current.isError).toBe(true);

        expect(result.current.error).toEqual(new Error('Network error'));
    });

    it('should use correct query key for caching', () => {
        mockedUseMetricsContext.mockReturnValue({
            timeRange: { label: 'Last hour', value: '1h', minutes: 60 },
            startTime: new Date('2024-01-01T00:00:00Z'),
            endTime: new Date('2024-01-01T01:00:00Z'),
            lastUpdated: null,
            isRefreshing: false,
            refreshTrigger: 5,
            setTimeRange: jest.fn(),
            triggerRefresh: jest.fn(),
            setRefreshing: jest.fn(),
            updateLastUpdated: jest.fn(),
        });

        renderHook(() => useServiceConnections('my-service', 'my-namespace'), {
            wrapper,
        });

        // Query should be cached with the correct key
        const queries = queryClient.getQueryCache().getAll();
        expect(queries).toHaveLength(1);
        expect(queries[0].queryKey).toEqual([
            'serviceConnections',
            'my-service',
            'my-namespace',
            5,
        ]);
    });
});
