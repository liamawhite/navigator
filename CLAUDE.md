# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

Navigator is an edge computing platform that provides Kubernetes service discovery and proxy configuration analysis through a distributed architecture of manager, edge, and CLI components.

### Core Components

**Manager (`manager/`)**
- Central coordination point for edge connections
- Maintains bidirectional streaming gRPC connections with edges
- Aggregates high-level state from multiple clusters
- Handles frontend API requests through gRPC and HTTP gateway
- Serves on port N (gRPC) and N+1 (HTTP gateway)

**Edge (`edge/`)**
- Connects to Kubernetes API servers and Envoy proxies
- Streams cluster state (services, pods, endpoints) to manager
- Provides on-demand proxy configuration analysis via Envoy admin API
- Can run in-cluster or externally with kubeconfig access
- Handles proxy information retrieval for detailed sidecar analysis

**Navctl (`navctl/`)**
- CLI tool for local development and orchestration
- `navctl local` command starts manager + edge + UI in coordinated fashion
- Automatically opens browser and manages graceful shutdown
- Supports single kubeconfig operation for development scenarios

**API Definitions (`api/`)**
- Frontend APIs in `api/frontend/v1alpha1/` for service registry
- Backend APIs in `api/backend/v1alpha1/` for manager-edge communication
- Protocol buffer definitions with HTTP annotations for REST gateway
- Uses buf for code generation and linting

**UI (`ui/`)**
- React TypeScript application for service visualization
- Service list and detail views with proxy sidecar detection
- Real-time updates via TanStack Query
- Embedded into navctl binary for local development

### Data Flow
1. Edge connects to manager via bidirectional gRPC stream
2. Edge syncs cluster state (services, pods, endpoints) to manager
3. Manager aggregates state from multiple edges and serves frontend APIs
4. UI queries manager's HTTP gateway for service information
5. On-demand proxy analysis: Manager requests detailed config from specific edges

## Development Commands

### Quick Start
```bash
# Enter development environment
nix develop

# Build and start all services locally (one command)
make local

# Or manually build then run
make build
./bin/navctl local
```

### Build Commands
```bash
# Build all binaries and UI assets
make build

# Build individual components
make build-manager    # Build manager binary
make build-edge      # Build edge binary  
make build-navctl    # Build navctl binary (includes embedded UI)
make build-ui        # Build UI assets only
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
# Start all services with debug logging
./bin/navctl local --log-level debug --log-format json

# Individual components with custom logging  
./bin/manager --log-level debug --log-format json
./bin/edge --log-level info --log-format text

# Available log levels: debug, info, warn, error
# Available formats: text, json
```

### Manual Development Setup

#### Local Development with Navctl (Recommended)
```bash
# Build navctl binary first
make build-navctl

# Start all services locally
./bin/navctl local --kube-config ~/.kube/config

# Custom ports and options
./bin/navctl local --manager-port 9090 --ui-port 3000 --no-browser

# With debug logging
./bin/navctl local --log-level debug --log-format json
```

#### Individual Component Development
```bash
# Manager only
make build-manager
./bin/manager --port 8080

# Edge only (requires running manager)
make build-edge  
./bin/edge --manager-endpoint localhost:8080 --kubeconfig ~/.kube/config

# UI development only
cd ui && npm run dev  # Uses Vite proxy to localhost:8081
```

#### UI Development
```bash
# Frontend only (connects to existing backend via proxy)
cd ui && npm run dev

# Full stack with backend build
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
# Run unit tests (default make target)
make test-unit

# Run tests with verbose output
go test -v ./...

# Run tests for specific packages
go test ./manager/pkg/...
go test ./edge/pkg/...
go test ./pkg/...

# Run tests with build tags
go test -tags=test -v ./...

# Run unit tests only (skip integration tests)
go test -short ./...
```


### Protocol Buffer Generation
```bash
# Generate all protobuf code (included in make generate)
make generate

# Manual generation
cd api && buf generate  # Backend APIs
cd api && buf generate --template buf.gen.frontend.yaml  # Frontend APIs

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


## Key Directory Structure

```
manager/               # Manager service (central coordination)
  main.go             # Manager entry point
  pkg/config/         # Manager configuration
  pkg/connections/    # Connection management for edges
  pkg/service/        # Manager service implementation
