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
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/liamawhite/navigator/edge/pkg/config"
	"github.com/liamawhite/navigator/edge/pkg/kubernetes"
	"github.com/liamawhite/navigator/edge/pkg/proxy"
	edgeService "github.com/liamawhite/navigator/edge/pkg/service"
	managerConfig "github.com/liamawhite/navigator/manager/pkg/config"
	"github.com/liamawhite/navigator/manager/pkg/connections"
	managerService "github.com/liamawhite/navigator/manager/pkg/service"
	"github.com/liamawhite/navigator/navctl/pkg/ui"
	"github.com/liamawhite/navigator/pkg/istio/proxy/client"
	"github.com/liamawhite/navigator/pkg/logging"
)

var (
	kubeconfig     string
	managerPort    int
	managerHost    string
	maxMessageSize int
	disableUI      bool
	uiPort         int
	noBrowser      bool
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Run manager and edge services locally",
	Long: `Run both Navigator manager and edge services locally.
This command starts the manager service first, then connects an edge service to it.
Both services will use the same kubeconfig file to discover and serve services
from the local Kubernetes cluster.`,
	RunE: runLocal,
}

func runLocal(cmd *cobra.Command, args []string) error {
	logger := logging.For("navctl-local")

	// Validate kubeconfig exists
	if err := validateKubeconfig(); err != nil {
		return fmt.Errorf("kubeconfig validation failed: %w", err)
	}

	// Extract cluster ID to show in startup logs
	clusterID, err := extractClusterIDFromKubeconfig()
	if err != nil {
		logger.Warn("failed to extract cluster ID from kubeconfig, using fallback", "error", err)
		clusterID = "local-cluster"
	}

	logger.Info("starting Navigator services",
		"kubeconfig", kubeconfig,
		"cluster_id", clusterID,
		"manager_port", managerPort,
		"manager_host", managerHost)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start manager service
	managerSvc, err := startManagerService(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to start manager service: %w", err)
	}
	defer func() {
		logger.Info("stopping manager service")
		if err := managerSvc.Stop(); err != nil {
			logger.Error("error stopping manager service", "error", err)
		}
	}()

	// Wait a moment for manager to start
	time.Sleep(2 * time.Second)

	// Start edge service
	edgeSvc, err := startEdgeService(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to start edge service: %w", err)
	}
	defer func() {
		logger.Info("stopping edge service")
		if err := edgeSvc.Stop(); err != nil {
			logger.Error("error stopping edge service", "error", err)
		}
	}()

	// Start UI server unless disabled
	var uiSvc *ui.Server
	if !disableUI {
		uiSvc, err = startUIServer(ctx, logger)
		if err != nil {
			return fmt.Errorf("failed to start UI server: %w", err)
		}
		defer func() {
			logger.Info("stopping UI server")
			if err := uiSvc.Stop(); err != nil {
				logger.Error("error stopping UI server", "error", err)
			}
		}()
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Navigator services started successfully")
	logger.Info("manager gRPC server listening", "port", managerPort)
	logger.Info("manager HTTP gateway listening", "port", managerPort+1)

	if !disableUI {
		logger.Info("UI server listening", "port", uiPort)
		if !noBrowser {
			// Open browser after a short delay
			go func() {
				time.Sleep(1 * time.Second)
				url := fmt.Sprintf("http://localhost:%d", uiPort)
				logger.Info("opening browser", "url", url)
				if err := openBrowser(url); err != nil {
					logger.Warn("failed to open browser", "error", err, "url", url)
				}
			}()
		}
	}

	logger.Info("press Ctrl+C to stop")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		logger.Info("context canceled")
	case sig := <-sigChan:
		logger.Info("received shutdown signal", "signal", sig.String())
		cancel()
	}

	logger.Info("shutting down Navigator services")
	return nil
}

func startManagerService(ctx context.Context, logger *slog.Logger) (*managerService.ManagerService, error) {
	// Create manager config
	cfg := &managerConfig.Config{
		Port:           managerPort,
		LogLevel:       logLevel,
		LogFormat:      logFormat,
		MaxMessageSize: maxMessageSize,
	}

	// Create connections manager
	connectionManager := connections.NewManager(logging.For("manager"))

	// Create manager service
	managerSvc, err := managerService.NewManagerService(cfg, connectionManager, logging.For("manager"))
	if err != nil {
		return nil, fmt.Errorf("failed to create manager service: %w", err)
	}

	// Start manager service in goroutine
	go func() {
		if err := managerSvc.Start(); err != nil {
			logger.Error("manager service error", "error", err)
		}
	}()

	return managerSvc, nil
}

