package delegate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

// chainFake is a provider that returns a preset error, or answers, and counts
// its calls. A billable fake that must never run sets fail to true.
type chainFake struct {
	name  string
	err   error
	caps  []string // nil means text only, the default
	calls int
	t     *testing.T
}

func (f *chainFake) Name() string { return f.name }

func (f *chainFake) Capabilities() []string {
	if f.caps != nil {
		return f.caps
	}
	return []string{CapText}
}

func (f *chainFake) Complete(ctx context.Context, req Request) (Response, error) {
	f.calls++
	if f.err != nil {
		return Response{}, f.err
	}
	return Response{Text: "answer from " + f.name}, nil
}

// chainCfg lists providers with the given billable flag, all enabled.
func chainCfg(names []string, billable map[string]bool) *config.DelegateConfig {
	cfg := &config.DelegateConfig{}
	for _, n := range names {
		cfg.Providers = append(cfg.Providers, config.DelegateProviderConfig{
			Name: n, Type: "cli", Enabled: true, Billable: billable[n], Command: "x",
		})
	}
	return cfg
}

// chainOpts builds options whose Build returns the fakes by name and whose Sleep
// records the requested waits instead of sleeping.
func chainOpts(fakes map[string]*chainFake, slept *[]time.Duration) ChainOptions {
	return ChainOptions{
		Build: func(e config.DelegateProviderConfig) (Provider, error) {
			return fakes[e.Name], nil
		},
		Sleep: func(ctx context.Context, d time.Duration) error {
			*slept = append(*slept, d)
			return nil
		},
	}
}

// TestRunChain_ErrorClasses is table-driven over each error class and checks the
// next step the chain takes.
func TestRunChain_ErrorClasses(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantNext   bool // true: the second provider is called and answers
		wantStop   bool
		wantWaited bool
	}{
		{"transient falls through", &Error{Class: Transient, Err: errors.New("503")}, true, false, false},
		{"rate limited with retry-after waits then falls through", &Error{Class: RateLimited, RetryAfter: 7 * time.Second, Err: errors.New("429")}, true, false, true},
		{"terminal stops the chain", &Error{Class: Terminal, Err: errors.New("401")}, false, true, false},
		{"unclassified error is terminal", errors.New("boom"), false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first := &chainFake{name: "first", err: tc.err}
			second := &chainFake{name: "second"}
			fakes := map[string]*chainFake{"first": first, "second": second}
			var slept []time.Duration
			cfg := chainCfg([]string{"first", "second"}, nil)

			resp, attempts, err := RunChain(context.Background(), cfg, []string{"first", "second"}, Request{Prompt: "p"}, chainOpts(fakes, &slept))

			if tc.wantNext {
				if err != nil || resp.Provider != "second" {
					t.Fatalf("want second to answer, got resp=%+v err=%v", resp, err)
				}
			}
			if tc.wantStop {
				var ce *ChainError
				if !errors.As(err, &ce) || !ce.Stopped {
					t.Fatalf("want a stopped ChainError, got %v", err)
				}
				if second.calls != 0 {
					t.Fatalf("terminal error must not fall through, but second was called %d times", second.calls)
				}
			}
			if tc.wantWaited && (len(slept) != 1 || slept[0] != 7*time.Second) {
				t.Fatalf("want one wait of 7s for retry-after, got %v", slept)
			}
			if len(attempts) == 0 {
				t.Fatal("attempts must be recorded")
			}
		})
	}
}

// TestRunChain_RetryAfterRespected checks the wait is exactly the Retry-After and
// that a wait above the maximum is not taken.
func TestRunChain_RetryAfterRespected(t *testing.T) {
	first := &chainFake{name: "first", err: &Error{Class: RateLimited, RetryAfter: 5 * time.Second, Err: errors.New("429")}}
	second := &chainFake{name: "second"}
	fakes := map[string]*chainFake{"first": first, "second": second}
	var slept []time.Duration
	cfg := chainCfg([]string{"first", "second"}, nil)
	if _, _, err := RunChain(context.Background(), cfg, []string{"first", "second"}, Request{Prompt: "p"}, chainOpts(fakes, &slept)); err != nil {
		t.Fatalf("RunChain: %v", err)
	}
	if len(slept) != 1 || slept[0] != 5*time.Second {
		t.Fatalf("want wait 5s, got %v", slept)
	}

	// Above the maximum, the chain moves on without waiting.
	first2 := &chainFake{name: "first", err: &Error{Class: RateLimited, RetryAfter: 10 * time.Minute, Err: errors.New("429")}}
	fakes2 := map[string]*chainFake{"first": first2, "second": {name: "second"}}
	var slept2 []time.Duration
	opts := chainOpts(fakes2, &slept2)
	opts.MaxRetryAfter = time.Minute
	if _, _, err := RunChain(context.Background(), cfg, []string{"first", "second"}, Request{Prompt: "p"}, opts); err != nil {
		t.Fatalf("RunChain: %v", err)
	}
	if len(slept2) != 0 {
		t.Fatalf("a retry-after above the maximum must not be waited on, got %v", slept2)
	}
}

