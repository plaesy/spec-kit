//go:build windows

package installer

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// removeFromUserPath removes dir from the user PATH in HKCU\Environment. It
// reports whether the value changed. The machine PATH is never touched.
func removeFromUserPath(dir string) (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false, fmt.Errorf("open HKCU\\Environment: %w", err)
	}
	defer key.Close()

	value, valType, err := key.GetStringValue("Path")
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read user PATH: %w", err)
	}

	updated, changed := dropPathEntry(value, dir, true)
	if !changed {
		return false, nil
	}
	// Keep the value type: a user PATH is usually REG_EXPAND_SZ.
	if valType == registry.EXPAND_SZ {
		err = key.SetExpandStringValue("Path", updated)
	} else {
		err = key.SetStringValue("Path", updated)
	}
	if err != nil {
		return false, fmt.Errorf("write user PATH: %w", err)
	}
	return true, nil
}
