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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/datastore/mock"
)

func TestServiceRegistryServer_ListServices(t *testing.T) {
	tests := []struct {
		name          string
		request       *v1alpha1.ListServicesRequest
		expectedCount int
		expectedError bool
	}{
		{
			name: "list services in specific namespace",
			request: &v1alpha1.ListServicesRequest{
				Namespace: stringPtr("default"),
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "list services in empty namespace",
			request: &v1alpha1.ListServicesRequest{
				Namespace: stringPtr("empty"),
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:    "list all services (no namespace specified)",
			request: &v1alpha1.ListServicesRequest{
				// Namespace is nil (not specified)
			},
			expectedCount: 3, // All services from all namespaces
			expectedError: false,
		},
	}

	// Set up mock datastore
	mockDS := &mock.Datastore{
		Services: map[string][]*v1alpha1.Service{
			"default": {
				{
					Id:        "default:service-1",
					Name:      "service-1",
					Namespace: "default",
					Instances: []*v1alpha1.ServiceInstance{
						{Ip: "10.0.0.1", Pod: "pod-1", Namespace: "default"},
					},
				},
				{
					Id:        "default:service-2",
					Name:      "service-2",
					Namespace: "default",
					Instances: []*v1alpha1.ServiceInstance{
						{Ip: "10.0.0.2", Pod: "pod-2", Namespace: "default"},
					},
				},
			},
			"kube-system": {
				{
					Id:        "kube-system:kube-dns",
					Name:      "kube-dns",
					Namespace: "kube-system",
					Instances: []*v1alpha1.ServiceInstance{
						{Ip: "10.0.0.3", Pod: "kube-dns-pod", Namespace: "kube-system"},
					},
				},
			},
		},
	}

	server := NewServiceRegistryServer(mockDS)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := server.ListServices(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Len(t, response.Services, tt.expectedCount)
		})
	}
}

func TestServiceRegistryServer_GetService(t *testing.T) {
	tests := []struct {
		name          string
		request       *v1alpha1.GetServiceRequest
		expectedName  string
		expectedError bool
	}{
		{
			name: "get existing service",
			request: &v1alpha1.GetServiceRequest{
				Id: "default:service-1",
			},
			expectedName:  "service-1",
			expectedError: false,
		},
		{
			name: "get non-existent service",
			request: &v1alpha1.GetServiceRequest{
				Id: "default:non-existent",
			},
			expectedError: true,
		},
	}

	// Set up mock datastore
	mockDS := &mock.Datastore{
		Services: map[string][]*v1alpha1.Service{
			"default": {
				{
					Id:        "default:service-1",
					Name:      "service-1",
					Namespace: "default",
					Instances: []*v1alpha1.ServiceInstance{
						{Ip: "10.0.0.1", Pod: "pod-1", Namespace: "default"},
					},
				},
			},
		},
	}

	server := NewServiceRegistryServer(mockDS)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := server.GetService(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, tt.expectedName, response.Service.Name)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
