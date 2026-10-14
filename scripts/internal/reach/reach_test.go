// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"testing"
	"time"
)

func TestPlatformConstants(t *testing.T) {
	// Test that all platforms are defined
	expectedPlatforms := []Platform{
		PlatformWeb,
		PlatformYouTube,
		PlatformGitHub,
		PlatformRSS,
		PlatformTwitter,
		PlatformReddit,
		PlatformBilibili,
		PlatformXiaoHongShu,
		PlatformLinkedIn,
		PlatformExaSearch,
		PlatformV2EX,
		PlatformXueqiu,
		PlatformXiaoyuzhou,
		PlatformBossZhipin,
		PlatformFacebook,
		PlatformInstagram,
	}

	if len(AllPlatforms) != len(expectedPlatforms) {
		t.Errorf("AllPlatforms length mismatch: got %d, expected %d", len(AllPlatforms), len(expectedPlatforms))
	}

	for i, p := range expectedPlatforms {
		if AllPlatforms[i] != p {
			t.Errorf("Platform mismatch at index %d: got %s, expected %s", i, AllPlatforms[i], p)
		}
	}
}

func TestZeroConfigPlatforms(t *testing.T) {
	zeroConfig := []Platform{
		PlatformWeb,
		PlatformYouTube,
		PlatformGitHub,
		PlatformRSS,
		PlatformBilibili,
		PlatformV2EX,
		PlatformXueqiu,
	}

	if len(ZeroConfigPlatforms) != len(zeroConfig) {
		t.Errorf("ZeroConfigPlatforms length mismatch")
	}

	for i, p := range zeroConfig {
		if ZeroConfigPlatforms[i] != p {
			t.Errorf("ZeroConfigPlatform mismatch at index %d", i)
		}
	}
}

