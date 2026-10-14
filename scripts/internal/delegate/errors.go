package delegate

import (
	"errors"
	"time"
)

// Class says what a failed call means for the rest of a chain.
type Class int

const (
	// Transient: a timeout, a network error, a 5xx response, or a CLI that did
	// not run. The next provider may succeed, so the chain falls through.
	Transient Class = iota + 1
	// RateLimited: the provider asked us to wait (HTTP 429). The chain honours
	// RetryAfter when it is within the maximum wait, then falls through.
	RateLimited
	// Terminal: an auth failure, a bad request, bad configuration, a budget
	// refusal, or an error we cannot classify. The chain stops, because the same
	// request would fail elsewhere too, or because falling through could repeat
	// a billed request.
	Terminal
)

// String names the class for output.
func (c Class) String() string {
	switch c {
	case Transient:
		return "transient"
	case RateLimited:
		return "rate-limited"
	case Terminal:
		return "terminal"
	default:
		return "unknown"
	}
}

// Error is a provider failure with its class. Its message is the message of
// the wrapped error, so callers that only print the text see no difference.
type Error struct {
	Class      Class
	RetryAfter time.Duration
	Err        error
}

func (e *Error) Error() string { return e.Err.Error() }

// Unwrap exposes the underlying error to errors.Is and errors.As.
func (e *Error) Unwrap() error { return e.Err }

// classify returns the class of err. An error that carries no class is Terminal:
// the chain must not guess that a failure is safe to retry elsewhere.
func classify(err error) (Class, time.Duration) {
	var de *Error
	if errors.As(err, &de) {
		return de.Class, de.RetryAfter
	}
	return Terminal, 0
}
