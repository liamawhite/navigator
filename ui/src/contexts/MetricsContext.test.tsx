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
import React from 'react';
import {
    MetricsProvider,
    useMetricsContext,
    TIME_RANGES,
} from './MetricsContext';

describe('MetricsContext', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        jest.setSystemTime(new Date('2024-01-01T12:00:00Z'));
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    describe('MetricsProvider', () => {
        it('should provide metrics context with default values', () => {
            const TestComponent = () => {
                const context = useMetricsContext();
                return (
                    <div>
                        <span data-testid="time-range">
                            {context.timeRange.value}
                        </span>
                        <span data-testid="is-refreshing">
                            {context.isRefreshing.toString()}
                        </span>
                        <span data-testid="refresh-trigger">
                            {context.refreshTrigger}
                        </span>
                        <span data-testid="last-updated">
                            {context.lastUpdated
                                ? context.lastUpdated.toISOString()
                                : 'null'}
                        </span>
                    </div>
                );
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            expect(getByTestId('time-range').textContent).toBe('5m');
            expect(getByTestId('is-refreshing').textContent).toBe('false');
            expect(getByTestId('refresh-trigger').textContent).toBe('0');
            expect(getByTestId('last-updated').textContent).toBe('null');
        });

        it('should calculate correct start and end times', () => {
            const TestComponent = () => {
                const { startTime, endTime, timeRange } = useMetricsContext();
                return (
                    <div>
                        <span data-testid="start-time">
                            {startTime.toISOString()}
                        </span>
                        <span data-testid="end-time">
                            {endTime.toISOString()}
                        </span>
                        <span data-testid="time-range">
                            {timeRange.minutes}
                        </span>
                    </div>
                );
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            // Default is 5 minutes, so start time should be 5 minutes before current time
            expect(getByTestId('start-time').textContent).toBe(
                '2024-01-01T11:55:00.000Z'
            );
            expect(getByTestId('end-time').textContent).toBe(
                '2024-01-01T12:00:00.000Z'
            );
            expect(getByTestId('time-range').textContent).toBe('5');
        });

        it('should update time range and recalculate times', () => {
            const TestComponent = () => {
                const { startTime, endTime, timeRange, setTimeRange } =
                    useMetricsContext();
                return (
                    <div>
                        <span data-testid="start-time">
                            {startTime.toISOString()}
                        </span>
                        <span data-testid="end-time">
                            {endTime.toISOString()}
                        </span>
                        <span data-testid="time-range">{timeRange.value}</span>
                        <button
                            onClick={() => setTimeRange(TIME_RANGES[0])} // 5m
                            data-testid="set-5m"
                        >
                            Set 5m
                        </button>
                    </div>
                );
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            act(() => {
                getByTestId('set-5m').click();
            });

            expect(getByTestId('time-range').textContent).toBe('5m');
            expect(getByTestId('start-time').textContent).toBe(
                '2024-01-01T11:55:00.000Z'
            );
            expect(getByTestId('end-time').textContent).toBe(
                '2024-01-01T12:00:00.000Z'
            );
        });

        it('should trigger refresh and increment counter', () => {
            const TestComponent = () => {
                const { refreshTrigger, triggerRefresh } = useMetricsContext();
                return (
                    <div>
                        <span data-testid="refresh-trigger">
                            {refreshTrigger}
                        </span>
                        <button
                            onClick={triggerRefresh}
                            data-testid="trigger-refresh"
                        >
                            Refresh
                        </button>
                    </div>
                );
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            expect(getByTestId('refresh-trigger').textContent).toBe('0');

            act(() => {
                getByTestId('trigger-refresh').click();
            });

            expect(getByTestId('refresh-trigger').textContent).toBe('1');

            act(() => {
                getByTestId('trigger-refresh').click();
            });

            expect(getByTestId('refresh-trigger').textContent).toBe('2');
        });

        it('should manage refreshing state', () => {
            const TestComponent = () => {
                const { isRefreshing, setRefreshing } = useMetricsContext();
                return (
                    <div>
                        <span data-testid="is-refreshing">
                            {isRefreshing.toString()}
                        </span>
                        <button
                            onClick={() => setRefreshing(true)}
                            data-testid="set-refreshing-true"
                        >
                            Set Refreshing True
                        </button>
                        <button
                            onClick={() => setRefreshing(false)}
                            data-testid="set-refreshing-false"
                        >
                            Set Refreshing False
                        </button>
                    </div>
                );
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            expect(getByTestId('is-refreshing').textContent).toBe('false');

            act(() => {
                getByTestId('set-refreshing-true').click();
            });

            expect(getByTestId('is-refreshing').textContent).toBe('true');

            act(() => {
                getByTestId('set-refreshing-false').click();
            });

            expect(getByTestId('is-refreshing').textContent).toBe('false');
        });

        it('should update last updated timestamp', () => {
            const TestComponent = () => {
                const { lastUpdated, updateLastUpdated } = useMetricsContext();
                return (
                    <div>
                        <span data-testid="last-updated">
                            {lastUpdated ? lastUpdated.toISOString() : 'null'}
                        </span>
                        <button
                            onClick={updateLastUpdated}
                            data-testid="update-last-updated"
                        >
                            Update Last Updated
                        </button>
                    </div>
                );
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            expect(getByTestId('last-updated').textContent).toBe('null');

            act(() => {
                getByTestId('update-last-updated').click();
            });

            expect(getByTestId('last-updated').textContent).toBe(
                '2024-01-01T12:00:00.000Z'
            );

            // Advance time and update again
            jest.setSystemTime(new Date('2024-01-01T12:05:00Z'));

            act(() => {
                getByTestId('update-last-updated').click();
            });

            expect(getByTestId('last-updated').textContent).toBe(
                '2024-01-01T12:05:00.000Z'
            );
        });

        it('should memoize context value to prevent unnecessary re-renders', () => {
            let renderCount = 0;
            const TestComponent = () => {
                useMetricsContext(); // Access context to test memoization
                renderCount++;
                return <div data-testid="render-count">{renderCount}</div>;
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            expect(getByTestId('render-count').textContent).toBe('1');

            // Re-render the provider without changing any values
            render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            // Component should only render once due to memoization
            expect(renderCount).toBe(2); // Once for each render call
        });

        it('should handle different time ranges correctly', () => {
            const TestComponent = () => {
                const { startTime, endTime, setTimeRange } =
                    useMetricsContext();
                return (
                    <div>
                        <span data-testid="start-time">
                            {startTime.toISOString()}
                        </span>
                        <span data-testid="end-time">
                            {endTime.toISOString()}
                        </span>
                        {TIME_RANGES.map((range, index) => (
                            <button
                                key={range.value}
                                onClick={() => setTimeRange(range)}
                                data-testid={`set-range-${index}`}
                            >
                                Set {range.value}
                            </button>
                        ))}
                    </div>
                );
            };

            const { getByTestId } = render(
                <MetricsProvider>
                    <TestComponent />
                </MetricsProvider>
            );

            // Test a few different time ranges
            const testRanges = [TIME_RANGES[2], TIME_RANGES[4]]; // 30m and 6h

            testRanges.forEach((range, _index) => {
                act(() => {
                    getByTestId(
                        `set-range-${TIME_RANGES.indexOf(range)}`
                    ).click();
                });

                const expectedStartTime = new Date(
                    new Date('2024-01-01T12:00:00Z').getTime() -
                        range.minutes * 60 * 1000
                );

                expect(getByTestId('start-time').textContent).toBe(
                    expectedStartTime.toISOString()
                );
                expect(getByTestId('end-time').textContent).toBe(
                    '2024-01-01T12:00:00.000Z'
                );
            });
        });
    });

    describe('useMetricsContext', () => {
        it('should throw error when used outside provider', () => {
            const consoleSpy = jest
                .spyOn(console, 'error')
                .mockImplementation(() => {});

            expect(() => {
                renderHook(() => useMetricsContext());
            }).toThrow(
                'useMetricsContext must be used within a MetricsProvider'
            );

            consoleSpy.mockRestore();
        });
    });

    describe('TIME_RANGES', () => {
        it('should export correct time ranges', () => {
            expect(TIME_RANGES).toHaveLength(7);
            expect(TIME_RANGES[0]).toEqual({
                label: 'Last 5 minutes',
                value: '5m',
                minutes: 5,
            });
            expect(TIME_RANGES[6]).toEqual({
                label: 'Last 24 hours',
                value: '24h',
                minutes: 1440,
            });
        });
    });
});
