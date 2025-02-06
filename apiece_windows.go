//go:build windows
package apathy

import "strings"

// Normalize will step the given APiece down from its posix guarantees to a
// path or system specific separator form.
//
// Note: Windows 10+ are generally pretty fine with being given \\ or / in
// most scenarios, things that go explicitly through cmd being amongst the
// main exceptions.
//
// Check if you really need to do this before calling it a lot of times, m'kay?
func (p APiece) Normalize() string {
	return strings.Map(func (r rune) rune {
		if r == '/' {
			return '\\'
		}
		return r
	}, p.String())
}
