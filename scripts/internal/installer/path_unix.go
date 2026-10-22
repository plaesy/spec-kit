//go:build !windows

package installer

// removeFromUserPath removes the export block install.sh added for dir from
// each shell profile it may have written to. It reports whether any profile
// changed.
func removeFromUserPath(dir string) (bool, error) {
	files, err := profileFiles()
	if err != nil {
		return false, err
	}
	removed := false
	for _, file := range files {
		changed, err := removeInstallerLines(file, dir)
		if err != nil {
			return removed, err
		}
		removed = removed || changed
	}
	return removed, nil
}
