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
	"net/http"
	"os/exec"
	"time"

	"github.com/liamawhite/navigator/pkg/localenv/kind"
)

// VerifyMicroserviceChain tests the full request chain through the microservices using default port
func (k *KustomizeManager) VerifyMicroserviceChain(ctx context.Context) error {
	return k.VerifyMicroserviceChainWithPort(ctx, kind.HTTPNodePort)
}

// VerifyMicroserviceChainWithPort tests the full request chain through the microservices using a custom port
func (k *KustomizeManager) VerifyMicroserviceChainWithPort(ctx context.Context, httpPort int) error {
	k.logger.Info("Verifying microservice chain connectivity", "http_port", httpPort)

	// Wait for all deployments to be ready with extended timeout for verification
	if err := k.WaitForMicroservicesReady(ctx, 5*time.Minute); err != nil {
		return fmt.Errorf("microservices not ready: %w", err)
	}

	// Test the actual HTTP request chain: Gateway -> Frontend -> Backend -> Database
	if err := k.testRequestChainWithPort(ctx, httpPort); err != nil {
		return fmt.Errorf("request chain test failed: %w", err)
	}

	k.logger.Info("Microservice chain verification completed successfully")
	return nil
}

// WaitForMicroservicesReady waits for all microservice deployments to be ready
func (k *KustomizeManager) WaitForMicroservicesReady(ctx context.Context, timeout time.Duration) error {
	k.logger.Info("Waiting for microservices to be ready", "timeout", timeout)

	// Microservices namespace deployments
	microservicesDeployments := []string{"frontend", "backend"}
	for _, deployment := range microservicesDeployments {
		if err := k.waitForDeploymentReady(ctx, deployment, "microservices", timeout); err != nil {
			return fmt.Errorf("deployment %s not ready: %w", deployment, err)
		}
	}

	// Database namespace deployments
	databaseDeployments := []string{"database"}
	for _, deployment := range databaseDeployments {
		if err := k.waitForDeploymentReady(ctx, deployment, "database", timeout); err != nil {
			return fmt.Errorf("deployment %s not ready: %w", deployment, err)
		}
	}

	k.logger.Info("All microservices are ready")
	return nil
}

// waitForDeploymentReady waits for a specific deployment to be ready in the specified namespace
func (k *KustomizeManager) waitForDeploymentReady(ctx context.Context, deployment string, namespace string, timeout time.Duration) error {
	args := []string{"rollout", "status", "deployment/" + deployment, "-n", namespace, "--timeout=" + timeout.String()}
	if k.kubeconfig != "" {
		args = append([]string{"--kubeconfig", k.kubeconfig}, args...)
	}

	k.logger.Info("Waiting for deployment to be ready", "deployment", deployment, "namespace", namespace, "timeout", timeout)
	cmd := exec.CommandContext(ctx, "kubectl", args...) //nolint:gosec // kubectl execution with controlled args
	output, err := cmd.CombinedOutput()
	if err != nil {
		k.logger.Error("Deployment not ready", "deployment", deployment, "namespace", namespace, "error", err, "output", string(output))
		return fmt.Errorf("deployment %s not ready: %w", deployment, err)
	}
	k.logger.Info("Deployment is ready", "deployment", deployment, "namespace", namespace)
	return nil
}

// testRequestChainWithPort performs an actual HTTP request to test the full microservice chain using a custom port
func (k *KustomizeManager) testRequestChainWithPort(ctx context.Context, httpPort int) error {
	k.logger.Info("Testing HTTP request: Gateway -> Frontend via Istio service mesh", "http_port", httpPort)

	// Get the gateway URL for testing
	gatewayURL, err := k.getGatewayURLWithPort(ctx, httpPort)
	if err != nil {
		k.logger.Error("Failed to determine gateway URL", "error", err)
		return fmt.Errorf("failed to get gateway URL: %w", err)
	}

	// Build the request URL for the microservice chain test
	// This will verify the full chain: Gateway -> Frontend -> Backend -> Database (cross-namespace)
	requestURL := fmt.Sprintf("%s/microservices/proxy/backend:8080/proxy/database.database:8080", gatewayURL)
	k.logger.Info("Making HTTP request to test full microservice chain", "url", requestURL)

	// Create HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request through the full chain
	k.logger.Debug("Sending GET request...", "url", requestURL, "timeout", "30s")
	resp, err := client.Get(requestURL)
	if err != nil {
		k.logger.Error("HTTP request failed",
			"url", requestURL,
			"error", err,
			"hint", "Check if Istio gateway is accessible and services are running")
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			k.logger.Warn("Failed to close response body", "error", err)
		}
	}()

	// Log response details for debugging
	k.logger.Debug("HTTP response received",
		"url", requestURL,
		"status_code", resp.StatusCode,
		"status", resp.Status,
		"headers", fmt.Sprintf("%v", resp.Header))

	// Check if we got a successful response
	if resp.StatusCode != http.StatusOK {
		k.logger.Error("HTTP request returned non-200 status",
			"url", requestURL,
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"hint", "Check microservice logs and Istio configuration")
		return fmt.Errorf("HTTP request returned status %d (%s), expected 200", resp.StatusCode, resp.Status)
	}

	k.logger.Info("✓ HTTP request test successful!",
		"url", requestURL,
		"status_code", resp.StatusCode,
		"chain", "Gateway -> Frontend -> Backend -> Database (via Istio service mesh)")

	return nil
}

// getGatewayURLWithPort determines the gateway URL for the Kind cluster using a custom port
func (k *KustomizeManager) getGatewayURLWithPort(ctx context.Context, httpPort int) (string, error) {
	k.logger.Info("Getting gateway URL for Kind cluster with NodePort access", "http_port", httpPort)

	// Use the specified NodePort configured in Kind cluster configuration
	// This avoids dynamic port detection and ensures consistent access
	gatewayURL := fmt.Sprintf("http://localhost:%d", httpPort)
	k.logger.Info("Gateway URL determined via fixed NodePort", "url", gatewayURL, "node_port", httpPort)

	return gatewayURL, nil
}
