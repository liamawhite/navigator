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
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

// cacheEntry represents an entry in the LRU cache
type cacheEntry struct {
	token      *TokenCache
	lastUsed   time.Time
	prev, next *cacheEntry
}

// lruCache implements a simple LRU cache with maximum size
type lruCache struct {
	maxSize int
	entries map[string]*cacheEntry
	head    *cacheEntry
	tail    *cacheEntry
}

// newLRUCache creates a new LRU cache with the specified maximum size
func newLRUCache(maxSize int) *lruCache {
	if maxSize <= 0 {
		maxSize = 100 // Default max size
	}

	head := &cacheEntry{}
	tail := &cacheEntry{}
	head.next = tail
	tail.prev = head

	return &lruCache{
		maxSize: maxSize,
		entries: make(map[string]*cacheEntry),
		head:    head,
		tail:    tail,
	}
}

// get retrieves a value from the cache and moves it to front
func (c *lruCache) get(key string) (*TokenCache, bool) {
	if entry, exists := c.entries[key]; exists {
		// Move to front
		c.moveToFront(entry)
		entry.lastUsed = time.Now()
		return entry.token, true
	}
	return nil, false
}

// put adds a value to the cache, evicting LRU entry if needed
func (c *lruCache) put(key string, token *TokenCache) {
	if entry, exists := c.entries[key]; exists {
		// Update existing entry
		entry.token = token
		entry.lastUsed = time.Now()
		c.moveToFront(entry)
		return
	}

	// Create new entry
	entry := &cacheEntry{
		token:    token,
		lastUsed: time.Now(),
	}

	// Add to front
	c.addToFront(entry)
	c.entries[key] = entry

	// Evict LRU entry if over capacity
	if len(c.entries) > c.maxSize {
		c.evictLRU()
	}
}

// remove removes an entry from the cache
func (c *lruCache) remove(key string) {
	if entry, exists := c.entries[key]; exists {
		c.removeEntry(entry)
		delete(c.entries, key)
	}
}

// removeExpired removes all expired entries
func (c *lruCache) removeExpired() int {
	var expiredKeys []string
	for key, entry := range c.entries {
		if entry.token.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		c.remove(key)
	}

	return len(expiredKeys)
}

// clear removes all entries
func (c *lruCache) clear() {
	c.entries = make(map[string]*cacheEntry)
	c.head.next = c.tail
	c.tail.prev = c.head
}

// size returns the current cache size
func (c *lruCache) size() int {
	return len(c.entries)
}

// moveToFront moves an entry to the front of the list
func (c *lruCache) moveToFront(entry *cacheEntry) {
	c.removeEntry(entry)
	c.addToFront(entry)
}

// addToFront adds an entry to the front of the list
func (c *lruCache) addToFront(entry *cacheEntry) {
	entry.prev = c.head
	entry.next = c.head.next
	c.head.next.prev = entry
	c.head.next = entry
}

// removeEntry removes an entry from the doubly linked list
func (c *lruCache) removeEntry(entry *cacheEntry) {
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
}

// evictLRU removes the least recently used entry
func (c *lruCache) evictLRU() {
	if c.tail.prev != c.head {
		lru := c.tail.prev
		c.removeEntry(lru)
		// Find and delete the corresponding map entry
		for key, entry := range c.entries {
			if entry == lru {
				delete(c.entries, key)
				break
			}
		}
	}
}

// TokenExecutor handles executing commands to retrieve bearer tokens
type TokenExecutor struct {
	cache           *lruCache
	mutex           sync.RWMutex
	logger          *slog.Logger
	cleanupTicker   *time.Ticker
	cleanupStopChan chan struct{}
}

// NewTokenExecutor creates a new TokenExecutor with LRU cache and periodic cleanup
func NewTokenExecutor(logger *slog.Logger) *TokenExecutor {
	const maxCacheSize = 100
	const cleanupInterval = 5 * time.Minute

	te := &TokenExecutor{
		cache:           newLRUCache(maxCacheSize),
		logger:          logger,
		cleanupStopChan: make(chan struct{}),
	}

	// Start periodic cleanup goroutine
	te.cleanupTicker = time.NewTicker(cleanupInterval)
	go te.periodicCleanup()

	return te
}

