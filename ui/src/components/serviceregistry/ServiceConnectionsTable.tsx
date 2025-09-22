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

import React, { useState, useEffect, useMemo, useCallback } from 'react';
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
    ChevronLeft,
    ChevronDown,
} from 'lucide-react';
import type {
    v1alpha1AggregatedServicePairMetrics,
    v1alpha1ServicePairMetrics,
} from '../../types/generated/openapi-metrics_service';

interface ServiceConnectionsTableProps {
    inbound: v1alpha1AggregatedServicePairMetrics[];
    outbound: v1alpha1AggregatedServicePairMetrics[];
}

interface ConnectionRowData {
    service: string;
    namespace: string;
    cluster: string;
    sourceCluster: string;
    destinationCluster: string;
    requestRate: number;
    successRate: number;
    latencyP99: string | undefined;
    latencyP99Ms: number;
    isClickable: boolean;
}

interface ServiceRowData {
    serviceName: string;
    namespace: string;
    requestRate: number;
    successRate: number;
    latencyP99: string | undefined;
    latencyP99Ms: number;
    clusterPairCount: number;
    servicePairs: ConnectionRowData[];
}

type SortField =
    | 'service'
    | 'cluster'
    | 'requestRate'
    | 'successRate'
    | 'latencyP99';
type SortDirection = 'asc' | 'desc' | null;

interface SortState {
    field: SortField | null;
    direction: SortDirection;
}

// Constants for better maintainability
const SORT_ICON_SIZE = 'w-3 h-3';
const TRANSITION_DURATION = 'duration-200';
const OPACITY_VALUES = {
    INACTIVE: 'opacity-40',
    HOVER: 'opacity-70',
    ACTIVE: 'opacity-100',
} as const;

const MIN_REQUEST_RATE = 0.01;
const SUCCESS_RATE_EXCELLENT = 99;
const SUCCESS_RATE_GOOD = 95;

const STORAGE_KEY = 'serviceConnections.sort';
const DEFAULT_SORT_FIELD = 'requestRate';
const DEFAULT_SORT_DIRECTION = 'desc';

export const ServiceConnectionsTable: React.FC<
    ServiceConnectionsTableProps
