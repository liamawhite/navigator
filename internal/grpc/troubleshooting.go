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
