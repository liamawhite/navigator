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

import type { v1alpha1Service } from '../../types/generated/openapi';
import { Server, Database, Hexagon } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface ServiceCardProps {
    service: v1alpha1Service;
    onClick?: () => void;
}

export const ServiceCard: React.FC<ServiceCardProps> = ({
    service,
    onClick,
}) => {
    const proxiedInstances =
        service.instances?.filter((i) => i.envoyPresent).length || 0;

    return (
        <Card
            className="cursor-pointer hover:shadow-lg transition-all duration-200 border-0 shadow-md hover:scale-[1.02]"
            onClick={onClick}
        >
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-lg font-semibold text-slate-900 flex items-center gap-2">
                        <Server className="w-5 h-5 text-blue-500" />
                        {service.name}
                    </CardTitle>
                    <Badge variant="secondary" className="text-xs">
                        {service.namespace}
                    </Badge>
                </div>
            </CardHeader>

            <CardContent className="space-y-3">
                <div className="flex items-center gap-2">
                    <Database className="w-4 h-4 text-green-500" />
                    <span className="text-sm text-slate-600">
                        {service.instances?.length || 0} instance
                        {(service.instances?.length || 0) !== 1 ? 's' : ''}
                    </span>
                </div>

                {proxiedInstances > 0 && (
                    <div className="flex items-center gap-2">
                        <Hexagon className="w-4 h-4 text-orange-500" />
                        <span className="text-sm text-slate-600">
                            {proxiedInstances} with Envoy
                        </span>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};
