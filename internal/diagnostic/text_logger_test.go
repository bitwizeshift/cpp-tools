package diagnostic_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic"
	"rodusek.dev/pkg/cpp-tools/internal/term"
)

type textLoggerEnabler bool

func (e textLoggerEnabler) EnableColour(io.Writer) bool { return bool(e) }

func diagnosticRecord(level slog.Level, id, title, message, file string, lineStart, lineEnd, colStart, colEnd int64) slog.Record {
	rec := slog.NewRecord(time.Time{}, level, message, 0)
	attrs := []slog.Attr{
		slog.String("id", id),
		slog.String("title", title),
		slog.String("message", message),
		slog.String("severity", level.String()),
	}
	var locAttrs []any
	if file != "" {
		locAttrs = append(locAttrs, slog.String("file", file))
	}
	if lineStart != 0 {
		locAttrs = append(locAttrs, slog.Int64("line_start", lineStart))
	}
	if lineEnd != 0 {
		locAttrs = append(locAttrs, slog.Int64("line_end", lineEnd))
	}
	if colStart != 0 {
		locAttrs = append(locAttrs, slog.Int64("column_start", colStart))
	}
	if colEnd != 0 {
		locAttrs = append(locAttrs, slog.Int64("column_end", colEnd))
	}
	if len(locAttrs) > 0 {
		attrs = append(attrs, slog.Group("source", locAttrs...))
	}
	rec.AddAttrs(attrs...)
	return rec
}

func TestNewTextLogger_NonTTYWriter_EmitsPlainText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input *diagnostic.Diagnostic
		want  string
	}{
		{
			name: "ErrorWithIDAndDistinctMessage",
			input: &diagnostic.Diagnostic{
				ID:      "E001",
				Title:   "headline",
				Message: "details",
			},
			want: "error[E001]: headline\n   |\n   = details\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			logger := diagnostic.NewTextLogger(&buf)

			// Act
			logger.Error(context.Background(), tc.input)

			// Assert
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Handle(...) got %q, want %q", got, want)
			}
		})
	}
}

func TestTextHandler_Enabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
		want  bool
	}{
		{
			name:  "Debug",
			level: slog.LevelDebug,
			want:  true,
		}, {
			name:  "Info",
			level: slog.LevelInfo,
			want:  true,
		}, {
			name:  "Warn",
			level: slog.LevelWarn,
			want:  true,
		}, {
			name:  "Error",
			level: slog.LevelError,
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := diagnostic.NewTextHandler(
				&bytes.Buffer{},
				textLoggerEnabler(false),
				term.FixedSizer(80),
			)

			// Act
			got := sut.Enabled(context.Background(), tc.level)

			// Assert
			if got, want := got, tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Enabled(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestTextHandler_WithAttrs_ReturnsReceiver(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := diagnostic.NewTextHandler(
		&bytes.Buffer{},
		textLoggerEnabler(false),
		term.FixedSizer(80),
	)

	// Act
	got := sut.WithAttrs([]slog.Attr{slog.String("k", "v")})

	// Assert
	if got, want := got, slog.Handler(sut); got != want {
		t.Errorf("TextHandler.WithAttrs(...) got %v, want %v", got, want)
	}
}

func TestTextHandler_WithGroup_ReturnsReceiver(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := diagnostic.NewTextHandler(
		&bytes.Buffer{},
		textLoggerEnabler(false),
		term.FixedSizer(80),
	)

	// Act
	got := sut.WithGroup("g")

	// Assert
	if got, want := got, slog.Handler(sut); got != want {
		t.Errorf("TextHandler.WithGroup(...) got %v, want %v", got, want)
	}
}

func TestTextHandler_Handle_PlainText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		level   slog.Level
		id      string
		title   string
		message string
		file    string
		want    string
	}{
		{
			name:    "WithIDTitleEqualsMessage",
			level:   slog.LevelError,
			id:      "E1",
			title:   "boom",
			message: "boom",
			want:    "error[E1]: boom\n\n",
		}, {
			name:    "NoID",
			level:   slog.LevelWarn,
			id:      "",
			title:   "t",
			message: "t",
			want:    "warn: t\n\n",
		}, {
			name:    "TitleAndMessageDistinct",
			level:   slog.LevelError,
			id:      "E2",
			title:   "headline",
			message: "details",
			want:    "error[E2]: headline\n   |\n   = details\n\n",
		}, {
			name:    "TitleEmptyMessageSet",
			level:   slog.LevelInfo,
			id:      "I1",
			title:   "",
			message: "m",
			want:    "info[I1]: m\n\n",
		}, {
			name:    "TitleSetMessageEmpty",
			level:   slog.LevelInfo,
			id:      "I2",
			title:   "t",
			message: "",
			want:    "info[I2]: t\n\n",
		}, {
			name:    "BothEmpty",
			level:   slog.LevelError,
			id:      "",
			title:   "",
			message: "",
			want:    "error: \n\n",
		}, {
			name:    "WithLocation",
			level:   slog.LevelError,
			id:      "E3",
			title:   "t",
			message: "m",
			file:    "main.cpp",
			want:    "error[E3]: t\n  --> \x1b[4mmain.cpp\x1b[0m\n   |\n   = m\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			sut := diagnostic.NewTextHandler(
				&buf,
				textLoggerEnabler(false),
				term.FixedSizer(80),
			)
			rec := diagnosticRecord(tc.level, tc.id, tc.title, tc.message, tc.file, 0, 0, 0, 0)

			// Act
			_ = sut.Handle(context.Background(), rec)

			// Assert
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Handle(...) got %q, want %q", got, want)
			}
		})
	}
}

