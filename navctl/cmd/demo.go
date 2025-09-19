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
	"os"
	"path/filepath"
	"sync"

	"github.com/liamawhite/navigator/pkg/localenv/database"
	"github.com/liamawhite/navigator/pkg/localenv/fortio"
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
	demoCleanup bool
	// kubeconfigMutex serializes operations that modify the kubeconfig file
	// This prevents concurrent access that causes locking issues
	kubeconfigMutex sync.Mutex
)

const (
	demoClusterName  = "navigator-demo"
	demoIstioVersion = "1.25.4"
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

		const clusterCount = 2 // Always create exactly 2 clusters

		// Create 2 clusters in parallel
		logger.Info("Starting parallel demo cluster creation", "count", clusterCount, "base_name", demoClusterName)

		type clusterResult struct {
			clusterName  string
			clusterIndex int
			err          error
		}

		resultCh := make(chan clusterResult, clusterCount)

		// Start both clusters in parallel
		for i := range clusterCount {
			clusterName := fmt.Sprintf("%s-%d", demoClusterName, i+1)
			clusterIndex := i

			go func(name string, index int) {
				logger.Info("Starting cluster creation", "cluster", name, "index", index+1)
				err := createSingleDemoCluster(ctx, name, index, logger)
				resultCh <- clusterResult{clusterName: name, clusterIndex: index, err: err}
			}(clusterName, clusterIndex)
		}

		// Collect results
		var successfulClusters []string
		var failures []error

		for range clusterCount {
			result := <-resultCh
			if result.err != nil {
				failures = append(failures, fmt.Errorf("cluster %s failed: %w", result.clusterName, result.err))
				logger.Error("Cluster creation failed", "cluster", result.clusterName, "error", result.err)
			} else {
				successfulClusters = append(successfulClusters, result.clusterName)
				logger.Info("✓ Cluster creation completed", "cluster", result.clusterName)
			}
		}

		// Report final results - fail if any cluster failed
		if len(failures) > 0 {
			logger.Error("Cluster creation failed", "successful", len(successfulClusters), "failed", len(failures))
			for _, err := range failures {
				logger.Error("Failure details", "error", err)
			}
			if len(successfulClusters) == 0 {
				return fmt.Errorf("all clusters failed to create")
			} else {
				return fmt.Errorf("%d out of %d clusters failed to create", len(failures), clusterCount)
			}
		}

		logger.Info("🎉 Parallel demo cluster creation completed!",
			"successful", len(successfulClusters),
			"failed", len(failures),
			"clusters", successfulClusters)

		// Print summary of all successful clusters
		fmt.Printf("\n🎉 Successfully created %d demo clusters:\n", len(successfulClusters))
		for i, clusterName := range successfulClusters {
			portOffset := i * 1000
			httpPort := kind.HTTPNodePort + portOffset
			prometheusPort := kind.PrometheusNodePort + portOffset

			fmt.Printf("\n📦 Cluster: %s\n", clusterName)
			fmt.Printf("   🧪 Test URL: http://localhost:%d\n", httpPort)
			fmt.Printf("   📊 Prometheus: http://localhost:%d\n", prometheusPort)
			fmt.Printf("   📄 Kubeconfig: %s-kubeconfig\n", clusterName)
		}
		fmt.Printf("\n🚀 To start Navigator with both demo clusters and metrics enabled:\n")
		fmt.Printf("   navctl local --demo\n\n")

		return nil
	},
}

