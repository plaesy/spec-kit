package delegate

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestAnthropicClassifiesHTTPStatus runs the Messages provider against a local
// test server, never the real API, and checks each status maps to its class.
func TestAnthropicClassifiesHTTPStatus(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		retryAfter string
		wantClass  Class
		wantWait   time.Duration
	}{
		{"429 with Retry-After seconds", http.StatusTooManyRequests, "7", RateLimited, 7 * time.Second},
		{"429 without Retry-After", http.StatusTooManyRequests, "", RateLimited, 0},
		{"503 is transient", http.StatusServiceUnavailable, "", Transient, 0},
		{"529 overloaded is transient", 529, "", Transient, 0},
		{"401 bad key is terminal", http.StatusUnauthorized, "", Terminal, 0},
		{"400 bad request is terminal", http.StatusBadRequest, "", Terminal, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.retryAfter != "" {
					w.Header().Set("Retry-After", tc.retryAfter)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"error":"x"}`))
			}))
			defer srv.Close()

			p := &anthropicProvider{
				name: "anthropic", model: "m", endpoint: srv.URL, keyEnv: "K",
				getenv: func(string) string { return "test-key" }, client: srv.Client(),
			}
			_, err := p.Complete(context.Background(), Request{Prompt: "p"})
			class, wait := classify(err)
			if class != tc.wantClass {
				t.Fatalf("class = %v, want %v (err=%v)", class, tc.wantClass, err)
			}
			if wait != tc.wantWait {
				t.Fatalf("retry-after = %v, want %v", wait, tc.wantWait)
			}
		})
	}
}

// TestAnthropicMissingKeyIsTerminal checks a missing key stops a chain rather
// than falling through, since another provider will not fix configuration.
func TestAnthropicMissingKeyIsTerminal(t *testing.T) {
	p := &anthropicProvider{name: "anthropic", model: "m", keyEnv: "ANTHROPIC_API_KEY", getenv: func(string) string { return "" }}
	_, err := p.Complete(context.Background(), Request{Prompt: "p"})
	if class, _ := classify(err); class != Terminal {
		t.Fatalf("missing key class = %v, want terminal", class)
	}
	var de *Error
	if !errors.As(err, &de) {
		t.Fatal("missing key error should carry a class")
	}
}

// TestClassifyUnclassifiedIsTerminal checks a plain error is terminal, so an
// unknown failure never triggers a second, billed call elsewhere.
func TestClassifyUnclassifiedIsTerminal(t *testing.T) {
	if class, _ := classify(errors.New("unknown")); class != Terminal {
		t.Fatalf("unclassified class = %v, want terminal", class)
	}
}
