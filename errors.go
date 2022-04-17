package taily

import (
	nativeErrors "errors"

	"github.com/juju/errors"
)

// isErrors checks if err was caused by otherError.
func isError(err error, otherError error) bool {
	cause := errors.Cause(err)

	return nativeErrors.Is(cause, otherError)
}
