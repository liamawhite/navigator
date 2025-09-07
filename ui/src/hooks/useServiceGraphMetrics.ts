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
import { metricsApi } from '../utils/api';
import type { v1alpha1ServicePairMetrics } from '../types/generated/openapi-metrics_service';

interface ServiceGraphMetricsParams {
    namespaces?: string[];
    clusters?: string[];
    startTime?: string;
    endTime?: string;
}

export const useServiceGraphMetrics = (
    params?: ServiceGraphMetricsParams,
    refetchInterval: number | false = 30000
) => {
    return useQuery<v1alpha1ServicePairMetrics[], Error>({
        queryKey: ['serviceGraphMetrics', params],
        queryFn: () => metricsApi.getServiceGraphMetrics(params),
        refetchInterval,
        staleTime:
            refetchInterval === false
                ? 0
                : Math.max(1000, refetchInterval - 5000), // Consider data stale slightly before refetch
    });
};
