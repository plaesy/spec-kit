// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/zalando/go-keyring"
)

// DefaultConfigPath is the default location for reach configuration (legacy).
var DefaultConfigPath = filepath.Join(".plaesy", "reach", "config.json")

// UnifiedConfigPath is the new unified config location.
var UnifiedConfigPath = config.UnifiedConfigPath()

// ConfigManager handles loading and saving reach configuration.
type ConfigManager struct {
	path       string
	mu         sync.Mutex
	cache      *ReachConfig
	useKeyring bool
	keyringSvc string
	unifiedMgr *config.ConfigManager[config.UnifiedConfig]
}

// NewConfigManager creates a new config manager.
func NewConfigManager(path string) *ConfigManager {
	if path == "" {
		path = UnifiedConfigPath
	}
	unifiedMgr := config.NewConfigManager[config.UnifiedConfig](
		UnifiedConfigPath,
		config.DefaultUnifiedConfig(),
	)
	return &ConfigManager{
		path:       path,
		useKeyring: true,
		keyringSvc: "plaesy-reach",
		unifiedMgr: unifiedMgr,
	}
}

// MigrateFromLegacy migrates from the old reach config location to unified config.
func (cm *ConfigManager) MigrateFromLegacy() error {
	legacyPath := filepath.Join(".plaesy", "reach", "config.json")
	return cm.unifiedMgr.MigrateFrom(legacyPath, func(data []byte) (*config.UnifiedConfig, error) {
		var legacy ReachConfig
		if err := json.Unmarshal(data, &legacy); err != nil {
			return nil, err
		}
		unified := config.DefaultUnifiedConfig()
		unified.Reach = &config.ReachConfig{
			CacheDir:        legacy.CacheDir,
			CacheSizeMB:     legacy.CacheSizeMB,
			CacheMemorySize: legacy.CacheMemorySize,
			DefaultCacheTTL: legacy.DefaultCacheTTL.String(),
			Platforms:       make(map[string]config.ReachPlatformConfig),
			RateLimits:      make(map[string]config.RateLimitConfig),
			Timeouts:        make(map[string]string),
			RetryConfig:     config.DefaultRetryConfig(),
		}
		for platform, pc := range legacy.Platforms {
			unified.Reach.Platforms[string(platform)] = config.ReachPlatformConfig{
				Enabled:        pc.Enabled,
				Backends:       []string{string(pc.PrimaryBackend)},
				DefaultBackend: string(pc.PrimaryBackend),
				Credentials:    pc.Credentials,
				Options:        pc.Options,
			}
		}
		if legacy.GlobalRateLimit != nil {
			unified.Reach.RateLimits["global"] = config.RateLimitConfig{
				Requests: int(legacy.GlobalRateLimit.RequestsPerSecond * 60),
				Window:   legacy.GlobalRateLimit.Window.String(),
				Burst:    legacy.GlobalRateLimit.Burst,
			}
		}
		return unified, nil
	})
}

// SetKeyringEnabled enables or disables keyring integration.
func (cm *ConfigManager) SetKeyringEnabled(enabled bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.useKeyring = enabled
}

// Load loads the configuration from disk.
func (cm *ConfigManager) Load() (*ReachConfig, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cache != nil {
		return cm.cache, nil
	}

	// Load from unified config
	unified, err := cm.unifiedMgr.Load()
	if err != nil {
		return nil, err
	}

	// Check if unified config has reach section with platforms
	hasReachConfig := unified.Reach != nil && len(unified.Reach.Platforms) > 0

	// Try legacy migration if unified config has no reach section or empty platforms
	if !hasReachConfig {
		if err := cm.MigrateFromLegacy(); err != nil {
			return nil, err
		}
		unified, err = cm.unifiedMgr.Load()
		if err != nil {
			return nil, err
		}
		hasReachConfig = unified.Reach != nil && len(unified.Reach.Platforms) > 0
	}

	// Convert unified reach config to legacy ReachConfig
	if hasReachConfig {
		cm.cache = cm.convertFromUnified(unified.Reach)
	} else {
		cm.cache = DefaultConfig()
	}

	// Decrypt credentials from keyring if enabled
	if cm.useKeyring {
		for platform, pc := range cm.cache.Platforms {
			if len(pc.Credentials) > 0 {
				decrypted := make(map[string]string)
				for key, encrypted := range pc.Credentials {
					if val, err := cm.getFromKeyring(platform, key); err == nil {
						decrypted[key] = val
					} else {
						decrypted[key] = encrypted // Fallback to stored value
					}
				}
				pc.Credentials = decrypted
				cm.cache.Platforms[platform] = pc
			}
		}
	}

	return cm.cache, nil
}

