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

package kind

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
)

// Fixed NodePort assignments for Kind cluster demo configuration
// These ports are bound 1:1 from host to Kind container for direct access
const (
	// HTTPNodePort is the fixed NodePort for HTTP traffic (port 80)
	HTTPNodePort = 30080
	// HTTPSNodePort is the fixed NodePort for HTTPS traffic (port 443)
	HTTPSNodePort = 30443
	// StatusNodePort is the fixed NodePort for status/health checks (port 15021)
	StatusNodePort = 31021
	// PrometheusNodePort is the fixed NodePort for Prometheus metrics UI access (port 9090)
	PrometheusNodePort = 30090
)

type KindManager struct {
	provider *cluster.Provider
	logger   *slog.Logger
}

type KindClusterConfig struct {
	Name            string
	Image           string
	KubeVersion     string
	ConfigPath      string
	ExtraMounts     []string
	ExtraPortMaps   []string
	DisableDefaults bool
}

func NewKindManager(logger *slog.Logger) *KindManager {
	if logger == nil {
		logger = slog.Default()
	}

	return &KindManager{
		provider: cluster.NewProvider(),
		logger:   logger,
	}
}

func (k *KindManager) CreateCluster(ctx context.Context, config KindClusterConfig) error {
	k.logger.Info("Creating Kind cluster", "name", config.Name)

	var createOptions []cluster.CreateOption

	if config.Image != "" {
		createOptions = append(createOptions, cluster.CreateWithNodeImage(config.Image))
	}

	// Create Kind config file if we need port mappings
	if len(config.ExtraPortMaps) > 0 {
		configPath, err := k.createKindConfigFile(config)
		if err != nil {
			return fmt.Errorf("failed to create Kind config file: %w", err)
		}
		createOptions = append(createOptions, cluster.CreateWithConfigFile(configPath))
		k.logger.Debug("Using generated Kind config", "path", configPath)
	} else if config.ConfigPath != "" {
		createOptions = append(createOptions, cluster.CreateWithConfigFile(config.ConfigPath))
	}

	createOptions = append(createOptions, cluster.CreateWithDisplayUsage(true))
	createOptions = append(createOptions, cluster.CreateWithDisplaySalutation(true))

	if err := k.provider.Create(config.Name, createOptions...); err != nil {
		k.logger.Error("Failed to create Kind cluster", "name", config.Name, "error", err)
		return fmt.Errorf("failed to create Kind cluster %s: %w", config.Name, err)
	}

	k.logger.Info("Successfully created Kind cluster", "name", config.Name)
	return nil
}

func (k *KindManager) DeleteCluster(ctx context.Context, name string) error {
	k.logger.Info("Deleting Kind cluster", "name", name)

	if err := k.provider.Delete(name, ""); err != nil {
		k.logger.Error("Failed to delete Kind cluster", "name", name, "error", err)
		return fmt.Errorf("failed to delete Kind cluster %s: %w", name, err)
	}

	k.logger.Info("Successfully deleted Kind cluster", "name", name)
	return nil
}

func (k *KindManager) ListClusters(ctx context.Context) ([]string, error) {
	k.logger.Debug("Listing Kind clusters")

	clusters, err := k.provider.List()
	if err != nil {
		k.logger.Error("Failed to list Kind clusters", "error", err)
		return nil, fmt.Errorf("failed to list Kind clusters: %w", err)
	}

	k.logger.Debug("Found Kind clusters", "count", len(clusters), "clusters", clusters)
	return clusters, nil
}

func (k *KindManager) ClusterExists(ctx context.Context, name string) (bool, error) {
	clusters, err := k.ListClusters(ctx)
	if err != nil {
		return false, err
	}

	for _, cluster := range clusters {
		if cluster == name {
			return true, nil
		}
	}

	return false, nil
}

