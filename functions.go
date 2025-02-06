// Free-standing functions.
package apathy

import (
	"path"
	"strings"
)

const Dot = APiece(".")

// Normalize will return a windows-separated representation of the piece
// either always (windows), or if the piece has a drive prefix (posix).
func Normalize(piece Piecer) string {
	return piece.Piece().Normalize()
}

// GetAwd is simply the APiece variant of the os.Getwd() function.
func GetAwd() (APiece, error) {
	pwd, err := Getwd()
	if err != nil {
		return "", err
	}
	return NewAPiece(pwd), nil
}

// Join implements path.Join for one or more path components,
// into a posix-styled, Clean()d string.
// TODO: We should be able to leverage our knowledge about the state
// of 'Piece's to implement our own, faster algorithm.
func Join[T ~string](pieces ...T) APiece {
	if len(pieces) == 0 {
		return Dot
	}
	// The new string is going to be all the existing characters
	// plus the separators between them.
	size := len(pieces) - 1 /* one less separator than pieces */
	for _, piece := range pieces {
		size += len(piece)
	}
	if size == 0 {
		return Dot
	}

	buf := append(make([]byte, 0, size), pieces[0]...)
	for _, piece := range pieces[1:] {
		str := string(piece) // demote Pieces etc to strings.
		buf = append(buf, '/')
		buf = append(buf, str...)
	}
	// Use the APiece constructor to ensure the path is clean and posix-styled.
	return NewAPiece(string(buf))
}

func ToSlash[Str ~string](path Str) string {
	return strings.Map(func(r rune) rune {
		if r == '\\' {
			return '/'
		}
		return r
	}, string(path))
}

func Base(piece Piecer) APiece {
	return APiece(path.Base(piece.Piece().String()))
}

func Dir(piecer Piecer) APiece {
	// Windows paths such as 'x:' and 'x:/' need to return 'x:/' as their
	// path.
	piece := piecer.Piece()
	switch piece.Len() {
	case 2: // e.g c:
		if hasDriveLetter(piece) {
			return APiece(piece.String() + "/")
		}
	case 3: // e.g. c:/
		if hasDriveLetter(piece) {
			// c:/ -> c:/
			if piece.String()[2] == '/' {
				return piece
			}
			// Anything else is just c:
			return APiece(piece.String()[:2])
		}
	}
	result := APiece(path.Dir(piece.String()))
	// x:something's parent is x:, rather than .
	if result == "." && hasDriveLetter(piece) {
		result = APiece(piece.String()[:2])
	}
	return result
}

func Ext(piece Piecer) APiece {
	return APiece(path.Ext(piece.Piece().String()))
}
