# Architecture

Navigator runs all components on your local machine or a dedicated server outside the cluster using the `navctl local` command.

## navctl local Architecture

![navctl Local Architecture](./diagrams/navctl-local-architecture.svg)

**Characteristics:**
- **Single process**: Manager, multiple Edge instances, and UI run within one `navctl local` process
- **Multi-cluster support**: Each edge instance manages a different Kubernetes cluster
- **External access**: Connects to Kubernetes clusters via kubeconfig
- **Simple deployment**: No in-cluster permissions or resources required
- **Dynamic cluster switching**: Can add/remove clusters by updating kubeconfig contexts
- **Development friendly**: Ideal for development, testing, and lightweight production

**Use Cases:**
- Development and testing
- CI/CD environments  
- Small to medium deployments
- Situations where in-cluster deployment is restricted
- Multi-cluster management from a central location

## Core Components

### Manager Service
- **Central coordination hub** for all edge connections
- **Bidirectional streaming** - maintains persistent gRPC connections with edges
- **State aggregation** - consolidates cluster state from multiple sources
- **Query routing** - directs proxy configuration requests to appropriate edges
- **API gateway** - serves HTTP REST API for UI and external integrations

**Deployment:**
- Embedded in navctl process on ports 8080 (gRPC) and 8081 (HTTP)

### Edge Service  
- **Cluster connector** - interfaces with Kubernetes API servers
- **State synchronization** - streams services, pods, and endpoints to manager
- **Proxy analysis** - connects to Envoy admin APIs for configuration retrieval
- **Flexible deployment** - runs externally via kubeconfig

**Deployment:**
- Multiple edge instances embedded in navctl process, each using different kubeconfig contexts

### Web UI
- **Service discovery interface** - browse and inspect Kubernetes services
- **Proxy visualization** - view Envoy configurations in structured format  
- **Real-time updates** - live data from manager API
- **Multi-cluster view** - unified interface across all connected clusters

**Deployment:**
- Embedded in navctl, served on port 3000