package cli

import (
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
	"k8s.io/client-go/util/homedir"

	"github.com/liamawhite/navigator/internal/grpc"
	"github.com/liamawhite/navigator/pkg/datastore/kubeconfig"
	"github.com/liamawhite/navigator/pkg/logging"
	troubleshootingKubeconfig "github.com/liamawhite/navigator/pkg/troubleshooting/kubeconfig"
)

var (
	logLevel       string
	logFormat      string
	rootPort       int
	kubeconfigPath string
	noUI           bool
	noBrowser      bool
)

var rootCmd = &cobra.Command{
	Use:   "navigator",
	Short: "Navigator is a Kubernetes service registry",
	Long: `Navigator provides a gRPC-based service registry for Kubernetes clusters.
It enables service discovery and provides Envoy configuration analysis capabilities.

When run without subcommands, Navigator starts the server with UI and opens a browser.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize global logger
		config := &logging.Config{
			Level:  logging.ParseLevel(logLevel),
			Format: logFormat,
		}
		logger := logging.NewLogger(config)
		slog.SetDefault(logger)

		// Log startup information
		logging.For(logging.ComponentCLI).Info("navigator starting",
			"version", "dev",
			"log_level", logLevel,
			"log_format", logFormat)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default behavior: start server with UI and open browser
		return runServerWithUI()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

// runServerWithUI starts the server with UI and optionally opens a browser
func runServerWithUI() error {
	logger := logging.For(logging.ComponentCLI)
	logger.Info("starting navigator server with UI",
		"port", rootPort,
		"kubeconfig", kubeconfigPath,
		"no_ui", noUI,
		"no_browser", noBrowser)

	// Create service datastore
	serviceDS, err := kubeconfig.New(kubeconfigPath)
	if err != nil {
		logger.Error("failed to create service datastore", "error", err)
		return fmt.Errorf("failed to create service datastore: %w", err)
	}

	// Create troubleshooting datastore
	troubleshootingDS, err := troubleshootingKubeconfig.New(kubeconfigPath)
	if err != nil {
		logger.Error("failed to create troubleshooting datastore", "error", err)
		return fmt.Errorf("failed to create troubleshooting datastore: %w", err)
	}

	// Create gRPC server with or without UI
	var server *grpc.Server
	if noUI {
		server, err = grpc.NewServer(serviceDS, troubleshootingDS, rootPort)
	} else {
		server, err = grpc.NewServerWithUI(serviceDS, troubleshootingDS, rootPort)
	}
	if err != nil {
		logger.Error("failed to create gRPC server", "error", err)
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	// Start server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- server.Start()
	}()

	// Open browser if UI is enabled and not disabled
	if !noUI && !noBrowser && server.UIEnabled() {
		go func() {
			// Wait a moment for server to start
			time.Sleep(1 * time.Second)
			url := fmt.Sprintf("http://localhost%s", server.UIAddress())
			logger.Info("opening browser", "url", url)
			if err := openBrowser(url); err != nil {
				logger.Warn("failed to open browser", "error", err, "url", url)
			}
		}()
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("navigator server started",
		"grpc_address", server.Address(),
		"http_address", server.HTTPAddress())

	if server.UIEnabled() {
		logger.Info("UI available", "ui_address", server.UIAddress())
	}

	select {
	case err := <-serverErrChan:
		logger.Error("server error", "error", err)
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		logger.Info("received shutdown signal", "signal", sig.String())
		server.Stop()
		logger.Info("server shutdown complete")
		return nil
	}
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

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format (text, json)")

	// Root command flags
	rootCmd.Flags().IntVarP(&rootPort, "port", "p", 8080, "Port to run the server on")

	// Default kubeconfig path
	defaultKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeconfig = filepath.Join(home, ".kube", "config")
	}
	rootCmd.Flags().StringVarP(&kubeconfigPath, "kubeconfig", "k", defaultKubeconfig, "Path to kubeconfig file")

	rootCmd.Flags().BoolVar(&noUI, "no-ui", false, "Disable UI server")
	rootCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser automatically")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(demoCmd)
}
