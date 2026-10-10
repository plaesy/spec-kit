// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms (web, social media, code
// hosting, video, RSS, search) with multi-backend routing and fallback.
package reach

import (
	"context"
	"time"
)

// Platform identifies a data source platform.
type Platform string

const (
	PlatformWeb         Platform = "web"
	PlatformYouTube     Platform = "youtube"
	PlatformGitHub      Platform = "github"
	PlatformRSS         Platform = "rss"
	PlatformTwitter     Platform = "twitter"
	PlatformReddit      Platform = "reddit"
	PlatformBilibili    Platform = "bilibili"
	PlatformXiaoHongShu Platform = "xiaohongshu"
	PlatformLinkedIn    Platform = "linkedin"
	PlatformExaSearch   Platform = "exa"
	PlatformV2EX        Platform = "v2ex"
	PlatformXueqiu      Platform = "xueqiu"
	PlatformXiaoyuzhou  Platform = "xiaoyuzhou"
	PlatformBossZhipin  Platform = "bosszhipin"
	PlatformFacebook    Platform = "facebook"
	PlatformInstagram   Platform = "instagram"
)

// AllPlatforms lists all supported platforms.
var AllPlatforms = []Platform{
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

// ZeroConfigPlatforms are platforms that work without authentication.
var ZeroConfigPlatforms = []Platform{
	PlatformWeb,
	PlatformYouTube,
	PlatformGitHub,
	PlatformRSS,
	PlatformBilibili,
	PlatformV2EX,
	PlatformXueqiu,
}

// AuthRequiredPlatforms are platforms that require authentication/configuration.
var AuthRequiredPlatforms = []Platform{
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

// Backend identifies a specific backend implementation for a platform.
type Backend string

// BackendResult represents the result of a backend operation.
type BackendResult struct {
	Backend   Backend
	Platform  Platform
	Success   bool
	Data      string
	Error     string
	Latency   time.Duration
	Timestamp time.Time
}

// BackendHealth represents the health status of a backend.
type BackendHealth struct {
	Backend      Backend
	Platform     Platform
	Available    bool
	LastChecked  time.Time
	Error        string
	Latency      time.Duration
	Version      string
	RequiresAuth bool
}

// PlatformConfig holds configuration for a platform.
type PlatformConfig struct {
	Platform         Platform
	Enabled          bool
	PrimaryBackend   Backend
	FallbackBackends []Backend
	Credentials      map[string]string
	Options          map[string]string
	// Advanced settings
	Timeout     time.Duration    `json:"timeout,omitempty"`
	CacheTTL    time.Duration    `json:"cache_ttl,omitempty"`
	RateLimit   *RateLimitConfig `json:"rate_limit,omitempty"`
	RetryConfig *RetryConfig     `json:"retry_config,omitempty"`
	MaxRetries  int              `json:"max_retries,omitempty"`
}

// RateLimitConfig holds rate limit configuration for a platform.
type RateLimitConfig struct {
	RequestsPerSecond float64       `json:"requests_per_second"`
	Burst             int           `json:"burst"`
	Window            time.Duration `json:"window"`
	Algorithm         string        `json:"algorithm"` // "token-bucket" or "sliding-window"
}

// RetryConfig holds retry configuration.
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	BaseDelay       time.Duration `json:"base_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	Multiplier      float64       `json:"multiplier"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// DefaultRetryConfig returns sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"connection reset",
			"temporary failure",
			"too many requests",
			"rate limit",
			"503",
			"504",
			"502",
		},
	}
}

// ReachConfig holds the complete reach configuration.
type ReachConfig struct {
	Platforms     map[Platform]PlatformConfig
	GlobalOptions map[string]string
	// Global settings
	DefaultTimeout    time.Duration    `json:"default_timeout"`
	DefaultCacheTTL   time.Duration    `json:"default_cache_ttl"`
	GlobalRateLimit   *RateLimitConfig `json:"global_rate_limit,omitempty"`
	GlobalRetryConfig *RetryConfig     `json:"global_retry_config,omitempty"`
	CacheDir          string           `json:"cache_dir,omitempty"`
	CacheSizeMB       int              `json:"cache_size_mb,omitempty"`
	CacheMemorySize   int              `json:"cache_memory_size,omitempty"`
}

// PlatformAdapter defines the interface for platform-specific operations.
type PlatformAdapter interface {
	// Name returns the platform identifier.
	Name() Platform

	// Backends returns the ordered list of backends for this platform
	// (primary first, then fallbacks).
	Backends() []Backend

	// Execute runs a query on the platform using the given backend.
	Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error)

	// Health checks if a backend is available.
	Health(ctx context.Context, backend Backend) (BackendHealth, error)

	// Configure sets up authentication/credentials for the platform.
	Configure(creds map[string]string) error

	// RequiresAuth returns true if the platform needs authentication.
	RequiresAuth() bool
}

// Router routes queries to the best available backend for a platform.
type Router interface {
	// Route executes a query on the platform, trying backends in order
	// until one succeeds.
	Route(ctx context.Context, platform Platform, query string, opts map[string]string) (BackendResult, error)

	// Health checks all backends for a platform.
	Health(ctx context.Context, platform Platform) ([]BackendHealth, error)

	// RegisterAdapter registers a platform adapter.
	RegisterAdapter(adapter PlatformAdapter)

	// GetAdapter returns the adapter for a platform.
	GetAdapter(platform Platform) (PlatformAdapter, bool)

	// ListPlatforms returns all registered platforms.
	ListPlatforms() []Platform
}
