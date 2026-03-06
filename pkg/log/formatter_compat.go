package log

// formatter_compat.go re-exports types from log/formatter for backward compatibility.

import "github.com/eth2030/eth2030/log/formatter"

// Log formatter type aliases.
type (
	LogLevel       = formatter.LogLevel
	LogEntry       = formatter.LogEntry
	LogFormatter   = formatter.LogFormatter
	TextFormatter  = formatter.TextFormatter
	JSONFormatter  = formatter.JSONFormatter
	ColorFormatter = formatter.ColorFormatter
)

// Log level constants.
const (
	DEBUG = formatter.DEBUG
	INFO  = formatter.INFO
	WARN  = formatter.WARN
	ERROR = formatter.ERROR
	FATAL = formatter.FATAL
)

// Log formatter function wrappers.
func LevelFromString(s string) LogLevel { return formatter.LevelFromString(s) }
