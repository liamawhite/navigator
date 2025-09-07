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

	"github.com/liamawhite/navigator/edge/pkg/config"
	"github.com/liamawhite/navigator/edge/pkg/kubernetes"
	"github.com/liamawhite/navigator/edge/pkg/metrics"
	"github.com/liamawhite/navigator/edge/pkg/metrics/prometheus"
	"github.com/liamawhite/navigator/edge/pkg/proxy"
	edgeService "github.com/liamawhite/navigator/edge/pkg/service"
	managerConfig "github.com/liamawhite/navigator/manager/pkg/config"
	"github.com/liamawhite/navigator/manager/pkg/connections"
	managerServer "github.com/liamawhite/navigator/manager/pkg/server"
	"github.com/liamawhite/navigator/navctl/pkg/ui"
	"github.com/liamawhite/navigator/pkg/istio/proxy/client"
	"github.com/liamawhite/navigator/pkg/logging"
)

var (
	kubeconfig     string
	contexts       []string
	managerPort    int
	managerHost    string
	maxMessageSize int
	disableUI      bool
	uiPort         int
	noBrowser      bool
	// Metrics flags (enabled is inferred from presence of endpoint)
	metricsType     string
	metricsEndpoint string
	metricsTimeout  int
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Run manager and edge services locally",
	Long:  "", // Will be set in init()
	RunE:  runLocal,
}

