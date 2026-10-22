package reach

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_BurstThenRefuses(t *testing.T) {
	ctx := context.Background()
	tb := NewTokenBucket(3, 0.001) // a burst of 3, refilling far too slowly to matter

	for i := 0; i < 3; i++ {
		if !tb.Allow(ctx) {
			t.Fatalf("request %d of the burst was refused", i+1)
		}
	}
	if tb.Allow(ctx) {
		t.Error("a request past the burst must be refused")
	}
	if got := tb.Remaining(); got != 0 {
		t.Errorf("Remaining = %d after the burst, want 0", got)
	}
	if got := tb.Limit(); got != 3 {
		t.Errorf("Limit = %d, want 3", got)
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	ctx := context.Background()
	tb := NewTokenBucket(1, 500) // one token every 2ms
	tb.Allow(ctx)
	if tb.Allow(ctx) {
		t.Fatal("the bucket should be empty right after its token is used")
	}
	time.Sleep(10 * time.Millisecond)
	if !tb.Allow(ctx) {
		t.Error("the bucket should have refilled after 10ms at 500/s")
	}
}

func TestTokenBucket_RefillIsCappedAtCapacity(t *testing.T) {
	tb := NewTokenBucket(2, 1000)
	time.Sleep(10 * time.Millisecond)
	if got := tb.Remaining(); got != 2 {
		t.Errorf("Remaining = %d after a long idle, want the capacity 2", got)
	}
}

func TestTokenBucket_ResetReportsWhenTheNextTokenComes(t *testing.T) {
	ctx := context.Background()
	full := NewTokenBucket(1, 1)
	if r := full.Reset(); time.Until(r) > time.Second {
		t.Errorf("a full bucket should reset now, got %v away", time.Until(r))
	}

	empty := NewTokenBucket(1, 1)
	empty.Allow(ctx)
	if r := empty.Reset(); time.Until(r) <= 0 || time.Until(r) > 2*time.Second {
		t.Errorf("an empty bucket at 1/s should reset about a second away, got %v", time.Until(r))
	}
}

func TestTokenBucket_WaitBlocksUntilATokenIsFree(t *testing.T) {
	ctx := context.Background()
	tb := NewTokenBucket(1, 100) // a token every 10ms
	tb.Allow(ctx)

	start := time.Now()
	if err := tb.Wait(ctx); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 2*time.Millisecond {
		t.Errorf("Wait returned after %v, before a token could be free", elapsed)
	}
}

func TestTokenBucket_WaitStopsWhenContextEnds(t *testing.T) {
	tb := NewTokenBucket(1, 0.5) // the next token is two seconds away
	tb.Allow(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := tb.Wait(ctx); err != context.DeadlineExceeded {
		t.Errorf("Wait = %v, want the context's deadline error", err)
	}
}

func TestSlidingWindowLog_LimitsWithinTheWindow(t *testing.T) {
	ctx := context.Background()
	swl := NewSlidingWindowLog(2, time.Minute)
	if !swl.Allow(ctx) || !swl.Allow(ctx) {
		t.Fatal("the first two requests in the window should pass")
	}
	if swl.Allow(ctx) {
		t.Error("a third request in the same window must be refused")
	}
	if got := swl.Remaining(); got != 0 {
		t.Errorf("Remaining = %d, want 0", got)
	}
	if got := swl.Limit(); got != 2 {
		t.Errorf("Limit = %d, want 2", got)
	}
	if r := swl.Reset(); !r.After(time.Now()) {
		t.Errorf("Reset = %v, want a time in the future while the window is full", r)
	}
}

func TestSlidingWindowLog_FreesSlotsAfterTheWindow(t *testing.T) {
	ctx := context.Background()
	swl := NewSlidingWindowLog(1, 20*time.Millisecond)
	swl.Allow(ctx)
	if swl.Allow(ctx) {
		t.Fatal("the single slot should be taken")
	}
	time.Sleep(25 * time.Millisecond)
	if !swl.Allow(ctx) {
		t.Error("the slot should be free once the window has passed")
	}
}

func TestSlidingWindowLog_ResetIsNowWhenEmpty(t *testing.T) {
	swl := NewSlidingWindowLog(5, time.Minute)
	if r := swl.Reset(); time.Until(r) > time.Second {
		t.Errorf("an empty window should reset now, got %v away", time.Until(r))
	}
}

func TestSlidingWindowLog_WaitStopsWhenContextEnds(t *testing.T) {
	swl := NewSlidingWindowLog(1, time.Minute)
	swl.Allow(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := swl.Wait(ctx); err != context.DeadlineExceeded {
		t.Errorf("Wait = %v, want the context's deadline error", err)
	}
}

func TestSlidingWindowLog_WaitSucceedsOnceAFreeSlotAppears(t *testing.T) {
	swl := NewSlidingWindowLog(1, 20*time.Millisecond)
	swl.Allow(context.Background())
	if err := swl.Wait(context.Background()); err != nil {
		t.Errorf("Wait = %v, want nil once the window has passed", err)
	}
}

func TestMultiRateLimiter_RequiresEveryLimiter(t *testing.T) {
	ctx := context.Background()
	strict := NewTokenBucket(1, 0.001)
	loose := NewTokenBucket(10, 1000)
	m := NewMultiRateLimiter(strict, loose)

	if !m.Allow(ctx) {
		t.Fatal("the first request should pass both limiters")
	}
	if m.Allow(ctx) {
		t.Error("the strict limiter is spent, so the combined limiter must refuse")
	}
	// The strict limiter needs about 1000 seconds for its next token, so Wait must
	// block until the context ends, and must report that.
	wctx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if err := m.Wait(wctx); err != context.DeadlineExceeded {
		t.Errorf("Wait = %v, want the deadline error while the strict limiter is empty", err)
	}
}

func TestMultiRateLimiter_ReportsTheTightestLimit(t *testing.T) {
	m := NewMultiRateLimiter(NewTokenBucket(5, 1), NewTokenBucket(2, 1))
	if got := m.Limit(); got != 2 {
		t.Errorf("Limit = %d, want 2 (the tightest of 5 and 2)", got)
	}
	if got := m.Remaining(); got > 2 {
		t.Errorf("Remaining = %d, want at most the tightest limiter's 2", got)
	}
	_ = m.Reset()
}

func TestDefaultRateLimits_CoversEveryPlatform(t *testing.T) {
	limits := DefaultRateLimits()
	if len(limits) == 0 {
		t.Fatal("no default rate limits")
	}
	for p, cfg := range limits {
		if cfg.Burst <= 0 {
			t.Errorf("%s has a burst of %d, want a positive burst", p, cfg.Burst)
		}
		if cfg.RequestsPerSecond <= 0 && cfg.Algorithm != "sliding-window" {
			t.Errorf("%s has rate %v, want a positive rate", p, cfg.RequestsPerSecond)
		}
	}
}

func TestPlatformRateLimiter_UnconfiguredPlatformIsNotLimited(t *testing.T) {
	ctx := context.Background()
	p := NewPlatformRateLimiter(map[Platform]RateLimitConfig{})
	for i := 0; i < 50; i++ {
		if !p.Allow(ctx, "never-configured") {
			t.Fatalf("an unconfigured platform was refused at request %d", i+1)
		}
	}
}

func TestPlatformRateLimiter_SetConfigReplacesTheLimit(t *testing.T) {
	ctx := context.Background()
	p := NewPlatformRateLimiter(map[Platform]RateLimitConfig{
		PlatformWeb: {RequestsPerSecond: 0.001, Burst: 1, Algorithm: "token-bucket"},
	})
	if !p.Allow(ctx, PlatformWeb) {
		t.Fatal("the first request should pass")
	}
	if p.Allow(ctx, PlatformWeb) {
		t.Fatal("the bucket should be empty")
	}

	p.SetConfig(PlatformWeb, RateLimitConfig{RequestsPerSecond: 1000, Burst: 5, Algorithm: "token-bucket"})
	if !p.Allow(ctx, PlatformWeb) {
		t.Error("after SetConfig with a bigger burst, the platform should allow again")
	}
	if got := p.Get(PlatformWeb).Limit(); got != 5 {
		t.Errorf("Limit = %d after SetConfig, want 5", got)
	}
}

func TestPlatformRateLimiter_SlidingWindowAlgorithmIsUsed(t *testing.T) {
	ctx := context.Background()
	p := NewPlatformRateLimiter(map[Platform]RateLimitConfig{
		PlatformGitHub: {Burst: 1, Window: time.Minute, Algorithm: "sliding-window"},
	})
	if _, ok := p.Get(PlatformGitHub).(*SlidingWindowLog); !ok {
		t.Fatal("a sliding-window config must build a SlidingWindowLog")
	}
	if !p.Allow(ctx, PlatformGitHub) {
		t.Fatal("the first request should pass")
	}
	if p.Allow(ctx, PlatformGitHub) {
		t.Error("the single slot is taken, so the second request must be refused")
	}
	// The window is a minute long, so Wait can only end through the context.
	wctx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if err := p.Wait(wctx, PlatformGitHub); err != context.DeadlineExceeded {
		t.Errorf("Wait = %v, want the deadline error while the window is full", err)
	}
}
