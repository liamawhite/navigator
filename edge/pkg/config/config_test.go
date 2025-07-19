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

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				LogLevel:        "info",
				LogFormat:       "text",
				MaxMessageSize:  10,
			},
			wantErr: false,
		},
		{
			name: "missing cluster ID",
			config: Config{
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				LogLevel:        "info",
				LogFormat:       "text",
				MaxMessageSize:  10,
			},
			wantErr: true,
			errMsg:  "cluster-id is required",
		},
		{
			name: "missing manager endpoint",
			config: Config{
				ClusterID:      "test-cluster",
				SyncInterval:   30,
				LogLevel:       "info",
				LogFormat:      "text",
				MaxMessageSize: 10,
			},
			wantErr: true,
			errMsg:  "manager-endpoint is required",
		},
		{
			name: "invalid sync interval",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    0,
				LogLevel:        "info",
				LogFormat:       "text",
				MaxMessageSize:  10,
			},
			wantErr: true,
			errMsg:  "sync-interval must be positive",
		},
		{
			name: "negative sync interval",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    -1,
				LogLevel:        "info",
				LogFormat:       "text",
				MaxMessageSize:  10,
			},
			wantErr: true,
			errMsg:  "sync-interval must be positive",
		},
		{
			name: "invalid log level",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				LogLevel:        "invalid",
				LogFormat:       "text",
				MaxMessageSize:  10,
			},
			wantErr: true,
			errMsg:  "log-level must be one of: debug, info, warn, error",
		},
		{
			name: "invalid log format",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				LogLevel:        "info",
				LogFormat:       "invalid",
				MaxMessageSize:  10,
			},
			wantErr: true,
			errMsg:  "log-format must be one of: text, json",
		},
		{
			name: "invalid max message size",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				LogLevel:        "info",
				LogFormat:       "text",
				MaxMessageSize:  0,
			},
			wantErr: true,
			errMsg:  "max-message-size must be greater than 0",
		},
		{
			name: "valid debug log level",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				LogLevel:        "debug",
				LogFormat:       "text",
				MaxMessageSize:  10,
			},
			wantErr: false,
		},
		{
			name: "valid json log format",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				LogLevel:        "info",
				LogFormat:       "json",
				MaxMessageSize:  10,
			},
			wantErr: false,
		},
		{
			name: "valid with kubeconfig path",
			config: Config{
				ClusterID:       "test-cluster",
				ManagerEndpoint: "localhost:8080",
				SyncInterval:    30,
				KubeconfigPath:  "/path/to/kubeconfig",
				LogLevel:        "info",
				LogFormat:       "text",
				MaxMessageSize:  10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err, "Config.Validate() expected error but got none")
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