// convertFromUnified converts unified ReachConfig to legacy ReachConfig.
func (cm *ConfigManager) convertFromUnified(u *config.ReachConfig) *ReachConfig {
	platforms := make(map[Platform]PlatformConfig)
	for platformStr, pc := range u.Platforms {
		platform := Platform(platformStr)
		primaryBackend := pc.DefaultBackend
		if primaryBackend == "" && len(pc.Backends) > 0 {
			primaryBackend = pc.Backends[0]
		}
		platforms[platform] = PlatformConfig{
			Platform:       platform,
			Enabled:        pc.Enabled,
			PrimaryBackend: Backend(primaryBackend),
			FallbackBackends: func() []Backend {
				var fb []Backend
				for _, b := range pc.Backends {
					if b != primaryBackend {
						fb = append(fb, Backend(b))
					}
				}
				return fb
			}(),
			Credentials: pc.Credentials,
			Options:     pc.Options,
			Timeout:     parseDurationOrDefault(u.Timeouts[platformStr], 30*time.Second),
			CacheTTL:    parseDurationOrDefault(u.DefaultCacheTTL, 5*time.Minute),
			RateLimit:   rateLimitPtrFromUnified(u.RateLimits[platformStr]),
			RetryConfig: retryConfigFromUnified(u.RetryConfig),
			MaxRetries:  3,
		}
	}
	return &ReachConfig{
		Platforms:         platforms,
		GlobalOptions:     make(map[string]string),
		DefaultTimeout:    30 * time.Second,
		DefaultCacheTTL:   parseDurationOrDefault(u.DefaultCacheTTL, 5*time.Minute),
		CacheDir:          u.CacheDir,
		CacheSizeMB:       u.CacheSizeMB,
		CacheMemorySize:   u.CacheMemorySize,
		GlobalRateLimit:   rateLimitPtrFromUnified(u.RateLimits["global"]),
		GlobalRetryConfig: retryConfigFromUnified(u.RetryConfig),
	}
}

// parseDurationOrDefault parses a duration string or returns default.
func parseDurationOrDefault(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}

// rateLimitPtrFromUnified converts unified RateLimitConfig to legacy pointer.
func rateLimitPtrFromUnified(rl config.RateLimitConfig) *RateLimitConfig {
	if rl.Requests == 0 && rl.Window == "" {
		return nil
	}
	return &RateLimitConfig{
		RequestsPerSecond: float64(rl.Requests) / 60.0, // Convert requests per minute to per second
		Window:            parseDurationOrDefault(rl.Window, time.Minute),
		Burst:             rl.Burst,
		Algorithm:         "token-bucket",
	}
}

// retryConfigFromUnified converts unified RetryConfig to legacy RetryConfig.
func retryConfigFromUnified(rc *config.RetryConfig) *RetryConfig {
	if rc == nil {
		return nil
	}
	return &RetryConfig{
		MaxRetries:      rc.MaxRetries,
		BaseDelay:       parseDurationOrDefault(rc.BaseDelay, 100*time.Millisecond),
		MaxDelay:        parseDurationOrDefault(rc.MaxDelay, 5*time.Second),
		Multiplier:      2.0,
		RetryableErrors: rc.RetryableErrors,
	}
}

