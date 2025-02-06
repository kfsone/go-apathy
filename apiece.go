package apathy

import (
	"path"
	"strings"
)

// APiece is a string intended for file-system path construction but with the
// guarantee that it has been sanitized by having path-separators normalized to
// the posix style ('/') and undergone a path.Clean() operation.
//
// Modern Windows is happy to accept posix-style paths in most places, and knowing
// the string is already clean allows various optimizations and simplifications
// in path manipulation.
//
// Example
//
//	// Without Apathy
//	p1, err := filepath.Abs(filepath.Clean(filepath.Join(f1, f2)))
//	// ...
//	p2, err := filepath.Abs(filepath.Clean(filepath.Join(f1, f3))  // recleaning f1?
//
//	// With Apathy
//	a1, err := apathy.NewAPath(f1),
//	p1 := apathy.NewAPiece(f2)
//	...
//	a2, err := apathy.NewAPath(a1, p1)
type APiece string

// NewAPiece cleans the given string and ensures it is in posix-form,
// regardless of the platform the code is running on.
func NewAPiece(str string) APiece {
	str = ToSlash(str)

	// For windows filepaths, an absolute drive root is `<letter>:/`, but clean
	// might interfere with the trailing slash. If this is an absolute drive
	// reference without a path, don't clean it.
	if len(str) <= 3 {
		piece := APiece(str)
		if hasAbsDrive(piece) {
			return piece
		}
	}

	// Ok, tidy away, and that's a piece.
	return APiece(path.Clean(str))
}

// Piece returns the APiece representation of this path component.
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

// Helpers.

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
	return hasAbsDrive(p)
}
