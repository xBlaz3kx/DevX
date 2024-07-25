package errors

import (
	"errors"
	"fmt"
)

type ApiError struct {
	// HTTP status code
	statusCode int

	// Message displayed in the http response
	message string

	// Code used for debugging and comparison
	internalCode int

	// Optional wrapped error
	err error
}

func (e ApiError) StatusCode() int {
	return e.statusCode
}

func (e ApiError) Message() string {
	return e.message
}

func (e ApiError) InternalCode() int {
	return e.internalCode
}

// Error returns the error message
//
// Do NOT use this for returning error messages to the user, because it returns the wrapped errors as well!
// Use the Message field instead.
func (e ApiError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

func (e ApiError) Unwrap() error {
	return e.err
}

func (e ApiError) Is(target error) bool {
	// Check if error is of type apiError
	if t, ok := target.(ApiError); ok {
		return e.internalCode == t.internalCode
	}
	// Else check underlying error
	return errors.Is(e.err, target)
}

func New(internalCode int, statusCode int, message string) ApiError {
	return ApiError{
		internalCode: internalCode,
		statusCode:   statusCode,
		message:      message,
	}
}

func (e ApiError) New(internalCode int, message string) ApiError {
	err := New(internalCode, e.statusCode, message)
	err.err = e
	return err
}
