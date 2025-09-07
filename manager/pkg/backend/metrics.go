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

package backend

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/liamawhite/navigator/manager/pkg/providers"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// MeshMetricsService handles service mesh metrics requests to edge clusters
type MeshMetricsService struct {
	connectionManager providers.ConnectionManager
	logger            *slog.Logger

	// Pending requests tracking
	mu              sync.RWMutex
	pendingRequests map[string]*PendingMeshMetricsRequest
}

// PendingMeshMetricsRequest tracks in-flight mesh metrics requests
type PendingMeshMetricsRequest struct {
	RequestID  string
	ClusterID  string
	CreatedAt  time.Time
	ResponseCh chan *MeshMetricsResult
	ctx        context.Context
	cancel     context.CancelFunc
}

// MeshMetricsResult contains the result of a mesh metrics request
type MeshMetricsResult struct {
	MeshMetrics *typesv1alpha1.ServiceGraphMetrics
	Error       error
}

// NewMeshMetricsService creates a new mesh metrics service
func NewMeshMetricsService(connectionManager providers.ConnectionManager, logger *slog.Logger) *MeshMetricsService {
	return &MeshMetricsService{
		connectionManager: connectionManager,
		logger:            logger,
		pendingRequests:   make(map[string]*PendingMeshMetricsRequest),
	}
}

// GetServiceGraphMetrics requests service graph metrics from a specific edge cluster
func (m *MeshMetricsService) GetServiceGraphMetrics(ctx context.Context, clusterID string, req *frontendv1alpha1.GetServiceGraphMetricsRequest) (*typesv1alpha1.ServiceGraphMetrics, error) {
	m.logger.Info("requesting mesh metrics",
		"cluster_id", clusterID,
		"filters", fmt.Sprintf("%+v", req))

	// Check if cluster is connected
	if !m.connectionManager.IsClusterConnected(clusterID) {
		return nil, fmt.Errorf("cluster not connected: %s", clusterID)
	}

	// Generate unique request ID
	requestID := uuid.New().String()

	// Create request context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)

	// Convert frontend filters to shared filters type
	var filters *typesv1alpha1.GraphMetricsFilters
	if len(req.Namespaces) > 0 || len(req.Clusters) > 0 {
		filters = &typesv1alpha1.GraphMetricsFilters{
			Namespaces: req.Namespaces,
			Clusters:   req.Clusters,
		}
	}

	pendingReq := &PendingMeshMetricsRequest{
		RequestID:  requestID,
		ClusterID:  clusterID,
		CreatedAt:  time.Now(),
		ResponseCh: make(chan *MeshMetricsResult, 1),
		ctx:        reqCtx,
		cancel:     cancel,
	}

	// Register pending request
	m.mu.Lock()
	m.pendingRequests[requestID] = pendingReq
	m.mu.Unlock()

	// Cleanup on exit
	defer func() {
		m.mu.Lock()
		delete(m.pendingRequests, requestID)
		m.mu.Unlock()
		close(pendingReq.ResponseCh)
	}()

	// Send service graph metrics request to edge
	message := &v1alpha1.ConnectResponse{
		Message: &v1alpha1.ConnectResponse_ServiceGraphMetricsRequest{
			ServiceGraphMetricsRequest: &v1alpha1.ServiceGraphMetricsRequest{
				RequestId: requestID,
				Filters:   filters,
				StartTime: req.StartTime,
				EndTime:   req.EndTime,
			},
		},
	}

	if err := m.connectionManager.SendMessageToCluster(clusterID, message); err != nil {
		return nil, fmt.Errorf("failed to send mesh metrics request: %w", err)
	}

	m.logger.Debug("mesh metrics request sent", "request_id", requestID, "cluster_id", clusterID)

	// Wait for response or timeout
	select {
	case result := <-pendingReq.ResponseCh:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.MeshMetrics, nil
	case <-reqCtx.Done():
		return nil, fmt.Errorf("mesh metrics request timed out for cluster %s", clusterID)
	}
}

// HandleServiceGraphMetricsResponse processes service graph metrics responses from edge clusters
func (m *MeshMetricsService) HandleServiceGraphMetricsResponse(response *v1alpha1.ServiceGraphMetricsResponse) {
	requestID := response.RequestId

	m.mu.RLock()
	pendingReq, exists := m.pendingRequests[requestID]
	m.mu.RUnlock()

	if !exists {
		m.logger.Warn("received response for unknown mesh metrics request", "request_id", requestID)
		return
	}

	m.logger.Debug("received mesh metrics response", "request_id", requestID, "cluster_id", pendingReq.ClusterID)

	// Prepare result
	result := &MeshMetricsResult{}

	switch res := response.Result.(type) {
	case *v1alpha1.ServiceGraphMetricsResponse_ServiceGraphMetrics:
		result.MeshMetrics = res.ServiceGraphMetrics
		m.logger.Info("service graph metrics retrieved successfully",
			"request_id", requestID,
			"cluster_id", pendingReq.ClusterID,
			"pairs_count", len(res.ServiceGraphMetrics.Pairs))
	case *v1alpha1.ServiceGraphMetricsResponse_ErrorMessage:
		result.Error = fmt.Errorf("edge error: %s", res.ErrorMessage)
		m.logger.Error("mesh metrics request failed on edge",
			"request_id", requestID,
			"cluster_id", pendingReq.ClusterID,
			"error", res.ErrorMessage)
	default:
		result.Error = fmt.Errorf("unknown mesh metrics response type: %T", res)
		m.logger.Error("unknown mesh metrics response type",
			"request_id", requestID,
			"cluster_id", pendingReq.ClusterID,
			"type", fmt.Sprintf("%T", res))
	}

	// Send result to waiting goroutine
	select {
	case pendingReq.ResponseCh <- result:
	case <-pendingReq.ctx.Done():
		// Request timed out, no one is waiting
	}
}

// GetPendingRequestCount returns the number of pending mesh metrics requests
func (m *MeshMetricsService) GetPendingRequestCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pendingRequests)
}
