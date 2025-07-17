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

package localenv

import (
	"context"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// LocalEnvironment provides an interface for managing local development environments
type LocalEnvironment interface {
	// Setup initializes the local environment with the given configuration
	Setup(ctx context.Context, config *Config) error

	// Teardown cleans up the local environment and all its resources
	Teardown(ctx context.Context) error

	// DeployScenario deploys a predefined scenario to the environment
	DeployScenario(ctx context.Context, scenario *Scenario) error

	// GetKubeconfig returns the path to the kubeconfig file for this environment
	GetKubeconfig() string

	// GetGRPCClient returns a gRPC client connected to the Navigator service running in this environment
	GetGRPCClient() v1alpha1.ServiceRegistryServiceClient

	// GetNamespace returns the primary namespace used for this environment
	GetNamespace() string

	// IsReady checks if the environment is ready for use
	IsReady(ctx context.Context) bool
}

// Config holds configuration for the local environment
type Config struct {
	// ClusterName is the name of the cluster to create
	ClusterName string

	// Namespace is the primary namespace to use for services
	Namespace string

	// Port is the port on which Navigator will listen for gRPC requests
	Port int

	// HTTPPort is the port on which Navigator will listen for HTTP requests
	HTTPPort int

	// IstioEnabled determines whether to install Istio in the cluster
	IstioEnabled bool

	// CleanupOnExit determines whether to clean up resources when the environment is torn down
	CleanupOnExit bool

	// WaitTimeout is the maximum time to wait for cluster and services to be ready
	WaitTimeout string
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ClusterName:   "navigator-demo",
		Namespace:     "demo",
		Port:          8080,
		HTTPPort:      8081,
		IstioEnabled:  false,
		CleanupOnExit: true,
		WaitTimeout:   "5m",
	}
}
