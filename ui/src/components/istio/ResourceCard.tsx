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

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ConfigActions } from '../envoy';

interface ResourceCardProps {
    name: string;
    namespace: string;
    resourceType: string;
    spec?: Record<string, unknown>;
    rawConfig: string;
}

export const ResourceCard: React.FC<ResourceCardProps> = ({
    name,
    namespace,
    resourceType,
    spec,
    rawConfig,
}) => {
    const renderSpecSummary = () => {
        if (!spec || typeof spec !== 'object') {
            return <span className="text-muted-foreground text-sm">No specification available</span>;
        }

        if (resourceType === 'Sidecar') {
            const workloadSelector = spec.workloadSelector?.labels || {};
            const ingress = spec.ingress || [];
            const egress = spec.egress || [];
            
            return (
                <div className="space-y-2">
                    {Object.keys(workloadSelector).length > 0 && (
                        <div>
                            <span className="text-xs text-muted-foreground">Workload Selector:</span>
                            <div className="flex flex-wrap gap-1 mt-1">
                                {Object.entries(workloadSelector).slice(0, 2).map(([key, value], i) => (
                                    <Badge key={i} variant="outline" className="text-xs">
                                        {key}={String(value)}
                                    </Badge>
                                ))}
                            </div>
                        </div>
                    )}
                    <div className="flex gap-4 text-xs text-muted-foreground">
                        <span>Ingress: {ingress.length}</span>
                        <span>Egress: {egress.length}</span>
                    </div>
                </div>
            );
        }

        if (resourceType === 'EnvoyFilter') {
            const workloadSelector = spec.workloadSelector?.labels || {};
            const configPatches = spec.configPatches || [];
            
            return (
                <div className="space-y-2">
                    {Object.keys(workloadSelector).length > 0 && (
                        <div>
                            <span className="text-xs text-muted-foreground">Workload Selector:</span>
                            <div className="flex flex-wrap gap-1 mt-1">
                                {Object.entries(workloadSelector).slice(0, 2).map(([key, value], i) => (
                                    <Badge key={i} variant="outline" className="text-xs">
                                        {key}={String(value)}
                                    </Badge>
                                ))}
                            </div>
                        </div>
                    )}
                    <div className="text-xs text-muted-foreground">
                        Config Patches: {configPatches.length}
                    </div>
                </div>
            );
        }

        if (resourceType === 'RequestAuthentication' || resourceType === 'PeerAuthentication') {
            const selector = spec.selector?.matchLabels || {};
            const rules = spec.rules || [];
            const action = spec.action;
            
            return (
                <div className="space-y-2">
                    {Object.keys(selector).length > 0 && (
                        <div>
                            <span className="text-xs text-muted-foreground">Selector:</span>
                            <div className="flex flex-wrap gap-1 mt-1">
                                {Object.entries(selector).slice(0, 2).map(([key, value], i) => (
                                    <Badge key={i} variant="outline" className="text-xs">
                                        {key}={String(value)}
                                    </Badge>
                                ))}
                            </div>
                        </div>
                    )}
                    <div className="flex gap-4 text-xs text-muted-foreground">
                        {rules.length > 0 && <span>Rules: {rules.length}</span>}
                        {action && <span>Action: {action}</span>}
                    </div>
                </div>
            );
        }

        if (resourceType === 'WasmPlugin') {
            const selector = spec.selector?.matchLabels || {};
            const url = spec.url;
            const phase = spec.phase;
            
            return (
                <div className="space-y-2">
                    {Object.keys(selector).length > 0 && (
                        <div>
                            <span className="text-xs text-muted-foreground">Selector:</span>
                            <div className="flex flex-wrap gap-1 mt-1">
                                {Object.entries(selector).slice(0, 2).map(([key, value], i) => (
                                    <Badge key={i} variant="outline" className="text-xs">
                                        {key}={String(value)}
                                    </Badge>
                                ))}
                            </div>
                        </div>
                    )}
                    <div className="text-xs text-muted-foreground">
                        {url && <div>URL: {url}</div>}
                        {phase && <div>Phase: {phase}</div>}
                    </div>
                </div>
            );
        }

        if (resourceType === 'ServiceEntry') {
            const hosts = spec.hosts || [];
            const ports = spec.ports || [];
            const location = spec.location;
            
            return (
                <div className="space-y-2">
                    {hosts.length > 0 && (
                        <div>
                            <span className="text-xs text-muted-foreground">Hosts:</span>
                            <div className="flex flex-wrap gap-1 mt-1">
                                {hosts.slice(0, 2).map((host: string, i: number) => (
                                    <Badge key={i} variant="outline" className="text-xs">
                                        {host}
                                    </Badge>
                                ))}
                                {hosts.length > 2 && (
                                    <Badge variant="outline" className="text-xs">
                                        +{hosts.length - 2} more
                                    </Badge>
                                )}
                            </div>
                        </div>
                    )}
                    <div className="flex gap-4 text-xs text-muted-foreground">
                        <span>Ports: {ports.length}</span>
                        {location && <span>Location: {location}</span>}
                    </div>
                </div>
            );
        }

        return (
            <div className="text-xs text-muted-foreground">
                Resource configuration available
            </div>
        );
    };

    return (
        <Card>
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-mono">
                        {name}
                    </CardTitle>
                    <div className="flex items-center gap-2">
                        <Badge variant="outline" className="text-xs">
                            {namespace}
                        </Badge>
                        <ConfigActions
                            name={name}
                            rawConfig={rawConfig}
                            configType={resourceType}
                            copyId={`${resourceType.toLowerCase()}-${name}`}
                        />
                    </div>
                </div>
            </CardHeader>
            <CardContent className="pt-0">
                {renderSpecSummary()}
            </CardContent>
        </Card>
    );
};