package models

import (
	"errors"
	"fmt"
)

var (
	// ErrTypeAssertionError is thrown when type an interface does not match the asserted type
	ErrTypeAssertionError = errors.New("unable to assert type")
)

// ParsingError indicates that an error has occurred when parsing request parameters
type ParsingError struct {
	Param string
	Err   error
}

func (e *ParsingError) Unwrap() error {
	return e.Err
}

func (e *ParsingError) Error() string {
	if e.Param == "" {
		return e.Err.Error()
	}

	return e.Param + ": " + e.Err.Error()
}

// RequiredError indicates that an error has occurred when parsing request parameters
type RequiredError struct {
	Field string
}

func (e *RequiredError) Error() string {
	return fmt.Sprintf("required field '%s' is zero value.", e.Field)
}
