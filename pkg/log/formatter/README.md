# log/formatter - Pluggable log entry formatters

## Overview

Package `formatter` defines the `LogFormatter` interface and three concrete
implementations — plain text, JSON, and ANSI-colored text — used by the ETH2030
structured logger. Each formatter turns a `LogEntry` (timestamp, level, message,
and arbitrary key-value fields) into a single printable string.

Fields within a log entry are always rendered in sorted key order so that output
is deterministic and easy to grep, regardless of the order in which fields were
attached to the entry.

## Functionality

**Types**

- `LogLevel` (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`) with `String()` and
  `LevelFromString(s string) LogLevel` (case-insensitive, defaults to `INFO`).
- `LogEntry` - `Timestamp time.Time`, `Level LogLevel`, `Message string`,
  `Fields map[string]interface{}`.
- `LogFormatter` interface - single method `Format(entry LogEntry) string`.

**Formatters**

- `TextFormatter` - renders `[2006-01-02 15:04:05] LEVEL message key=value`.
  Field `TimeFormat` is optional (defaults to the above layout).
- `JSONFormatter` - renders a single-line JSON object with keys `time`, `level`,
  `msg`, plus all entry fields. Defaults to `time.RFC3339`. Never panics: returns
  a best-effort fallback string on marshal failure.
- `ColorFormatter` - same layout as `TextFormatter` but wraps the level name in
  ANSI escape codes: gray (DEBUG), green (INFO), yellow (WARN), red (ERROR),
  bold red (FATAL).

All three formatters are zero-value usable after construction with `&TextFormatter{}`,
`&JSONFormatter{}`, or `&ColorFormatter{}`.

Parent package: [`log`](../)
