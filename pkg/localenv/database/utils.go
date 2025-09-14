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

package database

import (
	"context"
	"fmt"
	"os/exec"
)

// checkKubectl verifies that kubectl is available and working
func (k *KustomizeManager) checkKubectl() error {
	cmd := exec.Command("kubectl", "version", "--client", "--output=json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.logger.Error("kubectl not available", "error", err, "output", string(output))
		return fmt.Errorf("kubectl is required but not available: %w", err)
	}
	k.logger.Debug("kubectl is available", "output", string(output))
	return nil
}

// WaitForDatabaseReady waits for database deployment to be ready
func (k *KustomizeManager) WaitForDatabaseReady(ctx context.Context) error {
	k.logger.Info("Waiting for database to be ready")

	deployment := "database"
	args := []string{"rollout", "status", "deployment/" + deployment, "-n", "database", "--timeout=5m"}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	k.logger.Info("Waiting for deployment to be ready", "deployment", deployment, "namespace", "database")
	cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.logger.Error("Deployment not ready", "deployment", deployment, "namespace", "database", "error", err, "output", string(output))
		return fmt.Errorf("deployment %s not ready: %w", deployment, err)
	}
	k.logger.Info("Database deployment is ready", "deployment", deployment, "namespace", "database")
	return nil
}