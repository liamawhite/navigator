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
	"fmt"
	"net"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	frontendv1alpha1 "github.com/liamawhite/navigator/pkg/api/frontend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/grpc/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// setupGRPCServer configures and creates the gRPC server
func (s *ManagerServer) setupGRPCServer() error {
	// Create gRPC listener
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GetPort()))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.GetPort(), err)
	}
	s.listener = grpcListener

	// Create gRPC server with message size limits and validation interceptors
	maxMessageSize := s.config.GetMaxMessageSize()
	s.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMessageSize),
		grpc.MaxSendMsgSize(maxMessageSize),
		grpc.UnaryInterceptor(interceptors.ValidationInterceptor(s.logger)),
		grpc.StreamInterceptor(interceptors.StreamValidationInterceptor(s.logger)),
	)

	// Register backend services
	v1alpha1.RegisterManagerServiceServer(s.grpcServer, s)

	// Register frontend services
	frontendv1alpha1.RegisterServiceRegistryServiceServer(s.grpcServer, s.serviceRegistryService)
	frontendv1alpha1.RegisterMetricsServiceServer(s.grpcServer, s.metricsService)
	frontendv1alpha1.RegisterClusterRegistryServiceServer(s.grpcServer, s.clusterRegistryService)

	// Enable reflection for debugging
	reflection.Register(s.grpcServer)

	return nil
}
