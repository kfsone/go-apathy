package apathy

import (
	"io/fs"
	"os"
)

// APathType tells us what we discovered about a directory item when we last Lstat()ed it,
// either that it did not exist, or distinguishing between file/directory/symlink/other.
type APathType uint32

const (
	ANotExist    APathType = iota // ANotExist indicates the file was not found when last Lstat()d.
	ATypeFile                     // ATypeFile indicates the last Lstat() of the path found a regular file.
	ATypeDir                      // ATypeDir indicates the last LStat() of the path found a regular directory.
	ATypeSymlink                  // ATypeSymlink indicates the last LStat() of the path found a symbolic link.
	ATypeUnknown
)

func (a APathType) String() string {
	switch a {
	case ANotExist:
		return "NotExist"
	case ATypeFile:
		return "File"
	case ATypeDir:
		return "Dir"
	case ATypeSymlink:
		return "Symlink"
	default:
		return "Unknown"
	}
}

// fileInfoToAPathType converts a fs.FileInfo to an APathType by inspecting
// the mode of the file. If the fs.FileInfo is nil, it returns ANotExist.
// Note that APathType considers symlinks a distinct type separate from
// files and directories, as this makes life easier on Windows in most cases.
func fileInfoToAPathType(info fs.FileInfo, err error) (APathType, error) {
	if err != nil {
		if !os.IsNotExist(err) {
			return ANotExist, err
		}
		return ANotExist, nil
	}
	mode := info.Mode()
	switch {
	case mode.IsRegular():
		return ATypeFile, nil
	case mode.Type()&fs.ModeSymlink != 0:
		return ATypeSymlink, nil
	case mode.IsDir():
		return ATypeDir, nil
	default:
		return ATypeUnknown, nil
	}
}
