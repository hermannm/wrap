// Package wrap provides utility functions to wrap errors with extra context.
package wrap

import (
	"fmt"
	"log/slog"

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

// ErrorWithAttrs wraps the given error with a message and log attributes, to add structured context
// to the error when it is logged (see below for how to pass attributes).
//
// The returned error implements the following method:
//
//	LogAttrs() []slog.Attr
//
// A logging library can check for the existence of this method when an error is logged, to add
// these attributes to the log output. The [hermannm.dev/devlog/log] library, which wraps
// [log/slog], does this in its error-aware logging functions.
//
// If you're in a function with a [context.Context] parameter, consider using
// [hermannm.dev/wrap/ctxwrap.ErrorWithAttrs] instead. See the [hermannm.dev/wrap/ctxwrap] package
// docs for why you may want to do this.
//
// The returned error also implements the Unwrap method from the standard [errors] package, so it
// works with [errors.Is] and [errors.As].
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	wrap.ErrorWithAttrs(err, "error message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	wrap.ErrorWithAttrs(err, "error message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	wrap.ErrorWithAttrs(err, "error message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
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
//
// [hermannm.dev/devlog/log]: https://pkg.go.dev/hermannm.dev/devlog/log
func ErrorWithAttrs(wrapped error, message string, logAttributes ...any) error {
	return wrappedErrorWithAttrs{wrapped, message, internal.ParseAttrs(logAttributes)}
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

// ErrorsWithAttrs wraps the given errors with a message and log attributes, to add structured
// context to the error when it is logged (see below for how to pass attributes).
//
// The returned error implements the following method:
//
//	LogAttrs() []slog.Attr
//
// A logging library can check for the existence of this method when an error is logged, to add
// these attributes to the log output. The [hermannm.dev/devlog/log] library, which wraps
// [log/slog], does this in its error-aware logging functions.
//
// If you're in a function with a [context.Context] parameter, consider using
// [hermannm.dev/wrap/ctxwrap.ErrorsWithAttrs] instead. See the [hermannm.dev/wrap/ctxwrap] package
// docs for why you may want to do this.
//
// The returned error also implements the Unwrap method from the standard [errors] package, so it
// works with [errors.Is] and [errors.As].
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	wrap.ErrorsWithAttrs(errs, "error message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	wrap.ErrorsWithAttrs(errs, "error message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	wrap.ErrorsWithAttrs(errs, "error message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
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
//
// [hermannm.dev/devlog/log]: https://pkg.go.dev/hermannm.dev/devlog/log
func ErrorsWithAttrs(wrapped []error, message string, logAttributes ...any) error {
	return wrappedErrorsWithAttrs{wrapped, message, internal.ParseAttrs(logAttributes)}
}

// NewErrorWithAttrs returns a new error with the given message, and logging attributes to add
// structured context to the error when it is logged (see below for how to pass attributes).
//
// The returned error implements the following method:
//
//	LogAttrs() []slog.Attr
//
// A logging library can check for the existence of this method when an error is logged, to add
// these attributes to the log output. The [hermannm.dev/devlog/log] library, which wraps
// [log/slog], does this in its error-aware logging functions.
//
// If you're in a function with a [context.Context] parameter, consider using
// [hermannm.dev/wrap/ctxwrap.NewErrorWithAttrs] instead. See the [hermannm.dev/wrap/ctxwrap]
// package docs for why you may want to do this.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	wrap.NewErrorWithAttrs("error message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	wrap.NewErrorWithAttrs("error message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	wrap.NewErrorWithAttrs("error message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
//
// [hermannm.dev/devlog/log]: https://pkg.go.dev/hermannm.dev/devlog/log
func NewErrorWithAttrs(message string, logAttributes ...any) error {
	return errorWithAttrs{message, internal.ParseAttrs(logAttributes)}
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

type wrappedErrorWithAttrs struct {
	wrapped error
	message string
	attrs   []slog.Attr
}

func (err wrappedErrorWithAttrs) Error() string {
	return internal.BuildWrappedErrorString(err)
}

// Unwrap matches the signature for wrapped errors expected by the [errors] package.
func (err wrappedErrorWithAttrs) Unwrap() error {
	return err.wrapped
}

// WrappingMessage implements [hermannm.dev/devlog/log.wrappedError] for log message formatting.
//
// [hermannm.dev/devlog/log.wrappedError]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go#L7-L13
func (err wrappedErrorWithAttrs) WrappingMessage() string {
	return err.message
}

func (err wrappedErrorWithAttrs) LogAttrs() []slog.Attr {
	return err.attrs
}

type wrappedErrorsWithAttrs struct {
	wrapped []error
	message string
	attrs   []slog.Attr
}

func (err wrappedErrorsWithAttrs) Error() string {
	return internal.BuildWrappedErrorsString(err)
}

// Unwrap matches the signature for wrapped errors expected by the [errors] package.
func (err wrappedErrorsWithAttrs) Unwrap() []error {
	return err.wrapped
}

// WrappingMessage implements [hermannm.dev/devlog/log.wrappedErrors] for log message formatting.
//
// [hermannm.dev/devlog/log.wrappedErrors]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go#L15-L21
func (err wrappedErrorsWithAttrs) WrappingMessage() string {
	return err.message
}

func (err wrappedErrorsWithAttrs) LogAttrs() []slog.Attr {
	return err.attrs
}

type errorWithAttrs struct {
	message string
	attrs   []slog.Attr
}

func (err errorWithAttrs) Error() string {
	return err.message
}

func (err errorWithAttrs) LogAttrs() []slog.Attr {
	return err.attrs
}
