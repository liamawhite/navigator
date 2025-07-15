# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

Navigator is a gRPC-based service registry that provides Kubernetes service discovery through both gRPC and HTTP APIs.

### Core Components

**API Definitions (`api/`)**
- Protocol buffer definitions with HTTP annotations
- Generates Go code, gRPC stubs, and OpenAPI specs
- Uses buf for code generation and linting

**CLI Layer (`internal/cli/`)**
- Cobra-based command structure with root command and serve subcommand
- Handles server configuration (port, kubeconfig path)
- Manages graceful shutdown on SIGINT/SIGTERM

**Datastore Pattern (`pkg/datastore/`)**
- Abstract ServiceDatastore interface
- Kubeconfig implementation uses Kubernetes client-go
- Mock implementation for testing

**gRPC Server (`internal/grpc/`)**
- Dual-protocol server: gRPC (port N) + HTTP gateway (port N+1)
- Uses grpc-gateway for automatic HTTP API generation
- Implements ServiceRegistryService with reflection enabled

**Service Registry (`internal/grpc/service_registry.go`)**
- Core business logic for service discovery
- Implements ListServices and GetService RPCs
- Delegates to datastore abstraction

### Data Flow
1. CLI starts gRPC server with kubeconfig-based datastore
2. gRPC server registers ServiceRegistryService
3. HTTP gateway proxies REST calls to gRPC
4. Service calls query Kubernetes API via client-go
5. Kubernetes Services and Endpoints are transformed to proto messages

## Development Commands

### Quick Start
```bash
# Enter development environment
nix develop

# Start full-stack development (recommended - opens browser automatically)
make dev

# Frontend only development
make dev-ui

# Backend only development  
make dev-backend
```

### Build Commands
```bash
# Build backend binary
make build

# Build frontend for production
make build-ui
```

### Code Quality Commands
```bash
# Format both Go and UI code
make format

# Lint both Go and UI code (with auto-fix)
make lint

# Run all quality checks (used in CI)
make check

# Clean up development environment
make clean
```

### Logging Configuration
```bash
# Start with debug logging in JSON format
./navigator serve --log-level debug --log-format json

# Available log levels: debug, info, warn, error
# Available formats: text, json
```

### Manual Development Setup

#### Backend Only
```bash
# Build the navigator binary
go build -o navigator cmd/navigator/main.go

# Run the server (gRPC on port 8080, HTTP gateway on port 8081)
./navigator serve -p 8080 -k ~/.kube/config

# Run with custom kubeconfig
./navigator serve --kubeconfig /path/to/kubeconfig --port 9090

# Run with hot reloading
air
```

#### UI Development (Manual)
```bash
# Frontend only (connects to existing backend)
cd ui && npm run dev

# Full stack with one command (builds and starts backend + frontend)
cd ui && npm run dev:full

# Full stack with Go hot reloading
cd ui && npm run dev:air
```

#### Environment Configuration
- `ui/.env` - Development (uses Vite proxy - recommended)
- `ui/.env.local` - Local development with Vite proxy
- `ui/.env.production` - Production environment

### Testing
```bash
# Run all tests (unit tests only)
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/grpc/
go test ./pkg/datastore/kubeconfig/

# Run single test
go test -run TestServiceRegistryServer_ListServices ./internal/grpc/

# Run unit tests only (skip integration tests)
go test -short ./...
```

### Integration Testing
Integration tests use Kind (Kubernetes in Docker) to test against a real Kubernetes cluster. These tests require Docker to be running and `kind` and `kubectl` to be installed.

```bash
# Prerequisites: Install kind and kubectl
# macOS: brew install kind kubectl
# Linux: See https://kind.sigs.k8s.io/docs/user/quick-start/

# Run integration tests (requires Docker, kind, kubectl)
go test -v ./testing/integration/local/

# Run all integration tests
go test -v -run TestLocalServiceRegistry ./testing/integration/local/

# Run specific test categories
go test -v -run TestLocalBasic ./testing/integration/local/
go test -v -run TestLocalAdvanced ./testing/integration/local/
go test -v -run TestLocalIsolation ./testing/integration/local/

# Run individual test scenarios
go test -v -run TestLocalSingle/basic_service_discovery ./testing/integration/local/
go test -v -run TestLocalSingle/microservice_topology ./testing/integration/local/
go test -v -run TestLocalSingle/mixed_service_types ./testing/integration/local/

# Run integration tests with longer timeout
go test -v -timeout 15m ./testing/integration/local/

# Run integration tests and keep clusters on failure for debugging
CLEANUP_ON_FAILURE=false go test -v ./testing/integration/local/

# Clean up any leftover test clusters
kind delete cluster --name navigator-integration-test
kind delete cluster --name navigator-basic-test
kind delete cluster --name navigator-multi-test
```

