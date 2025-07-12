import { useParams, useNavigate } from 'react-router-dom';
import { useService } from '../hooks/useServices';
import { ArrowLeft, Server, Database, Shield, Circle } from 'lucide-react';

export const ServiceDetailPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { data: service, isLoading, error } = useService(id!);

    if (isLoading) {
        return (
            <div className="container mx-auto px-4 py-8">
                <div className="animate-pulse">
                    <div className="h-8 bg-gray-200 rounded w-1/3 mb-4"></div>
                    <div className="h-4 bg-gray-200 rounded w-1/2 mb-8"></div>
                    <div className="space-y-4">
                        <div className="h-32 bg-gray-200 rounded"></div>
                        <div className="h-32 bg-gray-200 rounded"></div>
                    </div>
                </div>
            </div>
        );
    }

    if (error || !service) {
        return (
            <div className="container mx-auto px-4 py-8">
                <button
                    onClick={() => navigate('/')}
                    className="flex items-center text-blue-600 hover:text-blue-800 mb-4"
                >
                    <ArrowLeft className="w-4 h-4 mr-2" />
                    Back to Services
                </button>
                <div className="text-center py-12">
                    <p className="text-red-600">Service not found</p>
                </div>
            </div>
        );
    }

    const proxiedInstances = service.instances.filter((i) => i.hasProxySidecar);
    const directInstances = service.instances.filter((i) => !i.hasProxySidecar);

    return (
        <div className="container mx-auto px-4 py-8">
            <button
                onClick={() => navigate('/')}
                className="flex items-center text-blue-600 hover:text-blue-800 mb-6"
            >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Services
            </button>

            <div className="bg-white border border-gray-200 rounded-lg p-6">
                <div className="flex items-center justify-between mb-6">
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900">
                            {service.name}
                        </h1>
                        <p className="text-gray-600">
                            Namespace: {service.namespace}
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        <Server className="w-5 h-5 text-blue-500" />
                        <span className="text-sm font-medium text-gray-700">
                            Service
                        </span>
                    </div>
                </div>

                <div className="grid grid-cols-1 gap-6">
                    <div>
                        <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
                            <Database className="w-5 h-5 text-green-500" />
                            Service Instances ({service.instances.length})
                        </h2>

                        {directInstances.length > 0 && (
                            <div className="mb-4">
                                <h3 className="text-md font-medium text-gray-900 mb-2">
                                    Direct Instances
                                </h3>
                                <div className="space-y-3">
                                    {directInstances.map((instance, index) => (
                                        <div
                                            key={index}
                                            className="bg-gray-50 p-3 rounded"
                                        >
                                            <div className="flex items-center gap-2 mb-1">
                                                <Circle className="w-3 h-3 text-green-500 fill-current" />
                                                <span className="font-medium text-sm">
                                                    {instance.ip}
                                                </span>
                                            </div>
                                            {instance.pod && (
                                                <div className="text-sm text-gray-600 ml-5">
                                                    Pod: {instance.pod}
                                                </div>
                                            )}
                                            <div className="text-sm text-gray-600 ml-5">
                                                Namespace: {instance.namespace}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        {proxiedInstances.length > 0 && (
                            <div>
                                <h3 className="text-md font-medium text-gray-900 mb-2 flex items-center gap-2">
                                    <Shield className="w-4 h-4 text-orange-500" />
                                    Instances with Proxy Sidecar
                                </h3>
                                <div className="space-y-3">
                                    {proxiedInstances.map((instance, index) => (
                                        <div
                                            key={index}
                                            className="bg-orange-50 p-3 rounded border border-orange-200"
                                        >
                                            <div className="flex items-center gap-2 mb-1">
                                                <Circle className="w-3 h-3 text-green-500 fill-current" />
                                                <span className="font-medium text-sm">
                                                    {instance.ip}
                                                </span>
                                                <span className="text-xs px-2 py-1 bg-orange-100 text-orange-800 rounded">
                                                    Proxied
                                                </span>
                                            </div>
                                            {instance.pod && (
                                                <div className="text-sm text-gray-600 ml-5">
                                                    Pod: {instance.pod}
                                                </div>
                                            )}
                                            <div className="text-sm text-gray-600 ml-5">
                                                Namespace: {instance.namespace}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        {service.instances.length === 0 && (
                            <p className="text-gray-500 text-sm">
                                No instances available
                            </p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};
