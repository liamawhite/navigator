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
import type { ListenerCollapseGroups } from '../../types/collapseGroups';
import {
    ChevronRight,
    ChevronDown,
    EthernetPort,
    Target,
    Asterisk,
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
import type { v1alpha1ListenerSummary } from '@/types/generated/openapi-service_registry';
import { v1alpha1ProxyMode } from '@/types/generated/openapi-service_registry';

interface ListenersTableProps {
    listeners: v1alpha1ListenerSummary[];
    proxyMode?: v1alpha1ProxyMode;
    serviceId?: string;
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

// Helper functions for listener type formatting and styling
const formatListenerType = (type?: string | number): string => {
    if (!type) return 'unknown';

    // Convert enum values to display format - handle both string and numeric types
    const typeStr = String(type).toUpperCase();
    switch (typeStr) {
        case '0':
        case 'VIRTUAL_INBOUND':
            return 'virtual_inbound';
        case '1':
        case 'VIRTUAL_OUTBOUND':
            return 'virtual_outbound';
        case '2':
        case 'SERVICE_OUTBOUND':
            return 'service_outbound';
        case '3':
        case 'PORT_OUTBOUND':
            return 'port_outbound';
        case '4':
        case 'PROXY_METRICS':
            return 'proxy_metrics';
        case '5':
        case 'PROXY_HEALTHCHECK':
            return 'proxy_healthcheck';
        case '6':
        case 'ADMIN_XDS':
            return 'admin_xds';
        case '7':
        case 'ADMIN_WEBHOOK':
            return 'admin_webhook';
        case '8':
        case 'ADMIN_DEBUG':
            return 'admin_debug';
        case '9':
        case 'GATEWAY_INBOUND':
            return 'gateway_inbound';
        default:
            return String(type).toLowerCase().replace(/\s+/g, '_');
    }
};

const getTypeVariant = (
    type?: string | number
): 'default' | 'secondary' | 'destructive' | 'outline' => {
    if (!type) return 'outline';

    const typeStr = String(type).toUpperCase();
    switch (typeStr) {
        case '0':
        case 'VIRTUAL_INBOUND':
            return 'default'; // Blue - primary virtual inbound (main traffic entry)
        case '1':
        case 'VIRTUAL_OUTBOUND':
            return 'default'; // Blue - primary virtual outbound (main traffic exit)
        case '2':
        case 'SERVICE_OUTBOUND':
            return 'secondary'; // Gray - service-specific listeners
        case '3':
        case 'PORT_OUTBOUND':
            return 'secondary'; // Gray - port-based listeners
        case '4':
        case 'PROXY_METRICS':
        case '5':
        case 'PROXY_HEALTHCHECK':
            return 'outline'; // Outlined - proxy operational listeners
        case '6':
        case 'ADMIN_XDS':
        case '7':
        case 'ADMIN_WEBHOOK':
        case '8':
        case 'ADMIN_DEBUG':
            return 'secondary'; // Gray - legacy admin types (now port-based)
        case '9':
        case 'GATEWAY_INBOUND':
            return 'default'; // Blue - gateway inbound traffic entry
        default:
            return 'outline';
    }
};

// Helper function to group listeners by type (4-category system for modern Istio architecture)
const groupListenersByType = (listeners: v1alpha1ListenerSummary[]) => {
    const groups = {
        virtual: [] as v1alpha1ListenerSummary[],
        service: [] as v1alpha1ListenerSummary[],
        port: [] as v1alpha1ListenerSummary[],
        proxy: [] as v1alpha1ListenerSummary[],
    };

    listeners.forEach((listener) => {
        const type = String(listener.type || '').toLowerCase();
        if (
            type === 'virtual_inbound' ||
            type === 'virtual_outbound' ||
            type === 'gateway_inbound' ||
            type === '0' ||
            type === '1' ||
            type === '9'
        ) {
            // Virtual listeners are the main traffic entry/exit points in Istio (including gateway inbound)
            groups.virtual.push(listener);
        } else if (type === 'service_outbound' || type === '2') {
            // Service-specific outbound listeners (specific service destinations)
            groups.service.push(listener);
        } else if (
            type === 'port_outbound' ||
            type === '3' ||
            type === 'admin_xds' ||
            type === '6' ||
            type === 'admin_webhook' ||
            type === '7' ||
            type === 'admin_debug' ||
            type === '8'
        ) {
            // Port-based outbound listeners (generic port traffic including admin ports)
            groups.port.push(listener);
        } else if (
            type === 'proxy_metrics' ||
            type === '4' ||
            type === 'proxy_healthcheck' ||
            type === '5'
        ) {
            // Proxy-specific operational endpoints
            groups.proxy.push(listener);
        } else {
            // Fallback - put unknown types in port group
            groups.port.push(listener);
        }
    });

    return groups;
};

// Helper component for rendering a group of listeners
const ListenerGroup: React.FC<{
    title: string;
    listeners: v1alpha1ListenerSummary[];
    sortConfig: SortConfig;
    handleSort: (key: string) => void;
    getSortIcon: (key: string) => React.ReactNode;
    isCollapsed: boolean;
    onToggleCollapse: () => void;
}> = ({
    title,
    listeners,
    sortConfig,
    handleSort,
    getSortIcon,
    isCollapsed,
    onToggleCollapse,
}) => {
    if (listeners.length === 0) return null;

    const sortedListeners = [...listeners].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined = a[
            sortConfig.key as keyof v1alpha1ListenerSummary
        ] as string | number | undefined;
        let bVal: string | number | undefined = b[
            sortConfig.key as keyof v1alpha1ListenerSummary
        ] as string | number | undefined;

        // Handle special cases - none currently needed

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
            case 'Virtual Listeners':
            case 'Gateway Listeners':
                return <EthernetPort className="w-4 h-4 text-blue-500" />;
            case 'Service-Specific Listeners':
                return <Target className="w-4 h-4 text-green-500" />;
            case 'Port-Based Listeners':
                return <Asterisk className="w-4 h-4 text-orange-500" />;
            case 'Proxy Listeners':
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
                {title} ({listeners.length})
            </h4>
            {!isCollapsed && (
                <Table className="table-fixed">
                    <TableHeader>
                        <TableRow>
                            <TableHead
                                className="cursor-pointer select-none hover:bg-muted/50 w-48"
                                onClick={() => handleSort('name')}
                            >
                                <div className="flex items-center">
                                    Name
                                    {getSortIcon('name')}
                                </div>
                            </TableHead>
                            <TableHead
                                className="cursor-pointer select-none hover:bg-muted/50 w-32"
                                onClick={() => handleSort('address')}
                            >
                                <div className="flex items-center">
                                    Address:Port
                                    {getSortIcon('address')}
                                </div>
                            </TableHead>
                            <TableHead
                                className="cursor-pointer select-none hover:bg-muted/50 w-32"
                                onClick={() => handleSort('type')}
                            >
                                <div className="flex items-center">
                                    Type
                                    {getSortIcon('type')}
                                </div>
                            </TableHead>
                            <TableHead className="w-20"></TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {sortedListeners.map((listener, index) => (
                            <TableRow key={index}>
                                <TableCell>
                                    <span className="font-mono text-sm">
                                        {listener.name || 'N/A'}
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <span className="font-mono text-sm">
                                        {listener.address && listener.port
                                            ? `${listener.address}:${listener.port}`
                                            : listener.address ||
                                              listener.port ||
                                              'N/A'}
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <Badge
                                        variant={getTypeVariant(listener.type)}
                                    >
                                        {formatListenerType(listener.type)}
                                    </Badge>
                                </TableCell>
                                <TableCell>
                                    <ConfigActions
                                        name={listener.name || 'Unknown'}
                                        rawConfig={listener.rawConfig || ''}
                                        configType="Listener"
                                        copyId={`listener-${listener.name}`}
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

export const ListenersTable: React.FC<ListenersTableProps> = ({
    listeners,
    proxyMode,
    serviceId,
}) => {
    const [sortConfig, setSortConfig] = useState<SortConfig>({
        key: 'port',
        direction: 'asc',
    });

    const storageKey = serviceId
        ? `listeners-collapsed-${serviceId}`
        : 'listeners-collapsed';
    const { collapsedGroups, toggleGroupCollapse } =
        useCollapsibleSections<ListenerCollapseGroups>(storageKey, {
            virtual: false,
            service: false,
            port: false,
            proxy: true, // Default closed for Proxy Listeners
        });

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

    if (listeners.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No listeners configured
            </p>
        );
    }

    const groups = groupListenersByType(listeners);

    return (
        <div className="space-y-6">
            <ListenerGroup
                title={
                    proxyMode === v1alpha1ProxyMode.GATEWAY
                        ? 'Gateway Listeners'
                        : 'Virtual Listeners'
                }
                listeners={groups.virtual}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.virtual}
                onToggleCollapse={() => toggleGroupCollapse('virtual')}
            />
            <ListenerGroup
                title="Service-Specific Listeners"
                listeners={groups.service}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.service}
                onToggleCollapse={() => toggleGroupCollapse('service')}
            />
            <ListenerGroup
                title="Port-Based Listeners"
                listeners={groups.port}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.port}
                onToggleCollapse={() => toggleGroupCollapse('port')}
            />
            <ListenerGroup
                title="Proxy Listeners"
                listeners={groups.proxy}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
                isCollapsed={collapsedGroups.proxy}
                onToggleCollapse={() => toggleGroupCollapse('proxy')}
            />
        </div>
    );
};
