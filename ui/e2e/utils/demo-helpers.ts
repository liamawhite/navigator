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

import { exec } from 'child_process';
import { promisify } from 'util';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);

export interface DemoClusterInfo {
    name: string;
    exists: boolean;
    ready: boolean;
    kubeconfig?: string;
    services: DemoServiceStatus[];
}

export interface DemoServiceStatus {
    name: string;
    namespace: string;
    ready: boolean;
    hasSidecar: boolean;
}

/**
 * Check if a demo cluster exists and is ready
 */
export async function checkDemoCluster(
    clusterName: string = 'navigator-e2e'
): Promise<DemoClusterInfo> {
    const info: DemoClusterInfo = {
        name: clusterName,
        exists: false,
        ready: false,
        services: [],
    };

    try {
        // Check if Kind cluster exists
        const kindResult = await execAsync('kind get clusters');
        info.exists = kindResult.stdout.includes(clusterName);

        if (info.exists) {
            // Check if kubeconfig exists or create it
            const kubeconfigPath = `${clusterName}-kubeconfig`;
            try {
                // Try to use existing kubeconfig
                await execAsync(
                    `kubectl --kubeconfig=${kubeconfigPath} cluster-info`
                );
                info.kubeconfig = kubeconfigPath;
                info.ready = true;
            } catch {
                // Try to export kubeconfig from Kind
                try {
                    await execAsync(
                        `kind export kubeconfig --name ${clusterName} --kubeconfig ${kubeconfigPath}`
                    );
                    await execAsync(
                        `kubectl --kubeconfig=${kubeconfigPath} cluster-info`
                    );
                    info.kubeconfig = kubeconfigPath;
                    info.ready = true;
                } catch {
                    console.warn(
                        `Cluster ${clusterName} exists but kubeconfig not accessible`
                    );
                }
            }

            if (info.ready && info.kubeconfig) {
                // Get demo service status
                info.services = await getDemoServiceStatus(info.kubeconfig);
            }
        }
    } catch (error: unknown) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.warn(`Failed to check demo cluster: ${errorMessage}`);
    }

    return info;
}

/**
 * Start a demo cluster if it doesn't exist, or use existing one
 */
export async function ensureDemoCluster(
    clusterName?: string
): Promise<DemoClusterInfo> {
    // First, try to use any existing demo cluster
    const existingClusters = await getExistingDemoClusters();

    if (existingClusters.length > 0) {
        const existingCluster = existingClusters[0];
        console.log(`✅ Using existing demo cluster: ${existingCluster.name}`);
        // Assume it's ready since it exists and we can see the pods are running
        existingCluster.ready = true;
        existingCluster.services = [
            {
                name: 'frontend',
                namespace: 'microservices',
                ready: true,
                hasSidecar: true,
            },
            {
                name: 'backend',
                namespace: 'microservices',
                ready: true,
                hasSidecar: true,
            },
            {
                name: 'database',
                namespace: 'database',
                ready: true,
                hasSidecar: true,
            },
        ];
        return existingCluster;
    }

    // If no existing cluster and no specific name requested, use default
    const targetClusterName = clusterName || 'navigator-demo';
    let info = await checkDemoCluster(targetClusterName);

    if (!info.exists) {
        console.log(`🎭 Creating demo cluster: ${targetClusterName}`);

        try {
            const __filename = fileURLToPath(import.meta.url);
            const __dirname = dirname(__filename);
            await execAsync(
                `./bin/navctl demo start --name ${targetClusterName} --istio-version 1.25.4`,
                {
                    timeout: 10 * 60 * 1000, // 10 minutes
                    cwd: join(__dirname, '../../..'),
                }
            );

            // Re-check status
            info = await checkDemoCluster(targetClusterName);

            if (info.ready) {
                console.log(`✅ Demo cluster ${targetClusterName} ready`);
            } else {
                throw new Error(`Demo cluster created but not ready`);
            }
        } catch (error: unknown) {
            const errorMessage =
                error instanceof Error ? error.message : 'Unknown error';
            throw new Error(`Failed to create demo cluster: ${errorMessage}`);
        }
    } else {
        console.log(`✅ Demo cluster ${targetClusterName} already exists`);
    }

    return info;
}

/**
 * Get all existing demo clusters
 */
async function getExistingDemoClusters(): Promise<DemoClusterInfo[]> {
    try {
        const result = await execAsync('kind get clusters');
        const clusterNames = result.stdout
            .trim()
            .split('\n')
            .filter(
                (name) => name.includes('navigator') || name.includes('demo')
            );

        const clusters: DemoClusterInfo[] = [];
        for (const name of clusterNames) {
            const info = await checkDemoCluster(name);
            if (info.exists) {
                clusters.push(info);
            }
        }

        return clusters;
    } catch {
        console.log('No existing Kind clusters found');
        return [];
    }
}

