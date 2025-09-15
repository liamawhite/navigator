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
import { ServiceList } from './ServiceList';
import { useServices } from '../../hooks/useServices';
import { useClusters } from '../../hooks/useClusters';

jest.mock('../../hooks/useServices', () => ({
    useServices: jest.fn(),
}));

jest.mock('../../hooks/useClusters', () => ({
    useClusters: jest.fn(),
}));

const mockedUseServices = useServices as jest.MockedFunction<
    typeof useServices
>;
const mockedUseClusters = useClusters as jest.MockedFunction<
    typeof useClusters
>;

describe('ServiceList', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedUseClusters.mockReturnValue({
            data: [],
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any
    });

    const mockServices = [
        {
            id: 'service-1',
            name: 'frontend',
            namespace: 'production',
            instances: [
                {
                    instanceId: 'instance-1',
                    envoyPresent: true,
                    clusterName: 'cluster-1',
                },
                {
                    instanceId: 'instance-2',
                    envoyPresent: false,
                    clusterName: 'cluster-1',
                },
            ],
        },
        {
            id: 'service-2',
            name: 'backend',
            namespace: 'production',
            instances: [
                {
                    instanceId: 'instance-3',
                    envoyPresent: true,
                    clusterName: 'cluster-2',
                },
            ],
        },
        {
            id: 'service-3',
            name: 'database',
            namespace: 'staging',
            instances: [
                {
                    instanceId: 'instance-4',
                    envoyPresent: false,
                    clusterName: 'cluster-1',
                },
            ],
        },
    ];

    it('should display loading state', () => {
        mockedUseServices.mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        expect(screen.getByText('Loading services...')).toBeTruthy();
    });

    it('should display error state', () => {
        const errorMessage = 'Failed to fetch services';
        mockedUseServices.mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error(errorMessage),
            isError: true,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        expect(screen.getByText(/Failed to load services/)).toBeTruthy();
        expect(screen.getByText(/Failed to fetch services/)).toBeTruthy();
    });

    it('should display no services message when list is empty', () => {
        mockedUseServices.mockReturnValue({
            data: [],
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        expect(screen.getByText('No services found')).toBeTruthy();
    });

    it('should display services in a table', async () => {
        mockedUseServices.mockReturnValue({
            data: mockServices,
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        await waitFor(() => {
            expect(screen.getByText('Discovered Services')).toBeTruthy();
            expect(screen.getByText('3 services')).toBeTruthy();
            expect(screen.getByText('frontend')).toBeTruthy();
            expect(screen.getByText('backend')).toBeTruthy();
            expect(screen.getByText('database')).toBeTruthy();
        });
    });

    it('should display service details correctly', async () => {
        mockedUseServices.mockReturnValue({
            data: mockServices,
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        await waitFor(() => {
            // Check namespaces
            expect(screen.getAllByText('production')).toHaveLength(2);
            expect(screen.getByText('staging')).toBeTruthy();

            // Check that service details are displayed correctly - just verify they're there
            expect(screen.getByText('frontend')).toBeTruthy();
            expect(screen.getByText('backend')).toBeTruthy();
            expect(screen.getByText('database')).toBeTruthy();
        });
    });

    it('should call onServiceSelect when service row is clicked', async () => {
        const mockOnServiceSelect = jest.fn();
        mockedUseServices.mockReturnValue({
            data: mockServices,
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList onServiceSelect={mockOnServiceSelect} />);

        await waitFor(() => {
            expect(screen.getByText('frontend')).toBeTruthy();
        });

        const serviceRow = screen.getByText('frontend').closest('tr');
        fireEvent.click(serviceRow!);

        expect(mockOnServiceSelect).toHaveBeenCalledWith('service-1');
    });

    it('should sort services by namespace by default', async () => {
        mockedUseServices.mockReturnValue({
            data: mockServices,
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        await waitFor(() => {
            const rows = screen.getAllByRole('row');
            // Skip header row
            const serviceRows = rows.slice(1);

            // Should be sorted by namespace (production services first, then staging)
            expect(serviceRows[0].textContent).toContain('backend'); // production
            expect(serviceRows[1].textContent).toContain('frontend'); // production
            expect(serviceRows[2].textContent).toContain('database'); // staging
        });
    });

    it('should handle sorting by service name', async () => {
        mockedUseServices.mockReturnValue({
            data: mockServices,
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        await waitFor(() => {
            expect(screen.getByText('Service')).toBeTruthy();
        });

        // Click on Service header to sort by name
        fireEvent.click(screen.getByText('Service'));

        await waitFor(() => {
            const rows = screen.getAllByRole('row');
            const serviceRows = rows.slice(1);

            // Should be sorted by name alphabetically
            expect(serviceRows[0].textContent).toContain('backend');
            expect(serviceRows[1].textContent).toContain('database');
            expect(serviceRows[2].textContent).toContain('frontend');
        });
    });

    it('should display Envoy proxy indicators', async () => {
        mockedUseServices.mockReturnValue({
            data: mockServices,
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        await waitFor(() => {
            // Should show Envoy indicators - check for hexagon icons which indicate Envoy proxy
            const hexagonIcons = document.querySelectorAll('.lucide-hexagon');
            expect(hexagonIcons.length).toBeGreaterThan(0);
        });
    });

    it('should display initializing clusters message', () => {
        mockedUseServices.mockReturnValue({
            data: [],
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        mockedUseClusters.mockReturnValue({
            data: [
                {
                    clusterId: 'cluster-1',
                    syncStatus: 'SYNC_STATUS_INITIALIZING',
                },
                {
                    clusterId: 'cluster-2',
                    syncStatus: 'SYNC_STATUS_INITIALIZING',
                },
            ],
            isLoading: false,
            error: null,
            isError: false,
        } as any); // eslint-disable-line @typescript-eslint/no-explicit-any

        render(<ServiceList />);

        expect(screen.getByText('Initializing clusters')).toBeTruthy();
        expect(screen.getByText(/Connected to 2 clusters/)).toBeTruthy();
    });
});
