package local

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed kind-config.yaml
var kindConfig []byte

// KindCluster manages a Kind cluster for integration testing
type KindCluster struct {
	Name       string
	ConfigPath string
	kubeconfig string
	configFile *os.File
}

// NewKindCluster creates a new Kind cluster manager
func NewKindCluster(name string) *KindCluster {
	return &KindCluster{
		Name: name,
	}
}

// Create creates a new Kind cluster with the given configuration
func (k *KindCluster) Create(ctx context.Context) error {
	// Check if kind is available
	if _, err := exec.LookPath("kind"); err != nil {
		return fmt.Errorf("kind not found in PATH: %w", err)
	}

	// Check if cluster already exists
	if k.Exists(ctx) {
		return fmt.Errorf("cluster %s already exists", k.Name)
	}

	// Create temporary config file from embedded content
	tmpFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	k.configFile = tmpFile
	k.ConfigPath = tmpFile.Name()

	if _, err := tmpFile.Write(kindConfig); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write config to temp file: %w", err)
	}
	tmpFile.Close()

	args := []string{"create", "cluster", "--name", k.Name, "--config", k.ConfigPath}

	cmd := exec.CommandContext(ctx, "kind", args...)
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
		return nil // Already deleted
	}

	cmd := exec.CommandContext(ctx, "kind", "delete", "cluster", "--name", k.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w", err)
	}

	return nil
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

// GetKubeconfig returns the kubeconfig path for the Kind cluster
func (k *KindCluster) GetKubeconfig(ctx context.Context) (string, error) {
	if k.kubeconfig != "" {
		return k.kubeconfig, nil
	}

	// Create temporary kubeconfig file
	tmpDir, err := os.MkdirTemp("", "kind-kubeconfig-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")

	cmd := exec.CommandContext(ctx, "kind", "get", "kubeconfig", "--name", k.Name)
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

// waitForReady waits for the Kind cluster to be ready
func (k *KindCluster) waitForReady(ctx context.Context) error {
	timeout := 2 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		kubeconfig, err := k.GetKubeconfig(ctx)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// Try to connect to the cluster
		cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "cluster-info")
		if err := cmd.Run(); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// Check if nodes are ready
		cmd = exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "wait", "--for=condition=Ready", "nodes", "--all", "--timeout=60s")
		if err := cmd.Run(); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("cluster not ready within %v", timeout)
}

// InstallIstio installs Istio on the Kind cluster
func (k *KindCluster) InstallIstio(ctx context.Context) error {
	kubeconfig, err := k.GetKubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Check if istioctl is available
	if _, err := exec.LookPath("istioctl"); err != nil {
		return fmt.Errorf("istioctl not found in PATH: %w", err)
	}

	// Install Istio with demo profile for testing
	cmd := exec.CommandContext(ctx, "istioctl", "install", "--set", "values.defaultRevision=default", "--kubeconfig", kubeconfig, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Istio: %w", err)
	}

	// Wait for Istio to be ready
	if err := k.waitForIstioReady(ctx, kubeconfig); err != nil {
		return fmt.Errorf("Istio not ready: %w", err)
	}

	return nil
}

// EnableIstioInjection enables automatic Istio sidecar injection for a namespace
func (k *KindCluster) EnableIstioInjection(ctx context.Context, namespace string) error {
	kubeconfig, err := k.GetKubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Label namespace for Istio injection
	cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "label", "namespace", namespace, "istio-injection=enabled", "--overwrite")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable Istio injection for namespace %s: %w", namespace, err)
	}

	return nil
}

// waitForIstioReady waits for Istio components to be ready
func (k *KindCluster) waitForIstioReady(ctx context.Context, kubeconfig string) error {
	timeout := 5 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Check if istiod deployment is ready
		cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "wait", "--for=condition=available", "deployment/istiod", "-n", "istio-system", "--timeout=30s")
		if err := cmd.Run(); err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		// Check if istio-proxy container is being injected by creating a test pod
		testPodYaml := `
apiVersion: v1
kind: Pod
metadata:
  name: istio-test-pod
  namespace: default
spec:
  containers:
  - name: test
    image: nginx
    ports:
    - containerPort: 80
`
		// Create test pod to verify injection
		cmd = exec.Command("kubectl", "--kubeconfig", kubeconfig, "apply", "-f", "-")
		cmd.Stdin = strings.NewReader(testPodYaml)
		if err := cmd.Run(); err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		// Wait a moment for injection to happen
		time.Sleep(5 * time.Second)

		// Check if the pod has istio-proxy container
		cmd = exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "pod", "istio-test-pod", "-o", "jsonpath={.spec.containers[*].name}")
		output, err := cmd.Output()
		if err != nil {
			// Clean up test pod and continue
			exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "delete", "pod", "istio-test-pod", "--ignore-not-found")
			time.Sleep(10 * time.Second)
			continue
		}

		// Clean up test pod
		exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "delete", "pod", "istio-test-pod", "--ignore-not-found")

		containers := strings.Fields(string(output))
		for _, container := range containers {
			if container == "istio-proxy" {
				return nil // Istio injection is working
			}
		}

		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("Istio not ready within %v", timeout)
}

// Cleanup removes temporary files created by the cluster
func (k *KindCluster) Cleanup() error {
	var errs []error

	// Clean up kubeconfig
	if k.kubeconfig != "" {
		dir := filepath.Dir(k.kubeconfig)
		if strings.Contains(dir, "kind-kubeconfig-") {
			if err := os.RemoveAll(dir); err != nil {
				errs = append(errs, fmt.Errorf("failed to remove kubeconfig dir: %w", err))
			}
		}
	}

	// Clean up config file
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
