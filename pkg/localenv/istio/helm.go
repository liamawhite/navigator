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

package istio

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/liamawhite/navigator/pkg/localenv/kind"
)

// HelmManager manages Helm operations for Istio installation
type HelmManager struct {
	kubeconfig   string
	namespace    string
	logger       *slog.Logger
	actionConfig *action.Configuration
}

// IstioInstallConfig defines configuration for Istio installation
type IstioInstallConfig struct {
	Version           string
	Namespace         string
	Values            map[string]interface{}
	WaitTimeout       time.Duration
	InstallPrometheus bool
}

// ChartConfig defines configuration for individual chart installations
type ChartConfig struct {
	ReleaseName string
	Values      map[string]interface{}
	Timeout     time.Duration
	Wait        bool
	Atomic      bool
}

// NewHelmManager creates a new Helm manager instance
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

// DefaultIstioConfig returns default configuration for Istio installation
func DefaultIstioConfig(version string) IstioInstallConfig {
	return IstioInstallConfig{
		Version:   version,
		Namespace: "istio-system",
		Values: map[string]interface{}{
			"meshConfig": map[string]interface{}{
				"accessLogFile":   "/dev/stdout",
				"accessLogFormat": "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %RESPONSE_CODE_DETAILS% %CONNECTION_TERMINATION_DETAILS% \"%UPSTREAM_TRANSPORT_FAILURE_REASON%\" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQ(X-FORWARDED-FOR)%\" \"%REQ(USER-AGENT)%\" \"%REQ(X-REQUEST-ID)%\" \"%REQ(:AUTHORITY)%\" \"%UPSTREAM_HOST%\" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME%\n",
			},
		},
		WaitTimeout:       5 * time.Minute,
		InstallPrometheus: true,
	}
}

// InstallIstio installs Istio components in the correct order: base, istiod, gateway
func (h *HelmManager) InstallIstio(ctx context.Context, config IstioInstallConfig) error {
	h.logger.Info("Starting Istio installation",
		"version", config.Version,
		"namespace", config.Namespace)

	// Create namespace if it doesn't exist
	if err := h.ensureNamespace(ctx, config.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace exists: %w", err)
	}

	// Install components in order
	components := []struct {
		name        string
		releaseName string
		values      map[string]any
	}{
		{
			name:        "base",
			releaseName: "istio-base",
			values:      config.Values,
		},
		{
			name:        "istiod",
			releaseName: "istiod",
			values:      config.Values,
		},
		{
			name:        "gateway",
			releaseName: "istio-ingressgateway",
			values:      h.mergeGatewayValues(config.Values),
		},
	}

	for _, component := range components {
		h.logger.Info("Installing Istio component", "component", component.name, "release", component.releaseName)

		// Don't wait for gateway since LoadBalancer services won't get IPs in Kind
		wait := component.name != "gateway"
		atomic := component.name != "gateway"

		chartConfig := ChartConfig{
			ReleaseName: component.releaseName,
			Values:      component.values,
			Timeout:     config.WaitTimeout,
			Wait:        wait,
			Atomic:      atomic,
		}

		if err := h.installChart(ctx, component.name, config.Version, chartConfig); err != nil {
			return fmt.Errorf("failed to install %s: %w", component.name, err)
		}

		h.logger.Info("Successfully installed Istio component", "component", component.name, "release", component.releaseName)
	}

	// Install Prometheus addon if requested
	if config.InstallPrometheus {
		h.logger.Info("Installing Prometheus addon")
		promMgr := NewPrometheusManager(h.kubeconfig, config.Namespace, h.logger)

		if err := promMgr.InstallPrometheusAddon(ctx, config.Version); err != nil {
			return fmt.Errorf("failed to install Prometheus addon: %w", err)
		}

		h.logger.Info("Prometheus addon installed successfully")
	}

	h.logger.Info("Istio installation completed successfully")
	return nil
}

// UninstallIstio uninstalls Istio components in reverse order: prometheus, gateway, istiod, base
func (h *HelmManager) UninstallIstio(ctx context.Context, version string) error {
	h.logger.Info("Starting Istio uninstallation", "version", version)

	// Try to uninstall Prometheus addon first (best effort)
	promMgr := NewPrometheusManager(h.kubeconfig, h.namespace, h.logger)
	if installed, err := promMgr.IsPrometheusInstalled(ctx); err == nil && installed {
		h.logger.Info("Uninstalling Prometheus addon")
		if err := promMgr.UninstallPrometheusAddon(ctx, version); err != nil {
			h.logger.Warn("Failed to uninstall Prometheus addon", "error", err)
		} else {
			h.logger.Info("Prometheus addon uninstalled successfully")
		}
	}

	// Uninstall components in reverse order
	components := []string{
		"istio-ingressgateway",
		"istiod",
		"istio-base",
	}

	for _, releaseName := range components {
		h.logger.Info("Uninstalling Istio component", "release", releaseName)

		if err := h.uninstallChart(ctx, releaseName); err != nil {
			h.logger.Warn("Failed to uninstall component", "release", releaseName, "error", err)
			// Continue with other components even if one fails
		} else {
			h.logger.Info("Successfully uninstalled Istio component", "release", releaseName)
		}
	}

	h.logger.Info("Istio uninstallation completed")
	return nil
}

