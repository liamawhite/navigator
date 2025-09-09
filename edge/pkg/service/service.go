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

package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/liamawhite/navigator/edge/pkg/interfaces"
	"github.com/liamawhite/navigator/edge/pkg/metrics"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KubernetesClient interface for dependency injection
type KubernetesClient interface {
	GetClusterState(ctx context.Context) (*v1alpha1.ClusterState, error)
	GetClusterStateWithMetrics(ctx context.Context, metricsProvider interfaces.MetricsProvider) (*v1alpha1.ClusterState, error)
	GetClusterName(ctx context.Context) (string, error)
}

// ProxyService interface for dependency injection
type ProxyService interface {
	GetProxyConfig(ctx context.Context, namespace, podName string) (*types.ProxyConfig, error)
	ValidateProxyAccess(ctx context.Context, namespace, podName string) error
}

// Config interface for dependency injection
type Config interface {
	GetClusterID() string
	GetManagerEndpoint() string
	GetSyncInterval() int
	GetMaxMessageSize() int
	GetMetricsConfig() metrics.Config
	Validate() error
}

// EdgeService manages the connection to the manager and handles cluster state synchronization
type EdgeService struct {
	config          Config
	k8sClient       KubernetesClient
	proxyService    ProxyService
	metricsProvider interfaces.MetricsProvider
	logger          *slog.Logger
	client          v1alpha1.ManagerServiceClient
	conn            *grpc.ClientConn
	stream          v1alpha1.ManagerService_ConnectClient
	connected       bool
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// NewEdgeService creates a new edge service
func NewEdgeService(config Config, k8sClient KubernetesClient, proxyService ProxyService, metricsProvider interfaces.MetricsProvider, logger *slog.Logger) (*EdgeService, error) {
	// Validate configuration first
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid edge configuration: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &EdgeService{
		config:          config,
		k8sClient:       k8sClient,
		proxyService:    proxyService,
		metricsProvider: metricsProvider,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start starts the edge service and begins cluster state synchronization
func (e *EdgeService) Start() error {
	e.logger.Info("starting edge service", "cluster_id", e.config.GetClusterID(), "manager_endpoint", e.config.GetManagerEndpoint())

	// Connect to manager
	if err := e.connect(); err != nil {
		return fmt.Errorf("failed to connect to manager: %w", err)
	}

	// Start the sync loop
	e.wg.Add(1)
	go e.syncLoop()

	// Start the message handling loop for incoming proxy config requests
	e.wg.Add(1)
	go e.handleIncomingMessages()

	return nil
}

// Stop gracefully stops the edge service
func (e *EdgeService) Stop() error {
	e.logger.Info("stopping edge service")

	// Cancel context to stop all operations
	e.cancel()

	// Wait for goroutines to finish
	e.wg.Wait()

	// Close metrics provider
	if e.metricsProvider != nil {
		if err := e.metricsProvider.Close(); err != nil {
			e.logger.Error("failed to close metrics provider", "error", err)
		}
	}

	// Close connection
	if e.conn != nil {
		return e.conn.Close()
	}

	return nil
}

// connect establishes a connection to the manager
func (e *EdgeService) connect() error {
	e.logger.Info("connecting to manager", "endpoint", e.config.GetManagerEndpoint())

	// Create gRPC connection with message size limits
	maxMessageSize := e.config.GetMaxMessageSize()
	conn, err := grpc.NewClient(e.config.GetManagerEndpoint(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create grpc connection: %w", err)
	}

	e.conn = conn
	e.client = v1alpha1.NewManagerServiceClient(conn)

	// Create streaming connection
	stream, err := e.client.Connect(e.ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	e.stream = stream

	// Send cluster identification
	if err := e.sendClusterIdentification(); err != nil {
		return fmt.Errorf("failed to send cluster identification: %w", err)
	}

	// Wait for connection acknowledgment
	if err := e.waitForConnectionAck(); err != nil {
		return fmt.Errorf("failed to get connection acknowledgment: %w", err)
	}

	e.mu.Lock()
	e.connected = true
	e.mu.Unlock()

	e.logger.Info("successfully connected to manager")

	return nil
}

// sendClusterIdentification sends the cluster identification to the manager
func (e *EdgeService) sendClusterIdentification() error {
	req := &v1alpha1.ConnectRequest{
		Message: &v1alpha1.ConnectRequest_ClusterIdentification{
			ClusterIdentification: &v1alpha1.ClusterIdentification{
				ClusterId: e.config.GetClusterID(),
				Capabilities: &v1alpha1.EdgeCapabilities{
					MetricsEnabled: e.metricsProvider != nil && e.metricsProvider.GetProviderInfo().Type != metrics.ProviderTypeNone,
				},
			},
		},
	}

	return e.stream.Send(req)
}

// waitForConnectionAck waits for the connection acknowledgment from the manager
func (e *EdgeService) waitForConnectionAck() error {
	resp, err := e.stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	switch msg := resp.Message.(type) {
	case *v1alpha1.ConnectResponse_ConnectionAck:
		if !msg.ConnectionAck.Accepted {
			return fmt.Errorf("connection rejected by manager")
		}
		e.logger.Info("connection accepted by manager")
		return nil
	case *v1alpha1.ConnectResponse_Error:
		return fmt.Errorf("connection error: %s", msg.Error.ErrorMessage)
	default:
		return fmt.Errorf("unexpected response type: %T", msg)
	}
}

// syncLoop periodically syncs cluster state with the manager
func (e *EdgeService) syncLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Duration(e.config.GetSyncInterval()) * time.Second)
	defer ticker.Stop()

	// Perform initial sync
	if err := e.syncClusterState(); err != nil {
		e.logger.Error("failed to sync cluster state", "error", err)
	}

	for {
		select {
		case <-e.ctx.Done():
			e.logger.Info("sync loop stopped")
			return
		case <-ticker.C:
			if err := e.syncClusterState(); err != nil {
				e.logger.Error("failed to sync cluster state", "error", err)

				// Try to reconnect if we lost connection
				if e.shouldReconnect(err) {
					e.logger.Info("attempting to reconnect")
					if err := e.reconnect(); err != nil {
						e.logger.Error("failed to reconnect", "error", err)
					}
				}
			}
		}
	}
}

// syncClusterState gets the current cluster state and sends it to the manager
func (e *EdgeService) syncClusterState() error {
	e.mu.RLock()
	connected := e.connected
	e.mu.RUnlock()

	if !connected {
		return fmt.Errorf("not connected to manager")
	}

	// Get cluster state from Kubernetes with metrics
	clusterState, err := e.k8sClient.GetClusterStateWithMetrics(e.ctx, e.metricsProvider)
	if err != nil {
		return fmt.Errorf("failed to get cluster state: %w", err)
	}

	// Send cluster state to manager
	req := &v1alpha1.ConnectRequest{
		Message: &v1alpha1.ConnectRequest_ClusterState{
			ClusterState: clusterState,
		},
	}

	if err := e.stream.Send(req); err != nil {
		return fmt.Errorf("failed to send cluster state: %w", err)
	}

	e.logger.Debug("sent cluster state", "services", len(clusterState.Services))

	return nil
}

// shouldReconnect determines if we should attempt to reconnect based on the error
func (e *EdgeService) shouldReconnect(err error) bool {
	if err == nil {
		return false
	}

	// Check for connection-related errors
	if err == io.EOF {
		return true
	}

	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
			return true
		}
	}

	return false
}

// reconnect attempts to reconnect to the manager
func (e *EdgeService) reconnect() error {
	e.mu.Lock()
	e.connected = false
	e.mu.Unlock()

	// Close existing connection
	if e.conn != nil {
		_ = e.conn.Close()
	}

	// Exponential backoff for reconnection
	maxAttempts := 5
	backoff := time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-e.ctx.Done():
			return fmt.Errorf("context canceled")
		default:
		}

		e.logger.Info("reconnection attempt", "attempt", attempt, "max_attempts", maxAttempts)

		if err := e.connect(); err != nil {
			e.logger.Error("reconnection failed", "attempt", attempt, "error", err)

			if attempt < maxAttempts {
				select {
				case <-e.ctx.Done():
					return fmt.Errorf("context canceled")
				case <-time.After(backoff):
					backoff *= 2 // Exponential backoff
				}
			}
		} else {
			e.logger.Info("reconnection successful")
			return nil
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts", maxAttempts)
}

// handleIncomingMessages continuously listens for incoming messages from the manager
// This includes proxy configuration requests that need to be processed
func (e *EdgeService) handleIncomingMessages() {
	defer e.wg.Done()

	e.logger.Info("starting message handling loop")

	for {
		select {
		case <-e.ctx.Done():
			e.logger.Info("message handling loop stopped")
			return
		default:
			// Check if we're connected
			e.mu.RLock()
			connected := e.connected
			stream := e.stream
			e.mu.RUnlock()

			if !connected || stream == nil {
				// Not connected, sleep and retry
				select {
				case <-e.ctx.Done():
					return
				case <-time.After(time.Second):
					continue
				}
			}

			// Receive message from manager
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					e.logger.Info("manager closed connection")
					return
				}

				e.logger.Error("failed to receive message from manager", "error", err)

				// Mark as disconnected and try to reconnect
				e.mu.Lock()
				e.connected = false
				e.mu.Unlock()

				if e.shouldReconnect(err) {
					e.logger.Info("attempting to reconnect after message receive error")
					if err := e.reconnect(); err != nil {
						e.logger.Error("failed to reconnect", "error", err)
					}
				}
				continue
			}

			// Process the message
			if err := e.processIncomingMessage(resp); err != nil {
				e.logger.Error("failed to process incoming message", "error", err)
			}
		}
	}
}

