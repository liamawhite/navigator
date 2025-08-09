---
sidebar_position: 1
---

# Getting Started with Navigator

Navigator is a Kubernetes service discovery and proxy configuration analysis tool that helps you visualize and understand your services across multiple clusters.

## What is Navigator?

Navigator helps you understand and visualize your Kubernetes services across multiple clusters with built-in support for Istio service mesh analysis. Key features include:

- **Multi-cluster Service Discovery**: Aggregate services from multiple Kubernetes clusters
- **Proxy Configuration Analysis**: Deep inspection of Envoy/Istio proxy configurations
- **Real-time Synchronization**: Live updates of cluster state through gRPC streaming
- **Web Interface**: Modern React UI for service visualization and configuration viewing
- **Simple Setup**: Single command to start all components

## Core Components

- **Manager**: Central coordination point that aggregates cluster state from multiple edges
- **Edge**: Connects to Kubernetes clusters and streams state to the manager
- **navctl**: CLI tool that orchestrates all components
- **UI**: React web interface for service discovery and proxy analysis

## Quick Start

The fastest way to get started with Navigator:

### Prerequisites

- A running Kubernetes cluster with kubeconfig access

### Installation

Download the latest release for your platform from [GitHub Releases](https://github.com/liamawhite/navigator/releases/latest).

### Start Navigator

Once installed, start Navigator with a single command:

```bash
navctl local
```

This will:
1. Start the manager service
2. Start an edge process connected to your current kubectl context
3. Launch the web UI
4. Automatically open your browser to http://localhost:3000

### Exploring the Interface

In the web interface, you can:

- **Browse Services**: View all Kubernetes services across connected clusters
- **Inspect Instances**: See detailed information about service instances and pods
- **Analyze Proxies**: View Envoy proxy configurations for Istio-enabled services
- **Monitor Health**: Check real-time status of services and their endpoints

## Next Steps

- Check the [API Documentation](../api-docs/frontend/frontend-api) for integration details
- Read the [Development Guide](../development-docs/intro) if you want to contribute

## Need Help?

- Check our [GitHub Issues](https://github.com/liamawhite/navigator/issues) for known problems
- Start a [GitHub Discussion](https://github.com/liamawhite/navigator/discussions) for questions
- Review the [API Documentation](../api-docs/frontend/frontend-api) for integration details