> = ({ inbound, outbound }) => {
    const navigate = useNavigate();
    const [expandedServices, setExpandedServices] = useState<Set<string>>(
        new Set()
    );

    // Load sort preferences from localStorage with fallback to default
    const loadSortState = (): SortState => {
        try {
            const saved = localStorage.getItem(STORAGE_KEY);
            if (saved) {
                const parsed = JSON.parse(saved);
                // Validate the loaded state
                if (parsed.field && parsed.direction) {
                    return parsed;
                }
            }
        } catch (error) {
            console.warn('Failed to load sort preferences:', error);
        }
        return { field: DEFAULT_SORT_FIELD, direction: DEFAULT_SORT_DIRECTION };
    };

    const [sortState, setSortState] = useState<SortState>(() =>
        loadSortState()
    );

    // Save sort preferences to localStorage
    useEffect(() => {
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(sortState));
        } catch (error) {
            console.warn('Failed to save sort preference:', error);
        }
    }, [sortState]);

    const formatRequestRate = (rate: number): string => {
        if (rate >= 1) {
            return rate.toFixed(0);
        } else {
            return rate.toFixed(2);
        }
    };

    const formatSuccessRate = (rate: number): string => {
        if (rate >= 10) {
            return `${rate.toFixed(1)}%`;
        } else {
            return `${rate.toFixed(2)}%`;
        }
    };

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

    const getSuccessRateColor = (rate: number): string => {
        if (rate >= SUCCESS_RATE_EXCELLENT) return 'text-green-600';
        if (rate >= SUCCESS_RATE_GOOD) return 'text-amber-600';
        return 'text-red-600';
    };

    const processConnections = useCallback(
        (
            connections: v1alpha1ServicePairMetrics[],
            type: 'inbound' | 'outbound'
        ): ConnectionRowData[] => {
            return connections
                .filter((conn) => (conn.requestRate || 0) >= MIN_REQUEST_RATE)
                .map((conn) => {
                    const service =
                        type === 'inbound'
                            ? conn.sourceService || 'unknown'
                            : conn.destinationService || 'unknown';
                    const ns =
                        type === 'inbound'
                            ? conn.sourceNamespace || 'unknown'
                            : conn.destinationNamespace || 'unknown';
                    const cluster =
                        type === 'inbound'
                            ? conn.sourceCluster || 'unknown'
                            : conn.destinationCluster || 'unknown';
                    const sourceCluster = conn.sourceCluster || 'unknown';
                    const destinationCluster =
                        conn.destinationCluster || 'unknown';

                    const requestRate = conn.requestRate || 0;
                    const errorRate = conn.errorRate || 0;
                    const latencyP99 = conn.latencyP99;
                    const successRate =
                        requestRate > 0
                            ? ((requestRate - errorRate) / requestRate) * 100
                            : 100;

                    return {
                        service,
                        namespace: ns,
                        cluster,
                        sourceCluster,
                        destinationCluster,
                        requestRate,
                        successRate,
                        latencyP99,
                        latencyP99Ms: parseDurationToMs(latencyP99),
                        isClickable: service !== 'unknown' && ns !== 'unknown',
                    };
                });
        },
        []
    );

    const processServiceAggregations = useCallback(
        (
            connections: v1alpha1AggregatedServicePairMetrics[],
            type: 'inbound' | 'outbound'
        ): ServiceRowData[] => {
            const filteredConnections = connections.filter(
                (conn) => (conn.requestRate || 0) >= MIN_REQUEST_RATE
            );

            // Data is already aggregated, just convert to ServiceRowData format
            return filteredConnections.map((agg) => {
                const totalRequestRate = agg.requestRate || 0;
                const totalErrorRate = agg.errorRate || 0;
                const successRate =
                    totalRequestRate > 0
                        ? ((totalRequestRate - totalErrorRate) /
                              totalRequestRate) *
                          100
                        : 100;

                // Get the service name and namespace based on inbound/outbound direction
                const serviceName =
                    type === 'inbound'
                        ? agg.sourceService
                        : agg.destinationService;
                const namespace =
                    type === 'inbound'
                        ? agg.sourceNamespace
                        : agg.destinationNamespace;

                // Process the detailed breakdown for drill-down
                const servicePairs = agg.detailedBreakdown
                    ? processConnections(agg.detailedBreakdown, type)
                    : [];

                return {
                    serviceName: serviceName || 'unknown',
                    namespace: namespace || 'unknown',
                    requestRate: totalRequestRate,
                    successRate,
                    latencyP99: agg.latencyP99 || '0s',
                    latencyP99Ms: parseDurationToMs(agg.latencyP99 || '0s'),
                    clusterPairCount: servicePairs.length,
                    servicePairs: servicePairs,
                };
            });
        },
        [processConnections]
    );

    const handleSort = (field: SortField) => {
        let newDirection: SortDirection;
        if (sortState.field === field) {
            if (sortState.direction === 'asc') {
                newDirection = 'desc';
            } else if (sortState.direction === 'desc') {
                newDirection = null;
            } else {
                newDirection = 'asc';
            }
        } else {
            newDirection = 'asc';
        }

        setSortState({
            field: newDirection ? field : null,
            direction: newDirection,
        });
    };

    const sortServiceAggregations = useCallback(
        (
            serviceData: ServiceRowData[],
            sortState: SortState
        ): ServiceRowData[] => {
            if (!sortState.field || !sortState.direction) {
                // Default sort: RPS descending (highest traffic first)
                return [...serviceData].sort((a, b) => {
                    const rpsComparison = b.requestRate - a.requestRate;
                    // If RPS is equal, sort by service name for stability
                    if (rpsComparison === 0) {
                        return a.serviceName.localeCompare(b.serviceName);
                    }
                    return rpsComparison;
                });
            }

            return [...serviceData].sort((a, b) => {
                const multiplier = sortState.direction === 'asc' ? 1 : -1;
                let primaryComparison = 0;

                switch (sortState.field) {
                    case 'service':
                        primaryComparison =
                            a.serviceName.localeCompare(b.serviceName) *
                            multiplier;
                        break;
                    case 'cluster':
                        // For service aggregations, treat "cluster" sort as namespace sort
                        primaryComparison =
                            a.namespace.localeCompare(b.namespace) * multiplier;
                        break;
                    case 'requestRate':
                        primaryComparison =
                            (a.requestRate - b.requestRate) * multiplier;
                        break;
                    case 'successRate':
                        primaryComparison =
                            (a.successRate - b.successRate) * multiplier;
                        break;
                    case 'latencyP99':
                        primaryComparison =
                            (a.latencyP99Ms - b.latencyP99Ms) * multiplier;
                        break;
                    default:
                        primaryComparison = 0;
                }

                // If primary comparison is equal, use secondary sort by service name for stability
                if (primaryComparison === 0) {
                    return a.serviceName.localeCompare(b.serviceName);
                }

                return primaryComparison;
            });
        },
        []
    );

    const getSortIcon = (field: SortField, sortState: SortState) => {
        const baseClasses = `${SORT_ICON_SIZE} transition-all ${TRANSITION_DURATION}`;

        if (sortState.field !== field) {
            return (
                <ArrowUpDown
                    className={`${baseClasses} ${OPACITY_VALUES.INACTIVE}`}
                />
            );
        }

        if (sortState.direction === 'asc') {
            return (
                <ArrowUp
                    className={`${baseClasses} ${OPACITY_VALUES.ACTIVE}`}
                />
            );
        } else if (sortState.direction === 'desc') {
            return (
                <ArrowDown
                    className={`${baseClasses} ${OPACITY_VALUES.ACTIVE}`}
                />
            );
        } else {
            return (
                <ArrowUpDown
                    className={`${baseClasses} ${OPACITY_VALUES.INACTIVE}`}
                />
            );
        }
    };

    const SortableHeader = ({
        field,
        children,
        className = '',
    }: {
        field: SortField;
        children: React.ReactNode;
        className?: string;
    }) => {
        const isActive =
            sortState.field === field && sortState.direction !== null;

        const getTooltipText = () => {
            const sortStates = {
                asc: { current: 'ascending', next: 'descending' },
                desc: { current: 'descending', next: 'default' },
            };

            if (!isActive) {
                return `Click to sort by ${children} (ascending first)`;
            }

            const state = sortStates[sortState.direction];
            return `Currently sorted ${state.current}. Click to sort ${state.next}.`;
        };

        // Determine alignment based on className
        const isRightAligned = className.includes('text-right');
        const flexClass = isRightAligned ? 'justify-end' : 'justify-start';

        return (
            <TooltipProvider>
                <Tooltip>
                    <TooltipTrigger asChild>
                        <TableHead
                            className={`group cursor-pointer hover:bg-muted/50 select-none text-xs text-gray-500 dark:text-gray-500 font-medium py-2 transition-colors ${className}`}
                            onClick={() => handleSort(field)}
                            tabIndex={0}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                    e.preventDefault();
                                    handleSort(field);
                                }
                            }}
                            role="button"
                            aria-label={`Sort by ${field} ${
                                isActive
                                    ? `(currently ${sortState.direction}ending)`
                                    : ''
                            }`}
                        >
                            <div
                                className={`flex items-center gap-1 ${flexClass}`}
                            >
                                <span>{children}</span>
                                <div
                                    className={`transition-all duration-150 ${
                                        isActive
                                            ? OPACITY_VALUES.ACTIVE
                                            : `${OPACITY_VALUES.INACTIVE} group-hover:${OPACITY_VALUES.HOVER}`
                                    }`}
                                >
                                    {getSortIcon(field, sortState)}
                                </div>
                            </div>
                        </TableHead>
                    </TooltipTrigger>
                    <TooltipContent>
                        <p className="text-xs">{getTooltipText()}</p>
                    </TooltipContent>
                </Tooltip>
            </TooltipProvider>
        );
    };

    const toggleService = (serviceKey: string) => {
        const newExpanded = new Set(expandedServices);
        if (newExpanded.has(serviceKey)) {
            newExpanded.delete(serviceKey);
        } else {
            newExpanded.add(serviceKey);
        }
        setExpandedServices(newExpanded);
    };

    const handleServiceClick = (service: string, namespace: string) => {
        if (service !== 'unknown' && namespace !== 'unknown') {
            navigate(`/services/${namespace}:${service}`);
        }
    };

    const sortedData = useMemo(() => {
        const inboundAggregated = processServiceAggregations(
            inbound,
            'inbound'
        );
        const outboundAggregated = processServiceAggregations(
            outbound,
            'outbound'
        );
        return {
            inbound: sortServiceAggregations(inboundAggregated, sortState),
            outbound: sortServiceAggregations(outboundAggregated, sortState),
        };
    }, [
        inbound,
        outbound,
        sortState,
        processServiceAggregations,
        sortServiceAggregations,
    ]);

    const { inbound: inboundConnections, outbound: outboundConnections } =
        sortedData;

    if (inboundConnections.length === 0 && outboundConnections.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
                <p className="text-center">No service connections found</p>
                <p className="text-sm text-center mt-2">
                    This service has no inbound or outbound traffic in the
                    selected time range
                </p>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {/* Section Headers */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <ArrowRight className="w-4 h-4 text-blue-600" />
                    <h3 className="text-sm font-medium">
                        Inbound Connections ({inboundConnections.length})
                    </h3>
                </div>
                <div className="flex items-center gap-2">
                    <h3 className="text-sm font-medium">
                        Outbound Connections ({outboundConnections.length})
                    </h3>
                    <ArrowRight className="w-4 h-4 text-green-600" />
                </div>
            </div>

            {/* Tables */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 relative">
                {/* Divider line for desktop */}
                <div className="hidden lg:block absolute left-1/2 top-0 bottom-0 w-px bg-border transform -translate-x-1/2"></div>
                <div>
                    {inboundConnections.length > 0 ? (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <SortableHeader field="service">
                                        Source Service
                                    </SortableHeader>
                                    <SortableHeader
                                        field="latencyP99"
                                        className="text-right"
                                    >
                                        P99
                                    </SortableHeader>
                                    <SortableHeader
                                        field="requestRate"
                                        className="text-right"
                                    >
                                        Throughput
                                    </SortableHeader>
                                    <SortableHeader
                                        field="successRate"
                                        className="text-right"
                                    >
                                        Success
                                    </SortableHeader>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {inboundConnections.map(
                                    (service, serviceIndex) => {
                                        const serviceKey = `inbound-${service.serviceName}-${service.namespace}`;
                                        const isExpanded =
                                            expandedServices.has(serviceKey);
                                        const hasMultipleClusters =
                                            service.clusterPairCount > 1;

                                        return (
                                            <React.Fragment key={serviceIndex}>
                                                {/* Service summary row */}
                                                <TableRow
                                                    className={`${
                                                        hasMultipleClusters
                                                            ? 'cursor-pointer hover:bg-muted/50'
                                                            : ''
                                                    } ${isExpanded ? 'bg-muted/30' : ''}`}
                                                    onClick={
                                                        hasMultipleClusters
                                                            ? () =>
                                                                  toggleService(
                                                                      serviceKey
                                                                  )
                                                            : undefined
                                                    }
                                                >
                                                    <TableCell className="font-medium">
                                                        <div className="flex items-center gap-2">
                                                            {hasMultipleClusters && (
                                                                <div className="w-4 h-4 flex items-center justify-center">
                                                                    {isExpanded ? (
                                                                        <ChevronDown className="w-4 h-4" />
                                                                    ) : (
                                                                        <ChevronRight className="w-4 h-4" />
                                                                    )}
                                                                </div>
                                                            )}
                                                            {service.serviceName !==
                                                                'unknown' &&
                                                            service.namespace !==
                                                                'unknown' ? (
                                                                <button
                                                                    onClick={(
                                                                        e
                                                                    ) => {
                                                                        e.stopPropagation();
                                                                        handleServiceClick(
                                                                            service.serviceName,
                                                                            service.namespace
                                                                        );
                                                                    }}
                                                                    className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline text-left hover:cursor-pointer"
                                                                >
                                                                    {
                                                                        service.serviceName
                                                                    }
                                                                    .
                                                                    {
                                                                        service.namespace
                                                                    }
                                                                </button>
                                                            ) : (
                                                                <span className="text-muted-foreground">
                                                                    {
                                                                        service.serviceName
                                                                    }
                                                                    .
                                                                    {
                                                                        service.namespace
                                                                    }
                                                                </span>
                                                            )}
                                                            {hasMultipleClusters && (
                                                                <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                                                                    {
                                                                        service.clusterPairCount
                                                                    }
                                                                </span>
                                                            )}
                                                        </div>
                                                    </TableCell>
                                                    <TableCell className="text-right font-mono text-sm">
                                                        {formatLatency(
                                                            service.latencyP99
                                                        )}
                                                    </TableCell>
                                                    <TableCell className="text-right font-mono text-sm">
                                                        {formatRequestRate(
                                                            service.requestRate
                                                        )}{' '}
                                                        rps
                                                    </TableCell>
                                                    <TableCell
                                                        className={`text-right font-mono text-sm ${getSuccessRateColor(service.successRate)}`}
                                                    >
                                                        {formatSuccessRate(
                                                            service.successRate
                                                        )}
                                                    </TableCell>
                                                </TableRow>

                                                {/* Expanded cluster pairs */}
                                                {hasMultipleClusters &&
                                                    isExpanded && (
                                                        <>
                                                            <TableRow className="bg-muted/10 border-b">
                                                                <TableCell
                                                                    colSpan={4}
                                                                >
                                                                    <div className="flex items-center gap-4 px-2 py-1">
                                                                        <span className="text-xs font-medium text-muted-foreground flex-1">
                                                                            Clusters
                                                                        </span>
                                                                        <span className="text-xs font-medium text-muted-foreground w-16 text-right">
                                                                            P99
                                                                        </span>
                                                                        <span className="text-xs font-medium text-muted-foreground w-20 text-right">
                                                                            Throughput
                                                                        </span>
                                                                        <span className="text-xs font-medium text-muted-foreground w-16 text-right">
                                                                            Success
                                                                        </span>
                                                                    </div>
                                                                </TableCell>
                                                            </TableRow>
                                                            {service.servicePairs.map(
                                                                (
                                                                    pair,
                                                                    pairIndex
                                                                ) => (
                                                                    <TableRow
                                                                        key={
                                                                            pairIndex
                                                                        }
                                                                        className="bg-muted/25"
                                                                    >
                                                                        <TableCell
                                                                            colSpan={
                                                                                4
                                                                            }
                                                                        >
                                                                            <div className="flex items-center gap-4 px-2">
                                                                                <span className="text-muted-foreground text-sm flex-1">
                                                                                    {pair.sourceCluster ||
                                                                                        'unknown'}{' '}
                                                                                    →{' '}
                                                                                    {pair.destinationCluster ||
                                                                                        'unknown'}
                                                                                </span>
                                                                                <span className="text-right font-mono text-xs w-16">
                                                                                    {formatLatency(
                                                                                        pair.latencyP99
                                                                                    )}
                                                                                </span>
                                                                                <span className="text-right font-mono text-xs w-20">
                                                                                    {formatRequestRate(
                                                                                        pair.requestRate
                                                                                    )}{' '}
                                                                                    rps
                                                                                </span>
                                                                                <span
                                                                                    className={`text-right font-mono text-xs w-16 ${getSuccessRateColor(pair.successRate)}`}
                                                                                >
                                                                                    {formatSuccessRate(
                                                                                        pair.successRate
                                                                                    )}
                                                                                </span>
                                                                            </div>
                                                                        </TableCell>
                                                                    </TableRow>
                                                                )
                                                            )}
                                                        </>
                                                    )}
                                            </React.Fragment>
                                        );
                                    }
                                )}
                            </TableBody>
                        </Table>
                    ) : (
                        <div className="flex items-center justify-center text-gray-600 dark:text-gray-400 h-full min-h-24">
                            No Inbound Connections
                        </div>
                    )}
                </div>

                <div>
                    {outboundConnections.length > 0 ? (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <SortableHeader
                                        field="successRate"
                                        className="text-left"
                                    >
                                        Success
                                    </SortableHeader>
                                    <SortableHeader
                                        field="requestRate"
                                        className="text-left"
                                    >
                                        Throughput
                                    </SortableHeader>
                                    <SortableHeader
                                        field="latencyP99"
                                        className="text-left"
                                    >
                                        P99
                                    </SortableHeader>
                                    <SortableHeader
                                        field="service"
                                        className="text-right"
                                    >
                                        Destination Service
                                    </SortableHeader>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {outboundConnections.map(
                                    (service, serviceIndex) => {
                                        const serviceKey = `outbound-${service.serviceName}-${service.namespace}`;
                                        const isExpanded =
                                            expandedServices.has(serviceKey);
                                        const hasMultipleClusters =
                                            service.clusterPairCount > 1;

                                        return (
                                            <React.Fragment key={serviceIndex}>
                                                {/* Service summary row */}
                                                <TableRow
                                                    className={`${
                                                        hasMultipleClusters
                                                            ? 'cursor-pointer hover:bg-muted/50'
                                                            : ''
                                                    } ${isExpanded ? 'bg-muted/30' : ''}`}
                                                    onClick={
                                                        hasMultipleClusters
                                                            ? () =>
                                                                  toggleService(
                                                                      serviceKey
                                                                  )
                                                            : undefined
                                                    }
                                                >
                                                    <TableCell
                                                        className={`font-mono text-sm ${getSuccessRateColor(service.successRate)}`}
                                                    >
                                                        {formatSuccessRate(
                                                            service.successRate
                                                        )}
                                                    </TableCell>
                                                    <TableCell className="font-mono text-sm">
                                                        {formatRequestRate(
                                                            service.requestRate
                                                        )}{' '}
                                                        rps
                                                    </TableCell>
                                                    <TableCell className="font-mono text-sm">
                                                        {formatLatency(
                                                            service.latencyP99
                                                        )}
                                                    </TableCell>
                                                    <TableCell className="text-right font-medium">
                                                        <div className="flex items-center justify-end gap-2">
                                                            {hasMultipleClusters && (
                                                                <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                                                                    {
                                                                        service.clusterPairCount
                                                                    }
                                                                </span>
                                                            )}
                                                            {service.serviceName !==
                                                                'unknown' &&
                                                            service.namespace !==
                                                                'unknown' ? (
                                                                <button
                                                                    onClick={(
                                                                        e
                                                                    ) => {
                                                                        e.stopPropagation();
                                                                        handleServiceClick(
                                                                            service.serviceName,
                                                                            service.namespace
                                                                        );
                                                                    }}
                                                                    className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline text-right hover:cursor-pointer"
                                                                >
                                                                    {
                                                                        service.serviceName
                                                                    }
                                                                    .
                                                                    {
                                                                        service.namespace
                                                                    }
                                                                </button>
                                                            ) : (
                                                                <span className="text-muted-foreground">
                                                                    {
                                                                        service.serviceName
                                                                    }
                                                                    .
                                                                    {
                                                                        service.namespace
                                                                    }
                                                                </span>
                                                            )}
                                                            {hasMultipleClusters && (
                                                                <div className="w-4 h-4 flex items-center justify-center ml-1">
                                                                    {isExpanded ? (
                                                                        <ChevronDown className="w-4 h-4" />
                                                                    ) : (
                                                                        <ChevronLeft className="w-4 h-4" />
                                                                    )}
                                                                </div>
                                                            )}
                                                        </div>
                                                    </TableCell>
                                                </TableRow>

                                                {/* Expanded cluster pairs */}
                                                {hasMultipleClusters &&
                                                    isExpanded && (
                                                        <>
                                                            <TableRow className="bg-muted/10 border-b">
                                                                <TableCell
                                                                    colSpan={4}
                                                                >
                                                                    <div className="flex items-center gap-4 px-2 py-1">
                                                                        <span className="text-xs font-medium text-muted-foreground w-16">
                                                                            Success
                                                                        </span>
                                                                        <span className="text-xs font-medium text-muted-foreground w-20">
                                                                            Throughput
                                                                        </span>
                                                                        <span className="text-xs font-medium text-muted-foreground w-16">
                                                                            P99
                                                                        </span>
                                                                        <span className="text-xs font-medium text-muted-foreground flex-1">
                                                                            Clusters
                                                                        </span>
                                                                    </div>
                                                                </TableCell>
                                                            </TableRow>
                                                            {service.servicePairs.map(
                                                                (
                                                                    pair,
                                                                    pairIndex
                                                                ) => (
                                                                    <TableRow
                                                                        key={
                                                                            pairIndex
                                                                        }
                                                                        className="bg-muted/25"
                                                                    >
                                                                        <TableCell
                                                                            colSpan={
                                                                                4
                                                                            }
                                                                        >
                                                                            <div className="flex items-center gap-4 px-2">
                                                                                <span
                                                                                    className={`font-mono text-xs w-16 ${getSuccessRateColor(pair.successRate)}`}
                                                                                >
                                                                                    {formatSuccessRate(
                                                                                        pair.successRate
                                                                                    )}
                                                                                </span>
                                                                                <span className="font-mono text-xs w-20">
                                                                                    {formatRequestRate(
                                                                                        pair.requestRate
                                                                                    )}{' '}
                                                                                    rps
                                                                                </span>
                                                                                <span className="font-mono text-xs w-16">
                                                                                    {formatLatency(
                                                                                        pair.latencyP99
                                                                                    )}
                                                                                </span>
                                                                                <span className="text-muted-foreground text-sm flex-1">
                                                                                    {pair.sourceCluster ||
                                                                                        'unknown'}{' '}
                                                                                    →{' '}
                                                                                    {pair.destinationCluster ||
                                                                                        'unknown'}
                                                                                </span>
                                                                            </div>
                                                                        </TableCell>
                                                                    </TableRow>
                                                                )
                                                            )}
                                                        </>
                                                    )}
                                            </React.Fragment>
                                        );
                                    }
                                )}
                            </TableBody>
                        </Table>
                    ) : (
                        <div className="flex items-center justify-center text-gray-600 dark:text-gray-400 h-full min-h-24">
                            No Outbound Connections
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