// TestRunChain_NoAnswerListsEveryAttempt checks a chain that never answers
// returns an error naming every attempt, not an empty success.
func TestRunChain_NoAnswerListsEveryAttempt(t *testing.T) {
	fakes := map[string]*chainFake{
		"a": {name: "a", err: &Error{Class: Transient, Err: errors.New("timeout")}},
		"b": {name: "b", err: &Error{Class: Transient, Err: errors.New("503")}},
	}
	cfg := chainCfg([]string{"a", "b"}, nil)
	var slept []time.Duration
	_, attempts, err := RunChain(context.Background(), cfg, []string{"a", "b"}, Request{Prompt: "p"}, chainOpts(fakes, &slept))
	if err == nil {
		t.Fatal("want an error when no provider answers")
	}
	msg := err.Error()
	for _, want := range []string{"a: failed", "b: failed"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not list attempt %q", msg, want)
		}
	}
	if len(attempts) != 2 {
		t.Fatalf("want 2 attempts, got %d", len(attempts))
	}
}

// TestRunChain_BillableNeedsConfirm checks a billable provider is skipped without
// --confirm and is never called. The fake fails the test if Complete runs.
func TestRunChain_BillableNeedsConfirm(t *testing.T) {
	paid := &chainFake{name: "paid", t: t}
	free := &chainFake{name: "free"}
	fakes := map[string]*chainFake{"paid": paid, "free": free}
	cfg := chainCfg([]string{"paid", "free"}, map[string]bool{"paid": true})
	var slept []time.Duration

	resp, attempts, err := RunChain(context.Background(), cfg, []string{"paid", "free"}, Request{Prompt: "p"}, chainOpts(fakes, &slept))
	if err != nil || resp.Provider != "free" {
		t.Fatalf("want free to answer after paid is skipped, got resp=%+v err=%v", resp, err)
	}
	if paid.calls != 0 {
		t.Fatalf("billable provider was called %d times without --confirm", paid.calls)
	}
	if attempts[0].Outcome != "skipped" {
		t.Fatalf("want paid skipped, got %+v", attempts[0])
	}

	// With --confirm the same chain uses the billable provider first.
	paid2 := &chainFake{name: "paid"}
	fakes2 := map[string]*chainFake{"paid": paid2, "free": {name: "free"}}
	opts := chainOpts(fakes2, &slept)
	opts.Confirm = true
	resp2, _, err := RunChain(context.Background(), cfg, []string{"paid", "free"}, Request{Prompt: "p"}, opts)
	if err != nil || resp2.Provider != "paid" {
		t.Fatalf("want paid to answer with --confirm, got resp=%+v err=%v", resp2, err)
	}
}

// TestRunChain_CapabilityGateSkips checks a provider that lacks the text
// capability is skipped, not called.
func TestRunChain_CapabilityGateSkips(t *testing.T) {
	imageOnly := &chainFake{name: "image-only", caps: []string{"image"}}
	text := &chainFake{name: "text"}
	fakes := map[string]*chainFake{"image-only": imageOnly, "text": text}
	cfg := chainCfg([]string{"image-only", "text"}, nil)
	var slept []time.Duration
	resp, attempts, err := RunChain(context.Background(), cfg, []string{"image-only", "text"}, Request{Prompt: "p"}, chainOpts(fakes, &slept))
	if err != nil || resp.Provider != "text" {
		t.Fatalf("want text to answer, got resp=%+v err=%v", resp, err)
	}
	if imageOnly.calls != 0 {
		t.Fatalf("provider without the text capability was called")
	}
	if attempts[0].Outcome != "skipped" || !strings.Contains(attempts[0].Reason, "text capability") {
		t.Fatalf("want a capability skip reason, got %+v", attempts[0])
	}
}

// TestRunChain_BudgetRefusalSendsNothing checks a budget refusal stops the chain
// before any call, and the refused provider is never called.
func TestRunChain_BudgetRefusalSendsNothing(t *testing.T) {
	paid := &chainFake{name: "paid"}
	free := &chainFake{name: "free"}
	fakes := map[string]*chainFake{"paid": paid, "free": free}
	cfg := chainCfg([]string{"paid", "free"}, map[string]bool{"paid": true})
	var slept []time.Duration
	opts := chainOpts(fakes, &slept)
	opts.Confirm = true
	opts.Check = func(e config.DelegateProviderConfig, req Request) error {
		return &Error{Class: Terminal, Err: fmt.Errorf("budget refused for %s", e.Name)}
	}
	_, _, err := RunChain(context.Background(), cfg, []string{"paid", "free"}, Request{Prompt: "p"}, opts)
	var ce *ChainError
	if !errors.As(err, &ce) || !ce.Stopped {
		t.Fatalf("want the chain to stop on the budget refusal, got %v", err)
	}
	if paid.calls != 0 || free.calls != 0 {
		t.Fatalf("a budget refusal must send no request; paid=%d free=%d", paid.calls, free.calls)
	}
}
