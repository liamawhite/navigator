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

import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { useServiceInstance, useProxyConfig } from '../hooks/useServices';
import { Navbar } from '../components/Navbar';
import {
    Server,
    Database,
    Circle,
    Copy,
    Activity,
    MapPin,
    Tag,
    Hexagon,
    Home,
    Globe,
    Calendar,
    HardDrive,
    Code,
    AlertCircle,
    Sailboat,
    Container,
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useState } from 'react';
import {
    ListenersTable,
    ClustersTable,
    EndpointsTable,
    RoutesTable,
    BootstrapConfig,
    ConfigActions,
} from '../components/envoy';
import { IstioResourcesView } from '../components/istio';
import { v1alpha1ProxyMode } from '@/types/generated/openapi-service_registry/index';

const formatProxyMode = (proxyMode: v1alpha1ProxyMode | undefined): string => {
    switch (proxyMode) {
        case v1alpha1ProxyMode.SIDECAR:
            return 'Sidecar';
        case v1alpha1ProxyMode.GATEWAY:
            return 'Gateway';
        case v1alpha1ProxyMode.ROUTER:
            return 'Router';
        case v1alpha1ProxyMode.UNKNOWN_PROXY_MODE:
        default:
            return 'Unknown';
    }
};

const getProxyModeVariant = (
    proxyMode: v1alpha1ProxyMode | undefined
): 'default' | 'secondary' | 'destructive' | 'outline' => {
    switch (proxyMode) {
        case v1alpha1ProxyMode.SIDECAR:
            return 'default';
        case v1alpha1ProxyMode.GATEWAY:
            return 'outline';
        case v1alpha1ProxyMode.ROUTER:
            return 'outline';
        case v1alpha1ProxyMode.UNKNOWN_PROXY_MODE:
        default:
            return 'secondary';
    }
};

