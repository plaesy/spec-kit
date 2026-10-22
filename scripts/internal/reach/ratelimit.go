// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	Allow(ctx context.Context) bool
	Wait(ctx context.Context) error
	Limit() int
	Remaining() int
	Reset() time.Time
}

// TokenBucket is a thread-safe token bucket rate limiter.
type TokenBucket struct {
	mu         sync.Mutex
	capacity   int
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewTokenBucket creates a new token bucket rate limiter.
func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     float64(capacity),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(float64(tb.capacity), tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now
}

func (tb *TokenBucket) Allow(ctx context.Context) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.Allow(ctx) {
			return nil
		}
		// Wait for next token
		tb.mu.Lock()
		waitTime := time.Duration((1-tb.tokens)/tb.refillRate*float64(time.Second)) + time.Millisecond
		tb.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

func (tb *TokenBucket) Limit() int {
	return tb.capacity
}

func (tb *TokenBucket) Remaining() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return int(tb.tokens)
}

func (tb *TokenBucket) Reset() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	if tb.tokens >= float64(tb.capacity) {
		return time.Now()
	}
	return tb.lastRefill.Add(time.Duration((float64(tb.capacity) - tb.tokens) / tb.refillRate * float64(time.Second)))
}

// SlidingWindowLog is a more precise rate limiter using sliding window log.
type SlidingWindowLog struct {
	mu       sync.Mutex
	requests []time.Time
	limit    int
	window   time.Duration
}

// NewSlidingWindowLog creates a sliding window log rate limiter.
func NewSlidingWindowLog(limit int, window time.Duration) *SlidingWindowLog {
	return &SlidingWindowLog{
		limit:    limit,
		window:   window,
		requests: make([]time.Time, 0, limit),
	}
}

func (swl *SlidingWindowLog) Allow(ctx context.Context) bool {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-swl.window)

	// Remove old entries
	valid := swl.requests[:0]
	for _, t := range swl.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	swl.requests = valid

	if len(swl.requests) < swl.limit {
		swl.requests = append(swl.requests, now)
		return true
	}
	return false
}

func (swl *SlidingWindowLog) Wait(ctx context.Context) error {
	for {
		if swl.Allow(ctx) {
			return nil
		}
		swl.mu.Lock()
		var waitTime time.Duration
		if len(swl.requests) > 0 {
			waitTime = swl.requests[0].Add(swl.window).Sub(time.Now()) + time.Millisecond
		} else {
			waitTime = time.Millisecond
		}
		swl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}
}

func (swl *SlidingWindowLog) Limit() int {
	return swl.limit
}

func (swl *SlidingWindowLog) Remaining() int {
	swl.mu.Lock()
	defer swl.mu.Unlock()
	swl.cleanup()
	return max(0, swl.limit-len(swl.requests))
}

func (swl *SlidingWindowLog) Reset() time.Time {
	swl.mu.Lock()
	defer swl.mu.Unlock()
	swl.cleanup()
	if len(swl.requests) == 0 {
		return time.Now()
	}
	return swl.requests[0].Add(swl.window)
}

func (swl *SlidingWindowLog) cleanup() {
	now := time.Now()
	cutoff := now.Add(-swl.window)
	valid := swl.requests[:0]
	for _, t := range swl.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	swl.requests = valid
}

// MultiRateLimiter combines multiple rate limiters (AND logic - all must allow).
type MultiRateLimiter struct {
	limiters []RateLimiter
}

// NewMultiRateLimiter creates a multi rate limiter.
func NewMultiRateLimiter(limiters ...RateLimiter) *MultiRateLimiter {
	return &MultiRateLimiter{limiters: limiters}
}

func (m *MultiRateLimiter) Allow(ctx context.Context) bool {
	for _, l := range m.limiters {
		if !l.Allow(ctx) {
			return false
		}
	}
	return true
}

func (m *MultiRateLimiter) Wait(ctx context.Context) error {
	for _, l := range m.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiRateLimiter) Limit() int {
	minLimit := -1
	for _, l := range m.limiters {
		if minLimit == -1 || l.Limit() < minLimit {
			minLimit = l.Limit()
		}
	}
	return minLimit
}

