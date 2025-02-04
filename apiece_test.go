package apathy

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

type apieceTestCase struct {
	name     string
	input    string
	expected string
}

func TestNewAPiece_Basic(t *testing.T) {
	// Under the hood, NewAPiece is mostly just a wrapper around filepath.Clean,
	// so I'm not going to go to town on testing every kind of path I can think of.

	for _, tc := range []apieceTestCase{
		{"empty", "", "."},
		{"dot", ".", "."},
		{"slash", "/", "/"},
		{"slash-dot", "/.", "/"},
		{"dot-slash", "./", "."},
		{"simple", "foo", "foo"},
		{"foo-slash-bar", "foo/bar", "foo/bar"},
		{"slash-foo-slash-slash-bar-slash", "//foo/bar/", "/foo/bar"},
		{"trailing-slashes", "foo///////", "foo"},
		{"parent-parent-child", "../../Child", "../../Child"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := NewAPiece(tc.input)
			assert.Equal(t, tc.expected, result.String())
		})

	}
}

func TestNewAPiece_WindowsSlahes(t *testing.T) {
	// Save and restore the natural slash handling state.
	// Test with Windows slashes and with handling enabled/disabled.
	for _, tc := range []struct{ name, input, posix, win string }{
		{"empty", "", ".", "."},
		{"slash", "/", "/", "/"},
		{"black-slash", "\\", "\\", "/"},
		{"slash-slash-dot-blackslash", "//.\\", "/.\\", "/"},
		{"backslash-slash-dot", "\\/.", "\\", "/"},
		{"parent/parent/child", "../../Child", "../../Child", "../../Child"},
		{"parent\\parent\\child", "..\\..\\Child", "..\\..\\Child", "../../Child"},
		{"dot/parent\\parent\\child", "./..\\..\\Child", "..\\..\\Child", "../../Child"},
		{"dot\\parent\\parent\\child", ".\\..\\..\\Child", ".\\..\\..\\Child", "../../Child"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			posix := WithWindowsSlashesSetTo(false, func() APiece { return NewAPiece(tc.input) })
			win := WithWindowsSlashesSetTo(true, func() APiece { return NewAPiece(tc.input) })
			assert.Equal(t, tc.posix, posix.String())
			assert.Equal(t, tc.win, win.String())
		})
	}
}

func TestAPiece_Piece(t *testing.T) {
	// Coercively create our own malformed string, to validate there's no processing.
	input := "\\x//"
	piece := APiece(input)
	assert.Equal(t, input, string(piece.Piece()))
}

func TestAPiece_String(t *testing.T) {
	t.Parallel()

	// Nothing should change when we call string, so garbage-in-garbage-out tells us everything.
	input := "/y\\..:c"
	piece := APiece(input)
	assert.Equal(t, input, piece.String())
}

func TestAPiece_Len(t *testing.T) {
	t.Parallel()

	// The length of the NAME.
	for _, tc := range []struct{ name, input string }{
		{"empty", ""},
		{"simple", "foo"},
		{"complex", "\\/.:/ \t/foo"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			piece := APiece(tc.input)
			assert.Equal(t, len(tc.input), piece.Len())
		})
	}
}

func TestAPiece_IsAbs(t *testing.T) {
	t.Parallel()

	type Case struct {
		name       string
		input      string
		winSlashes bool
		expected   bool
	}
	for _, tc := range []Case{
		{"empty", "", false, false},
		{"posix+ slash", "/", false, true},
		{"win+ slash", "/", true, true},
		{"posix+ backslash", "\\", false, false},
		{"win+ backslash", "\\", true, true},
		{"posix+ drive", "C:", false, false},
		{"posix+ drive-slash", "C:/", false, false},
		{"win+ drive", "C:", true, false}, // a drive reference is not absolute, but relative.
		{"win+ drive-slash", "C:/", true, true},
		{"win+ drive-backslash", "C:\\", true, true},
		{"posix+ drive-slash-stuff", "C:/Windows", false, false},
		{"win+ drive-slash-stuff", "C:/Windows", true, true},
		{"posix+ bad drive", "yo:/mamma", false, false},
		{"win+ bad drive", "yo:/mamma", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := WithWindowsSlashesSetTo(tc.winSlashes, func() bool {
				return NewAPiece(tc.input).IsAbs()
			})
			assert.Equal(t, tc.expected, result, fmt.Sprintf("input: `%s`, piece: `%s`", tc.input, NewAPiece(tc.input).String()))
		})
	}
}

func TestApiece_ToNative(t *testing.T) {
	// This is kind of the point of go-apathy, this should be about the only place in the tests
	// that we see any kind of platform variance.
	var nativeSep = "/"
	if runtime.GOOS == "windows" {
		nativeSep = "\\"
	}
	for _, tc := range []struct{ name, input, expectedPosix, expectWindows string }{
		{"dot", ".", ".", "."},
		{"simple", "foo", "foo", "foo"},
		{"root", "/", "/", nativeSep},
		{"slash-etc-motd", "/etc/motd", "/etc/motd", nativeSep + "etc" + nativeSep + "motd"},
		{"etc-motd", "etc/motd", "etc/motd", "etc" + nativeSep + "motd"},
		{"c-windows-notepad", "c:/windows/notepad.exe", "c:/windows/notepad.exe", "c:\\windows\\notepad.exe"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			piece := APiece(tc.input)
			withPosix := WithWindowsSlashesSetTo(false, func() string {
				return piece.ToNative()
			})
			withWin := WithWindowsSlashesSetTo(true, func() string {
				return piece.ToNative()
			})
			assert.Equal(t, tc.expectedPosix, withPosix, "posix native gave wrong result")
			assert.Equal(t, tc.expectWindows, withWin, "windows native gave wrong result")
		})
	}
}
