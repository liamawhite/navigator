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
import { ChevronUp, ChevronDown, Settings } from 'lucide-react';
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
import type { v1alpha1EnvoyFilter } from '@/types/generated/openapi-service_registry';

interface EnvoyFiltersTableProps {
    envoyFilters: v1alpha1EnvoyFilter[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const EnvoyFiltersTable: React.FC<EnvoyFiltersTableProps> = ({
    envoyFilters,
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

    const formatContext = (context?: string): string => {
        switch (context?.toLowerCase()) {
            case 'sidecar_inbound':
                return 'Sidecar Inbound';
            case 'sidecar_outbound':
                return 'Sidecar Outbound';
            case 'gateway':
                return 'Gateway';
            default:
                return context || 'Any';
        }
    };

    const getContextVariant = (context?: string): 'default' | 'secondary' | 'destructive' | 'outline' => {
        switch (context?.toLowerCase()) {
            case 'sidecar_inbound':
                return 'default';
            case 'sidecar_outbound':
                return 'secondary';
            case 'gateway':
                return 'outline';
            default:
                return 'outline';
        }
    };

    const sortedEnvoyFilters = [...envoyFilters].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined;
        let bVal: string | number | undefined;

        if (sortConfig.key === 'name') {
            aVal = a.name;
            bVal = b.name;
        } else if (sortConfig.key === 'namespace') {
            aVal = a.namespace;
            bVal = b.namespace;
        } else if (sortConfig.key === 'context') {
            aVal = a.spec?.configPatches?.[0]?.applyTo;
            bVal = b.spec?.configPatches?.[0]?.applyTo;
        } else if (sortConfig.key === 'patches') {
            aVal = a.spec?.configPatches?.length || 0;
            bVal = b.spec?.configPatches?.length || 0;
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

    if (envoyFilters.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No envoy filters matched for this instance
            </p>
        );
    }

    return (
        <div className="space-y-2">
            <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Settings className="w-4 h-4 text-red-500" />
                EnvoyFilters ({envoyFilters.length})
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
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-32"
                            onClick={() => handleSort('context')}
                        >
                            <div className="flex items-center">
                                Context
                                {getSortIcon('context')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-20"
                            onClick={() => handleSort('patches')}
                        >
                            <div className="flex items-center">
                                Patches
                                {getSortIcon('patches')}
                            </div>
                        </TableHead>
                        <TableHead className="w-20"></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedEnvoyFilters.map((filter, index) => (
                        <TableRow key={index}>
                            <TableCell>
                                <span className="font-mono text-sm">
                                    {filter.name || 'Unknown'} / {filter.namespace || 'Unknown'}
                                </span>
                            </TableCell>
                            <TableCell>
                                <Badge variant={getContextVariant(filter.spec?.workloadSelector?.labels ? 'sidecar' : 'gateway')}>
                                    {formatContext(filter.spec?.workloadSelector?.labels ? 'sidecar' : 'gateway')}
                                </Badge>
                            </TableCell>
                            <TableCell>
                                <span className="text-sm">
                                    {filter.spec?.configPatches?.length || 0}
                                </span>
                            </TableCell>
                            <TableCell>
                                <ConfigActions
                                    name={filter.name || 'Unknown'}
                                    rawConfig={filter.rawConfig || ''}
                                    configType="EnvoyFilter"
                                    copyId={`envoy-filter-${filter.name || index}`}
                                />
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
};