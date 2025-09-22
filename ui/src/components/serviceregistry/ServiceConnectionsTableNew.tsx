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

import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import {
    ArrowRight,
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
    ChevronRight,
    ChevronDown,
} from 'lucide-react';
import type { v1alpha1AggregatedServicePairMetrics } from '../../types/generated/openapi-metrics_service';

interface ServiceConnectionsTableProps {
    inbound: v1alpha1AggregatedServicePairMetrics[];
    outbound: v1alpha1AggregatedServicePairMetrics[];
}

interface ServiceRowData {
    serviceName: string;
    namespace: string;
    requestRate: number;
    successRate: number;
    latencyP99: string;
    latencyP99Ms: number;
    clusterInfo: string;
    isClickable: boolean;
}

type SortField = 'service' | 'requestRate' | 'successRate' | 'latencyP99';
type SortDirection = 'asc' | 'desc' | null;

interface SortState {
    field: SortField | null;
    direction: SortDirection;
}

const MIN_REQUEST_RATE = 0.01;
const SUCCESS_RATE_EXCELLENT = 99;
const SUCCESS_RATE_GOOD = 95;

export const ServiceConnectionsTable: React.FC<
    ServiceConnectionsTableProps
> = ({ inbound, outbound }) => {
    const navigate = useNavigate();
    const [expandedServices, setExpandedServices] = useState<Set<string>>(
        new Set()
    );
    const [sortState, setSortState] = useState<SortState>({
        field: 'requestRate',
        direction: 'desc',
    });

    // Parse duration string to milliseconds
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

    // Format latency for display
    const formatLatency = (duration: string | undefined): string => {
        const latencyMs = parseDurationToMs(duration);
        if (latencyMs === 0) {
            return '-';
        } else if (latencyMs >= 1000) {
            return `${(latencyMs / 1000).toFixed(1)}s`;
        } else {
            return `${latencyMs.toFixed(0)}ms`;
        }
    };

    // Format request rate
    const formatRequestRate = (rate: number): string => {
        if (rate >= 1) {
            return `${rate.toFixed(0)} RPS`;
        } else {
            return `${rate.toFixed(2)} RPS`;
        }
    };

    // Get success rate color
    const getSuccessRateColor = (rate: number): string => {
        if (rate >= SUCCESS_RATE_EXCELLENT) return 'text-green-600';
        if (rate >= SUCCESS_RATE_GOOD) return 'text-amber-600';
        return 'text-red-600';
    };

    // Convert aggregated metrics to display format
    const processAggregatedMetrics = (
        metrics: v1alpha1AggregatedServicePairMetrics[],
        type: 'inbound' | 'outbound'
    ): ServiceRowData[] => {
        return metrics
            .filter((metric) => (metric.requestRate || 0) >= MIN_REQUEST_RATE)
            .map((metric) => {
                const serviceName =
                    type === 'inbound'
                        ? metric.sourceService || 'unknown'
                        : metric.destinationService || 'unknown';
                const namespace =
                    type === 'inbound'
                        ? metric.sourceNamespace || 'unknown'
                        : metric.destinationNamespace || 'unknown';

                const requestRate = metric.requestRate || 0;
                const errorRate = metric.errorRate || 0;
                const successRate =
                    requestRate > 0
                        ? ((requestRate - errorRate) / requestRate) * 100
                        : 100;

                // Format cluster information
                const clusterInfo =
                    metric.clusterPairs
                        ?.map(
                            (cp) =>
                                `${cp.sourceCluster} → ${cp.destinationCluster}`
                        )
                        .join(', ') || 'unknown → unknown';

                return {
                    serviceName,
                    namespace,
                    requestRate,
                    successRate,
                    latencyP99: formatLatency(metric.latencyP99),
                    latencyP99Ms: parseDurationToMs(metric.latencyP99),
                    clusterInfo,
                    isClickable:
                        serviceName !== 'unknown' && namespace !== 'unknown',
                };
            });
    };

    // Process data
    const inboundServices = useMemo(
        () => processAggregatedMetrics(inbound, 'inbound'),
        [inbound]
    );
    const outboundServices = useMemo(
        () => processAggregatedMetrics(outbound, 'outbound'),
        [outbound]
    );

    // Sorting logic
    const sortData = (
        data: ServiceRowData[],
        field: SortField,
        direction: SortDirection
    ): ServiceRowData[] => {
        if (!field || !direction) return data;

        return [...data].sort((a, b) => {
            let aVal: any;
            let bVal: any;

            switch (field) {
                case 'service':
                    aVal = `${a.serviceName}.${a.namespace}`;
                    bVal = `${b.serviceName}.${b.namespace}`;
                    break;
                case 'requestRate':
                    aVal = a.requestRate;
                    bVal = b.requestRate;
                    break;
                case 'successRate':
                    aVal = a.successRate;
                    bVal = b.successRate;
                    break;
                case 'latencyP99':
                    aVal = a.latencyP99Ms;
                    bVal = b.latencyP99Ms;
                    break;
                default:
                    return 0;
            }

            if (aVal < bVal) return direction === 'asc' ? -1 : 1;
            if (aVal > bVal) return direction === 'asc' ? 1 : -1;
            return 0;
        });
    };

    // Handle sort click
    const handleSort = (field: SortField) => {
        if (sortState.field === field) {
            const newDirection =
                sortState.direction === 'asc'
                    ? 'desc'
                    : sortState.direction === 'desc'
                      ? null
                      : 'asc';
            setSortState({
                field: newDirection ? field : null,
                direction: newDirection,
            });
        } else {
            setSortState({ field, direction: 'desc' });
        }
    };

    // Handle service click
    const handleServiceClick = (serviceName: string, namespace: string) => {
        if (serviceName !== 'unknown' && namespace !== 'unknown') {
            navigate(`/services/${namespace}/${serviceName}`);
        }
    };

    // Render sort icon
    const renderSortIcon = (field: SortField) => {
        if (sortState.field !== field) {
            return <ArrowUpDown className="w-3 h-3 opacity-40" />;
        }
        return sortState.direction === 'asc' ? (
            <ArrowUp className="w-3 h-3" />
        ) : (
            <ArrowDown className="w-3 h-3" />
        );
    };

    // Render table section
    const renderTableSection = (title: string, data: ServiceRowData[]) => {
        const sortedData = sortData(
            data,
            sortState.field!,
            sortState.direction
        );

        if (data.length === 0) {
            return (
                <div className="mb-8">
                    <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                        {title === 'Inbound' ? '← ' : '→ '}
                        {title} ({data.length})
                    </h3>
                    <div className="text-muted-foreground text-sm">
                        No {title.toLowerCase()} connections found
                    </div>
                </div>
            );
        }

        return (
            <div className="mb-8">
                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                    {title === 'Inbound' ? '← ' : '→ '}
                    {title} ({data.length})
                </h3>
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead
                                className="cursor-pointer hover:bg-muted/50 select-none"
                                onClick={() => handleSort('service')}
                            >
                                <div className="flex items-center gap-1">
                                    Service
                                    {renderSortIcon('service')}
                                </div>
                            </TableHead>
                            <TableHead
                                className="cursor-pointer hover:bg-muted/50 select-none text-right"
                                onClick={() => handleSort('requestRate')}
                            >
                                <div className="flex items-center gap-1 justify-end">
                                    Request Rate
                                    {renderSortIcon('requestRate')}
                                </div>
                            </TableHead>
                            <TableHead
                                className="cursor-pointer hover:bg-muted/50 select-none text-right"
                                onClick={() => handleSort('successRate')}
                            >
                                <div className="flex items-center gap-1 justify-end">
                                    Success Rate
                                    {renderSortIcon('successRate')}
                                </div>
                            </TableHead>
                            <TableHead
                                className="cursor-pointer hover:bg-muted/50 select-none text-right"
                                onClick={() => handleSort('latencyP99')}
                            >
                                <div className="flex items-center gap-1 justify-end">
                                    P99 Latency
                                    {renderSortIcon('latencyP99')}
                                </div>
                            </TableHead>
                            <TableHead>Clusters</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {sortedData.map((service, index) => (
                            <TableRow
                                key={`${service.serviceName}-${service.namespace}-${index}`}
                            >
                                <TableCell>
                                    {service.isClickable ? (
                                        <button
                                            onClick={() =>
                                                handleServiceClick(
                                                    service.serviceName,
                                                    service.namespace
                                                )
                                            }
                                            className="text-blue-600 hover:text-blue-800 hover:underline text-left"
                                        >
                                            {service.serviceName}.
                                            {service.namespace}
                                        </button>
                                    ) : (
                                        <span className="text-muted-foreground">
                                            {service.serviceName}.
                                            {service.namespace}
                                        </span>
                                    )}
                                </TableCell>
                                <TableCell className="text-right">
                                    {formatRequestRate(service.requestRate)}
                                </TableCell>
                                <TableCell
                                    className={`text-right ${getSuccessRateColor(service.successRate)}`}
                                >
                                    {service.successRate.toFixed(1)}%
                                </TableCell>
                                <TableCell className="text-right">
                                    {service.latencyP99}
                                </TableCell>
                                <TableCell className="text-sm text-muted-foreground">
                                    {service.clusterInfo}
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>
        );
    };

    return (
        <TooltipProvider>
            <div className="space-y-6">
                {renderTableSection('Inbound', inboundServices)}
                {renderTableSection('Outbound', outboundServices)}
            </div>
        </TooltipProvider>
    );
};
