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

package localenv

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

//go:embed kind-config.yaml
var kindConfig []byte

// KindEnvironment implements LocalEnvironment using Kind (Kubernetes in Docker)
type KindEnvironment struct {
	config   *Config
	cluster  *KindCluster
	fixtures *Fixtures
}

// KindCluster manages a Kind cluster
type KindCluster struct {
	Name       string
	ConfigPath string
	kubeconfig string
	configFile *os.File
}

// NewKindEnvironment creates a new Kind-based local environment
func NewKindEnvironment() *KindEnvironment {
	return &KindEnvironment{}
}

// Setup initializes the Kind environment
func (e *KindEnvironment) Setup(ctx context.Context, config *Config) error {
	e.config = config

	// Create Kind cluster
	e.cluster = NewKindCluster(config.ClusterName)
	if err := e.cluster.Create(ctx); err != nil {
		return fmt.Errorf("failed to create Kind cluster: %w", err)
	}

	// Get kubeconfig and create Kubernetes client
	kubeconfig, err := e.cluster.GetKubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	kubeClient, err := e.createKubernetesClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Initialize fixtures manager
	e.fixtures = NewFixtures(kubeClient, config.Namespace)

	// Create namespace
	if err := e.fixtures.CreateNamespace(ctx); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Install Istio if enabled
	if config.IstioEnabled {
		if err := e.cluster.InstallIstio(ctx); err != nil {
			return fmt.Errorf("failed to install Istio: %w", err)
		}

		if err := e.cluster.EnableIstioInjection(ctx, config.Namespace); err != nil {
			return fmt.Errorf("failed to enable Istio injection: %w", err)
		}
	}

	// Note: Navigator server should be started separately
	// For now, we'll prepare for it but not auto-start it
	fmt.Printf("Kind cluster ready! Start Navigator with:\n")
	fmt.Printf("  ./navigator serve --kubeconfig %s --port %d\n", kubeconfig, config.Port)
	fmt.Println()

	return nil
}

