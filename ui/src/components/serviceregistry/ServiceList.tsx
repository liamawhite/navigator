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

import { useServices } from '../../hooks/useServices';
import { useClusters } from '../../hooks/useClusters';
import {
    Loader2,
    AlertCircle,
    Server,
    Database,
    ChevronUp,
    ChevronDown,
    Hexagon,
    Globe,
    ArrowRightToLine,
} from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import { useState, useMemo } from 'react';
import { v1alpha1ProxyMode } from '../../types/generated/openapi-service_registry/models/v1alpha1ProxyMode';

interface ServiceListProps {
    onServiceSelect?: (serviceId: string) => void;
}

type SortField = 'name' | 'namespace';
type SortDirection = 'asc' | 'desc';

export const ServiceList: React.FC<ServiceListProps> = ({
    onServiceSelect,
}) => {
    const { data: services, isLoading, error, isError } = useServices();
    const { data: clusters } = useClusters();
    const [sortField, setSortField] = useState<SortField>('namespace');
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

    const handleSort = (field: SortField) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('asc');
        }
    };

    const { gateways, regularServices } = useMemo(() => {
        if (!services) return { gateways: [], regularServices: [] };

        // Split services into gateways and regular services
        const gatewayList = services.filter(
            (service) => service.proxyMode === v1alpha1ProxyMode.ROUTER
        );
        const serviceList = services.filter(
            (service) => service.proxyMode !== v1alpha1ProxyMode.ROUTER
        );

        // Sort function for both lists
        const sortServices = (
            a: (typeof services)[0],
            b: (typeof services)[0]
        ) => {
            const namespaceA = a.namespace?.toLowerCase() || '';
            const namespaceB = b.namespace?.toLowerCase() || '';

            if (sortField === 'namespace') {
                // Primary sort by namespace
                if (namespaceA !== namespaceB) {
                    if (namespaceA < namespaceB)
                        return sortDirection === 'asc' ? -1 : 1;
                    if (namespaceA > namespaceB)
                        return sortDirection === 'asc' ? 1 : -1;
                }
                // Secondary sort by service name within same namespace
                const nameA = a.name?.toLowerCase() || '';
                const nameB = b.name?.toLowerCase() || '';
                if (nameA < nameB) return -1;
                if (nameA > nameB) return 1;
                return 0;
            } else {
                // Primary sort by service name
                const nameA = a.name?.toLowerCase() || '';
                const nameB = b.name?.toLowerCase() || '';
                if (nameA !== nameB) {
                    if (nameA < nameB) return sortDirection === 'asc' ? -1 : 1;
                    if (nameA > nameB) return sortDirection === 'asc' ? 1 : -1;
                }
                // Secondary sort by namespace when service names are same
                if (namespaceA < namespaceB) return -1;
                if (namespaceA > namespaceB) return 1;
                return 0;
            }
        };

        return {
            gateways: [...gatewayList].sort(sortServices),
            regularServices: [...serviceList].sort(sortServices),
        };
    }, [services, sortField, sortDirection]);

    const ServiceTable = ({
        serviceList,
        title,
        icon,
    }: {
        serviceList: typeof services;
        title: string;
        icon: React.ReactNode;
    }) => {
        if (!serviceList || serviceList.length === 0) return null;

        return (
            <div className="mb-4">
                <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                        {icon}
                        <h3 className="text-lg font-semibold text-foreground">
                            {title}
                        </h3>
                    </div>
                    <span className="text-sm text-muted-foreground bg-muted px-3 py-1 rounded-full">
                        {serviceList.length}{' '}
                        {title.toLowerCase().replace(/s$/, '')}
                        {serviceList.length !== 1 ? 's' : ''}
                    </span>
                </div>

                <div className="bg-background border rounded-lg shadow-sm overflow-hidden">
                    <Table>
                        <TableHeader>
                            <TableRow className="group">
                                <TableHead
                                    className="w-64 cursor-pointer hover:bg-muted/50 select-none"
                                    onClick={() => handleSort('name')}
                                >
                                    <div className="flex items-center gap-1">
                                        <div>Service</div>
                                        {sortField === 'name' &&
                                            (sortDirection === 'desc' ? (
                                                <ChevronDown className="w-4 h-4 text-foreground" />
                                            ) : (
                                                <ChevronUp className="w-4 h-4 text-foreground" />
                                            ))}
                                        {sortField !== 'name' && (
                                            <ChevronUp className="w-3 h-3 text-muted-foreground opacity-60" />
                                        )}
                                    </div>
                                </TableHead>
                                <TableHead
                                    className="w-32 cursor-pointer hover:bg-muted/50 select-none"
                                    onClick={() => handleSort('namespace')}
                                >
                                    <div className="flex items-center gap-1">
                                        <div>Namespace</div>
                                        {sortField === 'namespace' &&
                                            (sortDirection === 'desc' ? (
                                                <ChevronDown className="w-4 h-4 text-foreground" />
                                            ) : (
                                                <ChevronUp className="w-4 h-4 text-foreground" />
                                            ))}
                                        {sortField !== 'namespace' && (
                                            <ChevronUp className="w-3 h-3 text-muted-foreground opacity-60" />
                                        )}
                                    </div>
                                </TableHead>
                                <TableHead className="w-20">
                                    <div>Clusters</div>
                                </TableHead>
                                <TableHead className="w-20">
                                    <div>Instances</div>
                                </TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {serviceList.map((service) => {
                                const proxiedInstances =
                                    service.instances?.filter(
                                        (i) => i.envoyPresent
                                    ).length || 0;

                                const uniqueClusters = new Set(
                                    service.instances?.map(
                                        (i) => i.clusterName
                                    ) || []
                                ).size;

                                return (
                                    <TableRow
                                        key={service.id}
                                        className="cursor-pointer hover:bg-muted/50"
                                        onClick={() =>
                                            onServiceSelect?.(service.id || '')
                                        }
                                    >
                                        <TableCell className="w-64">
                                            <div className="flex items-center gap-2">
                                                <TooltipProvider>
                                                    <Tooltip>
                                                        <TooltipTrigger asChild>
                                                            <Hexagon
                                                                className={`w-4 h-4 ${
                                                                    proxiedInstances >
                                                                    0
                                                                        ? 'text-purple-600 dark:text-purple-400 fill-purple-100 dark:fill-purple-900'
                                                                        : 'text-transparent fill-transparent'
                                                                }`}
                                                            />
                                                        </TooltipTrigger>
                                                        {proxiedInstances >
                                                            0 && (
                                                            <TooltipContent>
                                                                <p>
                                                                    {proxiedInstances ===
                                                                    (service
                                                                        .instances
                                                                        ?.length ||
                                                                        0)
                                                                        ? 'Envoy present in all instances'
                                                                        : `Envoy present in ${proxiedInstances} of ${service.instances?.length || 0} instances`}
                                                                </p>
                                                            </TooltipContent>
                                                        )}
                                                    </Tooltip>
                                                </TooltipProvider>
                                                <span className="font-medium text-foreground truncate">
                                                    {service.name}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="w-32">
                                            <Badge
                                                variant="secondary"
                                                className="text-xs text-foreground bg-muted"
                                            >
                                                {service.namespace}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="w-20">
                                            <div className="flex items-center gap-2">
                                                <Globe className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                                                <span className="text-foreground">
                                                    {uniqueClusters}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="w-20">
                                            <div className="flex items-center gap-2">
                                                <Database className="w-4 h-4 text-green-600 dark:text-green-400" />
                                                <span className="text-foreground">
                                                    {service.instances
                                                        ?.length || 0}
                                                </span>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                </div>
            </div>
        );
    };

    if (isLoading) {
        return (
            <Card className="border-0 shadow-md">
                <CardContent className="flex items-center justify-center py-12">
                    <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
                    <span className="ml-3 text-muted-foreground font-medium">
                        Loading services...
                    </span>
                </CardContent>
            </Card>
        );
    }

    if (isError) {
        return (
            <Card className="border-0 shadow-md border-red-100 bg-red-50 dark:border-red-900 dark:bg-red-950">
                <CardContent className="flex items-center justify-center py-12">
                    <AlertCircle className="w-8 h-8 text-red-500" />
                    <span className="ml-3 text-red-700 dark:text-red-400 font-medium">
                        Failed to load services: {error?.message}
                    </span>
                </CardContent>
            </Card>
        );
    }

    if (!services || services.length === 0) {
        // Check if all clusters are initializing
        const allClustersInitializing =
            clusters &&
            clusters.length > 0 &&
            clusters.every(
                (cluster) => cluster.syncStatus === 'SYNC_STATUS_INITIALIZING'
            );

        if (allClustersInitializing) {
            return (
                <Card className="border-0 shadow-md border-blue-100 bg-blue-50 dark:border-blue-900 dark:bg-blue-950">
                    <CardContent className="text-center py-12">
                        <Loader2 className="w-16 h-16 text-blue-500 mx-auto mb-4 animate-spin" />
                        <h3 className="text-lg font-semibold text-foreground mb-2">
                            Initializing clusters
                        </h3>
                        <p className="text-muted-foreground mb-4">
                            Connected to {clusters.length} cluster
                            {clusters.length !== 1 ? 's' : ''} but still waiting
                            for initial service discovery. This usually takes a
                            few moments.
                        </p>
                        <p className="text-sm text-blue-600 dark:text-blue-400">
                            Services will appear here once the initial sync is
                            complete.
                        </p>
                    </CardContent>
                </Card>
            );
        }

        return (
            <Card className="border-0 shadow-md">
                <CardContent className="text-center py-12">
                    <Server className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-foreground mb-2">
                        No services found
                    </h3>
                    <p className="text-muted-foreground">
                        Your cluster doesn't have any services yet, or they
                        haven't been discovered.
                    </p>
                </CardContent>
            </Card>
        );
    }

    return (
        <div>
            <div className="flex items-center justify-between mb-3">
                <h2 className="text-2xl font-bold text-foreground">
                    Discovered Services
                </h2>
                <div className="flex items-center gap-2">
                    {gateways.length > 0 && (
                        <span className="text-sm text-muted-foreground bg-orange-100 dark:bg-orange-950 text-orange-700 dark:text-orange-300 px-3 py-1 rounded-full">
                            {gateways.length} gateway
                            {gateways.length !== 1 ? 's' : ''}
                        </span>
                    )}
                    <span className="text-sm text-muted-foreground bg-muted px-3 py-1 rounded-full">
                        {gateways.length + regularServices.length} total
                    </span>
                </div>
            </div>

            {gateways.length > 0 && (
                <ServiceTable
                    serviceList={gateways}
                    title="Gateways"
                    icon={
                        <ArrowRightToLine className="w-5 h-5 text-orange-600 dark:text-orange-400" />
                    }
                />
            )}

            {regularServices.length > 0 && (
                <ServiceTable
                    serviceList={regularServices}
                    title="Services"
                    icon={
                        <Server className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                    }
                />
            )}

            {gateways.length === 0 && regularServices.length === 0 && (
                <Card className="border-0 shadow-md">
                    <CardContent className="text-center py-12">
                        <Server className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                        <h3 className="text-lg font-semibold text-foreground mb-2">
                            No services found
                        </h3>
                        <p className="text-muted-foreground">
                            Your cluster doesn't have any services yet, or they
                            haven't been discovered.
                        </p>
                    </CardContent>
                </Card>
            )}
        </div>
    );
};
