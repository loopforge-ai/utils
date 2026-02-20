package fs

import (
	"path/filepath"
	"slices"
	"strings"
)

// IsSafePath reports whether path is a relative path that does not escape
// the current directory via ".." segments or absolute prefixes.
func IsSafePath(path string) bool {
	if filepath.IsAbs(path) {
		return false
	}
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	return !slices.Contains(parts, "..")
}
