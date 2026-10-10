// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// isRetryable checks if an error is retryable.
func isRetryable(err error, config RetryConfig) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	for _, pattern := range config.RetryableErrors {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	subLower := toLower(substr)
	return len(sLower) >= len(subLower) && (sLower == subLower || len(sLower) > len(subLower) && (sLower[:len(subLower)] == subLower || sLower[len(sLower)-len(subLower):] == subLower || containsSubstring(sLower, subLower)))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			b[i] = byte(c - 'A' + 'a')
		} else {
			b[i] = byte(c)
		}
	}
	return string(b)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// calculateBackoff calculates exponential backoff with jitter.
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	// Add jitter (±25%)
	jitter := delay * 0.25 * (2*float64(time.Now().UnixNano()%1000)/1000 - 1)
	return time.Duration(delay + jitter)
}

// RouterOption configures the router.
type RouterOption func(*DefaultRouter)

// WithCache sets the cache backend.
func WithCache(cache CacheBackend) RouterOption {
	return func(r *DefaultRouter) {
		r.cache = cache
	}
}

// WithRateLimiter sets the rate limiter.
func WithRateLimiter(limiter *PlatformRateLimiter) RouterOption {
	return func(r *DefaultRouter) {
		r.rateLimiter = limiter
	}
}

// WithRetryConfig sets the retry configuration.
func WithRetryConfig(config RetryConfig) RouterOption {
	return func(r *DefaultRouter) {
		r.retryConfig = config
	}
}

// WithTimeout sets the default timeout.
func WithTimeout(timeout time.Duration) RouterOption {
	return func(r *DefaultRouter) {
		r.defaultTimeout = timeout
	}
}

// WithLogger sets the logger.
func WithLogger(logger Logger) RouterOption {
	return func(r *DefaultRouter) {
		r.logger = logger
	}
}

// Logger defines the logging interface.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// DefaultLogger is a no-op logger.
type DefaultLogger struct{}

func (DefaultLogger) Debug(msg string, args ...any) {}
func (DefaultLogger) Info(msg string, args ...any)  {}
func (DefaultLogger) Warn(msg string, args ...any)  {}
func (DefaultLogger) Error(msg string, args ...any) {}

// DefaultRouter implements the Router interface with multi-backend routing.
type DefaultRouter struct {
	adapters       map[Platform]PlatformAdapter
	mu             sync.RWMutex
	config         *ReachConfig
	cache          CacheBackend
	rateLimiter    *PlatformRateLimiter
	retryConfig    RetryConfig
	defaultTimeout time.Duration
	logger         Logger
}

