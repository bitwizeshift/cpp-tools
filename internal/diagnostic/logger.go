package diagnostic

import (
	"context"
	"log/slog"
)

// Logger reports Diagnostics at slog severity levels through an underlying
// slog.Handler.
type Logger struct {
	slog *slog.Logger
}

// NewLogger returns a Logger that dispatches to handler.
func NewLogger(handler slog.Handler) *Logger {
	return &Logger{
		slog: slog.New(handler),
	}
}

// Debug reports d at slog.LevelDebug.
func (l *Logger) Debug(ctx context.Context, d *Diagnostic) {
	l.log(ctx, slog.LevelDebug, d)
}

// Error reports d at slog.LevelError.
func (l *Logger) Error(ctx context.Context, d *Diagnostic) {
	l.log(ctx, slog.LevelError, d)
}

// Warn reports d at slog.LevelWarn.
func (l *Logger) Warn(ctx context.Context, d *Diagnostic) {
	l.log(ctx, slog.LevelWarn, d)
}

// Info reports d at slog.LevelInfo.
func (l *Logger) Info(ctx context.Context, d *Diagnostic) {
	l.log(ctx, slog.LevelInfo, d)
}

func (l *Logger) log(ctx context.Context, level slog.Level, d *Diagnostic) {
	record := d.toRecord(level)

	_ = l.slog.Handler().Handle(ctx, record)
}
