# Navigator E2E Testing Framework

This directory contains end-to-end tests for the Navigator UI using Playwright. The tests verify user stories against a real Navigator deployment with a demo Kubernetes cluster.

## Overview

The E2E testing framework includes:

- **Service Discovery Tests**: Verify service list, filtering, and search functionality
- **Service Details Tests**: Test service instance views and proxy configuration
- **Metrics Visualization Tests**: Validate metrics display and service connections topology
- **Navigation Tests**: Ensure proper routing, breadcrumbs, and UI interactions

## Quick Start

### Prerequisites

1. **Nix Development Environment** (recommended):
   ```bash
   nix develop
   ```

2. **Or install dependencies manually**:
   - Go 1.21+
   - Node.js 18+
   - Docker
   - kubectl
   - kind

### Running Tests

1. **Simple E2E test run** (assumes demo cluster exists):
   ```bash
   make test-e2e
   ```

2. **Full setup with demo cluster creation**:
   ```bash
   make test-e2e-setup
   ```

3. **CI mode** (with automatic demo setup):
   ```bash
   make test-e2e-ci
   ```

4. **Debug mode** (headed browser with inspector):
   ```bash
   make test-e2e-debug
   ```

5. **UI mode** (Playwright's interactive test runner):
   ```bash
   make test-e2e-ui
   ```

### Development Workflow

For active development:

1. **Start demo cluster once**:
   ```bash
   ./bin/navctl demo start --name navigator-e2e
   ```

2. **Run tests repeatedly during development**:
   ```bash
   cd ui && npm run e2e
   ```

3. **Clean up when done**:
   ```bash
   ./bin/navctl demo stop --name navigator-e2e
   ```

## Framework Architecture

### Directory Structure

```
ui/e2e/
├── tests/                        # Test specifications
│   ├── service-discovery.spec.ts
│   ├── service-details.spec.ts
│   ├── metrics-visualization.spec.ts
│   └── navigation.spec.ts
├── fixtures/                     # Test fixtures and page objects
│   └── navigator-page.ts         # Page Object Model for Navigator UI
├── utils/                        # Utility functions
│   ├── build-helpers.ts          # Build and binary management
│   ├── demo-helpers.ts           # Demo cluster management
│   ├── wait-helpers.ts           # Custom wait conditions
│   └── api-helpers.ts            # API testing utilities
├── playwright.config.ts          # Playwright configuration
├── global-setup.ts              # Global test setup
└── README.md                    # This file
```

### Key Components

1. **Fresh Build Integration**: Every test run builds a fresh `navctl` binary to ensure tests run against the latest code changes.

2. **Demo Environment Management**: Automated setup and teardown of Kind clusters with Istio and microservices.

3. **Page Object Model**: Centralized UI interaction patterns for maintainable tests.

4. **Smart Waits**: Custom wait conditions that understand Navigator's async loading patterns.

5. **API Validation**: Direct API calls to verify UI accuracy and catch data inconsistencies.

## Test Scenarios

### Service Discovery (`service-discovery.spec.ts`)

Tests the main service list functionality:
- ✅ Display services from demo cluster
- ✅ Identify services with Istio sidecars
- ✅ Filter services by namespace
- ✅ Search for specific services
- ✅ Navigate to service details
- ✅ Real-time service updates
- ✅ API data consistency

### Service Details (`service-details.spec.ts`)

Tests individual service views:
- ✅ Service instance information
- ✅ Proxy configuration for sidecar-enabled services
- ✅ Navigation between service and instance views
- ✅ Istio resource display
- ✅ Breadcrumb navigation
- ✅ Error handling for missing services

### Metrics Visualization (`metrics-visualization.spec.ts`)

Tests metrics and topology features:
- ✅ Service metrics display (request rate, error rate, latency)
- ✅ Metrics time range selection
- ✅ Manual metrics refresh
- ✅ Service connections visualization
- ✅ Auto-refresh functionality
- ✅ Load generation metrics validation

### Navigation (`navigation.spec.ts`)

Tests UI navigation and interactions:
- ✅ Route navigation and URL handling
- ✅ Breadcrumb navigation
- ✅ Theme switching (light/dark)
- ✅ Browser back/forward support
- ✅ Responsive design
- ✅ Keyboard navigation
- ✅ Error page handling

## Configuration

### Environment Variables

- `E2E_SETUP_DEMO=true`: Automatically create demo cluster before tests
- `E2E_DEMO_NAME=navigator-e2e`: Name of the demo cluster to use
- `E2E_NAVCTL_COMMAND`: Custom navctl command for webServer

### Playwright Configuration

Key settings in `playwright.config.ts`:
- **Single worker**: Prevents conflicts with demo environment
- **Retry on CI**: Automatic retries for flaky tests
- **Artifacts**: Screenshots, videos, and traces on failure
- **Timeout**: 60 seconds per test, 120 seconds for server startup

## Troubleshooting

### Common Issues

1. **Demo cluster not ready**:
   ```bash
   # Check cluster status
   kind get clusters
   kubectl --kubeconfig=navigator-e2e-kubeconfig get pods -A
   
   # Test connectivity
   curl http://localhost:30080
   ```

2. **Stale navctl binary**:
   ```bash
   # Force rebuild
   make clean
   make build-navctl-dev
   ```

3. **Port conflicts**:
   ```bash
   # Check for processes using Navigator ports
   lsof -i :8082  # Navigator UI
   lsof -i :30080 # Demo gateway
   lsof -i :30090 # Prometheus
   ```

4. **Browser installation**:
   ```bash
   cd ui
   npx playwright install chromium
   ```

### Debug Mode

Run tests in headed mode with debug tools:

```bash
make test-e2e-debug
```

Or use Playwright's UI mode:

```bash
make test-e2e-ui
```

### Viewing Test Reports

After test runs:

```bash
cd ui
npm run e2e:report
```

## CI/CD Integration

The E2E tests run automatically in GitHub Actions:

- **Trigger**: PRs and pushes to main branch
- **Environment**: Ubuntu with Nix development shell
- **Artifacts**: Test results, screenshots, and HTML reports
- **Cleanup**: Automatic demo cluster teardown

### Workflow Features

- Fresh build verification
- Demo cluster setup with health checks
- Parallel test execution (when safe)
- Artifact collection on failure
- PR comments with test results

## Writing New Tests

### Best Practices

1. **Use Page Object Model**:
   ```typescript
   const navigatorPage = new NavigatorPage(page);
   await navigatorPage.goToService('frontend');
   await navigatorPage.expectServiceVisible('frontend');
   ```

2. **Wait for conditions**:
   ```typescript
   await navigatorPage.waitForServicesLoaded(3);
   await waitForServicesDiscovered(request, ['frontend', 'backend']);
   ```

3. **Validate API consistency**:
   ```typescript
   const apiServices = await getServices(request);
   // Compare with UI display
   ```

4. **Handle flaky operations**:
   ```typescript
   await retryOperation(async () => {
     await navigatorPage.refreshMetrics();
     await navigatorPage.expectMetricsVisible();
   });
   ```

### Test Structure Template

```typescript
import { test, expect } from '@playwright/test';
import { NavigatorPage } from '../fixtures/navigator-page';
import { ensureNavctlReady } from '../utils/build-helpers';
import { ensureDemoCluster } from '../utils/demo-helpers';

test.describe('Feature Name', () => {
  let navigatorPage: NavigatorPage;

  test.beforeAll(async () => {
    await ensureNavctlReady();
    const demoInfo = await ensureDemoCluster();
    // Setup assertions
  });

  test.beforeEach(async ({ page }) => {
    navigatorPage = new NavigatorPage(page);
  });

  test('should do something', async ({ page, request }) => {
    // Test implementation
  });
});
```

## Performance Considerations

- Tests build fresh binaries (adds ~1-2 minutes)
- Demo cluster creation takes ~3-5 minutes
- Individual tests typically run in 10-30 seconds
- Full test suite completes in ~10-15 minutes

## Contributing

When adding new E2E tests:

1. Follow existing patterns and page object model
2. Add appropriate wait conditions
3. Include both UI and API validation
4. Test error conditions
5. Update this README if adding new test categories

For questions or issues, refer to the main Navigator documentation or open an issue.