### Protocol Buffer Generation
```bash
# Generate Go code from proto files
cd api && buf generate

# Lint proto files
cd api && buf lint

# Check for breaking changes
cd api && buf breaking --against '.git#branch=main'
```

### Code Quality
```bash
# Run Go vet
go vet ./...

# Run Go fmt
go fmt ./...

# Check Go modules
go mod tidy
go mod verify
```

## Development Environment

This project uses Nix flakes for development environment management:

```bash
# Enter development shell
nix develop

# Development shell includes:
# - Go compiler
# - buf (Protocol Buffer tooling)
# - protobuf compiler
# - git
# - kind (Kubernetes in Docker)
# - kubectl (Kubernetes CLI)
# - docker (Container runtime)
# - air (Go hot reloading)
# - golangci-lint (Go linting)
```

### Air Hot Reloading
Air configuration (`.air.toml`) provides Go hot reloading:
- Builds to `./tmp/navigator`
- Serves on port 8080 with arguments: `serve -p 8080`
- Excludes UI, testing, and generated code directories
- Watches `.go`, `.tpl`, `.tmpl`, `.html` files
- Automatically restarts on file changes

## Key Directory Structure

```
cmd/navigator/           # Application entry point
internal/cli/           # CLI command implementations
internal/grpc/          # gRPC server and service logic
pkg/api/               # Generated protobuf code (do not edit)
pkg/datastore/         # Datastore interfaces and implementations
  kubeconfig/          # Kubernetes client implementation
  mock/                # Mock datastore for testing
pkg/logging/           # Structured logging with slog
  logger.go            # Centralized logger configuration
  interceptors.go      # gRPC request logging interceptors
  http.go              # HTTP middleware for request logging
  request.go           # Request context and correlation IDs
api/                   # Protocol buffer definitions
  backend/v1alpha1/    # Service registry API definitions
  buf.yaml             # Buf configuration
  buf.gen.yaml         # Code generation configuration
ui/                    # React frontend application
  src/                 # TypeScript React source code
  package.json         # UI dependencies and scripts
  vite.config.ts       # Vite dev server with proxy configuration
  .prettierrc          # Code formatting configuration
  eslint.config.js     # Linting configuration
testing/integration/   # Integration tests
  environment.go       # Abstract test environment interface
  test_cases.go        # Table-driven test case definitions
  service_registry_test.go  # Shared test suite (environment-agnostic)
  local/               # Kind-based integration tests
    kind.go            # Kind cluster management
    fixtures.go        # Kubernetes test fixtures
    environment.go     # Local environment implementation
    local_test.go      # Local-specific test runner
.github/workflows/     # GitHub Actions CI/CD
  hygiene.yml          # Code formatting and generation checks
  lint.yml             # Code quality and linting checks
```

## UI Frontend

The Navigator UI is a React application built with TypeScript and Vite, providing a web interface for service discovery.

### Features
- **Service List View**: Browse all Kubernetes services with instance counts
- **Service Detail View**: Detailed information about service instances and endpoints
- **Proxy Sidecar Detection**: Visual indication of services with Istio/Envoy sidecars
- **Real-time Updates**: Auto-refresh every 5 seconds
- **Responsive Design**: Works on mobile and desktop

### Technology Stack
- **React 19** with TypeScript
- **Vite** for fast development and builds
- **Tailwind CSS** for styling
- **shadcn/ui** for component library
- **TanStack Query** for API state management
- **React Router** for navigation
- **Lucide React** for icons

### Development Workflow
- **Hot reloading** for instant feedback
- **Vite proxy** to avoid CORS issues
- **ESLint + Prettier** for code quality
- **4-space indentation** for consistency

### shadcn/ui Components
- **NEVER edit shadcn/ui components directly** in `ui/src/components/ui/`
- Always use `npx shadcn@latest add <component> --overwrite` to install/update components
- If customization is needed, wrap components or extend them in separate files
- Keep all shadcn components in their original upstream form
- Use `components.json` for configuration, not direct component modification

