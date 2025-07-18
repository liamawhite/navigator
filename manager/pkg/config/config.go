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
	"flag"
	"fmt"
)

// Config holds the configuration for the manager service
type Config struct {
	Port           int
	LogLevel       string
	LogFormat      string
	MaxMessageSize int // Maximum gRPC message size in MB
}

// ParseFlags parses command line flags and returns a Config
func ParseFlags() (*Config, error) {
	config := &Config{}

	flag.IntVar(&config.Port, "port", 8080, "Port for the gRPC server")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&config.LogFormat, "log-format", "text", "Log format (text, json)")
	flag.IntVar(&config.MaxMessageSize, "max-message-size", 10, "Maximum gRPC message size in MB")

	flag.Parse()

	return config, config.Validate()
}

// Validate checks that required configuration is provided
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return fmt.Errorf("log-level must be one of: debug, info, warn, error")
	}

	if c.LogFormat != "text" && c.LogFormat != "json" {
		return fmt.Errorf("log-format must be one of: text, json")
	}

	if c.MaxMessageSize <= 0 {
		return fmt.Errorf("max-message-size must be greater than 0")
	}

	return nil
}

// GetPort returns the port for the gRPC server
func (c *Config) GetPort() int {
	return c.Port
}

// GetMaxMessageSize returns the maximum gRPC message size in bytes
func (c *Config) GetMaxMessageSize() int {
	return c.MaxMessageSize * 1024 * 1024 // Convert MB to bytes
}
