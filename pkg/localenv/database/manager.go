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

package database

import (
	"embed"
	"fmt"
	"log/slog"
)

//go:embed manifests/*
var manifestsFS embed.FS

// KustomizeManager manages Kustomize operations for database installation
type KustomizeManager struct {
	kubeconfig string
	logger     *slog.Logger
}

// NewKustomizeManager creates a new Kustomize manager instance for database
func NewKustomizeManager(kubeconfig string, logger *slog.Logger) (*KustomizeManager, error) {
	if logger == nil {
		logger = slog.Default()
	}

	k := &KustomizeManager{
		kubeconfig: kubeconfig,
		logger:     logger,
	}

	// Check if kubectl is available
	if err := k.checkKubectl(); err != nil {
		return nil, fmt.Errorf("kubectl not available: %w", err)
	}

	return k, nil
}
