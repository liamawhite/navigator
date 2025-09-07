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

package providers

import (
	"github.com/liamawhite/navigator/manager/pkg/connections"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// ConnectionManager interface for basic connection management
type ConnectionManager interface {
	RegisterConnection(clusterID string, stream v1alpha1.ManagerService_ConnectServer) error
	UnregisterConnection(clusterID string)
	UpdateClusterState(clusterID string, clusterState *v1alpha1.ClusterState) error
	UpdateCapabilities(clusterID string, capabilities *v1alpha1.EdgeCapabilities) error
	GetClusterState(clusterID string) (*v1alpha1.ClusterState, error)
	GetAllClusterStates() map[string]*v1alpha1.ClusterState
	IsClusterConnected(clusterID string) bool
	GetActiveClusterCount() int
	SendMessageToCluster(clusterID string, message *v1alpha1.ConnectResponse) error
}

// ReadOptimizedConnectionManager extends ConnectionManager with read-optimized methods
type ReadOptimizedConnectionManager interface {
	ConnectionManager
	ListAggregatedServices(namespace, clusterID string) []*connections.AggregatedService
	GetAggregatedService(serviceID string) (*connections.AggregatedService, bool)
	GetAggregatedServiceInstance(instanceID string) (*connections.AggregatedServiceInstance, bool)
	GetConnectionInfo() map[string]connections.ConnectionInfo
}
