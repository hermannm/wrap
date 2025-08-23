package internal

import "log/slog"

// ParseAttrs is adapted from the standard library:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/attr.go#L71
func ParseAttrs(unparsed []any) []slog.Attr {
	parsed := make([]slog.Attr, 0, len(unparsed))
	var current slog.Attr

	for len(unparsed) > 0 {
		// - If unparsed[0] is an Attr, use that and continue
		// - If unparsed[0] is a string, the first two elements are a key-value pair
		// - Otherwise, it treats args[0] as a value with a missing key.
		switch attr := unparsed[0].(type) {
		case slog.Attr:
			current, unparsed = attr, unparsed[1:]
		case string:
			if len(unparsed) == 1 {
				current, unparsed = slog.String(badKey, attr), nil
			} else {
				current, unparsed = slog.Any(attr, unparsed[1]), unparsed[2:]
			}
		default:
			current, unparsed = slog.Any(badKey, attr), unparsed[1:]
		}

		parsed = append(parsed, current)
	}

	return parsed
}

// Same key as the one the standard library uses for attributes that failed to parse:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/record.go#L160
const badKey = "!BADKEY"
