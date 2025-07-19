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
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:           8080,
				LogLevel:       "info",
				LogFormat:      "text",
				MaxMessageSize: 10,
			},
			wantError: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Port:           0,
				LogLevel:       "info",
				LogFormat:      "text",
				MaxMessageSize: 10,
			},
			wantError: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Port:           65536,
				LogLevel:       "info",
				LogFormat:      "text",
				MaxMessageSize: 10,
			},
			wantError: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				Port:           8080,
				LogLevel:       "invalid",
				LogFormat:      "text",
				MaxMessageSize: 10,
			},
			wantError: true,
		},
		{
			name: "invalid log format",
			config: &Config{
				Port:           8080,
				LogLevel:       "info",
				LogFormat:      "invalid",
				MaxMessageSize: 10,
			},
			wantError: true,
		},
		{
			name: "invalid max message size",
			config: &Config{
				Port:           8080,
				LogLevel:       "info",
				LogFormat:      "text",
				MaxMessageSize: 0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Config.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestConfig_GetPort(t *testing.T) {
	config := &Config{Port: 9090}
	if got := config.GetPort(); got != 9090 {
		t.Errorf("Config.GetPort() = %v, want %v", got, 9090)
	}
}
