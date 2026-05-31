package diagnostictest

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic"
)

// EquateDiagnostics returns a cmp.Option that compares *diagnostic.Diagnostic
// values by ID, Title, Message, and Location. When either side carries an
// Err, equality is decided by cmpopts.EquateErrors and the other fields are
// ignored.
func EquateDiagnostics() cmp.Option {
	return cmp.Comparer(func(lhs, rhs *diagnostic.Diagnostic) bool {
		if lhs == nil && rhs == nil {
			return true
		}
		if lhs == nil || rhs == nil {
			return false
		}
		if lhs.Err != nil || rhs.Err != nil {
			return cmp.Equal(lhs.Err, rhs.Err, cmpopts.EquateErrors())
		}
		return lhs.ID == rhs.ID &&
			lhs.Title == rhs.Title &&
			lhs.Message == rhs.Message &&
			cmp.Equal(lhs.Location, rhs.Location)
	})
}

// NewLogger returns a diagnostic.Logger that appends each reported
// Diagnostic to *recorder. Calls are safe for concurrent use.
func NewLogger(recorder *[]diagnostic.Diagnostic) *diagnostic.Logger {
	return diagnostic.NewLogger(&logRecorder{
		recorder: recorder,
	})
}

type logRecorder struct {
	mu       sync.Mutex
	recorder *[]diagnostic.Diagnostic
}

func (lr *logRecorder) Enabled(context.Context, slog.Level) bool {
	return true
}

func (lr *logRecorder) Handle(_ context.Context, r slog.Record) error {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	fields := map[string]any{}
	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindGroup {
			for _, inner := range a.Value.Group() {
				fields[inner.Key] = inner.Value.Any()
			}
		} else {
			fields[a.Key] = a.Value.Any()
		}
		return true
	})

	d := diagnostic.Diagnostic{
		ID:      fields["id"].(string),
		Title:   fields["title"].(string),
		Message: fields["message"].(string),
	}
	if file, ok := fields["file"].(string); ok {
		lineStart, _ := fields["line_start"].(int64)
		lineEnd, _ := fields["line_end"].(int64)
		columnStart, _ := fields["column_start"].(int64)
		columnEnd, _ := fields["column_end"].(int64)
		d.Location = &diagnostic.Location{
			File:        file,
			LineStart:   int(lineStart),
			LineEnd:     int(lineEnd),
			ColumnStart: int(columnStart),
			ColumnEnd:   int(columnEnd),
		}
	}

	*lr.recorder = append(*lr.recorder, d)
	return nil
}

func (lr *logRecorder) WithGroup(string) slog.Handler {
	return lr
}

func (lr *logRecorder) WithAttrs([]slog.Attr) slog.Handler {
	return lr
}

var _ slog.Handler = (*logRecorder)(nil)
