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

import { useParams, useNavigate } from 'react-router-dom';
import { useService } from '../hooks/useServices';
import { Navbar } from '../components/Navbar';
import {
    Server,
    Database,
    Circle,
    Copy,
    Activity,
    MapPin,
    Hexagon,
    Home,
    Network,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
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
    TooltipTrigger,
} from '@/components/ui/tooltip';
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { useState } from 'react';

export const ServiceDetailPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: service, isLoading, error } = useService(id!);
    const [copiedItem, setCopiedItem] = useState<string | null>(null);

    const copyToClipboard = async (text: string, itemId: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setCopiedItem(itemId);
            setTimeout(() => setCopiedItem(null), 2000);
        } catch (err) {
            console.error('Failed to copy text: ', err);
        }
    };

    if (isLoading) {
        return (
            <div className="min-h-screen bg-background">
                <Navbar />
                <div className="container mx-auto px-4 py-8">
                    <Breadcrumb className="mb-6">
                        <BreadcrumbList>
                            <BreadcrumbItem>
                                <BreadcrumbLink
                                    onClick={() => navigate('/')}
                                    className="cursor-pointer flex items-center gap-1"
                                >
                                    <Home className="w-4 h-4" />
                                    Services
                                </BreadcrumbLink>
                            </BreadcrumbItem>
                            <BreadcrumbSeparator />
                            <BreadcrumbItem>
                                <BreadcrumbPage>Loading...</BreadcrumbPage>
                            </BreadcrumbItem>
                        </BreadcrumbList>
                    </Breadcrumb>
                    <div className="animate-pulse space-y-6">
                        <div className="h-8 bg-muted rounded w-1/4"></div>
                        <Card>
                            <CardContent className="p-6">
                                <div className="space-y-4">
                                    <div className="h-8 bg-muted rounded w-1/3"></div>
                                    <div className="h-4 bg-muted rounded w-1/2"></div>
                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                        <div className="h-24 bg-muted rounded"></div>
                                        <div className="h-24 bg-muted rounded"></div>
                                        <div className="h-24 bg-muted rounded"></div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardContent className="p-6">
                                <div className="space-y-3">
                                    <div className="h-6 bg-muted rounded w-1/4"></div>
                                    <div className="h-20 bg-muted rounded"></div>
                                    <div className="h-20 bg-muted rounded"></div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        );
    }

    if (error || !service) {
        return (
            <div className="min-h-screen bg-background">
                <Navbar />
                <div className="container mx-auto px-4 py-8">
                    <Breadcrumb className="mb-6">
                        <BreadcrumbList>
                            <BreadcrumbItem>
                                <BreadcrumbLink
                                    onClick={() => navigate('/')}
                                    className="cursor-pointer flex items-center gap-1"
                                >
                                    <Home className="w-4 h-4" />
                                    Services
                                </BreadcrumbLink>
                            </BreadcrumbItem>
                            <BreadcrumbSeparator />
                            <BreadcrumbItem>
                                <BreadcrumbPage>Not Found</BreadcrumbPage>
                            </BreadcrumbItem>
                        </BreadcrumbList>
                    </Breadcrumb>
                    <Card>
                        <CardContent className="text-center py-12">
                            <Server className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                            <h3 className="text-lg font-semibold text-foreground mb-2">
                                Service not found
                            </h3>
                            <p className="text-muted-foreground">
                                The service "{id}" could not be found or no
                                longer exists.
                            </p>
                        </CardContent>
                    </Card>
                </div>
            </div>
        );
    }

    const proxiedInstances = service.instances.filter((i) => i.envoyPresent);
    const serviceMeshEnabled = proxiedInstances.length > 0;

    const uniqueClusters = [
        ...new Set(service.instances.map((i) => i.clusterName)),
    ];
    const clusterCount = uniqueClusters.length;

    return (
        <div className="min-h-screen bg-background">
            <Navbar />
            <div className="container mx-auto px-4 py-8">
                <Breadcrumb className="mb-6">
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            <BreadcrumbLink
                                onClick={() => navigate('/')}
                                className="cursor-pointer flex items-center gap-1"
                            >
                                <Home className="w-4 h-4" />
                                Services
                            </BreadcrumbLink>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                            <BreadcrumbPage>
                                {service?.name || id}
                            </BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>

                {/* Service Header */}
                <Card className="mb-6">
                    <CardHeader>
                        <div className="flex items-start justify-between">
                            <div>
                                <CardTitle className="text-3xl font-bold text-foreground flex items-center gap-3">
                                    <Server className="w-8 h-8 text-blue-500" />
                                    {service.name}
                                </CardTitle>
                                <div className="flex items-center gap-2 mt-2">
                                    <MapPin className="w-4 h-4 text-muted-foreground" />
                                    <span className="text-muted-foreground">
                                        Namespace:
                                    </span>
                                    <Badge variant="secondary">
                                        {service.namespace}
                                    </Badge>
                                </div>
                                <div className="flex items-center gap-2 mt-2">
                                    <Network className="w-4 h-4 text-muted-foreground" />
                                    <span className="text-muted-foreground">
                                        {clusterCount === 1
                                            ? 'Cluster:'
                                            : 'Clusters:'}
                                    </span>
                                    <div className="flex gap-1 flex-wrap">
                                        {uniqueClusters.map((cluster) => {
                                            const clusterIP =
                                                service.clusterIps?.[cluster];
                                            const externalIP =
                                                service.externalIps?.[cluster];

                                            // Show external IP if available, otherwise cluster IP
                                            const displayIP =
                                                externalIP || clusterIP;

                                            return (
                                                <Badge
                                                    key={cluster}
                                                    variant="outline"
                                                    className={
                                                        externalIP
                                                            ? 'border-green-500 text-green-700'
                                                            : ''
                                                    }
                                                >
                                                    {cluster}
                                                    {displayIP
                                                        ? `:${displayIP}`
                                                        : ''}
                                                </Badge>
                                            );
                                        })}
                                    </div>
                                    {clusterCount > 1 && (
                                        <Badge
                                            variant="secondary"
                                            className="ml-1"
                                        >
                                            {clusterCount} total
                                        </Badge>
                                    )}
                                </div>
                            </div>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() =>
                                            copyToClipboard(
                                                service.id,
                                                'service-id'
                                            )
                                        }
                                    >
                                        {copiedItem === 'service-id' ? (
                                            <>
                                                <Circle className="w-4 h-4 mr-2 fill-green-500 text-green-500" />
                                                Copied!
                                            </>
                                        ) : (
                                            <>
                                                <Copy className="w-4 h-4 mr-2" />
                                                Copy ID
                                            </>
                                        )}
                                    </Button>
                                </TooltipTrigger>
                                <TooltipContent>
                                    <p>Copy service ID: {service.id}</p>
                                </TooltipContent>
                            </Tooltip>
                        </div>
                    </CardHeader>
                </Card>

                {/* Service Metrics */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                Total Instances
                            </CardTitle>
                            <Database className="w-4 h-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">
                                {service.instances.length}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                Running pods backing this service
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                Envoy
                            </CardTitle>
                            <Hexagon className="w-4 h-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold flex items-center gap-2">
                                <Circle
                                    className={`w-3 h-3 fill-current ${serviceMeshEnabled ? 'text-green-500' : 'text-gray-400'}`}
                                />
                                {serviceMeshEnabled ? 'Enabled' : 'Disabled'}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                {proxiedInstances.length ===
                                service.instances.length
                                    ? 'Envoy present in all instances'
                                    : `Envoy present in ${proxiedInstances.length} of ${service.instances.length} instances`}
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                Health Status
                            </CardTitle>
                            <Activity className="w-4 h-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold flex items-center gap-2">
                                <Circle className="w-3 h-3 fill-green-500 text-green-500" />
                                Healthy
                            </div>
                            <p className="text-xs text-muted-foreground">
                                All instances are running
                            </p>
                        </CardContent>
                    </Card>
                </div>

                {/* Service Instances */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Database className="w-5 h-5 text-green-500" />
                            Service Instances
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        {service.instances.length > 0 ? (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Status</TableHead>
                                        <TableHead>IP Address</TableHead>
                                        <TableHead>Pod</TableHead>
                                        <TableHead>Namespace</TableHead>
                                        <TableHead>Cluster</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {service.instances.map(
                                        (instance, index) => (
                                            <TableRow
                                                key={index}
                                                className="hover:bg-muted/50 cursor-pointer"
                                                onClick={() =>
                                                    navigate(
                                                        `/services/${service.id}/instances/${instance.instanceId}`
                                                    )
                                                }
                                            >
                                                <TableCell>
                                                    <div className="flex items-center gap-1">
                                                        <Circle className="w-3 h-3 text-green-500 fill-current" />
                                                        {instance.envoyPresent && (
                                                            <Tooltip>
                                                                <TooltipTrigger asChild>
                                                                    <Hexagon className="w-3 h-3 text-purple-600 cursor-help" />
                                                                </TooltipTrigger>
                                                                <TooltipContent>
                                                                    <p>Envoy sidecar proxy is present</p>
                                                                </TooltipContent>
                                                            </Tooltip>
                                                        )}
                                                    </div>
                                                </TableCell>
                                                <TableCell>
                                                    <span className="font-mono text-sm">
                                                        {instance.ip}
                                                    </span>
                                                </TableCell>
                                                <TableCell>
                                                    <span className="font-mono text-sm">
                                                        {instance.podName ||
                                                            'N/A'}
                                                    </span>
                                                </TableCell>
                                                <TableCell>
                                                    <Badge variant="secondary">
                                                        {instance.namespace}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell>
                                                    <Badge variant="outline">
                                                        {instance.clusterName}
                                                    </Badge>
                                                </TableCell>
                                            </TableRow>
                                        )
                                    )}
                                </TableBody>
                            </Table>
                        ) : (
                            <div className="text-center py-8">
                                <Database className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                                <h3 className="text-lg font-semibold text-foreground mb-2">
                                    No instances available
                                </h3>
                                <p className="text-muted-foreground">
                                    This service currently has no running
                                    instances.
                                </p>
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
};
