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

import { useSearchParams } from 'react-router-dom';
import { useIstioResources } from '../../hooks/useServices';
import { Card, CardContent } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
    AlertCircle,
    Network,
    Route,
    Settings,
    Globe,
    ShieldCheck,
} from 'lucide-react';
import { VirtualServicesTable } from './VirtualServicesTable';
import { DestinationRulesTable } from './DestinationRulesTable';
import { GatewaysTable } from './GatewaysTable';
import { RequestAuthenticationsTable } from './RequestAuthenticationsTable';
import { PeerAuthenticationsTable } from './PeerAuthenticationsTable';
import { AuthorizationPoliciesTable } from './AuthorizationPoliciesTable';
import { SidecarsTable } from './SidecarsTable';
import { EnvoyFiltersTable } from './EnvoyFiltersTable';
import { WasmPluginsTable } from './WasmPluginsTable';
import { ResourceCard } from './ResourceCard';

interface IstioResourcesViewProps {
    serviceId: string;
    instanceId: string;
}

export const IstioResourcesView: React.FC<IstioResourcesViewProps> = ({
    serviceId,
    instanceId,
}) => {
    const {
        data: istioResources,
        isLoading,
        error,
    } = useIstioResources(serviceId, instanceId);
    const [searchParams, setSearchParams] = useSearchParams();

    // Get Istio tab from URL params, default to 'traffic'
    const availableIstioTabs = ['traffic', 'security', 'extensions'] as const;
    const currentIstioTab = searchParams.get('istio_tab') || 'traffic';
    const validIstioTab = availableIstioTabs.includes(
        currentIstioTab as (typeof availableIstioTabs)[number]
    )
        ? (currentIstioTab as 'traffic' | 'security' | 'extensions')
        : 'traffic';

    if (isLoading) {
        return (
            <div className="animate-pulse space-y-6">
                <div className="h-4 bg-muted rounded w-1/4"></div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="h-32 bg-muted rounded"></div>
                    <div className="h-32 bg-muted rounded"></div>
                    <div className="h-32 bg-muted rounded"></div>
                </div>
                <div className="h-64 bg-muted rounded"></div>
            </div>
        );
    }

    if (error || !istioResources) {
        return (
            <Card>
                <CardContent className="text-center py-12">
                    <AlertCircle className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-foreground mb-2">
                        Unable to load Istio resources
                    </h3>
                    <p className="text-muted-foreground">
                        Failed to retrieve Istio configuration for this
                        instance.
                    </p>
                </CardContent>
            </Card>
        );
    }

    const trafficResources = [
        ...(istioResources.gateways || []),
        ...(istioResources.virtualServices || []),
        ...(istioResources.destinationRules || []),
        ...(istioResources.serviceEntries || []),
    ];

    const securityResources = [
        ...(istioResources.requestAuthentications || []),
        ...(istioResources.peerAuthentications || []),
        ...(istioResources.authorizationPolicies || []),
    ];

    const extensionResources = [
        ...(istioResources.sidecars || []),
        ...(istioResources.envoyFilters || []),
        ...(istioResources.wasmPlugins || []),
    ];

    const totalResources =
        trafficResources.length +
        securityResources.length +
        extensionResources.length;

    if (totalResources === 0) {
        return (
            <div className="text-center py-12">
                <Network className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-foreground mb-2">
                    No Istio resources found
                </h3>
                <p className="text-muted-foreground">
                    No Istio configuration resources affect this service
                    instance.
                </p>
            </div>
        );
    }

    return (
        <Tabs
            value={validIstioTab}
            onValueChange={(tab) => {
                setSearchParams((prev) => {
                    const newParams = new URLSearchParams(prev);
                    newParams.set('istio_tab', tab);
                    return newParams;
                });
            }}
            className="w-full"
        >
            <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="traffic" className="cursor-pointer">
                    Traffic Management ({trafficResources.length})
                </TabsTrigger>
                <TabsTrigger value="security" className="cursor-pointer">
                    Security ({securityResources.length})
                </TabsTrigger>
                <TabsTrigger value="extensions" className="cursor-pointer">
                    Proxy Extensions ({extensionResources.length})
                </TabsTrigger>
            </TabsList>

            <TabsContent value="traffic" className="mt-4">
                <div className="space-y-6">
                    {istioResources.gateways &&
                        istioResources.gateways.length > 0 && (
                            <GatewaysTable gateways={istioResources.gateways} />
                        )}

                    {istioResources.virtualServices &&
                        istioResources.virtualServices.length > 0 && (
                            <VirtualServicesTable
                                virtualServices={istioResources.virtualServices}
                            />
                        )}

                    {istioResources.destinationRules &&
                        istioResources.destinationRules.length > 0 && (
                            <DestinationRulesTable
                                destinationRules={
                                    istioResources.destinationRules
                                }
                            />
                        )}

                    {istioResources.serviceEntries &&
                        istioResources.serviceEntries.length > 0 && (
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                                    <Globe className="w-4 h-4 text-teal-500" />
                                    ServiceEntries (
                                    {istioResources.serviceEntries.length})
                                </h4>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    {istioResources.serviceEntries.map(
                                        (entry, index) => (
                                            <ResourceCard
                                                key={index}
                                                name={entry.name || 'Unknown'}
                                                namespace={
                                                    entry.namespace || 'Unknown'
                                                }
                                                resourceType="ServiceEntry"
                                                spec={entry.spec}
                                                rawConfig={
                                                    entry.rawConfig || ''
                                                }
                                            />
                                        )
                                    )}
                                </div>
                            </div>
                        )}

                    {/* Show missing resource types */}
                    <div className="space-y-2">
                        {(!istioResources.gateways ||
                            istioResources.gateways.length === 0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No Gateways matched for this instance
                            </div>
                        )}
                        {(!istioResources.virtualServices ||
                            istioResources.virtualServices.length === 0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No VirtualServices matched for this instance
                            </div>
                        )}
                        {(!istioResources.destinationRules ||
                            istioResources.destinationRules.length === 0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No DestinationRules matched for this instance
                            </div>
                        )}
                        {(!istioResources.serviceEntries ||
                            istioResources.serviceEntries.length === 0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No ServiceEntries matched for this instance
                            </div>
                        )}
                    </div>

                    {trafficResources.length === 0 && (
                        <div className="text-center py-8">
                            <Route className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                            <h3 className="text-sm font-medium text-foreground mb-2">
                                No traffic management resources
                            </h3>
                            <p className="text-sm text-muted-foreground">
                                No Gateways, VirtualServices, DestinationRules,
                                or ServiceEntries affect this instance.
                            </p>
                        </div>
                    )}
                </div>
            </TabsContent>

            <TabsContent value="security" className="mt-4">
                <div className="space-y-6">
                    {istioResources.requestAuthentications &&
                        istioResources.requestAuthentications.length > 0 && (
                            <RequestAuthenticationsTable
                                requestAuthentications={
                                    istioResources.requestAuthentications
                                }
                            />
                        )}

                    {istioResources.peerAuthentications &&
                        istioResources.peerAuthentications.length > 0 && (
                            <PeerAuthenticationsTable
                                peerAuthentications={
                                    istioResources.peerAuthentications
                                }
                            />
                        )}

                    {istioResources.authorizationPolicies &&
                        istioResources.authorizationPolicies.length > 0 && (
                            <AuthorizationPoliciesTable
                                authorizationPolicies={
                                    istioResources.authorizationPolicies
                                }
                            />
                        )}

                    {/* Show missing resource types */}
                    <div className="space-y-2">
                        {(!istioResources.requestAuthentications ||
                            istioResources.requestAuthentications.length ===
                                0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No RequestAuthentications matched for this
                                instance
                            </div>
                        )}
                        {(!istioResources.peerAuthentications ||
                            istioResources.peerAuthentications.length ===
                                0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No PeerAuthentications matched for this instance
                            </div>
                        )}
                        {(!istioResources.authorizationPolicies ||
                            istioResources.authorizationPolicies.length ===
                                0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No AuthorizationPolicies matched for this instance
                            </div>
                        )}
                    </div>

                    {securityResources.length === 0 && (
                        <div className="text-center py-8">
                            <ShieldCheck className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                            <h3 className="text-sm font-medium text-foreground mb-2">
                                No security resources
                            </h3>
                            <p className="text-sm text-muted-foreground">
                                No RequestAuthentications, PeerAuthentications, or AuthorizationPolicies
                                affect this instance.
                            </p>
                        </div>
                    )}
                </div>
            </TabsContent>

            <TabsContent value="extensions" className="mt-4">
                <div className="space-y-6">
                    {istioResources.sidecars &&
                        istioResources.sidecars.length > 0 && (
                            <SidecarsTable sidecars={istioResources.sidecars} />
                        )}

                    {istioResources.envoyFilters &&
                        istioResources.envoyFilters.length > 0 && (
                            <EnvoyFiltersTable
                                envoyFilters={istioResources.envoyFilters}
                            />
                        )}

                    {istioResources.wasmPlugins &&
                        istioResources.wasmPlugins.length > 0 && (
                            <WasmPluginsTable
                                wasmPlugins={istioResources.wasmPlugins}
                            />
                        )}

                    {/* Show missing resource types */}
                    <div className="space-y-2">
                        {(!istioResources.sidecars ||
                            istioResources.sidecars.length === 0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No Sidecars matched for this instance
                            </div>
                        )}
                        {(!istioResources.envoyFilters ||
                            istioResources.envoyFilters.length === 0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No EnvoyFilters matched for this instance
                            </div>
                        )}
                        {(!istioResources.wasmPlugins ||
                            istioResources.wasmPlugins.length === 0) && (
                            <div className="text-xs text-muted-foreground bg-muted/20 rounded px-3 py-2">
                                No WasmPlugins matched for this instance
                            </div>
                        )}
                    </div>

                    {extensionResources.length === 0 && (
                        <div className="text-center py-8">
                            <Settings className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                            <h3 className="text-sm font-medium text-foreground mb-2">
                                No proxy extensions
                            </h3>
                            <p className="text-sm text-muted-foreground">
                                No Sidecars, EnvoyFilters, or WasmPlugins affect
                                this instance.
                            </p>
                        </div>
                    )}
                </div>
            </TabsContent>
        </Tabs>
    );
};