func (m *MultiRateLimiter) Remaining() int {
	minRem := -1
	for _, l := range m.limiters {
		rem := l.Remaining()
		if minRem == -1 || rem < minRem {
			minRem = rem
		}
	}
	return minRem
}

func (m *MultiRateLimiter) Reset() time.Time {
	var latest time.Time
	for _, l := range m.limiters {
		if r := l.Reset(); r.After(latest) {
			latest = r
		}
	}
	return latest
}

// PlatformRateLimiter manages per-platform rate limiters.
type PlatformRateLimiter struct {
	mu         sync.RWMutex
	limiters   map[Platform]RateLimiter
	defaultLim map[Platform]RateLimitConfig
}

// DefaultRateLimits returns default rate limits per platform.
func DefaultRateLimits() map[Platform]RateLimitConfig {
	return map[Platform]RateLimitConfig{
		PlatformWeb:         {RequestsPerSecond: 10, Burst: 20, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformYouTube:     {RequestsPerSecond: 5, Burst: 10, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformGitHub:      {RequestsPerSecond: 50, Burst: 100, Window: time.Hour, Algorithm: "sliding-window"},
		PlatformRSS:         {RequestsPerSecond: 2, Burst: 5, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformTwitter:     {RequestsPerSecond: 1, Burst: 3, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformReddit:      {RequestsPerSecond: 2, Burst: 5, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformBilibili:    {RequestsPerSecond: 5, Burst: 10, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformXiaoHongShu: {RequestsPerSecond: 1, Burst: 2, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformLinkedIn:    {RequestsPerSecond: 2, Burst: 5, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformExaSearch:   {RequestsPerSecond: 5, Burst: 10, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformV2EX:        {RequestsPerSecond: 10, Burst: 20, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformXueqiu:      {RequestsPerSecond: 10, Burst: 20, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformXiaoyuzhou:  {RequestsPerSecond: 2, Burst: 5, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformBossZhipin:  {RequestsPerSecond: 1, Burst: 3, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformFacebook:    {RequestsPerSecond: 1, Burst: 2, Window: time.Minute, Algorithm: "token-bucket"},
		PlatformInstagram:   {RequestsPerSecond: 1, Burst: 2, Window: time.Minute, Algorithm: "token-bucket"},
	}
}

// NewPlatformRateLimiter creates a new platform rate limiter manager.
func NewPlatformRateLimiter(configs map[Platform]RateLimitConfig) *PlatformRateLimiter {
	prl := &PlatformRateLimiter{
		limiters:   make(map[Platform]RateLimiter),
		defaultLim: configs,
	}
	for platform, config := range configs {
		prl.limiters[platform] = prl.createLimiter(config)
	}
	return prl
}

func (p *PlatformRateLimiter) createLimiter(config RateLimitConfig) RateLimiter {
	switch config.Algorithm {
	case "sliding-window":
		return NewSlidingWindowLog(config.Burst, config.Window)
	default:
		return NewTokenBucket(config.Burst, config.RequestsPerSecond)
	}
}

// Get returns the rate limiter for a platform.
func (p *PlatformRateLimiter) Get(platform Platform) RateLimiter {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if l, ok := p.limiters[platform]; ok {
		return l
	}
	// Return default if not configured
	if config, ok := p.defaultLim[platform]; ok {
		return p.createLimiter(config)
	}
	// Unlimited
	return NewTokenBucket(1000, 1000)
}

// SetConfig updates the rate limit config for a platform.
func (p *PlatformRateLimiter) SetConfig(platform Platform, config RateLimitConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.limiters[platform] = p.createLimiter(config)
	p.defaultLim[platform] = config
}

// Allow checks if a request is allowed for the platform.
func (p *PlatformRateLimiter) Allow(ctx context.Context, platform Platform) bool {
	return p.Get(platform).Allow(ctx)
}

// Wait waits until a request is allowed for the platform.
func (p *PlatformRateLimiter) Wait(ctx context.Context, platform Platform) error {
	return p.Get(platform).Wait(ctx)
}