// Save saves the configuration to disk.
func (cm *ConfigManager) Save(config *ReachConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Encrypt credentials to keyring if enabled
	if cm.useKeyring {
		for platform, pc := range config.Platforms {
			if len(pc.Credentials) > 0 {
				for key, value := range pc.Credentials {
					if err := cm.saveToKeyring(platform, key, value); err != nil {
						// Log error but continue
						fmt.Fprintf(os.Stderr, "Warning: failed to save credential to keyring: %v\n", err)
					}
				}
				// Store placeholder in config
				for key := range pc.Credentials {
					pc.Credentials[key] = "[keyring]"
				}
				config.Platforms[platform] = pc
			}
		}
	}

	// Convert legacy config to unified and save
	unified, err := cm.unifiedMgr.Load()
	if err != nil {
		return err
	}

	unified.Reach = cm.convertToUnified(config)
	return cm.unifiedMgr.Save(unified)
}

// convertToUnified converts legacy ReachConfig to unified ReachConfig.
func (cm *ConfigManager) convertToUnified(cfg *ReachConfig) *config.ReachConfig {
	platforms := make(map[string]config.ReachPlatformConfig)
	rateLimits := make(map[string]config.RateLimitConfig)
	timeouts := make(map[string]string)

	for platform, pc := range cfg.Platforms {
		platforms[string(platform)] = config.ReachPlatformConfig{
			Enabled:        pc.Enabled,
			Backends:       append([]string{string(pc.PrimaryBackend)}, pc.FallbackBackendsStr()...),
			DefaultBackend: string(pc.PrimaryBackend),
			Credentials:    pc.Credentials,
			Options:        pc.Options,
		}
		if pc.RateLimit != nil {
			rateLimits[string(platform)] = config.RateLimitConfig{
				Requests: int(pc.RateLimit.RequestsPerSecond * 60), // Convert per second to per minute
				Window:   pc.RateLimit.Window.String(),
				Burst:    pc.RateLimit.Burst,
			}
		}
		timeouts[string(platform)] = pc.Timeout.String()
	}

	if cfg.GlobalRateLimit != nil {
		rateLimits["global"] = config.RateLimitConfig{
			Requests: int(cfg.GlobalRateLimit.RequestsPerSecond * 60),
			Window:   cfg.GlobalRateLimit.Window.String(),
			Burst:    cfg.GlobalRateLimit.Burst,
		}
	}

	return &config.ReachConfig{
		CacheDir:        cfg.CacheDir,
		CacheSizeMB:     cfg.CacheSizeMB,
		CacheMemorySize: cfg.CacheMemorySize,
		DefaultCacheTTL: cfg.DefaultCacheTTL.String(),
		Platforms:       platforms,
		RateLimits:      rateLimits,
		Timeouts:        timeouts,
		RetryConfig:     retryConfigToUnified(cfg.GlobalRetryConfig),
	}
}

// retryConfigToUnified converts legacy RetryConfig to unified RetryConfig.
func retryConfigToUnified(rc *RetryConfig) *config.RetryConfig {
	if rc == nil {
		return nil
	}
	return &config.RetryConfig{
		MaxRetries:      rc.MaxRetries,
		BaseDelay:       rc.BaseDelay.String(),
		MaxDelay:        rc.MaxDelay.String(),
		RetryableErrors: rc.RetryableErrors,
	}
}

// FallbackBackendsStr returns fallback backends as string slice.
func (pc PlatformConfig) FallbackBackendsStr() []string {
	var fb []string
	for _, b := range pc.FallbackBackends {
		fb = append(fb, string(b))
	}
	return fb
}

func (cm *ConfigManager) keyringKey(platform Platform, key string) string {
	return fmt.Sprintf("%s/%s/%s", cm.keyringSvc, platform, key)
}

func (cm *ConfigManager) saveToKeyring(platform Platform, key, value string) error {
	if !cm.useKeyring {
		return nil
	}
	return keyring.Set(cm.keyringSvc, cm.keyringKey(platform, key), value)
}

