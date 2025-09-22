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

import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

export function formatLastUpdated(date: Date | null): string {
    if (!date) return 'Never';

    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
    const diffInMinutes = Math.floor(diffInSeconds / 60);
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
    const diffInHours = Math.floor(diffInMinutes / 60);
    return `${diffInHours}h ago`;
}

// Service pair metrics aggregation utilities
export interface ServicePairMetrics {
    sourceCluster?: string;
    sourceNamespace?: string;
    sourceService?: string;
    destinationCluster?: string;
    destinationNamespace?: string;
    destinationService?: string;
    errorRate?: number;
    requestRate?: number;
    latencyP99?: string;
    latencyDistribution?: {
        buckets?: Array<{ le?: number; count?: number }>;
        totalCount?: number;
        sum?: number;
    };
}

export interface ServiceAggregatedMetrics {
    serviceName: string;
    namespace: string;
    errorRate: number;
    requestRate: number;
    latencyP99?: string;
    latencyP99Ms: number;
    servicePairs: ServicePairMetrics[];
    clusterPairCount: number;
}

const parseDurationToMs = (duration: string | undefined): number => {
    if (!duration) return 0;

    // Parse protobuf duration string format (e.g., "0.150s", "25ms")
    const match = duration.match(/^(\d+(?:\.\d+)?)(s|ms|ns)$/);
    if (!match) return 0;

    const value = parseFloat(match[1]);
    const unit = match[2];

    switch (unit) {
        case 's':
            return value * 1000; // seconds to milliseconds
        case 'ms':
            return value; // already milliseconds
        case 'ns':
            return value / 1000000; // nanoseconds to milliseconds
        default:
            return 0;
    }
};

const formatDurationFromMs = (latencyMs: number): string => {
    if (latencyMs === 0) return '0ms';
    if (latencyMs >= 1000) {
        return `${(latencyMs / 1000).toFixed(3)}s`;
    }
    return `${latencyMs.toFixed(3)}ms`;
};

// Aggregate histogram distributions by properly combining individual bucket counts
const aggregateLatencyDistributions = (
    distributions: Array<{
        buckets?: Array<{ le?: number; count?: number }>;
        totalCount?: number;
        sum?: number;
    }>
) => {
    if (distributions.length === 0) {
        return { buckets: [], totalCount: 0, sum: 0 };
    }

    // Collect all unique bucket boundaries
    const allBucketBoundaries = new Set<number>();
    distributions.forEach((dist) => {
        if (dist.buckets) {
            dist.buckets.forEach((bucket) => {
                if (bucket.le !== undefined) {
                    allBucketBoundaries.add(bucket.le);
                }
            });
        }
    });

    const sortedBoundaries = Array.from(allBucketBoundaries).sort(
        (a, b) => a - b
    );

    // Convert cumulative counts to individual bucket counts for each distribution
    const convertedDistributions = distributions.map((dist) => {
        if (!dist.buckets) return { buckets: [] };

        const sortedBuckets = [...dist.buckets].sort(
            (a, b) => (a.le || 0) - (b.le || 0)
        );
        const individualBuckets: Array<{ le: number; count: number }> = [];

        let previousCumulativeCount = 0;
        for (const bucket of sortedBuckets) {
            const cumulativeCount = bucket.count || 0;
            const individualCount = cumulativeCount - previousCumulativeCount;
            individualBuckets.push({
                le: bucket.le || 0,
                count: individualCount,
            });
            previousCumulativeCount = cumulativeCount;
        }

        return { buckets: individualBuckets };
    });

    // For each boundary, sum the individual counts from all distributions
    const aggregatedIndividualBuckets = sortedBoundaries.map((boundary) => {
        let totalIndividualCount = 0;

        convertedDistributions.forEach((dist) => {
            const bucket = dist.buckets.find((b) => b.le === boundary);
            if (bucket) {
                totalIndividualCount += bucket.count;
            }
        });

        return { le: boundary, count: totalIndividualCount };
    });

    // Convert back to cumulative counts
    const aggregatedBuckets: Array<{ le: number; count: number }> = [];
    let cumulativeCount = 0;

    for (const bucket of aggregatedIndividualBuckets) {
        cumulativeCount += bucket.count;
        aggregatedBuckets.push({ le: bucket.le, count: cumulativeCount });
    }

    const totalCount = distributions.reduce(
        (sum, dist) => sum + (dist.totalCount || 0),
        0
    );
    const totalSum = distributions.reduce(
        (sum, dist) => sum + (dist.sum || 0),
        0
    );

    return { buckets: aggregatedBuckets, totalCount, sum: totalSum };
};

