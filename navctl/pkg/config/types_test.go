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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenCache_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: time.Now().Add(5 * time.Minute),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-5 * time.Minute),
			want:      true,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TokenCache{
				Token:     "test-token",
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.want, tc.IsExpired())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, "navigator.io/v1alpha1", config.APIVersion)
	assert.Equal(t, "NavctlConfig", config.Kind)
	assert.Equal(t, "localhost", config.Manager.Host)
	assert.Equal(t, 8080, config.Manager.Port)
	assert.Equal(t, 10, config.Manager.MaxMessageSize)
	assert.Equal(t, 8082, config.UI.Port)
	assert.False(t, config.UI.Disabled)
	assert.False(t, config.UI.NoBrowser)
}

func TestDefaultEdgeConfig(t *testing.T) {
	edge := DefaultEdgeConfig()
	
	assert.Equal(t, 30, edge.SyncInterval)
	assert.Equal(t, "info", edge.LogLevel)
	assert.Equal(t, "text", edge.LogFormat)
}

func TestDefaultMetricsConfig(t *testing.T) {
	metrics := DefaultMetricsConfig()
	
	assert.Equal(t, "prometheus", metrics.Type)
	assert.Equal(t, 30, metrics.QueryInterval)
	assert.Equal(t, 10, metrics.Timeout)
}