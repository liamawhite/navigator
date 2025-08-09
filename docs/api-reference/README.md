# API Reference

Navigator provides gRPC APIs for both frontend clients and backend edge-to-manager communication.

## Contents

- [Backend API](backend-api.md) - Manager-edge communication APIs
- [Frontend API](frontend-api.md) - Frontend service registry APIs  
- [Types](types.md) - Istio resource type definitions

## API Overview

Navigator uses Protocol Buffers for API definitions with HTTP gateway support for REST access.

### Backend APIs (`api/backend/v1alpha1/`)
- Manager-edge bidirectional streaming
- Cluster state synchronization
- Proxy information requests

### Frontend APIs (`api/frontend/v1alpha1/`)
- Service registry queries
- Real-time service discovery
- HTTP gateway with OpenAPI specs

### Type Definitions (`api/types/v1alpha1/`)
- Istio resource definitions
- Kubernetes object wrappers
- Proxy configuration types

## Code Generation

All API code is generated from Protocol Buffer definitions:

```bash
# Generate all APIs
make generate

# Individual generation
cd api && buf generate
```

Generated code is located in `pkg/api/` and should never be edited directly.