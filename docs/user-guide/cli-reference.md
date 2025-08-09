# CLI Reference

Complete reference for the `navctl` command-line interface.

## Global Options

These options are available for all commands:

```bash
--log-level string    Log level (debug, info, warn, error) (default "info")
--log-format string   Log format (text, json) (default "text")
--help               Show help information
--version            Show version information
```

## Commands

### `navctl local`

Start all Navigator components locally for development and testing.

#### Synopsis

```bash
navctl local [flags]
```

#### Description

The `local` command orchestrates starting all Navigator components:
- Manager service for central coordination
- Edge process connected to your Kubernetes cluster
- Web UI for service visualization
- Automatic browser launch (optional)

#### Options

```bash
--kube-config string      Path to kubeconfig file (default: ~/.kube/config)
--manager-port int        Port for manager service (default: 8080)
--ui-port int            Port for web UI (default: 3000)
--no-browser             Don't automatically open browser
--cluster-id string      Cluster identifier (default: auto-detected)
--sync-interval duration Cluster state sync interval (default: 30s)
--max-message-size int   Maximum gRPC message size in bytes (default: 4194304)
```

#### Examples

```bash
# Basic usage - start with defaults
navctl local

# Custom ports and disable browser
navctl local --manager-port 9090 --ui-port 8080 --no-browser

# Debug logging with JSON format
navctl local --log-level debug --log-format json

# Use specific kubeconfig
navctl local --kube-config /path/to/config

# Custom cluster identifier
navctl local --cluster-id production-east
```

### `navctl version`

Display version information for Navigator.

#### Synopsis

```bash
navctl version [flags]
```

#### Description

Shows detailed version information including:
- Version number
- Git commit hash
- Build date
- Go version used to build

#### Examples

```bash
navctl version
```

## Configuration

Navigator can be configured through command-line flags or environment variables.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NAVIGATOR_LOG_LEVEL` | Log level (debug, info, warn, error) | info |
| `NAVIGATOR_LOG_FORMAT` | Log format (text, json) | text |
| `NAVIGATOR_KUBECONFIG` | Path to kubeconfig file | ~/.kube/config |
| `NAVIGATOR_CLUSTER_ID` | Cluster identifier | auto-detected |

### Configuration Precedence

Configuration values are applied in the following order (highest to lowest priority):
1. Command-line flags
2. Environment variables
3. Default values

## Exit Codes

Navigator uses standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Invalid usage or configuration
- `130` - Interrupted by signal (Ctrl+C)

## Examples

### Development Workflow

```bash
# Start Navigator for local development
navctl local

# Start with debug logging
navctl local --log-level debug

# Start without opening browser
navctl local --no-browser
```

### Production Setup

For production deployments, run components separately rather than using `navctl local`:

```bash
# Run manager
navigator-manager --port 8080

# Run edge (on each cluster)
navigator-edge --manager-endpoint manager:8080 --cluster-id cluster-name
```

### Troubleshooting

```bash
# Check version information
navctl version

# Enable verbose logging
navctl local --log-level debug --log-format json
```