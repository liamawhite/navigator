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

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Network, AlertCircle } from 'lucide-react';
import { useServiceConnections } from '../../hooks/useServiceConnections';
import { ServiceConnectionsVisualization } from './ServiceConnectionsVisualization';

interface ServiceConnectionsCardProps {
    serviceName: string;
    namespace: string;
}

export const ServiceConnectionsCard: React.FC<ServiceConnectionsCardProps> = ({
    serviceName,
    namespace,
}) => {
    const {
        data: connections,
        isLoading,
        error,
    } = useServiceConnections(serviceName, namespace);

    // Always show the service connections card content - we know metrics are enabled if we get here
    const showCollapsed = false;

    return (
        <Card className={`mb-6 ${showCollapsed ? 'opacity-50' : ''}`}>
            <CardHeader>
                <CardTitle className="flex items-center justify-between -mb-1.5">
                    <div className="flex items-center gap-2">
                        <Network className="w-5 h-5 text-purple-600" />
                        Service Connections
                        <sup className="text-xs text-purple-500 font-medium -ml-1.5">
                            alpha
                        </sup>
                    </div>
                    {showCollapsed && (
                        <div className="flex items-center gap-1.5 text-muted-foreground text-sm font-normal">
                            <AlertCircle className="w-4 h-4" />
                            <span>
                                Requires metrics to be enabled on at least one
                                cluster
                            </span>
                        </div>
                    )}
                </CardTitle>
            </CardHeader>
            <CardContent>
                {isLoading ? (
                    <div className="flex items-center justify-center h-64">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-600"></div>
                    </div>
                ) : error ? (
                    <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
                        <Network className="w-16 h-16 mb-4" />
                        <p className="text-center">
                            Failed to load service connections
                        </p>
                        <p className="text-sm text-center mt-2">
                            {error instanceof Error
                                ? error.message
                                : 'Unknown error'}
                        </p>
                    </div>
                ) : connections &&
                  (connections.inbound?.length ||
                      connections.outbound?.length) ? (
                    <ServiceConnectionsVisualization
                        serviceName={serviceName}
                        namespace={namespace}
                        inbound={connections.inbound || []}
                        outbound={connections.outbound || []}
                    />
                ) : (
                    <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
                        <Network className="w-16 h-16 mb-4" />
                        <p className="text-center">
                            No service connections found
                        </p>
                        <p className="text-sm text-center mt-2">
                            This service has no inbound or outbound traffic in
                            the last 5 minutes
                        </p>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};
