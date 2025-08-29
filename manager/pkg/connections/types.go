// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connections

import (
	"time"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// Connection represents an active connection from an edge process
type Connection struct {
	ClusterID    string
	ConnectedAt  time.Time
	LastUpdate   time.Time
	Stream       v1alpha1.ManagerService_ConnectServer
	ClusterState *v1alpha1.ClusterState
}

// AggregatedService represents a service consolidated across multiple clusters
type AggregatedService struct {
	ID          string // namespace:service-name
	Name        string
	Namespace   string
	Instances   []*AggregatedServiceInstance            // All instances across clusters
	ClusterMap  map[string][]*AggregatedServiceInstance // cluster_id -> instances
	ClusterIPs  map[string]string                       // cluster_id -> cluster IP
	ExternalIPs map[string]string                       // cluster_id -> external IP
}

// Container represents a container running in a pod
type Container struct {
	Name         string
	Image        string
	Status       string
	Ready        bool
	RestartCount int32
}

// AggregatedServiceInstance represents a service instance with cluster context
type AggregatedServiceInstance struct {
	InstanceID     string // cluster_id:namespace:pod_name
	IP             string
	PodName        string
	Namespace      string
	ClusterName    string
	EnvoyPresent   bool
	Containers     []Container
	PodStatus      string
	NodeName       string
	CreatedAt      string
	Labels         map[string]string
	Annotations    map[string]string
	IsEnvoyPresent bool
}

// ReadOptimizedIndexes contains read-optimized data structures
type ReadOptimizedIndexes struct {
	Services            map[string]*AggregatedService           // service_id -> aggregated service
	ServicesByNamespace map[string][]*AggregatedService         // namespace -> services
	ServicesByCluster   map[string][]*AggregatedService         // cluster_id -> services
	Instances           map[string]*AggregatedServiceInstance   // instance_id -> instance
	InstancesByService  map[string][]*AggregatedServiceInstance // service_id -> instances
}

// ConnectionInfo provides information about an active connection
type ConnectionInfo struct {
	ClusterID     string
	ConnectedAt   time.Time
	LastUpdate    time.Time
	ServiceCount  int
	StateReceived bool // Whether the connection has received a full cluster state
}
