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

import React, { useEffect, useRef, useState, useCallback } from 'react';
import * as d3 from 'd3';
import {
    GRAPH_CONFIG,
    getErrorRateColor,
    getSuccessRateColor,
    getEdgeWidth,
    getClusterColor,
} from '../utils/graphConfig';
import {
    NamespaceOptimizer,
    type GraphNode,
    type GraphEdge,
} from '../utils/namespaceOptimizer';
import {
    ForceSimulationManager,
    type D3Node,
    type D3Edge,
    type NamespaceData,
} from '../utils/forceSimulation';
import { PerformanceMonitor } from '../utils/performanceMonitor';
import { getThemeColors } from '../utils/graphConfig';

// Type definitions for better type safety
interface ThemeColors {
    BACKGROUND: string;
    FOREGROUND: string;
    NODE_FILL: string;
    NODE_STROKE: string;
    CLUSTER_BACKGROUND: string;
    NAMESPACE_BACKGROUND: string;
    TEXT_STROKE: string;
}

interface ClusterData {
    cluster: string;
    x: number;
    y: number;
    width: number;
    height: number;
    color: string;
    namespaces: NamespaceData[];
}

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

interface Edge {
    id: string;
    source: string;
    target: string;
    data?: {
        errorRate?: number;
        requestRate?: number;
    };
}

interface D3GraphRendererProps {
    nodes: Node[];
    edges: Edge[];
    onNodeClick?: (node: D3Node) => void;
    onCanvasClick?: () => void;
    className?: string;
}

/**
 * High-performance D3.js graph renderer with architecture diagram layout
 * and constrained physics simulation
 */
