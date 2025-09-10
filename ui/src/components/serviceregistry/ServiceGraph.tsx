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

import React, { useMemo, useCallback } from 'react';
import { Loader2, AlertCircle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useServiceGraphMetrics } from '../../hooks/useServiceGraphMetrics';
import { transformMetricsToGraph } from '../../utils/graphTransform';
import { D3GraphRenderer } from './D3GraphRenderer';
import { GraphErrorBoundary } from './GraphErrorBoundary';
import { GRAPH_CONFIG } from '../../utils/graphConfig';

interface Node {
    id: string;
    label: string;
    cluster?: string;
    size?: number;
    data?: {
        errorRate?: number;
        requestRate?: number;
        cluster?: string;
        namespace?: string;
    };
}

interface ServiceGraphProps {
    className?: string;
    onNodeClick?: (nodeId: string) => void;
    fullScreen?: boolean;
}

/**
 * Main ServiceGraph component - now much cleaner and focused on data management
 * Complex rendering logic has been extracted to specialized components
 */
export const ServiceGraph: React.FC<ServiceGraphProps> = ({
    className = '',
    onNodeClick,
    fullScreen = false,
}) => {
    // Memoize time range to prevent infinite re-renders
    const timeRange = useMemo(() => {
        const endTime = new Date();
        const startTime = new Date(endTime.getTime() - 60 * 60 * 1000); // 1 hour ago
        return {
            startTime: startTime.toISOString(),
            endTime: endTime.toISOString(),
        };
    }, []);

    // Fetch metrics data
    const {
        data: metrics,
        isLoading,
        error,
        isError,
    } = useServiceGraphMetrics(timeRange);

    // Transform metrics to graph data with error handling
    const graphData = useMemo(() => {
        if (!metrics) return { nodes: [], edges: [] };

        try {
            return transformMetricsToGraph(metrics);
        } catch (error) {
            console.error('Error transforming metrics to graph:', error);
            return { nodes: [], edges: [] };
        }
    }, [metrics]);

    // Optimized node click handler
    const handleNodeClick = useCallback(
        (node: Node) => {
            onNodeClick?.(node.id);
        },
        [onNodeClick]
    );

    // Canvas click handler for deselection
    const handleCanvasClick = useCallback(() => {
        // Future: Reset any selection state if needed
    }, []);

    // Performance check - warn if dataset is large
    const shouldShowPerformanceWarning = useMemo(() => {
        const nodeCount = graphData.nodes.length;
        const edgeCount = graphData.edges.length;
        return (
            nodeCount > GRAPH_CONFIG.MAX_NODES_BEFORE_VIRTUALIZATION ||
            edgeCount > GRAPH_CONFIG.MAX_NODES_BEFORE_VIRTUALIZATION * 2
        );
    }, [graphData]);

    // Loading state
    if (isLoading) {
        return (
            <Card className={`${className} border-0 shadow-md`}>
                <CardHeader>
                    <CardTitle>Service Graph</CardTitle>
                </CardHeader>
                <CardContent className="flex items-center justify-center py-12">
                    <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
                    <span className="ml-3 text-muted-foreground font-medium">
                        Loading service graph...
                    </span>
                </CardContent>
            </Card>
        );
    }

    // Error state
    if (isError) {
        return (
            <Card
                className={`${className} border-0 shadow-md border-red-100 bg-red-50 dark:border-red-900 dark:bg-red-950`}
            >
                <CardHeader>
                    <CardTitle>Service Graph</CardTitle>
                </CardHeader>
                <CardContent className="flex items-center justify-center py-12">
                    <AlertCircle className="w-8 h-8 text-red-500" />
                    <span className="ml-3 text-red-700 dark:text-red-400 font-medium">
                        Failed to load service graph:{' '}
                        {error?.message || 'Unknown error'}
                    </span>
                </CardContent>
            </Card>
        );
    }

    // Empty state
    if (!metrics || metrics.length === 0) {
        return (
            <Card className={`${className} border-0 shadow-md`}>
                <CardHeader>
                    <CardTitle>Service Graph</CardTitle>
                </CardHeader>
                <CardContent className="text-center py-12">
                    <AlertCircle className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-foreground mb-2">
                        No service communication found
                    </h3>
                    <p className="text-muted-foreground">
                        No service-to-service communication metrics available
                        for the selected time period.
                    </p>
                </CardContent>
            </Card>
        );
    }

    // Full screen mode (for topology page)
    if (fullScreen) {
        return (
            <div className={`${className} h-full`}>
                {shouldShowPerformanceWarning && (
                    <div className="absolute top-4 left-4 z-10">
                        <Badge variant="destructive" className="text-xs">
                            Large dataset ({graphData.nodes.length} nodes) -
                            performance may be affected
                        </Badge>
                    </div>
                )}
                <div className="h-full w-full relative overflow-hidden">
                    <GraphErrorBoundary fallbackClassName="h-full">
                        <D3GraphRenderer
                            nodes={graphData.nodes}
                            edges={graphData.edges}
                            onNodeClick={handleNodeClick}
                            onCanvasClick={handleCanvasClick}
                        />
                    </GraphErrorBoundary>
                </div>
            </div>
        );
    }

    // Card mode (for dashboard/embedded use)
    return (
        <Card className={`${className} border-0 shadow-md`}>
            <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                    <CardTitle>Service Graph</CardTitle>
                    <div className="flex items-center gap-2">
                        <Badge variant="secondary" className="text-xs">
                            {graphData.nodes.length} services
                        </Badge>
                        <Badge variant="secondary" className="text-xs">
                            {graphData.edges.length} connections
                        </Badge>
                        {shouldShowPerformanceWarning && (
                            <Badge variant="destructive" className="text-xs">
                                Large dataset
                            </Badge>
                        )}
                    </div>
                </div>
            </CardHeader>
            <CardContent className="p-0">
                <div className="h-96 w-full relative overflow-hidden">
                    <GraphErrorBoundary>
                        <D3GraphRenderer
                            nodes={graphData.nodes}
                            edges={graphData.edges}
                            onNodeClick={handleNodeClick}
                            onCanvasClick={handleCanvasClick}
                        />
                    </GraphErrorBoundary>
                </div>
            </CardContent>
        </Card>
    );
};
