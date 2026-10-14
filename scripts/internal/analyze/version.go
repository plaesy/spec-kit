package analyze

import "github.com/plaesy/spec-kit/internal/common"

// versionFromCommon exposes common.Version (set via -ldflags at build time)
// without analyze.go needing to import internal/common directly, keeping
// the dependency in one small, obvious place.
func versionFromCommon() string {
	return common.Version
}
