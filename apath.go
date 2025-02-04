package apathy

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

// APath guarantees a path string is absolute and in posix-separated notation. APaths
// are constructed with Lstat info so we also know whether the path refered to an
// extant item, and if so whether it was either a file, directory, or symlink, and
// it's size/mtime.
//
// To refresh the metadata, use see Observe()/ObserveWithInfo.
type APath struct {
	APiece
	atype APathType
	mtime time.Time // the last-modified time
	size  int64     // the size
}

// APathFromLStat forms an absolute path and then performs an Lstat on it to capture the
// filesystem's metadata for that path. Use APathFromInfo when you have the info already.
func APathFromLstat(pieces ...APiece) (*APath, error) {
	abspath, err := resolvePieces(pieces...)
	if err != nil {
		return nil, err
	}

	apath := APath{abspath, ANotExist, time.Time{}, 0}
	err = apath.Observe()
	return &apath, err
}

// APathFromInfo forms an absolute path based on the given path component(s), and uses the
// provided fs.FileInfo to set the filesystem attributes of the path. This is useful when
// you are walking a filesystem and have already obtained the fs.FileInfo for the path.
func APathFromInfo(info fs.FileInfo, infoErr error, pieces ...APiece) (*APath, error) {
	abspath, err := resolvePieces(pieces...)
	if err != nil {
		return nil, err
	}
	apath := APath{abspath, ANotExist, time.Time{}, 0}
	err = apath.ObserveWithInfo(info, infoErr)
	return &apath, err
}

func (p *APath) Piece() APiece {
	return p.APiece
}
func (p *APath) String() string {
	return string(p.APiece)
}

func (p *APath) Type() APathType {
	return p.atype
}
func (p *APath) Mtime() time.Time {
	return p.mtime
}
func (p *APath) Size() int64 {
	return p.size
}

// Yes, it always is.
func (p *APath) IsAbs() bool {
	return true
}

// Exists returns true if our last Lstat of the filesystem object did not produce a NotExist error.
// Use Observe() to refresh.
func (p *APath) Exists() bool {
	return p.atype != ANotExist
}

// IsSymlink returns true if the last LStat of the filesystem object found a symbolink link,
// (an APath can only be one of file, directory, or symlink).
func (p *APath) IsSymlink() bool {
	return p.atype == ATypeSymlink
}

// IsFile returns true if the last LStat of the filesystem object found a regular file.
func (p *APath) IsFile() bool {
	return p.atype == ATypeFile
}

// IsDir returns true if the last LStat of the filesystem object found a regular directory.
func (p *APath) IsDir() bool {
	return p.atype == ATypeDir
}

// Base returns the last component of the path.
func (p *APath) Base() APiece {
	return APiece(path.Base(p.APiece.String()))
}

// Dir returns the directory component of the path.
func (p *APath) Dir() APiece {
	return APiece(path.Dir(p.APiece.String()))
}

// Ext returns the file extension of the path.
func (p *APath) Ext() APiece {
	return APiece(path.Ext(p.APiece.String()))
}

// resolvePieces will combine several pieces into an absolute path.
func resolvePieces(pieces ...APiece) (APiece, error) {
	if len(pieces) == 0 {
		return "", errors.New("missing path components in call")
	}
	// If the first path is not absolute, we need to prepend the current working directory.
	if !filepath.IsAbs(pieces[0].String()) {
		pwd, err := GetAwd()
		if err != nil {
			return "", err
		}
		pieces = append([]APiece{pwd}, pieces...)
	}
	if len(pieces) == 1 {
		return pieces[0], nil
	}
	fullpath := Join(pieces...)
	if len(fullpath) == 0 {
		return "", errors.New("empty path after joining components")
	}
	return fullpath, nil
}

// Observe whether the file exists and capture its lstat data. Note that if
// LStat returns a NotExist error, we treat this as data and return no error.
func (p *APath) Observe() error {
	return p.ObserveWithInfo(os.Lstat(p.APiece.String()))
}

// ObserveWithInfo refreshes the cached filesystem information for the path
// based on whether err is an ErrNotExist and the provided FileInfo data.
// Call this instead of Observe when you already have the fs.FileInfo.
func (p *APath) ObserveWithInfo(info fs.FileInfo, err error) error {
	p.atype, err = fileInfoToAPathType(info, err)
	switch p.atype {
	case ATypeDir:
		p.mtime = info.ModTime()
		p.size = 0
		return nil
	case ATypeFile:
		p.mtime = info.ModTime()
		p.size = info.Size()
		return nil
	default:
		p.mtime = time.Time{}
		p.size = 0
		return err
	}
}
