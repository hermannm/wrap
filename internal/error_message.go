package internal

import "strings"

// Same interface that the standard [errors] package uses to support error wrapping.
type wrappingError interface {
	error
	Unwrap() error
}

// Same interface that the standard [errors] package uses to support wrapping of multiple errors.
type wrappingErrors interface {
	error
	Unwrap() []error
}

// hasWrappingMessage is an interface for errors that wrap an inner error with a wrapping message.
// The errors returned by this library all implement this interface.
//
// We don't export this interface, as we don't want library consumers to depend on it directly. The
// interface type itself is an implementation detail - we only use it to check if errors logged by
// this library implicitly implement this method. This is the same approach that the standard
// [errors] package uses to support Unwrap().
type hasWrappingMessage interface {
	WrappingMessage() string
}

func BuildWrappedErrorString(
	err interface {
		wrappingError
		hasWrappingMessage
	},
) string {
	var builder errorBuilder
	_, _ = builder.WriteString(err.WrappingMessage())
	builder.writeErrorListItem(err.Unwrap(), 1, false)
	return builder.String()
}

func BuildWrappedErrorsString(
	err interface {
		wrappingErrors
		hasWrappingMessage
	},
) string {
	var builder errorBuilder
	_, _ = builder.WriteString(err.WrappingMessage())
	builder.writeErrorList(err.Unwrap(), 1)
	return builder.String()
}

// strings.Builder always returns a nil error, so we don't have to check write errors with this.
type errorBuilder struct {
	strings.Builder
}

func (builder *errorBuilder) writeErrorListItem(wrappedErr error, indent int, partOfList bool) {
	builder.writeListItemPrefix(indent)

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := wrappedErr.(type) {
	case wrappingError:
		wrapped, errMessage, errMessageIsWrappingMessage := unwrapError(err)

		builder.writeErrorMessage([]byte(errMessage), indent)
		if errMessageIsWrappingMessage {
			if partOfList {
				indent++
			}
			builder.writeErrorListItem(wrapped, indent, false)
		}
	case wrappingErrors:
		wrapped, errMessage, errMessageIsWrappingMessage := unwrapErrors(err)

		builder.writeErrorMessage([]byte(errMessage), indent)
		if errMessageIsWrappingMessage {
			if partOfList || len(wrapped) > 1 {
				indent++
			}
			builder.writeErrorList(wrapped, indent)
		}
	default:
		builder.writeErrorMessage([]byte(err.Error()), indent)
	}
}

func (builder *errorBuilder) writeErrorList(wrappedErrs []error, indent int) {
	for _, wrappedErr := range wrappedErrs {
		builder.writeErrorListItem(wrappedErr, indent, len(wrappedErrs) > 1)
	}
}

func (builder *errorBuilder) writeErrorMessage(message []byte, indent int) {
	indent++ // Since indent is made for list bullet points, we want to indent one level deeper

	lastWriteIndex := 0
	for i, char := range message {
		if char == '\n' {
			_, _ = builder.Write(message[lastWriteIndex : i+1])
			builder.writeIndent(indent)
			lastWriteIndex = i + 1
		}
	}

	_, _ = builder.Write(message[lastWriteIndex:])
}

func (builder *errorBuilder) writeListItemPrefix(indent int) {
	_ = builder.WriteByte('\n')
	builder.writeIndent(indent)
	_, _ = builder.WriteString("- ")
}

func (builder *errorBuilder) writeIndent(indent int) {
	for i := 1; i < indent; i++ {
		_, _ = builder.WriteString("  ")
	}
}

// If errMessageIsWrappingMessage is true, then the returned errMessage is the wrapping message
// around the wrapped error. Otherwise, the returned errMessage is the full error message of the
// given err.
//
// Same implementation that the [hermannm.dev/devlog/log] library uses for structuring error logs.
//
// [hermannm.dev/devlog/log]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func unwrapError(err wrappingError) (
	unwrapped error,
	errMessage string,
	errMessageIsWrappingMessage bool,
) {
	unwrapped = err.Unwrap()

	// If err has a WrappingMessage() string method, we use that as the wrapping message
	if wrapper, ok := err.(hasWrappingMessage); ok {
		return unwrapped, wrapper.WrappingMessage(), true
	}

	errMessage = err.Error()
	if unwrapped == nil {
		return nil, errMessage, false
	}

	// If err did not implement WrappingMessage(), we look for a common pattern for wrapping errors:
	//	fmt.Errorf("wrapping message: %w", unwrapped)
	// If the full error message is suffixed by the unwrapped error message, with a ": " separator,
	// we can get the wrapping message before the separator.
	unwrappedMessage := unwrapped.Error()

	// -2 for ": " separator between wrapping message and unwrapped error
	wrappingMessageEndIndex := len(errMessage) - len(unwrappedMessage) - 2

	if wrappingMessageEndIndex > 0 &&
		strings.HasSuffix(errMessage, unwrappedMessage) &&
		errMessage[wrappingMessageEndIndex] == ':' {
		// Check for either space or newline in character after colon
		charAfterColon := errMessage[wrappingMessageEndIndex+1]

		if charAfterColon == ' ' || charAfterColon == '\n' {
			wrappingMessage := errMessage[0:wrappingMessageEndIndex]
			return unwrapped, wrappingMessage, true
		}
	}

	return unwrapped, errMessage, false
}

// If errMessageIsWrappingMessage is true, then the returned errMessage is the wrapping message
// around the wrapped errors. Otherwise, the returned errMessage is the full error message of the
// given err.
//
// Same implementation that the [hermannm.dev/devlog/log] library uses for structuring error logs.
//
// [hermannm.dev/devlog/log]: https://github.com/hermannm/devlog/blob/v0.6.0/log/errors.go
func unwrapErrors(err wrappingErrors) (
	unwrapped []error,
	errMessage string,
	errMessageIsWrappingMessage bool,
) {
	unwrapped = err.Unwrap()

	if wrapper, ok := err.(hasWrappingMessage); ok {
		return unwrapped, wrapper.WrappingMessage(), true
	} else {
		return unwrapped, err.Error(), false
	}
}
