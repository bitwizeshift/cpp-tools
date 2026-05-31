package diagnostic

import (
	"io"
	"log/slog"
)

// NewJSONLogger returns a Logger that emits each Diagnostic as one JSON line
// to w.
func NewJSONLogger(w io.Writer) *Logger {
	return NewLogger(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}))
}
