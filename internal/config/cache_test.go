// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
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
	"sync"
	"testing"
	"time"
)

// TestConfigCache_NewConfigCache tests cache creation
func TestConfigCache_NewConfigCache(t *testing.T) {
	cache := NewConfigCache()
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.Size() != 0 {
		t.Errorf("expected empty cache, got size %d", cache.Size())
	}
}

// TestConfigCache_SetAndGet tests basic set and get operations
func TestConfigCache_SetAndGet(t *testing.T) {
	ctx := context.Background()
	cache := NewConfigCache()

	// Create a minimal test config
	testConfig := &Config{}

	// Test Get on empty cache
	config, found := cache.Get(ctx, "test-cluster")
	if found {
		t.Error("expected not found in empty cache")
	}
	if config != nil {
		t.Error("expected nil config from empty cache")
	}

	// Test Set
	cache.Set(ctx, "test-cluster", testConfig)

	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}

	// Test Get after Set
	config, found = cache.Get(ctx, "test-cluster")
	if !found {
		t.Error("expected to find config in cache")
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
}

// TestConfigCache_Invalidate tests invalidating a cache entry
func TestConfigCache_Invalidate(t *testing.T) {
	ctx := context.Background()
	cache := NewConfigCache()

	testConfig := &Config{}
	cache.Set(ctx, "test-cluster", testConfig)

	// Verify entry exists
	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}

	// Invalidate entry
	cache.Invalidate(ctx, "test-cluster")

	// Verify entry is gone
	if cache.Size() != 0 {
		t.Errorf("expected empty cache after invalidate, got size %d", cache.Size())
	}

	config, found := cache.Get(ctx, "test-cluster")
	if found {
		t.Error("expected not found after invalidate")
	}
	if config != nil {
		t.Error("expected nil config after invalidate")
	}
}

// TestConfigCache_Clear tests clearing all cache entries
func TestConfigCache_Clear(t *testing.T) {
	ctx := context.Background()
	cache := NewConfigCache()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		config := &Config{}
		cache.Set(ctx, "cluster-"+string(rune('0'+i)), config)
	}

	if cache.Size() != 5 {
		t.Errorf("expected cache size 5, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear(ctx)

	if cache.Size() != 0 {
		t.Errorf("expected empty cache after clear, got size %d", cache.Size())
	}

	// Verify all entries are gone
	for i := 0; i < 5; i++ {
		_, found := cache.Get(ctx, "cluster-"+string(rune('0'+i)))
		if found {
			t.Errorf("expected cluster-%d not found after clear", i)
		}
	}
}

// TestConfigCache_Expiration tests cache entry expiration
func TestConfigCache_Expiration(t *testing.T) {
	ctx := context.Background()
	cache := NewConfigCache()

	testConfig := &Config{}

	// Set with expiration in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	cache.SetWithExpiration(ctx, "test-cluster", testConfig, pastTime)

	// Should not find expired entry
	config, found := cache.Get(ctx, "test-cluster")
	if found {
		t.Error("expected not found for expired entry")
	}
	if config != nil {
		t.Error("expected nil config for expired entry")
	}

	// Set with expiration in the future
	futureTime := time.Now().Add(1 * time.Hour)
	cache.SetWithExpiration(ctx, "test-cluster", testConfig, futureTime)

	// Should find non-expired entry
	config, found = cache.Get(ctx, "test-cluster")
	if !found {
		t.Error("expected to find non-expired entry")
	}
	if config == nil {
		t.Error("expected non-nil config for non-expired entry")
	}
}

// TestConfigCache_ThreadSafety tests concurrent access to cache
func TestConfigCache_ThreadSafety(t *testing.T) {
	ctx := context.Background()
	cache := NewConfigCache()

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				clusterName := "cluster-" + string(rune('0'+(id%10)))

				// Perform random operations
				switch j % 4 {
				case 0: // Set
					config := &Config{}
					cache.Set(ctx, clusterName, config)

				case 1: // Get
					cache.Get(ctx, clusterName)

				case 2: // Invalidate
					cache.Invalidate(ctx, clusterName)

				case 3: // Size
					cache.Size()
				}
			}
		}(i)
	}

	wg.Wait()

	// If we get here without deadlock or panic, thread safety is working
	t.Log("Thread safety test completed successfully")
}

// TestConfigCache_MultipleEntries tests managing multiple cache entries
func TestConfigCache_MultipleEntries(t *testing.T) {
	ctx := context.Background()
	cache := NewConfigCache()

	// Add multiple different entries
	clusters := []string{"dev-cluster", "staging-cluster", "prod-cluster"}

	for _, name := range clusters {
		config := &Config{}
		cache.Set(ctx, name, config)
	}

	if cache.Size() != len(clusters) {
		t.Errorf("expected cache size %d, got %d", len(clusters), cache.Size())
	}

	// Verify all entries exist
	for _, name := range clusters {
		config, found := cache.Get(ctx, name)
		if !found {
			t.Errorf("expected to find %s in cache", name)
		}
		if config == nil {
			t.Errorf("expected non-nil config for %s", name)
		}
	}

	// Invalidate one entry
	cache.Invalidate(ctx, "staging-cluster")

	if cache.Size() != len(clusters)-1 {
		t.Errorf("expected cache size %d after invalidate, got %d", len(clusters)-1, cache.Size())
	}

	// Verify invalidated entry is gone
	_, found := cache.Get(ctx, "staging-cluster")
	if found {
		t.Error("expected staging-cluster not found after invalidate")
	}

	// Verify other entries still exist
	for _, name := range []string{"dev-cluster", "prod-cluster"} {
		_, found := cache.Get(ctx, name)
		if !found {
			t.Errorf("expected to find %s in cache after invalidating different entry", name)
		}
	}
}

// TestConfigCache_SetWithoutExpiration tests that entries without expiration never expire
func TestConfigCache_SetWithoutExpiration(t *testing.T) {
	ctx := context.Background()
	cache := NewConfigCache()

	testConfig := &Config{}

	// Set without expiration (using regular Set)
	cache.Set(ctx, "test-cluster", testConfig)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Should still find the entry
	config, found := cache.Get(ctx, "test-cluster")
	if !found {
		t.Error("expected to find non-expiring entry")
	}
	if config == nil {
		t.Error("expected non-nil config for non-expiring entry")
	}
}

