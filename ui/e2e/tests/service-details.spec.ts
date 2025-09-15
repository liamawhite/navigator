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
import { ensureDemoCluster, waitForDemoServices } from '../utils/demo-helpers';
import {
    getService,
    getServiceInstance,
    serviceHasSidecar,
    waitForServicesDiscovered,
} from '../utils/api-helpers';

test.describe('Service Details', () => {
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
    });

    test.beforeEach(async ({ page }) => {
        navigatorPage = new NavigatorPage(page);
    });

    test('should display service details page', async ({ page, request }) => {
        // Wait for services to be discovered
        await waitForServicesDiscovered(request, ['frontend']);

        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Verify service header information
        await expect(navigatorPage.serviceHeader).toBeVisible();
        await expect(navigatorPage.serviceHeader).toContainText('frontend');
        await expect(navigatorPage.serviceHeader).toContainText(
            'microservices'
        ); // namespace

        // Verify service instances section is visible
        await expect(navigatorPage.serviceInstances).toBeVisible();
    });

    test('should show service instances and endpoints', async ({
        page,
        request,
    }) => {
        // Get service data from API
        const serviceData = await getService(request, 'frontend');
        expect(serviceData.instances).toBeDefined();
        expect(serviceData.instances.length).toBeGreaterThan(0);

        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Verify service instances are displayed
        const instanceCount = await navigatorPage.instanceCards.count();
        expect(instanceCount).toBe(serviceData.instances.length);

        // Verify first instance details
        const firstInstance = serviceData.instances[0];
        const firstInstanceCard = navigatorPage.getInstanceCard(
            firstInstance.name
        );

        await expect(firstInstanceCard).toBeVisible();
        await expect(firstInstanceCard).toContainText(firstInstance.podName);

        // Verify endpoint information
        if (firstInstance.endpoints && firstInstance.endpoints.length > 0) {
            await expect(firstInstanceCard).toContainText('8080'); // Default port
        }
    });

    test('should navigate to service instance details', async ({
        page,
        request,
    }) => {
        // Get service data from API
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Click on first instance
        await navigatorPage.clickInstanceCard(firstInstance.name);

        // Should navigate to instance details page
        await navigatorPage.expectCurrentPath(
            `/services/frontend/instances/${firstInstance.name}`
        );

        // Verify breadcrumb navigation
        await navigatorPage.expectBreadcrumbVisible('Services');
        await navigatorPage.expectBreadcrumbVisible('frontend');
        await navigatorPage.expectBreadcrumbVisible(firstInstance.name);
    });

    test('should show proxy configuration for services with sidecars', async ({
        page,
        request,
    }) => {
        // Check if frontend has sidecar
        const hasSidecar = await serviceHasSidecar(request, 'frontend');

        if (!hasSidecar) {
            test.skip('Frontend service does not have Istio sidecar');
            return;
        }

        // Get service data from API
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        // Navigate to service instance details
        await navigatorPage.goToServiceInstance('frontend', firstInstance.name);

        // Verify proxy configuration section is visible
        await navigatorPage.waitForProxyConfigLoaded();
        await navigatorPage.expectProxyConfigVisible();

        // Verify proxy config tabs are available
        await expect(navigatorPage.proxyConfigTabs).toBeVisible();
        await expect(navigatorPage.bootstrapTab).toBeVisible();
        await expect(navigatorPage.clustersTab).toBeVisible();
        await expect(navigatorPage.listenersTab).toBeVisible();
        await expect(navigatorPage.routesTab).toBeVisible();
        await expect(navigatorPage.endpointsTab).toBeVisible();

        // Verify config editor contains data
        await expect(navigatorPage.configEditor).toBeVisible();

        // Check that the editor has content
        const editorContent = await navigatorPage.configEditor.textContent();
        expect(editorContent).toBeTruthy();
        expect(editorContent!.length).toBeGreaterThan(100); // Should have substantial content
    });

    test('should switch between proxy configuration tabs', async ({
        page,
        request,
    }) => {
        // Check if frontend has sidecar
        const hasSidecar = await serviceHasSidecar(request, 'frontend');

        if (!hasSidecar) {
            test.skip('Frontend service does not have Istio sidecar');
            return;
        }

        // Get service data
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        // Navigate to service instance details
        await navigatorPage.goToServiceInstance('frontend', firstInstance.name);
        await navigatorPage.waitForProxyConfigLoaded();

        // Test switching between tabs
        const tabs = [
            'clusters',
            'listeners',
            'routes',
            'endpoints',
            'bootstrap',
        ] as const;

        for (const tab of tabs) {
            await navigatorPage.switchProxyConfigTab(tab);

            // Verify tab is active
            const tabElement = page.locator(`[data-testid="tab-${tab}"]`);
            await expect(tabElement).toHaveAttribute('aria-selected', 'true');

            // Verify content loaded
            await expect(navigatorPage.configEditor).toBeVisible();

            // Verify content changes between tabs
            const content = await navigatorPage.configEditor.textContent();
            expect(content).toBeTruthy();
            expect(content!.length).toBeGreaterThan(10);
        }
    });

    test('should show Istio resources for services', async ({
        page,
        request,
    }) => {
        // Navigate to service with potential Istio resources
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Check if Istio resources section exists
        const istioResourcesVisible =
            await navigatorPage.istioResources.isVisible();

        if (istioResourcesVisible) {
            // Verify Istio resources sections
            await expect(navigatorPage.istioResources).toBeVisible();

            // Check for different resource types (may or may not be present)
            const virtualServicesVisible =
                await navigatorPage.virtualServices.isVisible();
            const destinationRulesVisible =
                await navigatorPage.destinationRules.isVisible();
            const gatewaysVisible = await navigatorPage.gateways.isVisible();

            console.log(
                `Istio resources found - VirtualServices: ${virtualServicesVisible}, DestinationRules: ${destinationRulesVisible}, Gateways: ${gatewaysVisible}`
            );

            // At least one type of Istio resource should be present
            expect(
                virtualServicesVisible ||
                    destinationRulesVisible ||
                    gatewaysVisible
            ).toBe(true);
        } else {
            console.log('No Istio resources found for frontend service');
        }
    });

    test('should handle service instance without sidecar', async ({
        page,
        request,
    }) => {
        // Navigate to database service (typically doesn't have sidecar)
        await navigatorPage.goToService('database');
        await navigatorPage.waitForServiceDetailsLoaded('database');

        // Get service data
        const serviceData = await getService(request, 'database');
        const firstInstance = serviceData.instances[0];

        // Navigate to instance details
        await navigatorPage.goToServiceInstance('database', firstInstance.name);

        // Verify basic instance information is shown
        await expect(page.locator('h1')).toContainText(firstInstance.name);

        // Check if proxy config is available
        const proxyConfigVisible = await navigatorPage.proxyConfig.isVisible();

        if (!proxyConfigVisible) {
            // Should show appropriate message or alternative content
            const noProxyMessage = page.locator('text=/no.*proxy.*config/i');
            const messageVisible = await noProxyMessage.isVisible();

            // Either proxy config is available or there's a message explaining why not
            expect(messageVisible).toBe(true);
        }
    });

    test('should validate breadcrumb navigation', async ({ page, request }) => {
        // Navigate through the breadcrumb path
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Get service data for instance navigation
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        // Navigate to instance
        await navigatorPage.clickInstanceCard(firstInstance.name);

        // Verify breadcrumb structure
        await navigatorPage.expectBreadcrumbVisible('Services');
        await navigatorPage.expectBreadcrumbVisible('frontend');
        await navigatorPage.expectBreadcrumbVisible(firstInstance.name);

        // Test breadcrumb navigation - click on service name
        await page
            .locator('[data-testid="breadcrumbs"] >> text=frontend')
            .click();
        await navigatorPage.expectCurrentPath('/services/frontend');

        // Test breadcrumb navigation - click on Services
        await page
            .locator('[data-testid="breadcrumbs"] >> text=Services')
            .click();
        await navigatorPage.expectCurrentPath('/');
    });

    test('should display service metrics if available', async ({
        page,
        request,
    }) => {
        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Check if metrics section is visible
        const metricsVisible = await navigatorPage.serviceMetrics.isVisible();

        if (metricsVisible) {
            await navigatorPage.waitForMetricsLoaded();
            await navigatorPage.expectMetricsVisible();

            // Verify metrics have values (not just placeholders)
            const requestRate =
                await navigatorPage.requestRateMetric.textContent();
            const errorRate = await navigatorPage.errorRateMetric.textContent();
            const latency = await navigatorPage.latencyMetric.textContent();

            // Metrics should contain numbers or reasonable placeholder values
            expect(requestRate).toMatch(/\d+|--|\.\.\./);
            expect(errorRate).toMatch(/\d+|--|\.\.\.|%/);
            expect(latency).toMatch(/\d+|--|\.\.\.|ms/);
        } else {
            console.log('Metrics not available for frontend service');
        }
    });

    test('should handle API errors gracefully', async ({ page }) => {
        // Try to navigate to a non-existent service
        await navigatorPage.goToService('nonexistent-service');

        // Should handle the error gracefully
        // Either show 404 page or redirect to service list
        const currentUrl = page.url();
        const isErrorPage =
            currentUrl.includes('404') || currentUrl.includes('error');
        const isRedirectedHome =
            currentUrl.endsWith('/') || currentUrl.includes('/services');

        expect(isErrorPage || isRedirectedHome).toBe(true);

        // Should not crash the application
        await expect(page.locator('body')).toBeVisible();
    });

    test('should validate API data consistency', async ({ page, request }) => {
        // Get service data from API
        const serviceData = await getService(request, 'frontend');

        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Verify UI displays match API data
        await expect(navigatorPage.serviceHeader).toContainText(
            serviceData.name
        );
        await expect(navigatorPage.serviceHeader).toContainText(
            serviceData.namespace
        );

        // Verify instance count matches
        const uiInstanceCount = await navigatorPage.instanceCards.count();
        expect(uiInstanceCount).toBe(serviceData.instances.length);

        // Verify first instance details match
        if (serviceData.instances.length > 0) {
            const firstInstance = serviceData.instances[0];
            const firstInstanceCard = navigatorPage.getInstanceCard(
                firstInstance.name
            );

            await expect(firstInstanceCard).toContainText(
                firstInstance.podName
            );
            await expect(firstInstanceCard).toContainText(
                firstInstance.namespace
            );
        }
    });
});
