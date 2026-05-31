package diagnostic_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic/diagnostictest"
)

func TestNewLogger_DispatchesToHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input *diagnostic.Diagnostic
		want  []diagnostic.Diagnostic
	}{
		{
			name: "Records",
			input: &diagnostic.Diagnostic{
				ID:      "E0",
				Title:   "title",
				Message: "message",
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E0",
					Title:   "title",
					Message: "message",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var recorded []diagnostic.Diagnostic
			logger := diagnostictest.NewLogger(&recorded)

			// Act
			logger.Error(context.Background(), tc.input)

			// Assert
			opts := diagnostictest.EquateDiagnostics()
			if got, want := recorded, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Logger.Error(...) mismatch (-got +want):\n%s", cmp.Diff(got, want, opts))
			}
		})
	}
}

func TestLogger_Error_RecordsDiagnostic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input *diagnostic.Diagnostic
		want  []diagnostic.Diagnostic
	}{
		{
			name: "NoLocation",
			input: &diagnostic.Diagnostic{
				ID:      "E1",
				Title:   "title",
				Message: "message",
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E1",
					Title:   "title",
					Message: "message",
				},
			},
		}, {
			name: "ZeroValueLocation",
			input: &diagnostic.Diagnostic{
				ID:       "E2",
				Title:    "title",
				Message:  "message",
				Location: &diagnostic.Location{},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E2",
					Title:   "title",
					Message: "message",
				},
			},
		}, {
			name: "FileOnly",
			input: &diagnostic.Diagnostic{
				ID:      "E3",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File: "a.cpp",
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E3",
					Title:   "title",
					Message: "message",
					Location: &diagnostic.Location{
						File: "a.cpp",
					},
				},
			},
		}, {
			name: "FileAndLineStart",
			input: &diagnostic.Diagnostic{
				ID:      "E4",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File:      "a.cpp",
					LineStart: 1,
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E4",
					Title:   "title",
					Message: "message",
					Location: &diagnostic.Location{
						File:      "a.cpp",
						LineStart: 1,
					},
				},
			},
		}, {
			name: "FileAndLineEnd",
			input: &diagnostic.Diagnostic{
				ID:      "E5",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File:    "a.cpp",
					LineEnd: 2,
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E5",
					Title:   "title",
					Message: "message",
					Location: &diagnostic.Location{
						File:    "a.cpp",
						LineEnd: 2,
					},
				},
			},
		}, {
			name: "FileAndColumnStart",
			input: &diagnostic.Diagnostic{
				ID:      "E6",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File:        "a.cpp",
					ColumnStart: 3,
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E6",
					Title:   "title",
					Message: "message",
					Location: &diagnostic.Location{
						File:        "a.cpp",
						ColumnStart: 3,
					},
				},
			},
		}, {
			name: "FileAndColumnEnd",
			input: &diagnostic.Diagnostic{
				ID:      "E7",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File:      "a.cpp",
					ColumnEnd: 4,
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E7",
					Title:   "title",
					Message: "message",
					Location: &diagnostic.Location{
						File:      "a.cpp",
						ColumnEnd: 4,
					},
				},
			},
		}, {
			name: "AllFields",
			input: &diagnostic.Diagnostic{
				ID:      "E8",
				Title:   "title",
				Message: "message",
				Location: &diagnostic.Location{
					File:        "a.cpp",
					LineStart:   1,
					LineEnd:     2,
					ColumnStart: 3,
					ColumnEnd:   4,
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E8",
					Title:   "title",
					Message: "message",
					Location: &diagnostic.Location{
						File:        "a.cpp",
						LineStart:   1,
						LineEnd:     2,
						ColumnStart: 3,
						ColumnEnd:   4,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var recorded []diagnostic.Diagnostic
			logger := diagnostictest.NewLogger(&recorded)
			opts := diagnostictest.EquateDiagnostics()

			// Act
			logger.Error(context.Background(), tc.input)

			// Assert
			if got, want := recorded, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Logger.Error(...) mismatch (-got +want):\n%s", cmp.Diff(got, want, opts))
			}
		})
	}
}

func TestLogger_Warn_RecordsDiagnostic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input *diagnostic.Diagnostic
		want  []diagnostic.Diagnostic
	}{
		{
			name: "Records",
			input: &diagnostic.Diagnostic{
				ID:      "W1",
				Title:   "title",
				Message: "message",
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "W1",
					Title:   "title",
					Message: "message",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var recorded []diagnostic.Diagnostic
			logger := diagnostictest.NewLogger(&recorded)
			opts := diagnostictest.EquateDiagnostics()

			// Act
			logger.Warn(context.Background(), tc.input)

			// Assert
			if got, want := recorded, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Logger.Warn(...) mismatch (-got +want):\n%s", cmp.Diff(got, want, opts))
			}
		})
	}
}

func TestLogger_Info_RecordsDiagnostic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input *diagnostic.Diagnostic
		want  []diagnostic.Diagnostic
	}{
		{
			name: "Records",
			input: &diagnostic.Diagnostic{
				ID:      "I1",
				Title:   "title",
				Message: "message",
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "I1",
					Title:   "title",
					Message: "message",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var recorded []diagnostic.Diagnostic
			logger := diagnostictest.NewLogger(&recorded)
			opts := diagnostictest.EquateDiagnostics()

			// Act
			logger.Info(context.Background(), tc.input)

			// Assert
			if got, want := recorded, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Logger.Info(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestLogger_Debug_RecordsDiagnostic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input *diagnostic.Diagnostic
		want  []diagnostic.Diagnostic
	}{
		{
			name: "Records",
			input: &diagnostic.Diagnostic{
				ID:      "D1",
				Title:   "title",
				Message: "message",
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "D1",
					Title:   "title",
					Message: "message",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var recorded []diagnostic.Diagnostic
			logger := diagnostictest.NewLogger(&recorded)
			opts := diagnostictest.EquateDiagnostics()

			// Act
			logger.Debug(context.Background(), tc.input)

			// Assert
			if got, want := recorded, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Logger.Debug(...) mismatch (-got +want):\n%s", cmp.Diff(got, want, opts))
			}
		})
	}
}
