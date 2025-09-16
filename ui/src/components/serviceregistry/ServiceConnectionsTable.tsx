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
    serviceName: string;
    namespace: string;
    inbound: v1alpha1ServicePairMetrics[];
    outbound: v1alpha1ServicePairMetrics[];
}

interface ConnectionRowData {
    service: string;
    namespace: string;
    cluster: string;
    requestRate: number;
    successRate: number;
    isClickable: boolean;
}

export const ServiceConnectionsTable: React.FC<
    ServiceConnectionsTableProps
> = ({ serviceName: _serviceName, namespace: _namespace, inbound, outbound }) => {
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
        if (rate >= 99.95) {
            return '100%';
        } else if (rate >= 10) {
            return `${rate.toFixed(1)}%`;
        } else {
            return `${rate.toFixed(2)}%`;
        }
    };

    const getSuccessRateColor = (rate: number): string => {
        if (rate >= 99) return 'text-green-600';
        if (rate >= 95) return 'text-amber-600';
        return 'text-red-600';
    };

    const processConnections = (
        connections: v1alpha1ServicePairMetrics[],
        type: 'inbound' | 'outbound'
    ): ConnectionRowData[] => {
        return connections
            .filter((conn) => (conn.requestRate || 0) >= 0.01)
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
                                                {conn.service}.{conn.namespace}
                                            </button>
                                        ) : (
                                            <span className="text-muted-foreground">
                                                {conn.service}.{conn.namespace}
                                            </span>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-muted-foreground text-xs">
                                        {conn.cluster}
                                    </TableCell>
                                    <TableCell className="text-right font-mono text-sm px-1">
                                        {formatRequestRate(conn.requestRate)}{' '}
                                        rps
                                    </TableCell>
                                    <TableCell
                                        className={`text-right font-mono text-sm px-1 ${getSuccessRateColor(
                                            conn.successRate
                                        )}`}
                                    >
                                        {formatSuccessRate(conn.successRate)}
                                    </TableCell>
                                </TableRow>
                            ))}
                            {inboundConnections.length === 0 && (
                                <TableRow>
                                    <TableCell
                                        colSpan={4}
                                        className="text-center text-muted-foreground py-8"
                                    >
                                        No inbound connections
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>

                <div>
                    <Table>
                        <TableBody>
                            {outboundConnections.map((conn, index) => (
                                <TableRow key={index}>
                                    <TableCell
                                        className={`font-mono text-sm px-1 ${getSuccessRateColor(
                                            conn.successRate
                                        )}`}
                                    >
                                        {formatSuccessRate(conn.successRate)}
                                    </TableCell>
                                    <TableCell className="font-mono text-sm px-1">
                                        {formatRequestRate(conn.requestRate)}{' '}
                                        rps
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
                                                {conn.service}.{conn.namespace}
                                            </button>
                                        ) : (
                                            <span className="text-muted-foreground">
                                                {conn.service}.{conn.namespace}
                                            </span>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))}
                            {outboundConnections.length === 0 && (
                                <TableRow>
                                    <TableCell
                                        colSpan={4}
                                        className="text-center text-muted-foreground py-8"
                                    >
                                        No outbound connections
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>
        </div>
    );
};
