package cache

import (
	"context"
	"sync"
	"time"

	"github.com/histopathai/main-service/internal/port/cache"
)

// MemoryCache is an in-memory implementation of the Cache interface
type MemoryCache struct {
	mu      sync.RWMutex
	data    map[string]*cacheEntry
	stats   *cacheStats
	janitor *janitor
}

type cacheEntry struct {
	value     interface{}
	expiresAt *time.Time
}

type cacheStats struct {
	mu        sync.RWMutex
	hits      int64
	misses    int64
	evictions int64
}

type janitor struct {
	interval time.Duration
	stop     chan bool
}

func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	mc := &MemoryCache{
		data:  make(map[string]*cacheEntry),
		stats: &cacheStats{},
	}

	if cleanupInterval > 0 {
		mc.janitor = &janitor{
			interval: cleanupInterval,
			stop:     make(chan bool),
		}
		go mc.runJanitor()
	}

	return mc
}

func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		mc.recordMiss()
		return nil, nil
	}

	// Check if expired
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		mc.recordMiss()
		return nil, nil
	}

	mc.recordHit()
	return entry.value, nil
}

func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		expiry := time.Now().Add(ttl)
		expiresAt = &expiry
	}

	mc.data[key] = &cacheEntry{
		value:     value,
		expiresAt: expiresAt,
	}

	return nil
}

func (mc *MemoryCache) Has(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		return false, nil
	}

	return true, nil
}

func (mc *MemoryCache) Delete(ctx context.Context, key string) (bool, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	_, exists := mc.data[key]
	if !exists {
		return false, nil
	}

	delete(mc.data, key)
	return true, nil
}

func (mc *MemoryCache) DeletePattern(ctx context.Context, pattern string) (int, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	count := 0
	keysToDelete := make([]string, 0)

	for key := range mc.data {
		if matchPattern(key, pattern) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(mc.data, key)
		count++
	}

	mc.stats.mu.Lock()
	mc.stats.evictions += int64(count)
	mc.stats.mu.Unlock()

	return count, nil
}

func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	count := len(mc.data)
	mc.data = make(map[string]*cacheEntry)

	mc.stats.mu.Lock()
	mc.stats.evictions += int64(count)
	mc.stats.mu.Unlock()

	return nil
}

func (mc *MemoryCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]interface{}, len(keys))
	now := time.Now()

	for _, key := range keys {
		entry, exists := mc.data[key]
		if !exists {
			mc.recordMiss()
			continue
		}

		// Check if expired
		if entry.expiresAt != nil && now.After(*entry.expiresAt) {
			mc.recordMiss()
			continue
		}

		mc.recordHit()
		result[key] = entry.value
	}

	return result, nil
}

func (mc *MemoryCache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		expiry := time.Now().Add(ttl)
		expiresAt = &expiry
	}

	for key, value := range items {
		mc.data[key] = &cacheEntry{
			value:     value,
			expiresAt: expiresAt,
		}
	}

	return nil
}

// GetStats returns cache statistics
func (mc *MemoryCache) GetStats(ctx context.Context) (*cache.Stats, error) {
	mc.mu.RLock()
	size := int64(len(mc.data))
	mc.mu.RUnlock()

	mc.stats.mu.RLock()
	defer mc.stats.mu.RUnlock()

	return &cache.Stats{
		Hits:      mc.stats.hits,
		Misses:    mc.stats.misses,
		Size:      size,
		Evictions: mc.stats.evictions,
	}, nil
}

func (mc *MemoryCache) Close() {
	if mc.janitor != nil {
		mc.janitor.stop <- true
	}
}

func (mc *MemoryCache) recordHit() {
	mc.stats.mu.Lock()
	mc.stats.hits++
	mc.stats.mu.Unlock()
}

func (mc *MemoryCache) recordMiss() {
	mc.stats.mu.Lock()
	mc.stats.misses++
	mc.stats.mu.Unlock()
}

func (mc *MemoryCache) runJanitor() {
	ticker := time.NewTicker(mc.janitor.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.deleteExpired()
		case <-mc.janitor.stop:
			return
		}
	}
}

func (mc *MemoryCache) deleteExpired() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	count := 0

	for key, entry := range mc.data {
		if entry.expiresAt != nil && now.After(*entry.expiresAt) {
			delete(mc.data, key)
			count++
		}
	}

	if count > 0 {
		mc.stats.mu.Lock()
		mc.stats.evictions += int64(count)
		mc.stats.mu.Unlock()
	}
}

func matchPattern(key, pattern string) bool {
	// If no wildcard, exact match
	if pattern == "*" {
		return true
	}

	// Split pattern by '*'
	parts := splitPattern(pattern)
	if len(parts) == 0 {
		return key == pattern
	}

	// Check if key matches the pattern
	keyPos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}

		// Find the part in the remaining key
		idx := indexOf(key[keyPos:], part)
		if idx == -1 {
			return false
		}

		// First part must match from the beginning
		if i == 0 && idx != 0 {
			return false
		}

		keyPos += idx + len(part)
	}

	// Last part must match to the end
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if lastPart != "" && !endsWith(key, lastPart) {
			// Unless the pattern ends with '*'
			if !endsWith(pattern, "*") {
				return false
			}
		}
	}

	return true
}

func splitPattern(pattern string) []string {
	var parts []string
	current := ""

	for _, ch := range pattern {
		if ch == '*' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}

	if current != "" || endsWith(pattern, "*") {
		parts = append(parts, current)
	}

	return parts
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func endsWith(s, suffix string) bool {
	if len(suffix) > len(s) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
