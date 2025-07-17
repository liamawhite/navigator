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

package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/liamawhite/navigator/pkg/localenv"
	"github.com/liamawhite/navigator/pkg/logging"
)

var (
	scenarioName    string
	clusterName     string
	namespace       string
	port            int
	httpPort        int
	withIstio       bool
	cleanupOnExit   bool
	waitTimeout     string
	listScenarios   bool
	teardownCluster bool
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Manage local development environments with Kind clusters",
	Long: `The demo command helps you set up and manage local development environments
using Kind (Kubernetes in Docker) with predefined scenarios of microservices.

This command can:
- Set up Kind clusters with Navigator and sample microservices
- Deploy predefined scenarios showcasing different use cases
- Manage cluster lifecycle and cleanup
- List available scenarios

Examples:
  navigator demo --scenario basic
  navigator demo --scenario microservice-topology --with-istio
  navigator demo list
  navigator demo teardown`,
	RunE: runDemo,
}

func runDemo(cmd *cobra.Command, args []string) error {
	logger := logging.For(logging.ComponentCLI)

	// Handle special subcommands
	if listScenarios {
		return listAvailableScenarios()
	}

	if teardownCluster {
		return teardownDemo()
	}

	// Validate scenario
	scenario := localenv.GetScenarioByName(scenarioName)
	if scenario == nil {
		return fmt.Errorf("unknown scenario: %s. Use 'navigator demo list' to see available scenarios", scenarioName)
	}

	logger.Info("starting demo environment setup",
		"scenario", scenario.Name,
		"cluster", clusterName,
		"namespace", namespace,
		"istio", withIstio || scenario.IstioEnabled)

	// Create configuration
	config := &localenv.Config{
		ClusterName:   clusterName,
		Namespace:     namespace,
		Port:          port,
		HTTPPort:      httpPort,
		IstioEnabled:  withIstio || scenario.IstioEnabled,
		CleanupOnExit: cleanupOnExit,
		WaitTimeout:   waitTimeout,
	}

	// Create Kind environment
	env := localenv.NewKindEnvironment()

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), parseTimeout(waitTimeout))
	defer cancel()

	// Setup environment
	logger.Info("setting up Kind cluster and Navigator server...")
	if err := env.Setup(ctx, config); err != nil {
		return fmt.Errorf("failed to setup environment: %w", err)
	}

	// Deploy scenario
	logger.Info("deploying scenario", "name", scenario.Name, "services", len(scenario.Services))
	if err := env.DeployScenario(ctx, scenario); err != nil {
		return fmt.Errorf("failed to deploy scenario: %w", err)
	}

	// Check if environment is ready
	logger.Info("verifying environment is ready...")
	if !env.IsReady(ctx) {
		return fmt.Errorf("environment not ready")
	}

	// Print success information
	printEnvironmentInfo(config, scenario)

	// Handle cleanup on exit if requested
	if cleanupOnExit {
		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		logger.Info("demo environment ready! Press Ctrl+C to clean up and exit")

		// Wait for signal
		<-sigChan

		logger.Info("cleaning up demo environment...")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := env.Teardown(ctx); err != nil {
			logger.Error("failed to clean up environment", "error", err)
			return err
		}

		logger.Info("demo environment cleaned up successfully")
	} else {
		logger.Info("demo environment ready! Use 'navigator demo teardown' to clean up when done")
	}

	return nil
}

func listAvailableScenarios() error {
	scenarios := localenv.ListScenarios()

	fmt.Println("Available demo scenarios:")
	fmt.Println()

	for _, scenario := range scenarios {
		fmt.Printf("  %s\n", scenario.Name)
		fmt.Printf("    %s\n", scenario.Description)
		fmt.Printf("    Services: %d\n", len(scenario.Services))
		if scenario.IstioEnabled {
			fmt.Printf("    Istio: enabled\n")
		}
		fmt.Println()
	}

	return nil
}

func teardownDemo() error {
	logger := logging.For(logging.ComponentCLI)

	logger.Info("tearing down demo environment", "cluster", clusterName)

	env := localenv.NewKindEnvironment()
	config := &localenv.Config{
		ClusterName: clusterName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// We only need to set the cluster name for teardown
	env.SetConfig(config)

	if err := env.Teardown(ctx); err != nil {
		return fmt.Errorf("failed to teardown environment: %w", err)
	}

	logger.Info("demo environment torn down successfully")
	return nil
}

func printEnvironmentInfo(config *localenv.Config, scenario *localenv.Scenario) {
	fmt.Println()
	fmt.Println("🎉 Demo environment is ready!")
	fmt.Println()
	fmt.Printf("  Scenario: %s\n", scenario.Name)
	fmt.Printf("  Cluster:  %s\n", config.ClusterName)
	fmt.Printf("  Namespace: %s\n", config.Namespace)
	fmt.Println()
	fmt.Printf("  Services deployed: %d\n", len(scenario.Services))
	for _, service := range scenario.Services {
		fmt.Printf("    - %s (%s, %d replicas)\n", service.Name, service.Type, service.Replicas)
	}
	fmt.Println()

	if config.IstioEnabled {
		fmt.Println("  🔧 Istio service mesh is enabled")
		fmt.Println()
	}

	fmt.Println("  To start Navigator and explore the services:")
	fmt.Printf("    ./navigator serve --kubeconfig $(kind get kubeconfig-path --name %s) --port %d\n", config.ClusterName, config.Port)
	fmt.Printf("  Then visit: http://localhost:%d\n", config.HTTPPort)
	fmt.Println()
}

func parseTimeout(timeout string) time.Duration {
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return 5 * time.Minute // Default fallback
	}
	return duration
}

func init() {
	demoCmd.Flags().StringVar(&scenarioName, "scenario", "basic", "Scenario to deploy (use 'list' to see available scenarios)")
	demoCmd.Flags().StringVar(&clusterName, "cluster", "navigator-demo", "Name of the Kind cluster")
	demoCmd.Flags().StringVar(&namespace, "namespace", "demo", "Kubernetes namespace for services")
	demoCmd.Flags().IntVar(&port, "port", 8080, "Port for Navigator gRPC server")
	demoCmd.Flags().IntVar(&httpPort, "http-port", 8081, "Port for Navigator HTTP server")
	demoCmd.Flags().BoolVar(&withIstio, "with-istio", false, "Enable Istio service mesh")
	demoCmd.Flags().BoolVar(&cleanupOnExit, "cleanup-on-exit", true, "Clean up resources when exiting")
	demoCmd.Flags().StringVar(&waitTimeout, "timeout", "5m", "Timeout for cluster setup and service deployment")
	demoCmd.Flags().BoolVar(&listScenarios, "list", false, "List available scenarios")
	demoCmd.Flags().BoolVar(&teardownCluster, "teardown", false, "Teardown existing demo cluster")

	// Make scenario and list mutually exclusive in help text
	demoCmd.MarkFlagsMutuallyExclusive("scenario", "list")
	demoCmd.MarkFlagsMutuallyExclusive("scenario", "teardown")
	demoCmd.MarkFlagsMutuallyExclusive("list", "teardown")
}
