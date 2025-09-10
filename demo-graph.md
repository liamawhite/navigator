# Service Graph Visualization POC

## Overview
Created a proof-of-concept service graph visualization using Reagraph WebGL library that transforms Navigator's service metrics adjacency list into an interactive network graph.

## Implementation Details

### 🎯 **Components Created**
- **ServiceGraph** component (`ui/src/components/serviceregistry/ServiceGraph.tsx`)
- **Data transformation utilities** (`ui/src/utils/graphTransform.ts`) 
- **Graph metrics hook** (leverages existing `useServiceGraphMetrics`)

### 🔗 **Integration Points**
- **API Integration**: Uses existing `metricsApi.getServiceGraphMetrics()` 
- **Theme Support**: Dynamically adapts to Navigator's dark/light/system themes
- **Routing**: Added to existing `/topology` page alongside current table view

### 📊 **Graph Features**

#### Visual Elements
- **Nodes**: Services sized by total request rate 
- **Edges**: Connections sized by request rate, colored by error rate
- **Clustering**: Services colored by cluster for easy identification
- **Labels**: Service names and request rates displayed

#### Interactive Features  
- **WebGL Rendering**: High performance for large service meshes
- **Force-Directed Layout**: Automatic positioning using physics simulation
- **Drag & Drop**: Interactive node positioning
- **Zoom & Pan**: Navigate large graphs easily
- **Node Click**: Callback for future service detail navigation

#### Data Transformation
- **Node Creation**: Extracts unique services from source/destination pairs
- **Metrics Aggregation**: Combines incoming/outgoing traffic per service
- **Edge Visualization**: Shows request rates, highlights error conditions
- **Color Coding**:
  - Healthy traffic: Green edges
  - Medium errors (5-10%): Amber edges  
  - High errors (>10%): Red edges
  - Cluster-based node colors

## Accessing the Graph

Navigate to `/topology` in Navigator UI - the graph appears above the existing service communication table.

## Next Steps
- Add time range controls for historical analysis
- Implement service filtering by namespace/cluster
- Add detailed tooltips with latency percentiles  
- Enable click-through navigation to service details
- Consider 3D layout option for complex topologies