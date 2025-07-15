import { useState } from 'react';
import { ChevronUp, ChevronDown } from 'lucide-react';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { RawConfigDialog } from '@/components/envoy/RawConfigDialog';

interface ListenerSummary {
    name: string;
    address: string;
    port: number;
    type: string;
    filterChains?: { length: number };
    listenerFilters?: { length: number };
    rawConfig: string; // JSON representation of the full listener config
}

interface ListenersTableProps {
    listeners: ListenerSummary[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

// Helper functions for listener type formatting and styling
const formatListenerType = (type: string): string => {
    if (!type) return 'unknown';

    // Convert enum values to display format
    switch (type.toUpperCase()) {
        case '0':
        case 'INBOUND':
            return 'inbound';
        case '1':
        case 'OUTBOUND':
            return 'outbound';
        case '2':
        case 'VIRTUAL_INBOUND':
            return 'virtual_inbound';
        case '3':
        case 'VIRTUAL_OUTBOUND':
            return 'virtual_outbound';
        case '4':
        case 'METRICS':
            return 'metrics';
        case '5':
        case 'HEALTHCHECK':
            return 'healthcheck';
        case '6':
        case 'ADMIN_XDS':
            return 'admin_xds';
        case '7':
        case 'ADMIN_WEBHOOK':
            return 'admin_webhook';
        case '8':
        case 'ADMIN_DEBUG':
            return 'admin_debug';
        default:
            return type.toLowerCase().replace(/\s+/g, '_');
    }
};

const getTypeVariant = (
    type: string
): 'default' | 'secondary' | 'destructive' | 'outline' => {
    if (!type) return 'outline';

    switch (type.toUpperCase()) {
        case '0':
        case 'INBOUND':
            return 'default'; // Blue - regular inbound
        case '1':
        case 'OUTBOUND':
            return 'secondary'; // Gray - outbound traffic
        case '2':
        case 'VIRTUAL_INBOUND':
            return 'default'; // Blue - virtual inbound
        case '3':
        case 'VIRTUAL_OUTBOUND':
            return 'secondary'; // Gray - virtual outbound
        case '4':
        case 'METRICS':
        case '5':
        case 'HEALTHCHECK':
        case '6':
        case 'ADMIN_XDS':
        case '7':
        case 'ADMIN_WEBHOOK':
        case '8':
        case 'ADMIN_DEBUG':
            return 'outline'; // Outlined - admin/system traffic
        default:
            return 'outline';
    }
};

// Helper function to group listeners by type
const groupListenersByType = (listeners: ListenerSummary[]) => {
    const groups = {
        virtual: [] as ListenerSummary[],
        inbound: [] as ListenerSummary[],
        outbound: [] as ListenerSummary[],
        admin: [] as ListenerSummary[],
    };

    listeners.forEach((listener) => {
        const type = listener.type?.toLowerCase();
        if (
            type === 'virtual_inbound' ||
            type === 'virtual_outbound' ||
            type === '2' ||
            type === '3'
        ) {
            groups.virtual.push(listener);
        } else if (type === 'inbound' || type === '0') {
            groups.inbound.push(listener);
        } else if (type === 'outbound' || type === '1') {
            groups.outbound.push(listener);
        } else {
            // Admin types: metrics, healthcheck, admin_xds, admin_webhook, admin_debug
            groups.admin.push(listener);
        }
    });

    return groups;
};

// Helper component for rendering a group of listeners
const ListenerGroup: React.FC<{
    title: string;
    listeners: ListenerSummary[];
    sortConfig: SortConfig;
    handleSort: (key: string) => void;
    getSortIcon: (key: string) => React.ReactNode;
}> = ({ title, listeners, sortConfig, handleSort, getSortIcon }) => {
    if (listeners.length === 0) return null;

    const sortedListeners = [...listeners].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: any = a[sortConfig.key as keyof ListenerSummary];
        let bVal: any = b[sortConfig.key as keyof ListenerSummary];

        // Handle special cases
        if (sortConfig.key === 'filterChains') {
            aVal = a.filterChains?.length || 0;
            bVal = b.filterChains?.length || 0;
        } else if (sortConfig.key === 'listenerFilters') {
            aVal = a.listenerFilters?.length || 0;
            bVal = b.listenerFilters?.length || 0;
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

    return (
        <div className="space-y-2">
            <h4 className="text-sm font-medium text-muted-foreground">
                {title} ({listeners.length})
            </h4>
            <Table>
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
                            className="cursor-pointer select-none hover:bg-muted/50 w-40"
                            onClick={() => handleSort('address')}
                        >
                            <div className="flex items-center">
                                Address
                                {getSortIcon('address')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-20"
                            onClick={() => handleSort('port')}
                        >
                            <div className="flex items-center">
                                Port
                                {getSortIcon('port')}
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
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-28"
                            onClick={() => handleSort('filterChains')}
                        >
                            <div className="flex items-center">
                                Filter Chains
                                {getSortIcon('filterChains')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-32"
                            onClick={() => handleSort('listenerFilters')}
                        >
                            <div className="flex items-center">
                                Listener Filters
                                {getSortIcon('listenerFilters')}
                            </div>
                        </TableHead>
                        <TableHead className="w-20">Actions</TableHead>
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
                                    {listener.address || 'N/A'}
                                </span>
                            </TableCell>
                            <TableCell>
                                <span className="font-mono text-sm">
                                    {listener.port || 'N/A'}
                                </span>
                            </TableCell>
                            <TableCell>
                                <Badge variant={getTypeVariant(listener.type)}>
                                    {formatListenerType(listener.type)}
                                </Badge>
                            </TableCell>
                            <TableCell>
                                <span className="text-sm">
                                    {listener.filterChains?.length || 0}
                                </span>
                            </TableCell>
                            <TableCell>
                                <span className="text-sm">
                                    {listener.listenerFilters?.length || 0}
                                </span>
                            </TableCell>
                            <TableCell>
                                <RawConfigDialog
                                    name={listener.name}
                                    rawConfig={listener.rawConfig}
                                    configType="Listener"
                                />
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
};

export const ListenersTable: React.FC<ListenersTableProps> = ({
    listeners,
}) => {
    const [sortConfig, setSortConfig] = useState<SortConfig>({
        key: 'port',
        direction: 'asc',
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
            <ChevronUp className="w-4 h-4 ml-1" />
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
                title="Inbound Listeners"
                listeners={groups.inbound}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
            />
            <ListenerGroup
                title="Outbound Listeners"
                listeners={groups.outbound}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
            />
            <ListenerGroup
                title="Virtual Listeners"
                listeners={groups.virtual}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
            />
            <ListenerGroup
                title="Admin Listeners"
                listeners={groups.admin}
                sortConfig={sortConfig}
                handleSort={handleSort}
                getSortIcon={getSortIcon}
            />
        </div>
    );
};
