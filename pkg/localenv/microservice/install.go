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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// InstallMicroservice installs microservice manifests
func (k *KustomizeManager) InstallMicroservice(ctx context.Context) error {
	k.logger.Info("Starting microservice installation", "namespace", "microservices")

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

	// Apply manifests using kubectl with 2 minute timeout (without waiting for deployments)
	if err := k.applyManifests(ctx, tempDir, 0); err != nil {
		return fmt.Errorf("failed to apply manifests: %w", err)
	}

	// Install monolith namespace separately
	monolithDir := filepath.Join(tempDir, "monolith")
	if err := k.applyMonolithManifests(ctx, monolithDir, 0); err != nil {
		return fmt.Errorf("failed to apply monolith manifests: %w", err)
	}

	// Wait for all deployments after everything is applied
	if err := k.waitForDeployments(ctx, 2*time.Minute); err != nil {
		return fmt.Errorf("failed waiting for deployments: %w", err)
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

	// Delete monolith manifests
	monolithDir := filepath.Join(tempDir, "monolith")
	if err := k.deleteMonolithManifests(ctx, monolithDir); err != nil {
		k.logger.Warn("Failed to delete monolith manifests", "error", err)
	}

	// Delete main manifests using kubectl
	if err := k.deleteManifests(ctx, tempDir); err != nil {
		return fmt.Errorf("failed to delete manifests: %w", err)
	}

	k.logger.Info("Microservice uninstallation completed")
	return nil
}

// IsMicroserviceInstalled checks if microservice is installed
func (k *KustomizeManager) IsMicroserviceInstalled(ctx context.Context) (bool, string, error) {
	// Check if microservices namespace exists
	args := []string{"get", "namespace", "microservices"}
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
		return false, "", fmt.Errorf("failed to check microservices namespace: %w", err)
	}

	// Check if microservices deployments exist
	args = []string{"get", "deployments", "-n", "microservices", "-l", "app.kubernetes.io/part-of=three-tier-microservice"}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	cmd = exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err = cmd.CombinedOutput()
	if err != nil {
		return false, "", fmt.Errorf("failed to check microservices deployments: %w", err)
	}

	hasMicroserviceDeployments := !strings.Contains(string(output), "No resources found")
	return hasMicroserviceDeployments, "latest", nil
}

// extractManifests extracts embedded manifests to a temporary directory recursively
func (k *KustomizeManager) extractManifests(tempDir string) error {
	return k.extractDirectory("manifests", tempDir)
}

// extractDirectory recursively extracts a directory from the embedded filesystem
func (k *KustomizeManager) extractDirectory(srcDir, destDir string) error {
	entries, err := manifestsFS.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read embedded directory %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			// Create the directory
			if err := os.MkdirAll(destPath, 0750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}

			// Recursively extract subdirectory
			if err := k.extractDirectory(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to extract subdirectory %s: %w", srcPath, err)
			}
		} else {
			// Extract file
			content, err := manifestsFS.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read manifest %s: %w", srcPath, err)
			}

			if err := os.WriteFile(destPath, content, 0600); err != nil {
				return fmt.Errorf("failed to write manifest %s: %w", destPath, err)
			}
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

// applyMonolithManifests applies the monolith Kustomize manifests separately
func (k *KustomizeManager) applyMonolithManifests(ctx context.Context, manifestDir string, timeout time.Duration) error {
	args := []string{"apply", "-k", manifestDir}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	k.logger.Info("Applying monolith Kustomize manifests", "directory", manifestDir)
	cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.logger.Error("Failed to apply monolith manifests", "error", err, "output", string(output))
		return fmt.Errorf("kubectl apply monolith failed: %w", err)
	}

	k.logger.Debug("kubectl apply monolith output", "output", string(output))
	return nil
}

// deleteMonolithManifests deletes the monolith Kustomize manifests
func (k *KustomizeManager) deleteMonolithManifests(ctx context.Context, manifestDir string) error {
	args := []string{"delete", "-k", manifestDir, "--ignore-not-found=true"}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	k.logger.Info("Deleting monolith Kustomize manifests", "directory", manifestDir)
	cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.logger.Error("Failed to delete monolith manifests", "error", err, "output", string(output))
		return fmt.Errorf("kubectl delete monolith failed: %w", err)
	}

	k.logger.Debug("kubectl delete monolith output", "output", string(output))
	return nil
}

// waitForDeployments waits for all deployments to be ready
func (k *KustomizeManager) waitForDeployments(ctx context.Context, timeout time.Duration) error {
	deployments := []string{"frontend", "backend"}
	monolithDeployments := []string{"monolith"}

	// Wait for microservices deployments
	for _, deployment := range deployments {
		args := []string{"rollout", "status", "deployment/" + deployment, "-n", "microservices", "--timeout=" + timeout.String()}
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

	// Wait for monolith deployments
	for _, deployment := range monolithDeployments {
		args := []string{"rollout", "status", "deployment/" + deployment, "-n", "monolith", "--timeout=" + timeout.String()}
		if k.kubeconfig != "" {
			args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
		}

		k.logger.Info("Waiting for monolith deployment to be ready", "deployment", deployment, "timeout", timeout)
		cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
		output, err := cmd.CombinedOutput()
		if err != nil {
			k.logger.Error("Monolith deployment not ready", "deployment", deployment, "error", err, "output", string(output))
			return fmt.Errorf("monolith deployment %s not ready: %w", deployment, err)
		}
		k.logger.Info("Monolith deployment is ready", "deployment", deployment)
	}

	return nil
}
