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

package fortio

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:embed manifests/*.yaml
var manifestFS embed.FS

const (
	fortioNamespace = "load-generator"
	fortioPodName   = "fortio-load"
)

// FortioManager manages Fortio load generation
type FortioManager struct {
	kubeconfig string
	namespace  string
	logger     *slog.Logger
}

// NewFortioManager creates a new Fortio manager instance
func NewFortioManager(kubeconfig, namespace string, logger *slog.Logger) *FortioManager {
	if logger == nil {
		logger = slog.Default()
	}
	if namespace == "" {
		namespace = fortioNamespace
	}

	return &FortioManager{
		kubeconfig: kubeconfig,
		namespace:  namespace,
		logger:     logger,
	}
}

// InstallFortio deploys the Fortio load generation pod
func (f *FortioManager) InstallFortio(ctx context.Context) error {
	f.logger.Info("Installing Fortio load generator", "namespace", f.namespace)

	// Get the namespace manifest file path
	namespaceManifestPath, err := f.getNamespaceManifestPath()
	if err != nil {
		return fmt.Errorf("failed to get namespace manifest path: %w", err)
	}

	// Apply the namespace manifest first
	if err := f.applyManifest(ctx, namespaceManifestPath); err != nil {
		return fmt.Errorf("failed to apply namespace manifest: %w", err)
	}

	// Get the Fortio manifest file path
	fortioManifestPath, err := f.getFortioManifestPath()
	if err != nil {
		return fmt.Errorf("failed to get Fortio manifest path: %w", err)
	}

	// Apply the Fortio manifest using kubectl
	if err := f.applyManifest(ctx, fortioManifestPath); err != nil {
		return fmt.Errorf("failed to apply Fortio manifest: %w", err)
	}

	f.logger.Info("Fortio load generator installed successfully", "namespace", f.namespace)
	return nil
}

// UninstallFortio removes the Fortio load generation pod
func (f *FortioManager) UninstallFortio(ctx context.Context) error {
	f.logger.Info("Uninstalling Fortio load generator", "namespace", f.namespace)

	// Delete the pod using kubectl
	// #nosec G204 -- fortioPodName and f.namespace are constants/validated inputs
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "pod", fortioPodName, "-n", f.namespace, "--ignore-not-found=true")
	if f.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", f.kubeconfig}, cmd.Args[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete Fortio pod: %w, output: %s", err, string(output))
	}

	f.logger.Info("Fortio load generator uninstalled successfully", "namespace", f.namespace)
	return nil
}

// IsFortioRunning checks if the Fortio pod is running
func (f *FortioManager) IsFortioRunning(ctx context.Context) (bool, error) {
	f.logger.Debug("Checking if Fortio is running", "namespace", f.namespace)

	// Check for Fortio pod
	// #nosec G204 -- fortioPodName and f.namespace are constants/validated inputs
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pod", fortioPodName, "-n", f.namespace, "--no-headers")
	if f.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", f.kubeconfig}, cmd.Args[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the command fails, Fortio is likely not running
		f.logger.Debug("Fortio pod not found", "error", err, "output", string(output))
		return false, nil
	}

	// Check if the pod is in Running state
	outputStr := string(output)
	isRunning := strings.Contains(outputStr, "Running")

	f.logger.Debug("Fortio pod status", "output", outputStr, "running", isRunning)
	return isRunning, nil
}

// WaitForFortioReady waits for the Fortio pod to become ready
func (f *FortioManager) WaitForFortioReady(ctx context.Context, timeout time.Duration) error {
	f.logger.Info("Waiting for Fortio to be ready", "timeout", timeout, "namespace", f.namespace)

	// Wait for the pod to be ready
	// #nosec G204 -- fortioPodName and f.namespace are constants/validated inputs
	cmd := exec.CommandContext(ctx, "kubectl", "wait", "--for=condition=ready",
		fmt.Sprintf("pod/%s", fortioPodName), "-n", f.namespace, fmt.Sprintf("--timeout=%s", timeout))

	if f.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", f.kubeconfig}, cmd.Args[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fortio pod not ready within timeout: %w, output: %s", err, string(output))
	}

	f.logger.Info("Fortio is ready and generating load", "namespace", f.namespace)
	return nil
}

// GetFortioLogs returns the logs from the Fortio pod
func (f *FortioManager) GetFortioLogs(ctx context.Context) (string, error) {
	// #nosec G204 -- fortioPodName and f.namespace are constants/validated inputs
	cmd := exec.CommandContext(ctx, "kubectl", "logs", fortioPodName, "-n", f.namespace)
	if f.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", f.kubeconfig}, cmd.Args[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Fortio logs: %w", err)
	}

	return string(output), nil
}

// getNamespaceManifestPath returns the path to a temporary file containing the namespace manifest
func (f *FortioManager) getNamespaceManifestPath() (string, error) {
	return f.writeManifestToTempFile("manifests/load-generator-namespace.yaml", "namespace-*.yaml")
}

// getFortioManifestPath returns the path to a temporary file containing the Fortio manifest
func (f *FortioManager) getFortioManifestPath() (string, error) {
	return f.writeManifestToTempFile("manifests/fortio.yaml", "fortio-*.yaml")
}

// writeManifestToTempFile reads a manifest from the embedded filesystem and writes it to a temporary file
func (f *FortioManager) writeManifestToTempFile(embedPath, tempPattern string) (string, error) {
	// Read manifest from embedded filesystem
	data, err := manifestFS.ReadFile(embedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded manifest %s: %w", embedPath, err)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", tempPattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			f.logger.Warn("Failed to close temporary file", "path", tempFile.Name(), "error", err)
		}
	}()

	// Write manifest data to temporary file
	if _, err := tempFile.Write(data); err != nil {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			f.logger.Warn("Failed to remove temporary file after write error", "path", tempFile.Name(), "error", removeErr)
		}
		return "", fmt.Errorf("failed to write manifest to temporary file: %w", err)
	}

	f.logger.Debug("Created temporary manifest file", "path", tempFile.Name(), "source", embedPath)
	return tempFile.Name(), nil
}

// applyManifest applies a Kubernetes manifest using kubectl
func (f *FortioManager) applyManifest(ctx context.Context, manifestPath string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", manifestPath)
	if f.kubeconfig != "" {
		cmd.Args = append([]string{"kubectl", "--kubeconfig", f.kubeconfig}, cmd.Args[1:]...)
	}

	f.logger.Debug("Applying Fortio manifest", "command", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %w, output: %s", err, string(output))
	}

	f.logger.Debug("Fortio manifest applied successfully", "output", string(output))
	return nil
}
