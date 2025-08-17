package errors

import (
	cockroach "errors"
)

func PropagateError(err error) error {
	if err == nil {
		return nil
	}

	// Already an ApplicationError - return as-is
	var applicationError ApplicationError
	if cockroach.As(err, &applicationError) {
		return err
	}

	// Raw error - convert to infrastructure error
	return NewInfrastructureError(err.Error(), err)
}
