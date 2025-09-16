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

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ServiceConnectionsCard } from './ServiceConnectionsCard';
import { useServiceConnections } from '../../hooks/useServiceConnections';
import { useClusters } from '../../hooks/useClusters';
import { useMetricsContext } from '../../contexts/MetricsContext';

jest.mock('../../hooks/useServiceConnections', () => ({
    useServiceConnections: jest.fn(),
}));

jest.mock('../../hooks/useClusters', () => ({
    useClusters: jest.fn(),
}));

jest.mock('../../contexts/MetricsContext', () => ({
    useMetricsContext: jest.fn(),
    TIME_RANGES: [
        { label: 'Last 5 minutes', value: '5m', minutes: 5 },
        { label: 'Last 1 hour', value: '1h', minutes: 60 },
    ],
}));

jest.mock('./ServiceConnectionsTable', () => ({
    ServiceConnectionsTable: ({
        inbound,
        outbound,
    }: {
        inbound: any[];
        outbound: any[];
    }) => (
        <div data-testid="service-connections-table">
            Connections: {inbound.length} inbound, {outbound.length} outbound
        </div>
    ),
}));

const mockedUseServiceConnections =
    useServiceConnections as jest.MockedFunction<typeof useServiceConnections>;
const mockedUseClusters = useClusters as jest.MockedFunction<
    typeof useClusters
>;
const mockedUseMetricsContext = useMetricsContext as jest.MockedFunction<
    typeof useMetricsContext
>;

describe('ServiceConnectionsCard', () => {
    const mockMetricsContext = {
        timeRange: { label: 'Last 1 hour', value: '1h', minutes: 60 },
        startTime: new Date('2024-01-01T11:00:00Z'),
        endTime: new Date('2024-01-01T12:00:00Z'),
        lastUpdated: new Date('2024-01-01T12:00:00Z'),
        isRefreshing: false,
        refreshTrigger: 1,
        setRefreshing: jest.fn(),
        updateLastUpdated: jest.fn(),
        triggerRefresh: jest.fn(),
        setTimeRange: jest.fn(),
    };

    beforeEach(() => {
        jest.clearAllMocks();
        mockedUseMetricsContext.mockReturnValue(mockMetricsContext);
        mockedUseClusters.mockReturnValue({
            data: [{ clusterId: 'cluster-1', metricsEnabled: true }],
            isLoading: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    });

    it('should render service connections card', () => {
        mockedUseServiceConnections.mockReturnValue({
            data: null,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        expect(screen.getByText('Service Connections')).toBeTruthy();
        expect(screen.getByText('alpha')).toBeTruthy();
    });

    it('should show loading state', () => {
        mockedUseServiceConnections.mockReturnValue({
            data: null,
            isLoading: true,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        expect(screen.getByText('Service Connections')).toBeTruthy();
        // Should show loading spinner in the content area
        const spinner = document.querySelector('.animate-spin');
        expect(spinner).toBeTruthy();
    });

    it('should show error state', () => {
        const errorMessage = 'Failed to load connections';
        mockedUseServiceConnections.mockReturnValue({
            data: null,
            isLoading: false,
            error: new Error(errorMessage),
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        expect(
            screen.getByText('Failed to load service connections')
        ).toBeTruthy();
        expect(screen.getByText(errorMessage)).toBeTruthy();
    });

    it('should show no connections message when no data', () => {
        mockedUseServiceConnections.mockReturnValue({
            data: { inbound: [], outbound: [] },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        expect(screen.getByText('No service connections found')).toBeTruthy();
    });

    it('should render visualization when connections exist', () => {
        const mockConnections = {
            inbound: [{ serviceName: 'frontend', namespace: 'default' }],
            outbound: [{ serviceName: 'database', namespace: 'default' }],
        };

        mockedUseServiceConnections.mockReturnValue({
            data: mockConnections,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        expect(screen.getByTestId('service-connections-table')).toBeTruthy();
        expect(screen.getByText('Connections: 1 inbound, 1 outbound')).toBeTruthy();
    });

    it('should show collapsed state when no metrics enabled', () => {
        mockedUseClusters.mockReturnValue({
            data: [{ clusterId: 'cluster-1', metricsEnabled: false }],
            isLoading: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        mockedUseServiceConnections.mockReturnValue({
            data: null,
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        expect(
            screen.getByText(
                'Requires metrics to be enabled on at least one cluster'
            )
        ).toBeTruthy();
    });

    it('should trigger refresh when refresh button is clicked', async () => {
        mockedUseServiceConnections.mockReturnValue({
            data: { inbound: [], outbound: [] },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        // Find refresh button by the RefreshCw icon
        const refreshButton = screen.getByRole('button', { name: '' });
        fireEvent.click(refreshButton);

        expect(mockMetricsContext.triggerRefresh).toHaveBeenCalled();
    });

    it('should handle time range selection', async () => {
        mockedUseServiceConnections.mockReturnValue({
            data: { inbound: [], outbound: [] },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        // Find and click the time range selector
        const timeRangeButton = screen.getByRole('combobox');
        fireEvent.click(timeRangeButton);

        await waitFor(() => {
            const option = screen.getByText('Last 5 minutes');
            fireEvent.click(option);
        });

        expect(mockMetricsContext.setTimeRange).toHaveBeenCalledWith({
            label: 'Last 5 minutes',
            value: '5m',
            minutes: 5,
        });
    });

    it('should disable refresh button when refreshing', () => {
        mockedUseMetricsContext.mockReturnValue({
            ...mockMetricsContext,
            isRefreshing: true,
        });

        mockedUseServiceConnections.mockReturnValue({
            data: { inbound: [], outbound: [] },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        // Find button by querying all buttons and checking for disabled state
        const buttons = screen.getAllByRole('button');
        const refreshButton = buttons.find((btn) =>
            btn.hasAttribute('disabled')
        );
        expect(refreshButton).toBeTruthy();
    });

    it('should show last updated time', () => {
        mockedUseServiceConnections.mockReturnValue({
            data: { inbound: [], outbound: [] },
            isLoading: false,
            error: null,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(
            <ServiceConnectionsCard
                serviceName="test-service"
                namespace="default"
            />
        );

        expect(screen.getByText(/Last updated:/)).toBeTruthy();
    });
});
