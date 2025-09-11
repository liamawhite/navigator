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

/**
 * Utilities for creating theme-aware colors for the graph visualization
 * that adapt to light/dark mode using CSS custom properties
 */

// Theme-aware color palette for clusters (simplified approach)
export const getClusterColors = () => {
    const isDark =
        typeof window !== 'undefined' &&
        document.documentElement.classList.contains('dark');

    if (isDark) {
        // Dark theme chart colors
        return [
            '#60a5fa', // blue-400
            '#34d399', // emerald-400
            '#fbbf24', // amber-400
            '#a78bfa', // violet-400
            '#f87171', // red-400
            '#38bdf8', // sky-400
            '#a3e635', // lime-400
            '#fb923c', // orange-400
        ];
    } else {
        // Light theme chart colors
        return [
            '#3b82f6', // blue-500
            '#10b981', // emerald-500
            '#f59e0b', // amber-500
            '#8b5cf6', // violet-500
            '#ef4444', // red-500
            '#06b6d4', // cyan-500
            '#84cc16', // lime-500
            '#f97316', // orange-500
        ];
    }
};

// Theme-aware edge colors based on health/error status
export const getEdgeColor = (requestRate: number, errorRate: number) => {
    if (errorRate > 0.1) {
        return '#ef4444'; // High error rate - red
    } else if (errorRate > 0.05) {
        return '#f59e0b'; // Medium error rate - amber
    } else if (requestRate > 0) {
        return '#10b981'; // Healthy traffic - emerald
    } else {
        return '#6b7280'; // No traffic - gray
    }
};

// Simplified color palette that works better with the limited OKLCH conversion
export const getSimpleThemeColors = () => {
    // Use a simpler approach that works reliably across themes
    const isDark = document.documentElement.classList.contains('dark');

    if (isDark) {
        return {
            background: '#0f0f0f',
            foreground: '#fafafa',
            primary: '#3b82f6',
            secondary: '#374151',
            muted: '#262626',
            mutedForeground: '#a3a3a3',
            border: 'rgba(255, 255, 255, 0.1)',
            destructive: '#ef4444',
        };
    } else {
        return {
            background: '#ffffff',
            foreground: '#0f0f0f',
            primary: '#1e40af',
            secondary: '#f3f4f6',
            muted: '#f3f4f6',
            mutedForeground: '#6b7280',
            border: '#e5e7eb',
            destructive: '#dc2626',
        };
    }
};

// Create a minimal, compatible theme for Reagraph
export const createSimpleGraphTheme = () => {
    const colors = getSimpleThemeColors();

    // Create a minimal theme that should work with Reagraph
    return {
        canvas: {
            background: colors.background,
        },
        node: {
            fill: colors.primary,
            activeFill: colors.primary,
            inactiveFill: colors.muted,
            opacity: 1.0,
            selectedOpacity: 1.0,
            inactiveOpacity: 0.3,
            label: {
                color: colors.foreground,
                activeColor: colors.foreground,
                inactiveColor: colors.mutedForeground,
                fontFamily: 'Inter, system-ui, sans-serif',
                fontSize: 12,
            },
        },
        edge: {
            fill: colors.mutedForeground,
            activeFill: colors.primary,
            inactiveFill: colors.muted,
            opacity: 0.8,
            selectedOpacity: 1.0,
            inactiveOpacity: 0.2,
            label: {
                stroke: colors.background,
                color: colors.foreground,
                activeColor: colors.foreground,
                inactiveColor: colors.mutedForeground,
                fontFamily: 'Inter, system-ui, sans-serif',
                fontSize: 10,
            },
        },
        lasso: {
            border: colors.primary,
            background: colors.muted,
        },
    };
};
