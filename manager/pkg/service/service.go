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
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// ConnectionManager interface for dependency injection
type ConnectionManager interface {
	RegisterConnection(clusterID string, stream v1alpha1.ManagerService_ConnectServer) error
	UnregisterConnection(clusterID string)
	UpdateClusterState(clusterID string, clusterState *v1alpha1.ClusterState) error
	GetClusterState(clusterID string) (*v1alpha1.ClusterState, error)
	GetAllClusterStates() map[string]*v1alpha1.ClusterState
	IsClusterConnected(clusterID string) bool
	GetActiveClusterCount() int
	SendMessageToCluster(clusterID string, message *v1alpha1.ConnectResponse) error
}

// Config interface for dependency injection
type Config interface {
	GetPort() int
	GetMaxMessageSize() int
	Validate() error
}

// ManagerService implements the gRPC ManagerService
type ManagerService struct {
	v1alpha1.UnimplementedManagerServiceServer
	config            Config
	connectionManager ConnectionManager
	proxyService      *ProxyService
	logger            *slog.Logger
	server            *grpc.Server
	listener          net.Listener
	httpServer        *http.Server
	httpListener      net.Listener
	frontendService   *FrontendService
	mu                sync.RWMutex
	running           bool
}

// NewManagerService creates a new manager service
func NewManagerService(config Config, connectionManager ConnectionManager, logger *slog.Logger) (*ManagerService, error) {
	// Validate configuration first
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manager configuration: %w", err)
	}
	// Cast to ReadOptimizedConnectionManager - this is safe since our actual implementation
	// (connections.Manager) implements all the required methods
	readOptimizedManager, ok := connectionManager.(ReadOptimizedConnectionManager)
	if !ok {
		// This should not happen in practice since we always pass connections.Manager
		panic("connectionManager must implement ReadOptimizedConnectionManager")
	}

	proxyService := NewProxyService(connectionManager, logger)
	frontendService := NewFrontendService(readOptimizedManager, proxyService, logger)

	return &ManagerService{
		config:            config,
		connectionManager: connectionManager,
		proxyService:      proxyService,
		logger:            logger,
		frontendService:   frontendService,
	}, nil
}

// Start starts the gRPC server and HTTP gateway
func (m *ManagerService) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("manager service is already running")
	}

	// Create gRPC listener
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", m.config.GetPort()))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", m.config.GetPort(), err)
	}

	m.listener = grpcListener

	// Create HTTP listener (gRPC port + 1)
	httpPort := m.config.GetPort() + 1
	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
	if err != nil {
		return fmt.Errorf("failed to listen on HTTP port %d: %w", httpPort, err)
	}

	m.httpListener = httpListener

	// Create gRPC server with message size limits
	maxMessageSize := m.config.GetMaxMessageSize()
	m.server = grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMessageSize),
		grpc.MaxSendMsgSize(maxMessageSize),
	)

	// Register services
	v1alpha1.RegisterManagerServiceServer(m.server, m)
	frontendv1alpha1.RegisterServiceRegistryServiceServer(m.server, m.frontendService)

	// Enable reflection for debugging
	reflection.Register(m.server)

	// Create HTTP gateway
	if err := m.setupHTTPGateway(); err != nil {
		return fmt.Errorf("failed to setup HTTP gateway: %w", err)
	}

	m.running = true

	// Start gRPC server in goroutine
	go func() {
		m.logger.Info("starting gRPC server", "port", m.config.GetPort())
		if err := m.server.Serve(grpcListener); err != nil {
			m.logger.Error("gRPC server error", "error", err)
		}
	}()

	// Start HTTP server in goroutine
	go func() {
		m.logger.Info("starting HTTP gateway", "port", httpPort)
		if err := m.httpServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
			m.logger.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server and HTTP gateway
func (m *ManagerService) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.logger.Info("stopping gRPC server and HTTP gateway")

	// Graceful shutdown of HTTP server
	if m.httpServer != nil {
		if err := m.httpServer.Shutdown(context.Background()); err != nil {
			m.logger.Error("error shutting down HTTP server", "error", err)
		}
	}

	// Graceful shutdown of gRPC server
	if m.server != nil {
		m.server.GracefulStop()
	}

	// Close listeners
	if m.listener != nil {
		_ = m.listener.Close()
	}
	if m.httpListener != nil {
		_ = m.httpListener.Close()
	}

	m.running = false

	return nil
}

// setupHTTPGateway sets up the HTTP gateway for the frontend API
func (m *ManagerService) setupHTTPGateway() error {
	// Create gRPC gateway mux
	mux := runtime.NewServeMux()

	// Register frontend service with the gateway
	grpcEndpoint := fmt.Sprintf("localhost:%d", m.config.GetPort())
	maxMessageSize := m.config.GetMaxMessageSize()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
	}

	err := frontendv1alpha1.RegisterServiceRegistryServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		grpcEndpoint,
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register frontend service handler: %w", err)
	}

	// Create HTTP server
	m.httpServer = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	return nil
}

