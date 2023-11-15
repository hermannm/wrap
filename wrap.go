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

// Errorf wraps the given error with a message for context. It forwards the given message format and
// args to [fmt.Sprintf] to construct the message.
//
// Example:
//
//	err := errors.New("username already taken")
//	wrapped := wrap.Errorf(err, "failed to create user with name '%s'", "hermannm")
//	fmt.Println(wrapped)
//	// failed to create user with name 'hermannm'
//	// - username already taken
func Errorf(wrapped error, messageFormat string, formatArgs ...any) error {
	return Error(wrapped, fmt.Sprintf(messageFormat, formatArgs...))
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
	return wrappedErrors{message: message, wrapped: wrapped}
}

type wrappedError struct {
	message string
	wrapped error
}

func (err wrappedError) Error() string {
	var builder errorBuilder
	builder.WriteString(err.message)
	builder.writeErrorListItem(err.wrapped, 1, false)
	return builder.String()
}

// Unwrap matches the signature for wrapped errors expected by the [errors] package.
func (err wrappedError) Unwrap() error {
	return err.wrapped
}

// WrappingMessage implements [hermannm.dev/devlog/log.WrappedError] for log message formatting.
func (err wrappedError) WrappingMessage() string {
	return err.message
}

type wrappedErrors struct {
	message string
	wrapped []error
}

func (err wrappedErrors) Error() string {
	var builder errorBuilder
	builder.WriteString(err.message)
	builder.writeErrorList(err.wrapped, 1)
	return builder.String()
}

// Unwrap matches the signature for wrapped errors expected by the [errors] package.
func (err wrappedErrors) Unwrap() []error {
	return err.wrapped
}

// WrappingMessage implements [hermannm.dev/devlog/log.WrappedError] for log message formatting.
func (err wrappedErrors) WrappingMessage() string {
	return err.message
}

type errorBuilder struct {
	strings.Builder
}

func (builder *errorBuilder) writeErrorListItem(wrappedErr error, indent int, partOfList bool) {
	builder.writeListItemPrefix(indent)

	switch err := wrappedErr.(type) {
	case wrappedError:
		builder.writeErrorMessage([]byte(err.message), indent)

		nextIndent := indent
		if partOfList {
			nextIndent++
		}
		builder.writeErrorListItem(err.wrapped, nextIndent, false)
	case wrappedErrors:
		builder.writeErrorMessage([]byte(err.message), indent)
		builder.writeErrorList(err.wrapped, indent+1)
	default:
		builder.writeExternalErrorMessage([]byte(err.Error()), indent, partOfList)
	}
}

func (builder *errorBuilder) writeErrorList(wrappedErrs []error, indent int) {
	for _, wrappedErr := range wrappedErrs {
		builder.writeErrorListItem(wrappedErr, indent, len(wrappedErrs) > 1)
	}
}

// Splits error messages longer than 64 characters at ": " (typically used for error wrapping), if
// present. Ensures that no splits are shorter than 16 characters (except the last one).
func (builder *errorBuilder) writeExternalErrorMessage(
	message []byte,
	indent int,
	partOfList bool,
) {
	const minSplitLength = 16
	const maxSplitLength = 64

	if len(message) <= maxSplitLength {
		builder.writeErrorMessage(message, indent)
		return
	}

	lastWriteIndex := 0

MessageLoop:
	for i := 0; i < len(message)-1; i++ {
		switch message[i] {
		case ':':
			// Safe to index [i+1], since we loop until the second-to-last index
			switch message[i+1] {
			case ' ', '\n':
				if i-lastWriteIndex < minSplitLength {
					continue MessageLoop
				}

				// First split
				if lastWriteIndex == 0 {
					if partOfList {
						indent++
					}
				} else {
					builder.writeListItemPrefix(indent)
				}

				builder.Write(message[lastWriteIndex:i])

				lastWriteIndex = i + 2 // +2 for ': '
				if len(message)-lastWriteIndex <= maxSplitLength {
					break MessageLoop // Remaining message is short enough, we're done
				}

				i++ // Skips next character, since we already looked at it
			}
		case '\n':
			// Once we hit a newline (not preceded by ':'), we only indent the remainder of the
			// message
			if lastWriteIndex != 0 {
				builder.writeListItemPrefix(indent)
			}
			builder.Write(message[lastWriteIndex : i+1])
			builder.writeIndent(indent + 1)
			builder.writeErrorMessage(message[i+1:], indent)
			return
		}
	}

	if lastWriteIndex == 0 {
		builder.writeErrorMessage(message, indent)
		return
	}

	// Writes remainder after last split
	builder.writeListItemPrefix(indent)
	builder.writeErrorMessage(message[lastWriteIndex:], indent)
}

func (builder *errorBuilder) writeErrorMessage(message []byte, indent int) {
	indent++ // Since indent is made for list bullet points, we want to indent one level deeper

	lastWriteIndex := 0
	for i := 0; i < len(message)-1; i++ {
		if message[i] == '\n' {
			builder.Write(message[lastWriteIndex : i+1])
			builder.writeIndent(indent)
			lastWriteIndex = i + 1
		}
	}

	builder.Write(message[lastWriteIndex:])
}

func (builder *errorBuilder) writeListItemPrefix(indent int) {
	builder.WriteByte('\n')
	builder.writeIndent(indent)
	builder.WriteString("- ")
}

func (builder *errorBuilder) writeIndent(indent int) {
	for i := 1; i < indent; i++ {
		builder.WriteString("  ")
	}
}
