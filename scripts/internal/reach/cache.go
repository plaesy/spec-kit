// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheEntry represents a cached response.
type CacheEntry struct {
	Key       string
	Value     BackendResult
	CreatedAt time.Time
	ExpiresAt time.Time
	Hits      int64
}

// CacheBackend defines the interface for cache storage.
type CacheBackend interface {
	Get(ctx context.Context, key string) (BackendResult, bool)
	Set(ctx context.Context, key string, value BackendResult, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Stats() CacheStats
}

// CacheStats holds cache statistics.
type CacheStats struct {
	Entries   int64
	Hits      int64
	Misses    int64
	HitRate   float64
	SizeBytes int64
}

// MemoryCache is an in-memory cache implementation with TTL and LRU eviction.
type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int
	stats   CacheStats
	ttl     time.Duration
	cleanup *time.Ticker
	stopCh  chan struct{}
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache(maxSize int, defaultTTL time.Duration) *MemoryCache {
	c := &MemoryCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     defaultTTL,
		stopCh:  make(chan struct{}),
	}
	c.cleanup = time.NewTicker(5 * time.Minute)
	go c.cleanupLoop()
	return c
}

func (c *MemoryCache) cleanupLoop() {
	for {
		select {
		case <-c.cleanup.C:
			c.evictExpired()
		case <-c.stopCh:
			c.cleanup.Stop()
			return
		}
	}
}

func (c *MemoryCache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
			c.stats.Entries--
		}
	}

	// LRU eviction if over max size
	if len(c.entries) > c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for key, entry := range c.entries {
			if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
				oldestKey = key
				oldestTime = entry.CreatedAt
			}
		}
		if oldestKey != "" {
			delete(c.entries, oldestKey)
			c.stats.Entries--
		}
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string) (BackendResult, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return BackendResult{}, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.stats.Entries--
		c.stats.Misses++
		c.mu.Unlock()
		return BackendResult{}, false
	}

	c.mu.Lock()
	entry.Hits++
	c.stats.Hits++
	c.mu.Unlock()
	return entry.Value, true
}

func (c *MemoryCache) Set(ctx context.Context, key string, value BackendResult, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.ttl
	}

	now := time.Now()
	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		Hits:      0,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity
	if len(c.entries) >= c.maxSize && c.maxSize > 0 {
		var oldestKey string
		var oldestTime time.Time
		for k, e := range c.entries {
			if oldestKey == "" || e.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = e.CreatedAt
			}
		}
		if oldestKey != "" {
			delete(c.entries, oldestKey)
			c.stats.Entries--
		}
	}

	c.entries[key] = entry
	c.stats.Entries++
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.entries[key]; ok {
		delete(c.entries, key)
		c.stats.Entries--
	}
	return nil
}

func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
	c.stats.Entries = 0
	c.stats.Hits = 0
	c.stats.Misses = 0
	return nil
}

func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	stats := c.stats
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}
	return stats
}

func (c *MemoryCache) Close() error {
	close(c.stopCh)
	return nil
}

// FileCache is a file-based persistent cache.
type FileCache struct {
	dir     string
	mu      sync.RWMutex
	maxSize int64
	stats   CacheStats
	ttl     time.Duration
}

// NewFileCache creates a new file-based cache.
func NewFileCache(dir string, maxSizeMB int, defaultTTL time.Duration) (*FileCache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}
	return &FileCache{
		dir:     dir,
		maxSize: int64(maxSizeMB) * 1024 * 1024,
		ttl:     defaultTTL,
	}, nil
}

func (f *FileCache) cachePath(key string) string {
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(f.dir, hex.EncodeToString(hash[:16])+".cache")
}

