# wrap

A small Go library for wrapping errors with extra context.

Run `go get hermannm.dev/wrap` to add it to your project!

**Docs:** [pkg.go.dev/hermannm.dev/wrap](https://pkg.go.dev/hermannm.dev/wrap)

**Contents:**

- [Motivation behind the library](#motivation-behind-the-library)
- [Usage](#usage)
- [The `ctxwrap` subpackage](#the-ctxwrap-subpackage)

## Motivation behind the library

Go's [`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf) is a great way to provide context to errors
before returning them. However, the common way of using `fmt.Errorf("extra context: %w", err)` can
lead to long and hard-to-read error messages. See the following example, albeit a bit contrived:

```
failed to register user: user creation failed: invalid user data: invalid username: username exceeds 30 characters
```

This library's `wrap.Error` aims to alleviate this, by instead formatting wrapped errors like this:

```
failed to register user
- user creation failed
- invalid user data
- invalid username
- username exceeds 30 characters
```

The library also provides:

- `wrap.Errorf` to use a format string for the wrapping message
- `wrap.Errors` to wrap multiple errors
- `wrap.ErrorWithAttrs` to attach structured log attributes (from the standard
  [`log/slog`](https://pkg.go.dev/log/slog) package), to provide better context when an error is
  logged
    - The error returned by this wrapper can be used by a logging library (such as
      [hermannm.dev/devlog/log](https://pkg.go.dev/hermannm.dev/devlog/log)) to add error attributes
      to the log output
- A `ctxwrap` subpackage to attach `context.Context` to errors
  ([see below](#the-ctxwrap-subpackage) for more on this)

## Usage

Basic usage:

```go
err := errors.New("expired token")
wrapped := wrap.Error(err, "user authentication failed")
fmt.Println(wrapped)
// user authentication failed
// - expired token
```

Wrapped errors can be nested. Wrapping an already wrapped error adds it to the error list, as
follows:

```go
wrapped2 := wrap.Error(wrapped, "failed to update username")
fmt.Println(wrapped2)
// failed to update username
// - user authentication failed
// - expired token
```

`wrap.Errorf` can be used to create the wrapping message with a format string:

```go
err := errors.New("username already taken")
wrapped := wrap.Errorf(err, "failed to create user with name '%s'", "hermannm")
fmt.Println(wrapped)
// failed to create user with name 'hermannm'
// - username already taken
```

...and `wrap.Errors` can be used to wrap multiple errors:

```go
errs := []error{errors.New("username too long"), errors.New("invalid email")}
wrapped := wrap.Errors(errs, "user creation failed")
fmt.Println(wrapped)
// user creation failed
// - username too long
// - invalid email
```

When combining `wrap.Errors` and `wrap.Error`, nested errors are indented as follows:

```go
errs := []error{errors.New("username too long"), errors.New("invalid email")}
inner := wrap.Errors(errs, "user creation failed")
outer := wrap.Error(inner, "failed to register new user")
fmt.Println(outer)
// failed to register new user
// - user creation failed
//   - username too long
//   - invalid email
```

Finally, `wrap.ErrorWithAttrs` lets you attach structured log attributes to errors. This can be used
by error-aware logging libraries, such as
[hermannm.dev/devlog/log](https://pkg.go.dev/hermannm.dev/devlog/log), to add the error's attributes
to the log output.

<!-- @formatter:off -->
```go
func example() error {
	req := ExternalServiceRequest { /* ... */ }
	resp, err := callExternalService(request)
	if != nil {
		// When this error is logged by a conformant logging library such as devlog/log,
		// the log output will have a "request" attribute with the given struct
		return wrap.ErrorWithAttrs(err, "Request to external service failed", "request", req) 
	}
}
```
<!-- @formatter:on -->

## The `ctxwrap` subpackage

This library also provides a `ctxwrap` subpackage, which tries to solve the problem of errors
escaping their context. It mirrors the API of `wrap`, but adds a
[`context.Context`](https://pkg.go.dev/context) parameter to every error wrapping function, so that
the error can carry its original context as it's returned up the stack.

Every error returned by this package implements the following method:

```go
Context() context.Context
```

Other libraries (e.g. a logging library) can check for this method, to use the error's original
context.

To see why you may want this, let's look at an example using the
[hermannm.dev/devlog/log](https://pkg.go.dev/hermannm.dev/devlog/log) logging library:

<!-- @formatter:off -->
```go
import (
	"context"

	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

func parentFunction(ctx context.Context) {
	if err := childFunction(ctx); err != nil {
		log.Error(ctx, err, "Child function failed")
	}
}

func childFunction(ctx context.Context) error {
	// log.AddContextAttrs adds log attributes to the context.
    // When ctx is logged, these attributes are included in the log output
	ctx = log.AddContextAttrs(ctx, "key", "value")

	if err := someFallibleOperation(ctx); err != nil {
		return wrap.Error(err, "Operation failed")
	}

	return nil
}
```
<!-- @formatter:on -->

In the above example, `childFunction` returns an error, and `parentFunction` logs it. This is a
typical pattern, as errors are often returned up the stack before being logged.

We see that we attach log attributes to the context in `childFunction`, using `log.AddContextAttrs`.
But when we return the error using wrap.Error, we lose those context attributes! This is not ideal,
as we want as much context as possible when an error is logged.

`ctxwrap` solves this by letting us attach the context to the error, so that the logging library
can get the context attributes from the error when it is logged. This revised example uses
`ctxwrap` instead of `wrap`, to propagate context attributes:

<!-- @formatter:off -->
```go
import (
	"context"

	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap/ctxwrap"
)

func parentFunction(ctx context.Context) {
	if err := childFunction(ctx); err != nil {
		log.Error(ctx, err, "Child function failed")
	}
}

func childFunction(ctx context.Context) error {
	ctx = log.AddContextAttrs(ctx, "key", "value")

	if err := someFallibleOperation(ctx); err != nil {
		// Uses ctxwrap to attach the context to the error
		return ctxwrap.Error(ctx, err, "Operation failed")
	}

	return nil
}
```
<!-- @formatter:on -->

Now, when `parentFunction` logs the error from `childFunction`, the context attributes carried by
the error will be logged, so we get more context in our error log!
