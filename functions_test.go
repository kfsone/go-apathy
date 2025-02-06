package apathy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAwd_Error(t *testing.T) {
	defer withSaved(&Getwd, func() (string, error) {
		return "", os.ErrNotExist
	})()
	_, err := GetAwd()
	assert.NotNil(t, err)
}

func TestGetAwd(t *testing.T) {
	executablePiece := NewAPiece(myExeFolder)

	defer withSaved(&Getwd, func() (string, error) {
		return executablePiece.String(), nil
	})()

	awd, err := GetAwd()
	assert.Nil(t, err, "got an error calling GetAwd()")
	assert.Equal(t, executablePiece, awd)
}

func TestJoin(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		inputs    []string
		expecting string
	}{
		{"no parts", []string{}, "."},
		{"one empty", []string{""}, "."},
		{"dot", []string{"."}, "."},
		{"dot, dot", []string{".", "."}, "."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected := APiece(tc.expecting)
			actual := Join(tc.inputs...)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()

	// Normalize is just a wrapper for Piece.Normalize, so we're just testing
	// basic pass-through.

	// Something that shouldn't change:
	piece := APiece("C:/Windows/System16")
	assert.Equal(t, "C:\\Windows\\System16", Normalize(piece))
}

func TestAPieceHelpers(t *testing.T) {
	t.Parallel()

	a := &aPath{"/usr/lib/postgres/fire.theres_actual_fire", ANotExist, fixedTime, 0}
	assert.Equal(t, "fire.theres_actual_fire", Base(a).String())
	assert.Equal(t, "/usr/lib/postgres", Dir(a).String())
	assert.Equal(t, ".theres_actual_fire", Ext(a).String())
}

func TestAPiece_Dir(t *testing.T) {
	t.Parallel()
	// The windows paths x: and x:/anything need to return x:/ as their parent.
	for _, tc := range [][]APiece {
		{".", "."},
		{".", "a"},
		{".", "ab"},	// test len==2 case
		{".", "abc"},	// test len==3 case
		{"a", "a/b"}, 
		{"T:/", "T:/"},
		{"T:", "T:"},
		{"A:", "A:."},
		{"B:", "B:x"},
		{"u:/", "u:/"},
		{"S:", "S:xyz"},
		{"c:/", "c:/windows"},
		{"c:windows/system32", "c:windows/system32/drivers"},
		{"/", "/"},
		{"/etc/apt", "/etc/apt/apt.d"},
	} {
		t.Run(tc[0].String(), func (t *testing.T) {
			parent := Dir(tc[1])
			assert.Equal(t, tc[0], parent)
		})
	}
}
