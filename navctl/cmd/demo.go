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
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/liamawhite/navigator/navctl/pkg/demo"
	"github.com/liamawhite/navigator/pkg/logging"
)

var (
	demoClusterName string
	demoTimeout     time.Duration
)

// demoCmd represents the demo command
var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Manage demonstration environments",
	Long: `Manage demonstration environments using Kind clusters.

This command allows you to create, manage, and teardown demo environments
with predefined scenarios for showcasing Navigator's capabilities.

Available subcommands:
  start   - Create and start a demo environment
  stop    - Stop and remove the demo environment  
  list    - List available demo scenarios
  status  - Show status of current demo environment`,
}

// demoStartCmd represents the demo start command
var demoStartCmd = &cobra.Command{
	Use:   "start [scenario]",
	Short: "Start a demo environment with the specified scenario",
	Long: `Start a demo environment with the specified scenario.

If no scenario is specified, the 'basic' scenario will be used.
Use 'navctl demo list' to see all available scenarios.

Examples:
  navctl demo start                    # Start with basic scenario
  navctl demo start istio-demo         # Start with Istio demo scenario
  navctl demo start complex-topology   # Start complex microservice demo`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDemoStart,
}

// demoStopCmd represents the demo stop command
var demoStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop and remove the demo environment",
	Long: `Stop and remove the demo environment.

This will delete the Kind cluster and clean up all resources.`,
	RunE: runDemoStop,
}

// demoListCmd represents the demo list command
var demoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available demo scenarios",
	Long: `List all available demo scenarios with their descriptions.

Each scenario provides a different configuration of services to demonstrate
various Navigator capabilities.`,
	RunE: runDemoList,
}

// demoStatusCmd represents the demo status command
var demoStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of current demo environment",
	Long: `Show the status of the current demo environment.

This displays information about the demo cluster, its readiness,
and configuration details.`,
	RunE: runDemoStatus,
}

func runDemoStart(cmd *cobra.Command, args []string) error {
	logger := logging.For("navctl-demo")

	// Determine scenario to use
	scenarioName := "basic"
	if len(args) > 0 {
		scenarioName = args[0]
	}

	// Validate scenario exists
	if err := demo.ValidateScenarioName(scenarioName); err != nil {
		return err
	}

	// Get scenario info
	scenarioInfo, err := demo.GetScenarioInfo(scenarioName)
	if err != nil {
		return fmt.Errorf("failed to get scenario info: %w", err)
	}

	// All scenarios now have Istio enabled by default
	istioEnabled := scenarioInfo.IstioEnabled

	logger.Info("starting demo environment",
		"scenario", scenarioName,
		"cluster", demoClusterName,
		"istio_enabled", istioEnabled)

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), demoTimeout)
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			logger.Info("received shutdown signal, canceling demo setup")
			cancel()
		case <-ctx.Done():
		}
	}()

	// Create cluster manager
	manager := demo.NewClusterManager()

	// Ensure cluster exists
	if err := manager.EnsureCluster(ctx, demoClusterName, istioEnabled); err != nil {
		return fmt.Errorf("failed to ensure demo cluster: %w", err)
	}

	// Deploy scenario
	if err := manager.DeployScenario(ctx, scenarioName); err != nil {
		return fmt.Errorf("failed to deploy scenario: %w", err)
	}

	// Get cluster info
	info := manager.GetClusterInfo(ctx)
	if info == nil {
		return fmt.Errorf("failed to get cluster info")
	}

	fmt.Printf("\n✓ Demo environment started successfully!\n\n")
	fmt.Printf("Cluster: %s\n", info.Name)
	fmt.Printf("Scenario: %s (%s)\n", scenarioName, scenarioInfo.Description)
	fmt.Printf("Namespace: %s\n", info.Namespace)
	fmt.Printf("Services: %s\n", strings.Join(scenarioInfo.Services, ", "))
	if info.IstioEnabled {
		fmt.Printf("Istio: enabled\n")
	} else {
		fmt.Printf("Istio: disabled\n")
	}
	fmt.Printf("Kubeconfig: %s\n", info.Kubeconfig)
	fmt.Printf("\nTo start Navigator and view the demo:\n")
	fmt.Printf("  navctl local --kube-config %s\n", info.Kubeconfig)
	fmt.Printf("\nTo stop the demo:\n")
	fmt.Printf("  navctl demo stop\n")

	return nil
}

func runDemoStop(cmd *cobra.Command, args []string) error {
	logger := logging.For("navctl-demo")

	logger.Info("stopping demo environment", "cluster", demoClusterName)

	ctx, cancel := context.WithTimeout(context.Background(), demoTimeout)
	defer cancel()

	// Create cluster manager
	manager := demo.NewClusterManager()
	manager.SetConfig(demoClusterName)

	// Check if cluster exists
	if !manager.IsReady(ctx) {
		fmt.Printf("Demo cluster '%s' is not running\n", demoClusterName)
		return nil
	}

	// Teardown cluster
	if err := manager.Teardown(ctx); err != nil {
		return fmt.Errorf("failed to teardown demo cluster: %w", err)
	}

	fmt.Printf("✓ Demo environment stopped and cleaned up\n")
	return nil
}

func runDemoList(cmd *cobra.Command, args []string) error {
	fmt.Print(demo.FormatScenarioList())
	fmt.Printf("Usage: navctl demo start [scenario]\n")
	return nil
}

func runDemoStatus(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create cluster manager
	manager := demo.NewClusterManager()
	manager.SetConfig(demoClusterName)

	// Get cluster info
	info := manager.GetClusterInfo(ctx)
	if info == nil {
		fmt.Printf("Demo cluster '%s' not found or not configured\n", demoClusterName)
		return nil
	}

	fmt.Printf("Demo Environment Status:\n\n")
	fmt.Printf("Cluster: %s\n", info.Name)
	fmt.Printf("Namespace: %s\n", info.Namespace)
	if info.Ready {
		fmt.Printf("Status: ✓ Ready\n")
	} else {
		fmt.Printf("Status: ✗ Not Ready\n")
	}
	if info.IstioEnabled {
		fmt.Printf("Istio: enabled\n")
	} else {
		fmt.Printf("Istio: disabled\n")
	}
	if info.Kubeconfig != "" {
		fmt.Printf("Kubeconfig: %s\n", info.Kubeconfig)
	}

	if info.Ready {
		fmt.Printf("\nTo view the demo:\n")
		fmt.Printf("  navctl local --kube-config %s\n", info.Kubeconfig)
	} else {
		fmt.Printf("\nTo start a demo:\n")
		fmt.Printf("  navctl demo start [scenario]\n")
	}

	return nil
}

func init() {
	// Add subcommands
	demoCmd.AddCommand(demoStartCmd)
	demoCmd.AddCommand(demoStopCmd)
	demoCmd.AddCommand(demoListCmd)
	demoCmd.AddCommand(demoStatusCmd)

	// Demo start flags
	demoStartCmd.Flags().DurationVar(&demoTimeout, "timeout", 10*time.Minute, "Timeout for demo setup")

	// Demo stop flags
	demoStopCmd.Flags().DurationVar(&demoTimeout, "timeout", 5*time.Minute, "Timeout for demo teardown")

	// Global demo flags
	demoCmd.PersistentFlags().StringVar(&demoClusterName, "cluster-name", "navigator-demo", "Name of the demo cluster")

	// Add to root command
	rootCmd.AddCommand(demoCmd)
}
