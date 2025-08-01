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

package demo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/liamawhite/navigator/pkg/localenv"
	"github.com/liamawhite/navigator/pkg/logging"
)

// ClusterManager manages Kind clusters for demos
type ClusterManager struct {
	logger *slog.Logger
	env    *localenv.KindEnvironment
	config *localenv.Config
}

// NewClusterManager creates a new cluster manager for demos
func NewClusterManager() *ClusterManager {
	return &ClusterManager{
		logger: logging.For("demo"),
		env:    localenv.NewKindEnvironment(),
	}
}

// EnsureCluster ensures a Kind cluster exists with the specified configuration
// If the cluster already exists, it returns without error
func (m *ClusterManager) EnsureCluster(ctx context.Context, clusterName string, istioEnabled bool) error {
	// Create config for the demo cluster
	m.config = &localenv.Config{
		ClusterName:  clusterName,
		Namespace:    "demo",
		Port:         8080,
		IstioEnabled: istioEnabled,
	}

	// Check if environment is already ready
	if m.env.IsReady(ctx) {
		m.logger.Info("demo cluster already exists and is ready", "cluster", clusterName)
		return nil
	}

	m.logger.Info("creating demo cluster", "cluster", clusterName, "istio_enabled", istioEnabled)

	// Setup the environment (this will create the cluster if needed)
	if err := m.env.Setup(ctx, m.config); err != nil {
		return fmt.Errorf("failed to setup demo cluster: %w", err)
	}

	m.logger.Info("demo cluster is ready", "cluster", clusterName)
	return nil
}

// DeployScenario deploys a predefined scenario to the demo cluster
func (m *ClusterManager) DeployScenario(ctx context.Context, scenarioName string) error {
	if m.config == nil {
		return fmt.Errorf("cluster not initialized - call EnsureCluster first")
	}

	// Get the scenario by name
	scenario := localenv.GetScenarioByName(scenarioName)
	if scenario == nil {
		return fmt.Errorf("scenario '%s' not found", scenarioName)
	}

	m.logger.Info("deploying scenario", "scenario", scenarioName, "description", scenario.Description)

	// Deploy the scenario
	if err := m.env.DeployScenario(ctx, scenario); err != nil {
		return fmt.Errorf("failed to deploy scenario '%s': %w", scenarioName, err)
	}

	m.logger.Info("scenario deployed successfully", "scenario", scenarioName)
	return nil
}

// GetKubeconfig returns the kubeconfig path for the demo cluster
func (m *ClusterManager) GetKubeconfig() string {
	if m.env == nil {
		return ""
	}
	return m.env.GetKubeconfig()
}

// GetNamespace returns the namespace used for the demo
func (m *ClusterManager) GetNamespace() string {
	if m.env == nil {
		return ""
	}
	return m.env.GetNamespace()
}

// IsReady checks if the demo cluster is ready
func (m *ClusterManager) IsReady(ctx context.Context) bool {
	if m.env == nil {
		return false
	}
	return m.env.IsReady(ctx)
}

// Teardown removes the demo cluster and cleans up resources
func (m *ClusterManager) Teardown(ctx context.Context) error {
	if m.env == nil {
		return nil
	}

	m.logger.Info("tearing down demo cluster")

	if err := m.env.Teardown(ctx); err != nil {
		return fmt.Errorf("failed to teardown demo cluster: %w", err)
	}

	m.logger.Info("demo cluster teardown complete")
	return nil
}

// SetConfig allows setting the config for existing cluster management
func (m *ClusterManager) SetConfig(clusterName string) {
	m.config = &localenv.Config{
		ClusterName: clusterName,
		Namespace:   "demo",
		Port:        8080,
	}
	m.env.SetConfig(m.config)
}

// DemoClusterInfo holds information about a demo cluster
type DemoClusterInfo struct {
	Name         string
	Namespace    string
	Ready        bool
	Kubeconfig   string
	IstioEnabled bool
}

// GetClusterInfo returns information about the current demo cluster
func (m *ClusterManager) GetClusterInfo(ctx context.Context) *DemoClusterInfo {
	if m.config == nil {
		return nil
	}

	return &DemoClusterInfo{
		Name:         m.config.ClusterName,
		Namespace:    m.GetNamespace(),
		Ready:        m.IsReady(ctx),
		Kubeconfig:   m.GetKubeconfig(),
		IstioEnabled: m.config.IstioEnabled,
	}
}
