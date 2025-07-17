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
