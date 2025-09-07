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

import { useState, useEffect } from 'react';
import { ModeToggle } from './mode-toggle';
import { ClusterSyncStatus } from './ClusterSyncStatus';
import { Button } from './ui/button';
import {
    Tooltip,
    TooltipTrigger,
    TooltipContent,
    TooltipProvider,
} from './ui/tooltip';
import { useLocation, Link } from 'react-router-dom';
import { List, Waypoints } from 'lucide-react';
import { serviceApi } from '../utils/api';
import type { v1alpha1ClusterSyncInfo } from '../types/generated/openapi-cluster_registry';

export const Navbar: React.FC = () => {
    const location = useLocation();
    const isTopologyView = location.pathname.startsWith('/topology');
    const [hasMetricsCapability, setHasMetricsCapability] = useState(false);
    const [loading, setLoading] = useState(true);

    const checkMetricsCapability = async () => {
        try {
            const clusterData = await serviceApi.listClusters();
            const hasAnyMetrics = clusterData.some(
                (cluster: v1alpha1ClusterSyncInfo) => cluster.metricsEnabled
            );
            setHasMetricsCapability(hasAnyMetrics);
        } catch (error) {
            console.error('Failed to check metrics capability:', error);
            setHasMetricsCapability(false);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        checkMetricsCapability();
        // Check every 30 seconds to stay in sync with cluster status
        const interval = setInterval(checkMetricsCapability, 30000);
        return () => clearInterval(interval);
    }, []);

    return (
        <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
            <div className="container mx-auto px-4">
                <div className="flex h-16 items-center justify-between">
                    <div className="flex items-center space-x-2">
                        <div className="w-9 h-9">
                            <img
                                src="/navigator.svg"
                                alt="Navigator"
                                className="w-full h-full"
                            />
                        </div>

                        <div className="flex items-center space-x-2">
                            <Button
                                variant={
                                    !isTopologyView ? 'secondary' : 'ghost'
                                }
                                size="sm"
                                asChild
                            >
                                <Link
                                    to="/"
                                    className="flex items-center gap-2"
                                >
                                    <List className="h-4 w-4" />
                                    Service Registry
                                </Link>
                            </Button>
                            {!loading && (
                                <>
                                    {hasMetricsCapability ? (
                                        <Button
                                            variant={
                                                isTopologyView
                                                    ? 'secondary'
                                                    : 'ghost'
                                            }
                                            size="sm"
                                            asChild
                                        >
                                            <Link
                                                to="/topology"
                                                className="flex items-center gap-2"
                                            >
                                                <Waypoints className="h-4 w-4" />
                                                Topology
                                            </Link>
                                        </Button>
                                    ) : (
                                        <TooltipProvider>
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        className="opacity-50 cursor-not-allowed"
                                                        onClick={(e) =>
                                                            e.preventDefault()
                                                        }
                                                    >
                                                        <span className="flex items-center gap-2">
                                                            <Waypoints className="h-4 w-4" />
                                                            Topology
                                                        </span>
                                                    </Button>
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    <p>
                                                        Topology view requires
                                                        at least one cluster
                                                        with metrics enabled
                                                    </p>
                                                </TooltipContent>
                                            </Tooltip>
                                        </TooltipProvider>
                                    )}
                                </>
                            )}
                        </div>
                    </div>

                    <div className="flex items-center gap-3">
                        <ClusterSyncStatus />
                        <ModeToggle />
                    </div>
                </div>
            </div>
        </nav>
    );
};
