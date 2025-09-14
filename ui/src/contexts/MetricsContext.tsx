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

import React, { createContext, useContext, useState, useCallback } from 'react';

export interface TimeRange {
    label: string;
    value: string;
    minutes: number;
}

export const TIME_RANGES: TimeRange[] = [
    { label: 'Last 5 minutes', value: '5m', minutes: 5 },
    { label: 'Last 15 minutes', value: '15m', minutes: 15 },
    { label: 'Last 30 minutes', value: '30m', minutes: 30 },
    { label: 'Last 1 hour', value: '1h', minutes: 60 },
    { label: 'Last 6 hours', value: '6h', minutes: 360 },
    { label: 'Last 12 hours', value: '12h', minutes: 720 },
    { label: 'Last 24 hours', value: '24h', minutes: 1440 },
];

interface MetricsContextType {
    timeRange: TimeRange;
    startTime: Date;
    endTime: Date;
    lastUpdated: Date | null;
    isRefreshing: boolean;
    refreshTrigger: number;
    setTimeRange: (timeRange: TimeRange) => void;
    triggerRefresh: () => void;
    setRefreshing: (refreshing: boolean) => void;
    updateLastUpdated: () => void;
}

const MetricsContext = createContext<MetricsContextType | undefined>(undefined);

export const useMetricsContext = () => {
    const context = useContext(MetricsContext);
    if (context === undefined) {
        throw new Error(
            'useMetricsContext must be used within a MetricsProvider'
        );
    }
    return context;
};

interface MetricsProviderProps {
    children: React.ReactNode;
}

export const MetricsProvider: React.FC<MetricsProviderProps> = ({
    children,
}) => {
    // Initialize time range from localStorage or default to 5 minutes
    const [timeRange, setTimeRangeState] = useState<TimeRange>(() => {
        try {
            const saved = localStorage.getItem('metrics-time-range');
            if (saved) {
                const savedValue = JSON.parse(saved);
                const found = TIME_RANGES.find((r) => r.value === savedValue);
                if (found) return found;
            }
        } catch (error) {
            console.warn('Failed to load saved time range:', error);
        }
        return TIME_RANGES[0]; // Default to 5 minutes
    });
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
    const [isRefreshing, setIsRefreshing] = useState(false);
    const [refreshTrigger, setRefreshTrigger] = useState(0); // Start with 0 so queries don't run initially

    const endTime = new Date();
    const startTime = new Date(
        endTime.getTime() - timeRange.minutes * 60 * 1000
    );

    const setTimeRange = useCallback((newTimeRange: TimeRange) => {
        setTimeRangeState(newTimeRange);
        try {
            localStorage.setItem(
                'metrics-time-range',
                JSON.stringify(newTimeRange.value)
            );
        } catch (error) {
            console.warn('Failed to save time range to localStorage:', error);
        }
    }, []);

    const triggerRefresh = useCallback(() => {
        setRefreshTrigger((prev) => prev + 1);
        setIsRefreshing(true);
    }, []);

    const setRefreshing = useCallback((refreshing: boolean) => {
        setIsRefreshing(refreshing);
    }, []);

    const updateLastUpdated = useCallback(() => {
        setLastUpdated(new Date());
        setIsRefreshing(false);
    }, []);

    const value: MetricsContextType = {
        timeRange,
        startTime,
        endTime,
        lastUpdated,
        isRefreshing,
        refreshTrigger,
        setTimeRange,
        triggerRefresh,
        setRefreshing,
        updateLastUpdated,
    };

    return (
        <MetricsContext.Provider value={value}>
            {children}
        </MetricsContext.Provider>
    );
};