func (cm *ConfigManager) getFromKeyring(platform Platform, key string) (string, error) {
	if !cm.useKeyring {
		return "", fmt.Errorf("keyring disabled")
	}
	return keyring.Get(cm.keyringSvc, cm.keyringKey(platform, key))
}

func (cm *ConfigManager) deleteFromKeyring(platform Platform, key string) error {
	if !cm.useKeyring {
		return nil
	}
	return keyring.Delete(cm.keyringSvc, cm.keyringKey(platform, key))
}

// GetPlatformConfig returns the config for a specific platform.
func (cm *ConfigManager) GetPlatformConfig(platform Platform) (PlatformConfig, error) {
	config, err := cm.Load()
	if err != nil {
		return PlatformConfig{}, err
	}

	if pc, ok := config.Platforms[platform]; ok {
		return pc, nil
	}

	// Return default config for platform
	return DefaultPlatformConfig(platform), nil
}

// SetPlatformConfig sets the config for a specific platform.
func (cm *ConfigManager) SetPlatformConfig(platform Platform, pc PlatformConfig) error {
	config, err := cm.Load()
	if err != nil {
		return err
	}

	if config.Platforms == nil {
		config.Platforms = make(map[Platform]PlatformConfig)
	}
	config.Platforms[platform] = pc

	return cm.Save(config)
}

// EnablePlatform enables a platform.
func (cm *ConfigManager) EnablePlatform(platform Platform) error {
	pc, err := cm.GetPlatformConfig(platform)
	if err != nil {
		return err
	}
	pc.Enabled = true
	return cm.SetPlatformConfig(platform, pc)
}

// DisablePlatform disables a platform.
func (cm *ConfigManager) DisablePlatform(platform Platform) error {
	pc, err := cm.GetPlatformConfig(platform)
	if err != nil {
		return err
	}
	pc.Enabled = false
	return cm.SetPlatformConfig(platform, pc)
}

// SetCredentials sets credentials for a platform.
func (cm *ConfigManager) SetCredentials(platform Platform, creds map[string]string) error {
	pc, err := cm.GetPlatformConfig(platform)
	if err != nil {
		return err
	}
	if pc.Credentials == nil {
		pc.Credentials = make(map[string]string)
	}
	for k, v := range creds {
		pc.Credentials[k] = v
	}
	return cm.SetPlatformConfig(platform, pc)
}

// GetCredential retrieves a single credential from keyring.
func (cm *ConfigManager) GetCredential(platform Platform, key string) (string, error) {
	if cm.useKeyring {
		if val, err := cm.getFromKeyring(platform, key); err == nil {
			return val, nil
		}
	}
	// Fallback to config
	config, err := cm.Load()
	if err != nil {
		return "", err
	}
	if pc, ok := config.Platforms[platform]; ok {
		if val, ok := pc.Credentials[key]; ok && val != "[keyring]" {
			return val, nil
		}
	}
	return "", fmt.Errorf("credential not found")
}

// DeleteCredential removes a credential.
func (cm *ConfigManager) DeleteCredential(platform Platform, key string) error {
	if cm.useKeyring {
		cm.deleteFromKeyring(platform, key)
	}
	config, err := cm.Load()
	if err != nil {
		return err
	}
	if pc, ok := config.Platforms[platform]; ok {
		delete(pc.Credentials, key)
		return cm.SetPlatformConfig(platform, pc)
	}
	return nil
}

// DefaultConfig returns the default reach configuration.
func DefaultConfig() *ReachConfig {
	platforms := make(map[Platform]PlatformConfig)
	for _, p := range AllPlatforms {
		platforms[p] = DefaultPlatformConfig(p)
	}
	return &ReachConfig{
		Platforms:         platforms,
		GlobalOptions:     make(map[string]string),
		DefaultTimeout:    30 * time.Second,
		DefaultCacheTTL:   5 * time.Minute,
		CacheDir:          filepath.Join(".plaesy", "reach", "cache"),
		CacheSizeMB:       100,
		CacheMemorySize:   1000,
		GlobalRateLimit:   nil,
		GlobalRetryConfig: nil,
	}
}

