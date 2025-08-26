// Package ctxwrap tries to solve the problem of errors escaping their context. It mirrors the API
// of [hermannm.dev/wrap], but adds a [context.Context] parameter to every error wrapping function,
// so that the error can carry its original context as it's returned up the stack.
//
// Every error returned by this package implements the following method:
//
//	Context() context.Context
//
// Other libraries (e.g. a logging library) can check for this method, to use the error's original
// context.
//
// # Motivation
//
// To see why you may want this, let's look at an example using the [hermannm.dev/devlog/log]
// logging library:
//
//	import (
//		"context"
//
//		"hermannm.dev/devlog/log"
//		"hermannm.dev/wrap"
//	)
//
//	func parentFunction(ctx context.Context) {
//		if err := childFunction(ctx); err != nil {
//			log.Error(ctx, err, "Child function failed")
//		}
//	}
//
//	func childFunction(ctx context.Context) error {
//		// log.AddContextAttrs adds log attributes to the context, to be attached when ctx is logged
//		ctx = log.AddContextAttrs(ctx, "key", "value")
//
//		if err := someFallibleOperation(ctx); err != nil {
//			return wrap.Error(err, "Operation failed")
//		}
//
//		return nil
//	}
//
// In the above example, childFunction returns an error, and parentFunction logs it. This is a
// typical pattern, as errors are often returned up the stack before being logged.
//
// We see that we attach log attributes to the context in childFunction, using log.AddContextAttrs.
// But when we return the error using wrap.Error, we lose those context attributes! This is not
// ideal, as we want as much context as possible when an error is logged.
//
// ctxwrap solves this by letting us attach the context to the error, so that the logging library
// can get the context attributes from the error when it is logged. This revised example uses
// ctxwrap instead of wrap, to propagate context attributes:
//
//	import (
//		"context"
//
//		"hermannm.dev/devlog/log"
//		"hermannm.dev/wrap/ctxwrap"
//	)
//
//	func parentFunction(ctx context.Context) {
//		if err := childFunction(ctx); err != nil {
//			log.Error(ctx, err, "Child function failed")
//		}
//	}
//
//	func childFunction(ctx context.Context) error {
//		ctx = log.AddContextAttrs(ctx, "key", "value")
//
//		if err := someFallibleOperation(ctx); err != nil {
//			// Uses ctxwrap to attach the context to the error
//			return ctxwrap.Error(ctx, err, "Operation failed")
//		}
//
//		return nil
//	}
//
// Now, when parentFunction logs the error from childFunction, the context attributes carried by
// the error will be logged, so we get more context in our error log!
//
// [hermannm.dev/devlog/log]: https://pkg.go.dev/hermannm.dev/devlog/log
package ctxwrap

import (
	"context"
	"fmt"
	"log/slog"

	"hermannm.dev/wrap/internal"
)

// Error wraps the given error with a message, to add context to the error.
//
// It takes a [context.Context] parameter, to preserve the error's context when it's returned up
// the stack (see the [ctxwrap] package docs for more on this). If you're in a function without
// a context parameter, you can use [hermannm.dev/wrap.Error] instead.
//
// The returned error implements the Unwrap method from the standard [errors] package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	err := errors.New("duplicate primary key")
//	wrapped := ctxwrap.Error(ctx, err, "database insert failed")
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
//	inner := ctxwrap.Error(ctx, err, "database insert failed")
//	outer := ctxwrap.Error(ctx, inner, "failed to store event")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	failed to store event
//	- database insert failed
//	- duplicate primary key
func Error(ctx context.Context, wrapped error, message string) error {
	return wrappedError{ctx, wrapped, message}
}

// Errorf wraps the given error with a formatted message, to add context to the error. It forwards
// the given message format and args to [fmt.Sprintf] to construct the message.
//
// It takes a [context.Context] parameter, to preserve the error's context when it's returned up
// the stack (see the [ctxwrap] package docs for more on this). If you're in a function without
// a context parameter, you can use [hermannm.dev/wrap.Errorf] instead.
//
// The returned error implements the Unwrap method from the standard [errors] package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	err := errors.New("unrecognized event type")
//	wrapped := ctxwrap.Errorf(ctx, err, "failed to process event of type '%s'", "ORDER_UPDATED")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	failed to process event of type 'ORDER_UPDATED'
//	- unrecognized event type
func Errorf(
	ctx context.Context,
	wrapped error,
	messageFormat string,
	formatArgs ...any,
) error {
	return wrappedError{ctx, wrapped, fmt.Sprintf(messageFormat, formatArgs...)}
}

