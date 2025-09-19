//go:build test

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
)

func TestLRUCache_BasicOperations(t *testing.T) {
	cache := newLRUCache(3)

	// Test put and get
	token1 := &TokenCache{Token: "token1", ExpiresAt: time.Now().Add(time.Hour)}
	cache.put("key1", token1)

	retrieved, exists := cache.get("key1")
	if !exists {
		t.Fatal("Expected key1 to exist")
	}
	if retrieved.Token != "token1" {
		t.Errorf("Expected token1, got %s", retrieved.Token)
	}

	// Test non-existent key
	_, exists = cache.get("nonexistent")
	if exists {
		t.Error("Expected nonexistent key to not exist")
	}
}

func TestLRUCache_LRUEviction(t *testing.T) {
	cache := newLRUCache(2) // Small cache for testing eviction

	// Fill cache to capacity
	token1 := &TokenCache{Token: "token1", ExpiresAt: time.Now().Add(time.Hour)}
	token2 := &TokenCache{Token: "token2", ExpiresAt: time.Now().Add(time.Hour)}
	cache.put("key1", token1)
	cache.put("key2", token2)

	if cache.size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.size())
	}

	// Add third item, should evict LRU (key1)
	token3 := &TokenCache{Token: "token3", ExpiresAt: time.Now().Add(time.Hour)}
	cache.put("key3", token3)

	if cache.size() != 2 {
		t.Errorf("Expected cache size 2 after eviction, got %d", cache.size())
	}

	// key1 should be evicted
	_, exists := cache.get("key1")
	if exists {
		t.Error("Expected key1 to be evicted")
	}

	// key2 and key3 should still exist
	_, exists = cache.get("key2")
	if !exists {
		t.Error("Expected key2 to still exist")
	}
	_, exists = cache.get("key3")
	if !exists {
		t.Error("Expected key3 to still exist")
	}
}

func TestLRUCache_LRUOrdering(t *testing.T) {
	cache := newLRUCache(3)

	// Add items
	token1 := &TokenCache{Token: "token1", ExpiresAt: time.Now().Add(time.Hour)}
	token2 := &TokenCache{Token: "token2", ExpiresAt: time.Now().Add(time.Hour)}
	token3 := &TokenCache{Token: "token3", ExpiresAt: time.Now().Add(time.Hour)}

	cache.put("key1", token1)
	cache.put("key2", token2)
	cache.put("key3", token3)

	// Access key1 to make it most recently used
	cache.get("key1")

	// Add key4, should evict key2 (least recently used)
	token4 := &TokenCache{Token: "token4", ExpiresAt: time.Now().Add(time.Hour)}
	cache.put("key4", token4)

	// key2 should be evicted, others should remain
	_, exists := cache.get("key2")
	if exists {
		t.Error("Expected key2 to be evicted")
	}

	_, exists = cache.get("key1")
	if !exists {
		t.Error("Expected key1 to still exist")
	}
	_, exists = cache.get("key3")
	if !exists {
		t.Error("Expected key3 to still exist")
	}
	_, exists = cache.get("key4")
	if !exists {
		t.Error("Expected key4 to still exist")
	}
}

