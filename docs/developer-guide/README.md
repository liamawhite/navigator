# Developer Guide

Welcome to the Navigator development documentation. This section is designed for developers, contributors, and coding agents who need technical details about Navigator's architecture, APIs, and implementation.

## Contents

- [Introduction](intro.md) - Development documentation overview
- [Architecture](architecture.md) - High-level system design and components
- [Cluster State Synchronization](cluster-state-sync.md) - Edge-to-manager state streaming
- [Proxy Information Retrieval](proxy-information-retrieval.md) - On-demand proxy analysis
- [Contributing](contributing.md) - How to contribute to Navigator
- [Diagrams](diagrams/) - Architecture diagrams and visual references

## Quick Links

- [Architecture Overview](architecture.md)
- [API Reference](../reference/api/)
- [User Guide](../user-guide/) - For user-focused documentation

## Development Setup

The only requirement to get started is to install `nix`. Nix will automatically handle all dependencies required for developement including:
- Go
- buf (Protocol Buffer tooling)
- protobuf compiler
- kind, kubectl, docker
- golangci-lint, gosec

> Note: Nix does not actually setup these tools in your default shell, instead it sets up a temporary shell enviroment with the correct binaries in the path using symlinks. All of this is handled transparently via the Makefile.

```bash
curl -fsSL https://install.determinate.systems/nix | sh -s -- install
# select **n** (no) when prompted to use the OSS installation.
   ```

### Quick Commands

```bash
# Build all components
make build

# Run quality checks
make check

# Generate all documentation and code
make generate
```

See [CLAUDE.md](../../CLAUDE.md) for detailed development setup instructions including:

- Building and running locally
- Code quality tools
- Testing patterns
- Protocol buffer generation
