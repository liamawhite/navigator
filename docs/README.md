# Navigator

Navigator is a service-focused analysis tool for Kubernetes and Istio that provides service discovery and proxy configuration analysis.

## Documentation Sections

### [User Guide](user-guide/)
Get started with Navigator, including installation and usage instructions.

- [Installation](user-guide/installation.md) - How to install Navigator
- [Getting Started](user-guide/getting-started.md) - Quick start guide  

### [Developer Guide](developer-guide/)
Technical documentation for developers and contributors.

- [Architecture](developer-guide/architecture.md) - System design and components
- [Cluster State Sync](developer-guide/cluster-state-sync.md) - Edge-to-manager streaming
- [Proxy Information](developer-guide/proxy-information-retrieval.md) - On-demand analysis
- [Contributing](developer-guide/contributing.md) - How to contribute

### [Reference Documentation](reference/)
Complete API and CLI reference documentation.

- [API Reference](reference/api/) - gRPC and HTTP interfaces
- [CLI Reference](reference/cli/) - Command-line interface documentation

## Quick Links

- **Get Started**: [Installation Guide](user-guide/installation.md)
- **Development**: [CLAUDE.md](../CLAUDE.md) for complete development setup
- **Architecture**: [System Overview](developer-guide/architecture.md)
- **APIs**: [Reference Documentation](reference/)

## Project Structure

Navigator consists of three main components:
- **navctl** - CLI tool for orchestration and local development
- **manager** - Central coordination service  
- **edge** - Kubernetes cluster connector

For detailed architecture information, see the [Developer Guide](developer-guide/).
