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
import {
    waitForServicesLoaded,
    waitForServiceDetailsLoaded,
    waitForMetricsLoaded,
    waitForServiceConnectionsLoaded,
    waitForProxyConfigLoaded,
    waitForPageReady,
} from '../utils/wait-helpers';

/**
 * Page Object Model for Navigator UI
 */
export class NavigatorPage {
    constructor(private page: Page) {}

    // Navigation methods
    async goto(path: string = '/'): Promise<void> {
        await this.page.goto(path);
        await waitForPageReady(this.page);
    }

    async goToHome(): Promise<void> {
        await this.goto('/');
    }

    async goToService(serviceName: string): Promise<void> {
        await this.goto(`/services/${serviceName}`);
    }

    async goToServiceInstance(
        serviceName: string,
        instanceId: string
    ): Promise<void> {
        await this.goto(`/services/${serviceName}/instances/${instanceId}`);
    }

    // Header and navigation elements
    get navbar(): Locator {
        return this.page.locator(':has-text("Service Registry")').first();
    }

    get logo(): Locator {
        return this.page.locator('img').first(); // The Navigator logo image
    }

    get themeToggle(): Locator {
        return this.page.getByRole('button', { name: /toggle.*theme/i });
    }

    get breadcrumbs(): Locator {
        return this.page.locator('navigation[aria-label="breadcrumb"]');
    }

    // Service list page elements
    get serviceList(): Locator {
        return this.page.locator('table').first(); // The services table
    }

    get serviceRows(): Locator {
        return this.page.locator('table tbody tr');
    }

    getServiceRow(serviceName: string): Locator {
        return this.page.locator(
            `table tbody tr:has(td:has-text("${serviceName}"))`
        );
    }

    get clusterFilter(): Locator {
        return this.page.locator('button:has-text("cluster")');
    }

    // Service details page elements
    get serviceHeader(): Locator {
        return this.page.getByRole('link', {
            name: /frontend|backend|database/i,
            disabled: true,
        });
    }

    get serviceInstances(): Locator {
        return this.page
            .locator(':has-text("Service Instances")')
            .locator('..')
            .locator('table');
    }

    get instanceRows(): Locator {
        return this.page.locator('table:has(th:has-text("Pod")) tbody tr');
    }

    getInstanceRow(podName: string): Locator {
        return this.page.locator(
            `table tbody tr:has(td:has-text("${podName}"))`
        );
    }

    // Metrics elements
    get serviceMetrics(): Locator {
        return this.page
            .locator(':has-text("Service Connections")')
            .locator('..');
    }

    get requestRateMetric(): Locator {
        return this.page.locator(':has-text("requests")');
    }

    get errorRateMetric(): Locator {
        return this.page.locator(':has-text("error")');
    }

    get latencyMetric(): Locator {
        return this.page.locator(':has-text("latency")');
    }

    get metricsTimeRange(): Locator {
        return this.page.locator('combobox:has-text("minutes")');
    }

    get metricsRefresh(): Locator {
        return this.page.locator('button[disabled]:has(svg)').first();
    }

    // Service connections visualization
    get serviceConnections(): Locator {
        return this.page
            .locator(':has-text("Service Connections")')
            .locator('..');
    }

    get connectionGraph(): Locator {
        return this.page.locator('canvas, svg').first();
    }

    // Proxy configuration elements
    get proxyConfig(): Locator {
        return this.page.locator(':has-text("proxy", "config")').first();
    }

    get proxyConfigTabs(): Locator {
        return this.page.locator('[role="tablist"]');
    }

    get bootstrapTab(): Locator {
        return this.page.locator('[role="tab"]:has-text("bootstrap")');
    }

    get clustersTab(): Locator {
        return this.page.locator('[role="tab"]:has-text("clusters")');
    }

    get listenersTab(): Locator {
        return this.page.locator('[role="tab"]:has-text("listeners")');
    }

    get routesTab(): Locator {
        return this.page.locator('[role="tab"]:has-text("routes")');
    }

