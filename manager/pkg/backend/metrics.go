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
	backendv1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// MeshMetricsService handles service mesh metrics requests to edge clusters
type MeshMetricsService struct {
	connectionManager providers.ConnectionManager
	logger            *slog.Logger

	// Pending requests tracking
	mu                                sync.RWMutex
	pendingServiceConnectionsRequests map[string]*PendingServiceConnectionsRequest
}

// PendingServiceConnectionsRequest tracks in-flight service connections requests
type PendingServiceConnectionsRequest struct {
	RequestID  string
	ClusterID  string
	CreatedAt  time.Time
	ResponseCh chan *ServiceConnectionsResult
}

// ServiceConnectionsResult contains the result of a service connections request
type ServiceConnectionsResult struct {
	ServiceConnections *typesv1alpha1.ServiceGraphMetrics
	Error              error
}

// NewMeshMetricsService creates a new mesh metrics service
func NewMeshMetricsService(connectionManager providers.ConnectionManager, logger *slog.Logger) *MeshMetricsService {
	return &MeshMetricsService{
		connectionManager:                 connectionManager,
		logger:                            logger,
		pendingServiceConnectionsRequests: make(map[string]*PendingServiceConnectionsRequest),
	}
}

// GetServiceConnections requests service connections metrics from a specific edge cluster
func (m *MeshMetricsService) GetServiceConnections(ctx context.Context, clusterID string, req *frontendv1alpha1.GetServiceConnectionsRequest, proxyMode typesv1alpha1.ProxyMode) (*typesv1alpha1.ServiceGraphMetrics, error) {
	m.logger.Info("requesting service connections from edge cluster",
		"cluster_id", clusterID,
		"service_name", req.ServiceName,
		"namespace", req.Namespace)

	// Generate unique request ID
	requestID := uuid.New().String()

	// Create response channel
	responseCh := make(chan *ServiceConnectionsResult, 1)

	// Track pending request
	pendingRequest := &PendingServiceConnectionsRequest{
		RequestID:  requestID,
		ClusterID:  clusterID,
		CreatedAt:  time.Now(),
		ResponseCh: responseCh,
	}

	m.mu.Lock()
	m.pendingServiceConnectionsRequests[requestID] = pendingRequest
	m.mu.Unlock()

	// Clean up request when done
	defer func() {
		m.mu.Lock()
		delete(m.pendingServiceConnectionsRequests, requestID)
		m.mu.Unlock()
	}()

	// Use timestamps from request directly
	startTime := req.StartTime
	endTime := req.EndTime

	// Create service connections request for edge
	serviceConnectionsReq := &backendv1alpha1.ServiceConnectionsRequest{
		RequestId:   requestID,
		ServiceName: req.ServiceName,
		Namespace:   req.Namespace,
		StartTime:   startTime,
		EndTime:     endTime,
		ProxyMode:   proxyMode,
	}

	// Send request to edge cluster
	if err := m.connectionManager.SendMessageToCluster(clusterID, &backendv1alpha1.ConnectResponse{
		Message: &backendv1alpha1.ConnectResponse_ServiceConnectionsRequest{
			ServiceConnectionsRequest: serviceConnectionsReq,
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to send service connections request to cluster %s: %w", clusterID, err)
	}

	// Wait for response with timeout
	select {
	case result := <-responseCh:
		if result.Error != nil {
			return nil, result.Error
		}
		return result.ServiceConnections, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for service connections response from cluster %s", clusterID)
	}
}

// HandleServiceConnectionsResponse processes a service connections response from an edge cluster
func (m *MeshMetricsService) HandleServiceConnectionsResponse(resp *backendv1alpha1.ServiceConnectionsResponse) {
	m.mu.Lock()
	pendingRequest, exists := m.pendingServiceConnectionsRequests[resp.RequestId]
	m.mu.Unlock()

	if !exists {
		m.logger.Warn("received service connections response for unknown request", "request_id", resp.RequestId)
		return
	}

	result := &ServiceConnectionsResult{}

	switch r := resp.Result.(type) {
	case *backendv1alpha1.ServiceConnectionsResponse_ServiceConnections:
		result.ServiceConnections = r.ServiceConnections
		m.logger.Info("received service connections from edge cluster",
			"cluster_id", pendingRequest.ClusterID,
			"request_id", resp.RequestId)
	case *backendv1alpha1.ServiceConnectionsResponse_ErrorMessage:
		result.Error = fmt.Errorf("edge error: %s", r.ErrorMessage)
		m.logger.Error("received service connections error from edge cluster",
			"cluster_id", pendingRequest.ClusterID,
			"request_id", resp.RequestId,
			"error", r.ErrorMessage)
	default:
		result.Error = fmt.Errorf("unknown service connections response type")
		m.logger.Error("received unknown service connections response type",
			"cluster_id", pendingRequest.ClusterID,
			"request_id", resp.RequestId)
	}

	// Send result to waiting goroutine
	select {
	case pendingRequest.ResponseCh <- result:
	default:
		m.logger.Warn("failed to send service connections response - channel full or closed", "request_id", resp.RequestId)
	}
}

// GetPendingServiceConnectionsRequestCount returns the number of pending service connections requests
func (m *MeshMetricsService) GetPendingServiceConnectionsRequestCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pendingServiceConnectionsRequests)
}
