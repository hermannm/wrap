package wrap_test

import (
	"errors"
	"testing"

	"hermannm.dev/wrap"
)

func TestError(t *testing.T) {
	err := errors.New("error")
	wrapped := wrap.Error(err, "wrapped error")

	expected := `wrapped error

Caused by:
- error`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestErrorf(t *testing.T) {
	err := errors.New("username already taken")
	wrapped := wrap.Errorf(err, "failed to create user with name '%s'", "hermannm")

	expected := `failed to create user with name 'hermannm'

Caused by:
- username already taken`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	wrapped := wrap.Errors("wrapped errors", err1, err2)

	expected := `wrapped errors

Caused by:
- error 1
- error 2`

	assertEqualErrorStrings(t, wrapped, expected)
}

func TestNestedError(t *testing.T) {
	err := errors.New("error")
	inner := wrap.Error(err, "inner wrapped error")
	outer := wrap.Error(inner, "outer wrapped error")

	expected := `outer wrapped error

Caused by:
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

Caused by:
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

Caused by:
- wrapped multiline
  errors
  - multiline
    error 1
  - multiline
    error 2`

	assertEqualErrorStrings(t, outer, expected)
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