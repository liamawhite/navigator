# Installation

This guide covers how to install Navigator on your system.

## Prerequisites

- A running Kubernetes cluster with kubeconfig access
- `kubectl` configured to access your cluster

## Download and Install

### Stable Release (Recommended)

Download the latest stable release from [GitHub Releases](https://github.com/liamawhite/navigator/releases/latest):

```bash
# Linux (x86_64)
curl -L https://github.com/liamawhite/navigator/releases/latest/download/navigator_Linux_x86_64.tar.gz | tar xz
chmod +x navctl && sudo mv navctl /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/liamawhite/navigator/releases/latest/download/navigator_Darwin_arm64.tar.gz | tar xz
chmod +x navctl && sudo mv navctl /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/liamawhite/navigator/releases/latest/download/navigator_Darwin_x86_64.tar.gz | tar xz
chmod +x navctl && sudo mv navctl /usr/local/bin/
```

### Nightly Builds (Development)

For early access to new features, use nightly builds which are automatically generated daily from the main branch:

```bash
# Linux (x86_64)
curl -L https://github.com/liamawhite/navigator/releases/download/$(curl -s https://api.github.com/repos/liamawhite/navigator/releases | jq -r '.[] | select(.prerelease == true and (.tag_name | contains("nightly"))) | .tag_name' | head -1)/navigator_Linux_x86_64.tar.gz | tar xz
chmod +x navctl && sudo mv navctl /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/liamawhite/navigator/releases/download/$(curl -s https://api.github.com/repos/liamawhite/navigator/releases | jq -r '.[] | select(.prerelease == true and (.tag_name | contains("nightly"))) | .tag_name' | head -1)/navigator_Darwin_arm64.tar.gz | tar xz
chmod +x navctl && sudo mv navctl /usr/local/bin/
```

⚠️ **Nightly Build Notice**: Nightly builds contain the latest features but may be unstable. Use stable releases for production.

## Verify Installation

Confirm Navigator is installed correctly:

```bash
navctl version
```

## Next Steps

Once installed, see the [Getting Started](getting-started.md) guide to begin using Navigator.