// Calculate P99 from aggregated histogram distribution
const calculateP99FromDistribution = (distribution: {
    buckets?: Array<{ le?: number; count?: number }>;
    totalCount?: number;
}) => {
    if (
        !distribution.buckets ||
        distribution.buckets.length === 0 ||
        !distribution.totalCount ||
        distribution.totalCount === 0
    ) {
        return 0;
    }

    // Sort buckets by le (upper bound) to ensure correct order
    const sortedBuckets = [...distribution.buckets].sort(
        (a, b) => (a.le || 0) - (b.le || 0)
    );

    // Calculate P99 from histogram buckets
    // Note: Istio histogram buckets are already in milliseconds, no conversion needed
    const p99Target = distribution.totalCount * 0.99;

    for (const bucket of sortedBuckets) {
        const cumulativeCount = bucket.count || 0;
        if (cumulativeCount >= p99Target) {
            return bucket.le || 0; // Already in milliseconds
        }
    }

    // If we reach here, return the last bucket's upper bound
    const lastBucket = sortedBuckets[sortedBuckets.length - 1];
    return lastBucket?.le || 0;
};

// Aggregate service pairs by destination service (for inbound traffic)
export const aggregateInboundByService = (
    servicePairs: ServicePairMetrics[]
): ServiceAggregatedMetrics[] => {
    const serviceMap = new Map<string, ServicePairMetrics[]>();

    // Group by destination service (namespace:serviceName)
    servicePairs.forEach((pair) => {
        const serviceKey = `${pair.destinationNamespace || 'unknown'}:${pair.destinationService || 'unknown'}`;
        if (!serviceMap.has(serviceKey)) {
            serviceMap.set(serviceKey, []);
        }
        serviceMap.get(serviceKey)!.push(pair);
    });

    // Aggregate metrics for each service
    return Array.from(serviceMap.entries()).map(([serviceKey, pairs]) => {
        const [namespace, serviceName] = serviceKey.split(':');
        const totalRequestRate = pairs.reduce(
            (sum, pair) => sum + (pair.requestRate || 0),
            0
        );
        const totalErrorRate = pairs.reduce(
            (sum, pair) => sum + (pair.errorRate || 0),
            0
        );

        // Aggregate latency distributions
        const distributions = pairs
            .map((pair) => pair.latencyDistribution)
            .filter((dist) => dist !== undefined);

        // Use weighted average of P99 values since backend provides correctly calculated values
        let p99Ms = 0;
        const validPairs = pairs.filter(
            (p) => p.latencyP99 && p.requestRate && p.requestRate > 0
        );
        if (validPairs.length > 0) {
            const totalWeight = validPairs.reduce(
                (sum, p) => sum + (p.requestRate || 0),
                0
            );
            if (totalWeight > 0) {
                const weightedSum = validPairs.reduce((sum, p) => {
                    const latencyMs = parseDurationToMs(p.latencyP99);
                    return sum + latencyMs * (p.requestRate || 0);
                }, 0);
                p99Ms = weightedSum / totalWeight;
            }
        }

        return {
            serviceName,
            namespace,
            errorRate: totalErrorRate,
            requestRate: totalRequestRate,
            latencyP99: formatDurationFromMs(p99Ms),
            latencyP99Ms: p99Ms,
            servicePairs: pairs,
            clusterPairCount: pairs.length,
        };
    });
};

// Aggregate service pairs by destination service (for outbound traffic)
export const aggregateOutboundByService = (
    servicePairs: ServicePairMetrics[]
): ServiceAggregatedMetrics[] => {
    const serviceMap = new Map<string, ServicePairMetrics[]>();

    // Group by destination service (namespace:serviceName)
    servicePairs.forEach((pair) => {
        const serviceKey = `${pair.destinationNamespace || 'unknown'}:${pair.destinationService || 'unknown'}`;
        if (!serviceMap.has(serviceKey)) {
            serviceMap.set(serviceKey, []);
        }
        serviceMap.get(serviceKey)!.push(pair);
    });

    // Aggregate metrics for each service
    return Array.from(serviceMap.entries()).map(([serviceKey, pairs]) => {
        const [namespace, serviceName] = serviceKey.split(':');
        const totalRequestRate = pairs.reduce(
            (sum, pair) => sum + (pair.requestRate || 0),
            0
        );
        const totalErrorRate = pairs.reduce(
            (sum, pair) => sum + (pair.errorRate || 0),
            0
        );

        // Aggregate latency distributions
        const distributions = pairs
            .map((pair) => pair.latencyDistribution)
            .filter((dist) => dist !== undefined);

        // Use weighted average of P99 values since backend provides correctly calculated values
        let p99Ms = 0;
        const validPairs = pairs.filter(
            (p) => p.latencyP99 && p.requestRate && p.requestRate > 0
        );
        if (validPairs.length > 0) {
            const totalWeight = validPairs.reduce(
                (sum, p) => sum + (p.requestRate || 0),
                0
            );
            if (totalWeight > 0) {
                const weightedSum = validPairs.reduce((sum, p) => {
                    const latencyMs = parseDurationToMs(p.latencyP99);
                    return sum + latencyMs * (p.requestRate || 0);
                }, 0);
                p99Ms = weightedSum / totalWeight;
            }
        }

        return {
            serviceName,
            namespace,
            errorRate: totalErrorRate,
            requestRate: totalRequestRate,
            latencyP99: formatDurationFromMs(p99Ms),
            latencyP99Ms: p99Ms,
            servicePairs: pairs,
            clusterPairCount: pairs.length,
        };
    });
};
