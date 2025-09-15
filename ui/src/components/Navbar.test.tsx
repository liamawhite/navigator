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
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { Navbar } from './Navbar';
import { useTheme } from './theme-provider';

jest.mock('./theme-provider', () => ({
    useTheme: jest.fn(),
}));

jest.mock('./ClusterSyncStatus', () => ({
    ClusterSyncStatus: () => (
        <div data-testid="cluster-sync-status">ClusterSyncStatus</div>
    ),
}));

jest.mock('./mode-toggle', () => ({
    ModeToggle: () => <div data-testid="mode-toggle">ModeToggle</div>,
}));

const mockedUseTheme = useTheme as jest.MockedFunction<typeof useTheme>;

const renderWithRouter = (component: React.ReactElement) => {
    return render(<BrowserRouter>{component}</BrowserRouter>);
};

describe('Navbar', () => {
    beforeEach(() => {
        mockedUseTheme.mockReturnValue({
            theme: 'light',
            setTheme: jest.fn(),
        });
    });

    it('should render Navigator logo', () => {
        renderWithRouter(<Navbar />);

        const logo = screen.getByAltText('Navigator');
        expect(logo).toBeTruthy();
        expect(logo.getAttribute('src')).toBe('/navigator.svg');
    });

    it('should render Service Registry link', () => {
        renderWithRouter(<Navbar />);

        const link = screen.getByRole('link', { name: /service registry/i });
        expect(link).toBeTruthy();
        expect(link.getAttribute('href')).toBe('/');
    });

    it('should render ClusterSyncStatus component', () => {
        renderWithRouter(<Navbar />);

        expect(screen.getByTestId('cluster-sync-status')).toBeTruthy();
    });

    it('should render ModeToggle component', () => {
        renderWithRouter(<Navbar />);

        expect(screen.getByTestId('mode-toggle')).toBeTruthy();
    });

    it('should have proper navigation structure', () => {
        renderWithRouter(<Navbar />);

        const nav = screen.getByRole('navigation');
        expect(nav).toBeTruthy();

        // Check that all main elements are present
        expect(screen.getByAltText('Navigator')).toBeTruthy();
        expect(screen.getByText(/service registry/i)).toBeTruthy();
        expect(screen.getByTestId('cluster-sync-status')).toBeTruthy();
        expect(screen.getByTestId('mode-toggle')).toBeTruthy();
    });
});
