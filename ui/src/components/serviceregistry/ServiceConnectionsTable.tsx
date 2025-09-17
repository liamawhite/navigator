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

import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, TableBody, TableCell, TableRow } from '@/components/ui/table';
import { ArrowRight } from 'lucide-react';
import type { v1alpha1ServicePairMetrics } from '../../types/generated/openapi-metrics_service';

interface ServiceConnectionsTableProps {
    inbound: v1alpha1ServicePairMetrics[];
    outbound: v1alpha1ServicePairMetrics[];
}

interface ConnectionRowData {
    service: string;
    namespace: string;
    cluster: string;
    requestRate: number;
    successRate: number;
    latencyP99: string | undefined;
    isClickable: boolean;
}

export const ServiceConnectionsTable: React.FC<
    ServiceConnectionsTableProps
> = ({ inbound, outbound }) => {
    const navigate = useNavigate();

    const formatRequestRate = (rate: number): string => {
        if (rate >= 100) {
            return rate.toFixed(0);
        } else if (rate >= 10) {
            return rate.toFixed(1);
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

    // Configuration constants
    const MIN_REQUEST_RATE = 0.01;
    const SUCCESS_RATE_EXCELLENT = 99;
    const SUCCESS_RATE_GOOD = 95;

    const processConnections = (
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
                    requestRate,
                    successRate,
                    latencyP99,
                    isClickable: service !== 'unknown' && ns !== 'unknown',
                };
            })
            .sort((a, b) => {
                // Sort by namespace first, then by service name
                if (a.namespace !== b.namespace) {
                    return a.namespace.localeCompare(b.namespace);
                }
                return a.service.localeCompare(b.service);
            });
    };

    const handleServiceClick = (service: string, namespace: string) => {
        if (service !== 'unknown' && namespace !== 'unknown') {
            navigate(`/services/${namespace}:${service}`);
        }
    };

    const formatServiceName = (service: string, namespace: string): string => {
        if (service === 'unknown' && namespace === 'unknown') {
            return 'unknown';
        }
        return `${service}.${namespace}`;
    };

    const inboundConnections = processConnections(inbound, 'inbound');
    const outboundConnections = processConnections(outbound, 'outbound');

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
                            <TableBody>
                                {inboundConnections.map((conn, index) => (
                                    <TableRow key={index}>
                                        <TableCell>
                                            {conn.isClickable ? (
                                                <button
                                                    onClick={() =>
                                                        handleServiceClick(
                                                            conn.service,
                                                            conn.namespace
                                                        )
                                                    }
                                                    className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline text-left hover:cursor-pointer"
                                                >
                                                    {formatServiceName(
                                                        conn.service,
                                                        conn.namespace
                                                    )}
                                                </button>
                                            ) : (
                                                <span className="text-muted-foreground">
                                                    {formatServiceName(
                                                        conn.service,
                                                        conn.namespace
                                                    )}
                                                </span>
                                            )}
                                        </TableCell>
                                        <TableCell className="text-muted-foreground text-xs">
                                            {conn.cluster}
                                        </TableCell>
                                        <TableCell className="text-right font-mono text-sm px-1">
                                            {formatLatency(conn.latencyP99)}
                                        </TableCell>
                                        <TableCell className="text-right font-mono text-sm px-1">
                                            {formatRequestRate(
                                                conn.requestRate
                                            )}{' '}
                                            rps
                                        </TableCell>
                                        <TableCell
                                            className={`text-right font-mono text-sm px-1 ${getSuccessRateColor(
                                                conn.successRate
                                            )}`}
                                        >
                                            {formatSuccessRate(
                                                conn.successRate
                                            )}
                                        </TableCell>
                                    </TableRow>
                                ))}
                                {Array.from(
                                    {
                                        length: Math.max(
                                            0,
                                            outboundConnections.length -
                                                inboundConnections.length
                                        ),
                                    },
                                    (_, index) => (
                                        <TableRow
                                            key={`empty-${index}`}
                                            className="border-0"
                                        >
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                        </TableRow>
                                    )
                                )}
                                <TableRow>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2">
                                        Service
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2">
                                        Cluster
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-right">
                                        P99
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-right">
                                        Throughput
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-right">
                                        Success
                                    </TableCell>
                                </TableRow>
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
                            <TableBody>
                                {outboundConnections.map((conn, index) => (
                                    <TableRow key={index}>
                                        <TableCell
                                            className={`font-mono text-sm px-1 ${getSuccessRateColor(
                                                conn.successRate
                                            )}`}
                                        >
                                            {formatSuccessRate(
                                                conn.successRate
                                            )}
                                        </TableCell>
                                        <TableCell className="font-mono text-sm px-1">
                                            {formatRequestRate(
                                                conn.requestRate
                                            )}{' '}
                                            rps
                                        </TableCell>
                                        <TableCell className="font-mono text-sm px-1">
                                            {formatLatency(conn.latencyP99)}
                                        </TableCell>
                                        <TableCell className="text-muted-foreground text-xs text-right">
                                            {conn.cluster}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            {conn.isClickable ? (
                                                <button
                                                    onClick={() =>
                                                        handleServiceClick(
                                                            conn.service,
                                                            conn.namespace
                                                        )
                                                    }
                                                    className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline text-right hover:cursor-pointer"
                                                >
                                                    {formatServiceName(
                                                        conn.service,
                                                        conn.namespace
                                                    )}
                                                </button>
                                            ) : (
                                                <span className="text-muted-foreground">
                                                    {formatServiceName(
                                                        conn.service,
                                                        conn.namespace
                                                    )}
                                                </span>
                                            )}
                                        </TableCell>
                                    </TableRow>
                                ))}
                                {Array.from(
                                    {
                                        length: Math.max(
                                            0,
                                            inboundConnections.length -
                                                outboundConnections.length
                                        ),
                                    },
                                    (_, index) => (
                                        <TableRow
                                            key={`empty-${index}`}
                                            className="border-0"
                                        >
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                            <TableCell className="border-0">
                                                &nbsp;
                                            </TableCell>
                                        </TableRow>
                                    )
                                )}
                                <TableRow>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-left">
                                        Success
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-left">
                                        Throughput
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-left">
                                        P99
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-right">
                                        Cluster
                                    </TableCell>
                                    <TableCell className="text-xs text-gray-500 dark:text-gray-500 font-medium py-2 text-right">
                                        Service
                                    </TableCell>
                                </TableRow>
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
