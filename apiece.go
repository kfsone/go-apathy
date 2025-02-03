// APiece encapsulates a string with the guarantee that any path separators it has
// had any old Windows-style path-separators ('\\') transformed to posix-style ('/'),
// and that it has undergone a path.Clean() operation.
//
// Modern Windows is happy to accept posix-style paths in most places, and knowing
// the string is already clean allows various optimizations and simplifications
// in path manipulation.
package apathy

import (
	"path/filepath"
	"runtime"
	"strings"
)

// APiece is a path component with posix-slashes and path.Clean()d.
type APiece string

// NewAPiece cleans the given string and ensures it is in posix-form,
// regardless of the platform the code is running on.
func NewAPiece(str string) APiece {
	// If we're on a posix system, it won't actually try and replace windows
	// slashes with posix slashes.
	if runtime.GOOS == "windows" || ExpectWindowsSlashes {
		str = strings.ReplaceAll(str, "\\", "/")
	}
	return APiece(filepath.Clean(str))
}

// Piece returns the the APiece representation of this path component.
func (p APiece) Piece() APiece {
	return p
}

// String returns the posix-notation string representation of this path component.
func (p APiece) String() string {
	return string(p)
}

// Len returns the length of the path component.
func (p APiece) Len() int {
	return len(p)
}

func (p APiece) IsAbs() bool {
	return filepath.IsAbs(p.String())
}

// ToNative returns the native rendering of the path component, which means on windows
// the '/'s are replaced with '\'s.
// Windows note: When using Windows 10 or above, most filesystem APIs and programs
// are happy to accept posix-style paths, so you may not need to call this function.
func (p APiece) ToNative() string {
	if runtime.GOOS == "windows" {
		return filepath.Clean(p.String())
	}
	return p.String()
}
