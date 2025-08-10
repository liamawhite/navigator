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
import { ChevronUp, ChevronDown, Network } from 'lucide-react';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { ConfigActions } from '@/components/envoy/ConfigActions';
import type { v1alpha1Sidecar } from '@/types/generated/openapi-service_registry';

interface SidecarsTableProps {
    sidecars: v1alpha1Sidecar[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const SidecarsTable: React.FC<SidecarsTableProps> = ({
    sidecars,
}) => {
    const [sortConfig, setSortConfig] = useState<SortConfig>({
        key: 'name',
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

    const sortedSidecars = [...sidecars].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined;
        let bVal: string | number | undefined;

        if (sortConfig.key === 'name') {
            aVal = `${a.name || 'Unknown'}/${a.namespace || 'Unknown'}`.toLowerCase();
            bVal = `${b.name || 'Unknown'}/${b.namespace || 'Unknown'}`.toLowerCase();
        } else {
            return 0;
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

    if (sidecars.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No sidecars matched for this instance
            </p>
        );
    }

    return (
        <div className="space-y-2">
            <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Network className="w-4 h-4 text-orange-500" />
                Sidecars ({sidecars.length})
            </h4>
            <Table className="table-fixed">
                <TableHeader>
                    <TableRow>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-48"
                            onClick={() => handleSort('name')}
                        >
                            <div className="flex items-center">
                                Name / Namespace
                                {getSortIcon('name')}
                            </div>
                        </TableHead>
                        <TableHead className="w-20"></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedSidecars.map((sidecar, index) => (
                        <TableRow key={index}>
                            <TableCell>
                                <span className="font-mono text-sm">
                                    {sidecar.name || 'Unknown'} / {sidecar.namespace || 'Unknown'}
                                </span>
                            </TableCell>
                            <TableCell>
                                <ConfigActions
                                    name={sidecar.name || 'Unknown'}
                                    rawConfig={sidecar.rawConfig || ''}
                                    configType="Sidecar"
                                    copyId={`sidecar-${sidecar.name || index}`}
                                />
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
};