// processIncomingMessage processes different types of messages from the manager
func (e *EdgeService) processIncomingMessage(resp *v1alpha1.ConnectResponse) error {
	switch msg := resp.Message.(type) {
	case *v1alpha1.ConnectResponse_ProxyConfigRequest:
		return e.processProxyConfigRequest(msg.ProxyConfigRequest)
	case *v1alpha1.ConnectResponse_ServiceGraphMetricsRequest:
		return e.processServiceGraphMetricsRequest(msg.ServiceGraphMetricsRequest)
	case *v1alpha1.ConnectResponse_Error:
		e.logger.Error("received error from manager", "error_code", msg.Error.ErrorCode, "error_message", msg.Error.ErrorMessage)
		return fmt.Errorf("manager error: %s", msg.Error.ErrorMessage)
	default:
		e.logger.Debug("received unknown message type", "type", fmt.Sprintf("%T", msg))
		return nil
	}
}

// processProxyConfigRequest handles proxy configuration requests from the manager
func (e *EdgeService) processProxyConfigRequest(req *v1alpha1.ProxyConfigRequest) error {
	e.logger.Info("processing proxy config request",
		"request_id", req.RequestId,
		"namespace", req.PodNamespace,
		"pod", req.PodName)

	// Create response message
	resp := &v1alpha1.ConnectRequest{
		Message: &v1alpha1.ConnectRequest_ProxyConfigResponse{
			ProxyConfigResponse: &v1alpha1.ProxyConfigResponse{
				RequestId: req.RequestId,
			},
		},
	}

	// Get proxy configuration
	proxyConfig, err := e.proxyService.GetProxyConfig(e.ctx, req.PodNamespace, req.PodName)
	if err != nil {
		e.logger.Error("failed to get proxy config",
			"request_id", req.RequestId,
			"namespace", req.PodNamespace,
			"pod", req.PodName,
			"error", err)

		// Set error in response
		resp.Message.(*v1alpha1.ConnectRequest_ProxyConfigResponse).ProxyConfigResponse.Result = &v1alpha1.ProxyConfigResponse_ErrorMessage{
			ErrorMessage: err.Error(),
		}
	} else {
		e.logger.Info("successfully retrieved proxy config",
			"request_id", req.RequestId,
			"namespace", req.PodNamespace,
			"pod", req.PodName,
			"version", proxyConfig.Version)

		// Set successful result
		resp.Message.(*v1alpha1.ConnectRequest_ProxyConfigResponse).ProxyConfigResponse.Result = &v1alpha1.ProxyConfigResponse_ProxyConfig{
			ProxyConfig: proxyConfig,
		}
	}

	// Send response back to manager
	e.mu.RLock()
	stream := e.stream
	e.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("no active stream to send proxy config response")
	}

	if err := stream.Send(resp); err != nil {
		e.logger.Error("failed to send proxy config response", "request_id", req.RequestId, "error", err)
		return fmt.Errorf("failed to send proxy config response: %w", err)
	}

	e.logger.Debug("proxy config response sent", "request_id", req.RequestId)
	return nil
}

