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
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	edgeConfig "github.com/liamawhite/navigator/edge/pkg/config"
	"github.com/liamawhite/navigator/edge/pkg/interfaces"
	"github.com/liamawhite/navigator/edge/pkg/kubernetes"
	"github.com/liamawhite/navigator/edge/pkg/metrics"
	"github.com/liamawhite/navigator/edge/pkg/metrics/prometheus"
	"github.com/liamawhite/navigator/edge/pkg/proxy"
	edgeService "github.com/liamawhite/navigator/edge/pkg/service"
	managerConfig "github.com/liamawhite/navigator/manager/pkg/config"
	"github.com/liamawhite/navigator/manager/pkg/connections"
	managerServer "github.com/liamawhite/navigator/manager/pkg/server"
	navctlConfig "github.com/liamawhite/navigator/navctl/pkg/config"
	"github.com/liamawhite/navigator/navctl/pkg/ui"
	"github.com/liamawhite/navigator/pkg/istio/proxy/client"
	"github.com/liamawhite/navigator/pkg/logging"
)

var (
	// Config file flag
	configFile string
	// Demo mode flag
	demoMode bool

	// Traditional CLI flags (used when no config file is specified)
	kubeconfig     string
	contexts       []string
	managerPort    int
	managerHost    string
	maxMessageSize int
	disableUI      bool
	uiPort         int
	noBrowser      bool
	// Metrics flags (enabled is inferred from presence of endpoint)
	metricsType       string
	metricsEndpoint   string
	metricsTimeout    int
	metricsAuthBearer string
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Run manager and edge services locally",
	Long:  "", // Will be set in init()
	RunE:  runLocal,
}

// LocalRuntime holds the configuration and services needed to run Navigator locally
type LocalRuntime struct {
	Logger        *slog.Logger
	ManagerConfig *managerConfig.Config
	UIConfig      *UIConfig
	EdgeConfigs   []EdgeRuntimeConfig
}

// EdgeRuntimeConfig holds configuration for a single edge service
type EdgeRuntimeConfig struct {
	KubeconfigPath string
	ContextName    string
	EdgeConfig     *edgeConfig.Config
}

// UIConfig holds UI server configuration
type UIConfig struct {
	Port      int
	Disabled  bool
	NoBrowser bool
}

func runLocal(cmd *cobra.Command, args []string) error {
	logger := logging.For("navctl-local")

	// Validate that conflicting flags aren't used together
	if demoMode && configFile != "" {
		return fmt.Errorf("cannot use --demo and --config flags together")
	}

	// Prepare runtime configuration based on mode
	var runtime *LocalRuntime
	var err error

	if demoMode || configFile != "" {
		runtime, err = prepareConfigFileRuntime(logger, logLevel, logFormat)
	} else {
		runtime, err = prepareCLIRuntime(logger, logLevel, logFormat)
	}

	if err != nil {
		return err
	}

	// Run Navigator services with the prepared configuration
	return runNavigatorServices(runtime)
}

