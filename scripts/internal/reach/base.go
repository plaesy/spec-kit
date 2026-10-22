// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BaseAdapter provides common functionality for platform adapters.
type BaseAdapter struct {
	name         Platform
	backends     []Backend
	requiresAuth bool
}

// Name returns the platform identifier.
func (b *BaseAdapter) Name() Platform {
	return b.name
}

// Backends returns the ordered list of backends for this platform.
func (b *BaseAdapter) Backends() []Backend {
	return b.backends
}

// RequiresAuth returns true if the platform needs authentication.
func (b *BaseAdapter) RequiresAuth() bool {
	return b.requiresAuth
}

// Execute runs a command and returns the result.
func (b *BaseAdapter) runCommand(ctx context.Context, backend Backend, args ...string) (BackendResult, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)

	result := BackendResult{
		Backend:   backend,
		Platform:  b.name,
		Timestamp: time.Now(),
		Latency:   latency,
	}

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("%v: %s", err, string(output))
		return result, nil
	}

	result.Success = true
	result.Data = strings.TrimSpace(string(output))
	return result, nil
}

// Health checks if a command is available.
func (b *BaseAdapter) checkCommand(ctx context.Context, backend Backend, cmd string, args ...string) BackendHealth {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, cmd, args...)
	err := c.Run()
	latency := time.Since(start)

	health := BackendHealth{
		Backend:      backend,
		Platform:     b.name,
		LastChecked:  time.Now(),
		Latency:      latency,
		RequiresAuth: b.requiresAuth,
	}

	if err != nil {
		health.Available = false
		health.Error = err.Error()
	} else {
		health.Available = true
		// Try to get version
		if v, err := b.getVersion(ctx, cmd); err == nil {
			health.Version = v
		}
	}
	return health
}

func (b *BaseAdapter) getVersion(ctx context.Context, cmd string) (string, error) {
	c := exec.CommandContext(ctx, cmd, "--version")
	output, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Configure sets up authentication/credentials.
func (b *BaseAdapter) Configure(creds map[string]string) error {
	// Base implementation does nothing
	return nil
}
