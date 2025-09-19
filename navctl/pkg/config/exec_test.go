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
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenExecutor_GetBearerToken_StaticToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerToken: "static-token",
	}

	token, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)
	assert.Equal(t, "static-token", token)
}

func TestTokenExecutor_GetBearerToken_NoAuth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	token, err := executor.GetBearerToken("test-edge", nil)
	require.NoError(t, err)
	assert.Empty(t, token)
}

func TestTokenExecutor_GetBearerToken_ExecCommand(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "echo",
			Args:    []string{"test-token"},
		},
	}

	token, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

func TestTokenExecutor_GetBearerToken_ExecCommand_WithTrimming(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "echo",
			Args:    []string{"  test-token  \n"},
		},
	}

	token, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)
	assert.Equal(t, "test-token", token)
}

func TestTokenExecutor_GetBearerToken_ExecCommand_WithTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "sleep",
			Args:    []string{"10"},
			Timeout: "1s",
		},
	}

	_, err := executor.GetBearerToken("test-edge", auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestTokenExecutor_GetBearerToken_ExecCommand_InvalidTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "echo",
			Args:    []string{"test-token"},
			Timeout: "invalid",
		},
	}

	_, err := executor.GetBearerToken("test-edge", auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid timeout format")
}

func TestTokenExecutor_GetBearerToken_ExecCommand_NonExistentCommand(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "non-existent-command",
		},
	}

	_, err := executor.GetBearerToken("test-edge", auth)
	assert.Error(t, err)
}

func TestTokenExecutor_GetBearerToken_ExecCommand_EmptyOutput(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "echo",
			Args:    []string{""}, // Empty output
		},
	}

	_, err := executor.GetBearerToken("test-edge", auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty output")
}

func TestTokenExecutor_GetBearerToken_ExecCommand_WithEnv(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "sh",
			Args:    []string{"-c", "echo $TEST_TOKEN"},
			Env: []EnvVar{
				{Name: "TEST_TOKEN", Value: "env-token"},
			},
		},
	}

	token, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)
	assert.Equal(t, "env-token", token)
}

func TestTokenExecutor_Caching(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "sh",
			Args:    []string{"-c", "echo cached-token-$(date +%s%N)"},
		},
	}

	// First call should execute the command
	token1, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(token1, "cached-token-"))

	// Second call should return cached token (same value)
	token2, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)
	assert.Equal(t, token1, token2)

	// Verify cache stats
	stats := executor.GetCacheStats()
	assert.Equal(t, 1, stats["total_tokens"])
	assert.Equal(t, 1, stats["active_tokens"])
	assert.Equal(t, 0, stats["expired_tokens"])
}

func TestTokenExecutor_RefreshToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "sh",
			Args:    []string{"-c", "echo refresh-token-$(date +%s%N)"},
		},
	}

	// Get initial token
	token1, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)

	// Sleep briefly to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Refresh should get a new token
	token2, err := executor.RefreshToken("test-edge", auth)
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2)
	assert.True(t, strings.HasPrefix(token2, "refresh-token-"))
}

func TestTokenExecutor_RefreshToken_StaticToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerToken: "static-token",
	}

	token, err := executor.RefreshToken("test-edge", auth)
	require.NoError(t, err)
	assert.Equal(t, "static-token", token)
}

func TestTokenExecutor_ClearCache(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	auth := &MetricsAuth{
		BearerTokenExec: &ExecConfig{
			Command: "echo",
			Args:    []string{"test-token"},
		},
	}

	// Get a token to populate cache
	_, err := executor.GetBearerToken("test-edge", auth)
	require.NoError(t, err)

	// Verify cache has content
	stats := executor.GetCacheStats()
	assert.Equal(t, 1, stats["total_tokens"])

	// Clear cache
	executor.ClearCache()

	// Verify cache is empty
	stats = executor.GetCacheStats()
	assert.Equal(t, 0, stats["total_tokens"])
}

func TestTokenExecutor_ClearExpiredTokens(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := NewTokenExecutor(logger)

	// Manually add an expired token to cache
	executor.cache["expired-key"] = &TokenCache{
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	// Add a valid token
	executor.cache["valid-key"] = &TokenCache{
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Verify cache has both tokens
	stats := executor.GetCacheStats()
	assert.Equal(t, 2, stats["total_tokens"])
	assert.Equal(t, 1, stats["expired_tokens"])
	assert.Equal(t, 1, stats["active_tokens"])

	// Clear expired tokens
	executor.ClearExpiredTokens()

	// Verify only valid token remains
	stats = executor.GetCacheStats()
	assert.Equal(t, 1, stats["total_tokens"])
	assert.Equal(t, 0, stats["expired_tokens"])
	assert.Equal(t, 1, stats["active_tokens"])
}