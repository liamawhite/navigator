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

import React, { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import * as d3 from 'd3';
import type { v1alpha1ServicePairMetrics } from '../../types/generated/openapi-metrics_service';
import { useTheme } from '../theme-provider';

interface ServiceConnectionsVisualizationProps {
    serviceName: string;
    namespace: string;
    inbound: v1alpha1ServicePairMetrics[];
    outbound: v1alpha1ServicePairMetrics[];
}

interface ConnectionNode {
    id: string;
    name: string;
    type: 'center' | 'inbound' | 'outbound';
    x?: number;
    y?: number;
    cluster?: string;
    namespace?: string;
    hasEnvoy?: boolean;
}

interface ConnectionLink {
    source: string;
    target: string;
    requestRate: number;
    errorRate: number;
    type: 'inbound' | 'outbound';
}

export const ServiceConnectionsVisualization: React.FC<
    ServiceConnectionsVisualizationProps
> = ({ serviceName, namespace, inbound, outbound }) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const navigate = useNavigate();
    const { theme } = useTheme();
    const isDark =
        theme === 'dark' ||
        (theme === 'system' &&
            window.matchMedia('(prefers-color-scheme: dark)').matches);

    useEffect(() => {
        if (!containerRef.current) return;

        // Clear any existing visualization
        d3.select(containerRef.current).selectAll('*').remove();

        // Process data
        const nodes: ConnectionNode[] = [
            {
                id: 'center',
                name: `${serviceName}.${namespace}`,
                type: 'center',
                x: 0,
                y: 0,
            },
        ];

        const links: ConnectionLink[] = [];

        // Add inbound connections
        inbound.forEach((conn) => {
            if ((conn.requestRate || 0) < 0.01) return;

            const sourceService = conn.sourceService || 'unknown';
            const sourceNamespace = conn.sourceNamespace || 'unknown';
            const sourceId = `${sourceService}-${sourceNamespace}-${conn.sourceCluster}`;

            // Show unknown services without namespace
            const sourceName =
                sourceService === 'unknown'
                    ? 'unknown'
                    : `${sourceService}.${sourceNamespace}`;

            nodes.push({
                id: sourceId,
                name: sourceName,
                type: 'inbound',
                cluster: conn.sourceCluster,
                namespace: sourceNamespace,
            });

            links.push({
                source: sourceId,
                target: 'center',
                requestRate: conn.requestRate || 0,
                errorRate: conn.errorRate || 0,
                type: 'inbound',
            });
        });

        // Add outbound connections
        outbound.forEach((conn) => {
            if ((conn.requestRate || 0) < 0.01) return;

            const destinationService = conn.destinationService || 'unknown';
            const destinationNamespace = conn.destinationNamespace || 'unknown';
            const targetId = `${destinationService}-${destinationNamespace}-${conn.destinationCluster}`;

            // Show unknown services without namespace
            const targetName =
                destinationService === 'unknown'
                    ? 'unknown'
                    : `${destinationService}.${destinationNamespace}`;

            nodes.push({
                id: targetId,
                name: targetName,
                type: 'outbound',
                cluster: conn.destinationCluster,
                namespace: destinationNamespace,
            });

            links.push({
                source: 'center',
                target: targetId,
                requestRate: conn.requestRate || 0,
                errorRate: conn.errorRate || 0,
                type: 'outbound',
            });
        });

        if (nodes.length === 1) {
            // No connections to show
            return;
        }

        // Set up dimensions
        const containerRect = containerRef.current.getBoundingClientRect();
        const width = Math.max(800, containerRect.width);
        const height = 400;
        const centerX = width / 2;
        const centerY = height / 2;

        // Create SVG
        const svg = d3
            .select(containerRef.current)
            .append('svg')
            .attr('width', '100%')
            .attr('height', height)
            .attr('viewBox', `0 0 ${width} ${height}`)
            .style('background', 'transparent');

        // Define colors based on theme
        const colors = {
            background: isDark ? '#1f2937' : '#ffffff',
            text: isDark ? '#e5e7eb' : '#374151', // Original color for rps labels (gray-200 / gray-700)
            textMuted: isDark ? '#9ca3af' : '#6b7280', // More muted for service names (gray-400 / gray-500)
            muted: isDark ? '#6b7280' : '#9ca3af',
            center: isDark ? '#3b82f6' : '#2563eb', // blue
            inbound: isDark ? '#6b7280' : '#9ca3af', // grey
            outbound: isDark ? '#6b7280' : '#9ca3af', // grey
            envoy: isDark ? '#8b5cf6' : '#7c3aed', // purple
            errorLow: isDark ? '#10b981' : '#059669', // green
            errorMed: isDark ? '#f59e0b' : '#d97706', // amber
            errorHigh: isDark ? '#ef4444' : '#dc2626', // red
        };

        // Remove arrow markers for simplicity - just using lines now

        // Position nodes - group by namespace, then sort alphabetically within namespace
        const inboundNodes = nodes
            .filter((n) => n.type === 'inbound')
            .sort((a, b) => {
                if (a.namespace !== b.namespace) {
                    return (a.namespace || '').localeCompare(b.namespace || '');
                }
                return a.name.localeCompare(b.name);
            });
        const outboundNodes = nodes
            .filter((n) => n.type === 'outbound')
            .sort((a, b) => {
                if (a.namespace !== b.namespace) {
                    return (a.namespace || '').localeCompare(b.namespace || '');
                }
                return a.name.localeCompare(b.name);
            });

        // Calculate max text width for inbound nodes (right-aligned)
        const maxInboundWidth = Math.max(
            ...inboundNodes.map(
                (node) =>
                    Math.max(node.name.length, (node.cluster || '').length) *
                        8 +
                    10 // Less padding
            ),
            100 // smaller minimum space
        );

        // Calculate max text width for outbound nodes (left-aligned)
        const maxOutboundWidth = Math.max(
            ...outboundNodes.map(
                (node) =>
                    Math.max(node.name.length, (node.cluster || '').length) *
                        8 +
                    10 // Less padding
            ),
            100 // smaller minimum space
        );

        // Pin inbound nodes with calculated spacing
        const leftX = maxInboundWidth;
        inboundNodes.forEach((node, i) => {
            const spacing = Math.max(60, height / (inboundNodes.length + 1));
            node.x = leftX;
            node.y = spacing * (i + 1);
        });

        // Pin outbound nodes with calculated spacing
        const rightX = width - maxOutboundWidth;
        outboundNodes.forEach((node, i) => {
            const spacing = Math.max(60, height / (outboundNodes.length + 1));
            node.x = rightX;
            node.y = spacing * (i + 1);
        });

        // Find center node
        const centerNode = nodes.find((n) => n.type === 'center');
        if (centerNode) {
            centerNode.x = centerX;
            centerNode.y = centerY;
        }

        // Calculate max request rate for relative sizing
        const maxRequestRate = Math.max(...links.map((d) => d.requestRate));

        // Create links
        const linkElements = svg
            .selectAll('.link')
            .data(links)
            .enter()
            .append('g')
            .attr('class', 'link');

        // Draw connection paths
        linkElements
            .append('path')
            .attr('class', 'connection-path')
            .attr('d', (d) => {
                const sourceNode = nodes.find((n) => n.id === d.source);
                const targetNode = nodes.find((n) => n.id === d.target);
                if (
                    !sourceNode ||
                    !targetNode ||
                    !sourceNode.x ||
                    !sourceNode.y ||
                    !targetNode.x ||
                    !targetNode.y
                ) {
                    return '';
                }

                // No arrows, so no need for arrow buffer
                let sourceEdgeX, sourceEdgeY, targetEdgeX, targetEdgeY;

                // Calculate direction vector
                const dx = targetNode.x - sourceNode.x;
                const dy = targetNode.y - sourceNode.y;
                const distance = Math.sqrt(dx * dx + dy * dy);

                // Source edge calculation
                if (sourceNode.type === 'center') {
                    // Circle edge with fixed buffer
                    const sourceRadius = 45 + 12; // Fixed gap from circle to line
                    sourceEdgeX = sourceNode.x + (dx / distance) * sourceRadius;
                    sourceEdgeY = sourceNode.y + (dy / distance) * sourceRadius;
                } else {
                    // Text node edge - point from center between service name and cluster name
                    const textBuffer = 12; // Smaller gap from text to arrow
                    const centerY = sourceNode.y + (sourceNode.cluster ? 3 : 0); // Offset to center between lines

                    // Recalculate direction from center point
                    const dxFromCenter = targetNode.x - sourceNode.x;
                    const dyFromCenter = targetNode.y - centerY;
                    const distanceFromCenter = Math.sqrt(
                        dxFromCenter * dxFromCenter +
                            dyFromCenter * dyFromCenter
                    );

                    sourceEdgeX =
                        sourceNode.x +
                        (dxFromCenter / distanceFromCenter) * textBuffer;
                    sourceEdgeY =
                        centerY +
                        (dyFromCenter / distanceFromCenter) * textBuffer;
                }

                // Target edge calculation
                if (targetNode.type === 'center') {
                    // Circle edge with fixed buffer - no arrows to worry about
                    const targetRadius = 45 + 12; // Fixed gap from line to circle
                    targetEdgeX = targetNode.x - (dx / distance) * targetRadius;
                    targetEdgeY = targetNode.y - (dy / distance) * targetRadius;
                } else {
                    // Text node edge - point to center between service name and cluster name
                    const textBuffer = 12; // Gap from line to text
                    const centerY = targetNode.y + (targetNode.cluster ? 3 : 0); // Offset to center between lines

                    // Recalculate direction to center point
                    const dxToCenter = targetNode.x - sourceNode.x;
                    const dyToCenter = centerY - sourceNode.y;
                    const distanceToCenter = Math.sqrt(
                        dxToCenter * dxToCenter + dyToCenter * dyToCenter
                    );

                    targetEdgeX =
                        targetNode.x -
                        (dxToCenter / distanceToCenter) * textBuffer;
                    targetEdgeY =
                        centerY - (dyToCenter / distanceToCenter) * textBuffer;
                }

                // Create smooth curved path for funds-flow style
                const curve = Math.abs(dx) * 0.3; // Horizontal curve amount

                // Use cubic bezier curve for smooth flow from edge to edge
                const midX1 = sourceEdgeX + curve;
                const midX2 = targetEdgeX - curve;

                return `M${sourceEdgeX},${sourceEdgeY}C${midX1},${sourceEdgeY} ${midX2},${targetEdgeY} ${targetEdgeX},${targetEdgeY}`;
            })
            .attr('fill', 'none')
            .attr('stroke', (d) => {
                // Color based on success rate (inverted error rate)
                const errorPercent =
                    d.requestRate > 0 ? (d.errorRate / d.requestRate) * 100 : 0;
                const successPercent = 100 - errorPercent;

                if (successPercent >= 99) return colors.errorLow; // Green for high success
                if (successPercent >= 95) return colors.errorMed; // Amber for medium success
                return colors.errorHigh; // Red for low success
            })
            .attr('stroke-width', (d) => {
                // Calculate thickness as percentage of max rate (1-8px range)
                const relativeThickness =
                    maxRequestRate > 0 ? d.requestRate / maxRequestRate : 0;
                return Math.max(1, Math.min(8, relativeThickness * 8));
            })
            // No arrow markers needed
            .style('opacity', 0.8);

        // Add request rate labels on links
        linkElements
            .append('text')
            .attr('class', 'rate-label')
            .attr('x', 0) // Will be positioned by transform
            .attr('y', 0) // Will be positioned by transform
            .attr('text-anchor', 'middle')
            .attr('font-size', '10px')
            .attr('fill', colors.text)
            .attr('transform', (d) => {
                const sourceNode = nodes.find((n) => n.id === d.source);
                const targetNode = nodes.find((n) => n.id === d.target);
                if (
                    !sourceNode ||
                    !targetNode ||
                    !sourceNode.x ||
                    !targetNode.x ||
                    !sourceNode.y ||
                    !targetNode.y
                )
                    return '';

                // Replicate the same edge calculation logic from the path drawing
                const dx = targetNode.x - sourceNode.x;
                const dy = targetNode.y - sourceNode.y;
                const distance = Math.sqrt(dx * dx + dy * dy);

                let sourceEdgeX, sourceEdgeY, targetEdgeX, targetEdgeY;

                // Source edge calculation
                if (sourceNode.type === 'center') {
                    const sourceRadius = 45 + 12;
                    sourceEdgeX = sourceNode.x + (dx / distance) * sourceRadius;
                    sourceEdgeY = sourceNode.y + (dy / distance) * sourceRadius;
                } else {
                    const textBuffer = 12;
                    const centerY = sourceNode.y + (sourceNode.cluster ? 3 : 0);
                    const dxFromCenter = targetNode.x - sourceNode.x;
                    const dyFromCenter = targetNode.y - centerY;
                    const distanceFromCenter = Math.sqrt(
                        dxFromCenter * dxFromCenter +
                            dyFromCenter * dyFromCenter
                    );
                    sourceEdgeX =
                        sourceNode.x +
                        (dxFromCenter / distanceFromCenter) * textBuffer;
                    sourceEdgeY =
                        centerY +
                        (dyFromCenter / distanceFromCenter) * textBuffer;
                }

                // Target edge calculation
                if (targetNode.type === 'center') {
                    const targetRadius = 45 + 12;
                    targetEdgeX = targetNode.x - (dx / distance) * targetRadius;
                    targetEdgeY = targetNode.y - (dy / distance) * targetRadius;
                } else {
                    const textBuffer = 12;
                    const centerY = targetNode.y + (targetNode.cluster ? 3 : 0);
                    const dxToCenter = targetNode.x - sourceNode.x;
                    const dyToCenter = centerY - sourceNode.y;
                    const distanceToCenter = Math.sqrt(
                        dxToCenter * dxToCenter + dyToCenter * dyToCenter
                    );
                    targetEdgeX =
                        targetNode.x -
                        (dxToCenter / distanceToCenter) * textBuffer;
                    targetEdgeY =
                        centerY - (dyToCenter / distanceToCenter) * textBuffer;
                }

                // Find a middle ground between flat curve middle and steep endpoints
                const edgeDx = targetEdgeX - sourceEdgeX;
                const edgeDy = targetEdgeY - sourceEdgeY;

                // Use 95% of the edge direction to get closer to curve feel without being too flat
                const adjustedDy = edgeDy * 0.95;

                const flowAngle =
                    Math.atan2(adjustedDy, edgeDx) * (180 / Math.PI);

                // Get midpoint for rotation center (centered on line - 15 offset)
                const midX = (sourceEdgeX + targetEdgeX) / 2;
                const midY = (sourceEdgeY + targetEdgeY) / 2 - 15;

                return `translate(${midX}, ${midY}) rotate(${flowAngle})`;
            })
            .text((d) => {
                if (d.requestRate < 0.01) return '';

                // Calculate success rate
                const errorPercent =
                    d.requestRate > 0 ? (d.errorRate / d.requestRate) * 100 : 0;
                const successPercent = 100 - errorPercent;

                // Format success rate to appropriate precision
                let successRateStr;
                if (successPercent >= 100) {
                    successRateStr = '100';
                } else if (successPercent >= 10) {
                    successRateStr = successPercent.toFixed(1);
                } else {
                    successRateStr = successPercent.toFixed(2);
                }

                return `${d.requestRate.toFixed(2)} rps · ${successRateStr}%`;
            });

        // Create node groups
        const nodeElements = svg
            .selectAll('.node')
            .data(nodes)
            .enter()
            .append('g')
            .attr('class', 'node')
            .attr('transform', (d) => `translate(${d.x || 0}, ${d.y || 0})`)
            .style('cursor', (d) =>
                d.type !== 'center' && d.name !== 'unknown'
                    ? 'pointer'
                    : 'default'
            )
            .on('click', (event, d) => {
                if (
                    d.type !== 'center' &&
                    d.name !== 'unknown' &&
                    d.namespace
                ) {
                    // Parse service name from the display format (service.namespace)
                    const serviceName = d.name.split('.')[0];
                    navigate(`/services/${d.namespace}:${serviceName}`);
                }
            });

        // Add node shapes - only circles for center nodes
        nodeElements
            .filter((d) => d.type === 'center')
            .append('circle')
            .attr('r', 45)
            .attr('fill', colors.center)
            .attr('stroke', colors.background)
            .attr('stroke-width', 2)
            .style('opacity', 0.9);

        // Add hover effects to edge nodes only
        nodeElements
            .filter((d) => d.type !== 'center' && d.name !== 'unknown')
            .on('mouseenter', function (_event, _d) {
                const nodeGroup = d3.select(this);

                // Stop any running transitions first
                nodeGroup.selectAll('*').transition().duration(0);

                // Hover effect for text - much brighter for better contrast
                nodeGroup
                    .selectAll('text')
                    .transition()
                    .duration(150)
                    .attr('fill', isDark ? '#f3f4f6' : '#111827'); // gray-100 / gray-900
            })
            .on('mouseleave', function (_event, _d) {
                const nodeGroup = d3.select(this);

                // Stop any running transitions first
                nodeGroup.selectAll('*').transition().duration(0);

                // Reset text colors
                nodeGroup
                    .select('text:first-of-type')
                    .transition()
                    .duration(150)
                    .attr('fill', colors.textMuted);

                // Reset cluster badge
                nodeGroup
                    .select('text:last-of-type')
                    .transition()
                    .duration(150)
                    .attr('fill', colors.muted);
            });

        // Add service names
        nodeElements
            .append('text')
            .attr('dy', (d) => {
                if (d.type === 'center') return 60;
                return d.cluster ? -3 : 4; // Higher if we have cluster text below
            })
            .attr('text-anchor', (d) => {
                if (d.type === 'center') return 'middle';
                return d.type === 'inbound' ? 'end' : 'start'; // Left align inbound, right align outbound
            })
            .attr('font-size', (d) => (d.type === 'center' ? '14px' : '12px'))
            .attr('font-weight', (d) =>
                d.type === 'center' ? 'bold' : 'medium'
            )
            .attr('fill', (d) =>
                d.type === 'center' ? colors.text : colors.textMuted
            )
            .text((d) => d.name);

        // Add cluster labels for non-center nodes
        nodeElements
            .filter((d) => d.type !== 'center' && d.cluster)
            .append('text')
            .attr('dy', 12) // Below service name
            .attr('text-anchor', (d) =>
                d.type === 'inbound' ? 'end' : 'start'
            ) // Match service name alignment
            .attr('font-size', '10px')
            .attr('fill', colors.muted)
            .text((d) => d.cluster || '');
    }, [serviceName, namespace, inbound, outbound, theme, isDark, navigate]);

    return (
        <div
            ref={containerRef}
            className="w-full"
            style={{ minHeight: '400px' }}
        />
    );
};
