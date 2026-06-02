package spdx

import (
	"encoding"
	"fmt"
	"strings"
)

// Expression is a parsed SPDX license expression in disjunctive normal form:
// a set of alternative [Choice]s. Satisfying any single [Choice] satisfies the
// expression as a whole.
type Expression struct {
	// Choices enumerates the alternative satisfying license sets.
	Choices []Choice
}

// String renders the expression as SPDX syntax, joining [Choice]s with OR.
func (e Expression) String() string {
	parts := make([]string, len(e.Choices))
	for i, c := range e.Choices {
		parts[i] = c.String()
	}
	return strings.Join(parts, " OR ")
}

// UnmarshalText implements [encoding.TextUnmarshaler] by parsing text as an SPDX
// license expression.
func (e *Expression) UnmarshalText(text []byte) error {
	expr, err := Parse(string(text))
	if err != nil {
		return err
	}
	*e = *expr
	return nil
}

var (
	_ encoding.TextUnmarshaler = (*Expression)(nil)
	_ fmt.Stringer             = Expression{}
)

// Choice is a set of licenses that must all be honored together for the
// choice to be satisfied.
type Choice struct {
	// Licenses are the [License]s that must all be honored.
	Licenses []License
}

// String renders the choice as SPDX syntax, joining [License]s with AND.
func (c Choice) String() string {
	parts := make([]string, len(c.Licenses))
	for i, l := range c.Licenses {
		parts[i] = l.String()
	}
	return strings.Join(parts, " AND ")
}

// License names a single SPDX license, optionally with an "or-later" suffix
// and an SPDX exception applied via WITH.
type License struct {
	// ID is the SPDX license identifier, e.g. "MIT" or "Apache-2.0".
	ID string

	// OrLater reports whether the source expression used the "+" suffix.
	OrLater bool

	// Exception is the SPDX exception identifier from a WITH clause; empty
	// when the license had no WITH clause.
	Exception string
}

// String renders the license as SPDX syntax, including any "+" suffix and
// WITH exception clause.
func (l License) String() string {
	var b strings.Builder
	b.WriteString(l.ID)
	if l.OrLater {
		b.WriteByte('+')
	}
	if l.Exception != "" {
		b.WriteString(" WITH ")
		b.WriteString(l.Exception)
	}
	return b.String()
}