edge/                 # Edge service (cluster connector)
  main.go            # Edge entry point  
  pkg/config/        # Edge configuration
  pkg/kubernetes/    # Kubernetes client wrapper
  pkg/proxy/         # Proxy configuration analysis
  pkg/service/       # Edge service implementation
navctl/               # CLI tool for orchestration
  main.go           # CLI entry point
  cmd/              # Cobra command implementations
  pkg/ui/           # Embedded UI server
api/                  # Protocol buffer definitions
  backend/v1alpha1/   # Manager-edge communication APIs
  frontend/v1alpha1/  # Frontend service registry APIs
  buf.yaml           # Buf configuration
  buf.gen.yaml       # Backend code generation
  buf.gen.frontend.yaml # Frontend code generation
pkg/api/             # Generated protobuf code (do not edit)
  backend/           # Generated backend APIs
  frontend/          # Generated frontend APIs
pkg/envoy/           # Envoy proxy analysis utilities
  admin/             # Envoy admin API clients
  configdump/        # Configuration dump parsing
pkg/logging/         # Structured logging with slog
  logger.go         # Centralized logger configuration
  interceptors.go   # gRPC request logging interceptors
  http.go           # HTTP middleware for request logging
  request.go        # Request context and correlation IDs
ui/                  # React frontend application
  src/              # TypeScript React source code
  components/ui/    # shadcn/ui components (do not edit directly)
  package.json      # UI dependencies and scripts
  vite.config.ts    # Vite dev server with proxy configuration
docs/               # Documentation (Docusaurus)
  development-docs/ # Architecture and development guides
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

## Testing Patterns

### Unit Tests
- Use `testify` for assertions and test structure
- Manager tests use connection mocks
- Edge tests use fake Kubernetes clients with `fake.NewSimpleClientset()`
- Test files follow Go conventions: `*_test.go`
- Tests use build tags: `go test -tags=test`

### Component Architecture
- **Manager**: Handles edge connections and aggregates cluster state
- **Edge**: Streams Kubernetes state to manager, provides proxy analysis
- **Navctl**: Orchestrates local development environments
- **UI**: Embedded React application served by navctl

### Development Patterns

**Code Generation**
- All protobuf code is generated, never edit files in `pkg/api/`
- Frontend APIs generate both Go code and OpenAPI specs for TypeScript client generation
- Use `make generate` to regenerate all protobuf and OpenAPI code

**Configuration Management**
- Each component (manager, edge, navctl) has its own config package
- Flag parsing is centralized in config packages
- Environment-specific configuration through command-line flags

**Logging Architecture**
- Structured logging with Go's `slog` package throughout all components
- Component-scoped loggers via `logging.For("component-name")`
- Request correlation IDs for distributed tracing
- Configurable log levels and formats (text/json)

**Error Handling**
- Standard Go error patterns with context propagation
- gRPC status codes for API errors
- Graceful shutdown handling in all services

### Proxy Configuration Analysis
Navigator includes comprehensive Envoy proxy analysis capabilities:
- **Admin API Access**: Multiple client types (kubectl port-forward, pilot-agent)
- **Configuration Parsing**: Bootstrap, clusters, listeners, routes, endpoints
- **Sidecar Detection**: Automatic Istio/Envoy proxy identification
- **Config Dump Processing**: Structured parsing of Envoy configuration dumps

## Continuous Integration

Navigator uses GitHub Actions for automated quality assurance:

### Hygiene Workflow (`.github/workflows/hygiene.yml`)
- **Format Check**: Runs `make format` to verify Go and TypeScript formatting
- **Lint Check**: Runs `make lint` with golangci-lint and ESLint
- **Generation Check**: Runs `make generate` to verify protobuf generation is up-to-date
- **Git Cleanliness**: Ensures no untracked changes after generation/formatting

### Test Workflow (`.github/workflows/test.yml`)
- **Unit Tests**: Runs `make test-unit` with build tags for all Go packages
- **Release Build Test**: Tests GoReleaser snapshot builds
- Uses Nix development environment for consistent tooling