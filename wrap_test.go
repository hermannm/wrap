package wrap_test

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"reflect"
	"testing"

	"hermannm.dev/wrap"
)

func TestError(t *testing.T) {
	err := errors.New("error")
	wrapped := wrap.Error(err, "wrapped error")

	expected := `wrapped error
- error`

	assertErrorString(t, wrapped, expected)
}

func TestErrorf(t *testing.T) {
	err := errors.New("unrecognized event type")
	wrapped := wrap.Errorf(err, "failed to process event of type '%s'", "ORDER_UPDATED")

	expected := `failed to process event of type 'ORDER_UPDATED'
- unrecognized event type`

	assertErrorString(t, wrapped, expected)
}

func TestErrorWithAttrs(t *testing.T) {
	err := errors.New("error")
	wrapped := wrap.ErrorWithAttrs(err, "wrapped error", "key1", "value1", slog.Int("key2", 2))

	expected := `wrapped error
- error`

	assertErrorString(t, wrapped, expected)
	assertLogAttrs(t, wrapped, slog.String("key1", "value1"), slog.Int("key2", 2))
}

func TestErrors(t *testing.T) {
	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
	wrapped := wrap.Errors(errs, "failed to parse event")

	expected := `failed to parse event
- invalid timestamp format
- id was not UUID`

	assertErrorString(t, wrapped, expected)
}

func TestErrorsWithAttrs(t *testing.T) {
	errs := []error{errors.New("error 1"), errors.New("error 2")}
	wrapped := wrap.ErrorsWithAttrs(errs, "wrapped errors", "key1", "value1", slog.Int("key2", 2))

	expected := `wrapped errors
- error 1
- error 2`

	assertErrorString(t, wrapped, expected)
	assertLogAttrs(t, wrapped, slog.String("key1", "value1"), slog.Int("key2", 2))
}

func TestErrorsf(t *testing.T) {
	errs := []error{errors.New("invalid timestamp format"), errors.New("id was not UUID")}
	wrapped := wrap.Errorsf(errs, "failed to parse event of type '%s'", "ORDER_UPDATED")

	expected := `failed to parse event of type 'ORDER_UPDATED'
- invalid timestamp format
- id was not UUID`

	assertErrorString(t, wrapped, expected)
}

func TestNewErrorWithAttrs(t *testing.T) {
	err := wrap.NewErrorWithAttrs("error message", "key1", "value1", slog.Int("key2", 2))

	assertErrorString(t, err, "error message")
	assertLogAttrs(t, err, slog.String("key1", "value1"), slog.Int("key2", 2))
}

func TestNestedError(t *testing.T) {
	err := errors.New("error")
	inner := wrap.Error(err, "inner wrapped error")
	outer := wrap.Error(inner, "outer wrapped error")

	expected := `outer wrapped error
- inner wrapped error
- error`

	assertErrorString(t, outer, expected)
}

func TestNestedErrors(t *testing.T) {
	wrappedErrs1 := []error{errors.New("error 1"), errors.New("error 2")}
	inner1 := wrap.Errors(wrappedErrs1, "inner wrapped errors 1")

	wrappedErrs2 := []error{errors.New("error 3"), errors.New("error 4")}
	inner2 := wrap.Errors(wrappedErrs2, "inner wrapped errors 2")

	inner3 := wrap.Error(inner2, "inner wrapped error 3")
	inner4 := wrap.Error(inner3, "inner wrapped error 4")

	outer := wrap.Errors([]error{inner1, inner4}, "outer wrapped error")

	expected := `outer wrapped error
- inner wrapped errors 1
  - error 1
  - error 2
- inner wrapped error 4
  - inner wrapped error 3
  - inner wrapped errors 2
    - error 3
    - error 4`

	assertErrorString(t, outer, expected)
}

func TestMultilineError(t *testing.T) {
	err1 := errors.New(
		`multiline
error 1`,
	)
	err2 := errors.New(
		`multiline
error 2`,
	)
	inner := wrap.Errors(
		[]error{err1, err2},
		`wrapped multiline
errors`,
	)
	outer := wrap.Error(inner, "outer wrapped error")

	expected := `outer wrapped error
- wrapped multiline
  errors
  - multiline
    error 1
  - multiline
    error 2`

	assertErrorString(t, outer, expected)
}

