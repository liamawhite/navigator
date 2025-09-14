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
import { useMetricsContext } from '../contexts/MetricsContext';
import { MetricsServiceService } from '../types/generated/openapi-metrics_service';

export function useServiceConnections(serviceName: string, namespace: string) {
    const { startTime, endTime, refreshTrigger } = useMetricsContext();

    return useQuery({
        queryKey: [
            'serviceConnections',
            serviceName,
            namespace,
            refreshTrigger,
        ],
        queryFn: async () => {
            const response =
                await MetricsServiceService.metricsServiceGetServiceConnections(
                    serviceName,
                    namespace,
                    startTime.toISOString(),
                    endTime.toISOString()
                );

            // Handle the union type - check if it's an error response
            if ('code' in response) {
                throw new Error(
                    `Failed to fetch service connections: ${response.message}`
                );
            }

            return response;
        },
        enabled: refreshTrigger > 0, // Only fetch when manually triggered (starts at 1)
        retry: 3,
        retryDelay: 1000,
    });
}