func (f *FileCache) Get(ctx context.Context, key string) (BackendResult, bool) {
	path := f.cachePath(key)
	f.mu.RLock()
	defer f.mu.RUnlock()

	data, err := os.ReadFile(path)
	if err != nil {
		f.mu.RUnlock()
		f.mu.Lock()
		f.stats.Misses++
		f.mu.Unlock()
		f.mu.RLock()
		return BackendResult{}, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		f.mu.RUnlock()
		f.mu.Lock()
		f.stats.Misses++
		f.mu.Unlock()
		f.mu.RLock()
		return BackendResult{}, false
	}

	if time.Now().After(entry.ExpiresAt) {
		os.Remove(path)
		f.mu.RUnlock()
		f.mu.Lock()
		f.stats.Misses++
		f.mu.Unlock()
		f.mu.RLock()
		return BackendResult{}, false
	}

	f.mu.RUnlock()
	f.mu.Lock()
	entry.Hits++
	f.stats.Hits++
	f.mu.Unlock()
	f.mu.RLock()

	// Update hit count
	entryData, _ := json.Marshal(entry)
	os.WriteFile(path, entryData, 0644)

	return entry.Value, true
}

func (f *FileCache) Set(ctx context.Context, key string, value BackendResult, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = f.ttl
	}

	entry := CacheEntry{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		Hits:      0,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	path := f.cachePath(key)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	f.mu.Lock()
	f.stats.Entries++
	f.mu.Unlock()
	return nil
}

func (f *FileCache) Delete(ctx context.Context, key string) error {
	path := f.cachePath(key)
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := os.Remove(path); err == nil {
		f.stats.Entries--
	}
	return nil
}

func (f *FileCache) Clear(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	entries, _ := os.ReadDir(f.dir)
	for _, e := range entries {
		os.Remove(filepath.Join(f.dir, e.Name()))
	}
	f.stats.Entries = 0
	f.stats.Hits = 0
	f.stats.Misses = 0
	return nil
}

func (f *FileCache) Stats() CacheStats {
	f.mu.RLock()
	defer f.mu.RUnlock()
	stats := f.stats
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}
	return stats
}

// TieredCache combines memory and file caches.
type TieredCache struct {
	memory *MemoryCache
	file   *FileCache
}

// NewTieredCache creates a tiered cache (memory L1 + file L2).
func NewTieredCache(memorySize int, fileDir string, fileSizeMB int, ttl time.Duration) (*TieredCache, error) {
	fileCache, err := NewFileCache(fileDir, fileSizeMB, ttl)
	if err != nil {
		return nil, err
	}
	return &TieredCache{
		memory: NewMemoryCache(memorySize, ttl),
		file:   fileCache,
	}, nil
}

func (t *TieredCache) Get(ctx context.Context, key string) (BackendResult, bool) {
	// Try L1 (memory)
	if val, ok := t.memory.Get(ctx, key); ok {
		return val, true
	}
	// Try L2 (file)
	if val, ok := t.file.Get(ctx, key); ok {
		// Promote to L1
		t.memory.Set(ctx, key, val, t.file.ttl)
		return val, true
	}
	return BackendResult{}, false
}

func (t *TieredCache) Set(ctx context.Context, key string, value BackendResult, ttl time.Duration) error {
	t.memory.Set(ctx, key, value, ttl)
	return t.file.Set(ctx, key, value, ttl)
}

func (t *TieredCache) Delete(ctx context.Context, key string) error {
	t.memory.Delete(ctx, key)
	return t.file.Delete(ctx, key)
}

func (t *TieredCache) Clear(ctx context.Context) error {
	t.memory.Clear(ctx)
	return t.file.Clear(ctx)
}

func (t *TieredCache) Stats() CacheStats {
	memStats := t.memory.Stats()
	fileStats := t.file.Stats()
	return CacheStats{
		Entries: memStats.Entries + fileStats.Entries,
		Hits:    memStats.Hits + fileStats.Hits,
		Misses:  memStats.Misses + fileStats.Misses,
		HitRate: 0, // Combined rate
	}
}

func (t *TieredCache) Close() error {
	return t.memory.Close()
}

// CacheKey generates a cache key from platform, backend, query, and options.
func CacheKey(platform Platform, backend Backend, query string, opts map[string]string) string {
	data := fmt.Sprintf("%s|%s|%s|%v", platform, backend, query, opts)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}
