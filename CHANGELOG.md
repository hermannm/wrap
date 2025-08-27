# Changelog

## [v0.4.0] - 2025-08-27

- Add `ctxwrap` subpackage for attaching `context.Context` to errors (see README for more on this)
- Add `wrap.ErrorWithAttrs` function for attaching structured log attributes to errors
- Add `wrap.Errorsf` function for wrapping multiple errors with a formatted message
- **Breaking:** Change variadic parameter in `wrap.Errors` to slice of errors
    - This makes the signature of this function more consistent with the other multi-error-wrapping
      functions in this package
    - To migrate, replace:
      ```go
      wrap.Errors("error message", err1, err2)
      ```
      ...with:
      ```go
      wrap.Errors([]error{err1, err2}, "error message")
      ```
- Make error unwrapping of plain errors more robust (check for `Unwrap() error` method instead of
  just splitting on ": ")

## [v0.3.1] - 2023-11-20

- Change  `wrap.Errors` to only indent error list if there is more than 1 error

## [v0.3.0] - 2023-11-15

- Implement unwrapping of long external errors
- Implement error interfaces from `hermannm.dev/devlog/log` to enable log message formatting
- Unexport `WrappedError` / `WrappedErrors`

## [v0.2.1] - 2023-11-11

- Make `WrappedError` and `WrappedErrors` types public

## [v0.2.0] - 2023-09-08

- Remove 'Caused by' header from error format

## [v0.1.0] - 2023-09-06

- Initial release

[Unreleased]: https://github.com/hermannm/wrap/compare/v0.4.0...HEAD

[v0.4.0]: https://github.com/hermannm/wrap/compare/v0.3.1...v0.4.0

[v0.3.1]: https://github.com/hermannm/wrap/compare/v0.3.0...v0.3.1

[v0.3.0]: https://github.com/hermannm/wrap/compare/v0.2.1...v0.3.0

[v0.2.1]: https://github.com/hermannm/wrap/compare/v0.2.0...v0.2.1

[v0.2.0]: https://github.com/hermannm/wrap/compare/v0.1.0...v0.2.0

[v0.1.0]: https://github.com/hermannm/wrap/compare/f9adbb2...v0.1.0