    get endpointsTab(): Locator {
        return this.page.locator('[role="tab"]:has-text("endpoints")');
    }

    get configEditor(): Locator {
        return this.page.locator('.cm-editor, textarea, pre');
    }

    // Istio resources elements
    get istioResources(): Locator {
        return this.page.locator('[data-testid="istio-resources"]');
    }

    get virtualServices(): Locator {
        return this.page.locator('[data-testid="virtual-services"]');
    }

    get destinationRules(): Locator {
        return this.page.locator('[data-testid="destination-rules"]');
    }

    get gateways(): Locator {
        return this.page.locator('[data-testid="gateways"]');
    }

    // Common action methods

    async clickServiceRow(serviceName: string): Promise<void> {
        await this.getServiceRow(serviceName).click();
        await waitForServiceDetailsLoaded(this.page, serviceName);
    }

    async clickInstanceRow(podName: string): Promise<void> {
        await this.getInstanceRow(podName).click();
        await waitForPageReady(this.page);
    }

    async toggleTheme(): Promise<void> {
        await this.themeToggle.click();
        await this.page.waitForTimeout(500); // Allow theme transition
    }

    async refreshMetrics(): Promise<void> {
        await this.metricsRefresh.click();
        await waitForMetricsLoaded(this.page);
    }

    async selectMetricsTimeRange(range: string): Promise<void> {
        await this.metricsTimeRange.selectOption(range);
        await waitForMetricsLoaded(this.page);
    }

    async switchProxyConfigTab(
        tab: 'bootstrap' | 'clusters' | 'listeners' | 'routes' | 'endpoints'
    ): Promise<void> {
        const tabLocator = this.page.locator(`[data-testid="tab-${tab}"]`);
        await tabLocator.click();
        await waitForProxyConfigLoaded(this.page);
    }

    // Wait methods
    async waitForServicesLoaded(minServices: number = 1): Promise<void> {
        await waitForServicesLoaded(this.page, minServices);
    }

    async waitForServiceDetailsLoaded(serviceName: string): Promise<void> {
        await waitForServiceDetailsLoaded(this.page, serviceName);
    }

    async waitForMetricsLoaded(): Promise<void> {
        await waitForMetricsLoaded(this.page);
    }

    async waitForServiceConnectionsLoaded(): Promise<void> {
        await waitForServiceConnectionsLoaded(this.page);
    }

    async waitForProxyConfigLoaded(): Promise<void> {
        await waitForProxyConfigLoaded(this.page);
    }

    // Validation methods
    async expectServiceVisible(serviceName: string): Promise<void> {
        await expect(this.getServiceRow(serviceName)).toBeVisible();
    }

    async expectServiceHasSidecar(serviceName: string): Promise<void> {
        const serviceRow = this.getServiceRow(serviceName);
        // Look for the specific purple hexagon sidecar indicator
        await expect(serviceRow.locator('svg.lucide-hexagon')).toBeVisible();
    }

    async expectMetricsVisible(): Promise<void> {
        await expect(this.serviceMetrics).toBeVisible();
        await expect(this.requestRateMetric).toBeVisible();
        await expect(this.errorRateMetric).toBeVisible();
        await expect(this.latencyMetric).toBeVisible();
    }

    async expectProxyConfigVisible(): Promise<void> {
        await expect(this.proxyConfig).toBeVisible();
        await expect(this.configEditor).toBeVisible();
    }

    async expectConnectionGraphVisible(): Promise<void> {
        await expect(this.serviceConnections).toBeVisible();
        // Check for either canvas or SVG (depending on visualization library)
        await Promise.race([
            expect(this.page.locator('canvas')).toBeVisible(),
            expect(this.page.locator('svg')).toBeVisible(),
        ]);
    }

    // Navigation validation
    async expectCurrentPath(path: string): Promise<void> {
        await expect(this.page).toHaveURL(
            new RegExp(path.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
        );
    }

    async expectBreadcrumbVisible(text: string): Promise<void> {
        await expect(this.breadcrumbs.locator('text=' + text)).toBeVisible();
    }
}