func TestLRUCache_ExpiredTokenRemoval(t *testing.T) {
	cache := newLRUCache(5)

	// Add mix of expired and valid tokens
	expiredToken := &TokenCache{Token: "expired", ExpiresAt: time.Now().Add(-time.Hour)}
	validToken := &TokenCache{Token: "valid", ExpiresAt: time.Now().Add(time.Hour)}

	cache.put("expired1", expiredToken)
	cache.put("valid1", validToken)
	cache.put("expired2", &TokenCache{Token: "expired2", ExpiresAt: time.Now().Add(-time.Hour)})

	if cache.size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.size())
	}

	// Remove expired tokens
	removed := cache.removeExpired()
	if removed != 2 {
		t.Errorf("Expected 2 expired tokens removed, got %d", removed)
	}

	if cache.size() != 1 {
		t.Errorf("Expected cache size 1 after cleanup, got %d", cache.size())
	}

	// Only valid token should remain
	_, exists := cache.get("valid1")
	if !exists {
		t.Error("Expected valid token to remain")
	}
	_, exists = cache.get("expired1")
	if exists {
		t.Error("Expected expired token to be removed")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := newLRUCache(5)

	// Add some tokens
	token1 := &TokenCache{Token: "token1", ExpiresAt: time.Now().Add(time.Hour)}
	cache.put("key1", token1)
	cache.put("key2", token1)

	if cache.size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.size())
	}

	// Clear cache
	cache.clear()

	if cache.size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.size())
	}

	// Verify entries are gone
	_, exists := cache.get("key1")
	if exists {
		t.Error("Expected key1 to be cleared")
	}
}

func TestLRUCache_UpdateExistingEntry(t *testing.T) {
	cache := newLRUCache(3)

	// Add initial token
	token1 := &TokenCache{Token: "token1", ExpiresAt: time.Now().Add(time.Hour)}
	cache.put("key1", token1)

	// Update with new token
	token2 := &TokenCache{Token: "token2", ExpiresAt: time.Now().Add(2 * time.Hour)}
	cache.put("key1", token2)

	// Should still have size 1
	if cache.size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.size())
	}

	// Should have updated token
	retrieved, exists := cache.get("key1")
	if !exists {
		t.Fatal("Expected key1 to exist")
	}
	if retrieved.Token != "token2" {
		t.Errorf("Expected updated token2, got %s", retrieved.Token)
	}
}

func TestCreateCacheKey_IncludesEnvironmentVariables(t *testing.T) {
	te := &TokenExecutor{}

	execConfig := &ExecConfig{
		Command: "kubectl",
		Args:    []string{"get", "secret"},
		Timeout: "30s",
		Env: []EnvVar{
			{Name: "KUBECONFIG", Value: "/path/to/config"},
			{Name: "NAMESPACE", Value: "default"},
		},
	}

	key1 := te.createCacheKey("edge1", execConfig)

	// Same config should produce same key
	key2 := te.createCacheKey("edge1", execConfig)
	if key1 != key2 {
		t.Error("Expected same config to produce same cache key")
	}

	// Different environment should produce different key
	execConfig2 := &ExecConfig{
		Command: "kubectl",
		Args:    []string{"get", "secret"},
		Timeout: "30s",
		Env: []EnvVar{
			{Name: "KUBECONFIG", Value: "/different/path"},
			{Name: "NAMESPACE", Value: "default"},
		},
	}

	key3 := te.createCacheKey("edge1", execConfig2)
	if key1 == key3 {
		t.Error("Expected different environment to produce different cache key")
	}

	// Different edge name should produce different key
	key4 := te.createCacheKey("edge2", execConfig)
	if key1 == key4 {
		t.Error("Expected different edge name to produce different cache key")
	}
}

func TestCreateCacheKey_DeterministicEnvironmentOrdering(t *testing.T) {
	te := &TokenExecutor{}

	// Same environment variables in different order should produce same key
	execConfig1 := &ExecConfig{
		Command: "kubectl",
		Args:    []string{"get", "secret"},
		Env: []EnvVar{
			{Name: "KUBECONFIG", Value: "/path/to/config"},
			{Name: "NAMESPACE", Value: "default"},
		},
	}

	execConfig2 := &ExecConfig{
		Command: "kubectl",
		Args:    []string{"get", "secret"},
		Env: []EnvVar{
			{Name: "NAMESPACE", Value: "default"},
			{Name: "KUBECONFIG", Value: "/path/to/config"},
		},
	}

	key1 := te.createCacheKey("edge1", execConfig1)
	key2 := te.createCacheKey("edge1", execConfig2)

	if key1 != key2 {
		t.Error("Expected same environment variables in different order to produce same cache key")
	}
}
