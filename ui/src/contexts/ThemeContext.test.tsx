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

import { render, act, renderHook } from '@testing-library/react';
import * as React from 'react';
import { ThemeProvider, useTheme } from './ThemeContext';

// Mock localStorage
const localStorageMock = (() => {
    let store: Record<string, string> = {};

    return {
        getItem: jest.fn((key: string) => {
            return store[key] || null;
        }),
        setItem: jest.fn((key: string, value: string) => {
            store[key] = value.toString();
        }),
        removeItem: jest.fn((key: string) => {
            delete store[key];
        }),
        clear: jest.fn(() => {
            store = {};
        }),
    };
})();

Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
});

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: jest.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: jest.fn(), // deprecated
        removeListener: jest.fn(), // deprecated
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
    })),
});

describe('ThemeContext', () => {
    beforeEach(() => {
        localStorageMock.clear();
        jest.clearAllMocks();
    });

    afterEach(() => {
        // Clean up DOM
        document.documentElement.classList.remove('dark', 'light');
    });

    describe('ThemeProvider', () => {
        it('should provide theme context to children', () => {
            const TestComponent = () => {
                const { theme, actualTheme } = useTheme();
                return (
                    <div>
                        <span data-testid="theme">{theme}</span>
                        <span data-testid="actual-theme">{actualTheme}</span>
                    </div>
                );
            };

            const { getByTestId } = render(
                <ThemeProvider>
                    <TestComponent />
                </ThemeProvider>
            );

            expect(getByTestId('theme').textContent).toBe('system');
            expect(getByTestId('actual-theme').textContent).toBe('light');
        });

        it('should load theme from localStorage', () => {
            localStorageMock.setItem('theme', 'dark');

            const TestComponent = () => {
                const { theme, actualTheme } = useTheme();
                return (
                    <div>
                        <span data-testid="theme">{theme}</span>
                        <span data-testid="actual-theme">{actualTheme}</span>
                    </div>
                );
            };

            const { getByTestId } = render(
                <ThemeProvider>
                    <TestComponent />
                </ThemeProvider>
            );

            expect(getByTestId('theme').textContent).toBe('dark');
            expect(getByTestId('actual-theme').textContent).toBe('dark');
        });

        it('should set theme and persist to localStorage', () => {
            const TestComponent = () => {
                const { theme, setTheme, actualTheme } = useTheme();
                return (
                    <div>
                        <span data-testid="theme">{theme}</span>
                        <span data-testid="actual-theme">{actualTheme}</span>
                        <button
                            onClick={() => setTheme('dark')}
                            data-testid="set-dark"
                        >
                            Set Dark
                        </button>
                    </div>
                );
            };

            const { getByTestId } = render(
                <ThemeProvider>
                    <TestComponent />
                </ThemeProvider>
            );

            act(() => {
                getByTestId('set-dark').click();
            });

            expect(getByTestId('theme').textContent).toBe('dark');
            expect(getByTestId('actual-theme').textContent).toBe('dark');
            expect(localStorageMock.setItem).toHaveBeenCalledWith(
                'theme',
                'dark'
            );
        });

        it('should handle system theme with light preference', () => {
            // Mock matchMedia to return light theme
            (window.matchMedia as jest.Mock).mockImplementation(() => ({
                matches: false, // false means light theme
                addEventListener: jest.fn(),
                removeEventListener: jest.fn(),
            }));

            const TestComponent = () => {
                const { theme, actualTheme } = useTheme();
                return (
                    <div>
                        <span data-testid="theme">{theme}</span>
                        <span data-testid="actual-theme">{actualTheme}</span>
                    </div>
                );
            };

            const { getByTestId } = render(
                <ThemeProvider>
                    <TestComponent />
                </ThemeProvider>
            );

            expect(getByTestId('theme').textContent).toBe('system');
            expect(getByTestId('actual-theme').textContent).toBe('light');
        });

        it('should handle system theme with dark preference', () => {
            // Mock matchMedia to return dark theme
            (window.matchMedia as jest.Mock).mockImplementation(() => ({
                matches: true, // true means dark theme
                addEventListener: jest.fn(),
                removeEventListener: jest.fn(),
            }));

            const TestComponent = () => {
                const { theme, actualTheme } = useTheme();
                return (
                    <div>
                        <span data-testid="theme">{theme}</span>
                        <span data-testid="actual-theme">{actualTheme}</span>
                    </div>
                );
            };

            const { getByTestId } = render(
                <ThemeProvider>
                    <TestComponent />
                </ThemeProvider>
            );

            expect(getByTestId('theme').textContent).toBe('system');
            expect(getByTestId('actual-theme').textContent).toBe('dark');
        });

        it('should apply dark class to document element', () => {
            const TestComponent = () => {
                const { setTheme } = useTheme();
                return (
                    <button
                        onClick={() => setTheme('dark')}
                        data-testid="set-dark"
                    >
                        Set Dark
                    </button>
                );
            };

            const { getByTestId } = render(
                <ThemeProvider>
                    <TestComponent />
                </ThemeProvider>
            );

            act(() => {
                getByTestId('set-dark').click();
            });

            expect(document.documentElement.classList.contains('dark')).toBe(
                true
            );
        });

        it('should remove dark class when switching to light', () => {
            // First set dark theme
            const TestComponent = () => {
                const { setTheme } = useTheme();
                return (
                    <div>
                        <button
                            onClick={() => setTheme('dark')}
                            data-testid="set-dark"
                        >
                            Set Dark
                        </button>
                        <button
                            onClick={() => setTheme('light')}
                            data-testid="set-light"
                        >
                            Set Light
                        </button>
                    </div>
                );
            };

            const { getByTestId } = render(
                <ThemeProvider>
                    <TestComponent />
                </ThemeProvider>
            );

            act(() => {
                getByTestId('set-dark').click();
            });

            expect(document.documentElement.classList.contains('dark')).toBe(
                true
            );

            act(() => {
                getByTestId('set-light').click();
            });

            expect(document.documentElement.classList.contains('dark')).toBe(
                false
            );
        });

        it('should listen for system theme changes', () => {
            const addEventListenerSpy = jest.fn();
            const removeEventListenerSpy = jest.fn();

            (window.matchMedia as jest.Mock).mockImplementation(() => ({
                matches: false,
                addEventListener: addEventListenerSpy,
                removeEventListener: removeEventListenerSpy,
            }));

            const { unmount } = render(
                <ThemeProvider>
                    <div>Test</div>
                </ThemeProvider>
            );

            expect(addEventListenerSpy).toHaveBeenCalledWith(
                'change',
                expect.any(Function)
            );

            unmount();

            expect(removeEventListenerSpy).toHaveBeenCalledWith(
                'change',
                expect.any(Function)
            );
        });
    });

    describe('useTheme', () => {
        it('should throw error when used outside provider', () => {
            // Suppress console.error for this test
            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {});

            expect(() => {
                renderHook(() => useTheme());
            }).toThrow('useTheme must be used within a ThemeProvider');

            consoleSpy.mockRestore();
        });
    });
});
