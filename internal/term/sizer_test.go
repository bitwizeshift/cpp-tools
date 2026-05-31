package term_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"rodusek.dev/pkg/cpp-tools/internal/term"
)

// fdWriter is a writer that exposes an Fd() method, simulating a terminal.
type fdWriter struct {
	bytes.Buffer
	fd uintptr
}

func (f *fdWriter) Fd() uintptr { return f.fd }

// fakeSizer is a test double for the Sizer interface.
type fakeSizer struct {
	result int
}

func (f fakeSizer) Columns(io.Writer) int { return f.result }

func TestEnvSizer_Columns(t *testing.T) {
	const varName = "TERM_TEST_COLUMNS"

	testCases := []struct {
		name     string
		variable string
		set      bool
		value    string
		want     int
	}{
		{
			name:     "VariableNotSet",
			variable: varName,
			set:      false,
			value:    "",
			want:     0,
		}, {
			name:     "VariableSetToValidInteger",
			variable: varName,
			set:      true,
			value:    "80",
			want:     80,
		}, {
			name:     "VariableSetToNonNumeric",
			variable: varName,
			set:      true,
			value:    "wide",
			want:     0,
		}, {
			name:     "VariableSetToEmpty",
			variable: varName,
			set:      true,
			value:    "",
			want:     0,
		}, {
			name:     "VariableSetToZero",
			variable: varName,
			set:      true,
			value:    "0",
			want:     0,
		}, {
			name:     "VariableSetToNegative",
			variable: varName,
			set:      true,
			value:    "-1",
			want:     -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			if tc.set {
				t.Setenv(tc.variable, tc.value)
			}
			sizer := term.EnvSizer{Variable: tc.variable}

			// Act
			cols := sizer.Columns(&bytes.Buffer{})

			// Assert
			if got, want := cols, tc.want; !cmp.Equal(got, want) {
				t.Errorf("EnvSizer.Columns(...) got %d, want %d", got, want)
			}
		})
	}
}

func TestFixedSizer_Columns(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		sizer term.FixedSizer
		want  int
	}{
		{
			name:  "Zero",
			sizer: term.FixedSizer(0),
			want:  0,
		}, {
			name:  "Positive",
			sizer: term.FixedSizer(80),
			want:  80,
		}, {
			name:  "Negative",
			sizer: term.FixedSizer(-1),
			want:  -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sizer := tc.sizer

			// Act
			cols := sizer.Columns(&bytes.Buffer{})

			// Assert
			if got, want := cols, tc.want; !cmp.Equal(got, want) {
				t.Errorf("FixedSizer.Columns(...) got %d, want %d", got, want)
			}
		})
	}
}

func TestTTYFuncSizer_Columns(t *testing.T) {
	t.Parallel()

	var errGetSize = errors.New("get size failed")

	testCases := []struct {
		name    string
		writer  io.Writer
		cols    int
		funcErr error
		want    int
	}{
		{
			name:    "WriterWithFdFuncSucceeds",
			writer:  &fdWriter{fd: 1},
			cols:    120,
			funcErr: nil,
			want:    120,
		}, {
			name:    "WriterWithFdFuncFails",
			writer:  &fdWriter{fd: 1},
			cols:    0,
			funcErr: errGetSize,
			want:    0,
		}, {
			name:    "WriterWithoutFd",
			writer:  &bytes.Buffer{},
			cols:    120,
			funcErr: nil,
			want:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sizer := term.TTYFuncSizer(func(int) (int, int, error) {
				return tc.cols, 0, tc.funcErr
			})

			// Act
			cols := sizer.Columns(tc.writer)

			// Assert
			if got, want := cols, tc.want; !cmp.Equal(got, want) {
				t.Errorf("TTYFuncSizer.Columns(...) got %d, want %d", got, want)
			}
		})
	}
}

func TestSaturateSizer_Columns(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		min   int
		max   int
		sizer fakeSizer
		want  int
	}{
		{
			name:  "InnerReturnsZero",
			min:   50,
			max:   100,
			sizer: fakeSizer{result: 0},
			want:  0,
		}, {
			name:  "InnerWithinRange",
			min:   50,
			max:   100,
			sizer: fakeSizer{result: 80},
			want:  80,
		}, {
			name:  "InnerBelowMin",
			min:   50,
			max:   100,
			sizer: fakeSizer{result: 20},
			want:  50,
		}, {
			name:  "InnerAboveMax",
			min:   50,
			max:   100,
			sizer: fakeSizer{result: 200},
			want:  100,
		}, {
			name:  "InnerExactlyMin",
			min:   50,
			max:   100,
			sizer: fakeSizer{result: 50},
			want:  50,
		}, {
			name:  "InnerExactlyMax",
			min:   50,
			max:   100,
			sizer: fakeSizer{result: 100},
			want:  100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sizer := term.SaturateSizer{
				Min:   tc.min,
				Max:   tc.max,
				Sizer: tc.sizer,
			}

			// Act
			cols := sizer.Columns(&bytes.Buffer{})

			// Assert
			if got, want := cols, tc.want; !cmp.Equal(got, want) {
				t.Errorf("SaturateSizer.Columns(...) got %d, want %d", got, want)
			}
		})
	}
}

func TestFallbackSizer_Columns(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		sizers []int
		want   int
	}{
		{
			name:   "Empty",
			sizers: nil,
			want:   0,
		}, {
			name:   "SingleNonZero",
			sizers: []int{80},
			want:   80,
		}, {
			name:   "SingleZero",
			sizers: []int{0},
			want:   0,
		}, {
			name:   "FirstNonZeroWins",
			sizers: []int{80, 100},
			want:   80,
		}, {
			name:   "FirstZeroFallsThrough",
			sizers: []int{0, 100},
			want:   100,
		}, {
			name:   "AllZero",
			sizers: []int{0, 0, 0},
			want:   0,
		}, {
			name:   "AllZeroThenNonZero",
			sizers: []int{0, 0, 120},
			want:   120,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sizer term.FallbackSizer
			for _, result := range tc.sizers {
				sizer = append(sizer, fakeSizer{result: result})
			}

			// Act
			cols := sizer.Columns(&bytes.Buffer{})

			// Assert
			if got, want := cols, tc.want; !cmp.Equal(got, want) {
				t.Errorf("FallbackSizer.Columns(...) got %d, want %d", got, want)
			}
		})
	}
}
