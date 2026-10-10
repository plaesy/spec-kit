// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Registry holds all platform adapters and creates a router.
type Registry struct {
	router      *DefaultRouter
	config      *ReachConfig
	cache       CacheBackend
	rateLimiter *PlatformRateLimiter
	logger      Logger
	mu          sync.RWMutex
	initialized bool
}

// RegistryOption configures the registry.
type RegistryOption func(*Registry)

// WithRegistryCache sets the cache backend for the registry.
func WithRegistryCache(cache CacheBackend) RegistryOption {
	return func(r *Registry) {
		r.cache = cache
	}
}

// WithRegistryRateLimiter sets the rate limiter for the registry.
func WithRegistryRateLimiter(limiter *PlatformRateLimiter) RegistryOption {
	return func(r *Registry) {
		r.rateLimiter = limiter
	}
}

// WithRegistryLogger sets the logger for the registry.
func WithRegistryLogger(logger Logger) RegistryOption {
	return func(r *Registry) {
		r.logger = logger
	}
}

// WithRegistryRetryConfig sets the retry configuration.
func WithRegistryRetryConfig(config RetryConfig) RegistryOption {
	return func(r *Registry) {
		r.router = NewRouter(r.config,
			WithCache(r.cache),
			WithRateLimiter(r.rateLimiter),
			WithLogger(r.logger),
			WithRetryConfig(config),
		)
		r.registerAdapters()
	}
}

// NewRegistry creates a new registry with all built-in adapters.
func NewRegistry(config *ReachConfig, opts ...RegistryOption) *Registry {
	r := &Registry{
		config:      config,
		logger:      DefaultLogger{},
		initialized: false,
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	// Create router with options
	routerOpts := []RouterOption{}
	if r.cache != nil {
		routerOpts = append(routerOpts, WithCache(r.cache))
	}
	if r.rateLimiter != nil {
		routerOpts = append(routerOpts, WithRateLimiter(r.rateLimiter))
	}
	routerOpts = append(routerOpts, WithLogger(r.logger))

	// Add global retry config if present
	if config.GlobalRetryConfig != nil {
		routerOpts = append(routerOpts, WithRetryConfig(*config.GlobalRetryConfig))
	}

	// Add global timeout if present
	if config.DefaultTimeout > 0 {
		routerOpts = append(routerOpts, WithTimeout(config.DefaultTimeout))
	}

	r.router = NewRouter(config, routerOpts...)
	r.registerAdapters()
	r.initialized = true

	return r
}

func (r *Registry) registerAdapters() {
	// Register all built-in adapters
	r.router.RegisterAdapter(NewWebAdapter())
	r.router.RegisterAdapter(NewYouTubeAdapter())
	r.router.RegisterAdapter(NewGitHubAdapter())
	r.router.RegisterAdapter(NewRSSAdapter())
	r.router.RegisterAdapter(NewTwitterAdapter())
	r.router.RegisterAdapter(NewRedditAdapter())
	r.router.RegisterAdapter(NewBilibiliAdapter())
	r.router.RegisterAdapter(NewXiaoHongShuAdapter())
	r.router.RegisterAdapter(NewLinkedInAdapter())
	r.router.RegisterAdapter(NewExaSearchAdapter())
	r.router.RegisterAdapter(NewV2EXAdapter())
	r.router.RegisterAdapter(NewXueqiuAdapter())
	r.router.RegisterAdapter(NewXiaoyuzhouAdapter())
	r.router.RegisterAdapter(NewBossZhipinAdapter())
	r.router.RegisterAdapter(NewFacebookAdapter())
	r.router.RegisterAdapter(NewInstagramAdapter())
}

// Router returns the router instance.
func (r *Registry) Router() Router {
	return r.router
}

// Query executes a query on a platform.
func (r *Registry) Query(ctx context.Context, platform Platform, query string, opts map[string]string) (BackendResult, error) {
	return r.router.Route(ctx, platform, query, opts)
}

// Health checks the health of a platform's backends.
func (r *Registry) Health(ctx context.Context, platform Platform) ([]BackendHealth, error) {
	return r.router.Health(ctx, platform)
}

// HealthAll checks health of all platforms.
func (r *Registry) HealthAll(ctx context.Context) (map[Platform][]BackendHealth, error) {
	result := make(map[Platform][]BackendHealth)
	for _, platform := range r.router.ListPlatforms() {
		health, err := r.router.Health(ctx, platform)
		if err != nil {
			result[platform] = []BackendHealth{{
				Platform:    platform,
				Available:   false,
				LastChecked: time.Now(),
				Error:       err.Error(),
			}}
		} else {
			result[platform] = health
		}
	}
	return result, nil
}

// ConfigurePlatform configures a platform with credentials.
func (r *Registry) ConfigurePlatform(platform Platform, creds map[string]string) error {
	if adapter, ok := r.router.GetAdapter(platform); ok {
		return adapter.Configure(creds)
	}
	return fmt.Errorf("no adapter for platform %s", platform)
}

// LoadConfig loads configuration and applies it to adapters.
func (r *Registry) LoadConfig(ctx context.Context) error {
	cm := NewConfigManager("")
	config, err := cm.Load()
	if err != nil {
		return err
	}

	for platform, pc := range config.Platforms {
		if !pc.Enabled {
			continue
		}
		if len(pc.Credentials) > 0 {
			if err := r.ConfigurePlatform(platform, pc.Credentials); err != nil {
				return fmt.Errorf("configuring platform %s: %w", platform, err)
			}
		}

		// Apply platform-specific settings to router
		if pc.Timeout > 0 {
			// Timeout is handled per-query via options
		}
		if pc.RateLimit != nil && r.rateLimiter != nil {
			r.rateLimiter.SetConfig(platform, *pc.RateLimit)
		}
	}

	return nil
}

// ReloadConfig reloads configuration from disk.
func (r *Registry) ReloadConfig(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cm := NewConfigManager("")
	config, err := cm.Load()
	if err != nil {
		return err
	}

	r.config = config
	r.router.config = config

	// Re-apply platform settings
	for platform, pc := range config.Platforms {
		if pc.RateLimit != nil && r.rateLimiter != nil {
			r.rateLimiter.SetConfig(platform, *pc.RateLimit)
		}
	}

	return nil
}

// GetCache returns the cache backend.
func (r *Registry) GetCache() CacheBackend {
	return r.cache
}

// GetRateLimiter returns the rate limiter.
func (r *Registry) GetRateLimiter() *PlatformRateLimiter {
	return r.rateLimiter
}

// SetLogger updates the logger.
func (r *Registry) SetLogger(logger Logger) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger = logger
	if r.router != nil {
		r.router.SetLogger(logger)
	}
}

// Stats returns registry statistics.
func (r *Registry) Stats() RegistryStats {
	var cacheStats CacheStats
	if r.cache != nil {
		cacheStats = r.cache.Stats()
	}

	return RegistryStats{
		Platforms:   len(r.router.ListPlatforms()),
		CacheStats:  cacheStats,
		Initialized: r.initialized,
	}
}

// RegistryStats holds registry statistics.
type RegistryStats struct {
	Platforms   int
	CacheStats  CacheStats
	Initialized bool
}
