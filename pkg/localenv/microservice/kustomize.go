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
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed manifests/*
var manifestsFS embed.FS

// KustomizeManager manages Kustomize operations for microservice installation
type KustomizeManager struct {
	kubeconfig string
	logger     *slog.Logger
}

// NewKustomizeManager creates a new Kustomize manager instance for microservices
func NewKustomizeManager(kubeconfig string, logger *slog.Logger) (*KustomizeManager, error) {
	if logger == nil {
		logger = slog.Default()
	}

	k := &KustomizeManager{
		kubeconfig: kubeconfig,
		logger:     logger,
	}

	// Check if kubectl is available
	if err := k.checkKubectl(); err != nil {
		return nil, fmt.Errorf("kubectl not available: %w", err)
	}

	return k, nil
}

// InstallMicroservice installs microservice manifests
func (k *KustomizeManager) InstallMicroservice(ctx context.Context) error {
	k.logger.Info("Starting microservice installation", "namespace", "microservice-demo")

	// Create temporary directory for manifests
	tempDir, err := os.MkdirTemp("", "navigator-microservice-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			k.logger.Warn("Failed to cleanup temp directory", "error", cleanupErr, "path", tempDir)
		}
	}()

	// Extract manifests to temporary directory
	if err := k.extractManifests(tempDir); err != nil {
		return fmt.Errorf("failed to extract manifests: %w", err)
	}

	// Use the namespace from kustomization.yaml (microservice-demo)

	// Apply manifests using kubectl with 2 minute timeout
	if err := k.applyManifests(ctx, tempDir, 2*time.Minute); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	k.logger.Info("Microservice installation completed successfully")
	return nil
}

// UninstallMicroservice uninstalls microservice manifests
func (k *KustomizeManager) UninstallMicroservice(ctx context.Context) error {
	k.logger.Info("Starting microservice uninstallation")

	// Create temporary directory for manifests
	tempDir, err := os.MkdirTemp("", "navigator-microservice-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			k.logger.Warn("Failed to cleanup temp directory", "error", cleanupErr, "path", tempDir)
		}
	}()

	// Extract manifests to temporary directory
	if err := k.extractManifests(tempDir); err != nil {
		return fmt.Errorf("failed to extract manifests: %w", err)
	}

	// Delete manifests using kubectl
	if err := k.deleteManifests(ctx, tempDir); err != nil {
		return fmt.Errorf("failed to delete manifests: %w", err)
	}

	k.logger.Info("Microservice uninstallation completed")
	return nil
}

// IsMicroserviceInstalled checks if microservice is installed
func (k *KustomizeManager) IsMicroserviceInstalled(ctx context.Context) (bool, string, error) {
	// Check if namespace exists and has our labeled resources
	args := []string{"get", "namespace", "microservice-demo"}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If namespace doesn't exist, microservice is not installed
		if strings.Contains(string(output), "not found") {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check namespace: %w", err)
	}

	// Check if deployments exist
	args = []string{"get", "deployments", "-n", "microservice-demo", "-l", "app.kubernetes.io/part-of=three-tier-microservice"}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	cmd = exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err = cmd.CombinedOutput()
	if err != nil {
		return false, "", fmt.Errorf("failed to check deployments: %w", err)
	}

	// If we have deployments, consider it installed
	hasDeployments := !strings.Contains(string(output), "No resources found")
	return hasDeployments, "latest", nil
}

// extractManifests extracts embedded manifests to a temporary directory
func (k *KustomizeManager) extractManifests(tempDir string) error {
	entries, err := manifestsFS.ReadDir("manifests")
	if err != nil {
		return fmt.Errorf("failed to read embedded manifests: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := manifestsFS.ReadFile(filepath.Join("manifests", entry.Name()))
		if err != nil {
			return fmt.Errorf("failed to read manifest %s: %w", entry.Name(), err)
		}

		filePath := filepath.Join(tempDir, entry.Name())
		if err := os.WriteFile(filePath, content, 0600); err != nil {
			return fmt.Errorf("failed to write manifest %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// applyManifests applies the Kustomize manifests using kubectl
func (k *KustomizeManager) applyManifests(ctx context.Context, manifestDir string, timeout time.Duration) error {
	args := []string{"apply", "-k", manifestDir}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	k.logger.Info("Applying Kustomize manifests", "directory", manifestDir)
	cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.logger.Error("Failed to apply manifests", "error", err, "output", string(output))
		return fmt.Errorf("kubectl apply failed: %w", err)
	}

	k.logger.Debug("kubectl apply output", "output", string(output))

	// Wait for deployments to be ready if timeout is specified
	if timeout > 0 {
		if err := k.waitForDeployments(ctx, timeout); err != nil {
			return fmt.Errorf("failed waiting for deployments: %w", err)
		}
	}

	return nil
}

// deleteManifests deletes the Kustomize manifests using kubectl
func (k *KustomizeManager) deleteManifests(ctx context.Context, manifestDir string) error {
	args := []string{"delete", "-k", manifestDir, "--ignore-not-found=true"}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	k.logger.Info("Deleting Kustomize manifests", "directory", manifestDir)
	cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.logger.Error("Failed to delete manifests", "error", err, "output", string(output))
		return fmt.Errorf("kubectl delete failed: %w", err)
	}

	k.logger.Debug("kubectl delete output", "output", string(output))
	return nil
}

// waitForDeployments waits for all deployments to be ready
func (k *KustomizeManager) waitForDeployments(ctx context.Context, timeout time.Duration) error {
	deployments := []string{"frontend", "backend", "database"}

	for _, deployment := range deployments {
		args := []string{"rollout", "status", "deployment/" + deployment, "-n", "microservice-demo", "--timeout=" + timeout.String()}
		if k.kubeconfig != "" {
			args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
		}

		k.logger.Info("Waiting for deployment to be ready", "deployment", deployment, "timeout", timeout)
		cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
		output, err := cmd.CombinedOutput()
		if err != nil {
			k.logger.Error("Deployment not ready", "deployment", deployment, "error", err, "output", string(output))
			return fmt.Errorf("deployment %s not ready: %w", deployment, err)
		}
		k.logger.Info("Deployment is ready", "deployment", deployment)
	}

	return nil
}

// checkKubectl verifies that kubectl is available
func (k *KustomizeManager) checkKubectl() error {
	cmd := exec.Command("kubectl", "version", "--client", "--output=yaml")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubectl not found or not working: %w", err)
	}
	return nil
}
