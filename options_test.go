package apathy

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var slashOptionMutex sync.Mutex

func WithWindowsSlashesSetTo[RetType any](state bool, call func() RetType) RetType {
	slashOptionMutex.Lock()
	defer slashOptionMutex.Unlock()
	priorState := ExpectWindowsSlashes
	defer func() { ExpectWindowsSlashes = priorState }()
	ExpectWindowsSlashes = state
	return call()
}

func TestExpectWindowsSlashes_State(t *testing.T) {
	const isWindows = (runtime.GOOS == "windows")
	assert.Equal(t, isWindows, ExpectWindowsSlashes, "ExpectWindowsSlashes should default to true on windows and windows only")
}

func TestWithPossibleWindowsSlashes(t *testing.T) {
	priorState := ExpectWindowsSlashes
	defer func() { ExpectWindowsSlashes = priorState }()

	child := func() {
		defer WithPossibleWindowsSlashes()()
		assert.True(t, ExpectWindowsSlashes, "ExpectWindowsSlashes should be true within the child")
	}

	for _, testState := range []bool{false, true} {
		t.Run(fmt.Sprintf("state=%t", testState), func(t *testing.T) {
			ExpectWindowsSlashes = testState
			child()
			assert.Equal(t, testState, ExpectWindowsSlashes, "ExpectWindowsSlashes should be false after the child")
		})
	}
}
