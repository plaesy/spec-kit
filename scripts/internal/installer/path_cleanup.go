package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// profileMarker is the comment line install.sh writes above the export it adds
// to a shell profile. Removal matches this line and the export line after it.
const profileMarker = "# Added by the plaesy installer"

// dropPathEntry returns value with every entry that names dir removed, and
// whether anything was removed. Entries are compared the way OnPath compares
// them: surrounding whitespace and quotes are ignored, and the comparison is
// case-insensitive when fold is true (Windows).
func dropPathEntry(value, dir string, fold bool) (string, bool) {
	target := filepath.Clean(dir)
	entries := strings.Split(value, ";")
	kept := make([]string, 0, len(entries))
	removed := false
	for _, entry := range entries {
		if samePathEntry(entry, target, fold) {
			removed = true
			continue
		}
		kept = append(kept, entry)
	}
	if !removed {
		return value, false
	}
	return strings.Join(kept, ";"), true
}

func samePathEntry(entry, target string, fold bool) bool {
	candidate := strings.Trim(strings.TrimSpace(entry), `"`)
	if candidate == "" {
		return false
	}
	candidate = filepath.Clean(candidate)
	if fold {
		return strings.EqualFold(candidate, target)
	}
	return candidate == target
}

// removeInstallerLines deletes the block install.sh appended to profile for
// dir: the marker line and the export line after it, plus the blank line the
// installer wrote in front of the marker. Other lines are left as they are.
// A missing profile is not an error.
func removeInstallerLines(profile, dir string) (bool, error) {
	info, err := os.Stat(profile)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", profile, err)
	}
	data, err := os.ReadFile(profile)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", profile, err)
	}

	want := fmt.Sprintf(`export PATH="%s:$PATH"`, dir)
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for i := 0; i < len(lines); i++ {
		if lines[i] == profileMarker && i+1 < len(lines) && lines[i+1] == want {
			if n := len(out); n > 0 && out[n-1] == "" {
				out = out[:n-1]
			}
			i++
			changed = true
			continue
		}
		out = append(out, lines[i])
	}
	if !changed {
		return false, nil
	}
	if err := os.WriteFile(profile, []byte(strings.Join(out, "\n")), info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("write %s: %w", profile, err)
	}
	return true, nil
}

// profileFiles lists the shell profiles install.sh may have written to.
func profileFiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}
	files := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
	}
	if runtime.GOOS == "darwin" {
		files = append(files, filepath.Join(home, ".bash_profile"))
	}
	return files, nil
}
