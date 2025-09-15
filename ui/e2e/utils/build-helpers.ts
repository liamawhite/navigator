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
import { access, constants } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);
const accessAsync = promisify(access);

export interface BuildResult {
    success: boolean;
    stdout: string;
    stderr: string;
    duration: number;
}

export interface BinaryInfo {
    path: string;
    version: string;
    exists: boolean;
    executable: boolean;
}

/**
 * Build Navigator components using make targets
 */
export async function buildNavigator(component?: string): Promise<BuildResult> {
    const startTime = Date.now();
    const target = component ? `build-${component}` : 'build';

    try {
        console.log(`🔨 Building Navigator component: ${target}`);
        const __filename = fileURLToPath(import.meta.url);
        const __dirname = dirname(__filename);
        const result = await execAsync(`make ${target}`, {
            timeout: 5 * 60 * 1000, // 5 minutes timeout
            cwd: join(__dirname, '../../..'), // Project root
        });

        const duration = Date.now() - startTime;

        return {
            success: true,
            stdout: result.stdout,
            stderr: result.stderr,
            duration,
        };
    } catch (error: unknown) {
        const duration = Date.now() - startTime;
        const execError = error as {
            stdout?: string;
            stderr?: string;
            message?: string;
        };

        return {
            success: false,
            stdout: execError.stdout || '',
            stderr:
                execError.stderr || execError.message || 'Unknown build error',
            duration,
        };
    }
}

/**
 * Verify that a binary exists and is executable
 */
export async function verifyBinary(binaryPath: string): Promise<BinaryInfo> {
    const info: BinaryInfo = {
        path: binaryPath,
        version: '',
        exists: false,
        executable: false,
    };

    try {
        // Check if file exists and is executable
        await accessAsync(binaryPath, constants.F_OK);
        info.exists = true;

        await accessAsync(binaryPath, constants.X_OK);
        info.executable = true;

        // Get version information
        const versionResult = await execAsync(`${binaryPath} version`, {
            timeout: 10 * 1000, // 10 seconds
        });
        info.version = versionResult.stdout.trim();
    } catch (error: unknown) {
        // Binary doesn't exist or isn't executable
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.warn(
            `Binary verification failed for ${binaryPath}:`,
            errorMessage
        );
    }

    return info;
}

/**
 * Ensure navctl is built and ready for testing
 */
export async function ensureNavctlReady(): Promise<BinaryInfo> {
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = dirname(__filename);
    const projectRoot = join(__dirname, '../../..');
    const binaryPath = join(projectRoot, 'bin/navctl');

    let info = await verifyBinary(binaryPath);

    if (!info.exists || !info.executable) {
        console.log(
            '🔨 navctl binary not found or not executable, building...'
        );

        const buildResult = await buildNavigator('navctl-dev');
        if (!buildResult.success) {
            throw new Error(`Failed to build navctl: ${buildResult.stderr}`);
        }

        // Verify again after build
        info = await verifyBinary(binaryPath);
        if (!info.exists || !info.executable) {
            throw new Error('navctl binary still not ready after build');
        }
    }

    console.log(`✅ navctl ready: ${info.version}`);
    return info;
}

/**
 * Clean build artifacts
 */
export async function cleanBuild(): Promise<void> {
    try {
        const __filename = fileURLToPath(import.meta.url);
        const __dirname = dirname(__filename);
        await execAsync('make clean', {
            cwd: join(__dirname, '../../..'),
            timeout: 30 * 1000, // 30 seconds
        });
        console.log('🧹 Build artifacts cleaned');
    } catch (error: unknown) {
        const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
        console.warn('⚠️  Failed to clean build artifacts:', errorMessage);
    }
}