// demoStopCmd represents the demo stop command
var demoStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop demo Kind clusters",
	Long: `Stop (delete) demo Kind clusters.

This command deletes the specified Kind cluster(s) and cleans up associated resources.
If --count is specified, it will stop multiple clusters with numbered suffixes.
Otherwise, it will stop the single cluster with the specified name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.For("demo")
		ctx := context.Background()

		// Stop both demo clusters
		const clusterCount = 2
		logger.Info("Stopping demo clusters", "count", clusterCount, "base_name", demoClusterName)

		var clustersToStop []string
		for i := 0; i < clusterCount; i++ {
			clusterName := fmt.Sprintf("%s-%d", demoClusterName, i+1)
			clustersToStop = append(clustersToStop, clusterName)
		}

		if len(clustersToStop) == 0 {
			logger.Info("No demo clusters found to stop", "base_name", demoClusterName)
			return nil
		}

		logger.Info("Found clusters to stop", "clusters", clustersToStop)

		// Stop clusters sequentially to avoid kubeconfig lock conflicts
		var successfulStops []string
		var failures []error

		for _, clusterName := range clustersToStop {
			logger.Info("Stopping cluster", "cluster", clusterName)
			err := stopSingleDemoCluster(ctx, clusterName, logger)
			if err != nil {
				failures = append(failures, fmt.Errorf("cluster %s: %w", clusterName, err))
				logger.Error("Cluster stop failed", "cluster", clusterName, "error", err)
			} else {
				successfulStops = append(successfulStops, clusterName)
				logger.Info("✓ Cluster stopped successfully", "cluster", clusterName)
			}
		}

		// Report final results
		if len(failures) > 0 {
			logger.Error("Some clusters failed to stop", "successful", len(successfulStops), "failed", len(failures))
			for _, err := range failures {
				logger.Error("Stop failure details", "error", err)
			}
		}

		logger.Info("Demo cluster stop completed!",
			"total_requested", len(clustersToStop),
			"successful", len(successfulStops),
			"failed", len(failures),
			"stopped_clusters", successfulStops)

		if len(failures) > 0 && len(successfulStops) == 0 {
			return fmt.Errorf("all clusters failed to stop")
		}

		return nil
	},
}

func init() {
	// Add flags to start command
	demoStartCmd.Flags().BoolVar(&demoCleanup, "cleanup", false, "Delete existing clusters if they exist")

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

// createSingleDemoCluster creates and configures a single demo cluster
func createSingleDemoCluster(ctx context.Context, clusterName string, clusterIndex int, logger *slog.Logger) error {
	logger.Info("Starting demo cluster creation", "cluster", clusterName, "index", clusterIndex)

	kindMgr := kind.NewKindManager(logger)

	// Check if cluster already exists
	exists, err := kindMgr.ClusterExists(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}

	if exists {
		logger.Info("Cluster already exists", "cluster", clusterName)
		if !demoCleanup {
			return fmt.Errorf("cluster %s already exists, use --cleanup to recreate", clusterName)
		}

		logger.Info("Deleting existing cluster", "cluster", clusterName)
		
		// Serialize cluster deletion to prevent kubeconfig locking conflicts
		// when multiple clusters are being deleted in parallel
		kubeconfigMutex.Lock()
		err := kindMgr.DeleteCluster(ctx, clusterName)
		kubeconfigMutex.Unlock()
		
		if err != nil {
			return fmt.Errorf("failed to delete existing cluster: %w", err)
		}
	}

	// Create the cluster with unique port mappings for parallel clusters
	config := kind.DemoKindConfigWithPorts(clusterName, clusterIndex)

	// Serialize cluster creation to prevent kubeconfig locking conflicts
	kubeconfigMutex.Lock()
	err = kindMgr.CreateCluster(ctx, config)
	kubeconfigMutex.Unlock()
	
	if err != nil {
		return fmt.Errorf("failed to create demo cluster: %w", err)
	}

	// Wait for cluster to be ready
	logger.Info("Waiting for cluster to be ready...", "cluster", clusterName)
	if err := kindMgr.WaitForClusterReady(ctx, clusterName); err != nil {
		return fmt.Errorf("cluster failed to become ready: %w", err)
	}

	// Export kubeconfig
	kubeconfigPath := fmt.Sprintf("%s-kubeconfig", clusterName)
	if err := kindMgr.ExportKubeconfig(ctx, clusterName, kubeconfigPath); err != nil {
		logger.Warn("Failed to export kubeconfig", "cluster", clusterName, "error", err)
	} else {
		logger.Info("Kubeconfig exported", "cluster", clusterName, "path", kubeconfigPath)
	}

	logger.Info("Demo cluster created successfully!", "cluster", clusterName, "kubeconfig", kubeconfigPath)

	// Install Istio
	logger.Info("Installing Istio service mesh", "cluster", clusterName, "version", demoIstioVersion)

	// Get absolute path for kubeconfig
	absKubeconfigPath, err := filepath.Abs(kubeconfigPath)
	if err != nil {
		logger.Warn("Failed to get absolute kubeconfig path", "cluster", clusterName, "error", err)
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

	logger.Info("Istio installed successfully", "cluster", clusterName)

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
	logger.Info("Labeling default namespace for Istio injection", "cluster", clusterName)
	if err := labelNamespaceForIstio(clientset, "default", logger); err != nil {
		return fmt.Errorf("failed to label default namespace for Istio injection: %w", err)
	}

	// Install microservices and database in parallel
	logger.Info("Installing microservices and database in parallel", "cluster", clusterName, "scenario", "three-tier")

	// Create managers
	microKustomizeMgr, err := microservice.NewKustomizeManager(absKubeconfigPath, logger)
	if err != nil {
		return fmt.Errorf("failed to create Kustomize manager for microservice installation: %w", err)
	}

	dbKustomizeMgr, err := database.NewKustomizeManager(absKubeconfigPath, logger)
	if err != nil {
		return fmt.Errorf("failed to create Kustomize manager for database installation: %w", err)
	}

	// Install both components in parallel using goroutines
	type installResult struct {
		component string
		err       error
	}

	resultCh := make(chan installResult, 2)

	// Install microservices
	go func() {
		logger.Info("Starting microservices installation...", "cluster", clusterName)
		err := microKustomizeMgr.InstallMicroservice(ctx)
		resultCh <- installResult{component: "microservices", err: err}
	}()

	// Install database
	go func() {
		logger.Info("Starting database installation...", "cluster", clusterName)
		err := dbKustomizeMgr.InstallDatabase(ctx)
		resultCh <- installResult{component: "database", err: err}
	}()

	// Wait for both installations to complete
	var microErr, dbErr error
	for i := 0; i < 2; i++ {
		result := <-resultCh
		switch result.component {
		case "microservices":
			microErr = result.err
			if microErr == nil {
				logger.Info("✓ Microservices installed successfully", "cluster", clusterName)
			}
		case "database":
			dbErr = result.err
			if dbErr == nil {
				logger.Info("✓ Database installed successfully", "cluster", clusterName)
			}
		}
	}

	// Check for any installation errors
	if microErr != nil {
		return fmt.Errorf("failed to install microservices: %w", microErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to install database: %w", dbErr)
	}

	logger.Info("All components installed successfully", "cluster", clusterName)

	// Verify the microservice chain is working
	logger.Info("Verifying microservice connectivity...", "cluster", clusterName)

	// Calculate ports for this cluster
	portOffset := clusterIndex * 1000
	httpPort := kind.HTTPNodePort + portOffset
	httpsPort := kind.HTTPSNodePort + portOffset
	statusPort := kind.StatusNodePort + portOffset
	prometheusPort := kind.PrometheusNodePort + portOffset

	// Verify Istio gateway readiness
	logger.Info("Step 1/3: Verifying Istio gateway readiness...", "cluster", clusterName)
	if err := helmMgr.VerifyIstioGateway(ctx); err != nil {
		logger.Error("Istio gateway verification failed", "cluster", clusterName, "error", err)

		logger.Info("Demo cluster ready but verification incomplete - try manual testing",
			"cluster", clusterName,
			"kubeconfig", kubeconfigPath,
			"istio_version", demoIstioVersion,
			"microservices_namespace", "microservices",
			"http_port", httpPort)
		return nil
	}
	logger.Info("✓ Istio gateway verification successful", "cluster", clusterName)

	// Verify Prometheus addon
	logger.Info("Step 2/3: Verifying Prometheus addon availability...", "cluster", clusterName)
	promMgr := istio.NewPrometheusManager(absKubeconfigPath, "istio-system", logger)
	if installed, err := promMgr.IsPrometheusInstalled(ctx); err != nil {
		logger.Warn("Could not verify Prometheus installation", "cluster", clusterName, "error", err)
	} else if !installed {
		logger.Warn("Prometheus addon not found - metrics collection may be limited", "cluster", clusterName)
	} else {
		logger.Info("✓ Prometheus addon verification successful", "cluster", clusterName)
	}

	// Verify microservice chain (including database connectivity)
	logger.Info("Step 3/3: Verifying microservice request chain...", "cluster", clusterName)

	if err := microKustomizeMgr.VerifyMicroserviceChainWithPort(ctx, httpPort); err != nil {
		logger.Error("Microservice verification failed", "cluster", clusterName, "error", err)
		return fmt.Errorf("microservice verification failed: %w", err)
	}
	logger.Info("✓ Microservice verification successful - full chain working!", "cluster", clusterName)

	// Start Fortio load generation
	logger.Info("Starting continuous load generation at 5 RPS...", "cluster", clusterName)
	fortioMgr := fortio.NewFortioManager(absKubeconfigPath, "load-generator", logger)
	if err := fortioMgr.InstallFortio(ctx); err != nil {
		logger.Warn("Failed to start Fortio load generator", "cluster", clusterName, "error", err)
	} else {
		logger.Info("✓ Load generation started - 5 RPS through full microservice chain", "cluster", clusterName)
	}

	logger.Info("🎉 Demo cluster ready and verified!",
		"cluster", clusterName,
		"kubeconfig", kubeconfigPath,
		"istio_version", demoIstioVersion,
		"microservices_namespace", "microservices",
		"database_namespace", "database",
		"test_url", "Gateway -> Frontend -> Backend -> Database chain verified",
		"http_port", httpPort,
		"https_port", httpsPort,
		"status_port", statusPort,
		"prometheus_port", prometheusPort)

	// Print curl examples for manual testing after all structured logging (only for first cluster)
	if clusterIndex == 0 {
		fmt.Printf("\n🧪 Test the microservice chain manually:\n")
		fmt.Printf("   curl -s \"http://localhost:%d/proxy/backend:8080/proxy/database.database:8080\"\n", httpPort)
		fmt.Printf("\n🔍 Test individual services:\n")
		fmt.Printf("   curl -s \"http://localhost:%d\"                              # Frontend only\n", httpPort)
		fmt.Printf("   curl -s \"http://localhost:%d/proxy/backend:8080\"             # Frontend -> Backend\n", httpPort)
		fmt.Printf("   curl -s \"http://localhost:%d/proxy/backend:8080/proxy/database.database:8080\" # Full chain\n", httpPort)
		fmt.Printf("\n📊 Access Prometheus metrics:\n")
		fmt.Printf("   http://localhost:%d\n", prometheusPort)
		fmt.Printf("\n")
	}

	return nil
}

// stopSingleDemoCluster stops and cleans up a single demo cluster
func stopSingleDemoCluster(ctx context.Context, clusterName string, logger *slog.Logger) error {
	logger.Info("Stopping demo cluster", "cluster", clusterName)

	kindMgr := kind.NewKindManager(logger)

	// Check if cluster exists
	exists, err := kindMgr.ClusterExists(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}

	if !exists {
		logger.Info("Cluster does not exist", "cluster", clusterName)
		return nil
	}

	// Delete the cluster directly using Kind library - this will clean up everything
	logger.Info("Deleting cluster", "cluster", clusterName)
	if err := kindMgr.DeleteCluster(ctx, clusterName); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	// Clean up the kubeconfig file if it exists (best effort)
	kubeconfigPath := fmt.Sprintf("%s-kubeconfig", clusterName)
	if _, err := os.Stat(kubeconfigPath); err == nil {
		logger.Debug("Removing kubeconfig file", "cluster", clusterName, "path", kubeconfigPath)
		if removeErr := os.Remove(kubeconfigPath); removeErr != nil {
			logger.Debug("Could not remove kubeconfig file", "cluster", clusterName, "path", kubeconfigPath, "error", removeErr)
		}
	}

	logger.Info("Demo cluster stopped successfully", "cluster", clusterName)

	return nil
}
