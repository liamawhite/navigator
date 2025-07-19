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

	"github.com/liamawhite/navigator/manager/pkg/config"
	"github.com/liamawhite/navigator/manager/pkg/connections"
	"github.com/liamawhite/navigator/manager/pkg/service"
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
	logger := logging.For("manager")

	// Create connections manager
	connectionManager := connections.NewManager(logger)

	// Create manager service
	managerService, err := service.NewManagerService(cfg, connectionManager, logger)
	if err != nil {
		logger.Error("failed to create manager service", "error", err)
		os.Exit(1)
	}

	// Start manager service
	if err := managerService.Start(); err != nil {
		logger.Error("failed to start manager service", "error", err)
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
	logger.Info("shutting down manager service")
	if err := managerService.Stop(); err != nil {
		logger.Error("error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("manager service stopped")
}