func (k *KindManager) GetKubeconfig(ctx context.Context, name string, internal bool) (string, error) {
	k.logger.Debug("Getting kubeconfig for Kind cluster", "name", name, "internal", internal)

	kubeconfig, err := k.provider.KubeConfig(name, internal)
	if err != nil {
		k.logger.Error("Failed to get kubeconfig for Kind cluster", "name", name, "error", err)
		return "", fmt.Errorf("failed to get kubeconfig for Kind cluster %s: %w", name, err)
	}

	return kubeconfig, nil
}

func (k *KindManager) ExportKubeconfig(ctx context.Context, name, path string) error {
	k.logger.Info("Exporting kubeconfig for Kind cluster", "name", name, "path", path)

	if err := k.provider.ExportKubeConfig(name, path, false); err != nil {
		k.logger.Error("Failed to export kubeconfig for Kind cluster", "name", name, "path", path, "error", err)
		return fmt.Errorf("failed to export kubeconfig for Kind cluster %s to %s: %w", name, path, err)
	}

	k.logger.Info("Successfully exported kubeconfig", "name", name, "path", path)
	return nil
}

func DefaultKindConfig(name string) KindClusterConfig {
	return KindClusterConfig{
		Name:            name,
		Image:           "",
		KubeVersion:     "",
		ConfigPath:      "",
		ExtraMounts:     []string{},
		ExtraPortMaps:   []string{},
		DisableDefaults: false,
	}
}

// DemoKindConfig returns a Kind configuration suitable for demo clusters with NodePort access
func DemoKindConfig(name string) KindClusterConfig {
	// Bind the specific fixed ports for Istio gateway access and Prometheus metrics
	portMaps := []string{
		fmt.Sprintf("%d:%d", HTTPNodePort, HTTPNodePort),             // HTTP - Microservices via Istio gateway
		fmt.Sprintf("%d:%d", HTTPSNodePort, HTTPSNodePort),           // HTTPS - Microservices via Istio gateway
		fmt.Sprintf("%d:%d", StatusNodePort, StatusNodePort),         // Status - Istio gateway health checks
		fmt.Sprintf("%d:%d", PrometheusNodePort, PrometheusNodePort), // Prometheus - Metrics UI access
	}

	return KindClusterConfig{
		Name:            name,
		Image:           "",
		KubeVersion:     "",
		ConfigPath:      "",
		ExtraMounts:     []string{},
		ExtraPortMaps:   portMaps,
		DisableDefaults: false,
	}
}

func (k *KindManager) createKindConfigFile(config KindClusterConfig) (string, error) {
	configYAML := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:`

	for _, portMap := range config.ExtraPortMaps {
		parts := strings.Split(portMap, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid port mapping format: %s", portMap)
		}

		hostPort := parts[0]
		containerPort := parts[1]

		configYAML += fmt.Sprintf(`
  - containerPort: %s
    hostPort: %s
    protocol: TCP`, containerPort, hostPort)
	}

	// Create temporary config file
	tempDir := filepath.Join(os.TempDir(), "kind-configs")
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	configPath := filepath.Join(tempDir, fmt.Sprintf("%s-config.yaml", config.Name))
	if err := os.WriteFile(configPath, []byte(configYAML), 0600); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	k.logger.Debug("Created Kind config file", "path", configPath)
	return configPath, nil
}

func (k *KindManager) WaitForClusterReady(ctx context.Context, name string) error {
	k.logger.Info("Waiting for Kind cluster to be ready", "name", name)

	exists, err := k.ClusterExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("cluster %s does not exist", name)
	}

	timeout := 5 * time.Minute
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			k.logger.Error("Timeout waiting for Kind cluster to be ready", "name", name)
			return fmt.Errorf("timeout waiting for Kind cluster %s to be ready", name)
		case <-ticker.C:
			kubeconfig, err := k.GetKubeconfig(ctx, name, false)
			if err == nil && kubeconfig != "" {
				k.logger.Info("Kind cluster is ready", "name", name)
				return nil
			}
			k.logger.Debug("Cluster not yet ready, retrying...", "name", name)
		}
	}
}
