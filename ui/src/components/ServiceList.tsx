import { useServices } from '../hooks/useServices';
import {
    Loader2,
    AlertCircle,
    Server,
    Database,
    Shield,
    ChevronUp,
    ChevronDown,
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
import { useState, useMemo } from 'react';

interface ServiceListProps {
    onServiceSelect?: (serviceId: string) => void;
}

type SortField = 'name' | 'namespace';
type SortDirection = 'asc' | 'desc';

export const ServiceList: React.FC<ServiceListProps> = ({
    onServiceSelect,
}) => {
    const { data: services, isLoading, error, isError } = useServices();
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

    const sortedServices = useMemo(() => {
        if (!services) return [];

        return [...services].sort((a, b) => {
            // First sort by namespace
            const namespaceA = a.namespace.toLowerCase();
            const namespaceB = b.namespace.toLowerCase();

            if (sortField === 'namespace') {
                // Primary sort by namespace
                if (namespaceA !== namespaceB) {
                    if (namespaceA < namespaceB)
                        return sortDirection === 'asc' ? -1 : 1;
                    if (namespaceA > namespaceB)
                        return sortDirection === 'asc' ? 1 : -1;
                }
                // Secondary sort by service name within same namespace
                const nameA = a.name.toLowerCase();
                const nameB = b.name.toLowerCase();
                if (nameA < nameB) return -1;
                if (nameA > nameB) return 1;
                return 0;
            } else {
                // Primary sort by service name
                const nameA = a.name.toLowerCase();
                const nameB = b.name.toLowerCase();
                if (nameA !== nameB) {
                    if (nameA < nameB) return sortDirection === 'asc' ? -1 : 1;
                    if (nameA > nameB) return sortDirection === 'asc' ? 1 : -1;
                }
                // Secondary sort by namespace when service names are same
                if (namespaceA < namespaceB) return -1;
                if (namespaceA > namespaceB) return 1;
                return 0;
            }
        });
    }, [services, sortField, sortDirection]);

    const SortableHeader = ({
        field,
        children,
    }: {
        field: SortField;
        children: React.ReactNode;
    }) => {
        const isActive = sortField === field;
        const isSecondary = sortField !== field; // This column is the secondary sort

        return (
            <TableHead
                className="cursor-pointer hover:bg-muted/50 select-none"
                onClick={() => handleSort(field)}
            >
                <div className="flex items-center gap-1">
                    {children}
                    <div className="flex flex-col">
                        {/* Primary sort indicator */}
                        {isActive &&
                            (sortDirection === 'desc' ? (
                                <ChevronDown className="w-4 h-4 text-foreground" />
                            ) : (
                                <ChevronUp className="w-4 h-4 text-foreground" />
                            ))}
                        {/* Secondary sort indicator (always ascending) */}
                        {isSecondary && (
                            <ChevronUp className="w-3 h-3 text-muted-foreground opacity-60" />
                        )}
                        {/* Hover hint for inactive columns */}
                        {!isActive && !isSecondary && (
                            <ChevronUp className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-30 transition-opacity" />
                        )}
                    </div>
                </div>
            </TableHead>
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
            <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-bold text-foreground">
                    Discovered Services
                </h2>
                <span className="text-sm text-muted-foreground bg-muted px-3 py-1 rounded-full">
                    {sortedServices.length} service
                    {sortedServices.length !== 1 ? 's' : ''}
                </span>
            </div>

            <Card className="border-0 shadow-md">
                <CardContent className="p-0">
                    <Table>
                        <TableHeader>
                            <TableRow className="group">
                                <SortableHeader field="name">
                                    Service
                                </SortableHeader>
                                <SortableHeader field="namespace">
                                    Namespace
                                </SortableHeader>
                                <TableHead>Instances</TableHead>
                                <TableHead>Proxy Sidecars</TableHead>
                                <TableHead>Status</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {sortedServices.map((service) => {
                                const proxiedInstances =
                                    service.instances.filter(
                                        (i) => i.hasProxySidecar
                                    ).length;

                                return (
                                    <TableRow
                                        key={service.id}
                                        className="cursor-pointer hover:bg-muted/50"
                                        onClick={() =>
                                            onServiceSelect?.(service.id)
                                        }
                                    >
                                        <TableCell>
                                            <div className="flex items-center gap-2">
                                                <Server className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                                                <span className="font-medium text-foreground">
                                                    {service.name}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge
                                                variant="secondary"
                                                className="text-xs text-foreground bg-muted"
                                            >
                                                {service.namespace}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex items-center gap-2">
                                                <Database className="w-4 h-4 text-green-600 dark:text-green-400" />
                                                <span className="text-foreground">
                                                    {service.instances.length}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            {proxiedInstances > 0 ? (
                                                <div className="flex items-center gap-2">
                                                    <Shield className="w-4 h-4 text-orange-600 dark:text-orange-400" />
                                                    <span className="text-foreground">
                                                        {proxiedInstances}
                                                    </span>
                                                </div>
                                            ) : (
                                                <span className="text-muted-foreground">
                                                    —
                                                </span>
                                            )}
                                        </TableCell>
                                        <TableCell>
                                            <Badge
                                                variant="outline"
                                                className="text-green-700 border-green-300 bg-green-50 dark:text-green-400 dark:border-green-700 dark:bg-green-950"
                                            >
                                                Running
                                            </Badge>
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
};
