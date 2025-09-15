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
import { render, screen, fireEvent } from '@testing-library/react';
import { ModeToggle } from './mode-toggle';
import { useTheme } from './theme-provider';

jest.mock('./theme-provider', () => ({
    useTheme: jest.fn(),
}));

const mockedUseTheme = useTheme as jest.MockedFunction<typeof useTheme>;

describe('ModeToggle', () => {
    const mockSetTheme = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('should render toggle button', () => {
        mockedUseTheme.mockReturnValue({
            theme: 'light',
            setTheme: mockSetTheme,
        });

        render(<ModeToggle />);

        expect(screen.getByRole('button')).toBeTruthy();
        expect(screen.getByText('Toggle theme')).toBeTruthy();
    });

    it('should toggle from light to dark theme', () => {
        mockedUseTheme.mockReturnValue({
            theme: 'light',
            setTheme: mockSetTheme,
        });

        render(<ModeToggle />);

        const button = screen.getByRole('button');
        fireEvent.click(button);

        expect(mockSetTheme).toHaveBeenCalledWith('dark');
    });

    it('should toggle from dark to light theme', () => {
        mockedUseTheme.mockReturnValue({
            theme: 'dark',
            setTheme: mockSetTheme,
        });

        render(<ModeToggle />);

        const button = screen.getByRole('button');
        fireEvent.click(button);

        expect(mockSetTheme).toHaveBeenCalledWith('light');
    });

    it('should handle non-light themes by switching to light', () => {
        mockedUseTheme.mockReturnValue({
            theme: 'system' as any, // eslint-disable-line @typescript-eslint/no-explicit-any
            setTheme: mockSetTheme,
        });

        render(<ModeToggle />);

        const button = screen.getByRole('button');
        fireEvent.click(button);

        expect(mockSetTheme).toHaveBeenCalledWith('light');
    });
});
