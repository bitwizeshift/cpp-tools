package modified_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/command"
	"rodusek.dev/pkg/cpp-tools/internal/command/commandtest"
	"rodusek.dev/pkg/cpp-tools/internal/modified"
)

func TestCommandYearRangerYearRange(t *testing.T) {
	t.Parallel()

	errStdoutPipe := errors.New("stdout pipe failed")
	errStart := errors.New("start failed")
	errWait := errors.New("wait failed")

	testCases := []struct {
		name      string
		creator   command.Creator
		wantStart int
		wantEnd   int
		wantErr   error
	}{
		{
			name:      "MultipleYearsReturnsMinAndMax",
			creator:   commandtest.Pipes("2020\n2018\n2025\n2018\n", ""),
			wantStart: 2018,
			wantEnd:   2025,
			wantErr:   nil,
		},
		{
			name:      "EmptyStdoutReturnsError",
			creator:   commandtest.Pipes("", ""),
			wantStart: 0,
			wantEnd:   0,
			wantErr:   cmpopts.AnyError,
		},
		{
			name:      "BlankAndNonNumericValuesAreSkipped",
			creator:   commandtest.Pipes("\n  \nnot-a-year\n2021\n   2019   \nabc\n2023\n", ""),
			wantStart: 2019,
			wantEnd:   2023,
			wantErr:   nil,
		},
		{
			name:      "SingleYearReturnsThatYear",
			creator:   commandtest.Pipes("2022\n", ""),
			wantStart: 2022,
			wantEnd:   2022,
			wantErr:   nil,
		},
		{
			name:      "StdoutPipeErrorReturnsError",
			creator:   commandtest.ErrOnStdoutPipe(errStdoutPipe),
			wantStart: 0,
			wantEnd:   0,
			wantErr:   errStdoutPipe,
		},
		{
			name:      "StartErrorReturnsError",
			creator:   commandtest.ErrOnStart(errStart),
			wantStart: 0,
			wantEnd:   0,
			wantErr:   errStart,
		},
		{
			name:      "WaitErrorReturnsError",
			creator:   commandtest.ErrOnWait(errWait),
			wantStart: 0,
			wantEnd:   0,
			wantErr:   errWait,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			sut := modified.CommandYearRanger{
				CommandCreator: tc.creator,
			}

			// Act
			start, end, err := sut.YearRange(ctx, "some/path")

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("CommandYearRanger.YearRange(...) err = %v, want %v", got, want)
			}
			if got, want := start, tc.wantStart; !cmp.Equal(got, want) {
				t.Errorf("CommandYearRanger.YearRange(...) start = %v, want %v", got, want)
			}
			if got, want := end, tc.wantEnd; !cmp.Equal(got, want) {
				t.Errorf("CommandYearRanger.YearRange(...) end = %v, want %v", got, want)
			}
		})
	}
}
