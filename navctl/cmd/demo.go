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

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/liamawhite/navigator/pkg/localenv/istio"
	"github.com/liamawhite/navigator/pkg/localenv/kind"
	"github.com/liamawhite/navigator/pkg/localenv/microservice"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	demoClusterName  string
	demoCleanup      bool
	demoIstioVersion string
)

// demoCmd represents the demo command
var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Manage demo Kind clusters for testing Navigator",
	Long: `Manage demo Kind clusters for testing Navigator functionality.

Use subcommands to start or stop demo clusters that can be used to test
Navigator's service discovery and proxy analysis features.`,
}

// demoStartCmd represents the demo start command
var demoStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a demo Kind cluster with Istio service mesh and microservices",
	Long: `Start a demo Kind cluster for testing Navigator functionality.

This command creates a basic Kind cluster, installs Istio service mesh, and 
deploys a microservice topology for testing Navigator's service discovery 
and proxy analysis features.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.For("demo")
		ctx := context.Background()

		logger.Info("Starting demo cluster creation", "cluster", demoClusterName)

		kindMgr := kind.NewKindManager(logger)

		// Check if cluster already exists
		exists, err := kindMgr.ClusterExists(ctx, demoClusterName)
		if err != nil {
			return fmt.Errorf("failed to check if cluster exists: %w", err)
		}

		if exists {
			logger.Info("Cluster already exists", "cluster", demoClusterName)
			if !demoCleanup {
				return fmt.Errorf("cluster %s already exists, use --cleanup to recreate", demoClusterName)
			}

			logger.Info("Deleting existing cluster", "cluster", demoClusterName)
			if err := kindMgr.DeleteCluster(ctx, demoClusterName); err != nil {
				return fmt.Errorf("failed to delete existing cluster: %w", err)
			}
		}

		// Create the cluster
		config := kind.DefaultKindConfig(demoClusterName)
		if err := kindMgr.CreateCluster(ctx, config); err != nil {
			return fmt.Errorf("failed to create demo cluster: %w", err)
		}

		// Wait for cluster to be ready
		logger.Info("Waiting for cluster to be ready...")
		if err := kindMgr.WaitForClusterReady(ctx, demoClusterName); err != nil {
			return fmt.Errorf("cluster failed to become ready: %w", err)
		}

		// Export kubeconfig
		kubeconfigPath := fmt.Sprintf("%s-kubeconfig", demoClusterName)
		if err := kindMgr.ExportKubeconfig(ctx, demoClusterName, kubeconfigPath); err != nil {
			logger.Warn("Failed to export kubeconfig", "error", err)
		} else {
			logger.Info("Kubeconfig exported", "path", kubeconfigPath)
		}

		logger.Info("Demo cluster created successfully!",
			"cluster", demoClusterName,
			"kubeconfig", kubeconfigPath)

		// Install Istio
		logger.Info("Installing Istio service mesh", "version", demoIstioVersion)

		// Get absolute path for kubeconfig
		absKubeconfigPath, err := filepath.Abs(kubeconfigPath)
		if err != nil {
			logger.Warn("Failed to get absolute kubeconfig path", "error", err)
			absKubeconfigPath = kubeconfigPath
		}

		// Create Helm manager
		helmMgr, err := istio.NewHelmManager(absKubeconfigPath, "istio-system", logger)
		if err != nil {
			return fmt.Errorf("failed to create Helm manager for Istio installation: %w", err)
		}

		// Install Istio
		istioConfig := istio.DefaultIstioConfig(demoIstioVersion)
		if err := helmMgr.InstallIstio(ctx, istioConfig); err != nil {
			return fmt.Errorf("failed to install Istio: %w", err)
		}

		logger.Info("Istio installed successfully")

		// Create Kubernetes client for namespace labeling
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", absKubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed to build kubeconfig for namespace labeling: %w", err)
		}

		clientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client for namespace labeling: %w", err)
		}

		// Label default namespace for Istio injection
		logger.Info("Labeling default namespace for Istio injection")
		if err := labelNamespaceForIstio(clientset, "default", logger); err != nil {
			return fmt.Errorf("failed to label default namespace for Istio injection: %w", err)
		}

		// Install microservices topology
		logger.Info("Installing microservices", "scenario", "three-tier")

		// Create microservice Helm manager (using default namespace for Helm release management)
		microHelmMgr, err := microservice.NewHelmManager(absKubeconfigPath, "default", logger)
		if err != nil {
			return fmt.Errorf("failed to create Helm manager for microservice installation: %w", err)
		}

		// Install microservices with custom values
		microConfig := microservice.DefaultMicroserviceConfig()
		microConfig.Namespace = "default"
		microConfig.ReleaseName = "microservice-demo"
		microConfig.CustomValues = microservice.CreateThreeTierApplicationValues()

		if err := microHelmMgr.InstallMicroservice(ctx, microConfig); err != nil {
			return fmt.Errorf("failed to install microservices: %w", err)
		}

		logger.Info("Microservices installed successfully")
		logger.Info("Demo cluster ready",
			"cluster", demoClusterName,
			"kubeconfig", kubeconfigPath,
			"istio_version", demoIstioVersion,
			"microservices_namespace", "default")

		return nil
	},
}

// demoStopCmd represents the demo stop command
var demoStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a demo Kind cluster",
	Long: `Stop (delete) a demo Kind cluster.

