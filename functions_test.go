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
	// Normalize is just a wrapper for Piece.Normalize, so we're just testing
	// basic pass-through.

	// Something that shouldn't change:
	piece := APiece("C:/Windows/System16")
	assert.Equal(t, "C:\\Windows\\System16", Normalize(piece))
}
