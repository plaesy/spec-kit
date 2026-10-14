package delegate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

// CapText is the capability every request needs. A provider that does not
// declare it is skipped, not asked to degrade.
const CapText = "text"

// DefaultMaxRetryAfter bounds how long a chain waits on one Retry-After. A
// longer wait moves on to the next provider instead.
const DefaultMaxRetryAfter = 60 * time.Second

// Attempt records what a chain did with one provider. Outcome is "answered",
// "skipped" (no call was made), or "failed" (a call was made and failed).
type Attempt struct {
	Provider string `json:"provider"`
	Outcome  string `json:"outcome"`
	Reason   string `json:"reason,omitempty"`
	Class    string `json:"class,omitempty"`
	WaitedMs int64  `json:"waited_ms,omitempty"`
}

// ChainError is returned when a chain produces no answer. It lists every
// attempt, so the caller can see why each provider was skipped or failed.
type ChainError struct {
	Chain    string
	Stopped  bool // true when a terminal error ended the chain early
	Attempts []Attempt
}

func (e *ChainError) Error() string {
	var b strings.Builder
	if e.Stopped {
		fmt.Fprintf(&b, "delegate chain %q stopped on a terminal error; no provider answered", e.Chain)
	} else {
		fmt.Fprintf(&b, "delegate chain %q: no provider answered", e.Chain)
	}
	for _, a := range e.Attempts {
		fmt.Fprintf(&b, "\n  - %s: %s", a.Provider, a.Outcome)
		if a.Class != "" {
			fmt.Fprintf(&b, " (%s)", a.Class)
		}
		if a.Reason != "" {
			fmt.Fprintf(&b, ": %s", a.Reason)
		}
	}
	return b.String()
}

// ChainOptions carries what a chain needs from its caller. Build, Check, and
// Sleep are injected so tests can use fakes and a fake clock.
type ChainOptions struct {
	Chain         string // name shown in errors; "" for a single provider
	Confirm       bool   // --confirm: allows billable providers to run
	MaxRetryAfter time.Duration
	// Build turns a config entry into a Provider. Required.
	Build func(config.DelegateProviderConfig) (Provider, error)
	// Check runs before each call. A non-nil error (a budget refusal) stops the
	// chain and no request is sent. Optional.
	Check func(config.DelegateProviderConfig, Request) error
	// Sleep waits for a Retry-After. Optional; defaults to a context-aware sleep.
	Sleep func(ctx context.Context, d time.Duration) error
}

// EnvBuild returns the default Build, which reads API keys from the process
// environment.
func EnvBuild() func(config.DelegateProviderConfig) (Provider, error) {
	return func(e config.DelegateProviderConfig) (Provider, error) {
		return Build(e, os.Getenv)
	}
}

// SleepCtx waits for d, or returns early with the context's error.
func SleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Capabilities returns what a provider can take. A provider that does not
// declare capabilities is assumed to take text only.
func Capabilities(p Provider) []string {
	if c, ok := p.(interface{ Capabilities() []string }); ok {
		return c.Capabilities()
	}
	return []string{CapText}
}

func hasCapability(caps []string, want string) bool {
	for _, c := range caps {
		if c == want {
			return true
		}
	}
	return false
}

// RunChain tries names in order and returns the first answer, with the attempts
// made along the way. It falls through on a Transient error, and on a
// RateLimited error after waiting for Retry-After when the wait is within
// MaxRetryAfter. A Terminal error stops the chain. A disabled provider, a
// billable provider without --confirm, and a provider that lacks the text
// capability are skipped without a call.
func RunChain(ctx context.Context, cfg *config.DelegateConfig, names []string, req Request, opts ChainOptions) (Response, []Attempt, error) {
	if opts.Build == nil {
		opts.Build = EnvBuild()
	}
	if opts.Sleep == nil {
		opts.Sleep = SleepCtx
	}
	if opts.MaxRetryAfter <= 0 {
		opts.MaxRetryAfter = DefaultMaxRetryAfter
	}

	var attempts []Attempt
	stop := func(a Attempt) (Response, []Attempt, error) {
		attempts = append(attempts, a)
		return Response{}, attempts, &ChainError{Chain: opts.Chain, Stopped: true, Attempts: attempts}
	}

	for _, name := range names {
		entry, err := Select(cfg, name, opts.Confirm)
		if err != nil {
			attempts = append(attempts, Attempt{Provider: name, Outcome: "skipped", Reason: err.Error()})
			continue
		}
		p, err := opts.Build(entry)
		if err != nil {
			return stop(Attempt{Provider: name, Outcome: "failed", Reason: err.Error(), Class: Terminal.String()})
		}
		if !hasCapability(Capabilities(p), CapText) {
			attempts = append(attempts, Attempt{Provider: name, Outcome: "skipped", Reason: "provider lacks the text capability"})
			continue
		}
		if opts.Check != nil {
			if err := opts.Check(entry, req); err != nil {
				class, _ := classify(err)
				return stop(Attempt{Provider: name, Outcome: "failed", Reason: err.Error(), Class: class.String()})
			}
		}

		resp, err := Run(ctx, p, req)
		if err == nil {
			attempts = append(attempts, Attempt{Provider: name, Outcome: "answered"})
			return resp, attempts, nil
		}
		class, retryAfter := classify(err)
		att := Attempt{Provider: name, Outcome: "failed", Reason: err.Error(), Class: class.String()}
		if class == Terminal {
			return stop(att)
		}
		if class == RateLimited && retryAfter > 0 {
			if retryAfter > opts.MaxRetryAfter {
				att.Reason += fmt.Sprintf("; retry-after %s exceeds the %s maximum, moving on", retryAfter, opts.MaxRetryAfter)
			} else {
				start := time.Now()
				if werr := opts.Sleep(ctx, retryAfter); werr != nil {
					att.Reason += "; interrupted while waiting: " + werr.Error()
					attempts = append(attempts, att)
					return Response{}, attempts, &ChainError{Chain: opts.Chain, Stopped: true, Attempts: attempts}
				}
				att.WaitedMs = time.Since(start).Milliseconds()
			}
		}
		attempts = append(attempts, att)
		if ctx.Err() != nil {
			return Response{}, attempts, &ChainError{Chain: opts.Chain, Stopped: true, Attempts: attempts}
		}
	}
	return Response{}, attempts, &ChainError{Chain: opts.Chain, Attempts: attempts}
}
