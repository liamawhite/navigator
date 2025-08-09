# Navigator Documentation

Welcome to the Navigator documentation. Navigator is a service-focused analysis tool for Kubernetes and Istio that provides service discovery and proxy configuration analysis.

## Documentation Sections

### [User Guide](user-guide/)
Get started with Navigator, including installation and usage instructions.

- [Installation](user-guide/installation.md) - How to install Navigator
- [Getting Started](user-guide/getting-started.md) - Quick start guide  
- [CLI Reference](user-guide/cli-reference.md) - Complete navctl command reference

### [Developer Guide](developer-guide/)
Technical documentation for developers and contributors.

- [Architecture](developer-guide/architecture.md) - System design and components
- [Cluster State Sync](developer-guide/cluster-state-sync.md) - Edge-to-manager streaming
- [Proxy Information](developer-guide/proxy-information-retrieval.md) - On-demand analysis
- [Contributing](developer-guide/contributing.md) - How to contribute

### [API Reference](api-reference/)
Complete API documentation for gRPC and HTTP interfaces.

- [Backend API](api-reference/backend-api.md) - Manager-edge communication
- [Frontend API](api-reference/frontend-api.md) - Service registry APIs
- [Types](api-reference/types.md) - Istio resource definitions

## Quick Links

- **Get Started**: [Installation Guide](user-guide/installation.md)
- **Development**: [CLAUDE.md](../CLAUDE.md) for complete development setup
- **Architecture**: [System Overview](developer-guide/architecture.md)
- **APIs**: [Reference Documentation](api-reference/)

## Project Structure

Navigator consists of three main components:
- **navctl** - CLI tool for orchestration and local development
- **manager** - Central coordination service  
- **edge** - Kubernetes cluster connector

For detailed architecture information, see the [Developer Guide](developer-guide/).