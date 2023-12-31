package wrap_test

import (
	"errors"
	"io/fs"
	"testing"

	"hermannm.dev/wrap"
)

func TestError(t *testing.T) {
	err := errors.New("error")
	wrapped := wrap.Error(err, "wrapped error")

	expected := `wrapped error
- error`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestErrorf(t *testing.T) {
	err := errors.New("username already taken")
	wrapped := wrap.Errorf(err, "failed to create user with name '%s'", "hermannm")

	expected := `failed to create user with name 'hermannm'
- username already taken`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	wrapped := wrap.Errors("wrapped errors", err1, err2)

	expected := `wrapped errors
- error 1
- error 2`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestNestedError(t *testing.T) {
	err := errors.New("error")
	inner := wrap.Error(err, "inner wrapped error")
	outer := wrap.Error(inner, "outer wrapped error")

	expected := `outer wrapped error
- inner wrapped error
- error`

	assertEqualErrorStrings(t, outer, expected)
}

func TestNestedErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	err4 := errors.New("error 4")

	inner1 := wrap.Errors("inner wrapped errors 1", err1, err2)
	inner2 := wrap.Errors("inner wrapped errors 2", err3, err4)
	inner3 := wrap.Error(inner2, "inner wrapped error 3")
	inner4 := wrap.Error(inner3, "inner wrapped error 4")

	outer := wrap.Errors("outer wrapped error", inner1, inner4)

	expected := `outer wrapped error
- inner wrapped errors 1
  - error 1
  - error 2
- inner wrapped error 4
  - inner wrapped error 3
  - inner wrapped errors 2
    - error 3
    - error 4`

	assertEqualErrorStrings(t, outer, expected)
}

func TestMultilineError(t *testing.T) {
	err1 := errors.New(`multiline
error 1`)
	err2 := errors.New(`multiline
error 2`)
	inner := wrap.Errors(`wrapped multiline
errors`, err1, err2)
	outer := wrap.Error(inner, "outer wrapped error")

	expected := `outer wrapped error
- wrapped multiline
  errors
  - multiline
    error 1
  - multiline
    error 2`

	assertEqualErrorStrings(t, outer, expected)
}

func TestSingleWrappedErrors(t *testing.T) {
	err1 := errors.New("error 1")
	wrapped1 := wrap.Errors("wrapped 1", err1)
	wrapped2 := wrap.Error(wrapped1, "wrapped 2")

	err2 := errors.New("error 2")
	wrapped3 := wrap.Errors("wrapped 3", err2)

	wrapped4 := wrap.Errors("wrapped 4", wrapped2, wrapped3)

	expected := `wrapped 4
- wrapped 2
  - wrapped 1
  - error 1
- wrapped 3
  - error 2`

	assertEqualErrorStrings(t, wrapped4, expected)
}

func TestLongErrorMessage(t *testing.T) {
	err := errors.New(
		"this error message is more than 16 characters: " +
			"less than 16: " +
			"now again longer than 16 characters: " +
			"this is a long error message, of barely less than 64 characters: " +
			"short message",
	)
	wrapped := wrap.Error(err, "wrapped error")

	expected := `wrapped error
- this error message is more than 16 characters
- less than 16: now again longer than 16 characters
- this is a long error message, of barely less than 64 characters
- short message`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestLongErrorMessageInList(t *testing.T) {
	err1 := errors.New(
		"this is a long error message, of barely less than 64 characters: " +
			"this error message is more than 16 characters: " +
			"and another one longer than 16 characters",
	)
	err2 := errors.New("other error")
	wrapped := wrap.Errors("wrapped error", err1, err2)

	expected := `wrapped error
- this is a long error message, of barely less than 64 characters
  - this error message is more than 16 characters
  - and another one longer than 16 characters
- other error`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestLongMultilineErrorMessage(t *testing.T) {
	err := errors.New(`this error message ends in a newline and colon:
more than 16 characters: this message ends in a newline
another message ending in a newline and colon:
another newline message`)
	wrapped := wrap.Error(err, "wrapped error")

	expected := `wrapped error
- this error message ends in a newline and colon
- more than 16 characters
- this message ends in a newline
  another message ending in a newline and colon:
  another newline message`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestUnsplittableErrorMessage(t *testing.T) {
	err := errors.New("this is a super long error message of more than 64 characters in total")
	wrapped := wrap.Error(err, "wrapped error")

	expected := `wrapped error
- this is a super long error message of more than 64 characters in total`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestErrorsIs(t *testing.T) {
	wrapped := wrap.Error(fs.ErrNotExist, "file not found")
	if !errors.Is(wrapped, fs.ErrNotExist) {
		t.Error("expected errors.Is to return true for wrapped error")
	}

	otherErr := errors.New("some other error")
	wrapped2 := wrap.Errors("failed to find file, and got other error", otherErr, fs.ErrNotExist)
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

	if pathErr != originalErr {
		t.Errorf(
			"expected error gotten from errors.As to equal original wrapped error; got %+v, want %+v",
			*pathErr,
			*originalErr,
		)
	}
}

func assertEqualErrorStrings(t *testing.T, errToTest error, expected string) {
	if actual := errToTest.Error(); actual != expected {
		t.Errorf(`unexpected error string
got:
----------------------------------------
%s
----------------------------------------

want:
----------------------------------------
%s
----------------------------------------
`, actual, expected)
	}
}
