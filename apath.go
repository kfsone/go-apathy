package apathy

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

// aPath is the underlying implementation of the APath interface.
type aPath struct {
	APiece
	aType APathType
	mtime time.Time
	size  int64
}

// NewAPath forms an absolute path and then performs an Lstat on it to capture the
// filesystem's metadata for that path. Use NewAPathWith when you have the info already.
func NewAPath(pieces ...APiece) (APath, error) {
	absPath, err := resolvePieces(pieces...)
	if err != nil {
		return nil, err
	}

	lstat, err := Lstat(absPath.String())
	return newAPathWith(absPath, lstat, err)
}

// NewAPathWith forms an absolute path based on the given path component(s), and uses the
// provided fs.FileInfo to set the filesystem attributes of the path. This is useful when
// you are walking a filesystem and have already obtained the fs.FileInfo for the path.
// We only take a single piece because if you just went and did an lstat, you must have
// assembled the path to pass to lstat.
func NewAPathWith(path APiece, info fs.FileInfo, infoErr error) (APath, error) {
	if !path.IsAbs() {
		panic(fmt.Errorf("%w: expected absolute path: %s", ErrInternal, path))
	}
	return newAPathWith(path, info, infoErr)
}

func newAPathWith(absolutePath APiece, info fs.FileInfo, err error) (APath, error) {
	// We've made them pass us the error so we can discriminate NotExists for the caller.
	if !absolutePath.IsAbs() {
		panic(fmt.Errorf("%w: non-absolute path leaked: %s", ErrInternal, absolutePath))
	}
	if err != nil {
		if !os.IsNotExist(err) {
			// Anything else is unrecoverable.
			return nil, err
		}
		// Fine, we'll represent a file that does not exist.
		return &aPath{APiece: absolutePath}, nil
	}
	// It exists, let's look and see what it is
	var aType APathType
	var mode = info.Mode().Type()
	switch {
	case mode.IsRegular():
		aType = ATypeFile
	case mode.IsDir():
		aType = ATypeDir
	case mode&fs.ModeSymlink != 0:
		aType = ATypeSymlink
	default:
		aType = ATypeUnknown
	}

	return &aPath{APiece: absolutePath, aType: aType, mtime: info.ModTime(), size: info.Size()}, nil
}

func (p *aPath) Piece() APiece {
	return p.APiece
}
func (p *aPath) String() string {
	return string(p.APiece)
}

func (p *aPath) Type() APathType {
	return p.aType
}
func (p *aPath) ModTime() time.Time {
	return p.mtime
}
func (p *aPath) Size() int64 {
	return p.size
}
func (p *aPath) IsAbs() bool {
	return true
}

// Exists returns true if our last Lstat of the filesystem object did not produce a NotExist error.
// Use Observe() to refresh.
func (p *aPath) Exists() bool {
	return p.aType != ANotExist
}

// IsSymlink returns true if the last LStat of the filesystem object found a symbolink link,
// (an APath can only be one of file, directory, or symlink).
func (p *aPath) IsSymlink() bool {
	return p.aType == ATypeSymlink
}

// IsFile returns true if the last LStat of the filesystem object found a regular file.
func (p *aPath) IsFile() bool {
	return p.aType == ATypeFile
}

// IsDir returns true if the last LStat of the filesystem object found a regular directory.
func (p *aPath) IsDir() bool {
	return p.aType == ATypeDir
}

// Base returns the last component of the path.
func (p *aPath) Base() APiece {
	return APiece(path.Base(p.APiece.String()))
}

// Dir returns the directory component of the path.
func (p *aPath) Dir() APiece {
	return APiece(path.Dir(p.APiece.String()))
}

// Ext returns the file extension of the path.
func (p *aPath) Ext() APiece {
	return APiece(path.Ext(p.APiece.String()))
}

// resolvePieces will combine several pieces into an absolute path.
func resolvePieces(pieces ...APiece) (APiece, error) {
	if len(pieces) == 0 {
		panic(fmt.Errorf("%w: resolvePieces requires at least one APiece", ErrMissingArgs))
	}
	fullPath := Join(pieces...).String()
	fullPath, err := Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("error resolving path: %w", err)
	}
	// filepath on windows will introduce native separators
	fullPath = filepath.ToSlash(fullPath)
	return APiece(fullPath), nil
}
