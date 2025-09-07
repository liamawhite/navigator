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

package frontend

import (
	"context"
	"log/slog"
	"time"

	"github.com/liamawhite/navigator/manager/pkg/connections"
	"github.com/liamawhite/navigator/manager/pkg/providers"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
)

// ClusterRegistryService implements the frontend ClusterRegistryService
type ClusterRegistryService struct {
	frontendv1alpha1.UnimplementedClusterRegistryServiceServer
	connectionManager providers.ReadOptimizedConnectionManager
	logger            *slog.Logger
}

// NewClusterRegistryService creates a new cluster registry service
func NewClusterRegistryService(connectionManager providers.ReadOptimizedConnectionManager, logger *slog.Logger) *ClusterRegistryService {
	return &ClusterRegistryService{
		connectionManager: connectionManager,
		logger:            logger,
	}
}

// ListClusters returns sync state information for all connected clusters
func (c *ClusterRegistryService) ListClusters(ctx context.Context, req *frontendv1alpha1.ListClustersRequest) (*frontendv1alpha1.ListClustersResponse, error) {
	c.logger.Debug("listing clusters")

	connectionInfos := c.connectionManager.GetConnectionInfo()
	clusters := make([]*frontendv1alpha1.ClusterSyncInfo, 0, len(connectionInfos))

	for _, connInfo := range connectionInfos {
		cluster := convertConnectionInfoToClusterSyncInfo(connInfo)
		clusters = append(clusters, cluster)
	}

	c.logger.Debug("listed clusters", "count", len(clusters))

	return &frontendv1alpha1.ListClustersResponse{
		Clusters: clusters,
	}, nil
}

// convertConnectionInfoToClusterSyncInfo converts a ConnectionInfo to the frontend API format
func convertConnectionInfoToClusterSyncInfo(connInfo connections.ConnectionInfo) *frontendv1alpha1.ClusterSyncInfo {
	// Safe conversion from int to int32 to avoid overflow
	var serviceCount int32
	if connInfo.ServiceCount > 2147483647 { // max int32
		serviceCount = 2147483647
	} else if connInfo.ServiceCount < -2147483648 { // min int32
		serviceCount = -2147483648
	} else {
		serviceCount = int32(connInfo.ServiceCount) // #nosec G115 - bounds checked above
	}

	return &frontendv1alpha1.ClusterSyncInfo{
		ClusterId:      connInfo.ClusterID,
		ConnectedAt:    connInfo.ConnectedAt.Format(time.RFC3339),
		LastUpdate:     connInfo.LastUpdate.Format(time.RFC3339),
		ServiceCount:   serviceCount,
		SyncStatus:     computeSyncStatus(connInfo),
		MetricsEnabled: connInfo.MetricsEnabled,
	}
}

// computeSyncStatus determines the sync health based on connection info
func computeSyncStatus(connInfo connections.ConnectionInfo) frontendv1alpha1.SyncStatus {
	// If no state has been received yet, connection is initializing
	if !connInfo.StateReceived {
		return frontendv1alpha1.SyncStatus_SYNC_STATUS_INITIALIZING
	}

	timeSince := time.Since(connInfo.LastUpdate)

	switch {
	case timeSince < 30*time.Second:
		return frontendv1alpha1.SyncStatus_SYNC_STATUS_HEALTHY
	case timeSince < 5*time.Minute:
		return frontendv1alpha1.SyncStatus_SYNC_STATUS_STALE
	default:
		return frontendv1alpha1.SyncStatus_SYNC_STATUS_DISCONNECTED
	}
}
