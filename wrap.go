// Package wrap provides utility functions to wrap errors with extra context in an easy-to-read
// format.
package wrap

import (
	"fmt"
	"strings"
)

// Error wraps the given error with a message for context.
//
// The error is displayed on the following format:
//
//	err := errors.New("expired token")
//	wrapped := wrap.Error(err, "user authentication failed")
//	fmt.Println(wrapped)
//	// user authentication failed
//	// - expired token
//
// Wrapped errors can be nested. Wrapping an already wrapped error adds it to the error list, as
// follows:
//
//	err := errors.New("expired token")
//	inner := wrap.Error(err, "user authentication failed")
//	outer := wrap.Error(inner, "failed to update username")
//	fmt.Println(outer)
//	// failed to update username
//	// - user authentication failed
//	// - expired token
//
// The returned error implements the Unwrap method from the standard errors package, so it works
// with [errors.Is] and [errors.As].
func Error(wrapped error, message string) error {
	return wrappedError{wrapped: wrapped, message: message}
}

// Errorf wraps the given error with a message for context. It forwards the given format string and
// args to [fmt.Sprintf] to construct the message.
//
// Example:
//
//	err := errors.New("username already taken")
//	wrapped := wrap.Errorf(err, "failed to create user with name '%s'", "hermannm")
//	fmt.Println(wrapped)
//	// failed to create user with name 'hermannm'
//	// - username already taken
func Errorf(wrapped error, format string, args ...any) error {
	return Error(wrapped, fmt.Sprintf(format, args...))
}

// Errors wraps the given errors with a message for context.
//
// The error is displayed on the following format:
//
//	err1 := errors.New("username too long")
//	err2 := errors.New("invalid email")
//	wrapped := wrap.Errors("user creation failed", err1, err2)
//	fmt.Println(wrapped)
//	// user creation failed
//	// - username too long
//	// - invalid email
//
// When combined with [Error], nested wrapped errors are indented as follows:
//
//	err1 := errors.New("username too long")
//	err2 := errors.New("invalid email")
//	inner := wrap.Errors("user creation failed", err1, err2)
//	outer := wrap.Error(inner, "failed to register new user")
//	fmt.Println(outer)
//	// failed to register new user
//	// - user creation failed
//	//   - username too long
//	//   - invalid email
//
// The returned error implements the Unwrap method from the standard errors package, so it works
// with [errors.Is] and [errors.As].
func Errors(message string, wrapped ...error) error {
	return wrappedErrors{wrapped: wrapped, message: message}
}

type wrappedError struct {
	wrapped error
	message string
}

func (err wrappedError) Error() string {
	return buildErrorString(err.message, []error{err.wrapped})
}

func (err wrappedError) Unwrap() error {
	return err.wrapped
}

type wrappedErrors struct {
	wrapped []error
	message string
}

func (err wrappedErrors) Error() string {
	return buildErrorString(err.message, err.wrapped)
}

func (err wrappedErrors) Unwrap() []error {
	return err.wrapped
}

func buildErrorString(message string, wrappedErrs []error) string {
	var errString strings.Builder
	errString.WriteString(message)
	buildErrorList(&errString, wrappedErrs, 1)
	return errString.String()
}

func buildErrorList(errString *strings.Builder, wrappedErrs []error, indent int) {
	for _, wrappedErr := range wrappedErrs {
		errString.WriteRune('\n')
		for i := 1; i < indent; i++ {
			errString.WriteString("  ")
		}
		errString.WriteString("- ")

		switch err := wrappedErr.(type) {
		case wrappedError:
			writeErrorMessage(errString, err.message, indent)

			nextIndent := indent
			if len(wrappedErrs) > 1 {
				nextIndent++
			}
			buildErrorList(errString, []error{err.wrapped}, nextIndent)
		case wrappedErrors:
			writeErrorMessage(errString, err.message, indent)
			buildErrorList(errString, err.wrapped, indent+1)
		default:
			writeErrorMessage(errString, err.Error(), indent)
		}
	}
}

func writeErrorMessage(errString *strings.Builder, message string, indent int) {
	lines := strings.SplitAfter(message, "\n")
	for i, line := range lines {
		if i > 0 {
			for i := 1; i <= indent; i++ {
				errString.WriteString("  ")
			}
		}
		errString.WriteString(line)
	}
}
