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

import { chromium } from '@playwright/test';

// Import the actual api helpers to test them
import {
    getServices,
    waitForServicesDiscovered,
} from './e2e/utils/api-helpers.js';

async function testApi() {
    const browser = await chromium.launch();
    const context = await browser.newContext();

    try {
        console.log('Testing API with Playwright request context...');

        // Test the API with full URL
        const response = await context.request.get(
            'http://localhost:8081/api/v1alpha1/services'
        );
        console.log('API Response status:', response.status());

        if (response.status() === 200) {
            const data = await response.json();
            console.log('Services found:', data.services?.length || 0);
            console.log(
                'Service names:',
                data.services?.map((s) => s.name).slice(0, 5) || []
            );
            console.log('✅ API test successful');
        } else {
            console.log('❌ API test failed - status:', response.status());
        }

        // Now test the actual helper functions
        console.log('\nTesting getServices helper...');
        try {
            const services = await getServices(context.request);
            console.log('getServices result:', services.length, 'services');
            console.log(
                'Service names via helper:',
                services.map((s) => s.name).slice(0, 5)
            );
            console.log('✅ getServices test successful');
        } catch (error) {
            const errorMessage =
                error instanceof Error ? error.message : 'Unknown error';
            const stackTrace =
                error instanceof Error ? error.stack : 'No stack trace';
            console.error('❌ getServices test error:', errorMessage);
            console.error('Stack trace:', stackTrace);
        }

        // Test waitForServicesDiscovered with a short timeout
        console.log('\nTesting waitForServicesDiscovered helper...');
        try {
            const result = await waitForServicesDiscovered(
                context.request,
                ['frontend', 'backend'],
                10000
            );
            console.log('waitForServicesDiscovered result:', result);
            console.log('✅ waitForServicesDiscovered test successful');
        } catch (error) {
            const errorMessage =
                error instanceof Error ? error.message : 'Unknown error';
            const stackTrace =
                error instanceof Error ? error.stack : 'No stack trace';
            console.error(
                '❌ waitForServicesDiscovered test error:',
                errorMessage
            );
            console.error('Stack trace:', stackTrace);
        }
    } catch (error) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.error('❌ API test error:', errorMessage);
    } finally {
        await browser.close();
    }
}

testApi();
