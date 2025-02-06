package apathy

import (
	"errors"
	"fmt"
)

var (
	// ErrInternal denotes an implementation error rather than anything a user can control.
	ErrInternal = errors.New("internal error")
	// ErrMissingArgs is an internal error caused by passing insufficient parameters to a variadic method.
	ErrMissingArgs = fmt.Errorf("%w: missing arguments", ErrInternal)
)
