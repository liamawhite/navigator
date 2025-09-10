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

import React, { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import { Loader2, AlertCircle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useServiceGraphMetrics } from '../../hooks/useServiceGraphMetrics';
import { transformMetricsToGraph } from '../../utils/graphTransform';

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

interface ServiceGraphProps {
    className?: string;
    onNodeClick?: (nodeId: string) => void;
    fullScreen?: boolean;
}

interface D3Node extends d3.SimulationNodeDatum {
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

interface D3Edge extends d3.SimulationLinkDatum<D3Node> {
    id: string;
    data?: {
        errorRate?: number;
        requestRate?: number;
    };
}

// D3 Graph Component
const D3Graph: React.FC<{
    nodes: Node[];
    edges: Edge[];
    onNodeClick?: (node: D3Node) => void;
    onCanvasClick?: () => void;
    className?: string;
}> = ({ nodes, edges, onNodeClick, onCanvasClick, className = '' }) => {
    const svgRef = useRef<SVGSVGElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [dimensions, setDimensions] = useState({ width: 800, height: 600 });

    // Color palette for clusters - using slate-based colors
    const getClusterColor = React.useCallback(
        (cluster?: string) => {
            const colors = [
                '#64748b', // slate-500
                '#6b7280', // gray-500
                '#71717a', // zinc-500
                '#737373', // neutral-500
                '#78716c', // stone-500
                '#475569', // slate-600
                '#4b5563', // gray-600
                '#52525b', // zinc-600
            ];
            if (!cluster) return colors[0];
            const clusters = [
                ...new Set(nodes.map((n) => n.cluster).filter(Boolean)),
            ];
            const index = clusters.indexOf(cluster);
            return colors[index % colors.length];
        },
        [nodes]
    );

    // Handle container resize
    useEffect(() => {
        const handleResize = () => {
            if (containerRef.current) {
                const rect = containerRef.current.getBoundingClientRect();
                const width = Math.max(rect.width, 800); // Minimum 800px width
                const height = Math.max(rect.height, 600); // Minimum 600px height
                setDimensions({ width, height });
            }
        };

        // Use setTimeout to ensure DOM is ready
        const timer = setTimeout(handleResize, 100);
        window.addEventListener('resize', handleResize);
        return () => {
            clearTimeout(timer);
            window.removeEventListener('resize', handleResize);
        };
    }, []);

    // Main D3 rendering effect
    useEffect(() => {
        if (!svgRef.current || nodes.length === 0) return;

        const svg = d3.select(svgRef.current);
        svg.selectAll('*').remove(); // Clear previous render

        // Define arrow markers first before any other elements
        const defs = svg.append('defs');

        // Create multiple arrow markers with different approaches
        defs.append('marker')
            .attr('id', 'arrowhead')
            .attr('viewBox', '0 0 20 20')
            .attr('refX', 18)
            .attr('refY', 10)
            .attr('orient', 'auto')
            .attr('markerWidth', 12)
            .attr('markerHeight', 12)
            .append('polygon')
            .attr('points', '0,5 0,15 20,10')
            .attr('fill', '#10b981')
            .attr('stroke', '#10b981')
            .attr('stroke-width', 1);

        // Alternative triangle marker
        defs.append('marker')
            .attr('id', 'triangle')
            .attr('viewBox', '0 0 10 10')
            .attr('refX', 8)
            .attr('refY', 5)
            .attr('orient', 'auto')
            .attr('markerWidth', 8)
            .attr('markerHeight', 8)
            .append('path')
            .attr('d', 'M 0 0 L 10 5 L 0 10 z')
            .attr('fill', '#10b981');

        // HUGE Circle marker for debugging
        defs.append('marker')
            .attr('id', 'circle')
            .attr('viewBox', '0 0 40 40')
            .attr('refX', 20)
            .attr('refY', 20)
            .attr('orient', 'auto')
            .attr('markerWidth', 30) // Huge marker
            .attr('markerHeight', 30) // Huge marker
            .append('circle')
            .attr('cx', 20)
            .attr('cy', 20)
            .attr('r', 15) // Huge radius
            .attr('fill', '#ef4444') // Bright red
            .attr('stroke', '#000000') // Black outline
            .attr('stroke-width', 2);

        const { width, height } = dimensions;

        // Create architecture diagram-style layout with literal boxes
        const clusters = [
            ...new Set(nodes.map((n) => n.data?.cluster).filter(Boolean)),
        ];

        // Calculate fixed layout positions like an architecture diagram
        const padding = 50;
        const clusterWidth = width - padding * 2;
        const clusterHeight = height - padding * 2;

        // Build cluster and namespace structure
        const clusterData = clusters.map((cluster, clusterIndex) => {
            const clusterNodes = nodes.filter(
                (n) => n.data?.cluster === cluster
            );
            const namespaces = [
                ...new Set(
                    clusterNodes.map((n) => n.data?.namespace).filter(Boolean)
                ),
            ];

            // Fitness function for optimal namespace placement
            const calculateNamespaceFitness = (namespaceOrder: string[]) => {
                let score = 0;

                // Factor 1: Minimize edge crossings (highest weight)
                let crossings = 0;
                for (let i = 0; i < edges.length; i++) {
                    for (let j = i + 1; j < edges.length; j++) {
                        const edge1 = edges[i];
                        const edge2 = edges[j];

                        // Get namespace positions for each edge's source and target
                        const e1SourceNs = nodes.find(
                            (n) => n.id === edge1.source
                        )?.data?.namespace;
                        const e1TargetNs = nodes.find(
                            (n) => n.id === edge1.target
                        )?.data?.namespace;
                        const e2SourceNs = nodes.find(
                            (n) => n.id === edge2.source
                        )?.data?.namespace;
                        const e2TargetNs = nodes.find(
                            (n) => n.id === edge2.target
                        )?.data?.namespace;

                        if (
                            e1SourceNs &&
                            e1TargetNs &&
                            e2SourceNs &&
                            e2TargetNs
                        ) {
                            const e1SourcePos =
                                namespaceOrder.indexOf(e1SourceNs);
                            const e1TargetPos =
                                namespaceOrder.indexOf(e1TargetNs);
                            const e2SourcePos =
                                namespaceOrder.indexOf(e2SourceNs);
                            const e2TargetPos =
                                namespaceOrder.indexOf(e2TargetNs);

                            // Check if edges cross (one goes left-to-right while other goes right-to-left in same span)
                            if (
                                (e1SourcePos < e1TargetPos &&
                                    e2SourcePos > e2TargetPos &&
                                    e1SourcePos < e2SourcePos &&
                                    e1TargetPos > e2TargetPos) ||
                                (e1SourcePos > e1TargetPos &&
                                    e2SourcePos < e2TargetPos &&
                                    e2SourcePos < e1SourcePos &&
                                    e2TargetPos > e1TargetPos)
                            ) {
                                crossings++;
                            }
                        }
                    }
                }
                score -= crossings * 100; // Heavy penalty for crossings

                // Factor 2: Reward left-to-right flow (pure sources on left, sinks on right)
                namespaceOrder.forEach((namespace, index) => {
                    const nsNodes = clusterNodes.filter(
                        (n) => n.data?.namespace === namespace
                    );
                    let sourceCount = 0;
                    let sinkCount = 0;

                    nsNodes.forEach((node) => {
                        const isSource = edges.some(
                            (edge) => edge.source === node.id
                        );
                        const isSink = edges.some(
                            (edge) => edge.target === node.id
                        );

                        if (isSource && !isSink) sourceCount++;
                        else if (!isSource && isSink) sinkCount++;
                    });

                    // Reward pure sources on the left (lower indices)
                    if (sourceCount > 0 && sinkCount === 0) {
                        score += (namespaceOrder.length - index) * 10;
                    }
                    // Reward pure sinks on the right (higher indices)
                    if (sinkCount > 0 && sourceCount === 0) {
                        score += index * 10;
                    }
                });

                // Factor 3: Minimize total edge length
                let totalDistance = 0;
                edges.forEach((edge) => {
                    const sourceNs = nodes.find((n) => n.id === edge.source)
                        ?.data?.namespace;
                    const targetNs = nodes.find((n) => n.id === edge.target)
                        ?.data?.namespace;

                    if (sourceNs && targetNs) {
                        const sourcePos = namespaceOrder.indexOf(sourceNs);
                        const targetPos = namespaceOrder.indexOf(targetNs);
                        totalDistance += Math.abs(targetPos - sourcePos);
                    }
                });
                score -= totalDistance * 5; // Penalty for long edges

                // Factor 4: Reward forward flow (left-to-right direction)
                let forwardFlows = 0;
                edges.forEach((edge) => {
                    const sourceNs = nodes.find((n) => n.id === edge.source)
                        ?.data?.namespace;
                    const targetNs = nodes.find((n) => n.id === edge.target)
                        ?.data?.namespace;

                    if (sourceNs && targetNs) {
                        const sourcePos = namespaceOrder.indexOf(sourceNs);
                        const targetPos = namespaceOrder.indexOf(targetNs);

                        if (targetPos > sourcePos) {
                            // Forward flow
                            forwardFlows++;
                        }
                    }
                });
                score += forwardFlows * 15; // Reward forward-flowing edges

                return score;
            };

            // Find optimal namespace ordering using simple optimization
            const optimizeNamespaceOrder = (namespaces: string[]) => {
                let bestOrder = [...namespaces];
                let bestScore = calculateNamespaceFitness(bestOrder);

                // Try all permutations for small namespace counts (brute force)
                if (namespaces.length <= 4) {
                    const permute = (arr: string[]): string[][] => {
                        if (arr.length <= 1) return [arr];
                        const result = [];
                        for (let i = 0; i < arr.length; i++) {
                            const rest = [
                                ...arr.slice(0, i),
                                ...arr.slice(i + 1),
                            ];
                            const perms = permute(rest);
                            for (const perm of perms) {
                                result.push([arr[i], ...perm]);
                            }
                        }
                        return result;
                    };

                    const allOrders = permute(namespaces);
                    for (const order of allOrders) {
                        const score = calculateNamespaceFitness(order);
                        if (score > bestScore) {
                            bestScore = score;
                            bestOrder = order;
                        }
                    }
                } else {
                    // For larger sets, use hill climbing optimization
                    for (let iteration = 0; iteration < 100; iteration++) {
                        let improved = false;
                        for (let i = 0; i < bestOrder.length - 1; i++) {
                            // Try swapping adjacent namespaces
                            const testOrder = [...bestOrder];
                            [testOrder[i], testOrder[i + 1]] = [
                                testOrder[i + 1],
                                testOrder[i],
                            ];

                            const score = calculateNamespaceFitness(testOrder);
                            if (score > bestScore) {
                                bestScore = score;
                                bestOrder = testOrder;
                                improved = true;
                            }
                        }
                        if (!improved) break; // Local optimum reached
                    }
                }

                return bestOrder;
            };

            // Apply fitness-based optimization to namespace ordering
            const optimizedNamespaces = optimizeNamespaceOrder(namespaces);

            // Position cluster box
            const clusterX = padding;
            const clusterY = padding + clusterIndex * (clusterHeight + 60);

            // Calculate namespace boxes within cluster
            const namespaceWidth =
                (clusterWidth - (optimizedNamespaces.length + 1) * 20) /
                optimizedNamespaces.length;
            const namespaceHeight = clusterHeight - 60; // Leave room for cluster label

            const namespaceData = optimizedNamespaces.map(
                (namespace, nsIndex) => {
                    const nsNodes = clusterNodes.filter(
                        (n) => n.data?.namespace === namespace
                    );
                    const nsX = clusterX + 20 + nsIndex * (namespaceWidth + 20);
                    const nsY = clusterY + 40; // Leave room for cluster label

                    // Analyze traffic flow to determine service positioning
                    const analyzeTrafficRole = (nodeId: string) => {
                        const isSource = edges.some(
                            (edge) => edge.source === nodeId
                        );
                        const isDestination = edges.some(
                            (edge) => edge.target === nodeId
                        );

                        if (isSource && !isDestination) return 'pure-source'; // Only sends traffic
                        if (!isSource && isDestination) return 'pure-sink'; // Only receives traffic
                        if (isSource && isDestination) return 'intermediate'; // Both sends and receives
                        return 'isolated'; // No traffic connections
                    };

                    // Sort nodes by traffic role for left-to-right positioning
                    const sortedNodes = [...nsNodes].sort((a, b) => {
                        const roleOrder = {
                            'pure-source': 0,
                            intermediate: 1,
                            'pure-sink': 2,
                            isolated: 3,
                        };
                        const roleA = analyzeTrafficRole(a.id);
                        const roleB = analyzeTrafficRole(b.id);
                        return roleOrder[roleA] - roleOrder[roleB];
                    });

                    // Position nodes in flow-based layout within namespace
                    const nodePositions = sortedNodes.map((node, nodeIndex) => {
                        const role = analyzeTrafficRole(node.id);
                        const nodesPerRow = Math.ceil(
                            Math.sqrt(nsNodes.length)
                        );
                        const row = Math.floor(nodeIndex / nodesPerRow);
                        const col = nodeIndex % nodesPerRow;

                        // Bias positioning based on traffic role
                        let xOffset = 0;
                        if (role === 'pure-source') {
                            xOffset = 0; // Far left
                        } else if (role === 'intermediate') {
                            xOffset = namespaceWidth * 0.3; // Middle-left
                        } else if (role === 'pure-sink') {
                            xOffset = namespaceWidth * 0.7; // Right side
                        } else {
                            xOffset = namespaceWidth * 0.5; // Center for isolated nodes
                        }

                        const nodeSpacing = Math.min(
                            60,
                            (namespaceWidth * 0.25) /
                                Math.max(1, nodesPerRow - 1)
                        );

                        return {
                            ...node,
                            x: nsX + 20 + xOffset + col * nodeSpacing,
                            y: nsY + 40 + row * 80 + 40, // Leave room for namespace label
                            role, // Store role for potential styling
                        };
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

        // Flatten node positions for D3
        const d3Nodes: D3Node[] = clusterData.flatMap((cluster) =>
            cluster.namespaces.flatMap((namespace) => namespace.nodes)
        );

        const d3Edges: D3Edge[] = edges.map((edge) => ({
            ...edge,
            source: edge.source,
            target: edge.target,
        }));

        // Enable constrained physics simulation - nodes can move within namespace boundaries
        const simulation = d3
            .forceSimulation<D3Node>(d3Nodes)
            .force(
                'link',
                d3
                    .forceLink<D3Node, D3Edge>(d3Edges)
                    .id((d) => d.id)
                    .distance(100)
            )
            .force('charge', d3.forceManyBody().strength(-200))
            .force('collision', d3.forceCollide().radius(30));

        // Create container groups
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

        // Draw literal architecture diagram boxes
        const isDark = document.documentElement.classList.contains('dark');
        const groupBoxes = container
            .append('g')
            .attr('class', 'architecture-boxes');

        // Draw cluster boxes first (background) - much more visible
        clusterData.forEach((cluster, clusterIndex) => {
            const clusterColor = getClusterColor(cluster.cluster);

            // Cluster box with strong, visible border
            groupBoxes
                .append('rect')
                .attr('class', `cluster-box-${clusterIndex}`)
                .attr('x', cluster.x)
                .attr('y', cluster.y)
                .attr('width', cluster.width)
                .attr('height', cluster.height)
                .attr('rx', 8)
                .attr('ry', 8)
                .attr(
                    'fill',
                    isDark
                        ? 'rgba(71, 85, 105, 0.15)'
                        : 'rgba(100, 116, 139, 0.08)'
                ) // Slate background
                .attr('stroke', clusterColor)
                .attr('stroke-width', 3) // Thicker border
                .style('opacity', 1); // Full opacity

            // Cluster title background - more prominent
            groupBoxes
                .append('rect')
                .attr('x', cluster.x)
                .attr('y', cluster.y)
                .attr('width', cluster.width)
                .attr('height', 35)
                .attr('rx', 8)
                .attr('ry', 8)
                .attr('fill', clusterColor)
                .style('opacity', 0.2); // More visible

            // Cluster label - white text on colored background
            groupBoxes
                .append('text')
                .attr('x', cluster.x + 15)
                .attr('y', cluster.y + 22)
                .attr('fill', isDark ? '#ffffff' : '#000000')
                .attr('font-size', '16px')
                .attr('font-weight', '700') // Bolder
                .text(`Cluster: ${cluster.cluster}`);

            // Draw namespace boxes within cluster - much more visible
            cluster.namespaces.forEach((namespace, nsIndex) => {
                // Use subtle, consistent namespace colors based on cluster color
                const namespaceColor = clusterColor; // Keep consistent with cluster

                // Namespace box with strong border
                groupBoxes
                    .append('rect')
                    .attr('class', `namespace-box-${clusterIndex}-${nsIndex}`)
                    .attr('x', namespace.x)
                    .attr('y', namespace.y)
                    .attr('width', namespace.width)
                    .attr('height', namespace.height)
                    .attr('rx', 6)
                    .attr('ry', 6)
                    .attr(
                        'fill',
                        isDark
                            ? 'rgba(255, 255, 255, 0.08)'
                            : 'rgba(0, 0, 0, 0.04)'
                    ) // More visible background
                    .attr('stroke', namespaceColor)
                    .attr('stroke-width', 2) // Thicker border
                    .attr('stroke-dasharray', '8,4') // Longer dashes
                    .style('opacity', 1); // Full opacity

                // Namespace label background for better readability
                groupBoxes
                    .append('rect')
                    .attr('x', namespace.x + 5)
                    .attr('y', namespace.y + 5)
                    .attr('width', namespace.namespace.length * 8 + 10)
                    .attr('height', 20)
                    .attr('rx', 3)
                    .attr('ry', 3)
                    .attr('fill', namespaceColor)
                    .style('opacity', 0.2);

                // Namespace label - better contrast
                groupBoxes
                    .append('text')
                    .attr('x', namespace.x + 10)
                    .attr('y', namespace.y + 18)
                    .attr('fill', isDark ? '#ffffff' : '#000000')
                    .attr('font-size', '12px')
                    .attr('font-weight', '600') // Bolder
                    .text(`${namespace.namespace}`);
            });
        });

        // Draw edges with error-rate-based coloring
        const linkGroup = container.append('g').attr('class', 'links');
        const links = linkGroup
            .selectAll('line')
            .data(d3Edges)
            .enter()
            .append('line')
            .attr('stroke', (d: D3Edge) => {
                // Color edges based on error rate from the edge data
                const errorRate = d.data?.errorRate || 0;
                if (errorRate > 0.1) {
                    return '#ef4444'; // Red for high error rate (>10%)
                } else if (errorRate > 0.05) {
                    return '#f59e0b'; // Orange/amber for medium error rate (>5%)
                } else if (errorRate > 0.01) {
                    return '#eab308'; // Yellow for low error rate (>1%)
                } else {
                    return '#10b981'; // Green for healthy traffic (≤1%)
                }
            })
            .attr('stroke-width', (d: D3Edge) => {
                // Thicker lines for higher traffic volume
                const requestRate = d.data?.requestRate || 0;
                return Math.max(1, Math.min(6, requestRate * 2 + 1));
            })
            .style('opacity', 0.8);

        // Create arrow triangles for each edge - positioned along the line, not at nodes
        const arrowTriangles = container
            .selectAll('.arrow-triangle')
            .data(d3Edges)
            .enter()
            .append('polygon')
            .attr('class', 'arrow-triangle')
            .attr('fill', (d: D3Edge) => {
                // Match arrow color with edge color based on error rate
                const errorRate = d.data?.errorRate || 0;
                if (errorRate > 0.1) {
                    return '#ef4444'; // Red for high error rate (>10%)
                } else if (errorRate > 0.05) {
                    return '#f59e0b'; // Orange/amber for medium error rate (>5%)
                } else if (errorRate > 0.01) {
                    return '#eab308'; // Yellow for low error rate (>1%)
                } else {
                    return '#10b981'; // Green for healthy traffic (≤1%)
                }
            })
            .attr('stroke', (d: D3Edge) => {
                // Darker version of fill color for outline
                const errorRate = d.data?.errorRate || 0;
                if (errorRate > 0.1) {
                    return '#dc2626'; // Darker red
                } else if (errorRate > 0.05) {
                    return '#d97706'; // Darker orange
                } else if (errorRate > 0.01) {
                    return '#ca8a04'; // Darker yellow
                } else {
                    return '#059669'; // Darker green
                }
            })
            .attr('stroke-width', 1);

        // Add success rate labels along the edge lines (like Kiali)
        const successLabels = linkGroup
            .selectAll('.success-label')
            .data(d3Edges)
            .enter()
            .append('text')
            .attr('class', 'success-label')
            .attr('text-anchor', 'middle')
            .attr('font-size', '12px')
            .attr('font-weight', '900')
            .attr('fill', (d: D3Edge) => {
                const errorRate = d.data?.errorRate || 0;
                const successRate = 1 - errorRate; // Convert error rate to success rate
                if (successRate >= 0.99) return '#10b981'; // Green for high success
                if (successRate >= 0.95) return '#f59e0b'; // Orange for medium success
                return '#ef4444'; // Red for low success
            })
            .attr('stroke', (_d: D3Edge) => {
                // Dark outline for contrast
                return document.documentElement.classList.contains('dark')
                    ? '#000000'
                    : '#ffffff';
            })
            .attr('stroke-width', 0.5)
            .style('pointer-events', 'none')
            .text((d: D3Edge) => {
                const errorRate = d.data?.errorRate || 0;
                const successRate = 1 - errorRate; // Convert error rate to success rate
                if (errorRate < 0.001 && successRate >= 0.999) return '100%'; // Show 100% for very low error rates
                return `${(successRate * 100).toFixed(1)}%`;
            });

        // Draw nodes
        const nodeGroup = container.append('g').attr('class', 'nodes');
        const nodeElements = nodeGroup
            .selectAll('g')
            .data(d3Nodes)
            .enter()
            .append('g')
            .style('cursor', 'pointer')
            .call(
                d3
                    .drag<SVGGElement, D3Node>()
                    .on('start', (event, _d) => {
                        if (!event.active)
                            simulation.alphaTarget(0.3).restart();
                        d.fx = d.x;
                        d.fy = d.y;
                    })
                    .on('drag', (event, d) => {
                        // Find the namespace boundaries for this node
                        const nodeNamespace = clusterData
                            .flatMap((c) => c.namespaces)
                            .find((ns) => ns.nodes.some((n) => n.id === d.id));

                        if (nodeNamespace) {
                            // Constrain to namespace boundaries with padding
                            const padding = 25;
                            d.fx = Math.max(
                                nodeNamespace.x + padding,
                                Math.min(
                                    nodeNamespace.x +
                                        nodeNamespace.width -
                                        padding,
                                    event.x
                                )
                            );
                            d.fy = Math.max(
                                nodeNamespace.y + padding,
                                Math.min(
                                    nodeNamespace.y +
                                        nodeNamespace.height -
                                        padding,
                                    event.y
                                )
                            );
                        } else {
                            d.fx = event.x;
                            d.fy = event.y;
                        }
                    })
                    .on('end', (event, d) => {
                        if (!event.active) simulation.alphaTarget(0);
                        d.fx = null;
                        d.fy = null;
                    })
            )
            .on('click', (event, d) => {
                event.stopPropagation();
                onNodeClick?.(d);
            });

        // Add circles for nodes with neutral styling
        nodeElements
            .append('circle')
            .attr('r', 24)
            .attr('fill', isDark ? '#374151' : '#e5e7eb') // Neutral gray
            .attr('stroke', isDark ? '#6b7280' : '#9ca3af') // Gray border
            .attr('stroke-width', 2)
            .style('transition', 'all 0.2s ease')
            .on('mouseenter', function () {
                d3.select(this).attr('r', 29);
            })
            .on('mouseleave', function () {
                d3.select(this).attr('r', 24);
            });

        // Add labels
        nodeElements
            .append('text')
            .text((d) => d.label)
            .attr('fill', isDark ? '#fafafa' : '#0f0f0f') // Use explicit colors for dark/light mode
            .attr('text-anchor', 'middle')
            .attr('dy', 44)
            .attr('font-size', '14px')
            .attr('font-weight', '500')
            .style('pointer-events', 'none')
            .style('user-select', 'none');

        // Update positions on simulation tick with boundary constraints
        simulation.on('tick', () => {
            // Apply boundary constraints to keep nodes within their namespaces
            d3Nodes.forEach((d) => {
                const nodeNamespace = clusterData
                    .flatMap((c) => c.namespaces)
                    .find((ns) => ns.nodes.some((n) => n.id === d.id));

                if (nodeNamespace) {
                    const padding = 25;
                    d.x = Math.max(
                        nodeNamespace.x + padding,
                        Math.min(
                            nodeNamespace.x + nodeNamespace.width - padding,
                            d.x || 0
                        )
                    );
                    d.y = Math.max(
                        nodeNamespace.y + padding,
                        Math.min(
                            nodeNamespace.y + nodeNamespace.height - padding,
                            d.y || 0
                        )
                    );
                }
            });

            // Update visual positions - lines should not overlap with node circles
            links
                .attr('x1', (d: D3Edge) => {
                    const dx =
                        (d.target as D3Node).x! - (d.source as D3Node).x!;
                    const dy =
                        (d.target as D3Node).y! - (d.source as D3Node).y!;
                    const length = Math.sqrt(dx * dx + dy * dy);
                    if (length === 0) return d.source.x;
                    const sourceRadius = 24 + 2;
                    const unitX = dx / length;
                    return (d.source as D3Node).x! + unitX * sourceRadius;
                })
                .attr('y1', (d: D3Edge) => {
                    const dx =
                        (d.target as D3Node).x! - (d.source as D3Node).x!;
                    const dy =
                        (d.target as D3Node).y! - (d.source as D3Node).y!;
                    const length = Math.sqrt(dx * dx + dy * dy);
                    if (length === 0) return d.source.y;
                    const sourceRadius = 24 + 2;
                    const unitY = dy / length;
                    return (d.source as D3Node).y! + unitY * sourceRadius;
                })
                .attr('x2', (d: D3Edge) => {
                    const dx =
                        (d.target as D3Node).x! - (d.source as D3Node).x!;
                    const dy =
                        (d.target as D3Node).y! - (d.source as D3Node).y!;
                    const length = Math.sqrt(dx * dx + dy * dy);
                    if (length === 0) return d.target.x;
                    const targetRadius = 24 + 2;
                    const unitX = dx / length;
                    return (d.target as D3Node).x! - unitX * targetRadius;
                })
                .attr('y2', (d: D3Edge) => {
                    const dx =
                        (d.target as D3Node).x! - (d.source as D3Node).x!;
                    const dy =
                        (d.target as D3Node).y! - (d.source as D3Node).y!;
                    const length = Math.sqrt(dx * dx + dy * dy);
                    if (length === 0) return d.target.y;
                    const targetRadius = 24 + 2;
                    const unitY = dy / length;
                    return (d.target as D3Node).y! - unitY * targetRadius;
                });

            // Update arrow triangles position - tip should touch the target node circle outline
            arrowTriangles.attr('points', (d: D3Edge) => {
                const dx = (d.target as D3Node).x! - (d.source as D3Node).x!;
                const dy = (d.target as D3Node).y! - (d.source as D3Node).y!;
                const length = Math.sqrt(dx * dx + dy * dy);

                if (length === 0) return '0,0 0,0 0,0'; // Avoid division by zero

                // Get target node radius
                const targetNodeRadius = 24 + 2; // Add 2 for stroke width

                // Calculate where arrow tip should be (at target node circle edge)
                const unitX = dx / length;
                const unitY = dy / length;
                const tipX = (d.target as D3Node).x! - unitX * targetNodeRadius;
                const tipY = (d.target as D3Node).y! - unitY * targetNodeRadius;

                // Arrow size for the triangle base
                const arrowSize = 12;

                // Triangle points: tip (at circle edge), left base, right base
                const leftX =
                    tipX - (unitY * arrowSize) / 2 - unitX * arrowSize;
                const leftY =
                    tipY + (unitX * arrowSize) / 2 - unitY * arrowSize;
                const rightX =
                    tipX + (unitY * arrowSize) / 2 - unitX * arrowSize;
                const rightY =
                    tipY - (unitX * arrowSize) / 2 - unitY * arrowSize;

                return `${tipX},${tipY} ${leftX},${leftY} ${rightX},${rightY}`;
            });

            // Update success rate labels position (along the line at midpoint)
            successLabels
                .attr(
                    'x',
                    (d: D3Edge) =>
                        ((d.source as D3Node).x! + (d.target as D3Node).x!) / 2
                )
                .attr(
                    'y',
                    (d: D3Edge) =>
                        ((d.source as D3Node).y! + (d.target as D3Node).y!) /
                            2 +
                        4
                ) // Slightly offset for readability
                .attr('transform', (d: D3Edge) => {
                    // Calculate angle of the line for text rotation
                    const dx =
                        (d.target as D3Node).x! - (d.source as D3Node).x!;
                    const dy =
                        (d.target as D3Node).y! - (d.source as D3Node).y!;
                    const angle = (Math.atan2(dy, dx) * 180) / Math.PI;
                    const midX =
                        ((d.source as D3Node).x! + (d.target as D3Node).x!) / 2;
                    const midY =
                        ((d.source as D3Node).y! + (d.target as D3Node).y!) /
                            2 +
                        4;

                    // Keep text readable - flip if angle would make it upside down
                    const adjustedAngle =
                        angle > 90 || angle < -90 ? angle + 180 : angle;

                    return `rotate(${adjustedAngle}, ${midX}, ${midY})`;
                });

            nodeElements.attr('transform', (d) => `translate(${d.x},${d.y})`);
        });

        // Cleanup
        return () => {
            simulation.stop();
        };
    }, [nodes, edges, dimensions, onNodeClick, onCanvasClick, getClusterColor]);

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

export const ServiceGraph: React.FC<ServiceGraphProps> = ({
    className = '',
    onNodeClick,
    fullScreen = false,
}) => {
    // Get metrics for the last hour by default - memoize to prevent infinite re-renders
    const timeRange = React.useMemo(() => {
        const endTime = new Date();
        const startTime = new Date(endTime.getTime() - 60 * 60 * 1000); // 1 hour ago
        return {
            startTime: startTime.toISOString(),
            endTime: endTime.toISOString(),
        };
    }, []); // Empty dependency array - only calculate once on mount

    const {
        data: metrics,
        isLoading,
        error,
        isError,
    } = useServiceGraphMetrics(timeRange);

    const graphData = React.useMemo(() => {
        if (!metrics) return { nodes: [], edges: [] };
        return transformMetricsToGraph(metrics);
    }, [metrics]);

    const handleNodeClick = React.useCallback(
        (node: D3Node) => {
            onNodeClick?.(node.id);
        },
        [onNodeClick]
    );

    const handleCanvasClick = React.useCallback(() => {
        // Reset any selection state if needed
    }, []);

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
                        Failed to load service graph: {error?.message}
                    </span>
                </CardContent>
            </Card>
        );
    }

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

    if (fullScreen) {
        return (
            <div className={`${className} h-full`}>
                <div className="h-full w-full relative overflow-hidden">
                    <D3Graph
                        nodes={graphData.nodes}
                        edges={graphData.edges}
                        onNodeClick={handleNodeClick}
                        onCanvasClick={handleCanvasClick}
                    />
                </div>
            </div>
        );
    }

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
                    </div>
                </div>
            </CardHeader>
            <CardContent className="p-0">
                <div className="h-96 w-full relative overflow-hidden">
                    <D3Graph
                        nodes={graphData.nodes}
                        edges={graphData.edges}
                        onNodeClick={handleNodeClick}
                        onCanvasClick={handleCanvasClick}
                    />
                </div>
            </CardContent>
        </Card>
    );
};
