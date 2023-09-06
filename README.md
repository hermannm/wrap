# wrap

A small Go package to wrap errors with extra context in an easy-to-read format.

Run `go get hermannm.dev/wrap` to add it to your project!

## Rationale

Go's [`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf) is a great way to provide context to errors
before returning them. However, the common way of using `fmt.Errorf("extra context: %w", err)` can
lead to long and hard-to-read error messages. See the following example, albeit a bit contrived:
```
failed to register user: user creation failed: invalid user data: invalid username: username exceeds 30 characters
```

This package's `wrap.Error` aims to alleviate this, by instead displaying wrapped errors like this:
```
failed to register user

Caused by:
- user creation failed
- invalid user data
- invalid username
- username exceeds 30 characters
```

The package also provides `wrap.Errorf` to use a format string for the wrapping message, and
`wrap.Errors` to wrap multiple errors.

## Usage

Basic usage:

```go
err := errors.New("expired token")
wrapped := wrap.Error(err, "user authentication failed")
fmt.Println(wrapped)
// user authentication failed
//
// Caused by:
// - expired token
```

Wrapped errors can be nested. Wrapping an already wrapped error adds it to the 'Caused by' list, as
follows:

```go
wrapped2 := wrap.Error(wrapped, "failed to update username")
fmt.Println(wrapped2)
// failed to update username
//
// Caused by:
// - user authentication failed
// - expired token
```

`wrap.Errorf` can be used to create the wrapping message with a format string:

```go
err := errors.New("username already taken")
wrapped := wrap.Errorf(err, "failed to create user with name '%s'", "hermannm")
fmt.Println(wrapped)
// failed to create user with name 'hermannm'
//
// Caused by:
// - username already taken
```

...and `wrap.Errors` can be used to wrap multiple errors:

```go
err1 := errors.New("username too long")
err2 := errors.New("invalid email")
wrapped := wrap.Errors("user creation failed", err1, err2)
fmt.Println(wrapped)
// user creation failed
//
// Caused by:
// - username too long
// - invalid email
```

When combining `wrap.Errors` and `wrap.Error`, nested errors are indented as follows:

```go
err1 := errors.New("username too long")
err2 := errors.New("invalid email")
inner := wrap.Errors("user creation failed", err1, err2)
outer := wrap.Error(inner, "failed to register new user")
fmt.Println(outer)
// failed to register new user
//
// Caused by:
// - user creation failed
//   - username too long
//   - invalid email
```

## Credits

- Rust's [anyhow](https://crates.io/crates/anyhow) crate, which inspired the error format
