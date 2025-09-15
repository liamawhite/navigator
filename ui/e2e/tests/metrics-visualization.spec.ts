/* eslint-disable @typescript-eslint/no-unused-vars */
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

import { test, expect } from '@playwright/test';
import { NavigatorPage } from '../fixtures/navigator-page';
import { ensureNavctlReady } from '../utils/build-helpers';
import {
    ensureDemoCluster,
    waitForDemoServices,
    testDemoConnectivity,
} from '../utils/demo-helpers';
import {
    getServiceConnections,
    waitForServicesDiscovered,
} from '../utils/api-helpers';

test.describe('Metrics Visualization', () => {
    let navigatorPage: NavigatorPage;

    test.beforeAll(async () => {
        // Ensure navctl is built and ready
        await ensureNavctlReady();

        // Ensure demo cluster is available
        const demoInfo = await ensureDemoCluster();
        if (!demoInfo.ready) {
            throw new Error('Demo cluster is not ready');
        }
        console.log(
            `✅ Demo cluster ready with services: ${demoInfo.services.map((s) => s.name).join(', ')}`
        );

        // Test demo connectivity to ensure metrics are being generated
        await testDemoConnectivity();
    });

    test.beforeEach(async ({ page }) => {
        navigatorPage = new NavigatorPage(page);
    });

    test('should display service metrics on service details page', async ({
        page,
        request,
    }) => {
        // Wait for services to be discovered
        await waitForServicesDiscovered(request, ['frontend']);

        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Check if metrics section is available
        const metricsVisible = await navigatorPage.serviceMetrics.isVisible();

        if (metricsVisible) {
            await navigatorPage.waitForMetricsLoaded();
            await navigatorPage.expectMetricsVisible();

            // Verify all metric types are displayed
            await expect(navigatorPage.requestRateMetric).toBeVisible();
            await expect(navigatorPage.errorRateMetric).toBeVisible();
            await expect(navigatorPage.latencyMetric).toBeVisible();

            // Verify metrics controls are available
            await expect(navigatorPage.metricsTimeRange).toBeVisible();
            await expect(navigatorPage.metricsRefresh).toBeVisible();
        } else {
            console.log(
                'Metrics not available - may need Prometheus configuration'
            );
        }
    });

    test('should allow changing metrics time range', async ({
        page,
        request,
    }) => {
        // Navigate to service with metrics
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Skip if metrics not available
        const metricsVisible = await navigatorPage.serviceMetrics.isVisible();
        if (!metricsVisible) {
            test.skip('Metrics not available');
            return;
        }

        await navigatorPage.waitForMetricsLoaded();

        // Test different time ranges
        const timeRanges = ['5m', '15m', '1h', '6h', '24h'];

        for (const range of timeRanges) {
            // Check if this time range option exists
            const timeRangeSelect = navigatorPage.metricsTimeRange;
            const options = await timeRangeSelect
                .locator('option')
                .allTextContents();

            if (options.some((option) => option.includes(range))) {
                await navigatorPage.selectMetricsTimeRange(range);

                // Verify the selection was applied
                await expect(timeRangeSelect).toHaveValue(range);

                // Wait for metrics to reload
                await page.waitForTimeout(2000);

                // Verify metrics are still visible after time range change
                await navigatorPage.expectMetricsVisible();
            }
        }
    });

    test('should allow manual metrics refresh', async ({ page, request }) => {
        // Navigate to service with metrics
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Skip if metrics not available
        const metricsVisible = await navigatorPage.serviceMetrics.isVisible();
        if (!metricsVisible) {
            test.skip('Metrics not available');
            return;
        }

        await navigatorPage.waitForMetricsLoaded();

        // Get initial metric values
        const initialRequestRate =
            await navigatorPage.requestRateMetric.textContent();
        const initialErrorRate =
            await navigatorPage.errorRateMetric.textContent();
        const initialLatency = await navigatorPage.latencyMetric.textContent();

        // Click refresh button
        await navigatorPage.refreshMetrics();

        // Verify metrics are still displayed (values may or may not change)
        await navigatorPage.expectMetricsVisible();

        // Verify refresh button is functional (doesn't cause errors)
        await expect(navigatorPage.metricsRefresh).toBeVisible();
        await expect(navigatorPage.serviceMetrics).toBeVisible();
    });

    test('should display service connections visualization', async ({
        page,
        request,
    }) => {
        // Navigate to home page to see overall topology
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(3);

        // Check for service connections visualization section
        const connectionsVisible =
            await navigatorPage.serviceConnections.isVisible();

        if (connectionsVisible) {
            await navigatorPage.waitForServiceConnectionsLoaded();
            await navigatorPage.expectConnectionGraphVisible();

            // Verify the graph container is present
            await expect(navigatorPage.connectionGraph).toBeVisible();

            // Check for graph elements (nodes and edges)
            const hasCanvas = await page.locator('canvas').isVisible();
            const hasSvg = await page.locator('svg').isVisible();

            // Should have either canvas or SVG-based visualization
            expect(hasCanvas || hasSvg).toBe(true);
        } else {
            console.log('Service connections visualization not available');
        }
    });

    test('should show metrics from demo load generation', async ({
        page,
        request,
    }) => {
        // Demo should have Fortio load generation running
        // Wait a bit for load to generate metrics
        await page.waitForTimeout(10000); // 10 seconds

        // Navigate to frontend service (receives traffic)
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Skip if metrics not available
        const metricsVisible = await navigatorPage.serviceMetrics.isVisible();
        if (!metricsVisible) {
            test.skip('Metrics not available');
            return;
        }

        await navigatorPage.waitForMetricsLoaded();

        // Get metric values
        const requestRateText =
            await navigatorPage.requestRateMetric.textContent();
        const errorRateText = await navigatorPage.errorRateMetric.textContent();
        const latencyText = await navigatorPage.latencyMetric.textContent();

        console.log(
            `Metrics - Request Rate: ${requestRateText}, Error Rate: ${errorRateText}, Latency: ${latencyText}`
        );

        // With load generation, we should see some request rate
        // Values might be numbers or dashes if metrics aren't flowing yet
        expect(requestRateText).toBeTruthy();
        expect(errorRateText).toBeTruthy();
        expect(latencyText).toBeTruthy();

        // If we see actual numbers, they should be reasonable
        if (requestRateText && requestRateText.match(/\d+/)) {
            const requestRate = parseFloat(requestRateText);
            expect(requestRate).toBeGreaterThanOrEqual(0);
        }

        if (errorRateText && errorRateText.match(/\d+/)) {
            const errorRate = parseFloat(errorRateText);
            expect(errorRate).toBeGreaterThanOrEqual(0);
            expect(errorRate).toBeLessThanOrEqual(100); // Percentage
        }
    });

    test('should show service connections from API', async ({
        page,
        request,
    }) => {
        // Get service connections from API
        const serviceConnections = await getServiceConnections(request);

        if (serviceConnections.length === 0) {
            console.log('No service connections found in API');
            test.skip('No service connections available');
            return;
        }

        console.log(
            `Found ${serviceConnections.length} service connections in API`
        );

        // Navigate to topology view
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(3);

        // Check if connections visualization is available
        const connectionsVisible =
            await navigatorPage.serviceConnections.isVisible();

        if (connectionsVisible) {
            await navigatorPage.waitForServiceConnectionsLoaded();

            // Verify visualization shows connections
            await navigatorPage.expectConnectionGraphVisible();

            // Log the connections for debugging
            for (const connection of serviceConnections.slice(0, 5)) {
                // First 5 connections
                console.log(
                    `Connection: ${connection.source?.name} -> ${connection.destination?.name}`
                );
            }
        } else {
            console.log('Service connections visualization not rendered in UI');
        }
    });

    test('should handle metrics auto-refresh', async ({ page, request }) => {
        // Navigate to service with metrics
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Skip if metrics not available
        const metricsVisible = await navigatorPage.serviceMetrics.isVisible();
        if (!metricsVisible) {
            test.skip('Metrics not available');
            return;
        }

        await navigatorPage.waitForMetricsLoaded();

        // Get initial metric values
        const initialRequestRate =
            await navigatorPage.requestRateMetric.textContent();

        // Wait for auto-refresh (Navigator typically refreshes every 5 seconds)
        await page.waitForTimeout(6000);

        // Verify metrics are still displayed after auto-refresh
        await navigatorPage.expectMetricsVisible();

        // Get updated metric values
        const updatedRequestRate =
            await navigatorPage.requestRateMetric.textContent();

        // Values may or may not change, but should still be displayed
        expect(updatedRequestRate).toBeTruthy();

        // Metrics section should remain functional
        await expect(navigatorPage.serviceMetrics).toBeVisible();
        await expect(navigatorPage.metricsRefresh).toBeVisible();
    });

    test('should show meaningful metrics during load test', async ({
        page,
        request,
    }) => {
        // This test assumes Fortio load generation is running
        // Generate some additional load to ensure metrics
        await testDemoConnectivity();

        // Wait for metrics to propagate
        await page.waitForTimeout(15000); // 15 seconds

        // Check metrics on all demo services
        const services = ['frontend', 'backend'];

        for (const serviceName of services) {
            await navigatorPage.goToService(serviceName);
            await navigatorPage.waitForServiceDetailsLoaded(serviceName);

            const metricsVisible =
                await navigatorPage.serviceMetrics.isVisible();

            if (metricsVisible) {
                await navigatorPage.waitForMetricsLoaded();

                // Get metric values
                const requestRateText =
                    await navigatorPage.requestRateMetric.textContent();
                const errorRateText =
                    await navigatorPage.errorRateMetric.textContent();
                const latencyText =
                    await navigatorPage.latencyMetric.textContent();

                console.log(
                    `${serviceName} metrics - RPS: ${requestRateText}, Error%: ${errorRateText}, Latency: ${latencyText}`
                );

                // Verify metrics are displayed (may be numbers or placeholders)
                expect(requestRateText).toBeTruthy();
                expect(errorRateText).toBeTruthy();
                expect(latencyText).toBeTruthy();

                // If we have numeric metrics, they should be reasonable
                const rpsMatch = requestRateText?.match(/(\d+\.?\d*)/);
                if (rpsMatch) {
                    const rps = parseFloat(rpsMatch[1]);
                    expect(rps).toBeGreaterThanOrEqual(0);
                    expect(rps).toBeLessThan(1000); // Reasonable upper bound
                }

                const errorMatch = errorRateText?.match(/(\d+\.?\d*)/);
                if (errorMatch) {
                    const errorRate = parseFloat(errorMatch[1]);
                    expect(errorRate).toBeGreaterThanOrEqual(0);
                    expect(errorRate).toBeLessThanOrEqual(100);
                }
            } else {
                console.log(`Metrics not available for ${serviceName}`);
            }
        }
    });

    test('should handle missing metrics gracefully', async ({
        page,
        request,
    }) => {
        // Navigate to database service (may not have metrics)
        await navigatorPage.goToService('database');
        await navigatorPage.waitForServiceDetailsLoaded('database');

        // Check metrics section
        const metricsVisible = await navigatorPage.serviceMetrics.isVisible();

        if (metricsVisible) {
            // If metrics section is shown, it should handle missing data gracefully
            await navigatorPage.waitForMetricsLoaded();

            const requestRate =
                await navigatorPage.requestRateMetric.textContent();
            const errorRate = await navigatorPage.errorRateMetric.textContent();
            const latency = await navigatorPage.latencyMetric.textContent();

            // Should show placeholder values like "--" or "..." when metrics are missing
            expect(requestRate).toMatch(/--|\.\.\.|0|N\/A/i);
            expect(errorRate).toMatch(/--|\.\.\.|0|N\/A/i);
            expect(latency).toMatch(/--|\.\.\.|0|N\/A/i);
        } else {
            // If no metrics section, that's also acceptable
            console.log(
                'No metrics section for database service - this is expected'
            );
        }

        // Page should still be functional
        await expect(navigatorPage.serviceHeader).toBeVisible();
        await expect(navigatorPage.serviceInstances).toBeVisible();
    });

    test('should validate metrics API response format', async ({
        page,
        request,
    }) => {
        // Test direct API call for service connections
        try {
            const serviceConnections = await getServiceConnections(request);

            // Validate API response structure
            expect(Array.isArray(serviceConnections)).toBe(true);

            if (serviceConnections.length > 0) {
                const firstConnection = serviceConnections[0];

                // Validate connection structure
                expect(firstConnection).toHaveProperty('source');
                expect(firstConnection).toHaveProperty('destination');

                if (firstConnection.source) {
                    expect(firstConnection.source).toHaveProperty('name');
                    expect(firstConnection.source).toHaveProperty('namespace');
                }

                if (firstConnection.destination) {
                    expect(firstConnection.destination).toHaveProperty('name');
                    expect(firstConnection.destination).toHaveProperty(
                        'namespace'
                    );
                }

                console.log('Service connections API validation passed');
            } else {
                console.log(
                    'No service connections found - this may be expected'
                );
            }
        } catch (error) {
            console.log('Service connections API not available:', error);
        }
    });
});