// processServiceGraphMetricsRequest handles service graph metrics requests from the manager
func (e *EdgeService) processServiceGraphMetricsRequest(req *v1alpha1.ServiceGraphMetricsRequest) error {
	e.logger.Info("processing mesh metrics request", "request_id", req.RequestId)

	// Create response message
	resp := &v1alpha1.ConnectRequest{
		Message: &v1alpha1.ConnectRequest_ServiceGraphMetricsResponse{
			ServiceGraphMetricsResponse: &v1alpha1.ServiceGraphMetricsResponse{
				RequestId: req.RequestId,
			},
		},
	}

	// Check if metrics provider is available and enabled
	if e.metricsProvider == nil || e.metricsProvider.GetProviderInfo().Type == metrics.ProviderTypeNone {
		err := fmt.Errorf("metrics provider not available")
		e.logger.Error("mesh metrics request failed", "request_id", req.RequestId, "error", err)
		resp.Message.(*v1alpha1.ConnectRequest_ServiceGraphMetricsResponse).ServiceGraphMetricsResponse.Result = &v1alpha1.ServiceGraphMetricsResponse_ErrorMessage{
			ErrorMessage: err.Error(),
		}
	} else {
		// Convert filters and execute query
		query, err := convertMeshMetricsFilters(req.Filters, req.StartTime, req.EndTime)
		if err != nil {
			e.logger.Error("invalid mesh metrics request", "request_id", req.RequestId, "error", err)
			resp.Message.(*v1alpha1.ConnectRequest_ServiceGraphMetricsResponse).ServiceGraphMetricsResponse.Result = &v1alpha1.ServiceGraphMetricsResponse_ErrorMessage{
				ErrorMessage: err.Error(),
			}
		} else {
			meshMetrics, err := e.metricsProvider.GetServiceGraphMetrics(e.ctx, query)

			if err != nil {
				e.logger.Error("failed to get mesh metrics",
					"request_id", req.RequestId,
					"error", err)

				resp.Message.(*v1alpha1.ConnectRequest_ServiceGraphMetricsResponse).ServiceGraphMetricsResponse.Result = &v1alpha1.ServiceGraphMetricsResponse_ErrorMessage{
					ErrorMessage: err.Error(),
				}
			} else {
				e.logger.Info("successfully retrieved mesh metrics",
					"request_id", req.RequestId,
					"pairs_count", len(meshMetrics.Pairs))

				// Convert to protobuf format
				e.logger.Debug("converting service graph metrics to protobuf", "pairs_count", len(meshMetrics.Pairs), "cluster_id", e.config.GetClusterID())
				protoMeshMetrics := convertServiceGraphMetricsToProto(meshMetrics, e.config.GetClusterID())

				resp.Message.(*v1alpha1.ConnectRequest_ServiceGraphMetricsResponse).ServiceGraphMetricsResponse.Result = &v1alpha1.ServiceGraphMetricsResponse_ServiceGraphMetrics{
					ServiceGraphMetrics: protoMeshMetrics,
				}
				e.logger.Debug("service graph metrics response prepared", "request_id", req.RequestId, "pairs_count", len(protoMeshMetrics.Pairs))
			}
		}
	}

	// Send response back to manager
	e.mu.RLock()
	stream := e.stream
	e.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("no active stream to send mesh metrics response")
	}

	if err := stream.Send(resp); err != nil {
		e.logger.Error("failed to send mesh metrics response", "request_id", req.RequestId, "error", err)
		return fmt.Errorf("failed to send mesh metrics response: %w", err)
	}

	e.logger.Debug("mesh metrics response sent", "request_id", req.RequestId)
	return nil
}

