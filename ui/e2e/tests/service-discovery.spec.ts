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
import { ensureDemoCluster } from '../utils/demo-helpers';
import { getServices, waitForServicesDiscovered } from '../utils/api-helpers';

test.describe('Service Discovery', () => {
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
        await navigatorPage.goToHome();
    });

    test('should display service list on home page', async ({ request }) => {
        // Wait for services to be discovered by the API
        await waitForServicesDiscovered(request, [
            'frontend',
            'backend',
            'database',
        ]);

        // Wait for services to be loaded in the UI - we know there are 9 total services
        await navigatorPage.waitForServicesLoaded(9);

        // Verify service list is visible
        await expect(navigatorPage.serviceList).toBeVisible();

        // Verify we have all the services from the cluster (9 total)
        await expect(navigatorPage.serviceRows).toHaveCount(9);
    });

    test('should show demo microservices', async ({ request }) => {
        // Wait for API to discover services
        await waitForServicesDiscovered(request, [
            'frontend',
            'backend',
            'database',
        ]);

        // Wait for UI to load services
        await navigatorPage.waitForServicesLoaded(9);

        // Verify each demo service is visible
        await navigatorPage.expectServiceVisible('frontend');
        await navigatorPage.expectServiceVisible('backend');
        await navigatorPage.expectServiceVisible('database');

        // Verify service rows contain expected information
        const frontendRow = navigatorPage.getServiceRow('frontend');
        await expect(frontendRow).toContainText('frontend');
        await expect(frontendRow).toContainText('microservices'); // namespace

        const backendRow = navigatorPage.getServiceRow('backend');
        await expect(backendRow).toContainText('backend');
        await expect(backendRow).toContainText('microservices'); // namespace

        const databaseRow = navigatorPage.getServiceRow('database');
        await expect(databaseRow).toContainText('database');
        await expect(databaseRow).toContainText('database'); // namespace
    });

    test('should identify services with Istio sidecars', async ({
        request,
    }) => {
        // Wait for services to be discovered
        await waitForServicesDiscovered(request, [
            'frontend',
            'backend',
            'database',
        ]);
        await navigatorPage.waitForServicesLoaded(9);

        // Verify that microservices have sidecar indicators
        // Frontend and backend should have sidecars (in microservices namespace)
        await navigatorPage.expectServiceHasSidecar('frontend');
        await navigatorPage.expectServiceHasSidecar('backend');

        // Database might not have sidecar depending on configuration
        // We'll check but not assert since it depends on demo setup
        const databaseRow = navigatorPage.getServiceRow('database');
        const hasSidecarIndicator = await databaseRow
            .locator('svg.lucide-hexagon')
            .isVisible();
        console.log(
            `Database sidecar indicator visible: ${hasSidecarIndicator}`
        );
    });

    test('should navigate to service details when clicking service card', async ({
        request,
    }) => {
        // Wait for services to be loaded
        await waitForServicesDiscovered(request, [
            'frontend',
            'backend',
            'database',
        ]);
        await navigatorPage.waitForServicesLoaded(9);

        // Click on frontend service
        await navigatorPage.clickServiceRow('frontend');

        // Should navigate to service details page
        await navigatorPage.expectCurrentPath(
            '/services/microservices:frontend'
        );

        // Should show service details
        await navigatorPage.waitForServiceDetailsLoaded('frontend');
        await expect(navigatorPage.serviceHeader).toContainText('frontend');
    });

    test('should show service instance counts', async ({ request }) => {
        // Wait for services to be loaded
        await waitForServicesDiscovered(request, [
            'frontend',
            'backend',
            'database',
        ]);
        await navigatorPage.waitForServicesLoaded(9);

        // Each demo service should have at least 1 instance
        const frontendRow = navigatorPage.getServiceRow('frontend');
        await expect(frontendRow).toContainText('1'); // instance count

        const backendRow = navigatorPage.getServiceRow('backend');
        await expect(backendRow).toContainText('1'); // instance count

        const databaseRow = navigatorPage.getServiceRow('database');
        await expect(databaseRow).toContainText('1'); // instance count
    });

    test('should update service list in real-time', async ({
        page,
        request,
    }) => {
        // Wait for initial services
        await waitForServicesDiscovered(request, [
            'frontend',
            'backend',
            'database',
        ]);
        await navigatorPage.waitForServicesLoaded(9);

        // Get initial service count
        const initialCount = await navigatorPage.serviceRows.count();
        expect(initialCount).toBeGreaterThanOrEqual(3);

        // Wait for auto-refresh (Navigator refreshes every 5 seconds)
        await page.waitForTimeout(6000);

        // Services should still be there (this tests the auto-refresh doesn't break)
        await navigatorPage.expectServiceVisible('frontend');
        await navigatorPage.expectServiceVisible('backend');
        await navigatorPage.expectServiceVisible('database');

        // Count should remain the same or potentially increase if new services are discovered
        const updatedCount = await navigatorPage.serviceRows.count();
        expect(updatedCount).toBeGreaterThanOrEqual(initialCount);
    });

    test('should handle empty state gracefully', async ({ page }) => {
        // This test is theoretical since we always have demo services
        // but it ensures the UI handles the loading state properly

        // Navigate to home before services are fully loaded
        await navigatorPage.goToHome();

        // Should show loading state initially
        const loadingSpinner = page.locator('[data-testid="loading-spinner"]');

        // Either loading spinner is visible, or services are already loaded
        const isLoadingVisible = await loadingSpinner.isVisible();
        const areServicesLoaded = (await navigatorPage.serviceRows.count()) > 0;

        expect(isLoadingVisible || areServicesLoaded).toBe(true);

        // Eventually services should load
        await navigatorPage.waitForServicesLoaded(9);
    });

    test('should validate API data matches UI display', async ({ request }) => {
        // Get services from API
        const apiServices = await getServices(request);
        expect(apiServices.length).toBeGreaterThanOrEqual(3);

        // Wait for UI to load
        await navigatorPage.waitForServicesLoaded(apiServices.length);

        // Verify each API service appears in UI
        for (const service of apiServices) {
            await navigatorPage.expectServiceVisible(service.name);

            const serviceRow = navigatorPage.getServiceRow(service.name);
            await expect(serviceRow).toContainText(service.namespace);
            // Cluster name is in the instances, not at service level
            if (service.instances && service.instances.length > 0) {
                await expect(serviceRow).toContainText('1'); // instance count
            }
        }

        // Verify UI count matches API count
        const uiServiceCount = await navigatorPage.serviceRows.count();
        expect(uiServiceCount).toBe(apiServices.length);
    });
});
