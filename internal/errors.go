package internal

import "fmt"

type Error struct {
	orig error
	msg  string
	code ErrorCode
}

type ErrorCode uint

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeNotFound
	ErrCodeInvalidArgument
)

func WrapErrorf(orig error, code ErrorCode, format string, args ...interface{}) error {
	return &Error{
		orig: orig,
		msg:  fmt.Sprintf(format, args...),
		code: code,
	}
}

// NewErrorf instantiates a new error.
func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapErrorf(nil, code, format, a...)
}

// Error returns the message, when wrapping errors the wrapped error is returned.
func (e *Error) Error() string {
	if e.orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.orig)
	}

	return e.msg
}

// Unwrap returns the wrapped error, if any.
func (e *Error) Unwrap() error {
	return e.orig
}

// Code returns the code representing this error.
func (e *Error) Code() ErrorCode {
	return e.code
}
