package reach

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// useTempConfig points reach at a config file inside the test's temp dir, so no
// test reads or writes the user's real configuration.
func useTempConfig(t *testing.T) string {
	t.Helper()
	orig := UnifiedConfigPath
	path := filepath.Join(t.TempDir(), "config.json")
	UnifiedConfigPath = path
	t.Cleanup(func() { UnifiedConfigPath = orig })
	return path
}

// newTestConfigManager is a config manager on a temp file with the OS keyring
// turned off, so tests never touch the real credential store.
func newTestConfigManager(t *testing.T) *ConfigManager {
	t.Helper()
	useTempConfig(t)
	cm := NewConfigManager(filepath.Join(t.TempDir(), "reach.json"))
	cm.SetKeyringEnabled(false)
	return cm
}

// fakeHTTP answers every HTTP request made through the default transport with
// the given status and body. Nothing reaches the network.
func fakeHTTP(t *testing.T, status int, body string) {
	t.Helper()
	orig := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = orig })
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// ---- ConfigManager ----

func TestConfigManager_EnableDisableRoundTrip(t *testing.T) {
	cm := newTestConfigManager(t)

	if err := cm.EnablePlatform(PlatformWeb); err != nil {
		t.Fatal(err)
	}
	pc, err := cm.GetPlatformConfig(PlatformWeb)
	if err != nil {
		t.Fatal(err)
	}
	if !pc.Enabled {
		t.Error("EnablePlatform did not enable the platform")
	}

	if err := cm.DisablePlatform(PlatformWeb); err != nil {
		t.Fatal(err)
	}
	pc, _ = cm.GetPlatformConfig(PlatformWeb)
	if pc.Enabled {
		t.Error("DisablePlatform did not disable the platform")
	}
}

func TestConfigManager_CredentialsWithoutKeyring(t *testing.T) {
	cm := newTestConfigManager(t)

	if err := cm.SetCredentials(PlatformGitHub, map[string]string{"token": "placeholder-token"}); err != nil {
		t.Fatal(err)
	}
	got, err := cm.GetCredential(PlatformGitHub, "token")
	if err != nil || got != "placeholder-token" {
		t.Fatalf("GetCredential = %q, %v; want the stored value", got, err)
	}

	if _, err := cm.GetCredential(PlatformGitHub, "absent"); err == nil {
		t.Error("an absent credential key must be an error")
	}

	if err := cm.DeleteCredential(PlatformGitHub, "token"); err != nil {
		t.Fatal(err)
	}
	if _, err := cm.GetCredential(PlatformGitHub, "token"); err == nil {
		t.Error("a deleted credential must not be returned")
	}
}

func TestConfigManager_SaveThenLoadKeepsSettings(t *testing.T) {
	cm := newTestConfigManager(t)

	cfg := DefaultConfig()
	cfg.CacheSizeMB = 7
	if err := cm.Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.CacheSizeMB != 7 {
		t.Errorf("loaded cache = %dMB, want 7MB", loaded.CacheSizeMB)
	}
	// Known gap, not asserted here: ReachConfig.DefaultTimeout is not in the
	// unified config schema (internal/config/schema.go), so Save drops it.
}

func TestConfigManager_LoadMissingFileUsesDefaults(t *testing.T) {
	cm := newTestConfigManager(t)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || len(cfg.Platforms) == 0 {
		t.Error("loading with no file should give the default platform set")
	}
}

