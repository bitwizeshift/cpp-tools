package ansi

import (
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

// Enabler decides whether colour should be emitted to a particular writer.
type Enabler interface {
	EnableColour(w io.Writer) bool
}

// IsTTYFuncEnabler adapts an "is this fd a TTY?" predicate (such as
// [golang.org/x/term.IsTerminal]) into an Enabler. It enables colour only
// when the writer exposes an Fd() method and the predicate returns true.
type IsTTYFuncEnabler func(fd int) bool

// EnableColour implements [Enabler].
func (f IsTTYFuncEnabler) EnableColour(w io.Writer) bool {
	if tty, ok := w.(interface{ Fd() uintptr }); ok {
		return f(int(tty.Fd()))
	}
	return false
}

var _ Enabler = (*IsTTYFuncEnabler)(nil)

// FixedEnabler is an Enabler whose decision is hard-coded. Useful when the
// caller has already resolved colour from configuration or a user flag.
type FixedEnabler bool

// EnableColour implements [Enabler].
func (e FixedEnabler) EnableColour(io.Writer) bool {
	return bool(e)
}

var _ Enabler = (*FixedEnabler)(nil)

// EnvEnabler enables colour iff the named environment variable parses as a
// truthy boolean (see [strconv.ParseBool]).
type EnvEnabler struct {
	Variable string
}

// EnableColour implements [Enabler].
func (e EnvEnabler) EnableColour(w io.Writer) bool {
	val, set := os.LookupEnv(e.Variable)
	enabled, _ := strconv.ParseBool(val)
	return set && enabled
}

var _ Enabler = (*EnvEnabler)(nil)

// InvertEnabler negates an inner Enabler's decision. Use it to turn an
// "opt-out" signal (such as NO_COLOUR) into the equivalent enable check.
type InvertEnabler struct {
	Enabler Enabler
}

// EnableColour implements [Enabler].
func (e InvertEnabler) EnableColour(w io.Writer) bool {
	return !e.Enabler.EnableColour(w)
}

var _ Enabler = (*InvertEnabler)(nil)

// ConjunctiveEnabler enables colour only when every member agrees. An empty
// value disables colour.
type ConjunctiveEnabler []Enabler

// EnableColour implements [Enabler].
func (c ConjunctiveEnabler) EnableColour(w io.Writer) bool {
	if len(c) == 0 {
		return false
	}
	for _, checker := range c {
		if !checker.EnableColour(w) {
			return false
		}
	}
	return true
}

var _ Enabler = (*ConjunctiveEnabler)(nil)

// DefaultEnabler is the standard policy: colour is enabled only when the
// writer is a real terminal and the NO_COLOUR environment variable is not
// set to a truthy value (see https://no-colour.org/).
var DefaultEnabler Enabler = ConjunctiveEnabler{
	IsTTYFuncEnabler(term.IsTerminal),
	InvertEnabler{
		Enabler: EnvEnabler{
			Variable: "NO_COLOUR",
		},
	},
}
