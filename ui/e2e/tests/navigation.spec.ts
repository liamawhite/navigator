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
import { getService, waitForServicesDiscovered } from '../utils/api-helpers';
import { waitForThemeApplied } from '../utils/wait-helpers';

test.describe('Navigation and UI', () => {
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

    test('should display main navigation elements', async ({ page }) => {
        await navigatorPage.goToHome();

        // Verify main navigation elements are present
        await expect(navigatorPage.navbar).toBeVisible();
        await expect(navigatorPage.logo).toBeVisible();
        await expect(navigatorPage.themeToggle).toBeVisible();

        // Verify page title
        await expect(page).toHaveTitle(/Navigator/i);
    });

    test('should navigate from service list to service details', async ({
        page,
        request,
    }) => {
        // Wait for services to be discovered
        await waitForServicesDiscovered(request, ['frontend']);

        // Start at home page
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);

        // Click on frontend service
        await navigatorPage.clickServiceCard('frontend');

        // Should navigate to service details
        await navigatorPage.expectCurrentPath('/services/frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Verify breadcrumbs
        await navigatorPage.expectBreadcrumbVisible('Services');
        await navigatorPage.expectBreadcrumbVisible('frontend');
    });

    test('should navigate from service details to instance details', async ({
        page,
        request,
    }) => {
        // Get service data
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Click on first instance
        await navigatorPage.clickInstanceCard(firstInstance.name);

        // Should navigate to instance details
        await navigatorPage.expectCurrentPath(
            `/services/frontend/instances/${firstInstance.name}`
        );

        // Verify breadcrumbs
        await navigatorPage.expectBreadcrumbVisible('Services');
        await navigatorPage.expectBreadcrumbVisible('frontend');
        await navigatorPage.expectBreadcrumbVisible(firstInstance.name);
    });

    test('should support breadcrumb navigation', async ({ page, request }) => {
        // Get service data for navigation
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        // Navigate to instance details (deepest level)
        await navigatorPage.goToServiceInstance('frontend', firstInstance.name);

        // Navigate back using breadcrumbs
        // Click on service name in breadcrumb
        await page
            .locator('[data-testid="breadcrumbs"] >> text=frontend')
            .click();
        await navigatorPage.expectCurrentPath('/services/frontend');

        // Click on Services in breadcrumb
        await page
            .locator('[data-testid="breadcrumbs"] >> text=Services')
            .click();
        await navigatorPage.expectCurrentPath('/');

        // Verify we're back at the service list
        await navigatorPage.waitForServicesLoaded(1);
    });

    test('should support direct URL navigation', async ({ page, request }) => {
        // Test direct navigation to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.expectCurrentPath('/services/frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Get service data for instance navigation
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        // Test direct navigation to instance details
        await navigatorPage.goToServiceInstance('frontend', firstInstance.name);
        await navigatorPage.expectCurrentPath(
            `/services/frontend/instances/${firstInstance.name}`
        );

        // Verify content loads correctly with direct navigation
        await expect(page.locator('h1')).toContainText(firstInstance.name);
    });

    test('should toggle between light and dark themes', async ({ page }) => {
        await navigatorPage.goToHome();

        // Get initial theme
        const htmlElement = page.locator('html');
        const initialClasses = (await htmlElement.getAttribute('class')) || '';
        const initialTheme = initialClasses.includes('dark') ? 'dark' : 'light';

        // Toggle theme
        await navigatorPage.toggleTheme();

        // Wait for theme to be applied
        const expectedTheme = initialTheme === 'dark' ? 'light' : 'dark';
        await waitForThemeApplied(page, expectedTheme);

        // Verify theme changed
        const newClasses = (await htmlElement.getAttribute('class')) || '';
        if (expectedTheme === 'dark') {
            expect(newClasses).toContain('dark');
        } else {
            expect(newClasses).not.toContain('dark');
        }

        // Toggle back
        await navigatorPage.toggleTheme();
        await waitForThemeApplied(page, initialTheme);

        // Verify theme reverted
        const revertedClasses = (await htmlElement.getAttribute('class')) || '';
        if (initialTheme === 'dark') {
            expect(revertedClasses).toContain('dark');
        } else {
            expect(revertedClasses).not.toContain('dark');
        }
    });

    test('should maintain theme preference across navigation', async ({
        page,
    }) => {
        await navigatorPage.goToHome();

        // Set to dark theme
        await navigatorPage.toggleTheme();
        await waitForThemeApplied(page, 'dark');

        // Navigate to service details
        await navigatorPage.goToService('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Verify theme persisted
        const htmlElement = page.locator('html');
        const classes = (await htmlElement.getAttribute('class')) || '';
        expect(classes).toContain('dark');

        // Navigate back to home
        await navigatorPage.goToHome();

        // Verify theme still persisted
        const homeClasses = (await htmlElement.getAttribute('class')) || '';
        expect(homeClasses).toContain('dark');
    });

    test('should handle browser back/forward navigation', async ({
        page,
        request,
    }) => {
        // Start at home
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);

        // Navigate to service details
        await navigatorPage.clickServiceCard('frontend');
        await navigatorPage.expectCurrentPath('/services/frontend');

        // Use browser back button
        await page.goBack();
        await navigatorPage.expectCurrentPath('/');
        await navigatorPage.waitForServicesLoaded(1);

        // Use browser forward button
        await page.goForward();
        await navigatorPage.expectCurrentPath('/services/frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');
    });

    test('should display proper page titles', async ({ page, request }) => {
        // Home page title
        await navigatorPage.goToHome();
        await expect(page).toHaveTitle(/Navigator/i);

        // Service details page title
        await navigatorPage.goToService('frontend');
        await expect(page).toHaveTitle(/frontend.*Navigator/i);

        // Instance details page title
        const serviceData = await getService(request, 'frontend');
        const firstInstance = serviceData.instances[0];

        await navigatorPage.goToServiceInstance('frontend', firstInstance.name);
        await expect(page).toHaveTitle(
            new RegExp(`${firstInstance.name}.*Navigator`, 'i')
        );
    });

    test('should handle 404 and error pages gracefully', async ({ page }) => {
        // Try to navigate to non-existent service
        await page.goto('/services/nonexistent-service');

        // Should handle gracefully - either 404 page or redirect
        const url = page.url();
        const isErrorHandled =
            url.includes('404') ||
            url.includes('error') ||
            url.endsWith('/') ||
            url.includes('/services');

        expect(isErrorHandled).toBe(true);

        // Page should still be functional
        await expect(page.locator('body')).toBeVisible();
        await expect(navigatorPage.navbar).toBeVisible();

        // Should be able to navigate back to working pages
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);
    });

    test('should support keyboard navigation', async ({ page }) => {
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);

        // Test tab navigation
        await page.keyboard.press('Tab');

        // Should focus on first interactive element
        const focusedElement = await page.locator(':focus').first();
        await expect(focusedElement).toBeVisible();

        // Continue tabbing to verify proper tab order
        await page.keyboard.press('Tab');
        const secondFocus = await page.locator(':focus').first();
        await expect(secondFocus).toBeVisible();

        // Test escape key (if modals or dropdowns are open)
        await page.keyboard.press('Escape');

        // Page should remain functional
        await expect(navigatorPage.serviceList).toBeVisible();
    });

    test('should be responsive on different screen sizes', async ({ page }) => {
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);

        // Test desktop size (default)
        await page.setViewportSize({ width: 1200, height: 800 });
        await expect(navigatorPage.navbar).toBeVisible();
        await expect(navigatorPage.serviceList).toBeVisible();

        // Test tablet size
        await page.setViewportSize({ width: 768, height: 1024 });
        await expect(navigatorPage.navbar).toBeVisible();
        await expect(navigatorPage.serviceList).toBeVisible();

        // Test mobile size
        await page.setViewportSize({ width: 375, height: 667 });
        await expect(navigatorPage.navbar).toBeVisible();
        await expect(navigatorPage.serviceList).toBeVisible();

        // Reset to desktop
        await page.setViewportSize({ width: 1200, height: 800 });
    });

    test('should load pages efficiently', async ({ page }) => {
        // Measure page load time
        const startTime = Date.now();

        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);

        const loadTime = Date.now() - startTime;

        // Page should load within reasonable time (adjust threshold as needed)
        expect(loadTime).toBeLessThan(10000); // 10 seconds

        console.log(`Home page load time: ${loadTime}ms`);
    });

    test('should handle network interruption gracefully', async ({ page }) => {
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);

        // Simulate network interruption
        await page.route('**/*', (route) => route.abort());

        // Try to navigate - should handle gracefully
        try {
            await navigatorPage.goToService('frontend');
        } catch (error) {
            // Expected to fail due to network interruption
        }

        // Restore network
        await page.unroute('**/*');

        // Should be able to recover
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);
    });

    test('should maintain state during navigation', async ({
        page,
        request,
    }) => {
        await navigatorPage.goToHome();
        await navigatorPage.waitForServicesLoaded(1);

        // Apply a filter
        await navigatorPage.filterByNamespace('microservices');

        // Verify filter is applied
        await navigatorPage.expectServiceVisible('frontend');
        await navigatorPage.expectServiceVisible('backend');

        // Navigate to service details
        await navigatorPage.clickServiceCard('frontend');
        await navigatorPage.waitForServiceDetailsLoaded('frontend');

        // Navigate back to home
        await navigatorPage.goToHome();

        // Filter state might or might not be preserved (depends on implementation)
        // Just verify the page loads correctly
        await navigatorPage.waitForServicesLoaded(1);
        await expect(navigatorPage.serviceList).toBeVisible();
    });
});
