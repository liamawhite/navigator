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
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TokenExecutor handles executing commands to retrieve bearer tokens
type TokenExecutor struct {
	cache  map[string]*TokenCache
	mutex  sync.RWMutex
	logger *slog.Logger
}

// NewTokenExecutor creates a new TokenExecutor
func NewTokenExecutor(logger *slog.Logger) *TokenExecutor {
	return &TokenExecutor{
		cache:  make(map[string]*TokenCache),
		logger: logger,
	}
}

// GetBearerToken retrieves a bearer token using the provided auth configuration
func (te *TokenExecutor) GetBearerToken(edgeName string, auth *MetricsAuth) (string, error) {
	if auth == nil {
		return "", nil
	}

	// Use static token if provided
	if auth.BearerToken != "" {
		return auth.BearerToken, nil
	}

	// Use exec command if provided
	if auth.BearerTokenExec != nil {
		return te.getBearerTokenFromExec(edgeName, auth.BearerTokenExec)
	}

	return "", nil
}

// getBearerTokenFromExec executes a command to retrieve a bearer token
func (te *TokenExecutor) getBearerTokenFromExec(edgeName string, execConfig *ExecConfig) (string, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("%s:%s:%v", edgeName, execConfig.Command, execConfig.Args)

	// Check cache first
	te.mutex.RLock()
	if cached, exists := te.cache[cacheKey]; exists && !cached.IsExpired() {
		te.mutex.RUnlock()
		te.logger.Debug("using cached bearer token", "edge", edgeName)
		return cached.Token, nil
	}
	te.mutex.RUnlock()

	te.logger.Debug("executing command to get bearer token", "edge", edgeName, "command", execConfig.Command)

	// Execute command
	token, err := te.executeCommand(execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to execute bearer token command for edge %s: %w", edgeName, err)
	}

	// Cache the token with a default TTL
	te.mutex.Lock()
	te.cache[cacheKey] = &TokenCache{
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute), // Default 15 minute cache
	}
	te.mutex.Unlock()

	te.logger.Debug("successfully retrieved bearer token", "edge", edgeName)
	return token, nil
}

// executeCommand executes the specified command and returns the output as a token
func (te *TokenExecutor) executeCommand(execConfig *ExecConfig) (string, error) {
	// Parse timeout
	timeout := 30 * time.Second // Default timeout
	if execConfig.Timeout != "" {
		var err error
		timeout, err = time.ParseDuration(execConfig.Timeout)
		if err != nil {
			return "", fmt.Errorf("invalid timeout format %s: %w", execConfig.Timeout, err)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create command - Note: This executes user-provided commands for token generation
	// This is the intended behavior for Kubernetes-style exec authentication
	cmd := exec.CommandContext(ctx, execConfig.Command, execConfig.Args...) // #nosec G204

	// Set environment variables
	cmd.Env = os.Environ() // Start with current environment
	for _, envVar := range execConfig.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
	}

	// Execute command
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", timeout)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("command failed with exit code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	// Clean up the output (remove trailing newlines/whitespace)
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("command returned empty output")
	}

	return token, nil
}

// ClearCache clears all cached tokens
func (te *TokenExecutor) ClearCache() {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	te.cache = make(map[string]*TokenCache)
}

// ClearExpiredTokens removes expired tokens from the cache
func (te *TokenExecutor) ClearExpiredTokens() {
	te.mutex.Lock()
	defer te.mutex.Unlock()

	for key, cached := range te.cache {
		if cached.IsExpired() {
			delete(te.cache, key)
		}
	}
}

// RefreshToken forces a refresh of the token for a specific edge
func (te *TokenExecutor) RefreshToken(edgeName string, auth *MetricsAuth) (string, error) {
	if auth == nil || auth.BearerTokenExec == nil {
		return te.GetBearerToken(edgeName, auth)
	}

	// Clear cached token for this edge
	cacheKey := fmt.Sprintf("%s:%s:%v", edgeName, auth.BearerTokenExec.Command, auth.BearerTokenExec.Args)
	te.mutex.Lock()
	delete(te.cache, cacheKey)
	te.mutex.Unlock()

	// Get fresh token
	return te.GetBearerToken(edgeName, auth)
}

// GetCacheStats returns statistics about the token cache
func (te *TokenExecutor) GetCacheStats() map[string]any {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	total := len(te.cache)
	expired := 0
	for _, cached := range te.cache {
		if cached.IsExpired() {
			expired++
		}
	}

	return map[string]any{
		"total_tokens":   total,
		"expired_tokens": expired,
		"active_tokens":  total - expired,
	}
}
