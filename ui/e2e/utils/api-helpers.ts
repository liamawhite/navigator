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

import { APIRequestContext, APIResponse, expect } from '@playwright/test';

export interface NavigatorApiEndpoints {
    services: string;
    service: (serviceName: string) => string;
    serviceInstance: (serviceName: string, instanceId: string) => string;
    clusters: string;
    metrics: string;
    serviceConnections: string;
}

/**
 * Navigator API endpoints
 */
export const API_ENDPOINTS: NavigatorApiEndpoints = {
    services: '/api/v1alpha1/services',
    service: (serviceName: string) => `/api/v1alpha1/services/${serviceName}`,
    serviceInstance: (serviceName: string, instanceId: string) =>
        `/api/v1alpha1/services/${serviceName}/instances/${instanceId}`,
    clusters: '/api/v1alpha1/clusters',
    metrics: '/api/v1alpha1/metrics',
    serviceConnections: '/api/v1alpha1/metrics/service-connections',
};

/**
 * Make API request and validate response
 */
export async function makeApiRequest(
    apiContext: APIRequestContext,
    endpoint: string,
    options: {
        method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
        data?: unknown;
        headers?: Record<string, string>;
        expectedStatus?: number;
    } = {}
): Promise<APIResponse> {
    const {
        method = 'GET',
        data,
        headers = {},
        expectedStatus = 200,
    } = options;

    // Use the API base URL (port 8081) instead of the UI base URL (port 8082)
    const apiBaseUrl = 'http://localhost:8081';
    const fullUrl = endpoint.startsWith('http')
        ? endpoint
        : `${apiBaseUrl}${endpoint}`;

    const response = await apiContext[
        method.toLowerCase() as 'get' | 'post' | 'put' | 'delete'
    ](fullUrl, {
        data,
        headers: {
            'Content-Type': 'application/json',
            ...headers,
        },
    });

    expect(response.status()).toBe(expectedStatus);
    return response;
}

/**
 * Get services list from API
 */
export interface Service {
    name: string;
    namespace: string;
    clusterName: string;
    instances: ServiceInstance[];
}

export async function getServices(
    apiContext: APIRequestContext
): Promise<Service[]> {
    const response = await makeApiRequest(apiContext, API_ENDPOINTS.services);
    const data = await response.json();
    return data.services || [];
}

/**
 * Get specific service details from API
 */
export async function getService(
    apiContext: APIRequestContext,
    serviceName: string
): Promise<Service> {
    const response = await makeApiRequest(
        apiContext,
        API_ENDPOINTS.service(serviceName)
    );
    return await response.json();
}

/**
 * Get service instance details from API
 */
export interface ServiceInstance {
    name: string;
    namespace: string;
    podName: string;
    proxyConfig?: unknown;
    hasSidecar?: boolean;
}

export async function getServiceInstance(
    apiContext: APIRequestContext,
    serviceName: string,
    instanceId: string
): Promise<ServiceInstance> {
    const response = await makeApiRequest(
        apiContext,
        API_ENDPOINTS.serviceInstance(serviceName, instanceId)
    );
    return await response.json();
}

/**
 * Get clusters from API
 */
export interface Cluster {
    name: string;
    endpoint: string;
}

export async function getClusters(
    apiContext: APIRequestContext
): Promise<Cluster[]> {
    const response = await makeApiRequest(apiContext, API_ENDPOINTS.clusters);
    const data = await response.json();
    return data.clusters || [];
}

/**
 * Get service connections from API
 */
export interface ServiceConnection {
    source: string;
    target: string;
    metrics?: unknown;
}

export async function getServiceConnections(
    apiContext: APIRequestContext
): Promise<ServiceConnection[]> {
    const response = await makeApiRequest(
        apiContext,
        API_ENDPOINTS.serviceConnections
    );
    const data = await response.json();
    return data.servicePairs || [];
}

/**
 * Wait for services to be discovered by the API
 */
export async function waitForServicesDiscovered(
    apiContext: APIRequestContext,
    expectedServices: string[],
    timeoutMs: number = 60000
): Promise<boolean> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
        try {
            const services = await getServices(apiContext);
            const serviceNames = services.map((s) => s.name);

            const allFound = expectedServices.every((name) =>
                serviceNames.includes(name)
            );

            if (allFound) {
                console.log('✅ All expected services discovered by API');
                return true;
            }

            console.log(
                `⏳ Waiting for services... Found: ${serviceNames.join(', ')}`
            );
            await new Promise((resolve) => setTimeout(resolve, 5000)); // Wait 5 seconds
        } catch {
            console.log('⏳ API not ready yet, waiting...');
            await new Promise((resolve) => setTimeout(resolve, 5000));
        }
    }

    console.warn('⚠️  Timeout waiting for services to be discovered');
    return false;
}

/**
 * Validate service has expected properties
 */
export function validateServiceStructure(service: Service): void {
    expect(service).toBeDefined();
    expect(service.name).toBeDefined();
    expect(service.namespace).toBeDefined();
    expect(service.clusterName).toBeDefined();
    expect(Array.isArray(service.instances)).toBe(true);
}

/**
 * Validate service instance has expected properties
 */
export function validateServiceInstanceStructure(
    instance: ServiceInstance
): void {
    expect(instance).toBeDefined();
    expect(instance.name).toBeDefined();
    expect(instance.namespace).toBeDefined();
    expect(instance.podName).toBeDefined();
}

/**
 * Check if service has Istio sidecar via API
 */
export async function serviceHasSidecar(
    apiContext: APIRequestContext,
    serviceName: string
): Promise<boolean> {
    try {
        const service = await getService(apiContext, serviceName);

        // Check if any instance has proxy configuration
        return (
            service.instances?.some(
                (instance: ServiceInstance) =>
                    instance.proxyConfig || instance.hasSidecar
            ) || false
        );
    } catch {
        console.warn(`Failed to check sidecar for service ${serviceName}:`);
        return false;
    }
}

/**
 * Get proxy configuration for a service instance
 */
export async function getProxyConfig(
    apiContext: APIRequestContext,
    serviceName: string,
    instanceId: string
): Promise<unknown> {
    const instance = await getServiceInstance(
        apiContext,
        serviceName,
        instanceId
    );
    return instance.proxyConfig;
}

/**
 * Validate API response time
 */
export async function validateResponseTime(
    apiContext: APIRequestContext,
    endpoint: string,
    maxResponseTimeMs: number = 5000
): Promise<number> {
    const startTime = Date.now();
    await makeApiRequest(apiContext, endpoint);
    const responseTime = Date.now() - startTime;

    expect(responseTime).toBeLessThan(maxResponseTimeMs);
    return responseTime;
}

/**
 * Check API health
 */
export async function checkApiHealth(
    apiContext: APIRequestContext
): Promise<boolean> {
    try {
        await makeApiRequest(apiContext, '/api/v1alpha1/services');
        return true;
    } catch (error) {
        console.warn('API health check failed:', error);
        return false;
    }
}
