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

import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import type { v1alpha1BootstrapSummary } from '@/types/generated/openapi-troubleshooting';

interface BootstrapConfigProps {
    bootstrap: v1alpha1BootstrapSummary | null;
}

export const BootstrapConfig: React.FC<BootstrapConfigProps> = ({
    bootstrap,
}) => {
    if (!bootstrap) {
        return (
            <p className="text-sm text-muted-foreground">
                Bootstrap configuration not available
            </p>
        );
    }

    return (
        <div className="space-y-6">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Property</TableHead>
                        <TableHead>Value</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    <TableRow>
                        <TableCell className="font-medium">Node ID</TableCell>
                        <TableCell>
                            <span className="font-mono text-sm">
                                {bootstrap.node?.id || 'N/A'}
                            </span>
                        </TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell className="font-medium">Cluster</TableCell>
                        <TableCell>
                            <span className="font-mono text-sm">
                                {bootstrap.node?.cluster || 'N/A'}
                            </span>
                        </TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell className="font-medium">
                            Admin Address
                        </TableCell>
                        <TableCell>
                            <span className="font-mono text-sm">
                                {bootstrap.adminAddress || 'N/A'}:
                                {bootstrap.adminPort || 'N/A'}
                            </span>
                        </TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell className="font-medium">
                            Static Resources Version
                        </TableCell>
                        <TableCell>
                            <span className="text-sm">
                                {bootstrap.staticResourcesVersion || 'N/A'}
                            </span>
                        </TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell className="font-medium">
                            Dynamic Resources
                        </TableCell>
                        <TableCell>
                            <Badge
                                variant={
                                    bootstrap.dynamicResourcesConfig
                                        ? 'default'
                                        : 'secondary'
                                }
                            >
                                {bootstrap.dynamicResourcesConfig
                                    ? 'Configured'
                                    : 'Not configured'}
                            </Badge>
                        </TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell className="font-medium">
                            Cluster Manager
                        </TableCell>
                        <TableCell>
                            <Badge
                                variant={
                                    bootstrap.clusterManager
                                        ? 'default'
                                        : 'secondary'
                                }
                            >
                                {bootstrap.clusterManager
                                    ? 'Configured'
                                    : 'Not configured'}
                            </Badge>
                        </TableCell>
                    </TableRow>
                    {bootstrap.node?.locality && (
                        <TableRow>
                            <TableCell className="font-medium">
                                Locality
                            </TableCell>
                            <TableCell>
                                <span className="text-sm">
                                    {bootstrap.node.locality.region || 'N/A'} /{' '}
                                    {bootstrap.node.locality.zone || 'N/A'}
                                </span>
                            </TableCell>
                        </TableRow>
                    )}
                </TableBody>
            </Table>

            {bootstrap.node?.metadata &&
                Object.keys(bootstrap.node.metadata).length > 0 && (
                    <div>
                        <h5 className="text-sm font-medium mb-2">
                            Node Metadata
                        </h5>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Key</TableHead>
                                    <TableHead>Value</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {Object.entries(bootstrap.node.metadata).map(
                                    ([key, value]) => (
                                        <TableRow key={key}>
                                            <TableCell className="font-mono text-sm">
                                                {key}
                                            </TableCell>
                                            <TableCell className="font-mono text-sm">
                                                {value}
                                            </TableCell>
                                        </TableRow>
                                    )
                                )}
                            </TableBody>
                        </Table>
                    </div>
                )}
        </div>
    );
};