// Connect handles bidirectional streaming connections from edge processes
func (m *ManagerService) Connect(stream v1alpha1.ManagerService_ConnectServer) error {
	m.logger.Info("new connection attempt")

	// Wait for cluster identification
	req, err := stream.Recv()
	if err != nil {
		m.logger.Error("failed to receive cluster identification", "error", err)
		return status.Errorf(codes.InvalidArgument, "failed to receive cluster identification: %v", err)
	}

	// Process cluster identification
	clusterID, err := m.processClusterIdentification(req)
	if err != nil {
		m.logger.Error("failed to process cluster identification", "error", err)

		// Send error response
		errorResp := &v1alpha1.ConnectResponse{
			Message: &v1alpha1.ConnectResponse_Error{
				Error: &v1alpha1.ErrorMessage{
					ErrorCode:    "INVALID_CLUSTER_IDENTIFICATION",
					ErrorMessage: err.Error(),
				},
			},
		}

		if sendErr := stream.Send(errorResp); sendErr != nil {
			m.logger.Error("failed to send error response", "error", sendErr)
		}

		return status.Errorf(codes.InvalidArgument, "invalid cluster identification: %v", err)
	}

	// Try to register connection
	if err := m.connectionManager.RegisterConnection(clusterID, stream); err != nil {
		m.logger.Error("failed to register connection", "cluster_id", clusterID, "error", err)

		// Send rejection response
		rejectionResp := &v1alpha1.ConnectResponse{
			Message: &v1alpha1.ConnectResponse_ConnectionAck{
				ConnectionAck: &v1alpha1.ConnectionAck{
					Accepted: false,
				},
			},
		}

		if sendErr := stream.Send(rejectionResp); sendErr != nil {
			m.logger.Error("failed to send rejection response", "error", sendErr)
		}

		return status.Errorf(codes.AlreadyExists, "connection rejected: %v", err)
	}

	// Send connection acceptance
	acceptanceResp := &v1alpha1.ConnectResponse{
		Message: &v1alpha1.ConnectResponse_ConnectionAck{
			ConnectionAck: &v1alpha1.ConnectionAck{
				Accepted: true,
			},
		},
	}

	if err := stream.Send(acceptanceResp); err != nil {
		m.logger.Error("failed to send acceptance response", "error", err)
		m.connectionManager.UnregisterConnection(clusterID)
		return status.Errorf(codes.Internal, "failed to send acceptance response: %v", err)
	}

	m.logger.Info("connection accepted", "cluster_id", clusterID)

	// Handle incoming messages
	defer func() {
		m.connectionManager.UnregisterConnection(clusterID)
		m.logger.Info("connection closed", "cluster_id", clusterID)
	}()

	for {
		req, err := stream.Recv()
		if err != nil {
			m.logger.Info("connection terminated", "cluster_id", clusterID, "error", err)
			return nil
		}

		if err := m.processIncomingMessage(clusterID, req); err != nil {
			m.logger.Error("failed to process message", "cluster_id", clusterID, "error", err)

			// Send error response
			errorResp := &v1alpha1.ConnectResponse{
				Message: &v1alpha1.ConnectResponse_Error{
					Error: &v1alpha1.ErrorMessage{
						ErrorCode:    "MESSAGE_PROCESSING_ERROR",
						ErrorMessage: err.Error(),
					},
				},
			}

			if sendErr := stream.Send(errorResp); sendErr != nil {
				m.logger.Error("failed to send error response", "error", sendErr)
			}

			return status.Errorf(codes.InvalidArgument, "message processing error: %v", err)
		}
	}
}

// processIncomingMessage processes different types of messages from edges
func (m *ManagerService) processIncomingMessage(clusterID string, req *v1alpha1.ConnectRequest) error {
	switch msg := req.Message.(type) {
	case *v1alpha1.ConnectRequest_ClusterState:
		return m.processClusterStateUpdate(clusterID, req)
	case *v1alpha1.ConnectRequest_ProxyConfigResponse:
		return m.processProxyConfigResponse(msg.ProxyConfigResponse)
	default:
		m.logger.Warn("received unknown message type", "cluster_id", clusterID, "type", fmt.Sprintf("%T", msg))
		return fmt.Errorf("unknown message type: %T", msg)
	}
}

// processProxyConfigResponse processes proxy configuration responses from edges
func (m *ManagerService) processProxyConfigResponse(response *v1alpha1.ProxyConfigResponse) error {
	m.logger.Debug("processing proxy config response", "request_id", response.RequestId)
	return m.proxyService.HandleProxyConfigResponse(response)
}

// processClusterIdentification processes cluster identification request
func (m *ManagerService) processClusterIdentification(req *v1alpha1.ConnectRequest) (string, error) {
	if req.Message == nil {
		return "", fmt.Errorf("empty message")
	}

	clusterIdentification, ok := req.Message.(*v1alpha1.ConnectRequest_ClusterIdentification)
	if !ok {
		return "", fmt.Errorf("expected cluster identification, got %T", req.Message)
	}

	if clusterIdentification.ClusterIdentification == nil {
		return "", fmt.Errorf("nil cluster identification")
	}

	clusterID := clusterIdentification.ClusterIdentification.ClusterId
	if clusterID == "" {
		return "", fmt.Errorf("empty cluster ID")
	}

	return clusterID, nil
}

// processClusterStateUpdate processes cluster state update request
func (m *ManagerService) processClusterStateUpdate(clusterID string, req *v1alpha1.ConnectRequest) error {
	if req.Message == nil {
		return fmt.Errorf("empty message")
	}

	clusterStateMsg, ok := req.Message.(*v1alpha1.ConnectRequest_ClusterState)
	if !ok {
		return fmt.Errorf("expected cluster state, got %T", req.Message)
	}

	if clusterStateMsg.ClusterState == nil {
		return fmt.Errorf("nil cluster state")
	}

	// Update cluster state
	if err := m.connectionManager.UpdateClusterState(clusterID, clusterStateMsg.ClusterState); err != nil {
		return fmt.Errorf("failed to update cluster state: %w", err)
	}

	m.logger.Debug("cluster state updated", "cluster_id", clusterID, "services", len(clusterStateMsg.ClusterState.Services))

	return nil
}

// GetProxyService returns the proxy service for external access
func (m *ManagerService) GetProxyService() *ProxyService {
	return m.proxyService
}
