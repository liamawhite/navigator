# Getting Started

This guide will help you get Navigator up and running quickly.

## Quick Start

The fastest way to get started with Navigator:

### 1. Start Navigator

Once [installed](installation.md), start Navigator with a single command:

```bash
navctl local
```

This command will:
1. Start the manager service
2. Start an edge process connected to your current kubectl context
3. Launch the web UI
4. Automatically open your browser to http://localhost:3000

### 2. Explore the Interface

In the web interface, you can:

- **Browse Services**: View all Kubernetes services across connected clusters
- **Inspect Instances**: See detailed information about service instances and pods
- **Analyze Proxies**: View Envoy proxy configurations for Istio-enabled services
- **Monitor Health**: Check real-time status of services and their endpoints

### 3. Understanding the Components

Navigator consists of three main components working together:

- **Manager**: Central coordination point that aggregates cluster state from multiple edges
- **Edge**: Connects to Kubernetes clusters and streams state to the manager
- **navctl**: CLI tool that orchestrates all components
- **UI**: React web interface for service discovery and proxy analysis

## Advanced Usage

### Custom Configuration

You can customize Navigator's behavior with command-line flags:

```bash
# Start with custom ports
navctl local --manager-port 9090 --ui-port 3000 --no-browser

# Enable debug logging
navctl local --log-level debug --log-format json

# Use specific kubeconfig
navctl local --kube-config /path/to/kubeconfig
```

### Multi-Cluster Setup

For production deployments across multiple clusters, see the [CLI Reference](cli-reference.md) for advanced configuration options.

## Troubleshooting

### Common Issues

**Port Already in Use**
- Navigator uses ports 8080 (manager) and 3000 (UI) by default
- Use `--manager-port` and `--ui-port` flags to specify different ports

**Kubernetes Access**
- Ensure `kubectl` can access your cluster
- Verify your current context with `kubectl config current-context`

**Browser Doesn't Open**
- Use `--no-browser` flag and manually navigate to http://localhost:3000

## Next Steps

- Learn about all available commands in the [CLI Reference](cli-reference.md)
- Explore the [Developer Guide](../developer-guide/) for architecture details
- Review the [API Reference](../api-reference/) for integration options