// prepareConfigFileRuntime prepares LocalRuntime from configuration file
func prepareConfigFileRuntime(logger *slog.Logger, globalLogLevel, globalLogFormat string) (*LocalRuntime, error) {
	var configManager *navctlConfig.Manager
	var err error

	// Load configuration based on mode
	if demoMode {
		// Load embedded demo configuration
		configManager, err = navctlConfig.LoadDemoConfig(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to load demo configuration: %w", err)
		}
		logger.Info("loaded embedded demo configuration")
	} else {
		// Load configuration from file
		configManager, err = navctlConfig.NewManager(configFile, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	config := configManager.GetConfig()

	// Validate configuration
	if err := configManager.ValidateEdges(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	logger.Info("loaded Navigator configuration",
		"config_file", configFile,
		"edge_count", len(config.Edges),
		"manager_host", config.Manager.Host,
		"manager_port", config.Manager.Port)

	// Prepare manager configuration
	managerCfg := configManager.GetManagerConfig()
	// Override with global CLI flags if provided
	if globalLogLevel != "" {
		managerCfg.LogLevel = globalLogLevel
	}
	if globalLogFormat != "" {
		managerCfg.LogFormat = globalLogFormat
	}

	// Prepare UI configuration
	uiConfig := configManager.GetUIConfig()

	// Prepare edge configurations
	var edgeConfigs []EdgeRuntimeConfig
	edgeCount := configManager.GetEdgeCount()
	for i := 0; i < edgeCount; i++ {
		edgeCfg, err := configManager.GetEdgeConfig(i, globalLogLevel, globalLogFormat)
		if err != nil {
			logger.Error("failed to get edge config", "edge_index", i, "error", err)
			continue
		}

		kubeconfigPath, err := configManager.GetEdgeKubeconfig(i)
		if err != nil {
			logger.Error("failed to get kubeconfig path", "edge_index", i, "error", err)
			continue
		}

		// Use default kubeconfig if not specified
		if kubeconfigPath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
		}

		contextName, err := configManager.GetEdgeKubeContext(i)
		if err != nil {
			logger.Error("failed to get kube context", "edge_index", i, "error", err)
			continue
		}

		edgeConfigs = append(edgeConfigs, EdgeRuntimeConfig{
			KubeconfigPath: kubeconfigPath,
			ContextName:    contextName,
			EdgeConfig:     edgeCfg,
		})
	}

	if len(edgeConfigs) == 0 {
		return nil, fmt.Errorf("no valid edge configurations found")
	}

	return &LocalRuntime{
		Logger:        logger,
		ManagerConfig: managerCfg,
		UIConfig: &UIConfig{
			Port:      uiConfig.Port,
			Disabled:  uiConfig.Disabled,
			NoBrowser: uiConfig.NoBrowser,
		},
		EdgeConfigs: edgeConfigs,
	}, nil
}

// prepareCLIRuntime prepares LocalRuntime from CLI flags
func prepareCLIRuntime(logger *slog.Logger, globalLogLevel, globalLogFormat string) (*LocalRuntime, error) {
	// Validate kubeconfig exists
	if err := validateKubeconfig(); err != nil {
		return nil, fmt.Errorf("kubeconfig validation failed: %w", err)
	}

	// Validate contexts
	if err := validateContexts(logger); err != nil {
		return nil, fmt.Errorf("context validation failed: %w", err)
	}

	// Get contexts to use
	contextsToUse, err := getContextsToUse(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to determine contexts: %w", err)
	}

	logger.Info("loaded Navigator CLI configuration",
		"kubeconfig", kubeconfig,
		"contexts", contextsToUse,
		"manager_port", managerPort,
		"manager_host", managerHost)

	// Prepare manager configuration
	managerCfg := &managerConfig.Config{
		Port:           managerPort,
		MaxMessageSize: maxMessageSize,
		LogLevel:       globalLogLevel,
		LogFormat:      globalLogFormat,
	}

	// Prepare edge configurations for each context
	var edgeConfigs []EdgeRuntimeConfig
	for _, contextName := range contextsToUse {
		edgeConfig := &edgeConfig.Config{
			ManagerEndpoint: fmt.Sprintf("%s:%d", managerHost, managerPort),
			SyncInterval:    30,
			LogLevel:        globalLogLevel,
			LogFormat:       globalLogFormat,
			MetricsConfig: metrics.Config{
				Enabled: metricsEndpoint != "",
			},
		}

		// Add metrics configuration if endpoint provided
		if metricsEndpoint != "" {
			edgeConfig.MetricsConfig.Type = "prometheus"
			edgeConfig.MetricsConfig.Endpoint = metricsEndpoint
			edgeConfig.MetricsConfig.QueryInterval = 30 // Default query interval
			edgeConfig.MetricsConfig.Timeout = 10       // Default timeout
		}

		edgeConfigs = append(edgeConfigs, EdgeRuntimeConfig{
			KubeconfigPath: kubeconfig,
			ContextName:    contextName,
			EdgeConfig:     edgeConfig,
		})
	}

	return &LocalRuntime{
		Logger:        logger,
		ManagerConfig: managerCfg,
		UIConfig: &UIConfig{
			Port:      uiPort,
			Disabled:  disableUI,
			NoBrowser: noBrowser,
		},
		EdgeConfigs: edgeConfigs,
	}, nil
}

// runNavigatorServices runs all Navigator services using the provided runtime configuration
func runNavigatorServices(runtime *LocalRuntime) error {
	logger := runtime.Logger

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start manager service
	managerSvc, err := startManagerServiceWithConfig(ctx, runtime.ManagerConfig, logger)
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

	// Start edge services
	var edgeServices []*edgeService.EdgeService
	for _, edgeConfig := range runtime.EdgeConfigs {
		logger.Info("starting edge service", "context", edgeConfig.ContextName)
		edgeSvc, err := startEdgeServiceFromRuntime(ctx, edgeConfig, logger)
		if err != nil {
			logger.Error("failed to start edge service", "context", edgeConfig.ContextName, "error", err)
			// Continue with other edges instead of failing completely
			continue
		}
		edgeServices = append(edgeServices, edgeSvc)
	}

	if len(edgeServices) == 0 {
		return fmt.Errorf("no edge services could be started")
	}

	// Setup cleanup for all edge services
	defer func() {
		logger.Info("stopping edge services", "count", len(edgeServices))
		for i, edgeSvc := range edgeServices {
			if err := edgeSvc.Stop(); err != nil {
				logger.Error("error stopping edge service", "service_index", i, "error", err)
			}
		}
	}()

	// Start UI server unless disabled
	var uiSvc *ui.Server
	if !runtime.UIConfig.Disabled {
		uiSvc, err = startUIServerFromRuntime(ctx, runtime.UIConfig, runtime.ManagerConfig.Port, logger)
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
	logger.Info("manager gRPC server listening", "port", runtime.ManagerConfig.Port)
	logger.Info("manager HTTP gateway listening", "port", runtime.ManagerConfig.Port+1)
	logger.Info("edge services running", "count", len(edgeServices))

	if !runtime.UIConfig.Disabled {
		logger.Info("UI server listening", "port", runtime.UIConfig.Port)
		if !runtime.UIConfig.NoBrowser {
			// Open browser after a short delay
			go func() {
				time.Sleep(1 * time.Second)
				url := fmt.Sprintf("http://localhost:%d", runtime.UIConfig.Port)
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

// startEdgeServiceFromRuntime starts an edge service using EdgeRuntimeConfig
func startEdgeServiceFromRuntime(ctx context.Context, edgeConfig EdgeRuntimeConfig, logger *slog.Logger) (*edgeService.EdgeService, error) {
	// Create Kubernetes client with specific context
	k8sLogger := logging.For(logging.ComponentServer).With("context", edgeConfig.ContextName, "component", "k8s")
	k8sClient, err := kubernetes.NewClientWithContext(edgeConfig.KubeconfigPath, edgeConfig.ContextName, k8sLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client for context '%s': %w", edgeConfig.ContextName, err)
	}

	// Auto-discover cluster name from Istio
	clusterName, err := k8sClient.GetClusterName(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to auto-discover cluster name from Istio control plane: %w", err)
	}
	
	logger.Info("discovered cluster name from Istio", "cluster_name", clusterName, "context", edgeConfig.ContextName)

	// Create admin client for proxy configuration access
	adminClient := client.NewAdminClient(k8sClient.GetClientset(), k8sClient.GetRestConfig())

	// Create proxy service
	proxyLogger := logging.For(logging.ComponentServer).With("cluster", clusterName, "component", "proxy")
	proxyService := proxy.NewProxyService(adminClient, proxyLogger)

	// Create metrics provider
	metricsLogger := logging.For(logging.ComponentServer).With("cluster", clusterName, "component", "metrics")
	var metricsProvider interfaces.MetricsProvider
	metricsConfig := edgeConfig.EdgeConfig.GetMetricsConfig()

	if metricsConfig.Enabled && metricsConfig.Type == metrics.ProviderTypePrometheus {
		metricsProvider, err = prometheus.Create(metricsConfig, metricsLogger, clusterName)
		if err != nil {
			return nil, fmt.Errorf("failed to create metrics provider for cluster '%s': %w", clusterName, err)
		}
	}

	// Create edge service
	edgeLogger := logging.For(logging.ComponentServer).With("cluster", clusterName, "component", "edge")
	edgeSvc, err := edgeService.NewEdgeService(edgeConfig.EdgeConfig, k8sClient, proxyService, metricsProvider, edgeLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create edge service for cluster '%s': %w", clusterName, err)
	}

	// Start edge service in goroutine
	go func() {
		if err := edgeSvc.Start(); err != nil {
			logger.Error("edge service error", "cluster", clusterName, "error", err)
		}
	}()

	return edgeSvc, nil
}

// startUIServerFromRuntime starts a UI server using UIConfig
func startUIServerFromRuntime(ctx context.Context, uiConfig *UIConfig, managerPort int, logger *slog.Logger) (*ui.Server, error) {
	// Create UI server
	uiSvc, err := ui.NewServer(uiConfig.Port, managerPort+1) // HTTP gateway port
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

// validateContexts validates that all specified contexts exist in the kubeconfig
// Expands patterns and validates that all resolved contexts exist
func validateContexts(logger *slog.Logger) error {
	if len(contexts) == 0 {
		return nil // No contexts specified, will use current context
	}

	// Expand patterns to actual context names
	expandedContexts, err := expandContextPatterns(contexts, logger)
	if err != nil {
		return fmt.Errorf("failed to expand context patterns: %w", err)
	}

	if len(expandedContexts) == 0 {
		return fmt.Errorf("no contexts matched the specified patterns: %v", contexts)
	}

	// Load the kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Validate each resolved context exists
	for _, contextName := range expandedContexts {
		if _, exists := config.Contexts[contextName]; !exists {
			return fmt.Errorf("context '%s' not found in kubeconfig", contextName)
		}
	}

	return nil
}

// getContextsToUse returns the list of contexts to use, defaulting to current context if none specified
// Expands patterns if any are provided
func getContextsToUse(logger *slog.Logger) ([]string, error) {
	if len(contexts) > 0 {
		// Expand patterns to actual context names
		expandedContexts, err := expandContextPatterns(contexts, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to expand context patterns: %w", err)
		}
		return expandedContexts, nil
	}

	// Load the kubeconfig to get current context
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if config.CurrentContext == "" {
		return nil, fmt.Errorf("no current context set in kubeconfig and no contexts specified")
	}

	return []string{config.CurrentContext}, nil
}

// getAvailableContexts returns all available contexts from kubeconfig
func getAvailableContexts(kubeconfigPath string) ([]string, string, error) {
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, "", err
	}

	var contextNames []string
	for name := range config.Contexts {
		contextNames = append(contextNames, name)
	}

	// Sort for consistent output
	sort.Strings(contextNames)

	return contextNames, config.CurrentContext, nil
}

// isPattern returns true if the string contains glob pattern characters
func isPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// expandContextPatterns expands glob patterns in context names to actual context names
func expandContextPatterns(patterns []string, logger *slog.Logger) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	// Load available contexts
	availableContexts, _, err := getAvailableContexts(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load contexts from kubeconfig: %w", err)
	}

	var expandedContexts []string
	var hasPatterns bool

	for _, pattern := range patterns {
		if !isPattern(pattern) {
			// Exact match - add as-is
			expandedContexts = append(expandedContexts, pattern)
			continue
		}

		hasPatterns = true
		var matches []string

		// Find all contexts that match this pattern
		for _, contextName := range availableContexts {
			matched, err := filepath.Match(pattern, contextName)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
			}
			if matched {
				matches = append(matches, contextName)
			}
		}

		if len(matches) == 0 {
			logger.Warn("pattern matched no contexts", "pattern", pattern)
		} else if logger.Enabled(context.Background(), slog.LevelDebug) {
			logger.Debug("pattern expanded", "pattern", pattern, "matches", matches, "count", len(matches))
		}

		expandedContexts = append(expandedContexts, matches...)
	}

	// Remove duplicates and sort
	expandedContexts = removeDuplicates(expandedContexts)
	sort.Strings(expandedContexts)

	// Log summary if patterns were used
	if hasPatterns && logger.Enabled(context.Background(), slog.LevelInfo) {
		logger.Info("expanded context patterns", "input_patterns", patterns, "resolved_contexts", expandedContexts, "count", len(expandedContexts))
	}

	return expandedContexts, nil
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// extractClusterIDFromKubeconfig extracts the cluster name from the specified context in kubeconfig
// If contextName is empty, uses the current context

// generateHelpText creates dynamic help text with available contexts
func generateHelpText(kubeconfigPath string) string {
	baseHelp := `Run both Navigator manager and edge services locally.
This command starts the manager service first, then connects edge services to it.
By default, it uses the current context from your kubeconfig. You can specify
multiple contexts using the --contexts flag to monitor multiple clusters simultaneously.

The --contexts flag supports both exact context names and glob patterns:
  * - matches any sequence of characters
  ? - matches any single character
  [abc] - matches any character in brackets
  [a-z] - matches any character in range

Examples:
  # Use current context
  navctl local

  # Use specific contexts
  navctl local --contexts context1,context2,context3

  # Use glob patterns to select multiple contexts
  navctl local --contexts "*-prod"
  navctl local --contexts "team-*"
  navctl local --contexts "*-prod,*-staging"
  navctl local --contexts "dev-*,test-?"

  # Mix exact names and patterns
  navctl local --contexts "production,*-staging"

  # Use custom kubeconfig with patterns
  navctl local --kube-config ~/.kube/config --contexts "*-prod"`

	// Try to get available contexts
	availableContexts, currentContext, err := getAvailableContexts(kubeconfigPath)
	if err != nil {
		return baseHelp + "\n\n(Note: Could not read kubeconfig to show available contexts)"
	}

	if len(availableContexts) > 0 {
		contextInfo := "\n\nAvailable contexts in " + kubeconfigPath + ":"
		for _, ctx := range availableContexts {
			if ctx == currentContext {
				contextInfo += "\n  * " + ctx + " (current)"
			} else {
				contextInfo += "\n  - " + ctx
			}
		}
		return baseHelp + contextInfo
	}

	return baseHelp
}

// startManagerServiceWithConfig starts the manager service using configuration
func startManagerServiceWithConfig(ctx context.Context, cfg *managerConfig.Config, logger *slog.Logger) (*managerServer.ManagerServer, error) {
	// Create connections manager
	connectionManager := connections.NewManager(logging.For("manager"))

	// Create manager server
	managerSvc, err := managerServer.NewManagerServer(cfg, connectionManager, logging.For("manager"))
	if err != nil {
		return nil, fmt.Errorf("failed to create manager server: %w", err)
	}

	// Start manager server in goroutine
	go func() {
		if err := managerSvc.Start(); err != nil {
			logger.Error("manager server error", "error", err)
		}
	}()

	return managerSvc, nil
}

// startEdgeServiceFromConfig starts an edge service using configuration

// startUIServerWithConfig starts the UI server using configuration

func init() {
	// Default kubeconfig path
	defaultKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Set initial help text with available contexts
	localCmd.Long = generateHelpText(defaultKubeconfig)

	// Command flags
	localCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to navctl configuration file (YAML or JSON)")
	localCmd.Flags().BoolVar(&demoMode, "demo", false, "Use embedded demo configuration for navigator-demo clusters")
	localCmd.Flags().StringVarP(&kubeconfig, "kube-config", "k", defaultKubeconfig, "Path to kubeconfig file (CLI mode only)")
	localCmd.Flags().StringSliceVar(&contexts, "contexts", nil, "Comma-separated list of kubeconfig contexts to use (CLI mode only)")
	localCmd.Flags().IntVar(&managerPort, "manager-port", 8080, "Port for manager service (CLI mode only)")
	localCmd.Flags().StringVar(&managerHost, "manager-host", "localhost", "Host for manager service (CLI mode only)")
	localCmd.Flags().IntVar(&maxMessageSize, "max-message-size", 10, "Maximum gRPC message size in MB (CLI mode only)")
	localCmd.Flags().BoolVar(&disableUI, "disable-ui", false, "Disable UI server (CLI mode only)")
	localCmd.Flags().IntVar(&uiPort, "ui-port", 8082, "Port for UI server (CLI mode only)")
	localCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser automatically (CLI mode only)")

	// Metrics flags (CLI mode only)
	localCmd.Flags().StringVar(&metricsType, "metrics-type", "prometheus", "Metrics provider type (CLI mode only)")
	localCmd.Flags().StringVar(&metricsEndpoint, "metrics-endpoint", "", "Metrics provider endpoint (CLI mode only)")
	localCmd.Flags().IntVar(&metricsTimeout, "metrics-timeout", 10, "Metrics query timeout in seconds (CLI mode only)")
	localCmd.Flags().StringVar(&metricsAuthBearer, "metrics-auth-bearer", "", "Bearer token for metrics provider authentication (CLI mode only)")

	// kube-config is optional with default value
}
