// Package mid provides app level middleware support.
package mid

import (
	"errors"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/foundation/web"
)

// isError tests if the Encoder has an error inside of it.
func isError(e web.Encoder) error {
	err, isError := e.(error)
	if isError {
		var appErr *errs.Error
		if errors.As(err, &appErr) && appErr.Code == errs.None {
			return nil
		}
		return err
	}
	return nil
}
