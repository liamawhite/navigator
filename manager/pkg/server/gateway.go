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
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// setupHTTPGateway sets up the HTTP gateway for the frontend API
func (s *ManagerServer) setupHTTPGateway() error {
	// Create HTTP listener (actual gRPC port + 1, or 0 if configured port was 0)
	var httpPort int
	if s.config.GetPort() == 0 {
		// If configured with port 0, use port 0 for HTTP listener too (system will assign)
		httpPort = 0
	} else {
		// Otherwise use configured port + 1
		httpPort = s.config.GetPort() + 1
	}
	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
	if err != nil {
		return fmt.Errorf("failed to listen on HTTP port %d: %w", httpPort, err)
	}
	s.httpListener = httpListener

	// Create gRPC gateway mux
	mux := runtime.NewServeMux()

	// Setup gRPC connection options
	grpcEndpoint := fmt.Sprintf("localhost:%d", s.config.GetPort())
	maxMessageSize := s.config.GetMaxMessageSize()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMessageSize),
			grpc.MaxCallSendMsgSize(maxMessageSize),
		),
	}

	// Register service registry service handler
	if err := frontendv1alpha1.RegisterServiceRegistryServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		grpcEndpoint,
		opts,
	); err != nil {
		return fmt.Errorf("failed to register service registry handler: %w", err)
	}

	// Register metrics service handler
	if err := frontendv1alpha1.RegisterMetricsServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		grpcEndpoint,
		opts,
	); err != nil {
		return fmt.Errorf("failed to register metrics service handler: %w", err)
	}

	// Register cluster registry service handler
	if err := frontendv1alpha1.RegisterClusterRegistryServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		grpcEndpoint,
		opts,
	); err != nil {
		return fmt.Errorf("failed to register cluster registry service handler: %w", err)
	}

	// Create HTTP server
	s.httpServer = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	return nil
}
