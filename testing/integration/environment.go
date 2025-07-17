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

package integration

import (
	"context"
	"testing"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// TestEnvironment provides an abstract interface for different test environments
// (local Kind clusters, remote clusters, etc.)
type TestEnvironment interface {
	// Setup initializes the test environment
	Setup(t *testing.T) error

	// Cleanup cleans up the test environment
	Cleanup(t *testing.T) error

	// GetGRPCClient returns a gRPC client for the Navigator service
	GetGRPCClient() v1alpha1.ServiceRegistryServiceClient

	// GetNamespace returns the test namespace
	GetNamespace() string

	// CreateServices creates test services in the environment
	CreateServices(ctx context.Context, services []ServiceSpec) error

	// WaitForServices waits for services to be ready
	WaitForServices(ctx context.Context, serviceNames []string) error

	// DeleteServices removes test services
	DeleteServices(ctx context.Context, serviceNames []string) error
}

// ServiceSpec defines how to create a test service
type ServiceSpec struct {
	Name     string
	Replicas int32
	Type     ServiceType
	// ExternalIPs is used for external services
	ExternalIPs []string
	// NextService is used for microservice topology chaining
	NextService string
}

// ServiceType represents different types of services for testing
type ServiceType string

const (
	ServiceTypeWeb      ServiceType = "web"
	ServiceTypeHeadless ServiceType = "headless"
	ServiceTypeExternal ServiceType = "external"
	ServiceTypeTopology ServiceType = "topology"
)

// TestContext provides common context for integration tests
type TestContext struct {
	Environment TestEnvironment
	Namespace   string
	GRPCClient  v1alpha1.ServiceRegistryServiceClient
}
