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
import type { v1alpha1WasmPlugin } from '@/types/generated/openapi-service_registry';

interface WasmPluginsTableProps {
    wasmPlugins: v1alpha1WasmPlugin[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const WasmPluginsTable: React.FC<WasmPluginsTableProps> = ({
    wasmPlugins,
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

    const formatPhase = (phase?: string): string => {
        switch (phase?.toLowerCase()) {
            case 'authn':
                return 'Authentication';
            case 'authz':
                return 'Authorization';
            case 'stats':
                return 'Statistics';
            case 'unspecified':
            default:
                return phase || 'Unspecified';
        }
    };

    const getPhaseVariant = (phase?: string): 'default' | 'secondary' | 'destructive' | 'outline' => {
        switch (phase?.toLowerCase()) {
            case 'authn':
                return 'default';
            case 'authz':
                return 'secondary';
            case 'stats':
                return 'outline';
            case 'unspecified':
            default:
                return 'outline';
        }
    };

    const sortedWasmPlugins = [...wasmPlugins].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string | number | undefined;
        let bVal: string | number | undefined;

        if (sortConfig.key === 'name') {
            aVal = a.name;
            bVal = b.name;
        } else if (sortConfig.key === 'namespace') {
            aVal = a.namespace;
            bVal = b.namespace;
        } else if (sortConfig.key === 'phase') {
            aVal = a.spec?.phase;
            bVal = b.spec?.phase;
        } else if (sortConfig.key === 'priority') {
            aVal = a.spec?.priority || 0;
            bVal = b.spec?.priority || 0;
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

    if (wasmPlugins.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No WASM plugins matched for this instance
            </p>
        );
    }

    return (
        <div className="space-y-2">
            <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Settings className="w-4 h-4 text-indigo-500" />
                WasmPlugins ({wasmPlugins.length})
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
                            onClick={() => handleSort('phase')}
                        >
                            <div className="flex items-center">
                                Phase
                                {getSortIcon('phase')}
                            </div>
                        </TableHead>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-20"
                            onClick={() => handleSort('priority')}
                        >
                            <div className="flex items-center">
                                Priority
                                {getSortIcon('priority')}
                            </div>
                        </TableHead>
                        <TableHead className="w-20"></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedWasmPlugins.map((plugin, index) => (
                        <TableRow key={index}>
                            <TableCell>
                                <span className="font-mono text-sm">
                                    {plugin.name || 'Unknown'} / {plugin.namespace || 'Unknown'}
                                </span>
                            </TableCell>
                            <TableCell>
                                <Badge variant={getPhaseVariant(plugin.spec?.phase)}>
                                    {formatPhase(plugin.spec?.phase)}
                                </Badge>
                            </TableCell>
                            <TableCell>
                                <span className="text-sm">
                                    {plugin.spec?.priority || 0}
                                </span>
                            </TableCell>
                            <TableCell>
                                <ConfigActions
                                    name={plugin.name || 'Unknown'}
                                    rawConfig={plugin.rawConfig || ''}
                                    configType="WasmPlugin"
                                    copyId={`wasm-plugin-${plugin.name || index}`}
                                />
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
};