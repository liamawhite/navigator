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
