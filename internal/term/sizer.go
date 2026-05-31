package term

import (
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

// Sizer reports the column width to use when rendering text for the writer
// w. A return value of 0 means "unknown" and is a signal that callers (or
// composing Sizers) should fall back to another source.
type Sizer interface {
	Columns(w io.Writer) int
}

// EnvSizer reads the column width from a named environment variable. A
// missing or non-numeric value yields 0.
type EnvSizer struct {
	Variable string
}

// Columns implements [Sizer].
func (s EnvSizer) Columns(io.Writer) int {
	val := os.Getenv(s.Variable)
	cols, _ := strconv.Atoi(val) // returns 0 on error
	return cols
}

var _ Sizer = (*EnvSizer)(nil)

// FixedSizer is a Sizer that always returns its integer value as the column
// width.
type FixedSizer int

// Columns implements [Sizer].
func (s FixedSizer) Columns(w io.Writer) int {
	return int(s)
}

var _ Sizer = (*FixedSizer)(nil)

// TTYFuncSizer adapts a "size of this fd" function (such as
// [golang.org/x/term.GetSize]) into a Sizer. It returns 0 for writers without
// an Fd() or when the underlying call fails.
type TTYFuncSizer func(fd int) (cols int, rows int, err error)

// Columns implements [Sizer].
func (f TTYFuncSizer) Columns(w io.Writer) int {
	if tty, ok := w.(interface{ Fd() uintptr }); ok {
		cols, _, err := f(int(tty.Fd()))
		if err == nil {
			return cols
		}
	}
	return 0
}

var _ Sizer = (*TTYFuncSizer)(nil)

// SaturateSizer clamps a non-zero reading from its inner Sizer to the
// inclusive range [Min, Max]. A zero reading is passed through unchanged so
// composing Sizers can still treat it as "unknown".
type SaturateSizer struct {
	Min, Max int
	Sizer    Sizer
}

// Columns implements [Sizer].
func (s SaturateSizer) Columns(w io.Writer) int {
	if cols := s.Sizer.Columns(w); cols > 0 {
		if cols < s.Min {
			return s.Min
		} else if cols > s.Max {
			return s.Max
		} else {
			return cols
		}
	}
	return 0
}

// FallbackSizer returns the first non-zero reading from its members.
type FallbackSizer []Sizer

// Columns implements [Sizer].
func (f FallbackSizer) Columns(w io.Writer) int {
	for _, sizer := range f {
		if cols := sizer.Columns(w); cols > 0 {
			return cols
		}
	}
	return 0
}

// DefaultSizer is the standard policy: prefer the COLUMNS environment
// variable, then the actual terminal width, then a final 80-column fallback,
// clamping the result to the inclusive range [50, 100].
var DefaultSizer = SaturateSizer{
	Min: 50,
	Max: 100,
	Sizer: FallbackSizer{
		EnvSizer{Variable: "COLUMNS"},
		TTYFuncSizer(term.GetSize),
		FixedSizer(80),
	},
}
