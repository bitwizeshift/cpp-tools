package ansi_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"rodusek.dev/pkg/cpp-tools/internal/ansi"
)

func TestColour_Format(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		colour ansi.Colour
		format string
		args   []any
		want   string
	}{
		{
			name:   "NoArgsPlainFormat",
			colour: ansi.Red,
			format: "hello",
			args:   nil,
			want:   "\x1b[31mhello\x1b[0m",
		}, {
			name:   "WithArgs",
			colour: ansi.Green,
			format: "%s %d",
			args:   []any{"foo", 1},
			want:   "\x1b[32mfoo 1\x1b[0m",
		}, {
			name:   "EmptyFormatString",
			colour: ansi.Blue,
			format: "",
			args:   nil,
			want:   "\x1b[34m\x1b[0m",
		}, {
			name:   "FormatContainsResetSequence",
			colour: ansi.Yellow,
			format: "before\x1b[0mafter",
			args:   nil,
			want:   "\x1b[33mbefore\x1b[0mafter\x1b[0m",
		}, {
			name:   "FormatContainsPercentLiteral",
			colour: ansi.Cyan,
			format: "100%% done",
			args:   nil,
			want:   "\x1b[36m100% done\x1b[0m",
		}, {
			name:   "ReceiverIsReset",
			colour: ansi.Reset,
			format: "body",
			args:   nil,
			want:   "\x1b[0mbody\x1b[0m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.colour
			format := tc.format
			args := tc.args

			// Act
			formatted := sut.Format(format, args...)

			// Assert
			if got, want := formatted, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Colour.Format(...) got %q, want %q", got, want)
			}
		})
	}
}