// NewRouter creates a new default router with options.
func NewRouter(config *ReachConfig, opts ...RouterOption) *DefaultRouter {
	r := &DefaultRouter{
		adapters:       make(map[Platform]PlatformAdapter),
		config:         config,
		retryConfig:    DefaultRetryConfig(),
		defaultTimeout: 30 * time.Second,
		logger:         DefaultLogger{},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// RegisterAdapter registers a platform adapter.
func (r *DefaultRouter) RegisterAdapter(adapter PlatformAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// Route executes a query on the platform, trying backends in order until one succeeds.
func (r *DefaultRouter) Route(ctx context.Context, platform Platform, query string, opts map[string]string) (BackendResult, error) {
	r.mu.RLock()
	adapter, ok := r.adapters[platform]
	r.mu.RUnlock()

	if !ok {
		return BackendResult{}, fmt.Errorf("no adapter registered for platform %s", platform)
	}

	backends := adapter.Backends()
	if len(backends) == 0 {
		return BackendResult{}, fmt.Errorf("no backends configured for platform %s", platform)
	}

	// Apply timeout from options or default
	timeout := r.defaultTimeout
	if toStr, ok := opts["timeout"]; ok {
		if d, err := time.ParseDuration(toStr); err == nil {
			timeout = d
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Rate limiting
	if r.rateLimiter != nil {
		if err := r.rateLimiter.Wait(ctx, platform); err != nil {
			return BackendResult{}, fmt.Errorf("rate limited: %w", err)
		}
	}

	// Check cache first
	cacheKey := ""
	if r.cache != nil && isCacheable(opts) {
		cacheKey = CacheKey(platform, "", query, opts)
		if cached, ok := r.cache.Get(ctx, cacheKey); ok {
			r.logger.Debug("Cache hit", "platform", platform, "query", query)
			return cached, nil
		}
		r.logger.Debug("Cache miss", "platform", platform, "query", query)
	}

	var lastErr error
	for _, backend := range backends {
		// Check if backend is enabled in config
		if r.config != nil {
			if pc, ok := r.config.Platforms[platform]; ok {
				enabled := false
				if pc.PrimaryBackend == backend {
					enabled = true
				} else {
					for _, fb := range pc.FallbackBackends {
						if fb == backend {
							enabled = true
							break
						}
					}
				}
				if !enabled {
					continue
				}
			}
		}

		// Execute with retries
		result, err := r.executeWithRetry(ctx, adapter, backend, query, opts)
		if err == nil && result.Success {
			// Cache successful result
			if r.cache != nil && cacheKey != "" && isCacheable(opts) {
				ttl := r.getCacheTTL(opts)
				r.cache.Set(ctx, cacheKey, result, ttl)
			}
			return result, nil
		}

		lastErr = err
		r.logger.Debug("Backend failed, trying next", "backend", backend, "error", err)
		// Continue to next backend on failure
	}

	return BackendResult{}, fmt.Errorf("all backends failed for platform %s: last error: %w", platform, lastErr)
}

func (r *DefaultRouter) executeWithRetry(ctx context.Context, adapter PlatformAdapter, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	var lastErr error
	var lastResult BackendResult

	for attempt := 0; attempt <= r.retryConfig.MaxRetries; attempt++ {
		result, err := adapter.Execute(ctx, backend, query, opts)
		if err == nil && result.Success {
			return result, nil
		}

		lastErr = err
		lastResult = result

		// Don't retry on last attempt
		if attempt == r.retryConfig.MaxRetries {
			break
		}

		// Check if error is retryable
		if !isRetryable(err, r.retryConfig) && result.Error != "" && !isRetryable(errors.New(result.Error), r.retryConfig) {
			break
		}

		// Wait before retry
		backoff := calculateBackoff(attempt, r.retryConfig)
		r.logger.Debug("Retrying", "attempt", attempt+1, "backend", backend, "backoff", backoff)

		select {
		case <-ctx.Done():
			return BackendResult{}, ctx.Err()
		case <-time.After(backoff):
		}
	}

	if lastErr != nil {
		return BackendResult{}, lastErr
	}
	return lastResult, fmt.Errorf("max retries exceeded")
}

func isCacheable(opts map[string]string) bool {
	// Don't cache if explicitly disabled
	if v, ok := opts["cache"]; ok && v == "false" {
		return false
	}
	// Don't cache mutations
	if action, ok := opts["action"]; ok {
		switch action {
		case "create", "update", "delete", "post", "comment", "vote":
			return false
		}
	}
	return true
}

func (r *DefaultRouter) getCacheTTL(opts map[string]string) time.Duration {
	if ttlStr, ok := opts["cache_ttl"]; ok {
		if d, err := time.ParseDuration(ttlStr); err == nil {
			return d
		}
	}
	// Default TTL based on platform
	return 5 * time.Minute
}

// Health checks all backends for a platform.
func (r *DefaultRouter) Health(ctx context.Context, platform Platform) ([]BackendHealth, error) {
	r.mu.RLock()
	adapter, ok := r.adapters[platform]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no adapter registered for platform %s", platform)
	}

	var healthResults []BackendHealth
	for _, backend := range adapter.Backends() {
		health, err := adapter.Health(ctx, backend)
		if err != nil {
			healthResults = append(healthResults, BackendHealth{
				Backend:     backend,
				Platform:    platform,
				Available:   false,
				LastChecked: time.Now(),
				Error:       err.Error(),
			})
		} else {
			healthResults = append(healthResults, health)
		}
	}
	return healthResults, nil
}

// GetAdapter returns the adapter for a platform.
func (r *DefaultRouter) GetAdapter(platform Platform) (PlatformAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[platform]
	return adapter, ok
}

// ListPlatforms returns all registered platforms.
func (r *DefaultRouter) ListPlatforms() []Platform {
	r.mu.RLock()
	defer r.mu.RUnlock()
	platforms := make([]Platform, 0, len(r.adapters))
	for p := range r.adapters {
		platforms = append(platforms, p)
	}
	return platforms
}

// SetLogger updates the logger.
func (r *DefaultRouter) SetLogger(logger Logger) {
	r.logger = logger
}

// SetCache updates the cache backend.
func (r *DefaultRouter) SetCache(cache CacheBackend) {
	r.cache = cache
}

// SetRateLimiter updates the rate limiter.
func (r *DefaultRouter) SetRateLimiter(limiter *PlatformRateLimiter) {
	r.rateLimiter = limiter
}
