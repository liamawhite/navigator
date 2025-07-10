# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
```bash
# Build the navigator binary
go build -o navigator cmd/navigator/main.go

# Run the server (gRPC on port 8080, HTTP gateway on port 8081)
./navigator serve -p 8080 -k ~/.kube/config

# Run with custom kubeconfig
./navigator serve --kubeconfig /path/to/kubeconfig --port 9090
```

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

## Architecture Overview

Navigator is a gRPC-based service registry that provides Kubernetes service discovery through both gRPC and HTTP APIs.

### Core Components

**CLI Layer (`internal/cli/`)**
- Cobra-based command structure with root command and serve subcommand
- Handles server configuration (port, kubeconfig path)
- Manages graceful shutdown on SIGINT/SIGTERM

**gRPC Server (`internal/grpc/`)**
- Dual-protocol server: gRPC (port N) + HTTP gateway (port N+1)
- Uses grpc-gateway for automatic HTTP API generation
- Implements ServiceRegistryService with reflection enabled

**Service Registry (`internal/grpc/service_registry.go`)**
- Core business logic for service discovery
- Implements ListServices and GetService RPCs
- Delegates to datastore abstraction

**Datastore Pattern (`pkg/datastore/`)**
- Abstract ServiceDatastore interface
- Kubeconfig implementation uses Kubernetes client-go
- Mock implementation for testing

**API Definitions (`api/`)**
- Protocol buffer definitions with HTTP annotations
- Generates Go code, gRPC stubs, and OpenAPI specs
- Uses buf for code generation and linting

### Data Flow
1. CLI starts gRPC server with kubeconfig-based datastore
2. gRPC server registers ServiceRegistryService
3. HTTP gateway proxies REST calls to gRPC
4. Service calls query Kubernetes API via client-go
5. Kubernetes Services and Endpoints are transformed to proto messages

## Key Directory Structure

```
cmd/navigator/           # Application entry point
internal/cli/           # CLI command implementations
internal/grpc/          # gRPC server and service logic
pkg/api/               # Generated protobuf code (do not edit)
pkg/datastore/         # Datastore interfaces and implementations
  kubeconfig/          # Kubernetes client implementation
  mock/                # Mock datastore for testing
api/                   # Protocol buffer definitions
  backend/v1alpha1/    # Service registry API definitions
  buf.yaml             # Buf configuration
  buf.gen.yaml         # Code generation configuration
testing/integration/   # Integration tests
  environment.go       # Abstract test environment interface
  test_cases.go        # Table-driven test case definitions
  service_registry_test.go  # Shared test suite (environment-agnostic)
  local/               # Kind-based integration tests
    kind.go            # Kind cluster management
    fixtures.go        # Kubernetes test fixtures
    environment.go     # Local environment implementation
    local_test.go      # Local-specific test runner
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
```

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

## Service IDs

Services are identified by `namespace/name` format (e.g., `default/nginx-service`). This convention is used throughout the API and datastore implementations.