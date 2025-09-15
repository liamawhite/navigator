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

import { Page, Locator, expect } from '@playwright/test';

/**
 * Wait for services to be loaded and displayed in the service list
 */
export async function waitForServicesLoaded(
    page: Page,
    minServices: number = 1
): Promise<void> {
    // Wait for the service table to be visible
    await expect(page.locator('table')).toBeVisible({ timeout: 30000 });

    // Wait for at least the minimum number of service rows to appear
    await expect(page.locator('table tbody tr')).toHaveCount(minServices, {
        timeout: 30000,
    });

    // Wait for services count to be displayed
    await expect(page.locator('span:has-text("services")').first()).toBeVisible(
        { timeout: 30000 }
    );
}

/**
 * Wait for a specific service to appear in the service list
 */
export async function waitForServiceToAppear(
    page: Page,
    serviceName: string
): Promise<Locator> {
    const serviceRow = page.locator(
        `table tbody tr:has(cell:has-text("${serviceName}"))`
    );
    await expect(serviceRow).toBeVisible({ timeout: 30000 });
    return serviceRow;
}

/**
 * Wait for service details to load
 */
export async function waitForServiceDetailsLoaded(
    page: Page,
    serviceName: string
): Promise<void> {
    // Wait for service instances table to load (more specific than service name)
    await expect(page.locator('table:has(th:has-text("Pod"))')).toBeVisible({
        timeout: 30000,
    });

    // Wait for breadcrumbs to load
    await expect(
        page.getByRole('navigation', { name: 'breadcrumb' })
    ).toBeVisible({ timeout: 30000 });

    // Wait for the service name in breadcrumb (more specific)
    await expect(
        page.locator(
            'span[role="link"][aria-current="page"]:has-text("' +
                serviceName +
                '")'
        )
    ).toBeVisible({ timeout: 30000 });
}

/**
 * Wait for metrics to load and be displayed
 */
export async function waitForMetricsLoaded(page: Page): Promise<void> {
    // Wait for metrics section to be visible
    await expect(page.locator('[data-testid="service-metrics"]')).toBeVisible({
        timeout: 30000,
    });

    // Wait for at least one metric value to appear (not "--" or loading)
    await page.waitForFunction(
        () => {
            const metricElements = document.querySelectorAll(
                '[data-testid^="metric-"]'
            );
            return Array.from(metricElements).some((el) => {
                const text = el.textContent || '';
                return text !== '--' && text !== '...' && text.trim() !== '';
            });
        },
        {},
        { timeout: 30000 }
    );
}

/**
 * Wait for the service connections visualization to load
 */
export async function waitForServiceConnectionsLoaded(
    page: Page
): Promise<void> {
    // Wait for the connections container to be visible
    await expect(
        page.locator('[data-testid="service-connections"]')
    ).toBeVisible({ timeout: 30000 });

    // Wait for the graph to render (either canvas or SVG elements)
    await Promise.race([
        page
            .locator('canvas')
            .first()
            .waitFor({ state: 'visible', timeout: 30000 }),
        page
            .locator('svg')
            .first()
            .waitFor({ state: 'visible', timeout: 30000 }),
    ]);

    // Wait a bit more for the visualization to settle
    await page.waitForTimeout(2000);
}

/**
 * Wait for proxy configuration to load
 */
export async function waitForProxyConfigLoaded(page: Page): Promise<void> {
    // Wait for proxy config section to be visible
    await expect(page.locator('[data-testid="proxy-config"]')).toBeVisible({
        timeout: 30000,
    });

    // Wait for config data to load (CodeMirror editor should be present)
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 30000 });

    // Wait for content to appear in the editor
    await page.waitForFunction(
        () => {
            const editor = document.querySelector('.cm-editor .cm-content');
            return (
                editor &&
                editor.textContent &&
                editor.textContent.trim().length > 10
            );
        },
        {},
        { timeout: 30000 }
    );
}

/**
 * Wait for navigation to complete and page to be ready
 */
export async function waitForPageReady(page: Page): Promise<void> {
    // Wait for network to be idle
    await page.waitForLoadState('networkidle');

    // Wait for React app to load - check for root div content
    await expect(page.locator('#root')).not.toBeEmpty({ timeout: 30000 });

    // Wait for any global loading indicators to disappear (if they exist)
    const globalLoading = page.locator('[data-testid="global-loading"]');
    const hasGlobalLoading = await globalLoading.isVisible();
    if (hasGlobalLoading) {
        await expect(globalLoading).not.toBeVisible({ timeout: 30000 });
    }
}

/**
 * Wait for API requests to complete
 */
export async function waitForApiRequests(
    page: Page,
    pattern: string | RegExp
): Promise<void> {
    await page.waitForResponse(
        (response) => {
            const url = response.url();
            if (typeof pattern === 'string') {
                return url.includes(pattern);
            }
            return pattern.test(url);
        },
        { timeout: 30000 }
    );
}

/**
 * Wait for theme to be applied
 */
export async function waitForThemeApplied(
    page: Page,
    theme: 'light' | 'dark'
): Promise<void> {
    await expect(page.locator('html')).toHaveClass(new RegExp(theme), {
        timeout: 10000,
    });
}

/**
 * Smart retry helper for flaky operations
 */
export async function retryOperation<T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    delayMs: number = 1000
): Promise<T> {
    let lastError: unknown;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
            return await operation();
        } catch (error) {
            lastError = error;

            if (attempt === maxRetries) {
                throw lastError;
            }

            console.log(
                `Attempt ${attempt} failed, retrying in ${delayMs}ms...`
            );
            await new Promise((resolve) => setTimeout(resolve, delayMs));
        }
    }

    throw lastError;
}

/**
 * Wait for element with custom text matching
 */
export async function waitForElementWithText(
    page: Page,
    selector: string,
    text: string | RegExp,
    options: { timeout?: number; exact?: boolean } = {}
): Promise<Locator> {
    const { timeout = 30000, exact = false } = options;

    const element = page.locator(selector);

    if (exact) {
        await expect(element).toHaveText(text, { timeout });
    } else {
        await expect(element).toContainText(text, { timeout });
    }

    return element;
}