// Teardown cleans up the Kind environment
func (e *KindEnvironment) Teardown(ctx context.Context) error {
	var errs []error

	// Delete Kind cluster
	if e.cluster != nil {
		if err := e.cluster.Delete(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete Kind cluster: %w", err))
		}

		if err := e.cluster.Cleanup(); err != nil {
			errs = append(errs, fmt.Errorf("failed to cleanup Kind cluster: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("teardown errors: %v", errs)
	}

	return nil
}

// DeployScenario deploys a predefined scenario to the environment
func (e *KindEnvironment) DeployScenario(ctx context.Context, scenario *Scenario) error {
	if e.fixtures == nil {
		return fmt.Errorf("environment not set up")
	}

	// Deploy services based on scenario
	for _, serviceSpec := range scenario.Services {
		switch serviceSpec.Type {
		case ServiceTypeWeb:
			if err := e.fixtures.CreateWebService(ctx, serviceSpec.Name, serviceSpec.Replicas); err != nil {
				return fmt.Errorf("failed to create web service %s: %w", serviceSpec.Name, err)
			}

		case ServiceTypeHeadless:
			if err := e.fixtures.CreateHeadlessService(ctx, serviceSpec.Name); err != nil {
				return fmt.Errorf("failed to create headless service %s: %w", serviceSpec.Name, err)
			}

		case ServiceTypeExternal:
			if err := e.fixtures.CreateExternalService(ctx, serviceSpec.Name, serviceSpec.ExternalIPs); err != nil {
				return fmt.Errorf("failed to create external service %s: %w", serviceSpec.Name, err)
			}

		case ServiceTypeTopology:
			if err := e.fixtures.CreateTopologyService(ctx, serviceSpec.Name, serviceSpec.Replicas, serviceSpec.NextService); err != nil {
				return fmt.Errorf("failed to create topology service %s: %w", serviceSpec.Name, err)
			}
		}
	}

	// Wait for deployments to be ready
	var serviceNames []string
	for _, spec := range scenario.Services {
		if spec.Type == ServiceTypeWeb || spec.Type == ServiceTypeTopology {
			serviceNames = append(serviceNames, spec.Name)
		}
	}

	if len(serviceNames) > 0 {
		for _, serviceName := range serviceNames {
			if err := e.fixtures.WaitForDeployment(ctx, serviceName); err != nil {
				return fmt.Errorf("failed waiting for deployment %s: %w", serviceName, err)
			}
		}
	}

	return nil
}

// GetKubeconfig returns the kubeconfig file path
func (e *KindEnvironment) GetKubeconfig() string {
	if e.cluster == nil {
		return ""
	}
	return e.cluster.kubeconfig
}

// GetGRPCClient returns the gRPC client (currently not implemented)
func (e *KindEnvironment) GetGRPCClient() v1alpha1.ServiceRegistryServiceClient {
	return nil // TODO: Implement when auto-starting Navigator server
}

// GetNamespace returns the primary namespace
func (e *KindEnvironment) GetNamespace() string {
	if e.config == nil {
		return ""
	}
	return e.config.Namespace
}

// IsReady checks if the environment is ready
func (e *KindEnvironment) IsReady(ctx context.Context) bool {
	if e.cluster == nil {
		return false
	}

	// For now, just check if the cluster exists
	// In the future, we could check if Navigator server is responding
	return e.cluster.Exists(ctx)
}

// SetConfig sets the configuration for the environment (used for teardown)
func (e *KindEnvironment) SetConfig(config *Config) {
	e.config = config
	if e.cluster == nil {
		e.cluster = NewKindCluster(config.ClusterName)
	}
}

// NewKindCluster creates a new Kind cluster manager
func NewKindCluster(name string) *KindCluster {
	return &KindCluster{
		Name: name,
	}
}

// Create creates a new Kind cluster
func (k *KindCluster) Create(ctx context.Context) error {
	// Check if kind is available
	if _, err := exec.LookPath("kind"); err != nil {
		return fmt.Errorf("kind not found in PATH: %w", err)
	}

	// Check if cluster already exists
	if k.Exists(ctx) {
		return fmt.Errorf("cluster %s already exists", k.Name)
	}

	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	k.configFile = tmpFile
	k.ConfigPath = tmpFile.Name()

	if _, err := tmpFile.Write(kindConfig); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write config to temp file: %w", err)
	}
	_ = tmpFile.Close()

	args := []string{"create", "cluster", "--name", k.Name, "--config", k.ConfigPath}

	cmd := exec.CommandContext(ctx, "kind", args...) // #nosec G204 - args are controlled by this package
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create kind cluster: %w", err)
	}

	// Wait for cluster to be ready
	if err := k.waitForReady(ctx); err != nil {
		return fmt.Errorf("cluster not ready: %w", err)
	}

	return nil
}

// Delete removes the Kind cluster
func (k *KindCluster) Delete(ctx context.Context) error {
	if !k.Exists(ctx) {
		return nil
	}

	cmd := exec.CommandContext(ctx, "kind", "delete", "cluster", "--name", k.Name) // #nosec G204 - k.Name is controlled by this package
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Exists checks if the Kind cluster exists
func (k *KindCluster) Exists(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if cluster == k.Name {
			return true
		}
	}

	return false
}

// GetKubeconfig returns the kubeconfig path
func (k *KindCluster) GetKubeconfig(ctx context.Context) (string, error) {
	if k.kubeconfig != "" {
		return k.kubeconfig, nil
	}

	tmpDir, err := os.MkdirTemp("", "kind-kubeconfig-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")

	cmd := exec.CommandContext(ctx, "kind", "get", "kubeconfig", "--name", k.Name) // #nosec G204 - k.Name is controlled by this package
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	if err := os.WriteFile(kubeconfigPath, output, 0600); err != nil {
		return "", fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	k.kubeconfig = kubeconfigPath
	return kubeconfigPath, nil
}

// InstallIstio installs Istio on the Kind cluster
func (k *KindCluster) InstallIstio(ctx context.Context) error {
	kubeconfig, err := k.GetKubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	if _, err := exec.LookPath("istioctl"); err != nil {
		return fmt.Errorf("istioctl not found in PATH: %w", err)
	}

	cmd := exec.CommandContext(ctx, "istioctl", "install", "--set", "values.defaultRevision=default", "--kubeconfig", kubeconfig, "-y") // #nosec G204 - kubeconfig is controlled by this package
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// EnableIstioInjection enables Istio injection for a namespace
func (k *KindCluster) EnableIstioInjection(ctx context.Context, namespace string) error {
	kubeconfig, err := k.GetKubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "label", "namespace", namespace, "istio-injection=enabled", "--overwrite") // #nosec G204 - kubeconfig and namespace are controlled by this package
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Cleanup removes temporary files
func (k *KindCluster) Cleanup() error {
	var errs []error

	if k.kubeconfig != "" {
		dir := filepath.Dir(k.kubeconfig)
		if strings.Contains(dir, "kind-kubeconfig-") {
			if err := os.RemoveAll(dir); err != nil {
				errs = append(errs, fmt.Errorf("failed to remove kubeconfig dir: %w", err))
			}
		}
	}

	if k.ConfigPath != "" {
		if err := os.Remove(k.ConfigPath); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("failed to remove config file: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// waitForReady waits for the cluster to be ready
func (k *KindCluster) waitForReady(ctx context.Context) error {
	timeout := 2 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		kubeconfig, err := k.GetKubeconfig(ctx)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "wait", "--for=condition=Ready", "nodes", "--all", "--timeout=60s") // #nosec G204 - kubeconfig is controlled by this package
		if err := cmd.Run(); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("cluster not ready within %v", timeout)
}

// createKubernetesClient creates a Kubernetes client from kubeconfig
func (e *KindEnvironment) createKubernetesClient(kubeconfigPath string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return client, nil
}
