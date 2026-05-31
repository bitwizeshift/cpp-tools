package diagnostictest_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic/diagnostictest"
)

func TestEquateDiagnostics(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("sentinel")
	other := errors.New("other")

	testCases := []struct {
		name string
		lhs  *diagnostic.Diagnostic
		rhs  *diagnostic.Diagnostic
		want bool
	}{
		{
			name: "BothNil",
			lhs:  nil,
			rhs:  nil,
			want: true,
		}, {
			name: "LhsNilRhsSet",
			lhs:  nil,
			rhs:  &diagnostic.Diagnostic{ID: "id"},
			want: false,
		}, {
			name: "LhsSetRhsNil",
			lhs:  &diagnostic.Diagnostic{ID: "id"},
			rhs:  nil,
			want: false,
		}, {
			name: "BothErrWrappedSentinel",
			lhs:  &diagnostic.Diagnostic{Err: sentinel},
			rhs:  &diagnostic.Diagnostic{Err: fmt.Errorf("wrap: %w", sentinel)},
			want: true,
		}, {
			name: "BothErrUnrelated",
			lhs:  &diagnostic.Diagnostic{Err: sentinel},
			rhs:  &diagnostic.Diagnostic{Err: other},
			want: false,
		}, {
			name: "OnlyLhsErr",
			lhs:  &diagnostic.Diagnostic{Err: sentinel},
			rhs:  &diagnostic.Diagnostic{},
			want: false,
		}, {
			name: "OnlyRhsErr",
			lhs:  &diagnostic.Diagnostic{},
			rhs:  &diagnostic.Diagnostic{Err: sentinel},
			want: false,
		}, {
			name: "AllFieldsEqualWithLocation",
			lhs: &diagnostic.Diagnostic{
				ID:      "id",
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
			rhs: &diagnostic.Diagnostic{
				ID:      "id",
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
			want: true,
		}, {
			name: "DifferInID",
			lhs:  &diagnostic.Diagnostic{ID: "id1", Title: "t", Message: "m"},
			rhs:  &diagnostic.Diagnostic{ID: "id2", Title: "t", Message: "m"},
			want: false,
		}, {
			name: "DifferInTitle",
			lhs:  &diagnostic.Diagnostic{ID: "id", Title: "t1", Message: "m"},
			rhs:  &diagnostic.Diagnostic{ID: "id", Title: "t2", Message: "m"},
			want: false,
		}, {
			name: "DifferInMessage",
			lhs:  &diagnostic.Diagnostic{ID: "id", Title: "t", Message: "m1"},
			rhs:  &diagnostic.Diagnostic{ID: "id", Title: "t", Message: "m2"},
			want: false,
		}, {
			name: "DifferInLocationOneNil",
			lhs:  &diagnostic.Diagnostic{ID: "id", Title: "t", Message: "m"},
			rhs: &diagnostic.Diagnostic{
				ID:       "id",
				Title:    "t",
				Message:  "m",
				Location: &diagnostic.Location{File: "a.cpp"},
			},
			want: false,
		}, {
			name: "DifferInLocationBothSet",
			lhs: &diagnostic.Diagnostic{
				ID:       "id",
				Title:    "t",
				Message:  "m",
				Location: &diagnostic.Location{File: "a.cpp", LineStart: 1},
			},
			rhs: &diagnostic.Diagnostic{
				ID:       "id",
				Title:    "t",
				Message:  "m",
				Location: &diagnostic.Location{File: "b.cpp", LineStart: 2},
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			opt := diagnostictest.EquateDiagnostics()

			// Act
			got := cmp.Equal(tc.lhs, tc.rhs, opt)

			// Assert
			if got, want := got, tc.want; !cmp.Equal(got, want) {
				t.Errorf("EquateDiagnostics()(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestNewLogger_RecordsDiagnostics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		inputs []*diagnostic.Diagnostic
		want   []diagnostic.Diagnostic
	}{
		{
			name: "SingleDiagnosticNoLocation",
			inputs: []*diagnostic.Diagnostic{
				{ID: "E1", Title: "t", Message: "m"},
			},
			want: []diagnostic.Diagnostic{
				{ID: "E1", Title: "t", Message: "m"},
			},
		}, {
			name: "SingleDiagnosticFullLocation",
			inputs: []*diagnostic.Diagnostic{
				{
					ID:      "E2",
					Title:   "t",
					Message: "m",
					Location: &diagnostic.Location{
						File:        "a.cpp",
						LineStart:   1,
						LineEnd:     2,
						ColumnStart: 3,
						ColumnEnd:   4,
					},
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E2",
					Title:   "t",
					Message: "m",
					Location: &diagnostic.Location{
						File:        "a.cpp",
						LineStart:   1,
						LineEnd:     2,
						ColumnStart: 3,
						ColumnEnd:   4,
					},
				},
			},
		}, {
			name: "SingleDiagnosticPartialLocation",
			inputs: []*diagnostic.Diagnostic{
				{
					ID:      "E3",
					Title:   "t",
					Message: "m",
					Location: &diagnostic.Location{
						File:      "a.cpp",
						LineStart: 5,
					},
				},
			},
			want: []diagnostic.Diagnostic{
				{
					ID:      "E3",
					Title:   "t",
					Message: "m",
					Location: &diagnostic.Location{
						File:      "a.cpp",
						LineStart: 5,
					},
				},
			},
		}, {
			name: "MultipleDiagnosticsInOrder",
			inputs: []*diagnostic.Diagnostic{
				{ID: "E1", Title: "t1", Message: "m1"},
				{ID: "E2", Title: "t2", Message: "m2"},
				{
					ID:      "E3",
					Title:   "t3",
					Message: "m3",
					Location: &diagnostic.Location{
						File:      "c.cpp",
						LineStart: 7,
					},
				},
			},
			want: []diagnostic.Diagnostic{
				{ID: "E1", Title: "t1", Message: "m1"},
				{ID: "E2", Title: "t2", Message: "m2"},
				{
					ID:      "E3",
					Title:   "t3",
					Message: "m3",
					Location: &diagnostic.Location{
						File:      "c.cpp",
						LineStart: 7,
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
			ctx := context.Background()

			// Act
			for _, d := range tc.inputs {
				logger.Error(ctx, d)
			}

			// Assert
			opts := diagnostictest.EquateDiagnostics()
			if got, want := recorded, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Logger.Error(...) mismatch (-got +want):\n%s", cmp.Diff(got, want, opts))
			}
		})
	}
}
