package ctxwrap_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"reflect"
	"testing"

	"hermannm.dev/wrap/ctxwrap"
)

var ctx = context.WithValue(context.Background(), "testkey", "testvalue")

func TestError(t *testing.T) {
	err := errors.New("error")
	wrapped := ctxwrap.Error(ctx, err, "wrapped error")

	expected := `wrapped error
- error`

	assertErrorString(t, wrapped, expected)
	assertContext(t, wrapped)
}

func TestErrorf(t *testing.T) {
	err := errors.New("username already taken")
	wrapped := ctxwrap.Errorf(ctx, err, "failed to create user with name '%s'", "hermannm")

	expected := `failed to create user with name 'hermannm'
- username already taken`

	assertErrorString(t, wrapped, expected)
	assertContext(t, wrapped)
}

func TestErrorWithAttrs(t *testing.T) {
	err := errors.New("error")
	wrapped := ctxwrap.ErrorWithAttrs(
		ctx,
		err,
		"wrapped error",
		"key1",
		"value1",
		slog.Int("key2", 2),
	)

	expected := `wrapped error
- error`

	assertErrorString(t, wrapped, expected)
	assertLogAttrs(t, wrapped, slog.String("key1", "value1"), slog.Int("key2", 2))
	assertContext(t, wrapped)
}

func TestErrors(t *testing.T) {
	errs := []error{errors.New("username too long"), errors.New("invalid email")}
	wrapped := ctxwrap.Errors(ctx, errs, "user creation failed")

	expected := `user creation failed
- username too long
- invalid email`

	assertErrorString(t, wrapped, expected)
	assertContext(t, wrapped)
}

func TestErrorsWithAttrs(t *testing.T) {
	errs := []error{errors.New("error 1"), errors.New("error 2")}
	wrapped := ctxwrap.ErrorsWithAttrs(
		ctx,
		errs,
		"wrapped errors",
		"key1",
		"value1",
		slog.Int("key2", 2),
	)

	expected := `wrapped errors
- error 1
- error 2`

	assertErrorString(t, wrapped, expected)
	assertLogAttrs(t, wrapped, slog.String("key1", "value1"), slog.Int("key2", 2))
	assertContext(t, wrapped)
}

func TestErrorsf(t *testing.T) {
	errs := []error{errors.New("username already taken"), errors.New("invalid email")}
	wrapped := ctxwrap.Errorsf(ctx, errs, "failed to create user with name '%s'", "hermannm")

	expected := `failed to create user with name 'hermannm'
- username already taken
- invalid email`

	assertErrorString(t, wrapped, expected)
	assertContext(t, wrapped)
}

func TestNewErrorWithAttrs(t *testing.T) {
	err := ctxwrap.NewErrorWithAttrs(ctx, "error message", "key1", "value1", slog.Int("key2", 2))

	assertErrorString(t, err, "error message")
	assertLogAttrs(t, err, slog.String("key1", "value1"), slog.Int("key2", 2))
	assertContext(t, err)
}

func TestNestedError(t *testing.T) {
	err := errors.New("error")
	inner := ctxwrap.Error(ctx, err, "inner wrapped error")
	outer := ctxwrap.Error(ctx, inner, "outer wrapped error")

	expected := `outer wrapped error
- inner wrapped error
- error`

	assertErrorString(t, outer, expected)
}

func TestNestedErrors(t *testing.T) {
	wrappedErrs1 := []error{errors.New("error 1"), errors.New("error 2")}
	inner1 := ctxwrap.Errors(ctx, wrappedErrs1, "inner wrapped errors 1")

	wrappedErrs2 := []error{errors.New("error 3"), errors.New("error 4")}
	inner2 := ctxwrap.Errors(ctx, wrappedErrs2, "inner wrapped errors 2")

	inner3 := ctxwrap.Error(ctx, inner2, "inner wrapped error 3")
	inner4 := ctxwrap.Error(ctx, inner3, "inner wrapped error 4")

	outer := ctxwrap.Errors(ctx, []error{inner1, inner4}, "outer wrapped error")

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
	inner := ctxwrap.Errors(
		ctx,
		[]error{err1, err2},
		`wrapped multiline
errors`,
	)
	outer := ctxwrap.Error(ctx, inner, "outer wrapped error")

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
	wrapped1 := ctxwrap.Errors(ctx, []error{err1}, "wrapped 1")
	wrapped2 := ctxwrap.Error(ctx, wrapped1, "wrapped 2")

	err2 := errors.New("error 2")
	wrapped3 := ctxwrap.Errors(ctx, []error{err2}, "wrapped 3")

	wrapped4 := ctxwrap.Errors(ctx, []error{wrapped2, wrapped3}, "wrapped 4")

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
	wrapped := ctxwrap.Error(ctx, err4, "wrapped error")

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
	wrapped := ctxwrap.Error(ctx, err2, "wrapped error")

	expected := `wrapped error
- context with newline
- error with
  newline`

	assertErrorString(t, wrapped, expected)
}

func TestErrorsIs(t *testing.T) {
	wrapped := ctxwrap.Error(ctx, fs.ErrNotExist, "file not found")
	if !errors.Is(wrapped, fs.ErrNotExist) {
		t.Error("expected errors.Is to return true for wrapped error")
	}

	otherErr := errors.New("some other error")
	wrapped2 := ctxwrap.Errors(
		ctx,
		[]error{otherErr, fs.ErrNotExist},
		"failed to find file, and got other error",
	)
	if !errors.Is(wrapped2, fs.ErrNotExist) {
		t.Error("expected errors.Is to return true for wrapped errors")
	}

	wrapped3 := ctxwrap.Error(ctx, wrapped2, "nested wrapped error")
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
	wrapped := ctxwrap.Error(ctx, originalErr, "failed to read file")

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
----------------------------------------
`,
			expected,
			actual,
		)
	}
}

func assertLogAttrs(t *testing.T, err error, expected ...slog.Attr) {
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

func assertContext(t *testing.T, err error) {
	errWithContext, ok := err.(interface{ Context() context.Context })
	if !ok {
		t.Fatalf("Expected error to implement Context() context.Context")
	}

	actual := errWithContext.Context()
	testValue := actual.Value("testkey")
	if testValue != "testvalue" {
		t.Fatalf("Expected error context to have value testkey=testvalue, but got '%v'", testValue)
	}
}