func runLocal(cmd *cobra.Command, args []string) error {
	logger := logging.For("navctl-local")

	// Validate kubeconfig exists
	if err := validateKubeconfig(); err != nil {
		return fmt.Errorf("kubeconfig validation failed: %w", err)
	}

	// Validate contexts
	if err := validateContexts(logger); err != nil {
		return fmt.Errorf("context validation failed: %w", err)
	}

	// Get contexts to use
	contextsToUse, err := getContextsToUse(logger)
	if err != nil {
		return fmt.Errorf("failed to determine contexts: %w", err)
	}

	logger.Info("starting Navigator services",
		"kubeconfig", kubeconfig,
		"contexts", contextsToUse,
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

	// Start edge services for each context
	var edgeServices []*edgeService.EdgeService
	for _, contextName := range contextsToUse {
		logger.Info("starting edge service for context", "context", contextName)
		edgeSvc, err := startEdgeServiceForContext(ctx, contextName, logger)
		if err != nil {
			return fmt.Errorf("failed to start edge service for context '%s': %w", contextName, err)
		}
		edgeServices = append(edgeServices, edgeSvc)
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
	logger.Info("edge services running", "contexts", contextsToUse, "count", len(edgeServices))

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

func startManagerService(ctx context.Context, logger *slog.Logger) (*managerServer.ManagerServer, error) {
	// Create manager config
	cfg := &managerConfig.Config{
		Port:           managerPort,
		LogLevel:       logLevel,
		LogFormat:      logFormat,
		MaxMessageSize: maxMessageSize,
	}

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

func startEdgeServiceForContext(ctx context.Context, contextName string, logger *slog.Logger) (*edgeService.EdgeService, error) {
	// Extract cluster ID from kubeconfig for the specific context
	clusterID, err := extractClusterIDFromKubeconfig(contextName)
	if err != nil {
		fallbackID := fmt.Sprintf("local-cluster-%s", contextName)
		if contextName == "" {
			fallbackID = "local-cluster"
		}
		logger.Warn("failed to extract cluster ID from kubeconfig, using fallback",
			"context", contextName, "error", err, "fallback_id", fallbackID)
		clusterID = fallbackID
	} else {
		logger.Info("extracted cluster ID from kubeconfig", "context", contextName, "cluster_id", clusterID)
	}

	// Create metrics configuration (enabled if endpoint provided)
	metricsConfig := metrics.Config{
		Enabled:       metricsEndpoint != "", // Infer enabled from presence of endpoint
		Type:          metrics.ProviderType(metricsType),
		Endpoint:      metricsEndpoint,
		QueryInterval: 30, // Default query interval
		Timeout:       metricsTimeout,
	}

	if metricsConfig.Enabled {
		logger.Info("metrics enabled", "context", contextName, "type", metricsType, "endpoint", metricsEndpoint)
	} else {
		logger.Info("metrics disabled", "context", contextName, "reason", "no endpoint provided")
	}

	// Create edge config with context-specific kubeconfig overrides
	cfg := &config.Config{
		KubeconfigPath:  kubeconfig,
		ManagerEndpoint: fmt.Sprintf("%s:%d", managerHost, managerPort),
		ClusterID:       clusterID,
		SyncInterval:    30,
		LogLevel:        logLevel,
		LogFormat:       logFormat,
		MaxMessageSize:  maxMessageSize,
		MetricsConfig:   metricsConfig,
	}

	// Create Kubernetes client with specific context
	k8sLogger := logging.For(logging.ComponentServer).With("context", contextName, "component", "k8s")
	k8sClient, err := kubernetes.NewClientWithContext(cfg.KubeconfigPath, contextName, k8sLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client for context '%s': %w", contextName, err)
	}

	// Create admin client for proxy configuration access
	adminClient := client.NewAdminClient(k8sClient.GetClientset(), k8sClient.GetRestConfig())

	// Create proxy service
	proxyLogger := logging.For(logging.ComponentServer).With("context", contextName, "component", "proxy")
	proxyService := proxy.NewProxyService(adminClient, proxyLogger)

	// Create metrics provider
	metricsLogger := logging.For(logging.ComponentServer).With("context", contextName, "component", "metrics")
	metricsRegistry := metrics.NewRegistry()
	prometheus.RegisterWithRegistry(metricsRegistry)

	metricsProvider, err := metricsRegistry.Create(cfg.GetMetricsConfig(), metricsLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics provider for context '%s': %w", contextName, err)
	}

	// Create edge service
	edgeLogger := logging.For(logging.ComponentServer).With("context", contextName, "component", "edge")
	edgeSvc, err := edgeService.NewEdgeService(cfg, k8sClient, proxyService, metricsProvider, edgeLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create edge service for context '%s': %w", contextName, err)
	}

	// Start edge service in goroutine
	go func() {
		if err := edgeSvc.Start(); err != nil {
			logger.Error("edge service error", "context", contextName, "error", err)
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
func extractClusterIDFromKubeconfig(contextName string) (string, error) {
	// Load the kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Determine which context to use
	targetContext := contextName
	if targetContext == "" {
		targetContext = config.CurrentContext
		if targetContext == "" {
			return "", fmt.Errorf("no current context set in kubeconfig")
		}
	}

	// Get the context details
	context, exists := config.Contexts[targetContext]
	if !exists {
		return "", fmt.Errorf("context '%s' not found in kubeconfig", targetContext)
	}

	// Get the cluster name
	clusterName := context.Cluster
	if clusterName == "" {
		return "", fmt.Errorf("no cluster set in context '%s'", targetContext)
	}

	return clusterName, nil
}

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

func init() {
	// Default kubeconfig path
	defaultKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Set initial help text with available contexts
	localCmd.Long = generateHelpText(defaultKubeconfig)

	// Command flags
	localCmd.Flags().StringVarP(&kubeconfig, "kube-config", "k", defaultKubeconfig, "Path to kubeconfig file")
	localCmd.Flags().StringSliceVar(&contexts, "contexts", nil, "Comma-separated list of kubeconfig contexts to use (uses current context if not specified)")
	localCmd.Flags().IntVar(&managerPort, "manager-port", 8080, "Port for manager service")
	localCmd.Flags().StringVar(&managerHost, "manager-host", "localhost", "Host for manager service")
	localCmd.Flags().IntVar(&maxMessageSize, "max-message-size", 10, "Maximum gRPC message size in MB")
	localCmd.Flags().BoolVar(&disableUI, "disable-ui", false, "Disable UI server")
	localCmd.Flags().IntVar(&uiPort, "ui-port", 8082, "Port for UI server")
	localCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser automatically")

	// Metrics flags
	localCmd.Flags().StringVar(&metricsType, "metrics-type", "prometheus", "Metrics provider type (prometheus)")
	localCmd.Flags().StringVar(&metricsEndpoint, "metrics-endpoint", "", "Metrics provider endpoint accessible from this machine (e.g., http://prometheus:9090). Enables metrics if provided.")
	localCmd.Flags().IntVar(&metricsTimeout, "metrics-timeout", 10, "Metrics query timeout in seconds")

	// kube-config is optional with default value
}