export const D3GraphRenderer: React.FC<D3GraphRendererProps> = ({
    nodes,
    edges,
    onNodeClick,
    onCanvasClick,
    className = '',
}) => {
    const svgRef = useRef<SVGSVGElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [dimensions, setDimensions] = useState({
        width: GRAPH_CONFIG.MIN_WIDTH,
        height: GRAPH_CONFIG.MIN_HEIGHT,
    });

    // Managers for complex operations
    const simulationManager = useRef(new ForceSimulationManager());
    const performanceMonitor = useRef(PerformanceMonitor.getInstance());
    const [themeVersion, setThemeVersion] = useState(0);

    // Optimize cluster color function with useCallback
    const getOptimizedClusterColor = useCallback(
        (cluster?: string) => {
            const clusters = [
                ...new Set(nodes.map((n) => n.cluster).filter(Boolean)),
            ];
            return getClusterColor(cluster, clusters);
        },
        [nodes]
    );

    // Initialize graph theme CSS and handle theme changes
    useEffect(() => {
        // Listen for theme changes
        const handleThemeChange = () => {
            setThemeVersion((prev) => prev + 1);
            simulationManager.current?.forceStabilityCheck();
        };

        // Monitor class changes on document element for theme updates
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
                    handleThemeChange();
                }
            });
        });

        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class']
        });

        return () => {
            observer?.disconnect();
        };
    }, []);

    // Handle container resize with debouncing
    useEffect(() => {
        let timeoutId: NodeJS.Timeout;

        const handleResize = () => {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => {
                if (containerRef.current) {
                    const rect = containerRef.current.getBoundingClientRect();
                    const width = Math.max(rect.width, GRAPH_CONFIG.MIN_WIDTH);
                    const height = Math.max(
                        rect.height,
                        GRAPH_CONFIG.MIN_HEIGHT
                    );
                    setDimensions({ width, height });
                }
            }, GRAPH_CONFIG.RESIZE_DEBOUNCE_MS);
        };

        // Initial size calculation
        const timer = setTimeout(handleResize, GRAPH_CONFIG.RESIZE_DEBOUNCE_MS);
        window.addEventListener('resize', handleResize);

        return () => {
            clearTimeout(timer);
            clearTimeout(timeoutId);
            window.removeEventListener('resize', handleResize);
        };
    }, []);

    // Main rendering effect
    useEffect(() => {
        if (!svgRef.current || nodes.length === 0) {
            return;
        }

        try {
            // Check browser support before rendering
            const browserSupport =
                performanceMonitor.current.checkBrowserSupport();
            if (!browserSupport.supported) {
                console.warn(
                    'Browser support issues:',
                    browserSupport.missingFeatures
                );
            }

            // Start performance monitoring
            performanceMonitor.current.startTiming('graph-render');

            renderGraph();

            // End performance monitoring and log results
            const renderTime =
                performanceMonitor.current.endTiming('graph-render');
            const metrics = performanceMonitor.current.analyzePerformance(
                nodes.length,
                edges.length,
                renderTime
            );
            performanceMonitor.current.logMetrics(metrics);
        } catch (error) {
            console.error('Error rendering graph:', error);

            // Enhanced error handling with fallback
            const svg = d3.select(svgRef.current);
            svg.selectAll('*').remove();

            const errorGroup = svg
                .append('g')
                .attr(
                    'transform',
                    `translate(${dimensions.width / 2}, ${dimensions.height / 2})`
                );

            errorGroup
                .append('text')
                .attr('text-anchor', 'middle')
                .attr('fill', 'red')
                .attr('font-size', '16px')
                .attr('font-weight', 'bold')
                .text('Graph Rendering Error');

            errorGroup
                .append('text')
                .attr('text-anchor', 'middle')
                .attr('fill', 'gray')
                .attr('font-size', '12px')
                .attr('y', 20)
                .text('Please check the console for details');

            // Re-throw to trigger error boundary
            throw error;
        }

        // Cleanup function
        return () => {
            simulationManager.current.stop();
        };
    }, [
        nodes,
        edges,
        dimensions,
        onNodeClick,
        onCanvasClick,
        getOptimizedClusterColor,
        themeVersion,
    ]);

    const renderGraph = () => {
        const svg = d3.select(svgRef.current!);
        svg.selectAll('*').remove();

        const { width, height } = dimensions;
        const isDark = document.documentElement.classList.contains('dark');
        const themeColors = getThemeColors(isDark);

        // Process data
        const { clusterData, d3Nodes, d3Edges } = processGraphData();

        // Set up SVG structure
        const container = svg.append('g');

        // Add zoom behavior
        const zoom = d3
            .zoom<SVGSVGElement, unknown>()
            .scaleExtent([0.1, 4])
            .on('zoom', (event) => {
                container.attr('transform', event.transform);
            });

        svg.call(zoom);

        // Add background click handler
        svg.append('rect')
            .attr('width', width)
            .attr('height', height)
            .attr('fill', 'transparent')
            .style('cursor', 'default')
            .on('click', (event) => {
                if (event.target === event.currentTarget) {
                    onCanvasClick?.();
                }
            });

        // Render architecture boxes
        renderArchitectureBoxes(container, clusterData, themeColors);

        // Render edges
        const { links, arrowTriangles, successLabels } = renderEdges(
            container,
            d3Edges,
            themeColors
        );

        // Render nodes
        const nodeElements = renderNodes(container, d3Nodes, themeColors);

        // Set up force simulation
        const namespaceData = clusterData.flatMap(
            (cluster) => cluster.namespaces
        );

        simulationManager.current.initialize(d3Nodes, d3Edges, namespaceData, {
            onTick: () =>
                updatePositions(
                    links,
                    arrowTriangles,
                    successLabels,
                    nodeElements
                ),
        });

        // Add drag behavior
        const dragBehavior = simulationManager.current.createDragBehavior({
            onTick: () =>
                updatePositions(
                    links,
                    arrowTriangles,
                    successLabels,
                    nodeElements
                ),
        });

        nodeElements.call(dragBehavior);
    };

    const processGraphData = () => {
        // Build cluster and namespace structure
        const clusters = [
            ...new Set(nodes.map((n) => n.data?.cluster).filter(Boolean)),
        ];

        const clusterData = clusters.map((cluster, clusterIndex) => {
            const clusterNodes = nodes.filter(
                (n) => n.data?.cluster === cluster
            );
            const namespaces = [
                ...new Set(
                    clusterNodes.map((n) => n.data?.namespace).filter(Boolean)
                ),
            ];

            // Optimize namespace ordering
            const optimizer = new NamespaceOptimizer(
                clusterNodes as GraphNode[],
                edges as GraphEdge[]
            );
            const optimizedNamespaces =
                optimizer.optimizeNamespaceOrder(namespaces);

            // Calculate layout positions
            const { width, height } = dimensions;
            const clusterWidth = width - GRAPH_CONFIG.CLUSTER_PADDING * 2;
            const clusterHeight = height - GRAPH_CONFIG.CLUSTER_PADDING * 2;

            const clusterX = GRAPH_CONFIG.CLUSTER_PADDING;
            const clusterY =
                GRAPH_CONFIG.CLUSTER_PADDING +
                clusterIndex * (clusterHeight + 60);

            const namespaceWidth =
                (clusterWidth -
                    (optimizedNamespaces.length + 1) *
                        GRAPH_CONFIG.NAMESPACE_PADDING) /
                optimizedNamespaces.length;
            const namespaceHeight =
                clusterHeight - GRAPH_CONFIG.NAMESPACE_HEIGHT_OFFSET;

            const namespaceData = optimizedNamespaces.map(
                (namespace, nsIndex) => {
                    const nsNodes = clusterNodes.filter(
                        (n) => n.data?.namespace === namespace
                    );
                    const nsX =
                        clusterX +
                        GRAPH_CONFIG.NAMESPACE_PADDING +
                        nsIndex *
                            (namespaceWidth + GRAPH_CONFIG.NAMESPACE_PADDING);
                    const nsY = clusterY + GRAPH_CONFIG.NAMESPACE_LABEL_HEIGHT;

                    // Sort nodes by traffic role
                    const sortedNodes = optimizer.sortNodesByTrafficRole(
                        nsNodes as GraphNode[]
                    );

                    // Position nodes within namespace
                    const nodePositions = sortedNodes.map((node, nodeIndex) => {
                        const role = optimizer.analyzeTrafficRole(node.id).role;
                        const nodesPerRow = Math.ceil(
                            Math.sqrt(nsNodes.length)
                        );
                        const row = Math.floor(nodeIndex / nodesPerRow);
                        const col = nodeIndex % nodesPerRow;

                        // Bias positioning based on traffic role
                        let xOffset = 0;
                        switch (role) {
                            case 'pure-source':
                                xOffset = 0;
                                break;
                            case 'intermediate':
                                xOffset = namespaceWidth * 0.3;
                                break;
                            case 'pure-sink':
                                xOffset = namespaceWidth * 0.7;
                                break;
                            default:
                                xOffset = namespaceWidth * 0.5;
                        }

                        const nodeSpacing = Math.min(
                            60,
                            (namespaceWidth * 0.25) /
                                Math.max(1, nodesPerRow - 1)
                        );

                        return {
                            ...node,
                            x: nsX + 20 + xOffset + col * nodeSpacing,
                            y: nsY + 40 + row * 80 + 40,
                            role,
                        } as D3Node;
                    });

                    return {
                        namespace,
                        x: nsX,
                        y: nsY,
                        width: namespaceWidth,
                        height: namespaceHeight,
                        nodes: nodePositions,
                    };
                }
            );

            return {
                cluster,
                x: clusterX,
                y: clusterY,
                width: clusterWidth,
                height: clusterHeight,
                namespaces: namespaceData,
            };
        });

        // Flatten for D3
        const d3Nodes: D3Node[] = clusterData.flatMap((cluster) =>
            cluster.namespaces.flatMap((namespace) => namespace.nodes)
        );

        const d3Edges: D3Edge[] = edges.map((edge) => ({
            ...edge,
            source: edge.source,
            target: edge.target,
        }));

        return { clusterData, d3Nodes, d3Edges };
    };

    const renderArchitectureBoxes = (
        container: d3.Selection<SVGGElement, unknown, null, undefined>,
        clusterData: ClusterData[],
        themeColors: ThemeColors
    ) => {
        const groupBoxes = container
            .append('g')
            .attr('class', 'architecture-boxes');

        clusterData.forEach((cluster, clusterIndex) => {
            const clusterColor = getOptimizedClusterColor(cluster.cluster);

            // Cluster box
            groupBoxes
                .append('rect')
                .attr('class', `cluster-box-${clusterIndex}`)
                .attr('x', cluster.x)
                .attr('y', cluster.y)
                .attr('width', cluster.width)
                .attr('height', cluster.height)
                .attr('rx', 8)
                .attr('ry', 8)
                .attr('fill', themeColors.CLUSTER_BACKGROUND)
                .attr('stroke', clusterColor)
                .attr('stroke-width', 3)
                .style('opacity', 1);

            // Cluster title background
            groupBoxes
                .append('rect')
                .attr('x', cluster.x)
                .attr('y', cluster.y)
                .attr('width', cluster.width)
                .attr('height', 35)
                .attr('rx', 8)
                .attr('ry', 8)
                .attr('fill', clusterColor)
                .style('opacity', 0.2);

            // Cluster label
            groupBoxes
                .append('text')
                .attr('x', cluster.x + 15)
                .attr('y', cluster.y + 22)
                .attr('fill', themeColors.FOREGROUND)
                .attr(
                    'font-size',
                    `${GRAPH_CONFIG.TYPOGRAPHY.CLUSTER_LABEL_SIZE}px`
                )
                .attr(
                    'font-weight',
                    GRAPH_CONFIG.TYPOGRAPHY.CLUSTER_LABEL_WEIGHT
                )
                .text(`Cluster: ${cluster.cluster}`);

            // Namespace boxes
            cluster.namespaces.forEach(
                (namespace: NamespaceData, nsIndex: number) => {
                    // Namespace box
                    groupBoxes
                        .append('rect')
                        .attr(
                            'class',
                            `namespace-box-${clusterIndex}-${nsIndex}`
                        )
                        .attr('x', namespace.x)
                        .attr('y', namespace.y)
                        .attr('width', namespace.width)
                        .attr('height', namespace.height)
                        .attr('rx', 6)
                        .attr('ry', 6)
                        .attr('fill', themeColors.NAMESPACE_BACKGROUND)
                        .attr('stroke', clusterColor)
                        .attr('stroke-width', 2)
                        .attr('stroke-dasharray', '8,4')
                        .style('opacity', 1);

                    // Namespace label background
                    groupBoxes
                        .append('rect')
                        .attr('x', namespace.x + 5)
                        .attr('y', namespace.y + 5)
                        .attr('width', namespace.namespace.length * 8 + 10)
                        .attr('height', 20)
                        .attr('rx', 3)
                        .attr('ry', 3)
                        .attr('fill', clusterColor)
                        .style('opacity', 0.2);

                    // Namespace label
                    groupBoxes
                        .append('text')
                        .attr('x', namespace.x + 10)
                        .attr('y', namespace.y + 18)
                        .attr('fill', themeColors.FOREGROUND)
                        .attr(
                            'font-size',
                            `${GRAPH_CONFIG.TYPOGRAPHY.NAMESPACE_LABEL_SIZE}px`
                        )
                        .attr(
                            'font-weight',
                            GRAPH_CONFIG.TYPOGRAPHY.NAMESPACE_LABEL_WEIGHT
                        )
                        .text(namespace.namespace);
                }
            );
        });
    };

    const renderEdges = (
        container: d3.Selection<SVGGElement, unknown, null, undefined>,
        d3Edges: D3Edge[],
        themeColors: ThemeColors
    ) => {
        const linkGroup = container.append('g').attr('class', 'links');

        const links = linkGroup
            .selectAll('line')
            .data(d3Edges)
            .enter()
            .append('line')
            .attr('stroke', (d: D3Edge) =>
                getErrorRateColor(d.data?.errorRate || 0)
            )
            .attr('stroke-width', (d: D3Edge) =>
                getEdgeWidth(d.data?.requestRate || 0)
            )
            .style('opacity', 0.8);

        const arrowTriangles = container
            .selectAll('.arrow-triangle')
            .data(d3Edges)
            .enter()
            .append('polygon')
            .attr('class', 'arrow-triangle')
            .attr('fill', (d: D3Edge) =>
                getErrorRateColor(d.data?.errorRate || 0)
            )
            .attr('stroke', (d: D3Edge) => {
                const baseColor = getErrorRateColor(d.data?.errorRate || 0);
                // Return darker version for outline
                const colorMap: Record<string, string> = {
                    '#ef4444': '#dc2626',
                    '#f59e0b': '#d97706',
                    '#eab308': '#ca8a04',
                    '#10b981': '#059669',
                };
                return colorMap[baseColor] || baseColor;
            })
            .attr('stroke-width', 1);

        const successLabels = linkGroup
            .selectAll('.success-label')
            .data(d3Edges)
            .enter()
            .append('text')
            .attr('class', 'success-label')
            .attr('text-anchor', 'middle')
            .attr('font-size', `${GRAPH_CONFIG.TYPOGRAPHY.EDGE_LABEL_SIZE}px`)
            .attr('font-weight', GRAPH_CONFIG.TYPOGRAPHY.EDGE_LABEL_WEIGHT)
            .attr('fill', (d: D3Edge) => {
                const errorRate = d.data?.errorRate || 0;
                const successRate = 1 - errorRate;
                return getSuccessRateColor(successRate);
            })
            .attr('stroke', themeColors.TEXT_STROKE)
            .attr('stroke-width', 0.5)
            .style('pointer-events', 'none')
            .text((d: D3Edge) => {
                const errorRate = d.data?.errorRate || 0;
                const successRate = 1 - errorRate;
                if (errorRate < 0.001 && successRate >= 0.999) return '100%';
                return `${(successRate * 100).toFixed(1)}%`;
            });

        return { links, arrowTriangles, successLabels };
    };

    const renderNodes = (
        container: d3.Selection<SVGGElement, unknown, null, undefined>,
        d3Nodes: D3Node[],
        themeColors: ThemeColors
    ) => {
        const nodeGroup = container.append('g').attr('class', 'nodes');

        const nodeElements = nodeGroup
            .selectAll('g')
            .data(d3Nodes)
            .enter()
            .append('g')
            .style('cursor', 'pointer')
            .on('click', (event, d) => {
                event.stopPropagation();
                onNodeClick?.(d);
            });

        // Add circles
        nodeElements
            .append('circle')
            .attr('r', GRAPH_CONFIG.NODE_RADIUS)
            .attr('fill', themeColors.NODE_FILL)
            .attr('stroke', themeColors.NODE_STROKE)
            .attr('stroke-width', GRAPH_CONFIG.NODE_STROKE_WIDTH)
            .style('transition', 'all 0.2s ease')
            .on('mouseenter', function () {
                d3.select(this).attr('r', GRAPH_CONFIG.NODE_HOVER_RADIUS);
            })
            .on('mouseleave', function () {
                d3.select(this).attr('r', GRAPH_CONFIG.NODE_RADIUS);
            });

        // Add labels
        nodeElements
            .append('text')
            .text((d) => d.label)
            .attr('fill', themeColors.FOREGROUND)
            .attr('text-anchor', 'middle')
            .attr('dy', 44)
            .attr('font-size', `${GRAPH_CONFIG.TYPOGRAPHY.NODE_LABEL_SIZE}px`)
            .attr('font-weight', GRAPH_CONFIG.TYPOGRAPHY.NODE_LABEL_WEIGHT)
            .style('pointer-events', 'none')
            .style('user-select', 'none');

        return nodeElements;
    };

    const updatePositions = (
        links: d3.Selection<SVGLineElement, D3Edge, SVGGElement, unknown>,
        arrowTriangles: d3.Selection<
            SVGPolygonElement,
            D3Edge,
            SVGGElement,
            unknown
        >,
        successLabels: d3.Selection<
            SVGTextElement,
            D3Edge,
            SVGGElement,
            unknown
        >,
        nodeElements: d3.Selection<SVGGElement, D3Node, SVGGElement, unknown>
    ) => {
        // Update edge positions (avoiding node overlap)
        links
            .attr('x1', (d: D3Edge) => {
                const dx = (d.target as D3Node).x! - (d.source as D3Node).x!;
                const dy = (d.target as D3Node).y! - (d.source as D3Node).y!;
                const length = Math.sqrt(dx * dx + dy * dy);
                if (length === 0) return (d.source as D3Node).x!;
                const sourceRadius =
                    GRAPH_CONFIG.NODE_RADIUS + GRAPH_CONFIG.EDGE_OFFSET;
                const unitX = dx / length;
                return (d.source as D3Node).x! + unitX * sourceRadius;
            })
            .attr('y1', (d: D3Edge) => {
                const dx = (d.target as D3Node).x! - (d.source as D3Node).x!;
                const dy = (d.target as D3Node).y! - (d.source as D3Node).y!;
                const length = Math.sqrt(dx * dx + dy * dy);
                if (length === 0) return (d.source as D3Node).y!;
                const sourceRadius =
                    GRAPH_CONFIG.NODE_RADIUS + GRAPH_CONFIG.EDGE_OFFSET;
                const unitY = dy / length;
                return (d.source as D3Node).y! + unitY * sourceRadius;
            })
            .attr('x2', (d: D3Edge) => {
                const dx = (d.target as D3Node).x! - (d.source as D3Node).x!;
                const dy = (d.target as D3Node).y! - (d.source as D3Node).y!;
                const length = Math.sqrt(dx * dx + dy * dy);
                if (length === 0) return (d.target as D3Node).x!;
                const targetRadius =
                    GRAPH_CONFIG.NODE_RADIUS + GRAPH_CONFIG.EDGE_OFFSET;
                const unitX = dx / length;
                return (d.target as D3Node).x! - unitX * targetRadius;
            })
            .attr('y2', (d: D3Edge) => {
                const dx = (d.target as D3Node).x! - (d.source as D3Node).x!;
                const dy = (d.target as D3Node).y! - (d.source as D3Node).y!;
                const length = Math.sqrt(dx * dx + dy * dy);
                if (length === 0) return (d.target as D3Node).y!;
                const targetRadius =
                    GRAPH_CONFIG.NODE_RADIUS + GRAPH_CONFIG.EDGE_OFFSET;
                const unitY = dy / length;
                return (d.target as D3Node).y! - unitY * targetRadius;
            });

        // Update arrow positions
        arrowTriangles.attr('points', (d: D3Edge) => {
            const dx = (d.target as D3Node).x! - (d.source as D3Node).x!;
            const dy = (d.target as D3Node).y! - (d.source as D3Node).y!;
            const length = Math.sqrt(dx * dx + dy * dy);

            if (length === 0) return '0,0 0,0 0,0';

            const targetNodeRadius =
                GRAPH_CONFIG.NODE_RADIUS + GRAPH_CONFIG.EDGE_OFFSET;
            const unitX = dx / length;
            const unitY = dy / length;
            const tipX = (d.target as D3Node).x! - unitX * targetNodeRadius;
            const tipY = (d.target as D3Node).y! - unitY * targetNodeRadius;

            const arrowSize = GRAPH_CONFIG.ARROW_SIZE;
            const leftX = tipX - (unitY * arrowSize) / 2 - unitX * arrowSize;
            const leftY = tipY + (unitX * arrowSize) / 2 - unitY * arrowSize;
            const rightX = tipX + (unitY * arrowSize) / 2 - unitX * arrowSize;
            const rightY = tipY - (unitX * arrowSize) / 2 - unitY * arrowSize;

            return `${tipX},${tipY} ${leftX},${leftY} ${rightX},${rightY}`;
        });

        // Update success rate labels
        successLabels
            .attr(
                'x',
                (d: D3Edge) =>
                    ((d.source as D3Node).x! + (d.target as D3Node).x!) / 2
            )
            .attr(
                'y',
                (d: D3Edge) =>
                    ((d.source as D3Node).y! + (d.target as D3Node).y!) / 2 + 4
            )
            .attr('transform', (d: D3Edge) => {
                const dx = (d.target as D3Node).x! - (d.source as D3Node).x!;
                const dy = (d.target as D3Node).y! - (d.source as D3Node).y!;
                const angle = (Math.atan2(dy, dx) * 180) / Math.PI;
                const midX =
                    ((d.source as D3Node).x! + (d.target as D3Node).x!) / 2;
                const midY =
                    ((d.source as D3Node).y! + (d.target as D3Node).y!) / 2 + 4;

                const adjustedAngle =
                    angle > 90 || angle < -90 ? angle + 180 : angle;
                return `rotate(${adjustedAngle}, ${midX}, ${midY})`;
            });

        // Update node positions
        nodeElements.attr(
            'transform',
            (d: D3Node) => `translate(${d.x},${d.y})`
        );
    };

    return (
        <div
            ref={containerRef}
            className={`w-full h-full bg-background ${className}`}
        >
            <svg
                ref={svgRef}
                width={dimensions.width}
                height={dimensions.height}
                className="w-full h-full"
            />
        </div>
    );
};
