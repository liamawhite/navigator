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
import { access, constants } from 'fs';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);
const accessAsync = promisify(access);

/**
 * Global setup runs once before all tests
 * Ensures fresh navctl binary is built before testing
 */
async function globalSetup() {
    console.log('🔨 Starting Navigator E2E test setup...');

    // Change to project root directory
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = dirname(__filename);
    const projectRoot = join(__dirname, '../..');
    process.chdir(projectRoot);

    try {
        // Build fresh navctl binary with embedded UI
        console.log('📦 Building fresh navctl binary...');
        const buildResult = await execAsync('make build-navctl-dev', {
            timeout: 5 * 60 * 1000, // 5 minutes timeout
        });

        if (buildResult.stderr && !buildResult.stderr.includes('warning')) {
            console.warn('Build warnings:', buildResult.stderr);
        }

        // Verify the binary was created and is executable
        const binaryPath = join(projectRoot, 'bin/navctl');
        await accessAsync(binaryPath, constants.F_OK | constants.X_OK);

        // Test the binary works
        console.log('✅ Verifying navctl binary...');
        const versionResult = await execAsync('./bin/navctl version');
        console.log('📋 Navigator version:', versionResult.stdout.trim());

        // Optionally setup demo cluster if requested
        if (process.env.E2E_SETUP_DEMO === 'true') {
            console.log('🎭 Setting up demo cluster...');
            const demoName = process.env.E2E_DEMO_NAME || 'navigator-e2e';

            try {
                // Check if demo cluster exists
                await execAsync(
                    `./bin/navctl demo start --name ${demoName} --istio-version 1.25.4`,
                    {
                        timeout: 10 * 60 * 1000, // 10 minutes for demo setup
                    }
                );
                console.log('✅ Demo cluster ready');
            } catch (demoError: unknown) {
                const errorMessage =
                    demoError instanceof Error
                        ? demoError.message
                        : 'Unknown error';
                console.warn(
                    '⚠️  Demo setup failed, tests will need existing cluster:',
                    errorMessage
                );
            }
        }

        console.log('🚀 Global setup completed successfully');
    } catch (error: unknown) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.error('❌ Global setup failed:', errorMessage);
        throw error;
    }
}

export default globalSetup;
