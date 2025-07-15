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
import type { v1alpha1RouteConfigSummary } from '@/types/generated/openapi-troubleshooting';

interface RoutesTableProps {
    routes: v1alpha1RouteConfigSummary[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const RoutesTable: React.FC<RoutesTableProps> = ({ routes }) => {
    const [sortConfig, setSortConfig] = useState<SortConfig>(null);

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
        } else if (sortConfig.key === 'internalOnlyHeaders') {
            aVal = a.internalOnlyHeaders?.length || 0;
            bVal = b.internalOnlyHeaders?.length || 0;
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

    if (routes.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No routes configured
            </p>
        );
    }

    return (
        <Table>
            <TableHeader>
                <TableRow>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('name')}
                    >
                        <div className="flex items-center">
                            Name
                            {getSortIcon('name')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('virtualHosts')}
                    >
                        <div className="flex items-center">
                            Virtual Hosts
                            {getSortIcon('virtualHosts')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('validateClusters')}
                    >
                        <div className="flex items-center">
                            Validate Clusters
                            {getSortIcon('validateClusters')}
                        </div>
                    </TableHead>
                    <TableHead>Internal Headers</TableHead>
                    <TableHead>Request Headers</TableHead>
                    <TableHead>Response Headers</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {sortedRoutes.map((route, index) => (
                    <TableRow key={index}>
                        <TableCell>
                            <span className="font-mono text-sm">
                                {route.name || 'N/A'}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                {route.virtualHosts?.length || 0}
                            </span>
                        </TableCell>
                        <TableCell>
                            <Badge
                                variant={
                                    route.validateClusters
                                        ? 'default'
                                        : 'secondary'
                                }
                            >
                                {route.validateClusters ? 'Yes' : 'No'}
                            </Badge>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                {route.internalOnlyHeaders?.length || 0}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                +{route.requestHeadersToAdd?.length || 0} -
                                {route.requestHeadersToRemove?.length || 0}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                +{route.responseHeadersToAdd?.length || 0} -
                                {route.responseHeadersToRemove?.length || 0}
                            </span>
                        </TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    );
};
