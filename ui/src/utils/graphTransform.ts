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

import type { v1alpha1ServicePairMetrics } from '../types/generated/openapi-metrics_service';
import { getClusterColors, getEdgeColor } from './graphTheme';

export interface GraphNode {
    id: string;
    label: string;
    fill?: string;
    size?: number;
    data?: {
        cluster: string;
        namespace: string;
        service: string;
        totalRequestRate: number;
        totalErrorRate: number;
    };
}

export interface GraphEdge {
    id: string;
    source: string;
    target: string;
    label?: string;
    size?: number;
    fill?: string;
    data?: {
        requestRate: number;
        errorRate: number;
        sourceCluster: string;
        destinationCluster: string;
    };
}

export interface GraphData {
    nodes: GraphNode[];
    edges: GraphEdge[];
}

function getClusterColorFromPalette(
    cluster: string,
    clusters: string[]
): string {
    const palette = getClusterColors();
    const index = clusters.indexOf(cluster);
    return palette[index % palette.length];
}

function createServiceKey(
    cluster: string,
    namespace: string,
    service: string
): string {
    return `${cluster}/${namespace}/${service}`;
}

function parseServiceKey(key: string): {
    cluster: string;
    namespace: string;
    service: string;
} {
    const [cluster, namespace, service] = key.split('/');
    return { cluster, namespace, service };
}

export function transformMetricsToGraph(
    metrics: v1alpha1ServicePairMetrics[]
): GraphData {
    if (!metrics || metrics.length === 0) {
        return { nodes: [], edges: [] };
    }

    // Extract unique services from source and destination
    const serviceKeys = new Set<string>();
    const clusters = new Set<string>();
    const namespaces = new Set<string>();

    metrics.forEach((metric) => {
        const sourceKey = createServiceKey(
            metric.sourceCluster || '',
            metric.sourceNamespace || '',
            metric.sourceService || ''
        );
        const destinationKey = createServiceKey(
            metric.destinationCluster || '',
            metric.destinationNamespace || '',
            metric.destinationService || ''
        );

        serviceKeys.add(sourceKey);
        serviceKeys.add(destinationKey);
        clusters.add(metric.sourceCluster || '');
        clusters.add(metric.destinationCluster || '');
        namespaces.add(metric.sourceNamespace || '');
        namespaces.add(metric.destinationNamespace || '');
    });

    const clusterList = Array.from(clusters).filter(Boolean).sort();

    // Calculate aggregated metrics for each service node
    const serviceMetrics = new Map<
        string,
        { totalRequestRate: number; totalErrorRate: number }
    >();

    metrics.forEach((metric) => {
        const sourceKey = createServiceKey(
            metric.sourceCluster || '',
            metric.sourceNamespace || '',
            metric.sourceService || ''
        );
        const destinationKey = createServiceKey(
            metric.destinationCluster || '',
            metric.destinationNamespace || '',
            metric.destinationService || ''
        );

        // Add to source service (outgoing traffic)
        const sourceData = serviceMetrics.get(sourceKey) || {
            totalRequestRate: 0,
            totalErrorRate: 0,
        };
        sourceData.totalRequestRate += metric.requestRate || 0;
        sourceData.totalErrorRate += metric.errorRate || 0;
        serviceMetrics.set(sourceKey, sourceData);

        // Add to destination service (incoming traffic)
        const destData = serviceMetrics.get(destinationKey) || {
            totalRequestRate: 0,
            totalErrorRate: 0,
        };
        destData.totalRequestRate += metric.requestRate || 0;
        destData.totalErrorRate += metric.errorRate || 0;
        serviceMetrics.set(destinationKey, destData);
    });

    // Create nodes
    const nodes: GraphNode[] = Array.from(serviceKeys).map((serviceKey) => {
        const { cluster, namespace, service } = parseServiceKey(serviceKey);
        const metrics = serviceMetrics.get(serviceKey) || {
            totalRequestRate: 0,
            totalErrorRate: 0,
        };

        // Use cluster for primary color, namespace for variation
        const clusterColor = getClusterColorFromPalette(cluster, clusterList);

        // Calculate node size based on total request rate
        const maxRequestRate = Math.max(
            ...Array.from(serviceMetrics.values()).map(
                (m) => m.totalRequestRate
            )
        );
        const minSize = 8;
        const maxSize = 32;
        const size =
            maxRequestRate > 0
                ? minSize +
                  (metrics.totalRequestRate / maxRequestRate) *
                      (maxSize - minSize)
                : minSize;

        return {
            id: serviceKey,
            label: service,
            fill: clusterColor,
            size: Math.round(size),
            data: {
                cluster,
                namespace,
                service,
                totalRequestRate: metrics.totalRequestRate,
                totalErrorRate: metrics.totalErrorRate,
            },
        };
    });

    // Create edges
    const edges: GraphEdge[] = metrics.map((metric, index) => {
        const sourceKey = createServiceKey(
            metric.sourceCluster || '',
            metric.sourceNamespace || '',
            metric.sourceService || ''
        );
        const destinationKey = createServiceKey(
            metric.destinationCluster || '',
            metric.destinationNamespace || '',
            metric.destinationService || ''
        );

        // Calculate edge thickness based on request rate
        const maxRequestRate = Math.max(
            ...metrics.map((m) => m.requestRate || 0)
        );
        const minSize = 1;
        const maxSize = 8;
        const size =
            maxRequestRate > 0
                ? minSize +
                  ((metric.requestRate || 0) / maxRequestRate) *
                      (maxSize - minSize)
                : minSize;

        // Color edge based on error rate using theme-aware colors
        const edgeColor = getEdgeColor(
            metric.requestRate || 0,
            metric.errorRate || 0
        );

        const label =
            (metric.requestRate || 0) > 0
                ? `${(metric.requestRate || 0).toFixed(1)} rps`
                : undefined;

        return {
            id: `edge-${index}`,
            source: sourceKey,
            target: destinationKey,
            label,
            size: Math.round(size),
            fill: edgeColor,
            data: {
                requestRate: metric.requestRate || 0,
                errorRate: metric.errorRate || 0,
                sourceCluster: metric.sourceCluster || '',
                destinationCluster: metric.destinationCluster || '',
            },
        };
    });

    return { nodes, edges };
}

export function getServiceDisplayName(serviceKey: string): string {
    const { cluster, namespace, service } = parseServiceKey(serviceKey);
    return `${service} (${namespace}/${cluster})`;
}

export function formatRequestRate(rate: number): string {
    if (rate < 1) {
        return `${(rate * 1000).toFixed(0)}m rps`;
    }
    return `${rate.toFixed(1)} rps`;
}

export function formatErrorRate(rate: number): string {
    return `${(rate * 100).toFixed(1)}%`;
}
