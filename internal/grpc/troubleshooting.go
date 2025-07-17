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

package grpc

import (
	"context"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/troubleshooting"
)

// TroubleshootingServer implements the TroubleshootingService gRPC interface.
type TroubleshootingServer struct {
	v1alpha1.UnimplementedTroubleshootingServiceServer
	datastore troubleshooting.ProxyDatastore
}

// NewTroubleshootingServer creates a new TroubleshootingServer with the given datastore.
func NewTroubleshootingServer(ds troubleshooting.ProxyDatastore) *TroubleshootingServer {
	return &TroubleshootingServer{
		datastore: ds,
	}
}

// GetProxyConfig returns the proxy configuration for a specific service instance.
func (s *TroubleshootingServer) GetProxyConfig(ctx context.Context, req *v1alpha1.GetProxyConfigRequest) (*v1alpha1.GetProxyConfigResponse, error) {
	proxyConfig, err := s.datastore.GetProxyConfig(ctx, req.ServiceId, req.InstanceId)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.GetProxyConfigResponse{
		ProxyConfig: proxyConfig,
	}, nil
}