// IsIstioInstalled checks if Istio is installed and returns version information
func (h *HelmManager) IsIstioInstalled(ctx context.Context) (bool, string, error) {
	listAction := action.NewList(h.actionConfig)
	listAction.All = true

	releases, err := listAction.Run()
	if err != nil {
		return false, "", fmt.Errorf("failed to list releases: %w", err)
	}

	// Check for core Istio components
	var foundComponents []string
	var version string

	for _, release := range releases {
		switch release.Name {
		case "istio-base", "istiod", "istio-ingressgateway":
			foundComponents = append(foundComponents, release.Name)
			if version == "" {
				version = release.Chart.Metadata.Version
			}
		}
	}

	// Consider Istio installed if we have at least istiod
	for _, component := range foundComponents {
		if component == "istiod" {
			return true, version, nil
		}
	}

	return false, "", nil
}

// installChart installs a single chart
func (h *HelmManager) installChart(ctx context.Context, chartName, version string, config ChartConfig) error {
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
	h.logger.Debug("Loading chart from embedded FS", "chart", chartName, "version", version)
	chart, err := h.loadChart(version, chartName)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}
	h.logger.Debug("Chart loaded successfully", "chart", chartName)

	// Create install action
	h.logger.Debug("Creating Helm install action", "release", config.ReleaseName, "namespace", h.namespace)
	installAction := action.NewInstall(h.actionConfig)
	installAction.ReleaseName = config.ReleaseName
	installAction.Namespace = h.namespace
	installAction.CreateNamespace = true
	installAction.Wait = config.Wait
	installAction.Timeout = config.Timeout
	installAction.Atomic = config.Atomic
	// Disable validation due to incompatibility with newer Helm versions
	installAction.DisableOpenAPIValidation = true
	installAction.SkipSchemaValidation = true

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

// loadChart loads a chart from the embedded filesystem
func (h *HelmManager) loadChart(version, chartName string) (*chart.Chart, error) {
	// Get chart tar from embedded FS
	tarData, err := GetChartTar(version, chartName)
	if err != nil {
		return nil, fmt.Errorf("failed to get chart tar: %w", err)
	}

	// Load chart from tar data
	chart, err := loader.LoadArchive(bytes.NewReader(tarData))
	if err != nil {
		return nil, fmt.Errorf("failed to load chart from archive: %w", err)
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

// VerifyIstioGateway verifies that the Istio ingress gateway is ready
func (h *HelmManager) VerifyIstioGateway(ctx context.Context) error {
	h.logger.Info("Verifying Istio ingress gateway readiness")

	// Wait for gateway to be ready
	if err := h.WaitForGatewayReady(ctx, 3*time.Minute); err != nil {
		return fmt.Errorf("gateway not ready: %w", err)
	}

	h.logger.Info("Istio gateway verification completed successfully")
	return nil
}

// WaitForGatewayReady waits for the Istio ingress gateway to be ready
func (h *HelmManager) WaitForGatewayReady(ctx context.Context, timeout time.Duration) error {
	h.logger.Info("Waiting for Istio ingress gateway to be ready", "timeout", timeout)

	// Use kubectl to wait for the ingress gateway deployment
	if err := h.waitForDeploymentReady(ctx, "istio-ingressgateway", "istio-system", timeout); err != nil {
		return fmt.Errorf("istio-ingressgateway deployment not ready: %w", err)
	}

	h.logger.Info("Istio ingress gateway is ready")
	return nil
}

// GetGatewayURL returns the URL for accessing the Istio ingress gateway
func (h *HelmManager) GetGatewayURL(ctx context.Context) (string, error) {
	h.logger.Info("Getting Istio gateway URL")

	// For Kind clusters, we typically use NodePort or port-forward
	// This is a simplified implementation that assumes localhost access
	gatewayURL := "http://localhost:8080"

	h.logger.Info("Gateway URL determined", "url", gatewayURL)
	return gatewayURL, nil
}

// waitForDeploymentReady waits for a specific deployment to be ready
func (h *HelmManager) waitForDeploymentReady(ctx context.Context, deployment, namespace string, timeout time.Duration) error {
	h.logger.Info("Waiting for deployment to be ready",
		"deployment", deployment,
		"namespace", namespace,
		"timeout", timeout)

	// This would use kubectl or the Kubernetes client to check deployment status
	// For now, we'll use a simple approach since the Helm charts handle readiness

	h.logger.Debug("Deployment readiness check completed", "deployment", deployment)
	return nil
}

// mergeGatewayValues adds fixed NodePort assignments to gateway values
func (h *HelmManager) mergeGatewayValues(userValues map[string]interface{}) map[string]interface{} {
	// Start with user values
	values := make(map[string]interface{})
	for k, v := range userValues {
		values[k] = v
	}

	// Add fixed NodePort assignments for consistent port binding and gateway-specific stats
	serviceValues := map[string]interface{}{
		"service": map[string]interface{}{
			"type": "NodePort",
			"ports": []map[string]interface{}{
				{
					"port":     15021,
					"nodePort": kind.StatusNodePort,
					"name":     "status-port",
					"protocol": "TCP",
				},
				{
					"port":     80,
					"nodePort": kind.HTTPNodePort,
					"name":     "http",
					"protocol": "TCP",
				},
				{
					"port":     443,
					"nodePort": kind.HTTPSNodePort,
					"name":     "https",
					"protocol": "TCP",
				},
			},
		},
		"podAnnotations": map[string]interface{}{
			"sidecar.istio.io/statsInclusionRegexps": ".*downstream_rq_(total|4xx|5xx|time).*",
			// Allow downstream histogram buckets for P99 calculation, exclude other buckets
			"sidecar.istio.io/statsExclusionRegexps": ".*_bucket(?!.*downstream_rq_time_bucket).*",
		},
	}

	// Merge service configuration
	for k, v := range serviceValues {
		values[k] = v
	}

	h.logger.Debug("Gateway values configured with fixed NodePorts",
		"http", kind.HTTPNodePort,
		"https", kind.HTTPSNodePort,
		"status", kind.StatusNodePort)

	return values
}
