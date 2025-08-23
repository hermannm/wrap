// Package wrap provides utility functions to wrap errors with extra context in an easy-to-read
// format.
package wrap

import (
	"fmt"

	"hermannm.dev/wrap/internal"
)

// Error wraps the given error with a message, to add context to the error.
//
// If you're in a function with a [context.Context] parameter, consider using
// [hermannm.dev/wrap/ctxwrap.Error] instead. See the [hermannm.dev/wrap/ctxwrap] package docs for
// why you may want to do this.
//
// The returned error implements the Unwrap method from the standard [errors] package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	err := errors.New("expired token")
//	wrapped := wrap.Error(err, "user authentication failed")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	user authentication failed
//	- expired token
//
// Wrapped errors can be nested. Wrapping an already wrapped error adds it to the error list, so
// this next example:
//
//	err := errors.New("expired token")
//	inner := wrap.Error(err, "user authentication failed")
//	outer := wrap.Error(inner, "failed to update username")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	failed to update username
//	- user authentication failed
//	- expired token
func Error(wrapped error, message string) error {
	return wrappedError{wrapped, message}
}

// Errorf wraps the given error with a formatted message, to add context to the error. It forwards
// the given message format and args to [fmt.Sprintf] to construct the message.
//
// If you're in a function with a [context.Context] parameter, consider using
// [hermannm.dev/wrap/ctxwrap.Errorf] instead. See the [hermannm.dev/wrap/ctxwrap] package docs for
// why you may want to do this.
//
// The returned error implements the Unwrap method from the standard [errors] package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	err := errors.New("username already taken")
//	wrapped := wrap.Errorf(err, "failed to create user with name '%s'", "hermannm")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	failed to create user with name 'hermannm'
//	- username already taken
func Errorf(wrapped error, messageFormat string, formatArgs ...any) error {
	return wrappedError{wrapped, fmt.Sprintf(messageFormat, formatArgs...)}
}

// Errors wraps the given errors with a message, to add context to the errors.
//
// If you're in a function with a [context.Context] parameter, consider using
// [hermannm.dev/wrap/ctxwrap.Errors] instead. See the [hermannm.dev/wrap/ctxwrap] package docs for
// why you may want to do this.
//
// The returned error implements the Unwrap method from the standard errors package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	err1 := errors.New("username too long")
//	err2 := errors.New("invalid email")
//	wrapped := wrap.Errors("user creation failed", err1, err2)
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	user creation failed
//	- username too long
//	- invalid email
//
// When combined with [wrap.Error], nested wrapped errors are indented, so this next example:
//
//	err1 := errors.New("username too long")
//	err2 := errors.New("invalid email")
//	inner := wrap.Errors("user creation failed", err1, err2)
//	outer := wrap.Error(inner, "failed to register new user")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	failed to register new user
//	- user creation failed
//	  - username too long
//	  - invalid email
func Errors(message string, wrapped ...error) error {
	return wrappedErrors{wrapped, message}
}

type wrappedError struct {
	wrapped error
	message string
}

func (err wrappedError) Error() string {
	return internal.BuildWrappedErrorString(err)
}

// Unwrap matches the signature for wrapped errors expected by the [errors] package.
func (err wrappedError) Unwrap() error {
	return err.wrapped
}

// WrappingMessage implements [hermannm.dev/devlog/log.hasWrappingMessage] for log message
// formatting.
//
// [hermannm.dev/devlog/log.hasWrappingMessage]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedError) WrappingMessage() string {
	return err.message
}

type wrappedErrors struct {
	wrapped []error
	message string
}

func (err wrappedErrors) Error() string {
	return internal.BuildWrappedErrorsString(err)
}

// Unwrap matches the signature for wrapped errors expected by the [errors] package.
func (err wrappedErrors) Unwrap() []error {
	return err.wrapped
}

// WrappingMessage implements [hermannm.dev/devlog/log.hasWrappingMessage] for log message
// formatting.
//
// [hermannm.dev/devlog/log.hasWrappingMessage]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedErrors) WrappingMessage() string {
	return err.message
}
