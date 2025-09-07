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

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Connect handles bidirectional streaming connections from edge processes
func (s *ManagerServer) Connect(stream v1alpha1.ManagerService_ConnectServer) error {
	s.logger.Info("new connection attempt")

	// Wait for cluster identification
	req, err := stream.Recv()
	if err != nil {
		s.logger.Error("failed to receive cluster identification", "error", err)
		return status.Errorf(codes.InvalidArgument, "failed to receive cluster identification: %v", err)
	}

	// Process cluster identification
	clusterID, capabilities, err := s.processClusterIdentification(req)
	if err != nil {
		s.logger.Error("failed to process cluster identification", "error", err)

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
			s.logger.Error("failed to send error response", "error", sendErr)
		}

		return status.Errorf(codes.InvalidArgument, "invalid cluster identification: %v", err)
	}

	// Try to register connection
	if err := s.connectionManager.RegisterConnection(clusterID, stream); err != nil {
		s.logger.Error("failed to register connection", "cluster_id", clusterID, "error", err)

		// Send rejection response
		rejectionResp := &v1alpha1.ConnectResponse{
			Message: &v1alpha1.ConnectResponse_ConnectionAck{
				ConnectionAck: &v1alpha1.ConnectionAck{
					Accepted: false,
				},
			},
		}

		if sendErr := stream.Send(rejectionResp); sendErr != nil {
			s.logger.Error("failed to send rejection response", "error", sendErr)
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
		s.logger.Error("failed to send acceptance response", "error", err)
		s.connectionManager.UnregisterConnection(clusterID)
		return status.Errorf(codes.Internal, "failed to send acceptance response: %v", err)
	}

	// Update capabilities for the connection
	if capabilities != nil {
		if err := s.connectionManager.UpdateCapabilities(clusterID, capabilities); err != nil {
			s.logger.Error("failed to update capabilities", "cluster_id", clusterID, "error", err)
		} else {
			s.logger.Info("connection capabilities updated",
				"cluster_id", clusterID,
				"metrics_enabled", capabilities.MetricsEnabled)
		}
	}

	s.logger.Info("connection accepted", "cluster_id", clusterID)

	// Handle incoming messages
	defer func() {
		s.connectionManager.UnregisterConnection(clusterID)
		s.logger.Info("connection closed", "cluster_id", clusterID)
	}()

	for {
		req, err := stream.Recv()
		if err != nil {
			s.logger.Info("connection terminated", "cluster_id", clusterID, "error", err)
			return nil
		}

		if err := s.processIncomingMessage(clusterID, req); err != nil {
			s.logger.Error("failed to process message", "cluster_id", clusterID, "error", err)

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
				s.logger.Error("failed to send error response", "error", sendErr)
			}

			return status.Errorf(codes.InvalidArgument, "message processing error: %v", err)
		}
	}
}

// processIncomingMessage processes different types of messages from edges
func (s *ManagerServer) processIncomingMessage(clusterID string, req *v1alpha1.ConnectRequest) error {
	switch msg := req.Message.(type) {
	case *v1alpha1.ConnectRequest_ClusterState:
		return s.processClusterStateUpdate(clusterID, req)
	case *v1alpha1.ConnectRequest_ProxyConfigResponse:
		return s.processProxyConfigResponse(msg.ProxyConfigResponse)
	case *v1alpha1.ConnectRequest_ServiceGraphMetricsResponse:
		return s.processServiceGraphMetricsResponse(msg.ServiceGraphMetricsResponse)
	default:
		s.logger.Warn("received unknown message type", "cluster_id", clusterID, "type", fmt.Sprintf("%T", msg))
		return fmt.Errorf("unknown message type: %T", msg)
	}
}

// processProxyConfigResponse processes proxy configuration responses from edges
func (s *ManagerServer) processProxyConfigResponse(response *v1alpha1.ProxyConfigResponse) error {
	s.logger.Debug("processing proxy config response", "request_id", response.RequestId)
	return s.proxyService.HandleProxyConfigResponse(response)
}

// processServiceGraphMetricsResponse processes service graph metrics responses from edges
func (s *ManagerServer) processServiceGraphMetricsResponse(response *v1alpha1.ServiceGraphMetricsResponse) error {
	s.logger.Debug("processing service graph metrics response", "request_id", response.RequestId)
	s.meshMetricsService.HandleServiceGraphMetricsResponse(response)
	return nil
}

// processClusterIdentification processes cluster identification request and returns clusterID and capabilities
func (s *ManagerServer) processClusterIdentification(req *v1alpha1.ConnectRequest) (string, *v1alpha1.EdgeCapabilities, error) {
	if req.Message == nil {
		return "", nil, fmt.Errorf("empty message")
	}

	clusterIdentification, ok := req.Message.(*v1alpha1.ConnectRequest_ClusterIdentification)
	if !ok {
		return "", nil, fmt.Errorf("expected cluster identification, got %T", req.Message)
	}

	if clusterIdentification.ClusterIdentification == nil {
		return "", nil, fmt.Errorf("nil cluster identification")
	}

	clusterID := clusterIdentification.ClusterIdentification.ClusterId
	if clusterID == "" {
		return "", nil, fmt.Errorf("empty cluster ID")
	}

	capabilities := clusterIdentification.ClusterIdentification.Capabilities
	// Capabilities are optional, but if present should be valid

	return clusterID, capabilities, nil
}

// processClusterStateUpdate processes cluster state update request
func (s *ManagerServer) processClusterStateUpdate(clusterID string, req *v1alpha1.ConnectRequest) error {
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
	if err := s.connectionManager.UpdateClusterState(clusterID, clusterStateMsg.ClusterState); err != nil {
		return fmt.Errorf("failed to update cluster state: %w", err)
	}

	s.logger.Debug("cluster state updated", "cluster_id", clusterID, "services", len(clusterStateMsg.ClusterState.Services))

	return nil
}
