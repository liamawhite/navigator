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
import { ChevronUp, ChevronDown, Shield } from 'lucide-react';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { ConfigActions } from '@/components/envoy/ConfigActions';
import type { v1alpha1AuthorizationPolicy } from '@/types/generated/openapi-service_registry';

interface AuthorizationPoliciesTableProps {
    authorizationPolicies: v1alpha1AuthorizationPolicy[];
}

type SortConfig = {
    key: string;
    direction: 'asc' | 'desc';
} | null;

export const AuthorizationPoliciesTable: React.FC<
    AuthorizationPoliciesTableProps
> = ({ authorizationPolicies }) => {
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

    const sortedAuthorizationPolicies = [...authorizationPolicies].sort(
        (a, b) => {
            if (!sortConfig) return 0;

            let aVal: string | number | undefined;
            let bVal: string | number | undefined;

            if (sortConfig.key === 'name') {
                aVal = a.name;
                bVal = b.name;
            } else if (sortConfig.key === 'namespace') {
                aVal = a.namespace;
                bVal = b.namespace;
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
        }
    );

    if (authorizationPolicies.length === 0) {
        return (
            <p className="text-sm text-muted-foreground">
                No authorization policies matched for this instance
            </p>
        );
    }

    return (
        <div className="space-y-2">
            <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Shield className="w-4 h-4 text-red-500" />
                AuthorizationPolicies ({authorizationPolicies.length})
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
                    {sortedAuthorizationPolicies.map((policy, index) => (
                        <TableRow key={index}>
                            <TableCell>
                                <span className="font-mono text-sm">
                                    {policy.name || 'Unknown'} /{' '}
                                    {policy.namespace || 'Unknown'}
                                </span>
                            </TableCell>
                            <TableCell>
                                <ConfigActions
                                    name={policy.name || 'Unknown'}
                                    rawConfig={policy.rawConfig || ''}
                                    configType="AuthorizationPolicy"
                                    copyId={`authorization-policy-${policy.name || index}`}
                                />
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
};
