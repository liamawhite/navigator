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

import { GRAPH_CONFIG } from './graphConfig';

export interface PerformanceMetrics {
    renderTime: number;
    nodeCount: number;
    edgeCount: number;
    memoryUsage?: number;
    isLargeDataset: boolean;
    warnings: string[];
}

/**
 * Performance monitoring utilities for graph visualization
 */
export class PerformanceMonitor {
    private static instance: PerformanceMonitor;
    private renderTimings: Map<string, number> = new Map();
    private memoryBaseline?: number;

    private constructor() {
        // Singleton pattern
        if (typeof window !== 'undefined' && 'performance' in window) {
            this.memoryBaseline = this.getCurrentMemoryUsage();
        }
    }

    public static getInstance(): PerformanceMonitor {
        if (!PerformanceMonitor.instance) {
            PerformanceMonitor.instance = new PerformanceMonitor();
        }
        return PerformanceMonitor.instance;
    }

    /**
     * Start timing a render operation
     */
    public startTiming(operationId: string): void {
        if (typeof window !== 'undefined' && 'performance' in window) {
            this.renderTimings.set(operationId, performance.now());
        }
    }

    /**
     * End timing and return elapsed time
     */
    public endTiming(operationId: string): number {
        if (typeof window === 'undefined' || !('performance' in window)) {
            return 0;
        }

        const startTime = this.renderTimings.get(operationId);
        if (!startTime) {
            console.warn(`No start time found for operation: ${operationId}`);
            return 0;
        }

        const elapsed = performance.now() - startTime;
        this.renderTimings.delete(operationId);
        return elapsed;
    }

    /**
     * Get current memory usage (if available)
     */
    private getCurrentMemoryUsage(): number | undefined {
        if (typeof window === 'undefined') return undefined;

        // @ts-expect-error - performance.memory is not in all TypeScript definitions
        const memory = (performance as { memory?: { usedJSHeapSize: number } })
            .memory;
        if (memory && memory.usedJSHeapSize) {
            return memory.usedJSHeapSize / 1024 / 1024; // Convert to MB
        }
        return undefined;
    }

    /**
     * Analyze performance metrics and generate warnings
     */
    public analyzePerformance(
        nodeCount: number,
        edgeCount: number,
        renderTime: number
    ): PerformanceMetrics {
        const warnings: string[] = [];
        const isLargeDataset =
            nodeCount > GRAPH_CONFIG.MAX_NODES_BEFORE_VIRTUALIZATION;

        // Check dataset size
        if (isLargeDataset) {
            warnings.push(
                `Large dataset detected (${nodeCount} nodes). Consider implementing virtualization.`
            );
        }

        // Check render time
        if (renderTime > 1000) {
            warnings.push(
                `Slow render time (${renderTime.toFixed(0)}ms). Consider optimizing visualization.`
            );
        }

        // Check edge density
        const edgeDensity = edgeCount / nodeCount;
        if (edgeDensity > 5) {
            warnings.push(
                `High edge density (${edgeDensity.toFixed(1)} edges per node). Consider edge bundling.`
            );
        }

        // Check memory usage
        const currentMemory = this.getCurrentMemoryUsage();
        let memoryIncrease: number | undefined;

        if (currentMemory && this.memoryBaseline) {
            memoryIncrease = currentMemory - this.memoryBaseline;
            if (memoryIncrease > 50) {
                // 50MB increase
                warnings.push(
                    `High memory usage increase (${memoryIncrease.toFixed(1)}MB). Monitor for memory leaks.`
                );
            }
        }

        return {
            renderTime,
            nodeCount,
            edgeCount,
            memoryUsage: currentMemory,
            isLargeDataset,
            warnings,
        };
    }

    /**
     * Log performance metrics to console (development only)
     */
    public logMetrics(
        metrics: PerformanceMetrics,
        operationName: string = 'Graph Render'
    ): void {
        if (process.env.NODE_ENV !== 'development') return;

        console.group(`📊 ${operationName} Performance`);
        console.log(`⏱️  Render Time: ${metrics.renderTime.toFixed(2)}ms`);
        console.log(
            `🔗 Nodes: ${metrics.nodeCount}, Edges: ${metrics.edgeCount}`
        );

        if (metrics.memoryUsage) {
            console.log(`💾 Memory Usage: ${metrics.memoryUsage.toFixed(1)}MB`);
        }

        if (metrics.warnings.length > 0) {
            console.warn('⚠️  Performance Warnings:');
            metrics.warnings.forEach((warning) =>
                console.warn(`   • ${warning}`)
            );
        }

        if (metrics.isLargeDataset) {
            console.info('💡 Optimization Suggestions:');
            console.info('   • Implement virtualization for large datasets');
            console.info('   • Consider edge bundling for dense graphs');
            console.info('   • Use level-of-detail rendering');
        }

        console.groupEnd();
    }

    /**
     * Check if browser supports required features
     */
    public checkBrowserSupport(): {
        supported: boolean;
        missingFeatures: string[];
    } {
        const missingFeatures: string[] = [];

        if (typeof window === 'undefined') {
            return {
                supported: false,
                missingFeatures: ['DOM not available (SSR)'],
            };
        }

        // Check for SVG support
        if (
            !document.createElementNS ||
            !document.createElementNS('http://www.w3.org/2000/svg', 'svg')
        ) {
            missingFeatures.push('SVG support');
        }

        // Check for D3 requirements
        if (!Array.from) {
            missingFeatures.push('Array.from (ES6)');
        }

        if (!Map) {
            missingFeatures.push('Map (ES6)');
        }

        if (!Set) {
            missingFeatures.push('Set (ES6)');
        }

        // Check for performance API
        if (!performance || !performance.now) {
            missingFeatures.push(
                'Performance API (timing measurements will be inaccurate)'
            );
        }

        return {
            supported: missingFeatures.length === 0,
            missingFeatures,
        };
    }

    /**
     * Get recommendations for large datasets
     */
    public getOptimizationRecommendations(
        nodeCount: number,
        edgeCount: number
    ): string[] {
        const recommendations: string[] = [];

        if (nodeCount > GRAPH_CONFIG.MAX_NODES_BEFORE_VIRTUALIZATION) {
            recommendations.push(
                'Consider implementing virtualization to render only visible nodes'
            );
            recommendations.push(
                'Use clustering/grouping to reduce visual complexity'
            );
        }

        if (edgeCount > nodeCount * 3) {
            recommendations.push(
                'Consider edge bundling or filtering to reduce visual clutter'
            );
            recommendations.push('Implement edge sampling for dense graphs');
        }

        if (nodeCount > 50) {
            recommendations.push(
                'Use level-of-detail rendering (show less detail when zoomed out)'
            );
            recommendations.push(
                'Implement progressive rendering (render in chunks)'
            );
        }

        return recommendations;
    }

    /**
     * Reset memory baseline (useful after major operations)
     */
    public resetMemoryBaseline(): void {
        this.memoryBaseline = this.getCurrentMemoryUsage();
    }

    /**
     * Force garbage collection if available (Chrome DevTools)
     */
    public forceGarbageCollection(): void {
        if (typeof window !== 'undefined') {
            // @ts-expect-error - gc is available in Chrome DevTools
            if ((window as { gc?: () => void }).gc) {
                // @ts-expect-error - window.gc is only available in Chrome DevTools
                (window as { gc: () => void }).gc();
                console.log('🗑️  Forced garbage collection');
            }
        }
    }
}
