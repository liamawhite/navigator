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

import { useState } from 'react';
import { useLocalStorage } from '../../hooks/useLocalStorage';
import {
    ChevronRight,
    ChevronDown,
    Target,
    Layers,
    Globe,
    Settings,
} from 'lucide-react';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { ConfigActions } from '@/components/envoy/ConfigActions';
import type { v1alpha1ClusterSummary } from '@/types/generated/openapi-service_registry';

interface ClustersTableProps {
    clusters: v1alpha1ClusterSummary[];
    serviceId?: string;
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

// Helper functions for cluster type formatting and styling
const formatClusterType = (type?: string): string => {
    if (!type) return 'unknown';

    const typeStr = type.toUpperCase();
    switch (typeStr) {
        case 'STATIC':
            return 'static';
        case 'STRICT_DNS':
            return 'strict_dns';
        case 'LOGICAL_DNS':
            return 'logical_dns';
        case 'EDS':
            return 'eds';
        case 'ORIGINAL_DST':
            return 'original_dst';
        default:
            return type.toLowerCase().replace(/_/g, '_');
    }
};

const getClusterTypeVariant = (
    type?: string
): 'default' | 'secondary' | 'destructive' | 'outline' => {
    if (!type) return 'outline';

    const typeStr = type.toUpperCase();
    switch (typeStr) {
        case 'EDS':
            return 'default'; // Blue - service discovery clusters
        case 'STATIC':
            return 'secondary'; // Gray - static clusters
        case 'STRICT_DNS':
        case 'LOGICAL_DNS':
            return 'outline'; // Outlined - DNS-based clusters
        case 'ORIGINAL_DST':
            return 'destructive'; // Red - special routing clusters
        default:
            return 'outline';
    }
};

// Helper function to group clusters by type
const groupClustersByType = (clusters: v1alpha1ClusterSummary[]) => {
    const groups = {
        serviceDiscovery: [] as v1alpha1ClusterSummary[], // EDS clusters
        static: [] as v1alpha1ClusterSummary[], // STATIC clusters
        dns: [] as v1alpha1ClusterSummary[], // DNS-based clusters
        special: [] as v1alpha1ClusterSummary[], // ORIGINAL_DST and others
    };

    clusters.forEach((cluster) => {
        const type = cluster.type?.toUpperCase() || '';

        if (type === 'EDS') {
            groups.serviceDiscovery.push(cluster);
        } else if (type === 'STATIC') {
            groups.static.push(cluster);
        } else if (type === 'STRICT_DNS' || type === 'LOGICAL_DNS') {
            groups.dns.push(cluster);
        } else if (
            type === 'ORIGINAL_DST' ||
            type === 'UNKNOWN' ||
            type === ''
        ) {
            groups.special.push(cluster);
        } else {
            // Fallback - put unknown types in special group
            groups.special.push(cluster);
        }
    });

    return groups;
};

// Helper component for rendering a group of clusters
const ClusterGroup: React.FC<{
    title: string;
    clusters: v1alpha1ClusterSummary[];
    sortConfig: SortConfig;
    handleSort: (key: string) => void;
    getSortIcon: (key: string) => React.ReactNode;
    isCollapsed: boolean;
    onToggleCollapse: () => void;
}> = ({ title, clusters, sortConfig, handleSort, getSortIcon, isCollapsed, onToggleCollapse }) => {
    if (clusters.length === 0) return null;

    const sortedClusters = [...clusters].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined = a[
            sortConfig.key as keyof v1alpha1ClusterSummary
        ] as string | number | undefined;
        let bVal: string | number | undefined = b[
            sortConfig.key as keyof v1alpha1ClusterSummary
        ] as string | number | undefined;

        // Convert to string for comparison if needed
        if (typeof aVal === 'string') aVal = aVal.toLowerCase();
        if (typeof bVal === 'string') bVal = bVal.toLowerCase();

        // Handle null/undefined values
        if (aVal == null && bVal == null) return 0;
        if (aVal == null) return sortConfig.direction === 'asc' ? -1 : 1;
        if (bVal == null) return sortConfig.direction === 'asc' ? 1 : -1;

        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
    });

    const getGroupIcon = (title: string) => {
        switch (title) {
            case 'Service Discovery Clusters':
                return <Target className="w-4 h-4 text-blue-500" />;
            case 'Static Clusters':
                return <Layers className="w-4 h-4 text-green-500" />;
            case 'DNS-Based Clusters':
                return <Globe className="w-4 h-4 text-orange-500" />;
            case 'Special Clusters':
                return <Settings className="w-4 h-4 text-purple-500" />;
            default:
                return null;
        }
    };

    return (
        <div className="space-y-2">
            <h4 
                className="text-sm font-medium text-muted-foreground flex items-center gap-2 cursor-pointer hover:text-foreground transition-colors"
                onClick={onToggleCollapse}
            >
                {isCollapsed ? (
                    <ChevronRight className="w-4 h-4" />
                ) : (
                    <ChevronDown className="w-4 h-4" />
                )}
                {getGroupIcon(title)}
                {title} ({clusters.length})
            </h4>
            {!isCollapsed && (
                <Table className="table-fixed">
                <TableHeader>
                    <TableRow>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50"
                            onClick={() => handleSort('serviceFqdn')}
                        >
                            <div className="flex items-center">
                                Service
                                {getSortIcon('serviceFqdn')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-20"
                            onClick={() => handleSort('direction')}
                        >
                            <div className="flex items-center">
                                Direction
                                {getSortIcon('direction')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-16"
                            onClick={() => handleSort('port')}
                        >
                            <div className="flex items-center">
                                Port
                                {getSortIcon('port')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-20"
                            onClick={() => handleSort('subset')}
                        >
                            <div className="flex items-center">
                                Subset
                                {getSortIcon('subset')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-24"
                            onClick={() => handleSort('type')}
                        >
                            <div className="flex items-center">
                                Type
                                {getSortIcon('type')}
                            </div>
                        </TableHead>
                        <TableHead className="w-32"></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedClusters.map((cluster, index) => (
                        <TableRow key={index}>
                            <TableCell>
                                <span className="font-mono text-sm">
                                    {cluster.serviceFqdn ||
                                        cluster.name ||
                                        'N/A'}
                                </span>
                            </TableCell>
                            <TableCell className="w-20">
                                <Badge
                                    variant={
                                        cluster.direction === 'INBOUND'
                                            ? 'default'
                                            : cluster.direction === 'OUTBOUND'
                                              ? 'secondary'
                                              : 'outline'
                                    }
                                >
                                    {cluster.direction?.toLowerCase() ||
                                        'unknown'}
                                </Badge>
                            </TableCell>
                            <TableCell className="w-16">
                                <span className="font-mono text-sm">
                                    {cluster.port || 'N/A'}
                                </span>
                            </TableCell>
                            <TableCell className="w-20">
                                <span className="text-sm">
                                    {cluster.subset || '-'}
                                </span>
                            </TableCell>
                            <TableCell className="w-24">
                                <Badge
                                    variant={getClusterTypeVariant(
                                        cluster.type
                                    )}
                                >
                                    {formatClusterType(cluster.type)}
                                </Badge>
                            </TableCell>
                            <TableCell className="w-32">
                                <ConfigActions
                                    name={cluster.name || 'Unknown'}
                                    rawConfig={cluster.rawConfig || ''}
                                    configType="Cluster"
                                    copyId={`cluster-${cluster.name}`}
                                />
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
                </Table>
            )}
        </div>
    );
};

export const ClustersTable: React.FC<ClustersTableProps> = ({ clusters, serviceId }) => {
    const [sortConfig, setSortConfig] = useState<SortConfig>({
        key: 'name',
        direction: 'asc',
    });

    const storageKey = serviceId ? `clusters-collapsed-${serviceId}` : 'clusters-collapsed';
    const [collapsedGroups, setCollapsedGroups] = useLocalStorage<Record<string, boolean>>(storageKey, {
        service: false,
        static: true, // Default closed for Static Clusters
        dns: false,
        special: false,
    });

    const toggleGroupCollapse = (groupKey: string) => {
        setCollapsedGroups(prev => ({
            ...prev,
            [groupKey]: !prev[groupKey]
        }));
    };

    const handleSort = (key: string) => {
        let direction: 'asc' | 'desc' = 'asc';
        if (
            sortConfig &&
            sortConfig.key === key &&
            sortConfig.direction === 'asc'
        ) {
            direction = 'desc';
        }
        setSortConfig({ key, direction });
    };

    const getSortIcon = (columnKey: string) => {
        if (!sortConfig || sortConfig.key !== columnKey) {
            return null;
        }
        return sortConfig.direction === 'asc' ? (
            <ChevronRight className="w-4 h-4 ml-1" />
        ) : (
            <ChevronDown className="w-4 h-4 ml-1" />
        );
    };

    if (clusters.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No clusters configured
            </p>
        );
    }

    const groups = groupClustersByType(clusters);

    return (
        <div className="space-y-6">
            <ClusterGroup
                title="Service Discovery Clusters"
                clusters={groups.serviceDiscovery}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.service}
                onToggleCollapse={() => toggleGroupCollapse('service')}
            />
            <ClusterGroup
                title="DNS-Based Clusters"
                clusters={groups.dns}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.dns}
                onToggleCollapse={() => toggleGroupCollapse('dns')}
            />
            <ClusterGroup
                title="Special Clusters"
                clusters={groups.special}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.special}
                onToggleCollapse={() => toggleGroupCollapse('special')}
            />
            <ClusterGroup
                title="Static Clusters"
                clusters={groups.static}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.static}
                onToggleCollapse={() => toggleGroupCollapse('static')}
            />
        </div>
    );
};
