// Interfaces
package apathy

// PieceMeal is any type that can present an APiece rendering of itself, i.e. a
// path.Clean() representation of some path component.
type PieceMeal interface {
	Piece() APiece
}

// ToNative will return a native representation of any Piece-type object where
// we know that it is currently in posix-style.
func ToNative(path PieceMeal) string {
	return path.Piece().ToNative()
}
