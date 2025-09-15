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

import React, { useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Network, AlertCircle, Clock, RefreshCw } from 'lucide-react';
import { useServiceConnections } from '../../hooks/useServiceConnections';
import { useClusters } from '../../hooks/useClusters';
import { useMetricsContext, TIME_RANGES } from '../../contexts/MetricsContext';
import { ServiceConnectionsVisualization } from './ServiceConnectionsVisualization';
import { Button } from '@/components/ui/button';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { formatLastUpdated } from '@/lib/utils';

interface ServiceConnectionsCardProps {
    serviceName: string;
    namespace: string;
}

export const ServiceConnectionsCard: React.FC<ServiceConnectionsCardProps> = ({
    serviceName,
    namespace,
}) => {
    const {
        timeRange,
        lastUpdated,
        isRefreshing,
        setRefreshing,
        updateLastUpdated,
        triggerRefresh,
        setTimeRange,
    } = useMetricsContext();

    const {
        data: connections,
        isLoading,
        error,
    } = useServiceConnections(serviceName, namespace);

    const { data: clusters, isLoading: clustersLoading } = useClusters();

    // Track if initial refresh has been triggered to prevent memory leaks
    const hasTriggeredInitialRefresh = useRef(false);

    // Trigger initial refresh on mount (only once)
    React.useEffect(() => {
        if (!hasTriggeredInitialRefresh.current) {
            triggerRefresh();
            hasTriggeredInitialRefresh.current = true;
        }
    }, [triggerRefresh]);

    // Update refreshing state and last updated timestamp
    React.useEffect(() => {
        let isMounted = true;

        if (isLoading && !isRefreshing) {
            if (isMounted) {
                setRefreshing(true);
            }
        } else if (!isLoading && isRefreshing) {
            if (isMounted) {
                setRefreshing(false);
                // Update last updated timestamp when loading completes
                if (connections) {
                    updateLastUpdated();
                }
            }
        }

        return () => {
            isMounted = false;
        };
    }, [
        isLoading,
        isRefreshing,
        setRefreshing,
        connections,
        updateLastUpdated,
    ]);

    const hasAnyMetrics =
        clusters?.some((cluster) => cluster.metricsEnabled) ?? false;
    const showCollapsed = !clustersLoading && !hasAnyMetrics;

    return (
        <Card className={`mb-6 ${showCollapsed ? 'opacity-50' : ''}`}>
            <CardHeader>
                <CardTitle className="flex items-center justify-between -mb-1.5">
                    <div className="flex items-center gap-2">
                        <Network className="w-5 h-5 text-purple-600" />
                        Service Connections
                        <sup className="text-xs text-purple-500 font-medium -ml-1.5">
                            alpha
                        </sup>
                    </div>
                    {showCollapsed ? (
                        <div className="flex items-center gap-1.5 text-muted-foreground text-sm font-normal">
                            <AlertCircle className="w-4 h-4" />
                            <span>
                                Requires metrics to be enabled on at least one
                                cluster
                            </span>
                        </div>
                    ) : (
                        <div className="flex items-center gap-3">
                            <div className="text-xs text-muted-foreground">
                                Last updated: {formatLastUpdated(lastUpdated)}
                            </div>
                            <Select
                                value={timeRange.value}
                                onValueChange={(value) => {
                                    const selectedRange = TIME_RANGES.find(
                                        (r) => r.value === value
                                    );
                                    if (selectedRange) {
                                        setTimeRange(selectedRange);
                                    }
                                }}
                            >
                                <SelectTrigger className="w-40 h-8 cursor-pointer">
                                    <div className="flex items-center gap-2">
                                        <Clock className="w-4 h-4" />
                                        <SelectValue />
                                    </div>
                                </SelectTrigger>
                                <SelectContent>
                                    {TIME_RANGES.map((range) => (
                                        <SelectItem
                                            key={range.value}
                                            value={range.value}
                                        >
                                            {range.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={triggerRefresh}
                                disabled={isRefreshing}
                                className="h-8 cursor-pointer"
                            >
                                <RefreshCw
                                    className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`}
                                />
                            </Button>
                        </div>
                    )}
                </CardTitle>
            </CardHeader>
            {!showCollapsed && (
                <CardContent className="relative">
                    {isLoading ? (
                        <div className="flex items-center justify-center h-64">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-600"></div>
                        </div>
                    ) : error ? (
                        <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
                            <Network className="w-16 h-16 mb-4" />
                            <p className="text-center">
                                Failed to load service connections
                            </p>
                            <p className="text-sm text-center mt-2">
                                {error instanceof Error
                                    ? error.message
                                    : 'Unknown error'}
                            </p>
                        </div>
                    ) : connections && !('code' in connections) ? (
                        // Type guard ensures we have the correct response type
                        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        (connections as any).inbound?.length ||
                        // eslint-disable-next-line @typescript-eslint/no-explicit-any
                        (connections as any).outbound?.length ? (
                            <ServiceConnectionsVisualization
                                serviceName={serviceName}
                                namespace={namespace}
                                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                inbound={(connections as any).inbound || []}
                                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                outbound={(connections as any).outbound || []}
                            />
                        ) : (
                            <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
                                <Network className="w-16 h-16 mb-4" />
                                <p className="text-center">
                                    No service connections found
                                </p>
                                <p className="text-sm text-center mt-2">
                                    This service has no inbound or outbound
                                    traffic in the selected time range
                                </p>
                            </div>
                        )
                    ) : null}
                </CardContent>
            )}
        </Card>
    );
};