// Close stops the periodic cleanup and releases resources
func (te *TokenExecutor) Close() {
	if te.cleanupTicker != nil {
		te.cleanupTicker.Stop()
	}
	close(te.cleanupStopChan)
}

// periodicCleanup runs in a goroutine to periodically clean expired tokens
func (te *TokenExecutor) periodicCleanup() {
	for {
		select {
		case <-te.cleanupTicker.C:
			te.mutex.Lock()
			removed := te.cache.removeExpired()
			te.mutex.Unlock()

			if removed > 0 {
				te.logger.Debug("cleaned up expired tokens", "removed_count", removed)
			}

		case <-te.cleanupStopChan:
			return
		}
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
	// Create comprehensive cache key that includes environment variables
	cacheKey := te.createCacheKey(edgeName, execConfig)

	// Check cache first
	te.mutex.RLock()
	if cached, exists := te.cache.get(cacheKey); exists && !cached.IsExpired() {
		te.mutex.RUnlock()
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cached.Token)))[:8]
		te.logger.Debug("using cached bearer token", "edge", edgeName, "token_hash", tokenHash, "token_length", len(cached.Token))
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
	te.cache.put(cacheKey, &TokenCache{
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute), // Default 15 minute cache
	})
	te.mutex.Unlock()

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))[:8]
	te.logger.Debug("successfully retrieved bearer token", "edge", edgeName, "token_hash", tokenHash, "token_length", len(token))
	return token, nil
}

// createCacheKey creates a comprehensive cache key including environment variables
func (te *TokenExecutor) createCacheKey(edgeName string, execConfig *ExecConfig) string {
	// Start with basic components
	parts := []string{
		edgeName,
		execConfig.Command,
		strings.Join(execConfig.Args, " "),
		execConfig.Timeout,
	}

	// Add environment variables in deterministic order
	if len(execConfig.Env) > 0 {
		envPairs := make([]string, 0, len(execConfig.Env))
		for _, env := range execConfig.Env {
			envPairs = append(envPairs, fmt.Sprintf("%s=%s", env.Name, env.Value))
		}
		sort.Strings(envPairs) // Ensure deterministic ordering
		parts = append(parts, strings.Join(envPairs, ","))
	}

	// Create hash of all components for a stable, collision-resistant key
	keyData := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Sprintf("%s:%x", edgeName, hash[:8]) // Use first 8 bytes of hash for readability
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
	te.cache.clear()
}

// ClearExpiredTokens removes expired tokens from the cache
func (te *TokenExecutor) ClearExpiredTokens() int {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	return te.cache.removeExpired()
}

// RefreshToken forces a refresh of the token for a specific edge
func (te *TokenExecutor) RefreshToken(edgeName string, auth *MetricsAuth) (string, error) {
	if auth == nil || auth.BearerTokenExec == nil {
		return te.GetBearerToken(edgeName, auth)
	}

	// Clear cached token for this edge
	cacheKey := te.createCacheKey(edgeName, auth.BearerTokenExec)
	te.mutex.Lock()
	te.cache.remove(cacheKey)
	te.mutex.Unlock()

	// Get fresh token
	return te.GetBearerToken(edgeName, auth)
}

// GetCacheStats returns statistics about the token cache
func (te *TokenExecutor) GetCacheStats() map[string]any {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	total := te.cache.size()
	expired := 0

	// Count expired entries without modifying cache
	for _, entry := range te.cache.entries {
		if entry.token.IsExpired() {
			expired++
		}
	}

	return map[string]any{
		"total_tokens":   total,
		"expired_tokens": expired,
		"active_tokens":  total - expired,
		"max_size":       te.cache.maxSize,
	}
}
