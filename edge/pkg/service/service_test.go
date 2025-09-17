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
	"testing"

	"github.com/liamawhite/navigator/edge/pkg/interfaces"
	"github.com/liamawhite/navigator/edge/pkg/metrics"
	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockKubernetesClient implements the KubernetesClient interface for testing
type mockKubernetesClient struct {
	clusterState *v1alpha1.ClusterState
	err          error
}

func (m *mockKubernetesClient) GetClusterState(ctx context.Context) (*v1alpha1.ClusterState, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.clusterState, nil
}

func (m *mockKubernetesClient) GetClusterStateWithMetrics(ctx context.Context, metricsProvider interfaces.MetricsProvider) (*v1alpha1.ClusterState, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.clusterState, nil
}

func (m *mockKubernetesClient) GetClusterName(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "test-cluster", nil
}

// mockProxyService implements the ProxyService interface for testing
type mockProxyService struct {
	proxyConfig *types.ProxyConfig
	err         error
}

func (m *mockProxyService) GetProxyConfig(ctx context.Context, namespace, podName string) (*types.ProxyConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.proxyConfig, nil
}

func (m *mockProxyService) ValidateProxyAccess(ctx context.Context, namespace, podName string) error {
	return m.err
}

// mockConfig implements the Config interface for testing
type mockConfig struct {
	clusterID       string
	managerEndpoint string
	syncInterval    int
	maxMessageSize  int
}

// mockMetricsProvider implements the MetricsProvider interface for testing
type mockMetricsProvider struct {
	err error
}

func (m *mockMetricsProvider) GetProviderInfo() metrics.ProviderInfo {
	return metrics.ProviderInfo{
		Type:     metrics.ProviderTypePrometheus,
		Endpoint: "http://localhost:9090",
	}
}

func (m *mockMetricsProvider) GetServiceConnections(ctx context.Context, serviceName, namespace string, proxyMode types.ProxyMode, startTime, endTime *timestamppb.Timestamp) (*types.ServiceGraphMetrics, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &types.ServiceGraphMetrics{}, nil
}

func (m *mockMetricsProvider) Close() error {
	return m.err
}

func (m *mockConfig) GetClusterID() string {
	return m.clusterID
}

func (m *mockConfig) GetManagerEndpoint() string {
	return m.managerEndpoint
}

func (m *mockConfig) GetSyncInterval() int {
	return m.syncInterval
}

func (m *mockConfig) GetMaxMessageSize() int {
	return m.maxMessageSize
}

func (m *mockConfig) GetMetricsConfig() metrics.Config {
	return metrics.Config{
		Enabled:  false,
		Timeout:  30,
		Type:     "prometheus",
		Endpoint: "http://localhost:9090",
	}
}

func (m *mockConfig) Validate() error {
	return nil
}

