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
 * Configuration constants for the D3.js graph visualization
 */

export const GRAPH_CONFIG = {
    // Node configuration
    NODE_RADIUS: 24,
    NODE_STROKE_WIDTH: 2,
    NODE_HOVER_RADIUS: 29,
    NODE_PADDING: 25,

    // Edge configuration
    ARROW_SIZE: 12,
    EDGE_OFFSET: 2, // Additional offset for stroke width
    MIN_EDGE_WIDTH: 1,
    MAX_EDGE_WIDTH: 6,

    // Layout configuration
    CLUSTER_PADDING: 50,
    NAMESPACE_PADDING: 20,
    NAMESPACE_HEIGHT_OFFSET: 60,
    NAMESPACE_LABEL_HEIGHT: 40,

    // Canvas configuration
    MIN_WIDTH: 800,
    MIN_HEIGHT: 600,

    // Force simulation configuration
    LINK_DISTANCE: 100,
    CHARGE_STRENGTH: -200,
    COLLISION_RADIUS: 30,
    ALPHA_TARGET: 0.3,

    // Performance thresholds
    MAX_NODES_BEFORE_VIRTUALIZATION: 100,
    MAX_SIMULATION_ITERATIONS: 300,
    SIMULATION_ALPHA_MIN: 0.001,

    // Animation and timing
    RESIZE_DEBOUNCE_MS: 100,
    THEME_CHANGE_DEBOUNCE_MS: 50,

    // Error rate thresholds for coloring
    ERROR_RATE_THRESHOLDS: {
        HIGH: 0.1, // 10% - Red
        MEDIUM: 0.05, // 5% - Orange
        LOW: 0.01, // 1% - Yellow
        // Below 1% - Green
    },

    // Success rate display thresholds
    SUCCESS_RATE_THRESHOLDS: {
        HIGH: 0.99, // 99% - Green
        MEDIUM: 0.95, // 95% - Orange
        // Below 95% - Red
    },

    // Colors (theme-aware)
    COLORS: {
        // Cluster colors (slate-based)
        CLUSTER_PALETTE: [
            '#64748b', // slate-500
            '#6b7280', // gray-500
            '#71717a', // zinc-500
            '#737373', // neutral-500
            '#78716c', // stone-500
            '#475569', // slate-600
            '#4b5563', // gray-600
            '#52525b', // zinc-600
        ],

        // Error rate colors
        ERROR_COLORS: {
            HIGH: '#ef4444', // Red
            MEDIUM: '#f59e0b', // Orange
            LOW: '#eab308', // Yellow
            HEALTHY: '#10b981', // Green
        },

        // Success rate colors
        SUCCESS_COLORS: {
            HIGH: '#10b981', // Green
            MEDIUM: '#f59e0b', // Orange
            LOW: '#ef4444', // Red
        },

        // Theme colors
        DARK_THEME: {
            BACKGROUND: '#0f0f0f',
            FOREGROUND: '#fafafa',
            NODE_FILL: '#374151',
            NODE_STROKE: '#6b7280',
            CLUSTER_BACKGROUND: 'rgba(71, 85, 105, 0.15)',
            NAMESPACE_BACKGROUND: 'rgba(255, 255, 255, 0.08)',
            TEXT_STROKE: '#000000',
        },

        LIGHT_THEME: {
            BACKGROUND: '#ffffff',
            FOREGROUND: '#0f0f0f',
            NODE_FILL: '#e5e7eb',
            NODE_STROKE: '#9ca3af',
            CLUSTER_BACKGROUND: 'rgba(100, 116, 139, 0.08)',
            NAMESPACE_BACKGROUND: 'rgba(0, 0, 0, 0.04)',
            TEXT_STROKE: '#ffffff',
        },
    },

    // Typography
    TYPOGRAPHY: {
        CLUSTER_LABEL_SIZE: 16,
        CLUSTER_LABEL_WEIGHT: '700',
        NAMESPACE_LABEL_SIZE: 12,
        NAMESPACE_LABEL_WEIGHT: '600',
        NODE_LABEL_SIZE: 14,
        NODE_LABEL_WEIGHT: '500',
        EDGE_LABEL_SIZE: 12,
        EDGE_LABEL_WEIGHT: '900',
    },

    // Fitness function weights for namespace optimization
    FITNESS_WEIGHTS: {
        CROSSING_PENALTY: 100,
        SOURCE_POSITION_BONUS: 10,
        SINK_POSITION_BONUS: 10,
        EDGE_LENGTH_PENALTY: 5,
        FORWARD_FLOW_BONUS: 15,
    },

    // Optimization limits
    OPTIMIZATION: {
        MAX_PERMUTATIONS_BRUTE_FORCE: 4, // For namespaces <= 4, use brute force
        MAX_HILL_CLIMBING_ITERATIONS: 100,
    },
} as const;

/**
 * Get color based on error rate
 */
export const getErrorRateColor = (errorRate: number): string => {
    const { ERROR_RATE_THRESHOLDS, COLORS } = GRAPH_CONFIG;

    if (errorRate > ERROR_RATE_THRESHOLDS.HIGH) {
        return COLORS.ERROR_COLORS.HIGH;
    } else if (errorRate > ERROR_RATE_THRESHOLDS.MEDIUM) {
        return COLORS.ERROR_COLORS.MEDIUM;
    } else if (errorRate > ERROR_RATE_THRESHOLDS.LOW) {
        return COLORS.ERROR_COLORS.LOW;
    } else {
        return COLORS.ERROR_COLORS.HEALTHY;
    }
};

/**
 * Get color based on success rate
 */
export const getSuccessRateColor = (successRate: number): string => {
    const { SUCCESS_RATE_THRESHOLDS, COLORS } = GRAPH_CONFIG;

    if (successRate >= SUCCESS_RATE_THRESHOLDS.HIGH) {
        return COLORS.SUCCESS_COLORS.HIGH;
    } else if (successRate >= SUCCESS_RATE_THRESHOLDS.MEDIUM) {
        return COLORS.SUCCESS_COLORS.MEDIUM;
    } else {
        return COLORS.SUCCESS_COLORS.LOW;
    }
};

/**
 * Get edge width based on request rate
 */
export const getEdgeWidth = (requestRate: number): number => {
    const { MIN_EDGE_WIDTH, MAX_EDGE_WIDTH } = GRAPH_CONFIG;
    return Math.max(
        MIN_EDGE_WIDTH,
        Math.min(MAX_EDGE_WIDTH, requestRate * 2 + 1)
    );
};

/**
 * Get theme colors based on current theme
 */
export const getThemeColors = (isDark: boolean) => {
    return isDark
        ? GRAPH_CONFIG.COLORS.DARK_THEME
        : GRAPH_CONFIG.COLORS.LIGHT_THEME;
};

/**
 * Get cluster color from palette
 */
export const getClusterColor = (
    cluster: string | undefined,
    clusters: string[]
): string => {
    const { CLUSTER_PALETTE } = GRAPH_CONFIG.COLORS;

    if (!cluster) return CLUSTER_PALETTE[0];

    const index = clusters.indexOf(cluster);
    return CLUSTER_PALETTE[index % CLUSTER_PALETTE.length];
};
