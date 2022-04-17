package taily

import (
	nativeErrors "errors"

	"github.com/juju/errors"
)

// isError checks if err was caused by otherError.
func IsError(err error, otherError error) bool {
	cause := errors.Cause(err)

	return nativeErrors.Is(cause, otherError)
}