export const ServiceInstanceDetailPage: React.FC = () => {
    const { serviceId, instanceId } = useParams<{
        serviceId: string;
        instanceId: string;
    }>();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();

    const availableTabs = [
        'listeners',
        'routes',
        'clusters',
        'endpoints',
        'bootstrap',
    ] as const;
    const currentTab = searchParams.get('proxy_config') || 'listeners';
    const validTab = availableTabs.includes(
        currentTab as (typeof availableTabs)[number]
    )
        ? currentTab
        : 'listeners';

    const {
        data: instance,
        isLoading,
        error,
    } = useServiceInstance(serviceId!, instanceId!);
    const { data: proxyConfig, isLoading: proxyLoading } = useProxyConfig(
        serviceId!,
        instanceId!
    );
    const [copiedItem, setCopiedItem] = useState<string | null>(null);

    // Get config view from URL params, default to 'proxy'
    const availableViews = ['proxy', 'istio', 'containers'] as const;
    const currentConfigView = searchParams.get('config_view') || 'proxy';
    const validConfigView = availableViews.includes(
        currentConfigView as (typeof availableViews)[number]
    )
        ? (currentConfigView as 'proxy' | 'istio' | 'containers')
        : 'proxy';

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
                    </div>
                </div>
            </div>
        );
    }

    if (error || !instance) {
        return (
            <div className="min-h-screen bg-background">
                <Navbar />
                <div className="container mx-auto px-4 py-8">
                    <Card>
                        <CardContent className="text-center py-12">
                            <Server className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                            <h3 className="text-lg font-semibold text-foreground mb-2">
                                Instance not found
                            </h3>
                            <p className="text-muted-foreground">
                                The service instance could not be found or no
                                longer exists.
                            </p>
                        </CardContent>
                    </Card>
                </div>
            </div>
        );
    }

    const formatDate = (dateString: string) => {
        try {
            return new Date(dateString).toLocaleString();
        } catch {
            return dateString;
        }
    };

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
                            <BreadcrumbLink
                                onClick={() =>
                                    navigate(`/services/${serviceId}`)
                                }
                                className="cursor-pointer"
                            >
                                {instance.serviceName}
                            </BreadcrumbLink>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                            <BreadcrumbPage>{instance.podName}</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>

                {/* Instance Header */}
                <Card className="mb-6">
                    <CardHeader>
                        <div className="flex items-start justify-between">
                            <div>
                                <CardTitle className="text-3xl font-bold text-foreground flex items-center gap-3">
                                    <Database className="w-8 h-8 text-blue-500" />
                                    {instance.podName}
                                </CardTitle>
                                <div className="flex flex-wrap items-center gap-4 mt-3">
                                    <div className="flex items-center gap-2">
                                        <MapPin className="w-4 h-4 text-muted-foreground" />
                                        <span className="text-muted-foreground">
                                            Namespace:
                                        </span>
                                        <Badge variant="secondary">
                                            {instance.namespace}
                                        </Badge>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Globe className="w-4 h-4 text-muted-foreground" />
                                        <span className="text-muted-foreground">
                                            Cluster:
                                        </span>
                                        <Badge variant="outline">
                                            {instance.clusterName}
                                        </Badge>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <HardDrive className="w-4 h-4 text-muted-foreground" />
                                        <span className="text-muted-foreground">
                                            Node:
                                        </span>
                                        <Badge variant="outline">
                                            {instance.nodeName}
                                        </Badge>
                                    </div>
                                </div>
                            </div>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() =>
                                            copyToClipboard(
                                                instance.instanceId,
                                                'instance-id'
                                            )
                                        }
                                    >
                                        {copiedItem === 'instance-id' ? (
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
                                    <p>
                                        Copy instance ID: {instance.instanceId}
                                    </p>
                                </TooltipContent>
                            </Tooltip>
                        </div>
                    </CardHeader>
                </Card>

                {/* Instance Metrics */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                Pod Status
                            </CardTitle>
                            <Activity className="w-4 h-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold flex items-center gap-2">
                                <Circle
                                    className={`w-3 h-3 fill-current ${
                                        instance.podStatus === 'Running'
                                            ? 'text-green-500'
                                            : 'text-yellow-500'
                                    }`}
                                />
                                {instance.podStatus}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                Current pod status
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                IP Address
                            </CardTitle>
                            <Globe className="w-4 h-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold font-mono">
                                {instance.ip}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                Pod network address
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                Envoy Proxy
                            </CardTitle>
                            <Hexagon className="w-4 h-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold flex items-center gap-2">
                                <Circle
                                    className={`w-3 h-3 fill-current ${
                                        instance.isEnvoyPresent
                                            ? 'text-green-500'
                                            : 'text-gray-400'
                                    }`}
                                />
                                {instance.isEnvoyPresent
                                    ? 'Present'
                                    : 'Not Present'}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                Service mesh sidecar
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                Created
                            </CardTitle>
                            <Calendar className="w-4 h-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-sm font-bold">
                                {formatDate(instance.createdAt)}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                Pod creation time
                            </p>
                        </CardContent>
                    </Card>
                </div>

                {/* Service Mesh Configuration */}
                {(instance.isEnvoyPresent ||
                    (instance.containers &&
                        instance.containers.length > 0)) && (
                    <Card className="mb-6">
                        <CardHeader>
                            <CardTitle className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <div className="flex items-center gap-3">
                                        <button
                                            onClick={() => {
                                                setSearchParams((prev) => {
                                                    const newParams =
                                                        new URLSearchParams(
                                                            prev
                                                        );
                                                    newParams.set(
                                                        'config_view',
                                                        'proxy'
                                                    );
                                                    // Keep existing proxy_config tab if it exists
                                                    if (
                                                        !prev.has(
                                                            'proxy_config'
                                                        )
                                                    ) {
                                                        newParams.set(
                                                            'proxy_config',
                                                            'listeners'
                                                        );
                                                    }
                                                    return newParams;
                                                });
                                            }}
                                            className={`flex items-center gap-2 px-3 py-1.5 rounded-md transition-colors cursor-pointer ${
                                                validConfigView === 'proxy'
                                                    ? 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300'
                                                    : 'text-muted-foreground hover:text-foreground'
                                            }`}
                                        >
                                            <Hexagon
                                                className={`w-4 h-4 ${
                                                    validConfigView === 'proxy'
                                                        ? ''
                                                        : 'text-purple-500'
                                                }`}
                                            />
                                            Proxy Configuration
                                            <sup className="text-xs text-purple-500 font-medium -ml-1.5">
                                                beta
                                            </sup>
                                        </button>
                                        <button
                                            onClick={() => {
                                                setSearchParams((prev) => {
                                                    const newParams =
                                                        new URLSearchParams(
                                                            prev
                                                        );
                                                    newParams.set(
                                                        'config_view',
                                                        'istio'
                                                    );
                                                    // Keep existing istio_tab if it exists
                                                    if (
                                                        !prev.has('istio_tab')
                                                    ) {
                                                        newParams.set(
                                                            'istio_tab',
                                                            'traffic'
                                                        );
                                                    }
                                                    return newParams;
                                                });
                                            }}
                                            className={`flex items-center gap-2 px-3 py-1.5 rounded-md transition-colors cursor-pointer ${
                                                validConfigView === 'istio'
                                                    ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                                                    : 'text-muted-foreground hover:text-foreground'
                                            }`}
                                        >
                                            <Sailboat
                                                className={`w-4 h-4 ${
                                                    validConfigView === 'istio'
                                                        ? ''
                                                        : 'text-blue-500'
                                                }`}
                                            />
                                            Istio Resources
                                            <sup className="text-xs text-blue-500 font-medium -ml-1.5">
                                                alpha
                                            </sup>
                                        </button>
                                        <button
                                            onClick={() => {
                                                setSearchParams((prev) => {
                                                    const newParams =
                                                        new URLSearchParams(
                                                            prev
                                                        );
                                                    newParams.set(
                                                        'config_view',
                                                        'containers'
                                                    );
                                                    return newParams;
                                                });
                                            }}
                                            className={`flex items-center gap-2 px-3 py-1.5 rounded-md transition-colors cursor-pointer ${
                                                validConfigView === 'containers'
                                                    ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                                                    : 'text-muted-foreground hover:text-foreground'
                                            }`}
                                        >
                                            <Container
                                                className={`w-4 h-4 ${
                                                    validConfigView ===
                                                    'containers'
                                                        ? ''
                                                        : 'text-green-500'
                                                }`}
                                            />
                                            Containers
                                        </button>
                                    </div>
                                </div>
                                {proxyConfig && validConfigView === 'proxy' && (
                                    <div className="flex items-center gap-3">
                                        <div className="flex items-center gap-2 bg-muted/50 rounded-md px-3 py-1.5 border">
                                            <Code className="w-3 h-3 text-blue-500" />
                                            <span className="text-xs font-medium text-muted-foreground">
                                                /config_dump
                                            </span>
                                            <ConfigActions
                                                name={
                                                    proxyConfig.proxyConfig
                                                        .bootstrap?.node?.id ||
                                                    'Proxy Configuration'
                                                }
                                                rawConfig={
                                                    proxyConfig.proxyConfig
                                                        .rawConfigDump || ''
                                                }
                                                configType="Configuration"
                                                copyId="full-config-dump"
                                            />
                                        </div>
                                        <div className="flex items-center gap-2 bg-muted/50 rounded-md px-3 py-1.5 border">
                                            <Database className="w-3 h-3 text-green-500" />
                                            <span className="text-xs font-medium text-muted-foreground">
                                                /clusters
                                            </span>
                                            <ConfigActions
                                                name="Raw Clusters"
                                                rawConfig={
                                                    proxyConfig.proxyConfig
                                                        .rawClusters || ''
                                                }
                                                configType="Clusters"
                                                copyId="raw-clusters"
                                            />
                                        </div>
                                    </div>
                                )}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            {validConfigView === 'proxy' ? (
                                proxyLoading ? (
                                    <div className="animate-pulse">
                                        <div className="h-4 bg-muted rounded w-1/4 mb-2"></div>
                                        <div className="h-32 bg-muted rounded"></div>
                                    </div>
                                ) : proxyConfig ? (
                                    <div className="space-y-4">
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                            <div>
                                                <label className="text-sm font-medium text-muted-foreground">
                                                    Proxy Mode
                                                </label>
                                                <div className="text-sm">
                                                    <Badge
                                                        variant={getProxyModeVariant(
                                                            proxyConfig
                                                                .proxyConfig
                                                                .bootstrap?.node
                                                                ?.proxyMode
                                                        )}
                                                    >
                                                        {formatProxyMode(
                                                            proxyConfig
                                                                .proxyConfig
                                                                .bootstrap?.node
                                                                ?.proxyMode
                                                        )}
                                                    </Badge>
                                                </div>
                                            </div>
                                            <div>
                                                <label className="text-sm font-medium text-muted-foreground">
                                                    Version
                                                </label>
                                                <div className="text-sm font-mono">
                                                    {
                                                        proxyConfig.proxyConfig
                                                            .version
                                                    }
                                                </div>
                                            </div>
                                        </div>

                                        <Tabs
                                            value={validTab}
                                            onValueChange={(tab) => {
                                                setSearchParams((prev) => {
                                                    const newParams =
                                                        new URLSearchParams(
                                                            prev
                                                        );
                                                    newParams.set(
                                                        'proxy_config',
                                                        tab
                                                    );
                                                    return newParams;
                                                });
                                            }}
                                            className="w-full"
                                        >
                                            <TabsList className="grid w-full grid-cols-5">
                                                <TabsTrigger
                                                    value="listeners"
                                                    className="cursor-pointer"
                                                >
                                                    Listeners
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value="routes"
                                                    className="cursor-pointer"
                                                >
                                                    Routes
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value="clusters"
                                                    className="cursor-pointer"
                                                >
                                                    Clusters
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value="endpoints"
                                                    className="cursor-pointer"
                                                >
                                                    Endpoints
                                                </TabsTrigger>
                                                <TabsTrigger
                                                    value="bootstrap"
                                                    className="cursor-pointer"
                                                >
                                                    Bootstrap
                                                </TabsTrigger>
                                            </TabsList>

                                            <TabsContent
                                                value="listeners"
                                                className="mt-4"
                                            >
                                                <ListenersTable
                                                    listeners={
                                                        proxyConfig.proxyConfig
                                                            .listeners || []
                                                    }
                                                    proxyMode={
                                                        proxyConfig.proxyConfig
                                                            .bootstrap?.node
                                                            ?.proxyMode
                                                    }
                                                    serviceId={serviceId}
                                                />
                                            </TabsContent>

                                            <TabsContent
                                                value="clusters"
                                                className="mt-4"
                                            >
                                                <ClustersTable
                                                    clusters={
                                                        proxyConfig.proxyConfig
                                                            .clusters || []
                                                    }
                                                    serviceId={serviceId}
                                                />
                                            </TabsContent>

                                            <TabsContent
                                                value="endpoints"
                                                className="mt-4"
                                            >
                                                <EndpointsTable
                                                    endpoints={
                                                        proxyConfig.proxyConfig
                                                            .endpoints || []
                                                    }
                                                    serviceId={serviceId}
                                                />
                                            </TabsContent>

                                            <TabsContent
                                                value="routes"
                                                className="mt-4"
                                            >
                                                <RoutesTable
                                                    routes={
                                                        proxyConfig.proxyConfig
                                                            .routes || []
                                                    }
                                                    serviceId={serviceId}
                                                />
                                            </TabsContent>

                                            <TabsContent
                                                value="bootstrap"
                                                className="mt-4"
                                            >
                                                <BootstrapConfig
                                                    bootstrap={
                                                        proxyConfig.proxyConfig
                                                            .bootstrap || null
                                                    }
                                                />
                                            </TabsContent>
                                        </Tabs>
                                    </div>
                                ) : (
                                    <div className="text-center py-8">
                                        <AlertCircle className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                                        <h3 className="text-lg font-semibold text-foreground mb-2">
                                            Configuration not available
                                        </h3>
                                        <p className="text-muted-foreground">
                                            Unable to retrieve proxy
                                            configuration for this instance.
                                        </p>
                                    </div>
                                )
                            ) : validConfigView === 'istio' ? (
                                <IstioResourcesView
                                    serviceId={serviceId!}
                                    instanceId={instanceId!}
                                />
                            ) : (
                                <div className="space-y-2">
                                    <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                                        <Container className="w-4 h-4 text-green-500" />
                                        Containers (
                                        {(instance.containers || []).length})
                                    </h4>
                                    <Table>
                                        <TableHeader>
                                            <TableRow>
                                                <TableHead>Status</TableHead>
                                                <TableHead>Name</TableHead>
                                                <TableHead>Image</TableHead>
                                                <TableHead>Restarts</TableHead>
                                            </TableRow>
                                        </TableHeader>
                                        <TableBody>
                                            {(instance.containers || []).map(
                                                (container, index) => (
                                                    <TableRow key={index}>
                                                        <TableCell>
                                                            <div className="flex items-center gap-1">
                                                                <Circle
                                                                    className={`w-3 h-3 fill-current ${
                                                                        container.ready &&
                                                                        container.status ===
                                                                            'Running'
                                                                            ? 'text-green-500'
                                                                            : container.ready
                                                                              ? 'text-yellow-500'
                                                                              : 'text-red-500'
                                                                    }`}
                                                                />
                                                            </div>
                                                        </TableCell>
                                                        <TableCell>
                                                            <span className="font-mono text-sm">
                                                                {container.name}
                                                            </span>
                                                        </TableCell>
                                                        <TableCell>
                                                            <span className="font-mono text-sm">
                                                                {
                                                                    container.image
                                                                }
                                                            </span>
                                                        </TableCell>
                                                        <TableCell>
                                                            <span className="font-mono">
                                                                {
                                                                    container.restartCount
                                                                }
                                                            </span>
                                                        </TableCell>
                                                    </TableRow>
                                                )
                                            )}
                                        </TableBody>
                                    </Table>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                )}

                {/* Labels and Annotations */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Tag className="w-5 h-5 text-blue-500" />
                                Labels
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            {Object.keys(instance.labels).length > 0 ? (
                                <div className="space-y-2">
                                    {Object.entries(instance.labels).map(
                                        ([key, value]) => (
                                            <div
                                                key={key}
                                                className="flex justify-between items-center"
                                            >
                                                <span className="text-sm font-medium text-muted-foreground">
                                                    {key}
                                                </span>
                                                <Badge
                                                    variant="outline"
                                                    className="font-mono text-xs"
                                                >
                                                    {value}
                                                </Badge>
                                            </div>
                                        )
                                    )}
                                </div>
                            ) : (
                                <p className="text-muted-foreground text-sm">
                                    No labels
                                </p>
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Code className="w-5 h-5 text-orange-500" />
                                Annotations
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            {Object.keys(instance.annotations).length > 0 ? (
                                <div className="space-y-2 max-h-64 overflow-y-auto">
                                    {Object.entries(instance.annotations).map(
                                        ([key, value]) => (
                                            <div
                                                key={key}
                                                className="flex justify-between items-start"
                                            >
                                                <span className="text-sm font-medium text-muted-foreground break-all mr-2">
                                                    {key}
                                                </span>
                                                <span className="text-xs font-mono bg-muted px-2 py-1 rounded max-w-xs break-all">
                                                    {value}
                                                </span>
                                            </div>
                                        )
                                    )}
                                </div>
                            ) : (
                                <p className="text-muted-foreground text-sm">
                                    No annotations
                                </p>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
};
