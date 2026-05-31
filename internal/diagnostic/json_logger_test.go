package diagnostic_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic"
)

func TestNewJSONLogger_WritesJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
		input *diagnostic.Diagnostic
		want  map[string]any
	}{
		{
			name:  "ErrorWithoutLocation",
			level: slog.LevelError,
			input: &diagnostic.Diagnostic{
				ID:      "E1",
				Title:   "title",
				Message: "message",
			},
			want: map[string]any{
				"level":    "ERROR",
				"msg":      "message",
				"id":       "E1",
				"title":    "title",
				"message":  "message",
				"severity": "ERROR",
			},
		}, {
			name:  "WarnWithPartialLocation",
			level: slog.LevelWarn,
			input: &diagnostic.Diagnostic{
				ID:      "W1",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File:      "a.cpp",
					LineStart: 5,
				},
			},
			want: map[string]any{
				"level":    "WARN",
				"msg":      "message",
				"id":       "W1",
				"title":    "title",
				"message":  "message",
				"severity": "WARN",
				"source": map[string]any{
					"file":       "a.cpp",
					"line_start": json.Number("5"),
				},
			},
		}, {
			name:  "InfoWithFullLocation",
			level: slog.LevelInfo,
			input: &diagnostic.Diagnostic{
				ID:      "I1",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File:        "b.cpp",
					LineStart:   1,
					LineEnd:     2,
					ColumnStart: 3,
					ColumnEnd:   4,
				},
			},
			want: map[string]any{
				"level":    "INFO",
				"msg":      "message",
				"id":       "I1",
				"title":    "title",
				"message":  "message",
				"severity": "INFO",
				"source": map[string]any{
					"file":         "b.cpp",
					"line_start":   json.Number("1"),
					"line_end":     json.Number("2"),
					"column_start": json.Number("3"),
					"column_end":   json.Number("4"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			logger := diagnostic.NewJSONLogger(&buf)
			ctx := context.Background()

			// Act
			switch tc.level {
			case slog.LevelError:
				logger.Error(ctx, tc.input)
			case slog.LevelWarn:
				logger.Warn(ctx, tc.input)
			case slog.LevelInfo:
				logger.Info(ctx, tc.input)
			}

			// Assert
			var got map[string]any
			dec := json.NewDecoder(bytes.NewReader(buf.Bytes()))
			dec.UseNumber()
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("json.Decode(%q) failed: %v", buf.String(), err)
			}
			delete(got, "time")
			if got, want := got, tc.want; !cmp.Equal(got, want) {
				t.Errorf("NewJSONLogger(...) got %v, want %v", got, want)
			}
		})
	}
}
