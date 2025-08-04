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

package microservice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// HelmManager manages Helm operations for microservice installation
type HelmManager struct {
	kubeconfig   string
	namespace    string
	logger       *slog.Logger
	actionConfig *action.Configuration
}

// MicroserviceInstallConfig defines configuration for microservice installation
type MicroserviceInstallConfig struct {
	Namespace    string
	ReleaseName  string
	CustomValues map[string]interface{}
	WaitTimeout  time.Duration
}

// ChartConfig defines configuration for individual chart installations
type ChartConfig struct {
	ReleaseName string
	Values      map[string]interface{}
	Timeout     time.Duration
	Wait        bool
	Atomic      bool
}

// NewHelmManager creates a new Helm manager instance for microservices
func NewHelmManager(kubeconfig, namespace string, logger *slog.Logger) (*HelmManager, error) {
	if logger == nil {
		logger = slog.Default()
	}

	h := &HelmManager{
		kubeconfig: kubeconfig,
		namespace:  namespace,
		logger:     logger,
	}

	// Initialize Helm action configuration
	actionConfig := new(action.Configuration)

	// Create CLI environment
	settings := cli.New()
	if kubeconfig != "" {
		settings.KubeConfig = kubeconfig
	}

	// Initialize the action configuration
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secret",
		func(format string, v ...interface{}) {
			h.logger.Debug(fmt.Sprintf(format, v...))
		}); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action config: %w", err)
	}

	h.actionConfig = actionConfig
	return h, nil
}

// DefaultMicroserviceConfig returns default configuration for microservice installation
func DefaultMicroserviceConfig() MicroserviceInstallConfig {
	return MicroserviceInstallConfig{
		Namespace:    "default",
		ReleaseName:  "microservice",
		CustomValues: map[string]interface{}{},
		WaitTimeout:  2 * time.Minute,
	}
}

// InstallMicroservice installs microservice chart with specified configuration
func (h *HelmManager) InstallMicroservice(ctx context.Context, config MicroserviceInstallConfig) error {
	h.logger.Info("Starting microservice installation",
		"namespace", config.Namespace,
		"release", config.ReleaseName)

	// Create namespace if it doesn't exist
	if err := h.ensureNamespace(ctx, config.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace exists: %w", err)
	}

	chartConfig := ChartConfig{
		ReleaseName: config.ReleaseName,
		Values:      config.CustomValues,
		Timeout:     config.WaitTimeout,
		Wait:        true,
		Atomic:      true,
	}

	if err := h.installChart(ctx, "microservice", chartConfig); err != nil {
		return fmt.Errorf("failed to install microservice: %w", err)
	}

	h.logger.Info("Microservice installation completed successfully")
	return nil
}

// UninstallMicroservice uninstalls microservice chart
func (h *HelmManager) UninstallMicroservice(ctx context.Context, releaseName string) error {
	h.logger.Info("Starting microservice uninstallation", "release", releaseName)

	if err := h.uninstallChart(ctx, releaseName); err != nil {
		return fmt.Errorf("failed to uninstall microservice: %w", err)
	}

	h.logger.Info("Microservice uninstallation completed")
	return nil
}

// IsMicroserviceInstalled checks if microservice is installed and returns version information
func (h *HelmManager) IsMicroserviceInstalled(ctx context.Context, releaseName string) (bool, string, error) {
	listAction := action.NewList(h.actionConfig)
	listAction.All = true

	releases, err := listAction.Run()
	if err != nil {
		return false, "", fmt.Errorf("failed to list releases: %w", err)
	}

	for _, release := range releases {
		if release.Name == releaseName {
			version := release.Chart.Metadata.Version
			return true, version, nil
		}
	}

	return false, "", nil
}

// installChart installs a single chart
func (h *HelmManager) installChart(ctx context.Context, chartName string, config ChartConfig) error {
	h.logger.Info("Starting chart installation", "chart", chartName, "release", config.ReleaseName)

	// Check if already installed
	h.logger.Debug("Checking if chart is already installed", "release", config.ReleaseName)
	if installed, err := h.isChartInstalled(ctx, config.ReleaseName); err != nil {
		return fmt.Errorf("failed to check if chart is installed: %w", err)
	} else if installed {
		h.logger.Info("Chart already installed, skipping", "release", config.ReleaseName)
		return nil
	}

	// Load chart from embedded FS
	h.logger.Debug("Loading chart from embedded FS", "chart", chartName)
	chart, err := h.loadChart(chartName)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}
	h.logger.Debug("Chart loaded successfully", "chart", chartName)

	// Create install action
	h.logger.Debug("Creating Helm install action", "release", config.ReleaseName, "namespace", h.namespace)
	installAction := action.NewInstall(h.actionConfig)
	installAction.ReleaseName = config.ReleaseName
	installAction.Namespace = h.namespace
	installAction.Wait = config.Wait
	installAction.Timeout = config.Timeout
	installAction.Atomic = config.Atomic

	// Log the actual namespace being used for debugging
	h.logger.Info("Installing to namespace", "namespace", installAction.Namespace, "helmManager.namespace", h.namespace)

	// Install the chart
	h.logger.Info("Installing chart with Helm", "chart", chartName, "release", config.ReleaseName, "wait", config.Wait, "timeout", config.Timeout)
	_, err = installAction.RunWithContext(ctx, chart, config.Values)
	if err != nil {
		return fmt.Errorf("failed to install chart %s: %w", chartName, err)
	}

	h.logger.Info("Chart installation completed", "chart", chartName, "release", config.ReleaseName)
	return nil
}

// uninstallChart uninstalls a single chart
func (h *HelmManager) uninstallChart(ctx context.Context, releaseName string) error {
	// Check if installed
	if installed, err := h.isChartInstalled(ctx, releaseName); err != nil {
		return fmt.Errorf("failed to check if chart is installed: %w", err)
	} else if !installed {
		h.logger.Info("Chart not installed, skipping", "release", releaseName)
		return nil
	}

	// Create uninstall action
	uninstallAction := action.NewUninstall(h.actionConfig)

	// Uninstall the chart
	_, err := uninstallAction.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall chart %s: %w", releaseName, err)
	}

	return nil
}

// isChartInstalled checks if a chart is installed
func (h *HelmManager) isChartInstalled(ctx context.Context, releaseName string) (bool, error) {
	getAction := action.NewGet(h.actionConfig)

	_, err := getAction.Run(releaseName)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// loadChart loads a chart from the local filesystem
func (h *HelmManager) loadChart(chartName string) (*chart.Chart, error) {
	// Load our wrapper chart from the local filesystem
	chartPath := "./pkg/localenv/microservice/chart"
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load microservice chart: %w", err)
	}
	return chart, nil
}

// ensureNamespace creates the namespace if it doesn't exist
func (h *HelmManager) ensureNamespace(ctx context.Context, namespace string) error {
	// For now, assume namespace exists or will be created by Helm
	// In a full implementation, you'd use the Kubernetes client to create it
	h.logger.Debug("Namespace handling delegated to Helm", "namespace", namespace)
	return nil
}
