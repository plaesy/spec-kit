package reach

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func result(data string) BackendResult {
	return BackendResult{Success: true, Data: data, Platform: "demo", Backend: "demo-cli"}
}

func TestMemoryCache_RoundTripAndMiss(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(10, time.Minute)
	defer c.Close()

	if _, ok := c.Get(ctx, "absent"); ok {
		t.Fatal("a key never set must miss")
	}
	if err := c.Set(ctx, "k", result("v"), 0); err != nil {
		t.Fatal(err)
	}
	got, ok := c.Get(ctx, "k")
	if !ok || got.Data != "v" {
		t.Fatalf("Get = %+v, %v; want the stored value", got, ok)
	}
}

func TestMemoryCache_ExpiredEntryIsAMiss(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(10, time.Minute)
	defer c.Close()

	if err := c.Set(ctx, "k", result("v"), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, ok := c.Get(ctx, "k"); ok {
		t.Fatal("an expired entry must not be returned")
	}
	if n := c.Stats().Entries; n != 0 {
		t.Errorf("Entries = %d after expiry, want 0", n)
	}
}

func TestMemoryCache_EvictsOldestAtCapacity(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(2, time.Minute)
	defer c.Close()

	c.Set(ctx, "a", result("1"), 0)
	time.Sleep(time.Millisecond)
	c.Set(ctx, "b", result("2"), 0)
	time.Sleep(time.Millisecond)
	c.Set(ctx, "c", result("3"), 0)

	if _, ok := c.Get(ctx, "a"); ok {
		t.Error("the oldest entry should have been evicted")
	}
	if _, ok := c.Get(ctx, "c"); !ok {
		t.Error("the newest entry should be kept")
	}
	if n := c.Stats().Entries; n != 2 {
		t.Errorf("Entries = %d, want 2", n)
	}
}

// Overwriting a key that is already present replaces it. It must not evict a
// different entry, and it must not count as a second entry.
func TestMemoryCache_OverwriteKeepsOtherEntriesAndCount(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(2, time.Minute)
	defer c.Close()

	c.Set(ctx, "a", result("1"), 0)
	time.Sleep(time.Millisecond)
	c.Set(ctx, "b", result("2"), 0)
	// b is the newer entry. Overwriting it must not evict the older a.
	c.Set(ctx, "b", result("2-new"), 0)

	if got, ok := c.Get(ctx, "a"); !ok || got.Data != "1" {
		t.Errorf("overwriting b evicted a: got %+v, %v", got, ok)
	}
	if got, ok := c.Get(ctx, "b"); !ok || got.Data != "2-new" {
		t.Errorf("b = %+v, %v; want the new value", got, ok)
	}
	if n := c.Stats().Entries; n != 2 {
		t.Errorf("Entries = %d after an overwrite, want 2", n)
	}
}

func TestMemoryCache_DeleteClearAndStats(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(10, time.Minute)
	defer c.Close()

	c.Set(ctx, "a", result("1"), 0)
	c.Set(ctx, "b", result("2"), 0)
	c.Get(ctx, "a")       // hit
	c.Get(ctx, "missing") // miss

	st := c.Stats()
	if st.Hits != 1 || st.Misses != 1 || st.HitRate != 0.5 {
		t.Errorf("stats = %+v, want 1 hit, 1 miss, rate 0.5", st)
	}

	if err := c.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get(ctx, "a"); ok {
		t.Error("a deleted key must miss")
	}

	if err := c.Clear(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get(ctx, "b"); ok {
		t.Error("Clear must remove every entry")
	}
}

func TestMemoryCache_CleanupLoopRemovesExpiredEntries(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache(10, time.Millisecond)
	defer c.Close()

	c.Set(ctx, "k", result("v"), time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	c.evictExpired()
	c.mu.RLock()
	_, present := c.entries["k"]
	c.mu.RUnlock()
	if present {
		t.Error("evictExpired must remove an entry past its expiry")
	}
}

func TestFileCache_RoundTripPersistsAcrossInstances(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	c, err := NewFileCache(dir, 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Set(ctx, "k", result("persisted"), 0); err != nil {
		t.Fatal(err)
	}

	again, err := NewFileCache(dir, 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := again.Get(ctx, "k")
	if !ok || got.Data != "persisted" {
		t.Fatalf("a new instance on the same dir = %+v, %v; want the stored value", got, ok)
	}
}

func TestFileCache_ExpiredAndCorruptFilesAreMisses(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	c, err := NewFileCache(dir, 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	c.Set(ctx, "expired", result("x"), time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if _, ok := c.Get(ctx, "expired"); ok {
		t.Error("an expired file entry must miss")
	}
	if _, err := os.Stat(c.cachePath("expired")); !os.IsNotExist(err) {
		t.Error("an expired file should be removed from disk")
	}

	if err := os.WriteFile(c.cachePath("broken"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get(ctx, "broken"); ok {
		t.Error("a corrupt file must miss, not panic or return a value")
	}
}

func TestFileCache_DeleteClearAndStats(t *testing.T) {
	ctx := context.Background()
	c, err := NewFileCache(t.TempDir(), 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	c.Set(ctx, "a", result("1"), 0)
	c.Get(ctx, "a")       // hit
	c.Get(ctx, "missing") // miss

	st := c.Stats()
	if st.Hits != 1 || st.Misses != 1 || st.HitRate != 0.5 {
		t.Errorf("stats = %+v, want 1 hit, 1 miss, rate 0.5", st)
	}

	if err := c.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get(ctx, "a"); ok {
		t.Error("a deleted key must miss")
	}

	c.Set(ctx, "b", result("2"), 0)
	if err := c.Clear(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get(ctx, "b"); ok {
		t.Error("Clear must remove every file")
	}
}

// Overwriting a file entry must not count it twice in Entries.
func TestFileCache_OverwriteDoesNotInflateEntries(t *testing.T) {
	ctx := context.Background()
	c, err := NewFileCache(t.TempDir(), 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	c.Set(ctx, "k", result("1"), 0)
	c.Set(ctx, "k", result("2"), 0)
	if n := c.Stats().Entries; n != 1 {
		t.Errorf("Entries = %d after overwriting one key, want 1", n)
	}
}

// The configured disk limit must hold: writing past it removes the oldest files.
func TestFileCache_EnforcesSizeLimit(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	c, err := NewFileCache(dir, 1, time.Minute) // 1 MB limit
	if err != nil {
		t.Fatal(err)
	}
	big := strings.Repeat("x", 600*1024)
	for _, k := range []string{"first", "second", "third"} {
		if err := c.Set(ctx, k, result(big), 0); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	var total int64
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		info, _ := e.Info()
		total += info.Size()
	}
	if total > 1024*1024 {
		t.Errorf("cache holds %d bytes, over its 1 MB limit", total)
	}
	if _, ok := c.Get(ctx, "third"); !ok {
		t.Error("the newest entry should survive the size limit")
	}
	if _, ok := c.Get(ctx, "first"); ok {
		t.Error("the oldest entry should have been removed to honour the limit")
	}
}

func TestFileCache_RejectsUnusableDirectory(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileCache(filepath.Join(blocker, "sub"), 1, time.Minute); err == nil {
		t.Error("a cache dir under a regular file must fail")
	}
}

func TestTieredCache_WritesBothTiersAndPromotes(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	tc, err := NewTieredCache(10, dir, 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer tc.Close()

	tc.Set(ctx, "k", result("v"), 0)
	if _, ok := tc.memory.Get(ctx, "k"); !ok {
		t.Error("Set must write the memory tier")
	}
	if _, ok := tc.file.Get(ctx, "k"); !ok {
		t.Error("Set must write the file tier")
	}

	// A value only on disk is found and copied up to memory.
	tc.memory.Delete(ctx, "k")
	if got, ok := tc.Get(ctx, "k"); !ok || got.Data != "v" {
		t.Fatalf("Get from disk = %+v, %v", got, ok)
	}
	if _, ok := tc.memory.Get(ctx, "k"); !ok {
		t.Error("a file hit must be promoted into memory")
	}
}

// A lookup that misses memory and hits disk is one hit overall, not a hit plus
// a miss. Stats must agree with what the caller saw.
func TestTieredCache_StatsCountEachLookupOnce(t *testing.T) {
	ctx := context.Background()
	tc, err := NewTieredCache(10, t.TempDir(), 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer tc.Close()

	tc.Set(ctx, "k", result("v"), 0)
	tc.memory.Delete(ctx, "k")
	tc.Get(ctx, "k")       // memory miss, disk hit
	tc.Get(ctx, "missing") // miss in both

	st := tc.Stats()
	if st.Hits != 1 || st.Misses != 1 {
		t.Errorf("stats = %+v, want 1 hit and 1 miss", st)
	}
	if st.HitRate != 0.5 {
		t.Errorf("HitRate = %v, want 0.5", st.HitRate)
	}
}

func TestTieredCache_ClearAndDelete(t *testing.T) {
	ctx := context.Background()
	tc, err := NewTieredCache(10, t.TempDir(), 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer tc.Close()

	tc.Set(ctx, "a", result("1"), 0)
	tc.Set(ctx, "b", result("2"), 0)
	if err := tc.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tc.Get(ctx, "a"); ok {
		t.Error("a deleted key must miss in both tiers")
	}
	if err := tc.Clear(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := tc.Get(ctx, "b"); ok {
		t.Error("Clear must empty both tiers")
	}
}

func TestCacheKey_StableAndSensitiveToEveryPart(t *testing.T) {
	base := CacheKey("web", "jina-reader", "q", map[string]string{"a": "1", "b": "2"})
	if again := CacheKey("web", "jina-reader", "q", map[string]string{"b": "2", "a": "1"}); again != base {
		t.Error("the key must not depend on map iteration order")
	}
	variants := []string{
		CacheKey("github", "jina-reader", "q", map[string]string{"a": "1", "b": "2"}),
		CacheKey("web", "curl", "q", map[string]string{"a": "1", "b": "2"}),
		CacheKey("web", "jina-reader", "other", map[string]string{"a": "1", "b": "2"}),
		CacheKey("web", "jina-reader", "q", map[string]string{"a": "9", "b": "2"}),
	}
	for i, v := range variants {
		if v == base {
			t.Errorf("variant %d has the same key as the base; parts must change the key", i)
		}
	}
}

func TestPlatformRateLimiter_EachPlatformGetsItsOwnBucket(t *testing.T) {
	ctx := context.Background()
	p := NewPlatformRateLimiter(map[Platform]RateLimitConfig{
		PlatformWeb:    {RequestsPerSecond: 0.001, Burst: 1, Algorithm: "token-bucket"},
		PlatformGitHub: {RequestsPerSecond: 0.001, Burst: 1, Algorithm: "token-bucket"},
	})
	if !p.Allow(ctx, PlatformWeb) {
		t.Fatal("the first web request should pass")
	}
	if p.Allow(ctx, PlatformWeb) {
		t.Error("web's burst of 1 is spent, so a second request must be refused")
	}
	if !p.Allow(ctx, PlatformGitHub) {
		t.Error("github has its own bucket and must not be affected by web")
	}
}