// ErrorWithAttrs wraps the given error with a message and log attributes, to add structured context
// to the error when it is logged (see below for how to pass attributes).
//
// It takes a [context.Context] parameter, to preserve the error's context when it's returned up
// the stack (see the [ctxwrap] package docs for more on this). If you're in a function without
// a context parameter, you can use [hermannm.dev/wrap.ErrorWithAttrs] instead.
//
// The returned error implements the following method:
//
//	LogAttrs() []slog.Attr
//
// A logging library can check for the existence of this method when an error is logged, to add
// these attributes to the log output. The [hermannm.dev/devlog/log] library, which wraps
// [log/slog], does this in its error-aware logging functions.
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
//	ctxwrap.ErrorWithAttrs(ctx, err, "error message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	ctxwrap.ErrorWithAttrs(ctx, err, "error message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	ctxwrap.ErrorWithAttrs(ctx, err, "error message", "key1", "value1", slog.Int("key2", 2))
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
//	wrapped := ctxwrap.Error(ctx, err, "database insert failed")
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
//	inner := ctxwrap.Error(ctx, err, "database insert failed")
//	outer := ctxwrap.Error(ctx, inner, "failed to store event")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	failed to store event
//	- database insert failed
//	- duplicate primary key
//
// [hermannm.dev/devlog/log]: https://pkg.go.dev/hermannm.dev/devlog/log
func ErrorWithAttrs(
	ctx context.Context,
	wrapped error,
	message string,
	logAttributes ...any,
) error {
	return wrappedErrorWithAttrs{ctx, wrapped, message, internal.ParseAttrs(logAttributes)}
}

// Errors wraps the given errors with a message, to add context to the errors.
//
// It takes a [context.Context] parameter, to preserve the error's context when it's returned up
// the stack (see the [ctxwrap] package docs for more on this). If you're in a function without
// a context parameter, you can use [hermannm.dev/wrap.Errors] instead.
//
// The returned error implements the Unwrap method from the standard errors package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
//	wrapped := ctxwrap.Errors(ctx, errs, "failed to parse event")
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
//	inner := ctxwrap.Errors(ctx, errs, "failed to parse event")
//	outer := ctxwrap.Error(ctx, inner, "event processing failed")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	event processing failed
//	- failed to parse event
//	  - invalid timestamp format
//	  - id was not UUID
func Errors(ctx context.Context, wrapped []error, message string) error {
	return wrappedErrors{ctx, wrapped, message}
}

// Errorsf wraps the given errors with a formatted message, to add context to the error. It forwards
// the given message format and args to [fmt.Sprintf] to construct the message.
//
// It takes a [context.Context] parameter, to preserve the error's context when it's returned up
// the stack (see the [ctxwrap] package docs for more on this). If you're in a function without
// a context parameter, you can use [hermannm.dev/wrap.Errorsf] instead.
//
// The returned error implements the Unwrap method from the standard errors package, so it works
// with [errors.Is] and [errors.As].
//
// # Error string format
//
// The following example:
//
//	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
//	wrapped := ctxwrap.Errorsf(ctx, errs, "failed to process event of type '%s'", "ORDER_UPDATED")
//	fmt.Println(wrapped)
//
// ...produces this error string:
//
//	failed to process event of type 'ORDER_UPDATED'
//	- invalid timestamp format
//	- id was not UUID
func Errorsf(ctx context.Context, wrapped []error, messageFormat string, formatArgs ...any) error {
	return wrappedErrors{ctx, wrapped, fmt.Sprintf(messageFormat, formatArgs...)}
}