func TestSingleWrappedErrors(t *testing.T) {
	err1 := errors.New("error 1")
	wrapped1 := wrap.Errors([]error{err1}, "wrapped 1")
	wrapped2 := wrap.Error(wrapped1, "wrapped 2")

	err2 := errors.New("error 2")
	wrapped3 := wrap.Errors([]error{err2}, "wrapped 3")

	wrapped4 := wrap.Errors([]error{wrapped2, wrapped3}, "wrapped 4")

	expected := `wrapped 4
- wrapped 2
  - wrapped 1
  - error 1
- wrapped 3
  - error 2`

	assertErrorString(t, wrapped4, expected)
}

func TestErrorWrappedWithFmt(t *testing.T) {
	err1 := errors.New("the underlying error")
	err2 := fmt.Errorf("something went wrong: %w", err1)
	err3 := fmt.Errorf("error string with : in the middle: %w", err2)
	err4 := fmt.Errorf("an error occurred: %w", err3)
	wrapped := wrap.Error(err4, "wrapped error")

	expected := `wrapped error
- an error occurred
- error string with : in the middle
- something went wrong
- the underlying error`

	assertErrorString(t, wrapped, expected)
}

func TestMultilineErrorWrappedWithFmt(t *testing.T) {
	err1 := errors.New(
		`error with
newline`,
	)
	err2 := fmt.Errorf("context with newline:\n%w", err1)
	wrapped := wrap.Error(err2, "wrapped error")

	expected := `wrapped error
- context with newline
- error with
  newline`

	assertErrorString(t, wrapped, expected)
}

func TestErrorsIs(t *testing.T) {
	wrapped := wrap.Error(fs.ErrNotExist, "file not found")
	if !errors.Is(wrapped, fs.ErrNotExist) {
		t.Error("expected errors.Is to return true for wrapped error")
	}

	otherErr := errors.New("some other error")
	wrapped2 := wrap.Errors(
		[]error{otherErr, fs.ErrNotExist},
		"failed to find file, and got other error",
	)
	if !errors.Is(wrapped2, fs.ErrNotExist) {
		t.Error("expected errors.Is to return true for wrapped errors")
	}

	wrapped3 := wrap.Error(wrapped2, "nested wrapped error")
	if !errors.Is(wrapped3, fs.ErrNotExist) {
		t.Error("expected errors.Is to return true for nested wrapped error")
	}
}

func TestErrorsAs(t *testing.T) {
	originalErr := &fs.PathError{
		Op:   "open",
		Path: "wrap/wra.go",
		Err:  errors.New("no such file or directory"),
	}
	wrapped := wrap.Error(originalErr, "failed to read file")

	var pathErr *fs.PathError
	if !errors.As(wrapped, &pathErr) {
		t.Error("expected errors.As to return true for wrapped error")
	}

	//goland:noinspection GoDirectComparisonOfErrors - We want to test errors.As here, not errors.Is
	if pathErr != originalErr {
		t.Errorf(
			"expected error gotten from errors.As to equal original wrapped error; got %+v, want %+v",
			*pathErr,
			*originalErr,
		)
	}
}

func assertErrorString(t *testing.T, errToTest error, expected string) {
	t.Helper()

	if actual := errToTest.Error(); actual != expected {
		t.Errorf(
			`Unexpected error string
Want:
----------------------------------------
%s
----------------------------------------
Got:
----------------------------------------
%s
----------------------------------------`,
			expected,
			actual,
		)
	}
}

func assertLogAttrs(t *testing.T, err error, expected ...slog.Attr) {
	t.Helper()

	errWithAttrs, ok := err.(interface{ LogAttrs() []slog.Attr })
	if !ok {
		t.Fatalf("Expected error to implement LogAttrs() []slog.Attr")
	}

	actual := errWithAttrs.LogAttrs()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(
			`Unexpected log attributes
Want: %v
Got: %v`,
			expected,
			actual,
		)
	}
}
