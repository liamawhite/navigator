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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/liamawhite/navigator/edge/pkg/config"
	"github.com/liamawhite/navigator/edge/pkg/interfaces"
	"github.com/liamawhite/navigator/edge/pkg/kubernetes"
	"github.com/liamawhite/navigator/edge/pkg/metrics"
	"github.com/liamawhite/navigator/edge/pkg/metrics/prometheus"
	"github.com/liamawhite/navigator/edge/pkg/proxy"
	"github.com/liamawhite/navigator/edge/pkg/service"
	"github.com/liamawhite/navigator/pkg/istio/proxy/client"
	"github.com/liamawhite/navigator/pkg/logging"
)

func main() {
	// Parse configuration
	cfg, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Setup logging
	logger := logging.For("edge")

	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient(cfg.KubeconfigPath, logger)
	if err != nil {
		logger.Error("failed to create kubernetes client", "error", err)
		os.Exit(1)
	}

	// Create admin client for Envoy proxy access
	adminClient := client.NewAdminClient(k8sClient.GetClientset(), k8sClient.GetRestConfig())

	// Create proxy service for handling proxy configuration requests
	proxyService := proxy.NewProxyService(adminClient, logger)

	// Create metrics provider directly
	var metricsProvider interfaces.MetricsProvider
	metricsConfig := cfg.GetMetricsConfig()

	if metricsConfig.Enabled && metricsConfig.Type == metrics.ProviderTypePrometheus {
		// Get cluster name from Istio for metrics filtering
		clusterName, err := k8sClient.GetClusterName(context.Background())
		if err != nil {
			logger.Warn("failed to get cluster name from istiod, metrics will not be cluster-filtered", "error", err)
			clusterName = ""
		} else {
			logger.Info("retrieved cluster name for metrics filtering", "cluster_name", clusterName)
		}

		metricsProvider, err = prometheus.Create(metricsConfig, logger, clusterName)
		if err != nil {
			logger.Error("failed to create metrics provider", "error", err)
			os.Exit(1)
		}
	}

	// Create edge service
	edgeService, err := service.NewEdgeService(cfg, k8sClient, proxyService, metricsProvider, logger)
	if err != nil {
		logger.Error("failed to create edge service", "error", err)
		os.Exit(1)
	}

	// Start edge service
	if err := edgeService.Start(); err != nil {
		logger.Error("failed to start edge service", "error", err)
		os.Exit(1)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		logger.Info("context canceled")
	case sig := <-sigChan:
		logger.Info("received signal", "signal", sig)
		cancel()
	}

	// Graceful shutdown
	logger.Info("shutting down edge service")
	if err := edgeService.Stop(); err != nil {
		logger.Error("error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("edge service stopped")
}
