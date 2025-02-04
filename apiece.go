// APiece encapsulates a string with the guarantee that any path separators it has
// had any old Windows-style path-separators ('\\') transformed to posix-style ('/'),
// and that it has undergone a path.Clean() operation.
//
// Modern Windows is happy to accept posix-style paths in most places, and knowing
// the string is already clean allows various optimizations and simplifications
// in path manipulation.
package apathy

import (
	"path"
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
	if ExpectWindowsSlashes {
		str = strings.ReplaceAll(str, "\\", "/")
		// Special case for 'C: and 'C:/'.
		piece := APiece(str)
		if hasAbsDrive(piece) {
			return piece
		}
	}
	str = path.Clean(str)
	return APiece(str)
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

// Simple drive-letter check, does not handle UNC paths or powershell mount names.
func hasDriveLetter(p APiece) bool {
	if len(p) < 2 || p[1] != ':' {
		return false
	}
	letter := strings.ToLower(string(p[0]))
	return letter >= "a" && letter <= "z"
}

func hasAbsDrive(p APiece) bool {
	return len(p) >= 3 && p[2] == '/' && hasDriveLetter(p)
}

func (p APiece) IsAbs() bool {
	if len(p) >= 1 && p[0] == '/' {
		return true
	}
	return ExpectWindowsSlashes && hasAbsDrive(p)
}

// ToNative returns the native rendering of the path component, which means on windows
// the '/'s are replaced with '\'s. When ExpectWindowsSlashes is true on a Posix
// host, it will replace '/'s with '\\'s for paths that start with a drive letter.
// Windows note: When using Windows 10 or above, most filesystem APIs and programs
// are happy to accept posix-style paths, so you may not need to call this function.
func (p APiece) ToNative() string {
	if ExpectWindowsSlashes {
		if runtime.GOOS == "windows" {
			return filepath.Clean(p.String())
		}
		if hasDriveLetter(p) {
			return strings.ReplaceAll(p.String(), "/", "\\")
		}
	}
	return p.String()
}