func TestAuthRequiredPlatforms(t *testing.T) {
	authRequired := []Platform{
		PlatformTwitter,
		PlatformReddit,
		PlatformXiaoHongShu,
		PlatformLinkedIn,
		PlatformExaSearch,
		PlatformXiaoyuzhou,
		PlatformBossZhipin,
		PlatformFacebook,
		PlatformInstagram,
	}

	if len(AuthRequiredPlatforms) != len(authRequired) {
		t.Errorf("AuthRequiredPlatforms length mismatch")
	}

	for i, p := range authRequired {
		if AuthRequiredPlatforms[i] != p {
			t.Errorf("AuthRequiredPlatform mismatch at index %d", i)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.Platforms == nil {
		t.Fatal("Platforms map is nil")
	}

	// Check all platforms have configs
	for _, p := range AllPlatforms {
		if _, ok := config.Platforms[p]; !ok {
			t.Errorf("Missing config for platform %s", p)
		}
	}
}

func TestDefaultPlatformConfig(t *testing.T) {
	for _, p := range AllPlatforms {
		pc := DefaultPlatformConfig(p)

		if pc.Platform != p {
			t.Errorf("Platform mismatch for %s: got %s", p, pc.Platform)
		}

		// Check zero-config platforms are enabled by default
		isZeroConfig := false
		for _, zc := range ZeroConfigPlatforms {
			if zc == p {
				isZeroConfig = true
				break
			}
		}

		if isZeroConfig && !pc.Enabled {
			t.Errorf("Zero-config platform %s should be enabled by default", p)
		}

		if !isZeroConfig && pc.Enabled {
			t.Errorf("Auth-required platform %s should be disabled by default", p)
		}

		// Check primary backend is set
		if pc.PrimaryBackend == "" {
			t.Errorf("Platform %s has empty primary backend", p)
		}
	}
}

func TestDefaultBackend(t *testing.T) {
	tests := []struct {
		platform Platform
		expected Backend
	}{
		{PlatformWeb, "jina-reader"},
		{PlatformYouTube, "yt-dlp"},
		{PlatformGitHub, "gh-cli"},
		{PlatformRSS, "builtin"},
		{PlatformTwitter, "twitter-cli"},
		{PlatformReddit, "opencli"},
		{PlatformBilibili, "bili-cli"},
		{PlatformXiaoHongShu, "opencli"},
		{PlatformLinkedIn, "mcp-linkedin"},
		{PlatformExaSearch, "exa-mcp"},
		{PlatformV2EX, "v2ex-api"},
		{PlatformXueqiu, "xueqiu-api"},
		{PlatformXiaoyuzhou, "xiaoyuzhou-mcp"},
		{PlatformBossZhipin, "boss-cdp"},
		{PlatformFacebook, "opencli"},
		{PlatformInstagram, "opencli"},
	}

	for _, tt := range tests {
		backend := DefaultBackend(tt.platform)
		if backend != tt.expected {
			t.Errorf("DefaultBackend(%s) = %s, expected %s", tt.platform, backend, tt.expected)
		}
	}
}

func TestFallbackBackends(t *testing.T) {
	// Platforms with fallbacks
	hasFallbacks := []Platform{
		PlatformTwitter,
		PlatformReddit,
		PlatformBilibili,
		PlatformXiaoHongShu,
		PlatformLinkedIn,
	}

	for _, p := range hasFallbacks {
		fallbacks := FallbackBackends(p)
		if len(fallbacks) == 0 {
			t.Errorf("Platform %s should have fallbacks", p)
		}
	}

	// Platforms without fallbacks (or minimal)
	noFallbacks := []Platform{
		PlatformWeb,
		PlatformYouTube,
		PlatformGitHub,
		PlatformRSS,
		PlatformV2EX,
		PlatformXueqiu,
		PlatformFacebook,
		PlatformInstagram,
	}

	for _, p := range noFallbacks {
		fallbacks := FallbackBackends(p)
		// These may have empty fallbacks, which is fine
		_ = fallbacks
	}
}

func TestCredentialKeys(t *testing.T) {
	tests := []struct {
		platform Platform
		keys     []string
	}{
		{PlatformTwitter, []string{"auth_token", "ct0"}},
		{PlatformReddit, []string{"cookie"}},
		{PlatformXiaoHongShu, []string{"cookie"}},
		{PlatformLinkedIn, []string{"mcp_config"}},
		{PlatformExaSearch, []string{"api_key"}},
		{PlatformXiaoyuzhou, []string{"whisper_key"}},
		{PlatformBossZhipin, []string{"cookie"}},
		{PlatformFacebook, []string{"cookie"}},
		{PlatformInstagram, []string{"cookie"}},
		{PlatformGitHub, []string{"token"}},
		{PlatformWeb, []string{}},
		{PlatformYouTube, []string{}},
	}

	for _, tt := range tests {
		keys := CredentialKeys(tt.platform)
		if len(keys) != len(tt.keys) {
			t.Errorf("CredentialKeys(%s) length: got %d, expected %d", tt.platform, len(keys), len(tt.keys))
			continue
		}
		for i, k := range tt.keys {
			if keys[i] != k {
				t.Errorf("CredentialKeys(%s)[%d] = %s, expected %s", tt.platform, i, keys[i], k)
			}
		}
	}
}

func TestConfigManager(t *testing.T) {
	cm := NewConfigManager("")
	cm.SetKeyringEnabled(false) // Disable keyring for tests

	config, err := cm.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if config == nil {
		t.Fatal("Config is nil")
	}

	// Test platform config
	pc, err := cm.GetPlatformConfig(PlatformWeb)
	if err != nil {
		t.Fatalf("GetPlatformConfig failed: %v", err)
	}

	if pc.Platform != PlatformWeb {
		t.Errorf("Platform mismatch: got %s", pc.Platform)
	}

	// Test setting config
	pc.Enabled = true
	pc.Credentials = map[string]string{"test": "value"}
	if err := cm.SetPlatformConfig(PlatformWeb, pc); err != nil {
		t.Fatalf("SetPlatformConfig failed: %v", err)
	}

	// Reload and verify
	cm2 := NewConfigManager("")
	config2, err := cm2.Load()
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if !config2.Platforms[PlatformWeb].Enabled {
		t.Error("Platform not enabled after save")
	}

	if config2.Platforms[PlatformWeb].Credentials["test"] != "value" {
		t.Error("Credentials not saved")
	}
}

func TestRouterRegistration(t *testing.T) {
	config := DefaultConfig()
	router := NewRouter(config)

	// Register a mock adapter
	mockAdapter := &mockPlatformAdapter{
		name:     PlatformWeb,
		backends: []Backend{"mock-backend"},
	}
	router.RegisterAdapter(mockAdapter)

	// Verify registration
	adapter, ok := router.GetAdapter(PlatformWeb)
	if !ok {
		t.Fatal("Adapter not registered")
	}
	if adapter.Name() != PlatformWeb {
		t.Errorf("Wrong adapter returned: %s", adapter.Name())
	}

	// List platforms
	platforms := router.ListPlatforms()
	found := false
	for _, p := range platforms {
		if p == PlatformWeb {
			found = true
			break
		}
	}
	if !found {
		t.Error("Registered platform not in list")
	}
}

func TestRegistryCreation(t *testing.T) {
	config := DefaultConfig()
	registry := NewRegistry(config)

	if registry == nil {
		t.Fatal("Registry is nil")
	}

	if registry.Router() == nil {
		t.Fatal("Router is nil")
	}

	// Check all platforms have adapters
	for _, p := range AllPlatforms {
		_, ok := registry.Router().GetAdapter(p)
		if !ok {
			t.Errorf("No adapter for platform %s", p)
		}
	}
}

func TestRouterHealth(t *testing.T) {
	config := DefaultConfig()
	router := NewRouter(config)

	// Register a mock adapter that always succeeds
	mockAdapter := &mockPlatformAdapter{
		name:     PlatformWeb,
		backends: []Backend{"mock-backend-1", "mock-backend-2"},
		healthResults: map[Backend]BackendHealth{
			"mock-backend-1": {Backend: "mock-backend-1", Platform: PlatformWeb, Available: true, LastChecked: time.Now()},
			"mock-backend-2": {Backend: "mock-backend-2", Platform: PlatformWeb, Available: false, LastChecked: time.Now(), Error: "not installed"},
		},
	}
	router.RegisterAdapter(mockAdapter)

	ctx := context.Background()
	health, err := router.Health(ctx, PlatformWeb)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if len(health) != 2 {
		t.Errorf("Expected 2 health results, got %d", len(health))
	}

	if !health[0].Available {
		t.Error("First backend should be available")
	}

	if health[1].Available {
		t.Error("Second backend should not be available")
	}
}

func TestRouterRouting(t *testing.T) {
	// Create config with mock backends enabled
	config := DefaultConfig()
	config.Platforms[PlatformWeb] = PlatformConfig{
		Platform:         PlatformWeb,
		Enabled:          true,
		PrimaryBackend:   "failing-backend",
		FallbackBackends: []Backend{"working-backend"},
		Credentials:      make(map[string]string),
		Options:          make(map[string]string),
	}
	router := NewRouter(config)

	// Register mock adapter where first backend fails, second succeeds
	mockAdapter := &mockPlatformAdapter{
		name:     PlatformWeb,
		backends: []Backend{"failing-backend", "working-backend"},
		executeResults: map[Backend]BackendResult{
			"failing-backend": {Backend: "failing-backend", Platform: PlatformWeb, Success: false, Error: "connection refused"},
			"working-backend": {Backend: "working-backend", Platform: PlatformWeb, Success: true, Data: "success data"},
		},
	}
	router.RegisterAdapter(mockAdapter)

	ctx := context.Background()
	result, err := router.Route(ctx, PlatformWeb, "test query", nil)
	if err != nil {
		t.Fatalf("Route failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Routing should succeed with fallback, got: %s", result.Error)
	}

	if result.Backend != "working-backend" {
		t.Errorf("Expected fallback to working-backend, got %s", result.Backend)
	}

	if result.Data != "success data" {
		t.Errorf("Unexpected data: %s", result.Data)
	}
}

func TestRegistryConfigurePlatform(t *testing.T) {
	config := DefaultConfig()
	registry := NewRegistry(config)

	// Configure Twitter with credentials
	err := registry.ConfigurePlatform(PlatformTwitter, map[string]string{
		"auth_token": "test-token",
		"ct0":        "test-ct0",
	})
	if err != nil {
		t.Fatalf("ConfigurePlatform failed: %v", err)
	}

	// Verify credentials were set
	adapter, ok := registry.Router().GetAdapter(PlatformTwitter)
	if !ok {
		t.Fatal("Twitter adapter not found")
	}

	twitterAdapter, ok := adapter.(*TwitterAdapter)
	if !ok {
		t.Fatal("Adapter is not TwitterAdapter")
	}

	if twitterAdapter.authToken != "test-token" {
		t.Errorf("Auth token not set: %s", twitterAdapter.authToken)
	}

	if twitterAdapter.ct0 != "test-ct0" {
		t.Errorf("CT0 not set: %s", twitterAdapter.ct0)
	}
}

// Mock adapter for testing
type mockPlatformAdapter struct {
	name           Platform
	backends       []Backend
	healthResults  map[Backend]BackendHealth
	executeResults map[Backend]BackendResult
}

func (m *mockPlatformAdapter) Name() Platform {
	return m.name
}

func (m *mockPlatformAdapter) Backends() []Backend {
	return m.backends
}

func (m *mockPlatformAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	if result, ok := m.executeResults[backend]; ok {
		return result, nil
	}
	return BackendResult{Backend: backend, Platform: m.name, Success: false, Error: "not configured"}, nil
}

func (m *mockPlatformAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	if health, ok := m.healthResults[backend]; ok {
		return health, nil
	}
	return BackendHealth{Backend: backend, Platform: m.name, Available: false, LastChecked: time.Now()}, nil
}

func (m *mockPlatformAdapter) Configure(creds map[string]string) error {
	return nil
}

func (m *mockPlatformAdapter) RequiresAuth() bool {
	return false
}
