package common

import (
	"regexp"
	"strings"
)

// Version is set at build time via -ldflags "-X .../common.Version=1.2.3".
// Falls back to "0.0.0" if not injected, matching common.sh's behavior when
// no VERSION file can be located.
var Version = "0.0.0"

var semverRE = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// NormalizeVersion mirrors get_plaesy_version()'s sanitize step: falls back
// to 0.0.0 for anything that isn't strict semver.
func NormalizeVersion(raw string) string {
	v := strings.TrimSpace(raw)
	if !semverRE.MatchString(v) {
		return "0.0.0"
	}
	return v
}
