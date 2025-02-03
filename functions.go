// Free-standing functions.
package apathy

import "os"

// GetAwd is simply the APiece variant of the os.Getwd() function.
func GetAwd() (APiece, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return NewAPiece(pwd), nil
}

// Join implements path.Join for one or more string-based path components,
// esp including APiece values, into a posix-styled, Clean()d string.
// Derived from the standard path.Join function.
func Join[T ~string](pieces ...T) APiece {
	// The new string is going to be all the existing characters
	// plus the separators between them.
	size := len(pieces) - 1 /* one less separator than pieces */
	for _, e := range pieces {
		size += len(e)
	}
	if size == 0 {
		return APiece(".")
	}

	buf := append(make([]byte, 0, size), pieces[0]...)
	for _, e := range pieces[1:] {
		str := string(e) // demote Pieces etc to strings.
		if len(str) == 0 {
			continue
		}
		buf = append(buf, '/')
		buf = append(buf, str...)
	}
	// Use the APiece constructor to ensure the path is clean and posix-styled.
	return NewAPiece(string(buf))
}