## Structured Logging

Navigator uses Go's built-in `slog` package for comprehensive structured logging throughout the application.

### Features
- **Centralized Configuration**: Single logger setup with `logging.For(Component)`
- **Request Correlation**: Automatic request ID generation and propagation
- **Component Scoping**: Separate loggers for CLI, server, gRPC, HTTP, and datastore
- **Performance Tracking**: Request timing and performance metrics
- **Configurable Output**: Text or JSON format with adjustable log levels

### Log Levels
- `debug` - Detailed debugging information
- `info` - General operational information (default)
- `warn` - Warning conditions
- `error` - Error conditions

### Request Tracing
All HTTP and gRPC requests automatically receive:
- **Unique request IDs** for correlation
- **Client information** (IP addresses, user agents)
- **Request timing** and response codes
- **Error context** with full details

## Service IDs

Services are identified by `namespace:name` format (e.g., `default:nginx-service`). This convention is used throughout the API and datastore implementations.

## Testing Patterns

### Unit Tests
- Use `testify` for assertions and test structure
- Kubeconfig datastore tests use `fake.NewSimpleClientset()`
- gRPC service tests use mock datastore
- Test files follow Go conventions: `*_test.go`
- Tests cover happy path, error cases, and edge cases (namespace isolation, etc.)

### Integration Tests
- **Table-Driven Tests**: All test scenarios defined in `testing/integration/test_cases.go`
- **Shared Test Suite**: Located in `testing/integration/service_registry_test.go`
- **Environment Abstraction**: `TestEnvironment` interface allows tests to run against different backends
- **Local Implementation**: `testing/integration/local/` provides Kind-based testing
- Use Kind (Kubernetes in Docker) for real cluster testing
- Automatically create and destroy test clusters
- Test against actual Kubernetes API and Navigator gRPC server
- Use `testing.Short()` check to skip in unit test runs
- Use `ghcr.io/liamawhite/microservice:latest` image for realistic HTTP services
- Test service discovery with actual working microservices
- Microservice image provides HTTP proxy capabilities for testing service topologies
- **Extensible**: Easy to add new environments (remote clusters, different K8s distributions, etc.)
- **Maintainable**: New test scenarios can be added by simply updating the test cases table

#### Adding New Test Environments
To add a new test environment (e.g., for EKS, GKE, or other K8s distributions):

1. Create a new package under `testing/integration/` (e.g., `testing/integration/eks/`)
2. Implement the `TestEnvironment` interface
3. Create test runner files that call the shared test functions
4. The same test suite will run against your new environment

#### Adding New Test Cases
To add a new test scenario:

1. Add a new `TestCase` struct to the `GetTestCases()` function in `testing/integration/test_cases.go`
2. Define the service setup, timeout, and assertions
3. The test will automatically run against all environments
4. Use the flexible assertion system with different `AssertionType` and `TargetType` combinations

## Demo Environment Management

Navigator includes a comprehensive demo system for local development and testing:

### Demo Commands
```bash
# List available scenarios  
./navigator demo --list

# Set up basic demo environment
./navigator demo --scenario basic

# Set up microservice topology with Istio
./navigator demo --scenario istio-demo

# Clean up demo environment
./navigator demo --teardown
```

### Available Scenarios
- **basic**: Single web service for basic testing
- **microservice-topology**: Chain of three interconnected services  
- **istio-demo**: Full Istio service mesh demonstration with proxy analysis
- **mixed-service-types**: Various service configurations for comprehensive testing

### Kind Cluster Configuration
Demo environments use Kind with custom configuration supporting:
- Control plane and worker nodes
- Port mappings for ingress (80, 443)
- Istio-compatible networking setup
- Cluster name: `navigator-demo`
- Kubeconfig stored at `/tmp/navigator-demo-kubeconfig`

## Continuous Integration

Navigator uses GitHub Actions for automated quality assurance:

### Hygiene Workflow (`.github/workflows/hygiene.yml`)
- Code formatting checks (Go + TypeScript)
- Linting with golangci-lint (excludes generated code)
- Protocol buffer generation verification
- Git cleanliness validation

### Test Workflow (`.github/workflows/test.yml`)
- Unit test execution across Go packages
- Integration tests with Kind clusters
- 20-minute timeout for complex test scenarios