package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigManager provides generic configuration management with hot-reload support.
type ConfigManager[T any] struct {
	path       string
	mu         sync.RWMutex
	cache      *T
	defaultCfg *T
	watchers   []func(*T)
	watcher    *fsnotify.Watcher
	stopWatch  context.CancelFunc
	useKeyring bool
	keyringSvc string
}

// NewConfigManager creates a new config manager with the given default configuration.
func NewConfigManager[T any](path string, defaultCfg *T) *ConfigManager[T] {
	if path == "" {
		path = UnifiedConfigPath()
	}
	return &ConfigManager[T]{
		path:       path,
		defaultCfg: defaultCfg,
		cache:      nil, // Start with nil cache, load from disk on first Load()
		useKeyring: true,
		keyringSvc: "plaesy-config",
	}
}

// UnifiedConfigPath returns the default unified config path.
func UnifiedConfigPath() string {
	return filepath.Join(".plaesy", "config.json")
}

// Load loads the configuration from disk, returning cached value if available.
func (cm *ConfigManager[T]) Load() (*T, error) {
	cm.mu.RLock()
	if cm.cache != nil {
		cfg := cm.cache
		cm.mu.RUnlock()
		return cfg, nil
	}
	cm.mu.RUnlock()

	return cm.loadFromDisk()
}

// loadFromDisk reads and parses the config file.
func (cm *ConfigManager[T]) loadFromDisk() (*T, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			cm.cache = cm.defaultCfg
			return cm.cache, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg T
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cm.cache = &cfg
	return cm.cache, nil
}

// Save writes the configuration to disk atomically.
func (cm *ConfigManager[T]) Save(cfg *T) error {
	cm.mu.Lock()

	if err := os.MkdirAll(filepath.Dir(cm.path), 0o755); err != nil {
		cm.mu.Unlock()
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		cm.mu.Unlock()
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Atomic write via temp file
	tmpPath := cm.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		cm.mu.Unlock()
		return fmt.Errorf("writing temp config: %w", err)
	}
	if err := os.Rename(tmpPath, cm.path); err != nil {
		cm.mu.Unlock()
		return fmt.Errorf("renaming config: %w", err)
	}

	cm.cache = cfg
	cm.mu.Unlock()

	// Notify watchers AFTER releasing the lock to avoid deadlock
	cm.notifyWatchers(cfg)
	return nil
}

// Watch registers a callback for configuration changes (hot-reload).
func (cm *ConfigManager[T]) Watch(ctx context.Context, callback func(*T)) error {
	cm.mu.Lock()
	cm.watchers = append(cm.watchers, callback)
	cm.mu.Unlock()

	// Start file watcher if not already running
	if cm.watcher == nil {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("creating file watcher: %w", err)
		}
		cm.watcher = w

		watchCtx, cancel := context.WithCancel(ctx)
		cm.stopWatch = cancel

		go cm.watchLoop(watchCtx)
	}

	// Ensure config directory is watched
	dir := filepath.Dir(cm.path)
	if err := cm.watcher.Add(dir); err != nil {
		return fmt.Errorf("watching config dir: %w", err)
	}

	return nil
}

// watchLoop handles file system events.
func (cm *ConfigManager[T]) watchLoop(ctx context.Context) {
	defer cm.watcher.Close()

	debounce := time.NewTimer(100 * time.Millisecond)
	if !debounce.Stop() {
		<-debounce.C
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				debounce.Reset(100 * time.Millisecond)
			}
		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err
		case <-debounce.C:
			// Reload and notify
			if cfg, err := cm.loadFromDisk(); err == nil {
				cm.notifyWatchers(cfg)
			}
		}
	}
}

// notifyWatchers calls all registered watchers with the new config.
func (cm *ConfigManager[T]) notifyWatchers(cfg *T) {
	cm.mu.RLock()
	watchers := make([]func(*T), len(cm.watchers))
	copy(watchers, cm.watchers)
	cm.mu.RUnlock()

	// Call watchers without holding any lock to avoid deadlock
	for _, w := range watchers {
		w(cfg)
	}
}

// StopWatch stops the file watcher.
func (cm *ConfigManager[T]) StopWatch() {
	if cm.stopWatch != nil {
		cm.stopWatch()
	}
}

// MigrateFrom migrates configuration from an old path to the unified config.
func (cm *ConfigManager[T]) MigrateFrom(oldPath string, migrateFn func([]byte) (*T, error)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No old config to migrate
		}
		return fmt.Errorf("reading old config: %w", err)
	}

	cfg, err := migrateFn(data)
	if err != nil {
		return fmt.Errorf("migrating config: %w", err)
	}

	// Save to new unified location
	if err := cm.saveUnlocked(cfg); err != nil {
		return err
	}

	// Remove old config file
	if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing old config: %w", err)
	}

	return nil
}

// saveUnlocked saves config without locking (caller must hold lock).
func (cm *ConfigManager[T]) saveUnlocked(cfg *T) error {
	if err := os.MkdirAll(filepath.Dir(cm.path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tmpPath := cm.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing temp config: %w", err)
	}
	if err := os.Rename(tmpPath, cm.path); err != nil {
		return fmt.Errorf("renaming config: %w", err)
	}

	cm.cache = cfg
	return nil
}

// Get returns the cached config (nil if not loaded).
func (cm *ConfigManager[T]) Get() *T {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.cache
}

// SetKeyringEnabled enables/disables keyring integration.
func (cm *ConfigManager[T]) SetKeyringEnabled(enabled bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.useKeyring = enabled
}
