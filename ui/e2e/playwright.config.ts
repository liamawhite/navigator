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

import { defineConfig, devices } from '@playwright/test';

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
    testDir: './tests',

    /* Run tests in files in parallel */
    fullyParallel: false, // Keep sequential for demo environment stability

    /* Fail the build on CI if you accidentally left test.only in the source code. */
    forbidOnly: !!process.env.CI,

    /* Retry on CI only */
    retries: process.env.CI ? 2 : 0,

    /* Opt out of parallel tests on CI. */
    workers: process.env.CI ? 1 : 1, // Single worker to avoid demo conflicts

    /* Reporter to use. See https://playwright.dev/docs/test-reporters */
    reporter: [
        ['list'],
        ['html', { outputFolder: 'playwright-report' }],
        ['junit', { outputFile: 'test-results/results.xml' }],
    ],

    /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
    use: {
        /* Base URL to use in actions like `await page.goto('/')`. */
        baseURL: 'http://localhost:8082',

        /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
        trace: 'on-first-retry',

        /* Take screenshot on failure */
        screenshot: 'only-on-failure',

        /* Record video on failure */
        video: 'retain-on-failure',

        /* Navigation timeout */
        navigationTimeout: 30 * 1000,

        /* Action timeout */
        actionTimeout: 10 * 1000,
    },

    /* Configure projects for major browsers */
    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],

    /* Assume Navigator is already running on port 8082 */
    // webServer: {
    //   command: process.env.E2E_NAVCTL_COMMAND || 'bin/navctl local --no-browser --metrics-endpoint http://localhost:30090',
    //   port: 8082,
    //   timeout: 120 * 1000, // 2 minutes for services to start
    //   reuseExistingServer: !process.env.CI, // Allow reuse in local development
    //   stdout: 'pipe',
    //   stderr: 'pipe',
    //   cwd: '..',  // Run from project root where bin/navctl exists
    // },

    /* Global timeout for each test */
    timeout: 60 * 1000, // 1 minute per test

    /* Global setup */
    globalSetup: './global-setup.ts',

    /* Test directory patterns */
    testMatch: ['**/*.spec.ts'],

    /* Output directories */
    outputDir: 'test-results',
});
