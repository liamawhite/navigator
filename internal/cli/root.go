package cli

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/liamawhite/navigator/pkg/logging"
)

var (
	logLevel  string
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "navigator",
	Short: "Navigator is a Kubernetes service registry",
	Long: `Navigator provides a gRPC-based service registry for Kubernetes clusters.
It enables service discovery and provides Envoy configuration analysis capabilities.`,
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
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format (text, json)")

	rootCmd.AddCommand(serveCmd)
}
