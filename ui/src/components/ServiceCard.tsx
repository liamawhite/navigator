import type { Service } from '../types/service';
import { Server, Database, Shield } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface ServiceCardProps {
    service: Service;
    onClick?: () => void;
}

export const ServiceCard: React.FC<ServiceCardProps> = ({
    service,
    onClick,
}) => {
    const proxiedInstances = service.instances.filter(
        (i) => i.hasProxySidecar
    ).length;

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
                        {service.instances.length} instance
                        {service.instances.length !== 1 ? 's' : ''}
                    </span>
                </div>

                {proxiedInstances > 0 && (
                    <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4 text-orange-500" />
                        <span className="text-sm text-slate-600">
                            {proxiedInstances} with proxy sidecar
                        </span>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};
