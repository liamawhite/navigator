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
	"log/slog"

	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/liamawhite/navigator/pkg/version"
	"github.com/spf13/cobra"
)

var (
	logLevel  string
	logFormat string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "navctl",
	Short: "Navigator control plane CLI",
	Long: `navctl is a CLI tool for managing Navigator services locally.
It provides commands for running and coordinating Navigator's manager and edge services.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize global logger
		config := &logging.Config{
			Level:  logging.ParseLevel(logLevel),
			Format: logFormat,
		}
		logger := logging.NewLogger(config)
		slog.SetDefault(logger)

		// Log startup information
		logging.For("navctl").Debug("navctl starting",
			"version", version.Get(),
			"log_level", logLevel,
			"log_format", logFormat)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format (text, json)")

	// Add subcommands
	rootCmd.AddCommand(localCmd)
	rootCmd.AddCommand(versionCmd)
}