// convertMeshMetricsFilters converts protobuf filters to Go types
func convertMeshMetricsFilters(protoFilters *types.GraphMetricsFilters, startTime, endTime *timestamppb.Timestamp) (metrics.MeshMetricsQuery, error) {
	var filters metrics.MeshMetricsFilters

	if protoFilters != nil {
		filters.Namespaces = protoFilters.Namespaces
	}

	// Validate that timestamps are provided
	if startTime == nil {
		return metrics.MeshMetricsQuery{}, fmt.Errorf("start_time is required")
	}
	if endTime == nil {
		return metrics.MeshMetricsQuery{}, fmt.Errorf("end_time is required")
	}

	// Convert timestamps
	start := startTime.AsTime()
	end := endTime.AsTime()

	// Validate time range
	if end.Before(start) {
		return metrics.MeshMetricsQuery{}, fmt.Errorf("end_time must be after start_time")
	}

	return metrics.MeshMetricsQuery{
		Filters:   filters,
		StartTime: start,
		EndTime:   end,
	}, nil
}

// convertServiceGraphMetricsToProto converts Go service graph metrics to protobuf format
func convertServiceGraphMetricsToProto(meshMetrics *metrics.ServiceGraphMetrics, clusterID string) *types.ServiceGraphMetrics {
	pairs := make([]*types.ServicePairMetrics, len(meshMetrics.Pairs))

	for i, pair := range meshMetrics.Pairs {
		pairs[i] = &types.ServicePairMetrics{
			SourceCluster:        pair.SourceCluster,
			SourceNamespace:      pair.SourceNamespace,
			SourceService:        pair.SourceService,
			DestinationCluster:   pair.DestinationCluster,
			DestinationNamespace: pair.DestinationNamespace,
			DestinationService:   pair.DestinationService,
			ErrorRate:            pair.ErrorRate,
			RequestRate:          pair.RequestRate,
		}
	}

	return &types.ServiceGraphMetrics{
		Pairs:     pairs,
		ClusterId: clusterID,
		Timestamp: meshMetrics.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
}
