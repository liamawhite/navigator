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
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	prometheusNamespace = "istio-system"
)

// PrometheusManager manages Prometheus addon installation
type PrometheusManager struct {
	kubeconfig string
	namespace  string
	logger     *slog.Logger
}

// NewPrometheusManager creates a new Prometheus manager instance
func NewPrometheusManager(kubeconfig, namespace string, logger *slog.Logger) *PrometheusManager {
	if logger == nil {
		logger = slog.Default()
	}
	if namespace == "" {
		namespace = prometheusNamespace
	}

	return &PrometheusManager{
		kubeconfig: kubeconfig,
		namespace:  namespace,
		logger:     logger,
	}
}

// InstallPrometheusAddon installs the Prometheus addon using the embedded manifest
func (p *PrometheusManager) InstallPrometheusAddon(ctx context.Context, version string) error {
	p.logger.Info("Installing Prometheus addon", "version", version, "namespace", p.namespace)

	// Get the Prometheus manifest from embedded files
	manifestData, err := GetPrometheusManifest(version)
	if err != nil {
		return fmt.Errorf("failed to get Prometheus manifest: %w", err)
	}

	// Create temporary file for the manifest
	tempFile, err := os.CreateTemp("", "prometheus-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			p.logger.Warn("Failed to remove temporary file", "file", tempFile.Name(), "error", removeErr)
		}
	}()

	// Write manifest to temporary file
	if _, err := tempFile.Write(manifestData); err != nil {
		return fmt.Errorf("failed to write manifest to temporary file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Apply the manifest using kubectl
	if err := p.applyManifest(ctx, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to apply Prometheus manifest: %w", err)
	}

	p.logger.Info("Prometheus addon installed successfully", "namespace", p.namespace)
	return nil
}

// UninstallPrometheusAddon uninstalls the Prometheus addon
func (p *PrometheusManager) UninstallPrometheusAddon(ctx context.Context, version string) error {
	p.logger.Info("Uninstalling Prometheus addon", "version", version, "namespace", p.namespace)

	// Get the Prometheus manifest from embedded files
	manifestData, err := GetPrometheusManifest(version)
	if err != nil {
		return fmt.Errorf("failed to get Prometheus manifest: %w", err)
	}

	// Create temporary file for the manifest
	tempFile, err := os.CreateTemp("", "prometheus-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			p.logger.Warn("Failed to remove temporary file", "file", tempFile.Name(), "error", removeErr)
		}
	}()

	// Write manifest to temporary file
	if _, err := tempFile.Write(manifestData); err != nil {
		return fmt.Errorf("failed to write manifest to temporary file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Delete the manifest using kubectl
	if err := p.deleteManifest(ctx, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to delete Prometheus manifest: %w", err)
	}

	p.logger.Info("Prometheus addon uninstalled successfully", "namespace", p.namespace)
	return nil
}

// IsPrometheusInstalled checks if Prometheus is installed in the cluster
func (p *PrometheusManager) IsPrometheusInstalled(ctx context.Context) (bool, error) {
	p.logger.Debug("Checking if Prometheus is installed", "namespace", p.namespace)

	// Check for Prometheus deployment
	// #nosec G204 -- deployment name and p.namespace are validated inputs
	cmd := exec.CommandContext(ctx, "kubectl", "get", "deployment", "prometheus", "-n", p.namespace)
	if p.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", p.kubeconfig}, cmd.Args[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the command fails, Prometheus is likely not installed
		p.logger.Debug("Prometheus deployment not found", "error", err, "output", string(output))
		return false, nil
	}

	p.logger.Debug("Prometheus deployment found", "output", string(output))
	return true, nil
}

// WaitForPrometheusReady waits for Prometheus to become ready
func (p *PrometheusManager) WaitForPrometheusReady(ctx context.Context, timeout time.Duration) error {
	p.logger.Info("Waiting for Prometheus to be ready", "timeout", timeout, "namespace", p.namespace)

	// Wait for the deployment to be ready
	// #nosec G204 -- deployment name and p.namespace are validated inputs
	cmd := exec.CommandContext(ctx, "kubectl", "wait", "--for=condition=available",
		"deployment/prometheus", "-n", p.namespace, fmt.Sprintf("--timeout=%s", timeout))

	if p.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", p.kubeconfig}, cmd.Args[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("prometheus deployment not ready within timeout: %w, output: %s", err, string(output))
	}

	p.logger.Info("Prometheus is ready", "namespace", p.namespace)
	return nil
}

// applyManifest applies a Kubernetes manifest using kubectl
func (p *PrometheusManager) applyManifest(ctx context.Context, manifestPath string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", manifestPath)
	if p.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", p.kubeconfig}, cmd.Args[1:]...)
	}

	p.logger.Debug("Applying Prometheus manifest", "command", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %w, output: %s", err, string(output))
	}

	p.logger.Debug("Prometheus manifest applied successfully", "output", string(output))
	return nil
}

// deleteManifest deletes a Kubernetes manifest using kubectl
func (p *PrometheusManager) deleteManifest(ctx context.Context, manifestPath string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "-f", manifestPath, "--ignore-not-found=true")
	if p.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", p.kubeconfig}, cmd.Args[1:]...)
	}

	p.logger.Debug("Deleting Prometheus manifest", "command", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl delete failed: %w, output: %s", err, string(output))
	}

	p.logger.Debug("Prometheus manifest deleted successfully", "output", string(output))
	return nil
}

// GetPrometheusURL returns the URL for accessing Prometheus
func (p *PrometheusManager) GetPrometheusURL(ctx context.Context) (string, error) {
	p.logger.Info("Getting Prometheus URL", "namespace", p.namespace)

	// For Kind clusters, we typically use port-forward or NodePort
	// This is a simplified implementation that assumes port-forward access
	prometheusURL := "http://localhost:9090"

	p.logger.Info("Prometheus URL determined", "url", prometheusURL)
	return prometheusURL, nil
}
