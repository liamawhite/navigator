import { useParams, useNavigate } from 'react-router-dom';
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
    Monitor,
    Code,
    AlertCircle,
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
import { v1alpha1ProxyMode } from '@/types/generated/openapi-troubleshooting';

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
                            <BreadcrumbPage>{instance.pod}</BreadcrumbPage>
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
                                    {instance.pod}
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

                {/* Containers */}
                <Card className="mb-6">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Monitor className="w-5 h-5 text-green-500" />
                            Containers
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Image</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Ready</TableHead>
                                    <TableHead>Restarts</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {instance.containers.map((container, index) => (
                                    <TableRow key={index}>
                                        <TableCell>
                                            <span className="font-mono text-sm">
                                                {container.name}
                                            </span>
                                        </TableCell>
                                        <TableCell>
                                            <span className="font-mono text-sm">
                                                {container.image}
                                            </span>
                                        </TableCell>
                                        <TableCell>
                                            <Badge
                                                variant={
                                                    container.status ===
                                                    'Running'
                                                        ? 'default'
                                                        : 'secondary'
                                                }
                                            >
                                                {container.status}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <Circle
                                                className={`w-3 h-3 fill-current ${
                                                    container.ready
                                                        ? 'text-green-500'
                                                        : 'text-red-500'
                                                }`}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <span className="font-mono">
                                                {container.restartCount}
                                            </span>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>

                {/* Proxy Configuration */}
                {instance.isEnvoyPresent && (
                    <Card className="mb-6">
                        <CardHeader>
                            <CardTitle className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <Hexagon className="w-5 h-5 text-purple-500" />
                                    Proxy Configuration
                                </div>
                                {proxyConfig && (
                                    <ConfigActions
                                        name={
                                            proxyConfig.proxyConfig.bootstrap
                                                ?.node?.id ||
                                            'Proxy Configuration'
                                        }
                                        rawConfig={
                                            proxyConfig.proxyConfig
                                                .rawConfigDump || ''
                                        }
                                        configType="Configuration"
                                        copyId="full-config-dump"
                                    />
                                )}
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            {proxyLoading ? (
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
                                                        proxyConfig.proxyConfig
                                                            .bootstrap?.node
                                                            ?.proxyMode
                                                    )}
                                                >
                                                    {formatProxyMode(
                                                        proxyConfig.proxyConfig
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
                                        defaultValue="listeners"
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
                                            <div className="space-y-4">
                                                <h4 className="text-sm font-medium">
                                                    Listeners (
                                                    {proxyConfig.proxyConfig
                                                        .listeners?.length || 0}
                                                    )
                                                </h4>
                                                <ListenersTable
                                                    listeners={
                                                        proxyConfig.proxyConfig
                                                            .listeners || []
                                                    }
                                                />
                                            </div>
                                        </TabsContent>

                                        <TabsContent
                                            value="clusters"
                                            className="mt-4"
                                        >
                                            <div className="space-y-4">
                                                <h4 className="text-sm font-medium">
                                                    Clusters (
                                                    {proxyConfig.proxyConfig
                                                        .clusters?.length || 0}
                                                    )
                                                </h4>
                                                <ClustersTable
                                                    clusters={
                                                        proxyConfig.proxyConfig
                                                            .clusters || []
                                                    }
                                                />
                                            </div>
                                        </TabsContent>

                                        <TabsContent
                                            value="endpoints"
                                            className="mt-4"
                                        >
                                            <div className="space-y-4">
                                                <h4 className="text-sm font-medium">
                                                    Endpoints (
                                                    {proxyConfig.proxyConfig
                                                        .endpoints?.length || 0}
                                                    )
                                                </h4>
                                                <EndpointsTable
                                                    endpoints={
                                                        proxyConfig.proxyConfig
                                                            .endpoints || []
                                                    }
                                                />
                                            </div>
                                        </TabsContent>

                                        <TabsContent
                                            value="routes"
                                            className="mt-4"
                                        >
                                            <div className="space-y-4">
                                                <h4 className="text-sm font-medium">
                                                    Routes (
                                                    {proxyConfig.proxyConfig
                                                        .routes?.length || 0}
                                                    )
                                                </h4>
                                                <RoutesTable
                                                    routes={
                                                        proxyConfig.proxyConfig
                                                            .routes || []
                                                    }
                                                />
                                            </div>
                                        </TabsContent>

                                        <TabsContent
                                            value="bootstrap"
                                            className="mt-4"
                                        >
                                            <div className="space-y-4">
                                                <h4 className="text-sm font-medium">
                                                    Bootstrap Configuration
                                                </h4>
                                                <BootstrapConfig
                                                    bootstrap={
                                                        proxyConfig.proxyConfig
                                                            .bootstrap || null
                                                    }
                                                />
                                            </div>
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
                                        Unable to retrieve proxy configuration
                                        for this instance.
                                    </p>
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