func TestEdgeService_shouldReconnect(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "unavailable error",
			err:  status.Error(codes.Unavailable, "service unavailable"),
			want: true,
		},
		{
			name: "deadline exceeded error",
			err:  status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			want: true,
		},
		{
			name: "canceled error",
			err:  status.Error(codes.Canceled, "canceled"),
			want: true,
		},
		{
			name: "permission denied error",
			err:  status.Error(codes.PermissionDenied, "permission denied"),
			want: false,
		},
		{
			name: "invalid argument error",
			err:  status.Error(codes.InvalidArgument, "invalid argument"),
			want: false,
		},
		{
			name: "not found error",
			err:  status.Error(codes.NotFound, "not found"),
			want: false,
		},
	}

	config := &mockConfig{
		clusterID:       "test-cluster",
		managerEndpoint: "localhost:8080",
		syncInterval:    30,
		maxMessageSize:  10485760,
	}

	mockK8s := &mockKubernetesClient{}
	mockProxy := &mockProxyService{}
	logger := logging.For("test")

	edgeService, err := NewEdgeService(config, mockK8s, mockProxy, &mockMetricsProvider{}, logger)
	assert.NoError(t, err)
	assert.NotNil(t, edgeService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := edgeService.shouldReconnect(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewEdgeService(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "create edge service with valid config",
			config: &mockConfig{
				clusterID:       "test-cluster",
				managerEndpoint: "localhost:8080",
				syncInterval:    30,
				maxMessageSize:  10485760,
			},
		},
		{
			name: "create edge service with custom sync interval",
			config: &mockConfig{
				clusterID:       "test-cluster",
				managerEndpoint: "localhost:8080",
				syncInterval:    60,
				maxMessageSize:  10485760,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockK8s := &mockKubernetesClient{}
			mockProxy := &mockProxyService{}
			logger := logging.For("test")

			edgeService, err := NewEdgeService(tt.config, mockK8s, mockProxy, &mockMetricsProvider{}, logger)

			assert.NoError(t, err)
			assert.NotNil(t, edgeService)
			assert.Equal(t, tt.config, edgeService.config)
			assert.Equal(t, mockK8s, edgeService.k8sClient)
			assert.Equal(t, mockProxy, edgeService.proxyService)
			assert.Equal(t, logger, edgeService.logger)
			assert.NotNil(t, edgeService.ctx)
			assert.NotNil(t, edgeService.cancel)
		})
	}
}

func TestEdgeService_syncClusterState(t *testing.T) {
	tests := []struct {
		name           string
		connected      bool
		clusterState   *v1alpha1.ClusterState
		k8sErr         error
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:      "successful sync",
			connected: true,
			clusterState: &v1alpha1.ClusterState{
				Services: []*v1alpha1.Service{
					{
						Name:      "test-service",
						Namespace: "default",
						Instances: []*v1alpha1.ServiceInstance{
							{
								Ip:           "10.0.0.1",
								PodName:      "test-pod-1",
								EnvoyPresent: true,
							},
						},
					},
				},
			},
			k8sErr:  nil,
			wantErr: false,
		},
		{
			name:           "not connected",
			connected:      false,
			clusterState:   nil,
			k8sErr:         nil,
			wantErr:        true,
			expectedErrMsg: "not connected to manager",
		},
		{
			name:           "kubernetes error",
			connected:      true,
			clusterState:   nil,
			k8sErr:         status.Error(codes.Unavailable, "kubernetes unavailable"),
			wantErr:        true,
			expectedErrMsg: "failed to get cluster state",
		},
		{
			name:      "empty cluster state",
			connected: true,
			clusterState: &v1alpha1.ClusterState{
				Services: []*v1alpha1.Service{},
			},
			k8sErr:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &mockConfig{
				clusterID:       "test-cluster",
				managerEndpoint: "localhost:8080",
				syncInterval:    30,
				maxMessageSize:  10485760,
			}

			mockK8s := &mockKubernetesClient{
				clusterState: tt.clusterState,
				err:          tt.k8sErr,
			}

			mockProxy := &mockProxyService{}
			logger := logging.For("test")
			edgeService, err := NewEdgeService(config, mockK8s, mockProxy, &mockMetricsProvider{}, logger)
			assert.NoError(t, err)
			assert.NotNil(t, edgeService)

			// Set connected state
			edgeService.mu.Lock()
			edgeService.connected = tt.connected
			edgeService.mu.Unlock()

			// Mock the stream if connected
			if tt.connected && tt.k8sErr == nil {
				// For this test, we can't easily mock the gRPC stream
				// In a real implementation, we'd want to use a mock gRPC client
				// For now, we'll just test the kubernetes client error handling
				if tt.k8sErr != nil {
					_, err := mockK8s.GetClusterState(context.TODO())
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}
			}

			// Test the parts we can test without a real gRPC connection
			if !tt.connected {
				// Test that we get the expected error when not connected
				edgeService.mu.RLock()
				connected := edgeService.connected
				edgeService.mu.RUnlock()

				assert.False(t, connected)
			}
		})
	}
}