// DefaultPlatformConfig returns the default configuration for a platform.
func DefaultPlatformConfig(platform Platform) PlatformConfig {
	enabled := false
	for _, p := range ZeroConfigPlatforms {
		if p == platform {
			enabled = true
			break
		}
	}

	return PlatformConfig{
		Platform:         platform,
		Enabled:          enabled,
		PrimaryBackend:   DefaultBackend(platform),
		FallbackBackends: FallbackBackends(platform),
		Credentials:      make(map[string]string),
		Options:          make(map[string]string),
		Timeout:          30 * time.Second,
		CacheTTL:         5 * time.Minute,
		RateLimit:        rateLimitPtr(DefaultRateLimits()[platform]),
		RetryConfig:      DefaultRetryConfigPtr(),
		MaxRetries:       3,
	}
}

// DefaultRetryConfigPtr returns a pointer to default retry config.
func DefaultRetryConfigPtr() *RetryConfig {
	cfg := DefaultRetryConfig()
	return &cfg
}

// rateLimitPtr returns a pointer to a RateLimitConfig.
func rateLimitPtr(cfg RateLimitConfig) *RateLimitConfig {
	return &cfg
}

// DefaultBackend returns the default primary backend for a platform.
func DefaultBackend(platform Platform) Backend {
	switch platform {
	case PlatformWeb:
		return "jina-reader"
	case PlatformYouTube:
		return "yt-dlp"
	case PlatformGitHub:
		return "gh-cli"
	case PlatformRSS:
		return "builtin"
	case PlatformTwitter:
		return "twitter-cli"
	case PlatformReddit:
		return "opencli"
	case PlatformBilibili:
		return "bili-cli"
	case PlatformXiaoHongShu:
		return "opencli"
	case PlatformLinkedIn:
		return "mcp-linkedin"
	case PlatformExaSearch:
		return "exa-mcp"
	case PlatformV2EX:
		return "v2ex-api"
	case PlatformXueqiu:
		return "xueqiu-api"
	case PlatformXiaoyuzhou:
		return "xiaoyuzhou-mcp"
	case PlatformBossZhipin:
		return "boss-cdp"
	case PlatformFacebook:
		return "opencli"
	case PlatformInstagram:
		return "opencli"
	default:
		return "unknown"
	}
}

// FallbackBackends returns the fallback backends for a platform.
func FallbackBackends(platform Platform) []Backend {
	switch platform {
	case PlatformTwitter:
		return []Backend{"opencli", "bird"}
	case PlatformReddit:
		return []Backend{"rdt-cli"}
	case PlatformBilibili:
		return []Backend{"opencli", "bili-api"}
	case PlatformXiaoHongShu:
		return []Backend{"xiaohongshu-mcp", "xhs-cli"}
	case PlatformLinkedIn:
		return []Backend{"jina-reader"}
	case PlatformRSS:
		return []Backend{"feedparser"}
	case PlatformFacebook:
		return []Backend{}
	case PlatformInstagram:
		return []Backend{}
	default:
		return []Backend{}
	}
}

// CredentialKeys returns the required credential keys for a platform.
func CredentialKeys(platform Platform) []string {
	switch platform {
	case PlatformTwitter:
		return []string{"auth_token", "ct0"}
	case PlatformReddit:
		return []string{"cookie"}
	case PlatformXiaoHongShu:
		return []string{"cookie"}
	case PlatformLinkedIn:
		return []string{"mcp_config"}
	case PlatformExaSearch:
		return []string{"api_key"}
	case PlatformXiaoyuzhou:
		return []string{"whisper_key"}
	case PlatformBossZhipin:
		return []string{"cookie"}
	case PlatformFacebook:
		return []string{"cookie"}
	case PlatformInstagram:
		return []string{"cookie"}
	case PlatformGitHub:
		return []string{"token"}
	default:
		return []string{}
	}
}
