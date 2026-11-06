package accountsecurity

import (
	"errors"
)

type ValidationError struct{ message string }

func (e *ValidationError) Error() string { return e.message }

func validationError(message string) error { return &ValidationError{message: message} }

func ClientErrorMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	if errors.Is(err, ErrCurrentPasswordInvalid) || errors.Is(err, ErrPasswordReused) {
		return err.Error()
	}
	var validation *ValidationError
	if errors.As(err, &validation) {
		return validation.Error()
	}
	return fallback
}
