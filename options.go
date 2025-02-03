package apathy

import "runtime"

// ExpectWindowsSlashes can be set to true to tell APiece to clean for windows slashes,
// even if the runtime is not Windows. This is useful when you anticipate that some of
// your inputs (files, network, config, command line, etc) might have winslashes.
//
// Only affects newly-created APieces. See also WithPossibleWindowsSlashes.
var ExpectWindowsSlashes = runtime.GOOS == "windows"

// WithPossibleWindowsSlashes sets ExpectWindowsSlashes to true and returns a capture
// that will reset it to its original state, so you can temporarily enable it with defer.
//
// Example:
//
//	defer WithPossibleWindowsSlashes()()
//	return apathy.NewAPiece("C:\\Windows\\System32/hosts\\etc")
func WithPossibleWindowsSlashes() func() {
	var priorState bool
	priorState, ExpectWindowsSlashes = ExpectWindowsSlashes, true

	return func() {
		ExpectWindowsSlashes = priorState
	}
}
