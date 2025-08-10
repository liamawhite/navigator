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
import {
    ChevronUp,
    ChevronDown,
    ChevronRight,
    Globe,
    Link,
    MapPin,
    Target,
    Layers,
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
import type { v1alpha1EndpointSummary } from '@/types/generated/openapi-service_registry';
import {
    v1alpha1ClusterType,
    v1alpha1AddressType,
} from '@/types/generated/openapi-service_registry';

interface EndpointsTableProps {
    endpoints: v1alpha1EndpointSummary[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

// Helper function to get address type icon and label
const getAddressTypeInfo = (addressType?: v1alpha1AddressType) => {
    switch (addressType) {
        case v1alpha1AddressType.SOCKET_ADDRESS:
            return { icon: Globe, label: 'Socket', color: 'text-blue-600' };
        case v1alpha1AddressType.PIPE_ADDRESS:
            return { icon: Link, label: 'Pipe', color: 'text-orange-600' };
        default:
            return { icon: Globe, label: 'Unknown', color: 'text-gray-600' };
    }
};

// Helper function to get health status circle
const getHealthIndicator = (health?: string) => {
    switch (health) {
        case 'HEALTHY':
            return <div className="w-2 h-2 rounded-full bg-green-500" />;
        case 'UNHEALTHY':
            return <div className="w-2 h-2 rounded-full bg-red-500" />;
        case 'DRAINING':
            return <div className="w-2 h-2 rounded-full bg-yellow-500" />;
        default:
            return <div className="w-2 h-2 rounded-full bg-gray-400" />;
    }
};

// Helper function to format locality
const formatLocality = (locality?: { region?: string; zone?: string }) => {
    if (!locality) return 'N/A';
    if (locality.region && locality.zone) {
        return `${locality.region}/${locality.zone}`;
    }
    if (locality.region) return locality.region;
    if (locality.zone) return locality.zone;
    return 'N/A';
};

// Helper function to group endpoint summaries by cluster type
const groupEndpointsByClusterType = (endpoints: v1alpha1EndpointSummary[]) => {
    const groups = {
        serviceDiscovery: [] as v1alpha1EndpointSummary[], // EDS clusters
        static: [] as v1alpha1EndpointSummary[], // STATIC clusters
        dns: [] as v1alpha1EndpointSummary[], // DNS-based clusters
        special: [] as v1alpha1EndpointSummary[], // ORIGINAL_DST and others
    };

    endpoints.forEach((endpoint) => {
        const type = endpoint.clusterType;

        if (type === v1alpha1ClusterType.CLUSTER_EDS) {
            groups.serviceDiscovery.push(endpoint);
        } else if (type === v1alpha1ClusterType.CLUSTER_STATIC) {
            groups.static.push(endpoint);
        } else if (
            type === v1alpha1ClusterType.CLUSTER_STRICT_DNS ||
            type === v1alpha1ClusterType.CLUSTER_LOGICAL_DNS
        ) {
            groups.dns.push(endpoint);
        } else if (
            type === v1alpha1ClusterType.CLUSTER_ORIGINAL_DST ||
            type === v1alpha1ClusterType.UNKNOWN_CLUSTER_TYPE ||
            type === undefined
        ) {
            groups.special.push(endpoint);
        } else {
            // Fallback - put unknown types in special group
            groups.special.push(endpoint);
        }
    });

    return groups;
};

// Helper component for rendering a group of endpoint summaries
const EndpointsGroup: React.FC<{
    title: string;
    endpoints: v1alpha1EndpointSummary[];
    sortConfig: SortConfig;
    handleSort: (key: string) => void;
    getSortIcon: (key: string) => React.ReactNode;
    expandedClusters: Set<string>;
    toggleCluster: (clusterName: string) => void;
}> = ({
    title,
    endpoints,
    sortConfig,
    handleSort,
    getSortIcon,
    expandedClusters,
    toggleCluster,
}) => {
    if (endpoints.length === 0) return null;

    const sortedClusters = [...endpoints].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined;
        let bVal: string | number | undefined;

        if (sortConfig.key === 'clusterName') {
            aVal = a.clusterName;
            bVal = b.clusterName;
        } else if (sortConfig.key === 'serviceFqdn') {
            aVal = a.serviceFqdn || a.clusterName;
            bVal = b.serviceFqdn || b.clusterName;
        } else if (sortConfig.key === 'port') {
            aVal = a.port;
            bVal = b.port;
        } else if (sortConfig.key === 'subset') {
            aVal = a.subset;
            bVal = b.subset;
        }

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
            <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                {getGroupIcon(title)}
                {title} ({endpoints.length})
            </h4>
            <Table className="table-fixed">
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-8">
                            {/* Expand/collapse column */}
                        </TableHead>
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
                        <TableHead className="text-right w-32">
                            Health Summary
                        </TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedClusters.map((cluster) => {
                        const isExpanded = expandedClusters.has(
                            cluster.clusterName || ''
                        );
                        const totalEndpoints = cluster.endpoints?.length || 0;
                        const healthyEndpoints =
                            cluster.endpoints?.filter(
                                (ep) => ep.health === 'HEALTHY'
                            ).length || 0;
                        const unhealthyEndpoints =
                            cluster.endpoints?.filter(
                                (ep) => ep.health === 'UNHEALTHY'
                            ).length || 0;
                        const otherEndpoints =
                            totalEndpoints -
                            healthyEndpoints -
                            unhealthyEndpoints;

                        return (
                            <>
                                {/* Cluster row */}
                                <TableRow
                                    key={cluster.clusterName}
                                    className="cursor-pointer hover:bg-muted/50"
                                    onClick={() =>
                                        toggleCluster(cluster.clusterName || '')
                                    }
                                >
                                    <TableCell className="w-8">
                                        {totalEndpoints > 0 && (
                                            <div className="flex items-center justify-center">
                                                {isExpanded ? (
                                                    <ChevronDown className="w-4 h-4" />
                                                ) : (
                                                    <ChevronRight className="w-4 h-4" />
                                                )}
                                            </div>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <span className="font-mono text-sm">
                                            {cluster.serviceFqdn ||
                                                cluster.clusterName ||
                                                'N/A'}
                                        </span>
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
                                    <TableCell className="text-right w-32">
                                        <div className="flex gap-1 justify-end">
                                            {healthyEndpoints > 0 && (
                                                <Badge
                                                    variant="default"
                                                    className="bg-green-100 text-green-800 border-green-200 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800"
                                                >
                                                    {healthyEndpoints} Healthy
                                                </Badge>
                                            )}
                                            {unhealthyEndpoints > 0 && (
                                                <Badge
                                                    variant="destructive"
                                                    className="bg-red-100 text-red-800 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800"
                                                >
                                                    {unhealthyEndpoints}{' '}
                                                    Unhealthy
                                                </Badge>
                                            )}
                                            {otherEndpoints > 0 && (
                                                <Badge variant="secondary">
                                                    {otherEndpoints} Other
                                                </Badge>
                                            )}
                                        </div>
                                    </TableCell>
                                </TableRow>

                                {/* Expanded endpoint header */}
                                {isExpanded && (
                                    <TableRow className="bg-muted/10 border-b">
                                        <TableCell className="w-8"></TableCell>
                                        <TableCell colSpan={4}>
                                            <div className="flex items-center gap-2 sm:gap-4 px-2">
                                                <span className="text-xs font-medium text-muted-foreground w-4 flex-shrink-0">
                                                    {/* Health indicator */}
                                                </span>
                                                <span className="text-xs font-medium text-muted-foreground flex-1 min-w-0">
                                                    Host Identifier
                                                </span>
                                                <span className="text-xs font-medium text-muted-foreground w-12 sm:w-16 flex-shrink-0 hidden sm:block">
                                                    Type
                                                </span>
                                                <span className="text-xs font-medium text-muted-foreground w-16 sm:w-24 flex-shrink-0 hidden md:block">
                                                    Locality
                                                </span>
                                                <span className="text-xs font-medium text-muted-foreground w-12 sm:w-16 flex-shrink-0 hidden lg:block">
                                                    Priority
                                                </span>
                                                <span className="text-xs font-medium text-muted-foreground w-12 sm:w-16 flex-shrink-0">
                                                    Weight
                                                </span>
                                                <span className="text-xs font-medium text-muted-foreground w-12 sm:w-16 flex-shrink-0">
                                                </span>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                )}

                                {/* Expanded endpoint rows */}
                                {isExpanded &&
                                    cluster.endpoints?.map(
                                        (endpoint, index) => {
                                            const addressTypeInfo =
                                                getAddressTypeInfo(
                                                    endpoint.addressType
                                                );
                                            const Icon = addressTypeInfo.icon;

                                            return (
                                                <TableRow
                                                    key={`${cluster.clusterName}-${index}`}
                                                    className="bg-muted/25"
                                                >
                                                    <TableCell className="w-8"></TableCell>
                                                    <TableCell colSpan={4}>
                                                        <div className="flex items-center gap-2 sm:gap-4 px-2">
                                                            <div className="w-4 flex items-center justify-start flex-shrink-0">
                                                                {getHealthIndicator(
                                                                    endpoint.health
                                                                )}
                                                            </div>
                                                            <span className="font-mono text-xs text-muted-foreground flex-1 min-w-0 break-all">
                                                                {endpoint.hostIdentifier ||
                                                                    'N/A'}
                                                            </span>
                                                            <div className="flex items-center gap-1 w-12 sm:w-16 flex-shrink-0 hidden sm:flex">
                                                                <Icon
                                                                    className={`w-3 h-3 ${addressTypeInfo.color}`}
                                                                />
                                                                <span
                                                                    className={`text-xs ${addressTypeInfo.color} hidden sm:inline`}
                                                                >
                                                                    {
                                                                        addressTypeInfo.label
                                                                    }
                                                                </span>
                                                            </div>
                                                            <div className="flex items-center gap-1 w-16 sm:w-24 flex-shrink-0 hidden md:flex">
                                                                <MapPin className="w-3 h-3 text-blue-600 flex-shrink-0" />
                                                                <span className="text-xs text-muted-foreground truncate">
                                                                    {formatLocality(
                                                                        endpoint.locality
                                                                    )}
                                                                </span>
                                                            </div>
                                                            <span className="text-xs text-muted-foreground w-12 sm:w-16 flex-shrink-0 hidden lg:block text-center">
                                                                {endpoint.priority ||
                                                                    0}
                                                            </span>
                                                            <span className="text-xs text-muted-foreground w-12 sm:w-16 flex-shrink-0 text-center">
                                                                {endpoint.weight ||
                                                                    'N/A'}
                                                            </span>
                                                            <div className="w-12 sm:w-16 flex-shrink-0">
                                                                <ConfigActions
                                                                    name={
                                                                        endpoint.hostIdentifier ||
                                                                        'Unknown'
                                                                    }
                                                                    rawConfig={JSON.stringify(
                                                                        endpoint,
                                                                        null,
                                                                        2
                                                                    )}
                                                                    configType="Endpoint"
                                                                    copyId={`endpoint-${cluster.clusterName}-${index}`}
                                                                />
                                                            </div>
                                                        </div>
                                                    </TableCell>
                                                </TableRow>
                                            );
                                        }
                                    )}
                            </>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
};

export const EndpointsTable: React.FC<EndpointsTableProps> = ({
    endpoints,
}) => {
    const [sortConfig, setSortConfig] = useState<SortConfig>({
        key: 'serviceFqdn',
        direction: 'asc',
    });
    const [expandedClusters, setExpandedClusters] = useState<Set<string>>(
        new Set()
    );

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
            <ChevronUp className="w-4 h-4 ml-1" />
        ) : (
            <ChevronDown className="w-4 h-4 ml-1" />
        );
    };

    const toggleCluster = (clusterName: string) => {
        const newExpanded = new Set(expandedClusters);
        if (newExpanded.has(clusterName)) {
            newExpanded.delete(clusterName);
        } else {
            newExpanded.add(clusterName);
        }
        setExpandedClusters(newExpanded);
    };

    if (endpoints.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No endpoints configured
            </p>
        );
    }

    const groups = groupEndpointsByClusterType(endpoints);

    return (
        <div className="space-y-6">
            <EndpointsGroup
                title="Service Discovery Clusters"
                endpoints={groups.serviceDiscovery}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                expandedClusters={expandedClusters}
                toggleCluster={toggleCluster}
            />
            <EndpointsGroup
                title="Static Clusters"
                endpoints={groups.static}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                expandedClusters={expandedClusters}
                toggleCluster={toggleCluster}
            />
            <EndpointsGroup
                title="DNS-Based Clusters"
                endpoints={groups.dns}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                expandedClusters={expandedClusters}
                toggleCluster={toggleCluster}
            />
            <EndpointsGroup
                title="Special Clusters"
                endpoints={groups.special}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                expandedClusters={expandedClusters}
                toggleCluster={toggleCluster}
            />
        </div>
    );
};
