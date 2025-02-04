package apathy

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fixedTime = time.Now()

func TestAPath_Acessors(t *testing.T) {
	t.Parallel()

	// Pass garbage in, if we get the same garbage back, Piece is doing its thing.
	a := APath{"x/y/z", ANotExist, time.Time{}, 0}
	assert.Equal(t, "x/y/z", string(a.Piece()))
	assert.Equal(t, "x/y/z", a.String())

	// defaults
	assert.Equal(t, ANotExist, a.Type())
	assert.Equal(t, time.Time{}, a.Mtime())
	assert.Equal(t, int64(0), a.Size())
	assert.True(t, a.IsAbs()) // demonstrate the blind faith
	assert.False(t, a.Exists())
	assert.False(t, a.IsSymlink())
	assert.False(t, a.IsFile())
	assert.False(t, a.IsDir())

	// value check
	a.atype = ATypeFile
	a.mtime = fixedTime
	a.size = 4352

	a.APiece = APiece("")
	assert.True(t, a.IsAbs())

	a.APiece = APiece(".")
	assert.True(t, a.IsAbs())

	assert.Equal(t, fixedTime, a.Mtime())
	assert.Equal(t, int64(4352), a.Size())

	t.Run("type: file", func(t *testing.T) {
		assert.Equal(t, ATypeFile, a.Type())
		assert.True(t, a.Exists())
		assert.False(t, a.IsSymlink())
		assert.True(t, a.IsFile())
		assert.False(t, a.IsDir())
	})

	t.Run("type: dir", func(t *testing.T) {
		a.atype = ATypeDir
		assert.True(t, a.IsDir())
		assert.True(t, a.Exists())
		assert.False(t, a.IsFile())
		assert.False(t, a.IsSymlink())
	})

	t.Run("type: symlink", func(t *testing.T) {

		a.atype = ATypeSymlink
		assert.True(t, a.IsSymlink())
		assert.True(t, a.Exists())
		assert.False(t, a.IsFile())
		assert.False(t, a.IsDir())
	})
}

func TestAPath_Helpers(t *testing.T) {
	t.Parallel()

	a := APath{"/usr/lib/postgres/fire.theres_actual_fire", ANotExist, time.Time{}, 0}

	assert.Equal(t, "fire.theres_actual_fire", a.Base().String())
	assert.Equal(t, "/usr/lib/postgres", a.Dir().String())
	assert.Equal(t, ".theres_actual_fire", a.Ext().String())
}

func TestAPath_ObserveWithInfo(t *testing.T) {
	type TestCase struct {
		name        string
		filename    string
		info        mockFileInfo
		err         error
		expectErr   error
		expectType  APathType
		expectMtime time.Time
		expectSize  int64
	}
	for _, tc := range []TestCase{
		{name: "bad error", err: os.ErrPermission, expectErr: os.ErrPermission},
		{name: "not exist", err: os.ErrNotExist, expectErr: nil},
		{name: "exists/file", info: mockFileInfo{mode: 0644, mtime: fixedTime, size: 6378}, expectType: ATypeFile, expectMtime: fixedTime, expectSize: 6378},
		{name: "exists/dir", info: mockFileInfo{mode: fs.ModeDir, mtime: fixedTime, size: 7777}, expectType: ATypeDir, expectMtime: fixedTime, expectSize: 0},
		{name: "exists/symlink", info: mockFileInfo{mode: fs.ModeSymlink, mtime: fixedTime, size: 5213}, expectType: ATypeSymlink},
		{name: "exists/device", info: mockFileInfo{mode: 0644 | fs.ModeDevice, mtime: fixedTime, size: 6378}, expectType: ATypeUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &APath{APiece: APiece(tc.filename)}
			err := a.ObserveWithInfo(tc.info, tc.err)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectType, a.Type())
			assert.Equal(t, tc.expectMtime, a.Mtime())
			assert.Equal(t, tc.expectSize, a.Size())
		})
	}
}

func TestAPath_Observe(t *testing.T) {
	// There are only two things we can be sure of that we can Lstat:
	// - this process,
	// - the directory
	// and since all we're really trying to detect is that Lstat is being used,
	awd, err := GetAwd()
	assert.Nil(t, err, "unable to locate current directory")

	a := &APath{APiece: awd}

	dirLstat, err := os.Lstat(awd.String())
	assert.Nil(t, err, "unable to lstat working directory")

	// Ask APath to refresh itself so its loads the info
	err = a.Observe()
	assert.Nil(t, err)
	assert.Equal(t, dirLstat.ModTime(), a.mtime)
}

func TestAPathFromLstat(t *testing.T) {
	// We can rely on the fact that our process is a file and it exists in
	// a directory to give us something meaningful to look at on the filesystem.
	exeFile, err := os.Executable()
	assert.Nil(t, err, "unable to locate current my own executable")
	exeDir := filepath.Dir(exeFile)

	exeAbs, err := filepath.Abs(exeFile)
	assert.Nil(t, err)

	dirPiece := NewAPiece(exeDir)
	dirAbs, err := filepath.Abs(dirPiece.String())
	assert.Nil(t, err)

	// If we give it a file that we are fairly confident won't exist, then
	// we should get a nil error but the file type should be ANotExist.
	t.Run("nosuch file", func(t *testing.T) {
		badFilename := NewAPiece(filepath.Base(exeFile) + ".nosuch.thing")
		apath, err := APathFromLstat(dirPiece, badFilename)
		assert.Nil(t, err)
		assert.NotNil(t, apath)
		assert.False(t, apath.Exists())
		assert.Equal(t, exeAbs+".nosuch.thing", apath.ToNative())
	})

	t.Run("directory", func(t *testing.T) {
		dirStat, err := os.Lstat(exeDir)
		assert.Nil(t, err, "unable to lstat our executable dir")

		apath, err := APathFromLstat(dirPiece)
		assert.Nil(t, err)
		assert.NotNil(t, apath)
		assert.True(t, apath.Exists())
		assert.True(t, apath.IsDir())
		assert.Equal(t, dirStat.ModTime(), apath.Mtime())
		assert.Equal(t, dirAbs, apath.ToNative())
	})

	t.Run("file w/join", func(t *testing.T) {
		exeStat, err := os.Lstat(exeFile)
		assert.Nil(t, err, "unable to lstat our executable")

		apath, err := APathFromLstat(dirPiece, ".", NewAPiece(filepath.Base(exeFile)))
		assert.Nil(t, err)
		assert.NotNil(t, apath)
		assert.True(t, apath.Exists())
		assert.True(t, apath.IsFile())
		assert.Equal(t, exeStat.ModTime(), apath.Mtime())
		assert.Equal(t, exeAbs, apath.ToNative())
	})

	t.Run("file single", func(t *testing.T) {
		exeStat, err := os.Lstat(exeFile)
		assert.Nil(t, err, "unable to lstat our executable")

		exePiece := NewAPiece(exeFile)
		apath, err := APathFromLstat(exePiece)
		assert.Nil(t, err)
		assert.NotNil(t, apath)
		assert.True(t, apath.Exists())
		assert.True(t, apath.IsFile())
		assert.Equal(t, exeStat.ModTime(), apath.Mtime())
		assert.Equal(t, exeAbs, apath.ToNative())
	})
}
