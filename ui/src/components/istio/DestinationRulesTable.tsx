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
import { Target, ChevronUp, ChevronDown } from 'lucide-react';
import { ConfigActions } from '../envoy';
import type { v1alpha1DestinationRule } from '../../types/generated/openapi-service_registry';

interface DestinationRulesTableProps {
    destinationRules: v1alpha1DestinationRule[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const DestinationRulesTable: React.FC<DestinationRulesTableProps> = ({
    destinationRules,
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

    const sortedDestinationRules = [...destinationRules].sort((a, b) => {
        if (!sortConfig) return 0;

        let aVal: string;
        let bVal: string;

        if (sortConfig.key === 'name') {
            aVal = `${a.name || 'Unknown'}/${a.namespace || 'Unknown'}`.toLowerCase();
            bVal = `${b.name || 'Unknown'}/${b.namespace || 'Unknown'}`.toLowerCase();
        } else {
            return 0;
        }

        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
    });

    return (
        <div className="space-y-2">
            <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Target className="w-4 h-4 text-green-500" />
                DestinationRules ({destinationRules.length})
            </h4>
            <Table className="table-fixed">
                <TableHeader>
                    <TableRow>
                        <TableHead
                            className="cursor-pointer select-none hover:bg-muted/50 w-40"
                            onClick={() => handleSort('name')}
                        >
                            <div className="flex items-center">
                                Name / Namespace
                                {getSortIcon('name')}
                            </div>
                        </TableHead>
                        <TableHead className="w-64">Host</TableHead>
                        <TableHead className="w-32">Subsets</TableHead>
                        <TableHead className="w-20"></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedDestinationRules.map((dr, index) => {
                        const subsets = dr.subsets || [];
                        
                        return (
                            <TableRow key={index}>
                                <TableCell>
                                    <span className="font-mono text-sm">
                                        {dr.name || 'Unknown'} / {dr.namespace || 'Unknown'}
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <span className="font-mono text-sm">
                                        {dr.host || '-'}
                                    </span>
                                </TableCell>
                                <TableCell>
                                    <div className="flex flex-wrap gap-1">
                                        {subsets.length > 0 ? (
                                            subsets.slice(0, 3).map((subset, i) => (
                                                <Badge key={i} variant="secondary" className="text-xs">
                                                    {subset.name || `subset-${i}`}
                                                </Badge>
                                            ))
                                        ) : (
                                            <Badge variant="outline" className="text-xs text-muted-foreground">
                                                none
                                            </Badge>
                                        )}
                                        {subsets.length > 3 && (
                                            <Badge variant="outline" className="text-xs">
                                                +{subsets.length - 3} more
                                            </Badge>
                                        )}
                                    </div>
                                </TableCell>
                                <TableCell>
                                    <ConfigActions
                                        name={dr.name || 'DestinationRule'}
                                        rawConfig={dr.rawConfig || ''}
                                        configType="DestinationRule"
                                        copyId={`dr-${index}`}
                                    />
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
};