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
	types "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// ProxyService handles proxy configuration requests to edge clusters
type ProxyService struct {
	connectionManager providers.ConnectionManager
	logger            *slog.Logger

	// Pending requests tracking
	mu              sync.RWMutex
	pendingRequests map[string]*PendingProxyRequest
}

// PendingProxyRequest tracks in-flight proxy configuration requests
type PendingProxyRequest struct {
	RequestID  string
	ClusterID  string
	Namespace  string
	PodName    string
	CreatedAt  time.Time
	ResponseCh chan *ProxyConfigResult
	ctx        context.Context
	cancel     context.CancelFunc
}

// ProxyConfigResult contains the result of a proxy configuration request
type ProxyConfigResult struct {
	ProxyConfig *types.ProxyConfig
	Error       error
}

// NewProxyService creates a new proxy service
func NewProxyService(connectionManager providers.ConnectionManager, logger *slog.Logger) *ProxyService {
	return &ProxyService{
		connectionManager: connectionManager,
		logger:            logger,
		pendingRequests:   make(map[string]*PendingProxyRequest),
	}
}

// GetProxyConfig requests proxy configuration from a specific edge cluster
func (p *ProxyService) GetProxyConfig(ctx context.Context, clusterID, namespace, podName string) (*types.ProxyConfig, error) {
	p.logger.Info("requesting proxy config",
		"cluster_id", clusterID,
		"namespace", namespace,
		"pod", podName)

	// Check if cluster is connected
	if !p.connectionManager.IsClusterConnected(clusterID) {
		return nil, fmt.Errorf("cluster %s is not connected", clusterID)
	}

	// Generate unique request ID
	requestID := uuid.New().String()

	// Create pending request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pendingReq := &PendingProxyRequest{
		RequestID:  requestID,
		ClusterID:  clusterID,
		Namespace:  namespace,
		PodName:    podName,
		CreatedAt:  time.Now(),
		ResponseCh: make(chan *ProxyConfigResult, 1),
		ctx:        reqCtx,
		cancel:     cancel,
	}

	// Register pending request
	p.mu.Lock()
	p.pendingRequests[requestID] = pendingReq
	p.mu.Unlock()

	// Cleanup on exit
	defer func() {
		p.mu.Lock()
		delete(p.pendingRequests, requestID)
		p.mu.Unlock()
		close(pendingReq.ResponseCh)
	}()

	// Send proxy config request to edge
	message := &v1alpha1.ConnectResponse{
		Message: &v1alpha1.ConnectResponse_ProxyConfigRequest{
			ProxyConfigRequest: &v1alpha1.ProxyConfigRequest{
				RequestId:    requestID,
				PodNamespace: namespace,
				PodName:      podName,
			},
		},
	}

	if err := p.connectionManager.SendMessageToCluster(clusterID, message); err != nil {
		return nil, fmt.Errorf("failed to send proxy config request: %w", err)
	}

	p.logger.Debug("proxy config request sent", "request_id", requestID, "cluster_id", clusterID)

	// Wait for response or timeout
	select {
	case result := <-pendingReq.ResponseCh:
		if result.Error != nil {
			p.logger.Error("proxy config request failed",
				"request_id", requestID,
				"cluster_id", clusterID,
				"error", result.Error)
			return nil, result.Error
		}

		p.logger.Info("proxy config request completed",
			"request_id", requestID,
			"cluster_id", clusterID,
			"version", result.ProxyConfig.Version)
		return result.ProxyConfig, nil

	case <-reqCtx.Done():
		p.logger.Error("proxy config request timed out",
			"request_id", requestID,
			"cluster_id", clusterID)
		return nil, fmt.Errorf("proxy config request timed out after 30 seconds")
	}
}

// HandleProxyConfigResponse processes proxy configuration responses from edges
func (p *ProxyService) HandleProxyConfigResponse(response *v1alpha1.ProxyConfigResponse) error {
	requestID := response.RequestId

	p.logger.Debug("received proxy config response", "request_id", requestID)

	// Find pending request
	p.mu.RLock()
	pendingReq, exists := p.pendingRequests[requestID]
	p.mu.RUnlock()

	if !exists {
		p.logger.Warn("received response for unknown request", "request_id", requestID)
		return fmt.Errorf("unknown request ID: %s", requestID)
	}

	// Check if request context is still valid
	select {
	case <-pendingReq.ctx.Done():
		p.logger.Warn("received response for expired request", "request_id", requestID)
		return fmt.Errorf("request %s has expired", requestID)
	default:
	}

	// Build result
	var result *ProxyConfigResult

	switch responseResult := response.Result.(type) {
	case *v1alpha1.ProxyConfigResponse_ProxyConfig:
		result = &ProxyConfigResult{
			ProxyConfig: responseResult.ProxyConfig,
			Error:       nil,
		}
		p.logger.Debug("proxy config response successful",
			"request_id", requestID,
			"version", responseResult.ProxyConfig.Version)

	case *v1alpha1.ProxyConfigResponse_ErrorMessage:
		result = &ProxyConfigResult{
			ProxyConfig: nil,
			Error:       fmt.Errorf("edge error: %s", responseResult.ErrorMessage),
		}
		p.logger.Error("proxy config response error",
			"request_id", requestID,
			"error", responseResult.ErrorMessage)

	default:
		result = &ProxyConfigResult{
			ProxyConfig: nil,
			Error:       fmt.Errorf("unknown response type: %T", responseResult),
		}
		p.logger.Error("unknown proxy config response type",
			"request_id", requestID,
			"type", fmt.Sprintf("%T", responseResult))
	}

	// Send result to waiting goroutine
	select {
	case pendingReq.ResponseCh <- result:
		return nil
	case <-pendingReq.ctx.Done():
		p.logger.Warn("failed to deliver response - request expired", "request_id", requestID)
		return fmt.Errorf("failed to deliver response - request %s expired", requestID)
	}
}

// GetPendingRequestCount returns the number of pending proxy requests (for monitoring)
func (p *ProxyService) GetPendingRequestCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pendingRequests)
}

// CleanupExpiredRequests removes expired requests (can be called periodically)
func (p *ProxyService) CleanupExpiredRequests() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for requestID, req := range p.pendingRequests {
		if now.Sub(req.CreatedAt) > 30*time.Second {
			p.logger.Warn("cleaning up expired request", "request_id", requestID)
			req.cancel()
			delete(p.pendingRequests, requestID)
		}
	}
}
