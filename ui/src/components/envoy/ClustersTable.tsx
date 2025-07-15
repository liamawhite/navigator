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
import type { v1alpha1ClusterSummary } from '@/types/generated/openapi-troubleshooting';

interface ClustersTableProps {
    clusters: v1alpha1ClusterSummary[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const ClustersTable: React.FC<ClustersTableProps> = ({ clusters }) => {
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

    const sortedClusters = [...clusters].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined = a[
            sortConfig.key as keyof v1alpha1ClusterSummary
        ] as string | number | undefined;
        let bVal: string | number | undefined = b[
            sortConfig.key as keyof v1alpha1ClusterSummary
        ] as string | number | undefined;

        // Handle special cases
        if (sortConfig.key === 'endpoints') {
            aVal = a.loadAssignment?.endpoints?.length || 0;
            bVal = b.loadAssignment?.endpoints?.length || 0;
        } else if (sortConfig.key === 'healthChecks') {
            aVal = a.healthChecks?.length || 0;
            bVal = b.healthChecks?.length || 0;
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

    if (clusters.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No clusters configured
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
                        onClick={() => handleSort('type')}
                    >
                        <div className="flex items-center">
                            Type
                            {getSortIcon('type')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('loadBalancingPolicy')}
                    >
                        <div className="flex items-center">
                            LB Policy
                            {getSortIcon('loadBalancingPolicy')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('connectTimeout')}
                    >
                        <div className="flex items-center">
                            Connect Timeout
                            {getSortIcon('connectTimeout')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('endpoints')}
                    >
                        <div className="flex items-center">
                            Endpoints
                            {getSortIcon('endpoints')}
                        </div>
                    </TableHead>
                    <TableHead
                        className="cursor-pointer select-none hover:bg-muted/50"
                        onClick={() => handleSort('healthChecks')}
                    >
                        <div className="flex items-center">
                            Health Checks
                            {getSortIcon('healthChecks')}
                        </div>
                    </TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {sortedClusters.map((cluster, index) => (
                    <TableRow key={index}>
                        <TableCell>
                            <span className="font-mono text-sm">
                                {cluster.name || 'N/A'}
                            </span>
                        </TableCell>
                        <TableCell>
                            <Badge variant="outline">
                                {cluster.type || 'UNKNOWN'}
                            </Badge>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                {cluster.loadBalancingPolicy ||
                                    cluster.lbPolicy ||
                                    'N/A'}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm font-mono">
                                {cluster.connectTimeout || 'N/A'}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                {cluster.loadAssignment?.endpoints?.length || 0}
                            </span>
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                {cluster.healthChecks?.length || 0}
                            </span>
                        </TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    );
};