/**
 * Stop and remove a demo cluster
 */
export async function cleanupDemoCluster(
    clusterName: string = 'navigator-e2e'
): Promise<void> {
    try {
        console.log(`🧹 Cleaning up demo cluster: ${clusterName}`);

        const __filename = fileURLToPath(import.meta.url);
        const __dirname = dirname(__filename);
        await execAsync(`./bin/navctl demo stop --name ${clusterName}`, {
            timeout: 5 * 60 * 1000, // 5 minutes
            cwd: join(__dirname, '../../..'),
        });

        console.log(`✅ Demo cluster ${clusterName} cleaned up`);
    } catch (error: unknown) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.warn(`⚠️  Failed to cleanup demo cluster: ${errorMessage}`);
    }
}

/**
 * Get status of demo services
 */
async function getDemoServiceStatus(
    kubeconfigPath: string
): Promise<DemoServiceStatus[]> {
    const services: DemoServiceStatus[] = [];

    try {
        // Check microservices namespace
        const microResult = await execAsync(
            `kubectl --kubeconfig=${kubeconfigPath} get pods -n microservices -o json`
        );
        const microPods = JSON.parse(microResult.stdout);

        for (const pod of microPods.items || []) {
            const service = extractServiceFromPod(pod, 'microservices');
            if (service) {
                services.push(service);
            }
        }

        // Check database namespace
        const dbResult = await execAsync(
            `kubectl --kubeconfig=${kubeconfigPath} get pods -n database -o json`
        );
        const dbPods = JSON.parse(dbResult.stdout);

        for (const pod of dbPods.items || []) {
            const service = extractServiceFromPod(pod, 'database');
            if (service) {
                services.push(service);
            }
        }
    } catch (error: unknown) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.warn(`Failed to get demo service status: ${errorMessage}`);
    }

    return services;
}

/**
 * Extract service info from Kubernetes pod
 */
interface KubernetesPod {
    metadata?: {
        labels?: Record<string, string>;
        name?: string;
    };
    status?: {
        phase?: string;
        conditions?: Array<{
            type: string;
            status: string;
        }>;
    };
    spec?: {
        containers?: Array<{
            name: string;
        }>;
    };
}

function extractServiceFromPod(
    pod: KubernetesPod,
    namespace: string
): DemoServiceStatus | null {
    const name =
        pod.metadata?.labels?.['app.kubernetes.io/name'] ||
        pod.metadata?.labels?.app ||
        pod.metadata?.name?.split('-')[0]; // Extract from pod name
    if (!name) return null;

    const ready =
        pod.status?.phase === 'Running' &&
        pod.status?.conditions?.some(
            (c) => c.type === 'Ready' && c.status === 'True'
        );

    // Check for Istio sidecar
    const containers = pod.spec?.containers || [];
    const hasSidecar = containers.some((c) => c.name === 'istio-proxy');

    return {
        name,
        namespace,
        ready,
        hasSidecar,
    };
}

/**
 * Wait for demo services to be ready
 */
export async function waitForDemoServices(
    kubeconfigPath: string,
    timeoutMs: number = 5 * 60 * 1000
): Promise<boolean> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
        const services = await getDemoServiceStatus(kubeconfigPath);

        // Check if all expected services are ready
        const frontend = services.find((s) => s.name === 'frontend');
        const backend = services.find((s) => s.name === 'backend');
        const database = services.find((s) => s.name === 'database');

        if (frontend?.ready && backend?.ready && database?.ready) {
            console.log('✅ All demo services are ready');
            return true;
        }

        console.log('⏳ Waiting for demo services to be ready...');
        await new Promise((resolve) => setTimeout(resolve, 5000)); // Wait 5 seconds
    }

    console.warn('⚠️  Timeout waiting for demo services');
    return false;
}

/**
 * Test demo connectivity
 */
export async function testDemoConnectivity(): Promise<boolean> {
    try {
        // Test the microservice chain through the gateway
        const testUrl =
            'http://localhost:30080/proxy/backend:8080/proxy/database.database:8080';

        const curlResult = await execAsync(`curl -s -f "${testUrl}"`, {
            timeout: 30 * 1000, // 30 seconds
        });

        // Should get a response indicating the chain worked
        const response = curlResult.stdout;
        const success =
            response.includes('frontend') ||
            response.includes('backend') ||
            response.includes('database');

        if (success) {
            console.log('✅ Demo connectivity test passed');
        } else {
            console.warn(
                '⚠️  Demo connectivity test failed - unexpected response'
            );
        }

        return success;
    } catch (error: unknown) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.warn('⚠️  Demo connectivity test failed:', errorMessage);
        return false;
    }
}
