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

### Multiple Kubernetes Contexts

Navigator can connect to multiple Kubernetes contexts simultaneously using the `--contexts` flag with glob pattern support:

```bash
# Use current context (default behavior)
navctl local

# Connect to specific contexts
navctl local --contexts context1,context2,context3

# Use glob patterns to select multiple contexts
navctl local --contexts "*-prod"
navctl local --contexts "team-*"
navctl local --contexts "*-prod,*-staging"

# Mix exact names and patterns
navctl local --contexts "production,*-staging"

# Use custom kubeconfig with context patterns
navctl local --kube-config ~/.kube/config --contexts "*-prod"
```

### Multi-Cluster Service Discovery

When connected to multiple contexts, Navigator creates one edge service per context, all connecting to the same manager instance. This provides:

- **Aggregated Service View**: See services from all connected clusters in a single interface
- **Cross-Cluster Analysis**: Compare service configurations across environments  
- **Cluster Identification**: Each service is tagged with its originating cluster
- **Unified Proxy Analysis**: Analyze Istio configurations across your entire mesh

## Troubleshooting

### Common Issues

**Port Already in Use**
- Navigator uses ports 8080 (manager), 8081 (manager HTTP gateway), and 8082 (UI) by default
- Use `--manager-port` and `--ui-port` flags to specify different ports

**gRPC Message Size Exceeded (Large Clusters)**
- Large clusters may exceed the default gRPC message size limit
- Increase the limit with `--max-message-size` flag (e.g., `--max-message-size 16` for 16MB)

**Kubernetes Access**
- Ensure `kubectl` can access your cluster
- Verify your current context with `kubectl config current-context`

**Browser Doesn't Open**
- Use `--no-browser` flag and manually navigate to http://localhost:8082

## Next Steps

- Learn about all available commands in the [CLI Reference](../reference/cli/)
- Explore the [Developer Guide](../developer-guide/) for architecture details
- Review the [API Reference](../reference/api/) for integration options