// ErrorsWithAttrs wraps the given errors with a message and log attributes, to add structured
// context to the error when it is logged (see below for how to pass attributes).
//
// It takes a [context.Context] parameter, to preserve the error's context when it's returned up
// the stack (see the [ctxwrap] package docs for more on this). If you're in a function without
// a context parameter, you can use [hermannm.dev/wrap.ErrorsWithAttrs] instead.
//
// The returned error implements the following method:
//
//	LogAttrs() []slog.Attr
//
// A logging library can check for the existence of this method when an error is logged, to add
// these attributes to the log output. The [hermannm.dev/devlog/log] library, which wraps
// [log/slog], does this in its error-aware logging functions.
//
// The returned error implements the Unwrap method from the standard [errors] package, so it works
// with [errors.Is] and [errors.As].
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	ctxwrap.ErrorsWithAttrs(ctx, errs, "error message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	ctxwrap.ErrorsWithAttrs(ctx, errs, "error message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	ctxwrap.ErrorsWithAttrs(ctx, errs, "error message", "key1", "value1", slog.Int("key2", 2))
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
//	wrapped := ctxwrap.Errors(ctx, errs, "failed to parse event")
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
//	inner := ctxwrap.Errors(ctx, errs, "failed to parse event")
//	outer := ctxwrap.Error(ctx, inner, "event processing failed")
//	fmt.Println(outer)
//
// ...produces this error string:
//
//	event processing failed
//	- failed to parse event
//	  - invalid timestamp format
//	  - id was not UUID
func ErrorsWithAttrs(
	ctx context.Context,
	wrapped []error,
	message string,
	logAttributes ...any,
) error {
	return wrappedErrorsWithAttrs{ctx, wrapped, message, internal.ParseAttrs(logAttributes)}
}

// NewError returns a new error with the given message. It takes a [context.Context] parameter, to
// preserve the error's context when it's returned up the stack (see the [ctxwrap] package docs for
// more on this). If you're in a function without a context parameter, you can use [errors.New]
// instead.
func NewError(ctx context.Context, message string) error {
	return errorWithContext{ctx, message}
}

// NewErrorf returns a new error with the given message. It takes a [context.Context] parameter, to
// preserve the error's context when it's returned up the stack (see the [ctxwrap] package docs for
// more on this). If you're in a function without a context parameter, you can use [fmt.Errorf]
// instead.
func NewErrorf(ctx context.Context, messageFormat string, formatArgs ...any) error {
	return errorWithContext{ctx, fmt.Sprintf(messageFormat, formatArgs...)}
}

// NewErrorWithAttrs returns a new error with the given message, and logging attributes to add
// structured context to the error when it is logged (see below for how to pass attributes).
//
// It takes a [context.Context] parameter, to preserve the error's context when it's returned up
// the stack (see the [ctxwrap] package docs for more on this). If you're in a function without
// a context parameter, you can use [hermannm.dev/wrap.NewErrorWithAttrs] instead.
//
// The returned error implements the following method:
//
//	LogAttrs() []slog.Attr
//
// A logging library can check for the existence of this method when an error is logged, to add
// these attributes to the log output. The [hermannm.dev/devlog/log] library, which wraps
// [log/slog], does this in its error-aware logging functions.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	ctxwrap.NewErrorWithAttrs(ctx, "error message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	ctxwrap.NewErrorWithAttrs(ctx, "error message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	ctxwrap.NewErrorWithAttrs(ctx, "error message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
//
// [hermannm.dev/devlog/log]: https://pkg.go.dev/hermannm.dev/devlog/log
func NewErrorWithAttrs(ctx context.Context, message string, logAttributes ...any) error {
	return errorWithAttrs{ctx, message, internal.ParseAttrs(logAttributes)}
}

type wrappedError struct {
	ctx     context.Context
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

// Context returns the original [context.Context] in which the error was created. See the [ctxwrap]
// package docs for the motivation behind this.
//
// This implements the [hermannm.dev/devlog/log.hasContext] interface, which is used in that library
// to attach error context attributes to the log.
//
// [hermannm.dev/devlog/log.hasContext]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedError) Context() context.Context {
	return err.ctx
}

type wrappedErrors struct {
	ctx     context.Context
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

// Context returns the original [context.Context] in which the error was created. See the [ctxwrap]
// package docs for the motivation behind this.
//
// This implements the [hermannm.dev/devlog/log.hasContext] interface, which is used in that library
// to attach error context attributes to the log.
//
// [hermannm.dev/devlog/log.hasContext]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedErrors) Context() context.Context {
	return err.ctx
}

type wrappedErrorWithAttrs struct {
	ctx     context.Context
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

// Context returns the original [context.Context] in which the error was created. See the [ctxwrap]
// package docs for the motivation behind this.
//
// This implements the [hermannm.dev/devlog/log.hasContext] interface, which is used in that library
// to attach error context attributes to the log.
//
// [hermannm.dev/devlog/log.hasContext]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedErrorWithAttrs) Context() context.Context {
	return err.ctx
}

type wrappedErrorsWithAttrs struct {
	ctx     context.Context
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

// Context returns the original [context.Context] in which the error was created. See the [ctxwrap]
// package docs for the motivation behind this.
//
// This implements the [hermannm.dev/devlog/log.hasContext] interface, which is used in that library
// to attach error context attributes to the log.
//
// [hermannm.dev/devlog/log.hasContext]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err wrappedErrorsWithAttrs) Context() context.Context {
	return err.ctx
}

type errorWithContext struct {
	ctx     context.Context
	message string
}

func (err errorWithContext) Error() string {
	return err.message
}

// Context returns the original [context.Context] in which the error was created. See the [ctxwrap]
// package docs for the motivation behind this.
//
// This implements the [hermannm.dev/devlog/log.hasContext] interface, which is used in that library
// to attach error context attributes to the log.
//
// [hermannm.dev/devlog/log.hasContext]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err errorWithContext) Context() context.Context {
	return err.ctx
}

type errorWithAttrs struct {
	ctx     context.Context
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

// Context returns the original [context.Context] in which the error was created. See the [ctxwrap]
// package docs for the motivation behind this.
//
// This implements the [hermannm.dev/devlog/log.hasContext] interface, which is used in that library
// to attach error context attributes to the log.
//
// [hermannm.dev/devlog/log.hasContext]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func (err errorWithAttrs) Context() context.Context {
	return err.ctx
}
