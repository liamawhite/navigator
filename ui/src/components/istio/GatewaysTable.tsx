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
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { ArrowRightToLine, ChevronRight, ChevronDown } from 'lucide-react';
import { ConfigActions } from '../envoy';
import type { v1alpha1Gateway } from '../../types/generated/openapi-service_registry';

interface GatewaysTableProps {
    gateways: v1alpha1Gateway[];
    isCollapsed?: boolean;
    onToggleCollapse?: () => void;
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const GatewaysTable: React.FC<GatewaysTableProps> = ({ 
    gateways,
    isCollapsed = false,
    onToggleCollapse,
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
            <ChevronRight className="w-4 h-4 ml-1" />
        ) : (
            <ChevronDown className="w-4 h-4 ml-1" />
        );
    };

    const sortedGateways = [...gateways].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string;
        let bVal: string;

        if (sortConfig.key === 'name') {
            aVal =
                `${a.name || 'Unknown'}/${a.namespace || 'Unknown'}`.toLowerCase();
            bVal =
                `${b.name || 'Unknown'}/${b.namespace || 'Unknown'}`.toLowerCase();
        } else {
            return 0;
        }

        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
    });

    return (
        <div className="space-y-2">
            <h4 
                className={`text-sm font-medium text-muted-foreground flex items-center gap-2 ${
                    onToggleCollapse ? 'cursor-pointer hover:text-foreground transition-colors' : ''
                }`}
                onClick={onToggleCollapse}
            >
                {onToggleCollapse && (isCollapsed ? (
                    <ChevronRight className="w-4 h-4" />
                ) : (
                    <ChevronDown className="w-4 h-4" />
                ))}
                <ArrowRightToLine className="w-4 h-4 text-purple-500" />
                Gateways ({gateways.length})
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
                                Name / Namespace
                                {getSortIcon('name')}
                            </div>
                        </TableHead>
                        <TableHead className="w-48">Selector</TableHead>
                        <TableHead className="w-20"></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedGateways.map((gateway, index) => {
                        const selector = gateway.selector || {};

                        return (
                            <TableRow key={index}>
                                <TableCell>
                                    <span className="font-mono text-sm">
                                        {gateway.name || 'Unknown'} /{' '}
                                        {gateway.namespace || 'Unknown'}
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <div className="flex flex-wrap gap-1">
                                        {Object.entries(selector).length > 0 ? (
                                            Object.entries(selector)
                                                .slice(0, 2)
                                                .map(([key, value], i) => (
                                                    <Badge
                                                        key={i}
                                                        variant="secondary"
                                                        className="text-xs"
                                                    >
                                                        {key}: {value}
                                                    </Badge>
                                                ))
                                        ) : (
                                            <span className="text-muted-foreground text-sm">
                                                -
                                            </span>
                                        )}
                                        {Object.entries(selector).length >
                                            2 && (
                                            <Badge
                                                variant="outline"
                                                className="text-xs"
                                            >
                                                +
                                                {Object.entries(selector)
                                                    .length - 2}{' '}
                                                more
                                            </Badge>
                                        )}
                                    </div>
                                </TableCell>
                                <TableCell>
                                    <ConfigActions
                                        name={gateway.name || 'Gateway'}
                                        rawConfig={gateway.rawConfig || ''}
                                        configType="Gateway"
                                        copyId={`gateway-${index}`}
                                    />
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
                </Table>
            )}
        </div>
    );
};
