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

import { useState, useEffect } from 'react';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from './ui/dropdown-menu';
import { serviceApi } from '../utils/api';
import type {
    v1alpha1ClusterSyncInfo,
    v1alpha1SyncStatus,
} from '../types/generated/openapi-service_registry';
import { ChevronDown, Server, Circle } from 'lucide-react';

const getSyncStatusColor = (status?: v1alpha1SyncStatus): string => {
    switch (status) {
        case 'SYNC_STATUS_INITIALIZING':
            return 'bg-blue-500';
        case 'SYNC_STATUS_HEALTHY':
            return 'bg-green-500';
        case 'SYNC_STATUS_STALE':
            return 'bg-yellow-500';
        case 'SYNC_STATUS_DISCONNECTED':
            return 'bg-red-500';
        default:
            return 'bg-gray-500';
    }
};

const getSyncStatusText = (status?: v1alpha1SyncStatus): string => {
    switch (status) {
        case 'SYNC_STATUS_INITIALIZING':
            return 'Initializing';
        case 'SYNC_STATUS_HEALTHY':
            return 'Healthy';
        case 'SYNC_STATUS_STALE':
            return 'Stale';
        case 'SYNC_STATUS_DISCONNECTED':
            return 'Disconnected';
        default:
            return 'Unknown';
    }
};

const formatTimestamp = (timestamp?: string): string => {
    if (!timestamp) return 'Unknown';
    try {
        const date = new Date(timestamp);
        return date.toLocaleString();
    } catch {
        return 'Invalid date';
    }
};

const getTimeAgo = (timestamp?: string): string => {
    if (!timestamp) return 'Unknown';
    try {
        const date = new Date(timestamp);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffSeconds = Math.floor(diffMs / 1000);
        const diffMinutes = Math.floor(diffSeconds / 60);
        const diffHours = Math.floor(diffMinutes / 60);
        const diffDays = Math.floor(diffHours / 24);

        if (diffSeconds < 60) return `${diffSeconds}s ago`;
        if (diffMinutes < 60) return `${diffMinutes}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        return `${diffDays}d ago`;
    } catch {
        return 'Unknown';
    }
};

export const ClusterSyncStatus: React.FC = () => {
    const [clusters, setClusters] = useState<v1alpha1ClusterSyncInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isOpen, setIsOpen] = useState(false);

    const fetchClusters = async () => {
        try {
            setLoading(true);
            setError(null);
            const clusterData = await serviceApi.listClusters();
            setClusters(clusterData);
        } catch (err) {
            console.error('Failed to fetch clusters:', err);
            setError('Failed to load cluster sync status');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchClusters();
        // Refresh every 30 seconds
        const interval = setInterval(fetchClusters, 30000);
        return () => clearInterval(interval);
    }, []);

    // Refresh when dropdown is opened
    useEffect(() => {
        if (isOpen) {
            fetchClusters();
        }
    }, [isOpen]);

    const getOverallStatus = (): v1alpha1SyncStatus => {
        if (clusters.length === 0) return 'SYNC_STATUS_DISCONNECTED';

        const statuses = clusters.map((c) => c.syncStatus);
        if (statuses.includes('SYNC_STATUS_DISCONNECTED'))
            return 'SYNC_STATUS_DISCONNECTED';
        if (statuses.includes('SYNC_STATUS_STALE')) return 'SYNC_STATUS_STALE';
        if (statuses.includes('SYNC_STATUS_INITIALIZING'))
            return 'SYNC_STATUS_INITIALIZING';
        return 'SYNC_STATUS_HEALTHY';
    };

    const overallStatus = getOverallStatus();

    return (
        <DropdownMenu onOpenChange={setIsOpen}>
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="sm"
                    className="gap-2 cursor-pointer"
                >
                    <Server className="h-4 w-4" />
                    <Circle
                        className={`h-2 w-2 rounded-full ${getSyncStatusColor(overallStatus)}`}
                    />
                    <span className="hidden sm:inline">
                        {clusters.length} cluster
                        {clusters.length !== 1 ? 's' : ''}
                    </span>
                    <ChevronDown className="h-3 w-3" />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-80">
                <DropdownMenuLabel className="flex items-center gap-2">
                    <Server className="h-4 w-4" />
                    Cluster Sync Status
                </DropdownMenuLabel>
                <DropdownMenuSeparator />

                {loading && (
                    <DropdownMenuItem disabled>
                        Loading cluster status...
                    </DropdownMenuItem>
                )}

                {error && (
                    <DropdownMenuItem disabled className="text-red-600">
                        {error}
                    </DropdownMenuItem>
                )}

                {!loading && !error && clusters.length === 0 && (
                    <DropdownMenuItem disabled>
                        No clusters connected
                    </DropdownMenuItem>
                )}

                {!loading &&
                    !error &&
                    clusters.map((cluster) => (
                        <DropdownMenuItem
                            key={cluster.clusterId}
                            className="flex-col items-start gap-1 p-3"
                        >
                            <div className="flex items-center gap-2 w-full">
                                <Circle
                                    className={`h-2 w-2 rounded-full ${getSyncStatusColor(cluster.syncStatus)}`}
                                />
                                <span className="font-medium">
                                    {cluster.clusterId}
                                </span>
                                <Badge variant="secondary" className="ml-auto">
                                    {getSyncStatusText(cluster.syncStatus)}
                                </Badge>
                            </div>
                            <div className="text-xs text-muted-foreground ml-4 space-y-1">
                                <div>Services: {cluster.serviceCount || 0}</div>
                                <div>
                                    Last sync: {getTimeAgo(cluster.lastUpdate)}
                                </div>
                                <div>
                                    Connected:{' '}
                                    {formatTimestamp(cluster.connectedAt)}
                                </div>
                            </div>
                        </DropdownMenuItem>
                    ))}

                {!loading && !error && clusters.length > 0 && (
                    <>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                            className="text-xs text-muted-foreground justify-center"
                            disabled
                        >
                            Refreshes every 30 seconds
                        </DropdownMenuItem>
                    </>
                )}
            </DropdownMenuContent>
        </DropdownMenu>
    );
};
