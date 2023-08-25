package errors

import "fmt"

type ErrorType string

var (
	ErrorTypeNotFound ErrorType = "not_found"
)

type Error struct {
	message string
	eType   ErrorType
}

func NewError(t ErrorType, message string, args ...any) *Error {
	return &Error{
		message: fmt.Sprintf(message, args...),
		eType:   t,
	}
}

func NewNotFoundError(message string, args ...any) *Error {
	return NewError(ErrorTypeNotFound, message, args...)
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.eType, e.message)
}

func IsNotFound(err error) bool {
	return errIs(err, ErrorTypeNotFound)
}

func errIs(err error, t ErrorType) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*Error); ok {
		return e.eType == t
	}

	return false
}
