package guardlog

import (
	nativeErrors "errors"

	"github.com/juju/errors"
)

func isError(err error, otherError error) bool {
	cause := errors.Cause(err)

	return nativeErrors.Is(cause, otherError)
}