func startEdgeService(ctx context.Context, logger *slog.Logger) (*edgeService.EdgeService, error) {
	// Extract cluster ID from kubeconfig
	clusterID, err := extractClusterIDFromKubeconfig()
	if err != nil {
		logger.Warn("failed to extract cluster ID from kubeconfig, using fallback", "error", err)
		clusterID = "local-cluster"
	} else {
		logger.Info("extracted cluster ID from kubeconfig", "cluster_id", clusterID)
	}

	// Create edge config
	cfg := &config.Config{
		KubeconfigPath:  kubeconfig,
		ManagerEndpoint: fmt.Sprintf("%s:%d", managerHost, managerPort),
		ClusterID:       clusterID,
		SyncInterval:    30,
		LogLevel:        logLevel,
		LogFormat:       logFormat,
		MaxMessageSize:  maxMessageSize,
	}

	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient(cfg.KubeconfigPath, logging.For("edge-k8s"))
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create admin client for proxy configuration access
	adminClient := client.NewAdminClient(k8sClient.GetClientset(), k8sClient.GetRestConfig())

	// Create proxy service
	proxyService := proxy.NewProxyService(adminClient, logging.For("edge-proxy"))

	// Create edge service
	edgeSvc, err := edgeService.NewEdgeService(cfg, k8sClient, proxyService, logging.For("edge"))
	if err != nil {
		return nil, fmt.Errorf("failed to create edge service: %w", err)
	}

	// Start edge service in goroutine
	go func() {
		if err := edgeSvc.Start(); err != nil {
			logger.Error("edge service error", "error", err)
		}
	}()

	return edgeSvc, nil
}

func startUIServer(ctx context.Context, logger *slog.Logger) (*ui.Server, error) {
	// Create UI server
	uiSvc, err := ui.NewServer(uiPort, managerPort+1) // API port is manager port + 1
	if err != nil {
		return nil, fmt.Errorf("failed to create UI server: %w", err)
	}

	// Start UI server in goroutine
	go func() {
		if err := uiSvc.Start(); err != nil {
			logger.Error("UI server error", "error", err)
		}
	}()

	return uiSvc, nil
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func validateKubeconfig() error {
	if kubeconfig == "" {
		return fmt.Errorf("kubeconfig path is required")
	}

	// Check if file exists
	if _, err := os.Stat(kubeconfig); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("kubeconfig file does not exist: %s", kubeconfig)
		}
		return fmt.Errorf("cannot access kubeconfig file: %w", err)
	}

	return nil
}

// extractClusterIDFromKubeconfig extracts the cluster name from the current context in kubeconfig
func extractClusterIDFromKubeconfig() (string, error) {
	// Load the kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Get the current context
	currentContext := config.CurrentContext
	if currentContext == "" {
		return "", fmt.Errorf("no current context set in kubeconfig")
	}

	// Get the context details
	context, exists := config.Contexts[currentContext]
	if !exists {
		return "", fmt.Errorf("current context '%s' not found in kubeconfig", currentContext)
	}

	// Get the cluster name
	clusterName := context.Cluster
	if clusterName == "" {
		return "", fmt.Errorf("no cluster set in current context '%s'", currentContext)
	}

	return clusterName, nil
}

func init() {
	// Default kubeconfig path
	defaultKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Command flags
	localCmd.Flags().StringVarP(&kubeconfig, "kube-config", "k", defaultKubeconfig, "Path to kubeconfig file")
	localCmd.Flags().IntVar(&managerPort, "manager-port", 8080, "Port for manager service")
	localCmd.Flags().StringVar(&managerHost, "manager-host", "localhost", "Host for manager service")
	localCmd.Flags().IntVar(&maxMessageSize, "max-message-size", 10, "Maximum gRPC message size in MB")
	localCmd.Flags().BoolVar(&disableUI, "disable-ui", false, "Disable UI server")
	localCmd.Flags().IntVar(&uiPort, "ui-port", 8082, "Port for UI server")
	localCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser automatically")

	// kube-config is optional with default value
}
