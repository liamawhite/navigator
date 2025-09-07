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

import React, { useMemo, useState, useEffect } from 'react';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '../components/ui/card';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '../components/ui/table';
import {
    Waypoints,
    ArrowRight,
    Activity,
    AlertTriangle,
    Loader2,
    RefreshCw,
    AlertCircle,
} from 'lucide-react';
import { useServiceGraphMetrics } from '../hooks/useServiceGraphMetrics';
import { Badge } from '../components/ui/badge';
import { Button } from '../components/ui/button';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '../components/ui/select';
import { Navbar } from '../components/Navbar';
import { serviceApi } from '../utils/api';
import type { v1alpha1ClusterSyncInfo } from '../types/generated/openapi-cluster_registry';

export const TopologyPage: React.FC = () => {
    // Check if any clusters have metrics capability
    const [hasMetricsCapability, setHasMetricsCapability] = useState<
        boolean | null
    >(null);

    // Refresh interval options (in milliseconds, 0 = manual)
    const refreshIntervals = [
        { value: 0, label: 'Manual' },
        { value: 5000, label: '5 seconds' },
        { value: 10000, label: '10 seconds' },
        { value: 30000, label: '30 seconds' },
        { value: 60000, label: '1 minute' },
        { value: 300000, label: '5 minutes' },
    ];

    const [refreshInterval, setRefreshInterval] = useState(0); // Default to manual

    // Stable time range that only updates periodically
    const [timeRange, setTimeRange] = useState(() => {
        const endTime = new Date().toISOString();
        const startTime = new Date(Date.now() - 5 * 60 * 1000).toISOString(); // 5 minutes ago
        return { startTime, endTime };
    });

    const checkMetricsCapability = async () => {
        try {
            const clusterData = await serviceApi.listClusters();
            const hasAnyMetrics = clusterData.some(
                (cluster: v1alpha1ClusterSyncInfo) => cluster.metricsEnabled
            );
            setHasMetricsCapability(hasAnyMetrics);
        } catch (error) {
            console.error('Failed to check metrics capability:', error);
            setHasMetricsCapability(false);
        }
    };

    useEffect(() => {
        checkMetricsCapability();
    }, []);

    // Update time range based on selected refresh interval (skip if manual)
    useEffect(() => {
        if (refreshInterval === 0) return; // Manual refresh, no automatic updates

        const interval = setInterval(() => {
            const endTime = new Date().toISOString();
            const startTime = new Date(
                Date.now() - 5 * 60 * 1000
            ).toISOString(); // 5 minutes ago
            setTimeRange({ startTime, endTime });
        }, refreshInterval);

        return () => clearInterval(interval);
    }, [refreshInterval]);

    const {
        data: servicePairs,
        isLoading,
        error,
    } = useServiceGraphMetrics(
        {
            startTime: timeRange.startTime,
            endTime: timeRange.endTime,
        },
        refreshInterval === 0 ? false : refreshInterval // Disable automatic refetch for manual mode
    );

    // Debug logging to see what data we're getting
    useEffect(() => {
        if (servicePairs && servicePairs.length > 0) {
            console.log('Service pairs data:', servicePairs);
        }
    }, [servicePairs]);

    const stats = useMemo(() => {
        if (!servicePairs)
            return { totalPairs: 0, totalRequests: 0, totalErrors: 0 };

        const totalRequests = servicePairs.reduce(
            (sum, pair) => sum + (pair.requestRate || 0),
            0
        );
        const totalErrors = servicePairs.reduce(
            (sum, pair) => sum + (pair.errorRate || 0),
            0
        );

        return {
            totalPairs: servicePairs.length,
            totalRequests: Math.round(totalRequests * 100) / 100,
            totalErrors: Math.round(totalErrors * 100) / 100,
        };
    }, [servicePairs]);

    // Manual refresh function
    const handleManualRefresh = () => {
        const endTime = new Date().toISOString();
        const startTime = new Date(Date.now() - 5 * 60 * 1000).toISOString(); // 5 minutes ago
        setTimeRange({ startTime, endTime });
    };

    const formatRate = (rate?: number) => {
        if (!rate || rate === 0) return '0';
        return rate.toFixed(2);
    };

    const getErrorRateBadgeVariant = (errorRate?: number) => {
        if (!errorRate || errorRate === 0) return 'secondary';
        if (errorRate < 1) return 'default';
        if (errorRate < 5) return 'destructive';
        return 'destructive';
    };

    const getServiceDisplayName = (
        serviceName?: string,
        namespace?: string,
        cluster?: string
    ) => {
        if (serviceName) return serviceName;
        if (namespace && cluster) return `${namespace} (${cluster})`;
        if (namespace) return namespace;
        if (cluster) return cluster;
        return 'Unknown Service';
    };

    // If we haven't checked yet, show loading
    if (hasMetricsCapability === null) {
        return (
            <div className="min-h-screen bg-background">
                <Navbar />
                <div className="container mx-auto px-4 py-8">
                    <div className="flex items-center justify-center min-h-96">
                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                    </div>
                </div>
            </div>
        );
    }

    // If no clusters have metrics capability, show the no-metrics state
    if (!hasMetricsCapability) {
        return (
            <div className="min-h-screen bg-background">
                <Navbar />
                <div className="container mx-auto px-4 py-8">
                    <div className="flex flex-col items-center justify-center min-h-96 text-center space-y-6">
                        <div className="flex items-center justify-center w-16 h-16 bg-orange-100 dark:bg-orange-900/20 rounded-full">
                            <AlertCircle className="h-8 w-8 text-orange-600 dark:text-orange-400" />
                        </div>
                        <div className="space-y-2">
                            <h1 className="text-2xl font-semibold text-foreground">
                                No Metrics Available
                            </h1>
                            <p className="text-muted-foreground max-w-md">
                                The topology view requires metrics data from
                                connected clusters. Currently, no edge services
                                are configured with metrics capabilities.
                            </p>
                        </div>
                        <div className="bg-muted p-4 rounded-lg max-w-md">
                            <p className="text-sm text-muted-foreground">
                                To enable topology view, configure your edge
                                services with metrics providers (e.g.,
                                Prometheus) and restart the connections.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            <Navbar />
            <div className="container mx-auto px-4 py-6">
                <div className="mb-6">
                    <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-3">
                            <Waypoints className="h-6 w-6" />
                            <h1 className="text-2xl font-bold">
                                Service Graph
                            </h1>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="text-sm text-muted-foreground">
                                Refresh:
                            </span>
                            <Select
                                value={refreshInterval.toString()}
                                onValueChange={(value) =>
                                    setRefreshInterval(Number(value))
                                }
                            >
                                <SelectTrigger className="w-32">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {refreshIntervals.map((interval) => (
                                        <SelectItem
                                            key={interval.value}
                                            value={interval.value.toString()}
                                        >
                                            {interval.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            {refreshInterval === 0 && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={handleManualRefresh}
                                    disabled={isLoading}
                                    className="px-3 cursor-pointer"
                                >
                                    <RefreshCw
                                        className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`}
                                    />
                                </Button>
                            )}
                        </div>
                    </div>
                    <p className="text-muted-foreground">
                        Service-to-service communication metrics across your
                        mesh
                    </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium">
                                Service Pairs
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">
                                {stats.totalPairs}
                            </div>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium">
                                Total Request Rate
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">
                                {stats.totalRequests}{' '}
                                <span className="text-sm font-normal text-muted-foreground">
                                    req/s
                                </span>
                            </div>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium">
                                Total Error Rate
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">
                                {stats.totalErrors}{' '}
                                <span className="text-sm font-normal text-muted-foreground">
                                    err/s
                                </span>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Activity className="h-5 w-5" />
                            Service Communication Graph
                        </CardTitle>
                        <CardDescription>
                            Service-to-service metrics (last 5 minutes)
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        {isLoading && (
                            <div className="flex items-center justify-center py-8">
                                <Loader2 className="h-6 w-6 animate-spin" />
                                <span className="ml-2">
                                    Loading service graph metrics...
                                </span>
                            </div>
                        )}

                        {error && (
                            <div className="flex items-center justify-center py-8 text-muted-foreground">
                                <AlertTriangle className="h-5 w-5 mr-2" />
                                <span>
                                    Failed to load service graph metrics
                                </span>
                            </div>
                        )}

                        {!isLoading && !error && servicePairs && (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Source</TableHead>
                                        <TableHead className="w-12"></TableHead>
                                        <TableHead>Destination</TableHead>
                                        <TableHead className="text-right">
                                            Request Rate
                                        </TableHead>
                                        <TableHead className="text-right">
                                            Error Rate
                                        </TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {servicePairs.length === 0 ? (
                                        <TableRow>
                                            <TableCell
                                                colSpan={5}
                                                className="text-center py-8 text-muted-foreground"
                                            >
                                                No service communication
                                                detected
                                            </TableCell>
                                        </TableRow>
                                    ) : (
                                        servicePairs.map((pair, index) => (
                                            <TableRow key={index}>
                                                <TableCell>
                                                    <div className="space-y-1">
                                                        <div className="font-medium">
                                                            {getServiceDisplayName(
                                                                pair.sourceService,
                                                                pair.sourceNamespace,
                                                                pair.sourceCluster
                                                            )}
                                                        </div>
                                                        <div className="text-sm text-muted-foreground">
                                                            {pair.sourceNamespace && (
                                                                <span>
                                                                    {
                                                                        pair.sourceNamespace
                                                                    }
                                                                </span>
                                                            )}
                                                            {pair.sourceCluster &&
                                                                pair.sourceCluster !==
                                                                    pair.sourceNamespace && (
                                                                    <span>
                                                                        {' '}
                                                                        •{' '}
                                                                        {
                                                                            pair.sourceCluster
                                                                        }
                                                                    </span>
                                                                )}
                                                        </div>
                                                    </div>
                                                </TableCell>
                                                <TableCell className="text-center">
                                                    <ArrowRight className="h-4 w-4 text-muted-foreground" />
                                                </TableCell>
                                                <TableCell>
                                                    <div className="space-y-1">
                                                        <div className="font-medium">
                                                            {getServiceDisplayName(
                                                                pair.destinationService,
                                                                pair.destinationNamespace,
                                                                pair.destinationCluster
                                                            )}
                                                        </div>
                                                        <div className="text-sm text-muted-foreground">
                                                            {pair.destinationNamespace && (
                                                                <span>
                                                                    {
                                                                        pair.destinationNamespace
                                                                    }
                                                                </span>
                                                            )}
                                                            {pair.destinationCluster &&
                                                                pair.destinationCluster !==
                                                                    pair.destinationNamespace && (
                                                                    <span>
                                                                        {' '}
                                                                        •{' '}
                                                                        {
                                                                            pair.destinationCluster
                                                                        }
                                                                    </span>
                                                                )}
                                                        </div>
                                                    </div>
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <Badge variant="secondary">
                                                        {formatRate(
                                                            pair.requestRate
                                                        )}{' '}
                                                        req/s
                                                    </Badge>
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <Badge
                                                        variant={getErrorRateBadgeVariant(
                                                            pair.errorRate
                                                        )}
                                                    >
                                                        {formatRate(
                                                            pair.errorRate
                                                        )}{' '}
                                                        err/s
                                                    </Badge>
                                                </TableCell>
                                            </TableRow>
                                        ))
                                    )}
                                </TableBody>
                            </Table>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
};
