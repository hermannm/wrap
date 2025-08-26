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
//	err := errors.New("duplicate primary key")
//	wrapped := wrap.Error(err, "database insert failed")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	database insert failed
//	- duplicate primary key
//
// Wrapped errors can be nested. Wrapping an already wrapped error adds it to the error list, so
// this next example:
//
//	err := errors.New("duplicate primary key")
//	inner := wrap.Error(err, "database insert failed")
//	outer := wrap.Error(inner, "failed to store event")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	failed to store event
//	- database insert failed
//	- duplicate primary key
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
//	err := errors.New("unrecognized event type")
//	wrapped := wrap.Errorf(err, "failed to process event of type '%s'", "ORDER_UPDATED")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	failed to process event of type 'ORDER_UPDATED'
//	- unrecognized event type
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
//	err := errors.New("duplicate primary key")
//	wrapped := wrap.Error(err, "database insert failed")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	database insert failed
//	- duplicate primary key
//
// Wrapped errors can be nested. Wrapping an already wrapped error adds it to the error list, so
// this next example:
//
//	err := errors.New("duplicate primary key")
//	inner := wrap.Error(err, "database insert failed")
//	outer := wrap.Error(inner, "failed to store event")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	failed to store event
//	- database insert failed
//	- duplicate primary key
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
//	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
//	wrapped := wrap.Errors(errs, "failed to parse event")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	failed to parse event
//	- invalid timestamp format
//	- id was not UUID
//
// When combined with [wrap.Error], nested wrapped errors are indented, so this next example:
//
//	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
//	inner := wrap.Errors(errs, "failed to parse event")
//	outer := wrap.Error(inner, "event processing failed")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	event processing failed
//	- failed to parse event
//	  - invalid timestamp format
//	  - id was not UUID
func Errors(wrapped []error, message string) error {
	return wrappedErrors{wrapped, message}
}

// Errorsf wraps the given errors with a formatted message, to add context to the error. It forwards
// the given message format and args to [fmt.Sprintf] to construct the message.
//
// If you're in a function with a [context.Context] parameter, consider using
// [hermannm.dev/wrap/ctxwrap.Errorsf] instead. See the [hermannm.dev/wrap/ctxwrap] package docs for
// why you may want to do this.
//
// The returned error implements the Unwrap method from the standard errors package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
//	wrapped := wrap.Errorsf(errs, "failed to process event of type '%s'", "ORDER_UPDATED")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	failed to process event of type 'ORDER_UPDATED'
//	- invalid timestamp format
//	- id was not UUID
func Errorsf(wrapped []error, messageFormat string, formatArgs ...any) error {
	return wrappedErrors{wrapped, fmt.Sprintf(messageFormat, formatArgs...)}
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
//	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
//	wrapped := wrap.Errors(errs, "failed to parse event")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	failed to parse event
//	- invalid timestamp format
//	- id was not UUID
//
// When combined with [wrap.Error], nested wrapped errors are indented, so this next example:
//
//	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
//	inner := wrap.Errors(errs, "failed to parse event")
//	outer := wrap.Error(inner, "event processing failed")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	event processing failed
//	- failed to parse event
//	  - invalid timestamp format
//	  - id was not UUID
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

// WrappingMessage implements [hermannm.dev/devlog/log.hasWrappingMessage] to support structured
// error logging.
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

// WrappingMessage implements [hermannm.dev/devlog/log.hasWrappingMessage] to support structured
// error logging.
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

// WrappingMessage implements [hermannm.dev/devlog/log.hasWrappingMessage] to support structured
// error logging.
//
// [hermannm.dev/devlog/log.hasWrappingMessage]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedErrorWithAttrs) WrappingMessage() string {
	return err.message
}

// LogAttrs implements [hermannm.dev/devlog/log.hasLogAttributes] to attach structured logging
// context to errors.
//
// [hermannm.dev/devlog/log.hasLogAttributes]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
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

// WrappingMessage implements [hermannm.dev/devlog/log.hasWrappingMessage] to support structured
// error logging.
//
// [hermannm.dev/devlog/log.hasWrappingMessage]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedErrorsWithAttrs) WrappingMessage() string {
	return err.message
}

// LogAttrs implements [hermannm.dev/devlog/log.hasLogAttributes] to attach structured logging
// context to errors.
//
// [hermannm.dev/devlog/log.hasLogAttributes]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
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

// LogAttrs implements [hermannm.dev/devlog/log.hasLogAttributes] to attach structured logging
// context to errors.
//
// [hermannm.dev/devlog/log.hasLogAttributes]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err errorWithAttrs) LogAttrs() []slog.Attr {
	return err.attrs
}
