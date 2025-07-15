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
import type { v1alpha1EndpointSummary } from '@/types/generated/openapi-troubleshooting';
import type { v1alpha1EndpointInfo } from '@/types/generated/openapi-troubleshooting';

interface EndpointsTableProps {
    endpoints: v1alpha1EndpointSummary[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const EndpointsTable: React.FC<EndpointsTableProps> = ({
    endpoints,
}) => {
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

    // Flatten the endpoints data
    const flatEndpoints = endpoints.flatMap(
        (endpoint) =>
            endpoint.endpoints?.map((ep) => ({
                ...ep,
                clusterName: endpoint.clusterName,
            })) || []
    );

    const sortedEndpoints = [...flatEndpoints].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined = a[
            sortConfig.key as keyof (v1alpha1EndpointInfo & {
                clusterName?: string;
            })
        ] as string | number | undefined;
        let bVal: string | number | undefined = b[
            sortConfig.key as keyof (v1alpha1EndpointInfo & {
                clusterName?: string;
            })
        ] as string | number | undefined;

        // Handle weight special case
        if (sortConfig.key === 'weight') {
            aVal = a.weight || a.loadBalancingWeight || 0;
            bVal = b.weight || b.loadBalancingWeight || 0;
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

    if (flatEndpoints.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No endpoints configured
            </p>
        );
    }

    return (
        <Table>
            <TableHeader>
                <TableRow>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('clusterName')}
                    >
                        <div className="flex items-center">
                            Cluster Name
                            {getSortIcon('clusterName')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('address')}
                    >
                        <div className="flex items-center">
                            Address:Port
                            {getSortIcon('address')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('health')}
                    >
                        <div className="flex items-center">
                            Health Status
                            {getSortIcon('health')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('priority')}
                    >
                        <div className="flex items-center">
                            Priority
                            {getSortIcon('priority')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('weight')}
                    >
                        <div className="flex items-center">
                            Weight
                            {getSortIcon('weight')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('hostIdentifier')}
                    >
                        <div className="flex items-center">
                            Host Identifier
                            {getSortIcon('hostIdentifier')}
                        </div>
                    </TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {sortedEndpoints.map((ep, index) => (
                    <TableRow key={`${ep.clusterName}-${index}`}>
                        <TableCell>
                            <span className="font-mono text-sm">
                                {ep.clusterName || 'N/A'}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="font-mono text-sm">
                                {ep.address || 'N/A'}:{ep.port || 'N/A'}
                            </span>
                        </TableCell>
                        <TableCell>
                            <Badge
                                variant={
                                    ep.health === 'HEALTHY'
                                        ? 'default'
                                        : ep.health === 'UNHEALTHY'
                                          ? 'destructive'
                                          : 'secondary'
                                }
                            >
                                {ep.health || 'UNKNOWN'}
                            </Badge>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">{ep.priority || 0}</span>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                {ep.weight || ep.loadBalancingWeight || 'N/A'}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="font-mono text-xs">
                                {ep.hostIdentifier || 'N/A'}
                            </span>
                        </TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    );
};
