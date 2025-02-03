package apathy

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
