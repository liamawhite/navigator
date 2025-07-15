import { useQuery } from '@tanstack/react-query';
import { serviceApi } from '../utils/api';

export const useServices = () => {
    return useQuery({
        queryKey: ['services'],
        queryFn: serviceApi.listServices,
        refetchInterval: 5000,
    });
};

export const useService = (id: string) => {
    return useQuery({
        queryKey: ['service', id],
        queryFn: () => serviceApi.getService(id),
        enabled: !!id,
    });
};

export const useServiceInstance = (serviceId: string, instanceId: string) => {
    return useQuery({
        queryKey: ['serviceInstance', serviceId, instanceId],
        queryFn: () => serviceApi.getServiceInstance(serviceId, instanceId),
        enabled: !!serviceId && !!instanceId,
    });
};

export const useProxyConfig = (serviceId: string, instanceId: string) => {
    return useQuery({
        queryKey: ['proxyConfig', serviceId, instanceId],
        queryFn: () => serviceApi.getProxyConfig(serviceId, instanceId),
        enabled: !!serviceId && !!instanceId,
    });
};
