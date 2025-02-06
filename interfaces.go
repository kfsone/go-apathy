// Interfaces
package apathy

import (
	"os"
	"path/filepath"
	"time"
)

// APath guarantees a path string is absolute and in posix-separated notation. APaths
// are constructed with Lstat info so we also know whether the path refered to an
// extant item, and if so whether it was either a file, directory, or symlink, and
// it's size/mtime.
//
// To refresh the metadata, use see Observe()/ObserveWithInfo.
type APath interface {
	Exists
	IsDir
	IsFile
	IsSymlink
	ModTimed
	Sized
	Normalizer
	Piecer
	Stringer
	Type() APathType
}

type Exists interface {
	Exists() bool
}
type IsDir interface {
	IsDir() bool
}
type IsFile interface {
	IsFile() bool
}
type IsSymlink interface {
	IsSymlink() bool
}
type ModTimed interface {
	ModTime() time.Time
}
type Sized interface {
	Size() int64
}
type Normalizer interface {
	Normalize() string
}
type Piecer interface {
	Piece() APiece
}
type Stringer interface {
	String() string
}
type Lengthed interface {
	Len() int
}

// For mocking/testing/etc.
var Abs = filepath.Abs
var Getwd = os.Getwd
var Lstat = os.Lstat
