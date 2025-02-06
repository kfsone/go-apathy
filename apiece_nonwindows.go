//go:build !windows

package apathy

import "strings"

// Normalize will step the given APiece down from its posix guarantees to a
// path or system specific separator form. On posix, if the path does not
// appear to start with a windows-style drive letter, this just returns the
// string.
//
// Thus "c:/windows" -> "c:\windows", but "/windows" -> "/windows".
//
// Note the behavior is different on Windows, where the slashes are *always*
// replaced.
func (p APiece) Normalize() string {
	if hasDriveLetter(p) {
		return strings.Map(func(r rune) rune {
			if r == '/' {
				return '\\'
			}
			return r
		}, p.String())
	}
	return p.String()
}