This command deletes the specified Kind cluster and cleans up associated resources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.For("demo")
		ctx := context.Background()

		logger.Info("Stopping demo cluster", "cluster", demoClusterName)

		kindMgr := kind.NewKindManager(logger)

		// Check if cluster exists
		exists, err := kindMgr.ClusterExists(ctx, demoClusterName)
		if err != nil {
			return fmt.Errorf("failed to check if cluster exists: %w", err)
		}

		if !exists {
			logger.Info("Cluster does not exist", "cluster", demoClusterName)
			return nil
		}

		// Try to clean up Istio if it exists (best effort)
		kubeconfigPath := fmt.Sprintf("%s-kubeconfig", demoClusterName)
		absKubeconfigPath, err := filepath.Abs(kubeconfigPath)
		if err != nil {
			logger.Debug("Could not get absolute kubeconfig path", "error", err)
			absKubeconfigPath = kubeconfigPath
		}

		// Check if kubeconfig exists and try cleanup
		if _, err := filepath.Abs(kubeconfigPath); err == nil {
			// Try to clean up microservices first (best effort)
			logger.Info("Attempting microservice cleanup")

			microHelmMgr, err := microservice.NewHelmManager(absKubeconfigPath, "default", logger)
			if err != nil {
				logger.Debug("Could not create microservice Helm manager for cleanup", "error", err)
			} else {
				// Check if microservice is installed
				if installed, version, err := microHelmMgr.IsMicroserviceInstalled(ctx, "microservice-demo"); err == nil && installed {
					logger.Info("Found microservice installation, cleaning up", "version", version)
					if err := microHelmMgr.UninstallMicroservice(ctx, "microservice-demo"); err != nil {
						logger.Warn("Failed to uninstall microservices", "error", err)
					} else {
						logger.Info("Microservices uninstalled successfully")
					}
				} else {
					logger.Debug("No microservice installation found or could not check")
				}
			}

			// Then clean up Istio
			logger.Info("Attempting Istio cleanup")

			helmMgr, err := istio.NewHelmManager(absKubeconfigPath, "istio-system", logger)
			if err != nil {
				logger.Debug("Could not create Helm manager for cleanup", "error", err)
			} else {
				// Check if Istio is installed and get version
				if installed, version, err := helmMgr.IsIstioInstalled(ctx); err == nil && installed {
					logger.Info("Found Istio installation, cleaning up", "version", version)
					if err := helmMgr.UninstallIstio(ctx, version); err != nil {
						logger.Warn("Failed to uninstall Istio", "error", err)
					} else {
						logger.Info("Istio uninstalled successfully")
					}
				} else {
					logger.Debug("No Istio installation found or could not check")
				}
			}
		}

		// Delete the cluster
		logger.Info("Deleting cluster", "cluster", demoClusterName)
		if err := kindMgr.DeleteCluster(ctx, demoClusterName); err != nil {
			return fmt.Errorf("failed to delete cluster: %w", err)
		}

		logger.Info("Demo cluster stopped successfully", "cluster", demoClusterName)

		return nil
	},
}

func init() {
	// Add flags to start command
	demoStartCmd.Flags().StringVar(&demoClusterName, "name", "navigator-demo", "Name of the demo cluster")
	demoStartCmd.Flags().BoolVar(&demoCleanup, "cleanup", false, "Delete existing cluster if it exists")
	demoStartCmd.Flags().StringVar(&demoIstioVersion, "istio-version", "1.25.4", "Istio version to install")

	// Add flags to stop command
	demoStopCmd.Flags().StringVar(&demoClusterName, "name", "navigator-demo", "Name of the demo cluster")

	// Add subcommands to demo
	demoCmd.AddCommand(demoStartCmd)
	demoCmd.AddCommand(demoStopCmd)
}

// labelNamespaceForIstio labels a namespace for Istio injection using Kubernetes client-go
func labelNamespaceForIstio(clientset kubernetes.Interface, namespace string, logger *slog.Logger) error {
	logger.Info("Labeling namespace for Istio injection", "namespace", namespace)

	// Get the namespace first to ensure it exists
	ns, err := clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get namespace %s: %w", namespace, err)
	}

	// Add the Istio injection label
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels["istio-injection"] = "enabled"

	// Update the namespace
	_, err = clientset.CoreV1().Namespaces().Update(context.Background(), ns, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update namespace %s with Istio injection label: %w", namespace, err)
	}

	logger.Info("Namespace labeled for Istio injection", "namespace", namespace)
	return nil
}
