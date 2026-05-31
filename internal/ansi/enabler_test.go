package ansi_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"rodusek.dev/pkg/cpp-tools/internal/ansi"
)

type fdWriter struct {
	bytes.Buffer
	fd uintptr
}

func (f *fdWriter) Fd() uintptr { return f.fd }

func TestIsTTYFuncEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		hasFd    bool
		fnResult bool
		want     bool
	}{
		{
			name:     "FdWriterFuncReturnsTrue",
			hasFd:    true,
			fnResult: true,
			want:     true,
		}, {
			name:     "FdWriterFuncReturnsFalse",
			hasFd:    true,
			fnResult: false,
			want:     false,
		}, {
			name:     "WriterWithoutFd",
			hasFd:    false,
			fnResult: true,
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var writer io.Writer
			if tc.hasFd {
				writer = &fdWriter{fd: 1}
			} else {
				writer = &bytes.Buffer{}
			}
			enabler := ansi.IsTTYFuncEnabler(func(int) bool { return tc.fnResult })

			// Act
			got := enabler.EnableColour(writer)

			// Assert
			if got, want := got, tc.want; !cmp.Equal(got, want) {
				t.Errorf("IsTTYFuncEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestFixedEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		result bool
		want   bool
	}{
		{
			name:   "True",
			result: true,
			want:   true,
		}, {
			name:   "False",
			result: false,
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := ansi.FixedEnabler(tc.result)

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("FixedEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestEnvEnabler_EnableColour(t *testing.T) {
	const varName = "ANSI_TEST_ENV_ENABLER_VAR"

	testCases := []struct {
		name  string
		set   bool
		value string
		want  bool
	}{
		{
			name:  "NotSet",
			set:   false,
			value: "",
			want:  false,
		}, {
			name:  "SetTrue",
			set:   true,
			value: "true",
			want:  true,
		}, {
			name:  "SetOne",
			set:   true,
			value: "1",
			want:  true,
		}, {
			name:  "SetFalse",
			set:   true,
			value: "false",
			want:  false,
		}, {
			name:  "SetZero",
			set:   true,
			value: "0",
			want:  false,
		}, {
			name:  "SetEmpty",
			set:   true,
			value: "",
			want:  false,
		}, {
			name:  "SetGarbage",
			set:   true,
			value: "garbage",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			if tc.set {
				t.Setenv(varName, tc.value)
			}
			sut := ansi.EnvEnabler{Variable: varName}

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("EnvEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestInvertEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		innerVal bool
		want     bool
	}{
		{
			name:     "InnerTrue",
			innerVal: true,
			want:     false,
		}, {
			name:     "InnerFalse",
			innerVal: false,
			want:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := ansi.InvertEnabler{
				Enabler: ansi.FixedEnabler(tc.innerVal),
			}

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("InvertEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestConjunctiveEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		enablers []ansi.Enabler
		want     bool
	}{
		{
			name:     "Empty",
			enablers: nil,
			want:     false,
		}, {
			name:     "SingleTrue",
			enablers: []ansi.Enabler{ansi.FixedEnabler(true)},
			want:     true,
		}, {
			name:     "SingleFalse",
			enablers: []ansi.Enabler{ansi.FixedEnabler(false)},
			want:     false,
		}, {
			name: "MultipleAllTrue",
			enablers: []ansi.Enabler{
				ansi.FixedEnabler(true),
				ansi.FixedEnabler(true),
				ansi.FixedEnabler(true),
			},
			want: true,
		}, {
			name: "MultipleFirstFalse",
			enablers: []ansi.Enabler{
				ansi.FixedEnabler(false),
				ansi.FixedEnabler(true),
				ansi.FixedEnabler(true),
			},
			want: false,
		}, {
			name: "MultipleLastFalse",
			enablers: []ansi.Enabler{
				ansi.FixedEnabler(true),
				ansi.FixedEnabler(true),
				ansi.FixedEnabler(false),
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := ansi.ConjunctiveEnabler(tc.enablers)

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("ConjunctiveEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}
