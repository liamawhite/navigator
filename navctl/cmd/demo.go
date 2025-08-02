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
	"path/filepath"

	"github.com/liamawhite/navigator/pkg/localenv/istio"
	"github.com/liamawhite/navigator/pkg/localenv/kind"
	"github.com/liamawhite/navigator/pkg/localenv/microservice"
	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/spf13/cobra"
)

var (
	demoClusterName           string
	demoCleanup               bool
	demoIstioVersion          string
	demoMicroserviceNamespace string
	demoMicroserviceScenario  string
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
	Short: "Start a demo Kind cluster with Istio service mesh",
	Long: `Start a demo Kind cluster for testing Navigator functionality.

This command creates a basic Kind cluster and installs Istio service mesh
for testing Navigator's service discovery and proxy analysis features.`,
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
		fmt.Printf("\nInstalling Istio %s...\n", demoIstioVersion)

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
		fmt.Printf("Istio %s installed successfully!\n", demoIstioVersion)

		// Install microservices if scenario is specified
		if demoMicroserviceScenario != "" {
			logger.Info("Installing microservices", "scenario", demoMicroserviceScenario)
			fmt.Printf("\nInstalling microservices (scenario: %s)...\n", demoMicroserviceScenario)

			// Create microservice Helm manager
			microHelmMgr, err := microservice.NewHelmManager(absKubeconfigPath, demoMicroserviceNamespace, logger)
			if err != nil {
				return fmt.Errorf("failed to create Helm manager for microservice installation: %w", err)
			}

			// Install microservices
			microConfig := microservice.DefaultMicroserviceConfig()
			microConfig.Namespace = demoMicroserviceNamespace
			microConfig.ReleaseName = "microservice-demo"
			microConfig.Scenario = demoMicroserviceScenario

			if err := microHelmMgr.InstallMicroservice(ctx, microConfig); err != nil {
				return fmt.Errorf("failed to install microservices: %w", err)
			}

			logger.Info("Microservices installed successfully")
			fmt.Printf("Microservices (scenario: %s) installed successfully!\n", demoMicroserviceScenario)
		}

		fmt.Printf("\nDemo cluster '%s' is ready!\n", demoClusterName)
		fmt.Printf("Kubeconfig exported to: %s\n", kubeconfigPath)
		fmt.Printf("Istio %s is installed and ready\n", demoIstioVersion)

		if demoMicroserviceScenario != "" {
			fmt.Printf("Microservices (scenario: %s) deployed in namespace %s\n", demoMicroserviceScenario, demoMicroserviceNamespace)
		}

		fmt.Printf("\nTo use this cluster:\n")
		fmt.Printf("  export KUBECONFIG=%s\n", kubeconfigPath)
		fmt.Printf("  kubectl get nodes\n")
		fmt.Printf("  kubectl get pods -n istio-system\n")

		if demoMicroserviceScenario != "" {
			fmt.Printf("  kubectl get pods -n %s\n", demoMicroserviceNamespace)
		}

		fmt.Printf("\nTo stop this cluster:\n")
		fmt.Printf("  navctl demo stop\n")

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
			fmt.Printf("Cluster '%s' does not exist or is already stopped.\n", demoClusterName)
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
			fmt.Printf("Cleaning up microservices if installed...\n")

			microHelmMgr, err := microservice.NewHelmManager(absKubeconfigPath, demoMicroserviceNamespace, logger)
			if err != nil {
				logger.Debug("Could not create microservice Helm manager for cleanup", "error", err)
			} else {
				// Check if microservice is installed
				if installed, version, err := microHelmMgr.IsMicroserviceInstalled(ctx, "microservice-demo"); err == nil && installed {
					logger.Info("Found microservice installation, cleaning up", "version", version)
					if err := microHelmMgr.UninstallMicroservice(ctx, "microservice-demo"); err != nil {
						logger.Warn("Failed to uninstall microservices", "error", err)
						fmt.Printf("Warning: Could not cleanly uninstall microservices: %v\n", err)
					} else {
						logger.Info("Microservices uninstalled successfully")
						fmt.Printf("Microservices uninstalled successfully\n")
					}
				} else {
					logger.Debug("No microservice installation found or could not check")
				}
			}

			// Then clean up Istio
			logger.Info("Attempting Istio cleanup")
			fmt.Printf("Cleaning up Istio if installed...\n")

			helmMgr, err := istio.NewHelmManager(absKubeconfigPath, "istio-system", logger)
			if err != nil {
				logger.Debug("Could not create Helm manager for cleanup", "error", err)
			} else {
				// Check if Istio is installed and get version
				if installed, version, err := helmMgr.IsIstioInstalled(ctx); err == nil && installed {
					logger.Info("Found Istio installation, cleaning up", "version", version)
					if err := helmMgr.UninstallIstio(ctx, version); err != nil {
						logger.Warn("Failed to uninstall Istio", "error", err)
						fmt.Printf("Warning: Could not cleanly uninstall Istio: %v\n", err)
					} else {
						logger.Info("Istio uninstalled successfully")
						fmt.Printf("Istio uninstalled successfully\n")
					}
				} else {
					logger.Debug("No Istio installation found or could not check")
				}
			}
		}

		// Delete the cluster
		fmt.Printf("Deleting cluster...\n")
		if err := kindMgr.DeleteCluster(ctx, demoClusterName); err != nil {
			return fmt.Errorf("failed to delete cluster: %w", err)
		}

		logger.Info("Demo cluster stopped successfully", "cluster", demoClusterName)
		fmt.Printf("Demo cluster '%s' stopped successfully.\n", demoClusterName)

		return nil
	},
}

func init() {
	// Add flags to start command
	demoStartCmd.Flags().StringVar(&demoClusterName, "name", "navigator-demo", "Name of the demo cluster")
	demoStartCmd.Flags().BoolVar(&demoCleanup, "cleanup", false, "Delete existing cluster if it exists")
	demoStartCmd.Flags().StringVar(&demoIstioVersion, "istio-version", "1.25.4", "Istio version to install")
	demoStartCmd.Flags().StringVar(&demoMicroserviceScenario, "microservice-scenario", "", "Microservice scenario to deploy (three-services)")
	demoStartCmd.Flags().StringVar(&demoMicroserviceNamespace, "microservice-namespace", "microservices", "Namespace for microservice deployment")

	// Add flags to stop command
	demoStopCmd.Flags().StringVar(&demoClusterName, "name", "navigator-demo", "Name of the demo cluster")
	demoStopCmd.Flags().StringVar(&demoMicroserviceNamespace, "microservice-namespace", "microservices", "Namespace for microservice cleanup")

	// Add subcommands to demo
	demoCmd.AddCommand(demoStartCmd)
	demoCmd.AddCommand(demoStopCmd)
}
