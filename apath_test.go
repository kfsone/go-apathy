package apathy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fixedTime = time.Now()

const onWindows = runtime.GOOS == "windows"

// Capture the path to the executable so we can test against it later.
var myExecutable string
var myExeFolder string
var myExeFile string

func withSaved[T any](thing *T, newValue T) func() {
	oldValue := *thing
	*thing = newValue
	return func() {
		*thing = oldValue
	}
}

func init() {
	exe, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("couldn't get the path of the executable: %w", err))
	}
	absExe, err := Abs(exe)
	if err != nil {
		panic(fmt.Errorf("couldn't resolve the path of the executable: %w", err))
	}

	myExecutable = absExe

	myExeFolder = filepath.Dir(myExecutable)
	myExeFile = filepath.Base(myExecutable)
}

func Test_aPath_Accessors(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		name      string
		apath     aPath
		wantStr   string
		wantType  APathType
		wantTime  time.Time
		wantSize  int64
		exists    bool
		isDir     bool
		isFile    bool
		isSymlink bool
	}
	for _, tc := range []TestCase{
		{name: "defaulted", apath: aPath{APiece: "x/y/z"}, wantStr: "x/y/z", wantType: ANotExist},
		{
			name:    "file",
			apath:   aPath{APiece: "a.file", aType: ATypeFile, mtime: fixedTime, size: 342},
			wantStr: "a.file", wantType: ATypeFile, wantTime: fixedTime, wantSize: 342,
			exists: true, isFile: true,
		},
		{
			name:    "dir",
			apath:   aPath{APiece: "my/dir", aType: ATypeDir, mtime: time.UnixMicro(333), size: 9},
			wantStr: "my/dir", wantType: ATypeDir, wantTime: time.UnixMicro(333), wantSize: 9,
			exists: true, isDir: true,
		},
		{
			name:    "symlink",
			apath:   aPath{APiece: "some/sym/link", aType: ATypeSymlink, mtime: fixedTime, size: 601},
			wantStr: "some/sym/link", wantType: ATypeSymlink, wantTime: fixedTime, wantSize: 601,
			exists: true, isSymlink: true,
		},
	} {
		assert.True(t, tc.apath.IsAbs())
		assert.Equal(t, tc.wantStr, tc.apath.Piece().String())
		assert.Equal(t, tc.wantStr, tc.apath.String())
		assert.Equal(t, tc.wantType, tc.apath.Type())
		assert.Equal(t, tc.wantTime, tc.apath.ModTime())
		assert.Equal(t, tc.wantSize, tc.apath.Size())
		assert.Equal(t, tc.exists, tc.apath.Exists())
		assert.Equal(t, tc.isDir, tc.apath.IsDir())
		assert.Equal(t, tc.isFile, tc.apath.IsFile())
		assert.Equal(t, tc.isSymlink, tc.apath.IsSymlink())
	}
}

func Test_aPath_resolvePieces_error(t *testing.T) {
	var myError = errors.New("disk full")
	defer withSaved(&Abs, func(string) (string, error) {
		return "", myError
	})()
	_, err := resolvePieces(APiece("."))
	assert.ErrorIs(t, err, myError)
}

func Test_aPath_resolvePieces(t *testing.T) {
	t.Parallel()

	// The use of "filepath.Abs" in the underlying code generate paths that
	// depend on the runtime context. So, let's find out what that is and
	// use it accordingly.
	awd, err := GetAwd()
	assert.NoError(t, err)

	t.Run("err on no pieces", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = resolvePieces()
		})
	})
	t.Run("Pwd on empty product", func(t *testing.T) {
		p, err := resolvePieces(APiece(""))
		assert.NoError(t, err)
		assert.Equal(t, awd, p)
	})
	t.Run("Dot with many dot dirs", func(t *testing.T) {
		p, err := resolvePieces(APiece("."), APiece("."), APiece("."))
		assert.NoError(t, err)
		assert.Equal(t, awd, p)
	})

	t.Run("simple string", func(t *testing.T) {
		p, err := resolvePieces(awd.Piece())
		assert.NoError(t, err)
		assert.Equal(t, awd, p)
	})

	t.Run("multiple strings", func(t *testing.T) {
		awdChild := Join(awd.String(), "child")
		p, err := resolvePieces(awd, "child")
		assert.NoError(t, err)
		assert.Equal(t, awdChild, p)
	})

	t.Run("relative component", func(t *testing.T) {
		assert.NoError(t, err)
		p, err := resolvePieces("child", "grandchild")
		assert.NoError(t, err)
		expected := awd.String() + "/child/grandchild"
		assert.Equal(t, expected, p.String())
	})
}