func TestConfigManager_MigrateFromLegacyReadsOldLocation(t *testing.T) {
	cm := newTestConfigManager(t)
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	legacy := filepath.Join(dir, ".plaesy", "reach")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	legacyJSON := `{"cache_dir":"/tmp/old-cache","cache_size_mb":3}`
	if err := os.WriteFile(filepath.Join(legacy, "config.json"), []byte(legacyJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	// Load is what triggers the migration when the unified file has no reach
	// platforms. The legacy file is moved over and removed, and its settings stay.
	migrated, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(legacy, "config.json")); !os.IsNotExist(err) {
		t.Error("the legacy file should be removed once it has been migrated")
	}
	if migrated.CacheSizeMB != 3 {
		t.Errorf("migrated cache = %dMB, want the legacy 3MB", migrated.CacheSizeMB)
	}
}

func TestConfigManager_PlatformDefaultsAreConsistent(t *testing.T) {
	for _, p := range []Platform{PlatformWeb, PlatformGitHub, PlatformYouTube} {
		pc := DefaultPlatformConfig(p)
		if pc.Platform != p {
			t.Errorf("DefaultPlatformConfig(%s).Platform = %s", p, pc.Platform)
		}
		if got := DefaultBackend(p); got == "" {
			t.Errorf("DefaultBackend(%s) is empty", p)
		}
		if got := FallbackBackends(p); got == nil {
			t.Errorf("FallbackBackends(%s) is nil", p)
		}
		_ = CredentialKeys(p)
	}
	if got := DefaultRetryConfigPtr(); got == nil {
		t.Error("DefaultRetryConfigPtr returned nil")
	}
	if got := parseDurationOrDefault("not-a-duration", time.Second); got != time.Second {
		t.Errorf("parseDurationOrDefault(bad) = %v, want the default", got)
	}
	if got := parseDurationOrDefault("2m", time.Second); got != 2*time.Minute {
		t.Errorf("parseDurationOrDefault(2m) = %v", got)
	}
	if got := (PlatformConfig{FallbackBackends: []Backend{"a", "b"}}).FallbackBackendsStr(); len(got) != 2 {
		t.Errorf("FallbackBackendsStr = %v, want two entries", got)
	}
}

// ---- Registry ----

func TestRegistry_QueryHealthAndStats(t *testing.T) {
	ctx := context.Background()
	fakeTool(t, `{"ok":true}`)
	cfg := DefaultConfig()
	reg := NewRegistry(cfg, WithRegistryLogger(DefaultLogger{}))

	if reg.Router() == nil {
		t.Fatal("Router() returned nil")
	}
	res, err := reg.Query(ctx, PlatformV2EX, "topics", map[string]string{"action": "topics"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if !res.Success {
		t.Errorf("Query result = %+v, want success with the fake tool", res)
	}

	health, err := reg.Health(ctx, PlatformV2EX)
	if err != nil || len(health) == 0 {
		t.Errorf("Health = %v, %v; want one entry per backend", health, err)
	}
	all, err := reg.HealthAll(ctx)
	if err != nil || len(all) == 0 {
		t.Errorf("HealthAll = %d platforms, %v", len(all), err)
	}
	_ = reg.Stats()
}

func TestRegistry_OptionsAreApplied(t *testing.T) {
	cache := NewMemoryCache(5, time.Minute)
	defer cache.Close()
	limiter := NewPlatformRateLimiter(DefaultRateLimits())
	reg := NewRegistry(DefaultConfig(),
		WithRegistryCache(cache),
		WithRegistryRateLimiter(limiter),
		WithRegistryLogger(DefaultLogger{}),
		WithRegistryRetryConfig(DefaultRetryConfig()),
	)
	if reg.GetCache() != cache {
		t.Error("WithRegistryCache was not applied")
	}
	if reg.GetRateLimiter() != limiter {
		t.Error("WithRegistryRateLimiter was not applied")
	}
	reg.SetLogger(DefaultLogger{})
}

func TestRegistry_ConfigurePlatformAndReload(t *testing.T) {
	ctx := context.Background()
	useTempConfig(t)
	// GitHub's Configure runs `gh auth setup-git`. A fake keeps the real tool out.
	fakeTool(t, "")
	reg := NewRegistry(DefaultConfig())

	if err := reg.ConfigurePlatform(PlatformGitHub, map[string]string{"token": "placeholder"}); err != nil {
		t.Fatalf("ConfigurePlatform: %v", err)
	}
	if err := reg.ConfigurePlatform("no-such-platform", nil); err == nil {
		t.Error("configuring an unknown platform must be an error")
	}

	if err := reg.LoadConfig(ctx); err != nil {
		t.Errorf("LoadConfig: %v", err)
	}
	if err := reg.ReloadConfig(ctx); err != nil {
		t.Errorf("ReloadConfig: %v", err)
	}
}

// ---- Router ----

func newMockAdapter(name Platform, backends []Backend, results map[Backend]BackendResult) *mockPlatformAdapter {
	return &mockPlatformAdapter{name: name, backends: backends, executeResults: results}
}

func TestRouter_RoutesToTheFirstWorkingBackend(t *testing.T) {
	ctx := context.Background()
	r := NewRouter(DefaultConfig())
	r.RegisterAdapter(newMockAdapter("demo", []Backend{"primary", "fallback"}, map[Backend]BackendResult{
		"primary":  {Success: false, Error: "down"},
		"fallback": {Success: true, Data: "from fallback"},
	}))
	res, err := r.Route(ctx, "demo", "q", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Data != "from fallback" {
		t.Errorf("Data = %q, want the fallback's answer", res.Data)
	}
}

func TestRouter_RejectsUnknownOrBackendlessPlatforms(t *testing.T) {
	ctx := context.Background()
	r := NewRouter(DefaultConfig())
	if _, err := r.Route(ctx, "nowhere", "q", nil); err == nil {
		t.Error("an unregistered platform must be an error")
	}
	r.RegisterAdapter(newMockAdapter("empty", nil, nil))
	if _, err := r.Route(ctx, "empty", "q", nil); err == nil {
		t.Error("a platform with no backends must be an error")
	}
}

func TestRouter_ReportsWhenEveryBackendFails(t *testing.T) {
	ctx := context.Background()
	r := NewRouter(DefaultConfig())
	r.RegisterAdapter(newMockAdapter("demo", []Backend{"a", "b"}, map[Backend]BackendResult{
		"a": {Success: false, Error: "x"},
		"b": {Success: false, Error: "y"},
	}))
	_, err := r.Route(ctx, "demo", "q", nil)
	if err == nil || !strings.Contains(err.Error(), "all backends failed") {
		t.Errorf("err = %v, want an all-backends-failed error", err)
	}
}

// When config disables every backend, no backend is tried, so there is no error
// to wrap. The message must still say why, not print a formatting placeholder.
func TestRouter_DisabledBackendsGiveAReadableError(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.Platforms[PlatformWeb] = PlatformConfig{Platform: PlatformWeb, Enabled: false, PrimaryBackend: "off", FallbackBackends: nil}
	r := NewRouter(cfg)
	r.RegisterAdapter(newMockAdapter(PlatformWeb, []Backend{"curl"}, map[Backend]BackendResult{
		"curl": {Success: true, Data: "should not run"},
	}))
	_, err := r.Route(ctx, PlatformWeb, "q", nil)
	if err == nil {
		t.Fatal("expected an error when every backend is disabled")
	}
	if strings.Contains(err.Error(), "%!") {
		t.Errorf("error %q contains a formatting placeholder", err)
	}
}

func TestRouter_ServesRepeatQueriesFromCache(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache(10, time.Minute)
	defer cache.Close()
	r := NewRouter(DefaultConfig(), WithCache(cache), WithTimeout(5*time.Second))
	adapter := newMockAdapter("demo", []Backend{"only"}, map[Backend]BackendResult{
		"only": {Success: true, Data: "first"},
	})
	r.RegisterAdapter(adapter)

	if _, err := r.Route(ctx, "demo", "q", nil); err != nil {
		t.Fatal(err)
	}
	adapter.executeResults["only"] = BackendResult{Success: true, Data: "changed"}
	res, err := r.Route(ctx, "demo", "q", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Data != "first" {
		t.Errorf("Data = %q, want the cached answer", res.Data)
	}
}

func TestRouter_OptionsAndAccessors(t *testing.T) {
	cache := NewMemoryCache(5, time.Minute)
	defer cache.Close()
	r := NewRouter(DefaultConfig(),
		WithCache(cache),
		WithRateLimiter(NewPlatformRateLimiter(DefaultRateLimits())),
		WithRetryConfig(DefaultRetryConfig()),
		WithTimeout(3*time.Second),
		WithLogger(DefaultLogger{}),
	)
	r.SetLogger(DefaultLogger{})
	r.SetCache(cache)
	r.SetRateLimiter(NewPlatformRateLimiter(DefaultRateLimits()))

	adapter := newMockAdapter("demo", []Backend{"a"}, map[Backend]BackendResult{"a": {Success: true}})
	r.RegisterAdapter(adapter)
	if got, ok := r.GetAdapter("demo"); !ok || got != adapter {
		t.Error("GetAdapter did not return the registered adapter")
	}
	if _, ok := r.GetAdapter("missing"); ok {
		t.Error("GetAdapter found an unregistered platform")
	}
	if got := r.ListPlatforms(); len(got) == 0 {
		t.Error("ListPlatforms returned nothing")
	}

	health, err := r.Health(context.Background(), "demo")
	if err != nil || len(health) == 0 {
		t.Errorf("Health = %v, %v", health, err)
	}
	if _, err := r.Health(context.Background(), "missing"); err == nil {
		t.Error("Health of an unregistered platform must be an error")
	}
}

func TestRouter_CacheabilityAndTTL(t *testing.T) {
	r := NewRouter(&ReachConfig{DefaultCacheTTL: time.Hour})
	if isCacheable(map[string]string{"cache": "false"}) {
		t.Error("a request with cache=false must not be cached")
	}
	if isCacheable(map[string]string{"action": "post"}) {
		t.Error("a mutating action must not be cached")
	}
	if !isCacheable(nil) {
		t.Error("a plain request should be cacheable")
	}
	if got := r.getCacheTTL(map[string]string{"cache_ttl": "5m"}); got != 5*time.Minute {
		t.Errorf("getCacheTTL(5m) = %v", got)
	}
	if got := r.getCacheTTL(nil); got != time.Hour {
		t.Errorf("getCacheTTL default = %v, want the config's hour", got)
	}
}

// ---- RSS builtin: the HTTP path, with the network replaced by a fake ----

func TestRSS_BuiltinParsesRSS2AndAtom(t *testing.T) {
	ctx := context.Background()
	rss2 := `<?xml version="1.0"?><rss version="2.0"><channel><title>Feed</title>
<item><title>First post</title><link>https://example.com/1</link></item>
<item><title>Second post</title><link>https://example.com/2</link></item>
</channel></rss>`
	fakeHTTP(t, http.StatusOK, rss2)
	res, err := NewRSSAdapter().Execute(ctx, "builtin", "https://example.com/feed.xml", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || !strings.Contains(res.Data, "First post") {
		t.Errorf("RSS 2.0 result = %+v, want the item titles", res)
	}

	atom := `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Atom</title>
<entry><title>Atom entry</title><link href="https://example.com/a"/></entry></feed>`
	fakeHTTP(t, http.StatusOK, atom)
	res, err = NewRSSAdapter().Execute(ctx, "builtin", "https://example.com/atom.xml", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success || !strings.Contains(res.Data, "Atom entry") {
		t.Errorf("Atom result = %+v, want the entry title", res)
	}
}

func TestRSS_BuiltinReportsHTTPFailure(t *testing.T) {
	fakeHTTP(t, http.StatusInternalServerError, "server error")
	res, err := NewRSSAdapter().Execute(context.Background(), "builtin", "https://example.com/feed.xml", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Success {
		t.Errorf("a 500 response must not be a success: %+v", res)
	}
}
