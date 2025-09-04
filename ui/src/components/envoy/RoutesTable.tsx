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
import { useCollapsibleSections } from '../../hooks/useCollapsibleSections';
import type { RouteCollapseGroups } from '../../types/collapseGroups';
import {
    ChevronRight,
    ChevronDown,
    Target,
    Asterisk,
    Layers,
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
import type { v1alpha1RouteConfigSummary } from '@/types/generated/openapi-service_registry';

interface RoutesTableProps {
    routes: v1alpha1RouteConfigSummary[];
    serviceId?: string;
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

// Helper functions for route type formatting and styling
const formatRouteType = (type?: string | number): string => {
    if (!type) return 'unknown';

    // Convert enum values to display format - handle both string and numeric types
    const typeStr = String(type).toUpperCase();
    switch (typeStr) {
        case '0':
        case 'PORT_BASED':
            return 'port';
        case '1':
        case 'SERVICE_SPECIFIC':
            return 'service';
        case '2':
        case 'STATIC':
            return 'static';
        default:
            return String(type).toLowerCase().replace(/\s+/g, '_');
    }
};

const getRouteTypeVariant = (
    type?: string | number
): 'default' | 'secondary' | 'destructive' | 'outline' => {
    if (!type) return 'outline';

    const typeStr = String(type).toUpperCase();
    switch (typeStr) {
        case '0':
        case 'PORT_BASED':
            return 'default'; // Blue - port-based routes
        case '1':
        case 'SERVICE_SPECIFIC':
            return 'secondary'; // Gray - service-specific routes
        case '2':
        case 'STATIC':
            return 'outline'; // Outlined - static/system routes
        default:
            return 'outline';
    }
};

// Helper function to group routes by type
const groupRoutesByType = (routes: v1alpha1RouteConfigSummary[]) => {
    const groups = {
        portBased: [] as v1alpha1RouteConfigSummary[],
        serviceSpecific: [] as v1alpha1RouteConfigSummary[],
        static: [] as v1alpha1RouteConfigSummary[],
    };

    routes.forEach((route) => {
        const type = String(route.type || '').toLowerCase();

        // If type is set by backend, use it
        if (type === 'port_based' || type === '0') {
            groups.portBased.push(route);
        } else if (type === 'service_specific' || type === '1') {
            groups.serviceSpecific.push(route);
        } else if (type === 'static' || type === '2') {
            groups.static.push(route);
        } else {
            // Fallback: determine type from route name when backend type is not set
            const routeName = route.name || '';

            // Port-based: just numbers (80, 443, 15010, etc.)
            if (/^[1-9]\d{0,4}$/.test(routeName) && routeName.length <= 5) {
                groups.portBased.push(route);
            }
            // Static: Istio internal patterns AND empty/whitespace names
            else if (
                routeName === '' ||
                routeName.trim() === '' ||
                routeName === 'InboundPassthroughCluster' ||
                routeName === 'BlackHoleCluster' ||
                routeName === 'PassthroughCluster' ||
                /^inbound\|\d+\|\|/.test(routeName) ||
                /^outbound\|\d+\|\|/.test(routeName)
            ) {
                groups.static.push(route);
            }
            // Service-specific: everything else (service.namespace.svc.cluster.local:port format)
            else {
                groups.serviceSpecific.push(route);
            }
        }
    });

    return groups;
};

// Helper component for rendering a group of routes
const RouteGroup: React.FC<{
    title: string;
    routes: v1alpha1RouteConfigSummary[];
    sortConfig: SortConfig;
    handleSort: (key: string) => void;
    getSortIcon: (key: string) => React.ReactNode;
    isCollapsed: boolean;
    onToggleCollapse: () => void;
}> = ({
    title,
    routes,
    sortConfig,
    handleSort,
    getSortIcon,
    isCollapsed,
    onToggleCollapse,
}) => {
    if (routes.length === 0) return null;

    const sortedRoutes = [...routes].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | boolean | undefined = a[
            sortConfig.key as keyof v1alpha1RouteConfigSummary
        ] as string | number | boolean | undefined;
        let bVal: string | number | boolean | undefined = b[
            sortConfig.key as keyof v1alpha1RouteConfigSummary
        ] as string | number | boolean | undefined;

        // Handle special cases
        if (sortConfig.key === 'virtualHosts') {
            aVal = a.virtualHosts?.length || 0;
            bVal = b.virtualHosts?.length || 0;
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
            case 'Service-Specific Routes':
                return <Target className="w-4 h-4 text-green-500" />;
            case 'Port-Based Routes':
                return <Asterisk className="w-4 h-4 text-orange-500" />;
            case 'Static Routes':
                return <Layers className="w-4 h-4 text-purple-500" />;
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
                {title} ({routes.length})
            </h4>
            {!isCollapsed && (
                <Table className="table-fixed w-full">
                    <TableHeader>
                        <TableRow>
                            <TableHead
                                className="cursor-pointer select-none hover:bg-muted/50"
                                style={{ width: '60%' }}
                                onClick={() => handleSort('name')}
                            >
                                <div className="flex items-center">
                                    Name
                                    {getSortIcon('name')}
                                </div>
                            </TableHead>
                            <TableHead
                                className="cursor-pointer select-none hover:bg-muted/50"
                                style={{ width: '15%' }}
                                onClick={() => handleSort('type')}
                            >
                                <div className="flex items-center">
                                    Type
                                    {getSortIcon('type')}
                                </div>
                            </TableHead>
                            <TableHead
                                className="cursor-pointer select-none hover:bg-muted/50"
                                style={{ width: '15%' }}
                                onClick={() => handleSort('virtualHosts')}
                            >
                                <div className="flex items-center">
                                    Virtual Hosts
                                    {getSortIcon('virtualHosts')}
                                </div>
                            </TableHead>
                            <TableHead style={{ width: '10%' }}></TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {sortedRoutes.map((route, index) => (
                            <TableRow key={index}>
                                <TableCell>
                                    <span className="font-mono text-sm truncate block">
                                        {route.name || 'N/A'}
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <Badge
                                        variant={getRouteTypeVariant(
                                            route.type
                                        )}
                                    >
                                        {formatRouteType(route.type)}
                                    </Badge>
                                </TableCell>
                                <TableCell>
                                    <span className="text-sm">
                                        {route.virtualHosts?.length || 0}
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <ConfigActions
                                        name={route.name || 'Unknown Route'}
                                        rawConfig={route.rawConfig || ''}
                                        configType="Route"
                                        copyId={`route-${route.name || `unknown-${index}`}`}
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

export const RoutesTable: React.FC<RoutesTableProps> = ({
    routes,
    serviceId,
}) => {
    const [sortConfig, setSortConfig] = useState<SortConfig>({
        key: 'name',
        direction: 'asc',
    });

    const storageKey = serviceId
        ? `routes-collapsed-${serviceId}`
        : 'routes-collapsed';
    const { collapsedGroups, toggleGroupCollapse } = useCollapsibleSections<RouteCollapseGroups>(
        storageKey,
        {
            serviceSpecific: false,
            portBased: false,
            static: true, // Default closed for Static Routes
        }
    );
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

    if (routes.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No routes configured
            </p>
        );
    }

    const groups = groupRoutesByType(routes);

    return (
        <div className="space-y-6">
            <RouteGroup
                title="Service-Specific Routes"
                routes={groups.serviceSpecific}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.serviceSpecific}
                onToggleCollapse={() => toggleGroupCollapse('serviceSpecific')}
            />
            <RouteGroup
                title="Port-Based Routes"
                routes={groups.portBased}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.portBased}
                onToggleCollapse={() => toggleGroupCollapse('portBased')}
            />
            <RouteGroup
                title="Static Routes"
                routes={groups.static}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.static}
                onToggleCollapse={() => toggleGroupCollapse('static')}
            />
        </div>
    );
};