func Test_newAPathWith_PanicRelativeArg(t *testing.T) {
	t.Parallel()
	wantErr := fmt.Errorf("%w: non-absolute path leaked: %s", ErrInternal, ".")
	assert.PanicsWithError(t, fmt.Sprint(wantErr), func() {
		_, err := newAPathWith(Dot, nil, nil)
		assert.NoError(t, err)
	})
}

func Test_newAPathWith_HardError(t *testing.T) {
	t.Parallel()
	fakeErr := errors.New("no cookie for you")
	p, err := newAPathWith("/", nil, fakeErr)
	assert.ErrorIs(t, err, fakeErr)
	assert.Nil(t, p)
}

func Test_newAPathWith_NotExist(t *testing.T) {
	t.Parallel()
	p, err := newAPathWith("/", nil, os.ErrNotExist)
	assert.NoError(t, err, "error being NotExist is not an error for newAPathWith")
	assert.NotNil(t, p)
	assert.Equal(t, p.String(), "/")
	assert.False(t, p.Exists())
}

func Test_newAPathWith_Exists(t *testing.T) {
	t.Parallel()

	var fixedSize int64 = 12345
	mockLstat := func(mode os.FileMode) mockFileInfo {
		return mockFileInfo{mode: mode, mtime: fixedTime, size: fixedSize}
	}

	for _, tc := range []struct {
		name, path string
		info       mockFileInfo
		aType      APathType
	}{
		{"folder", "/foo", mockLstat(os.ModeDir), ATypeDir},
		{"file", "/file", mockLstat(0), ATypeFile},
		{"symlink", "/symlink", mockLstat(os.ModeSymlink), ATypeSymlink},
		{"device", "/device", mockLstat(os.ModeDevice), ATypeUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// deliberately pass it a different path, it shouldn't be looking at the path.
			result, err := newAPathWith("/fiddle/sticks", tc.info, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.aType, result.Type())
			assert.Equal(t, fixedTime, result.ModTime())
			assert.Equal(t, fixedSize, result.Size())
		})
	}
}

func TestNewAPath_PanicsOnZeroPieces(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() {
		_, _ = NewAPath()
	})
}

func TestNewAPath_Errors(t *testing.T) {
	// Can't be parallel because it modifies globals.

	// First, make Abs return a daft root.
	defer withSaved(&Abs, func(child string) (string, error) {
		return "/x/" + child, nil
	})()
	// Make lstat return some other kind of error.
	var lstatErr = errors.New("no cookie for you")
	defer withSaved(&Lstat, func(in string) (os.FileInfo, error) {
		assert.Equal(t, in, "/x/cookie")
		return nil, lstatErr
	})()
	p, err := NewAPath("cookie")
	assert.ErrorIs(t, err, lstatErr)
	assert.Nil(t, p)

	// Now make sure that if Abs returns an error, we handle that
	var absErr = errors.New("lost in space")
	defer withSaved(&Abs, func(string) (string, error) {
		return "", absErr
	})()
	p, err = NewAPath("cookie")
	assert.ErrorIs(t, err, absErr)
	assert.Nil(t, p)
}

func TestNewAPath_Basic(t *testing.T) {
	// I want to go the whole hog and test the real Abs method, and the only
	// things we can reasonably rely on are the path to the executable and
	// the current working directory.
	t.Parallel()

	exePiece := NewAPiece(myExecutable)
	folderPiece := NewAPiece(myExeFolder)
	filePiece := NewAPiece(myExeFile)

	for _, tc := range []struct {
		name        string
		inputs      []APiece
		expectErr   error
		expectPiece APiece
		exists      bool
		isDir       bool
	}{
		{"myexe", []APiece{exePiece}, nil, exePiece, true, false},
		{"exe parent", []APiece{folderPiece}, nil, folderPiece, true, true},
		{"parent + file", []APiece{folderPiece, filePiece}, nil, exePiece, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			aPath, err := NewAPath(tc.inputs...)
			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
				assert.Nil(t, aPath)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectPiece, aPath.Piece())
			}
		})
	}
}

func TestNewAPathWith_PanicOnRelative(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() {
		_, _ = NewAPathWith(Dot, nil, nil)
	})
}

func TestNewAPathWith(t *testing.T) {
	t.Parallel()
	// We only need to test that it is forwarding to newAPathWith, which we test elsewhere.
	apath, err := NewAPathWith("/biscuit/gravy", mockFileInfo{mode: os.ModeSymlink}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, apath)
	assert.True(t, apath.Exists())
	assert.True(t, apath.IsSymlink())
}
