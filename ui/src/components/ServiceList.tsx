import { useServices } from '../hooks/useServices';
import { ServiceCard } from './ServiceCard';
import { Loader2, AlertCircle, Server } from 'lucide-react';

interface ServiceListProps {
    onServiceSelect?: (serviceId: string) => void;
}

export const ServiceList: React.FC<ServiceListProps> = ({
    onServiceSelect,
}) => {
    const { data: services, isLoading, error, isError } = useServices();

    if (isLoading) {
        return (
            <div className="flex items-center justify-center py-12">
                <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
                <span className="ml-2 text-gray-600">Loading services...</span>
            </div>
        );
    }

    if (isError) {
        return (
            <div className="flex items-center justify-center py-12">
                <AlertCircle className="w-8 h-8 text-red-500" />
                <span className="ml-2 text-red-600">
                    Failed to load services: {error?.message}
                </span>
            </div>
        );
    }

    if (!services || services.length === 0) {
        return (
            <div className="text-center py-12">
                <Server className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-600">No services found</p>
            </div>
        );
    }

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {services.map((service) => (
                <ServiceCard
                    key={service.id}
                    service={service}
                    onClick={() => onServiceSelect?.(service.id)}
                />
            ))}
        </div>
    );
};
