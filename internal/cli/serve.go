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
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"github.com/liamawhite/navigator/internal/grpc"
	"github.com/liamawhite/navigator/pkg/datastore/kubeconfig"
	"github.com/liamawhite/navigator/pkg/logging"
	troubleshootingKubeconfig "github.com/liamawhite/navigator/pkg/troubleshooting/kubeconfig"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Navigator gRPC server",
	Long: `Start the Navigator gRPC server that provides APIs for service discovery 
and Envoy configuration analysis.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		kubeconfigPath, _ := cmd.Flags().GetString("kubeconfig")

		logger := logging.For(logging.ComponentCLI)
		logger.Info("starting navigator server",
			"port", port,
			"kubeconfig", kubeconfigPath)

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

		// Create gRPC server
		server, err := grpc.NewServer(serviceDS, troubleshootingDS, port)
		if err != nil {
			logger.Error("failed to create gRPC server", "error", err)
			return fmt.Errorf("failed to create gRPC server: %w", err)
		}

		// Start server in a goroutine
		serverErrChan := make(chan error, 1)
		go func() {
			serverErrChan <- server.Start()
		}()

		// Wait for interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		logger.Info("navigator server started",
			"grpc_address", server.Address(),
			"http_address", server.HTTPAddress())

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
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")

	// Default kubeconfig path
	defaultKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeconfig = filepath.Join(home, ".kube", "config")
	}
	serveCmd.Flags().StringP("kubeconfig", "k", defaultKubeconfig, "Path to kubeconfig file")
}