func TestTextHandler_Handle_SeverityColours(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
		want  string
	}{
		{
			name:  "Error",
			level: slog.LevelError,
			want:  "\x1b[91merror[\x1b[0m\x1b[97mE1\x1b[0m\x1b[91m]\x1b[0m: t\n   \x1b[94m|\x1b[0m\n   \x1b[94m=\x1b[0m m\n\n",
		}, {
			name:  "Warn",
			level: slog.LevelWarn,
			want:  "\x1b[93mwarn[\x1b[0m\x1b[97mE1\x1b[0m\x1b[93m]\x1b[0m: t\n   \x1b[94m|\x1b[0m\n   \x1b[94m=\x1b[0m m\n\n",
		}, {
			name:  "Info",
			level: slog.LevelInfo,
			want:  "\x1b[92minfo[\x1b[0m\x1b[97mE1\x1b[0m\x1b[92m]\x1b[0m: t\n   \x1b[94m|\x1b[0m\n   \x1b[94m=\x1b[0m m\n\n",
		}, {
			name:  "DebugUsesDefaultColour",
			level: slog.LevelDebug,
			want:  "\x1b[96mdebug[\x1b[0m\x1b[97mE1\x1b[0m\x1b[96m]\x1b[0m: t\n   \x1b[94m|\x1b[0m\n   \x1b[94m=\x1b[0m m\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			sut := diagnostic.NewTextHandler(
				&buf,
				textLoggerEnabler(true),
				term.FixedSizer(80),
			)
			rec := diagnosticRecord(tc.level, "E1", "t", "m", "", 0, 0, 0, 0)

			// Act
			_ = sut.Handle(context.Background(), rec)

			// Assert
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Handle(...) got %q, want %q", got, want)
			}
		})
	}
}

func TestTextHandler_Handle_LocationVariants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		file      string
		lineStart int64
		lineEnd   int64
		colStart  int64
		colEnd    int64
		want      string
	}{
		{
			name:      "AllZero",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "error[E1]: t\n   |\n   = m\n\n",
		}, {
			name:      "FileOnly",
			file:      "a.cpp",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp\x1b[0m\n   |\n   = m\n\n",
		}, {
			name:      "FileAndLineStart",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp:10\x1b[0m\n   |\n   = m\n\n",
		}, {
			name:      "FileAndLineRangeEqual",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   10,
			colStart:  0,
			colEnd:    0,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp:10\x1b[0m\n   |\n   = m\n\n",
		}, {
			name:      "FileAndLineRangeDifferent",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   12,
			colStart:  0,
			colEnd:    0,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp:10-12\x1b[0m\n   |\n   = m\n\n",
		}, {
			name:      "FileLineColumnStart",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  5,
			colEnd:    0,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp:10:5\x1b[0m\n   |\n   = m\n\n",
		}, {
			name:      "FileLineColumnRangeEqual",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  5,
			colEnd:    5,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp:10:5\x1b[0m\n   |\n   = m\n\n",
		}, {
			name:      "FileLineColumnRangeDifferent",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  5,
			colEnd:    8,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp:10:5-8\x1b[0m\n   |\n   = m\n\n",
		}, {
			name:      "FileLineRangeColumnRange",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   12,
			colStart:  5,
			colEnd:    8,
			want:      "error[E1]: t\n  --> \x1b[4ma.cpp:10-12:5-8\x1b[0m\n   |\n   = m\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			sut := diagnostic.NewTextHandler(
				&buf,
				textLoggerEnabler(false),
				term.FixedSizer(80),
			)
			rec := diagnosticRecord(
				slog.LevelError, "E1", "t", "m",
				tc.file, tc.lineStart, tc.lineEnd, tc.colStart, tc.colEnd,
			)

			// Act
			_ = sut.Handle(context.Background(), rec)

			// Assert
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("TextHandler.Handle(...) got %q, want %q", got, want)
			}
		})
	}
}
