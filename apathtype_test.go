package apathy

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockFileInfo implements fs.FileInfo for testing
type mockFileInfo struct {
	name  string
	size  int64
	mode  fs.FileMode
	mtime time.Time
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.mtime }
func (m mockFileInfo) IsDir() bool        { return m.mode.IsDir() }
func (m mockFileInfo) Sys() interface{}   { return nil }

func Test_fileInfoToAPathType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		info     fs.FileInfo
		err      error
		wantType APathType
		wantErr  bool
		errIs    error // expected specific error
	}{
		{
			name:     "regular file",
			info:     mockFileInfo{mode: 0644},
			wantType: ATypeFile,
		},
		{
			name:     "directory",
			info:     mockFileInfo{mode: fs.ModeDir},
			wantType: ATypeDir,
		},
		{
			name:     "symlink",
			info:     mockFileInfo{mode: fs.ModeSymlink},
			wantType: ATypeSymlink,
		},
		{
			name:     "special file (device)",
			info:     mockFileInfo{mode: fs.ModeDevice},
			wantType: ATypeUnknown,
		},
		{
			name:     "special file (named pipe)",
			info:     mockFileInfo{mode: fs.ModeNamedPipe},
			wantType: ATypeUnknown,
		},
		{
			name:     "nil FileInfo",
			info:     nil,
			err:      os.ErrNotExist,
			wantType: ANotExist,
		},
		{
			name:     "not exist error",
			info:     mockFileInfo{},
			err:      os.ErrNotExist,
			wantType: ANotExist,
		},
		{
			name:     "permission error",
			info:     mockFileInfo{},
			err:      os.ErrPermission,
			wantType: ANotExist,
			wantErr:  true,
			errIs:    os.ErrPermission,
		},
		{
			name:     "generic error",
			info:     mockFileInfo{},
			err:      errors.New("something went wrong"),
			wantType: ANotExist,
			wantErr:  true,
		},
		{
			name:     "symlink with extra bits",
			info:     mockFileInfo{mode: fs.ModeSymlink | fs.ModeDir},
			wantType: ATypeSymlink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, err := fileInfoToAPathType(tt.info, tt.err)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantType, gotType)
		})
	}
}

func TestAPathType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		info APathType
		name string
	}{
		{ANotExist, "NotExist"},
		{ATypeFile, "File"},
		{ATypeDir, "Dir"},
		{ATypeSymlink, "Symlink"},
		{ATypeUnknown, "Unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.name, tc.info.String())
		})
	}
}
