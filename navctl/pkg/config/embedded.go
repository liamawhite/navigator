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

package config

import (
	_ "embed"
	"fmt"
	"log/slog"
)

//go:embed demo-config.yaml
var demoConfigYAML string

// LoadDemoConfig loads the embedded demo configuration
func LoadDemoConfig(logger *slog.Logger) (*Manager, error) {
	// Parse the embedded demo config
	config, err := parseConfig([]byte(demoConfigYAML), "demo-config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded demo config: %w", err)
	}

	// Apply defaults and validate
	if err := applyDefaultsAndValidate(config); err != nil {
		return nil, fmt.Errorf("demo config validation failed: %w", err)
	}

	// Perform post-load processing
	config.PostLoad()

	return &Manager{
		config:        config,
		tokenExecutor: NewTokenExecutor(logger),
		logger:        logger,
	}, nil
}