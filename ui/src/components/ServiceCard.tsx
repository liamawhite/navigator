import type { Service } from '../types/service';
import { Server, Database, Shield } from 'lucide-react';

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
        <div
            className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer"
            onClick={onClick}
        >
            <div className="flex items-center justify-between mb-2">
                <h3 className="text-lg font-semibold text-gray-900">
                    {service.name}
                </h3>
                <span className="text-sm text-gray-500">
                    {service.namespace}
                </span>
            </div>

            <div className="flex items-center gap-2 mb-3">
                <Server className="w-4 h-4 text-blue-500" />
                <span className="text-sm text-gray-600">Service</span>
            </div>

            <div className="flex items-center gap-2 mb-2">
                <Database className="w-4 h-4 text-green-500" />
                <span className="text-sm text-gray-600">
                    {service.instances.length} instance
                    {service.instances.length !== 1 ? 's' : ''}
                </span>
            </div>

            {proxiedInstances > 0 && (
                <div className="flex items-center gap-2 mb-2">
                    <Shield className="w-4 h-4 text-orange-500" />
                    <span className="text-sm text-gray-600">
                        {proxiedInstances} with proxy sidecar
                    </span>
                </div>
            )}
        </div>
    );
};
