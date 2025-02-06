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
	// NewAPiece should always transform backslashes to slash and - unless this
	// is a drive-root-reference, e.g. x:/ or c:, running a clean. We don't need
	// to extensively test path.Clean here.
	for _, tc := range []apieceTestCase{
		{"empty", "", "."},
		{"non-slashes", "foo", "foo"},
		{"cleanable", "./../foo/bar/../.", "../foo"},
		{"backslash", "\\", "/"},
		{"mixed foo", "foo\\bar/.\\..", "foo"},
		{"mixed foo/bar", "foo/ignore/.\\baz/..\\../bar\\baz/..", "foo/bar"},
		{"drive", "T:", "T:"},
		{"driveroot", "x:\\", "x:/"},
		{"posixdriveroot", "x:/", "x:/"},
		{"notepad", "c:\\windows\\notepad.exe", "c:/windows/notepad.exe"},
		{"notemix", "c:/windows\\notepad.exe", "c:/windows/notepad.exe"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := NewAPiece(tc.input)
			assert.Equal(t, tc.expected, result.String())
		})

	}
}

func TestAPiece_Piece(t *testing.T) {
	t.Parallel()

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
	// We're not going to test cases where the piece has invalid data that
	// doesn't meet our expectations. The string should be posix-styled.
	t.Parallel()

	type Case struct {
		name       string
		input      string
		expected   bool
	}
	for _, tc := range []Case{
		{"empty", "", false},
		{"slash", "/", true},
		{"slash-something", "/foo/bar/slurp", true},
		{"c:", "C:", false},
		{"x:", "x:", false},
		{"drive-with-slash", "u:/", true},
		{"drive-slash-stuff", "C:/Windows", true},
		{"named drive", "yo:/mamma", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := NewAPiece(tc.input).IsAbs()
			assert.Equal(t, tc.expected, result, fmt.Sprintf("input: `%s`, piece: `%s`", tc.input, NewAPiece(tc.input).String()))
		})
	}
}

func TestApiece_Normalize(t *testing.T) {
	// This is kind of the point of go-apathy, this should be about the only place in the tests
	// that we see any kind of platform variance.
	var nativeSep = "/"
	if runtime.GOOS == "windows" {
		nativeSep = "\\"
	}
	for _, tc := range []struct{ name, input, expected string }{
		{"dot", ".", "."},
		{"simple", "foo", "foo"},
		{"root", "/", nativeSep},
		{"slash-etc-motd", "/etc/motd", nativeSep + "etc" + nativeSep + "motd"},
		{"etc-motd", "etc/motd", "etc" + nativeSep + "motd"},
		{"c-windows-notepad", "c:/windows/notepad.exe", "c:\\windows\\notepad.exe"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			normalized := APiece(tc.input).Normalize()
			assert.Equal(t, normalized, tc.expected)
		})
	}
}

func Test_hasDriveLetter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		result bool
		cases  []string
	}{
		{true, []string{"a:", "c:", "A:", "C:", "z:", "Z:", "C:\\windows", "c:/", "w:/"}},
		{false, []string{"", ":", "::", " ", " :", "c :", "c/:", "c\\:", "1:", ";:", "c;", "cd:", "http://", "https://"}},
	} {
		t.Run(fmt.Sprintf("result=%t", tt.result), func(t *testing.T) {
			for _, c := range tt.cases {
				t.Run("case=`"+c+"`", func(t *testing.T) {
					assert.Equal(t, tt.result, hasDriveLetter(APiece(c)))
				})
			}
		})
	}
}
