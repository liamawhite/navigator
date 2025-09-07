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

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/liamawhite/navigator/manager/pkg/backend"
	"github.com/liamawhite/navigator/manager/pkg/frontend"
	"github.com/liamawhite/navigator/manager/pkg/providers"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"google.golang.org/grpc"
)

// ManagerServer orchestrates all manager services
type ManagerServer struct {
	v1alpha1.UnimplementedManagerServiceServer
	config            providers.Config
	connectionManager providers.ReadOptimizedConnectionManager
	logger            *slog.Logger
	grpcServer        *grpc.Server
	httpServer        *http.Server
	listener          net.Listener
	httpListener      net.Listener
	mu                sync.RWMutex
	running           bool

	// Backend services
	proxyService       *backend.ProxyService
	meshMetricsService *backend.MeshMetricsService

	// Provider implementations
	istioProvider providers.IstioResourcesProvider

	// Frontend services
	serviceRegistryService *frontend.ServiceRegistryService
	metricsService         *frontend.MetricsService
	clusterRegistryService *frontend.ClusterRegistryService
}

// NewManagerServer creates a new manager server
func NewManagerServer(config providers.Config, connectionManager providers.ReadOptimizedConnectionManager, logger *slog.Logger) (*ManagerServer, error) {
	// Validate configuration first
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manager configuration: %w", err)
	}

	// Create backend services
	proxyService := backend.NewProxyService(connectionManager, logger)
	meshMetricsService := backend.NewMeshMetricsService(connectionManager, logger)

	// Create provider implementations
	istioProvider := backend.NewIstioService(connectionManager, logger)

	// Create frontend services
	serviceRegistryService := frontend.NewServiceRegistryService(connectionManager, proxyService, istioProvider, logger)
	metricsService := frontend.NewMetricsService(connectionManager, meshMetricsService, logger)
	clusterRegistryService := frontend.NewClusterRegistryService(connectionManager, logger)

	return &ManagerServer{
		config:                 config,
		connectionManager:      connectionManager,
		logger:                 logger,
		proxyService:           proxyService,
		meshMetricsService:     meshMetricsService,
		istioProvider:          istioProvider,
		serviceRegistryService: serviceRegistryService,
		metricsService:         metricsService,
		clusterRegistryService: clusterRegistryService,
	}, nil
}

// Start starts the gRPC server and HTTP gateway
func (s *ManagerServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("manager server is already running")
	}

	// Setup gRPC server
	if err := s.setupGRPCServer(); err != nil {
		return fmt.Errorf("failed to setup gRPC server: %w", err)
	}

	// Setup HTTP gateway
	if err := s.setupHTTPGateway(); err != nil {
		return fmt.Errorf("failed to setup HTTP gateway: %w", err)
	}

	s.running = true

	// Start both servers in goroutines
	s.startServers()

	return nil
}

// Stop stops the gRPC server and HTTP gateway
func (s *ManagerServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("stopping gRPC server and HTTP gateway")

	// Graceful shutdown of HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(context.Background()); err != nil {
			s.logger.Error("error shutting down HTTP server", "error", err)
		}
	}

	// Graceful shutdown of gRPC server
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	// Close listeners
	if s.listener != nil {
		_ = s.listener.Close()
	}
	if s.httpListener != nil {
		_ = s.httpListener.Close()
	}

	s.running = false

	return nil
}

// startServers starts the gRPC and HTTP servers in separate goroutines
func (s *ManagerServer) startServers() {
	// Start gRPC server
	go func() {
		s.logger.Info("starting gRPC server", "port", s.config.GetPort())
		if err := s.grpcServer.Serve(s.listener); err != nil {
			s.logger.Error("gRPC server error", "error", err)
		}
	}()

	// Start HTTP server
	go func() {
		// Get the actual port from the listener
		actualPort := s.httpListener.Addr().(*net.TCPAddr).Port
		s.logger.Info("starting HTTP gateway", "port", actualPort)
		if err := s.httpServer.Serve(s.httpListener); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()
}

// GetProxyService returns the proxy service for external access (backward compatibility)
func (s *ManagerServer) GetProxyService() *backend.ProxyService {
	return s.proxyService
}
