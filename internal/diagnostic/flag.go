package diagnostic

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	"rodusek.dev/pkg/cpp-tools/internal/ansi"
	"rodusek.dev/pkg/cpp-tools/internal/term"
)

// FormatType selects one of the diagnostic output renderings.
type FormatType string

// Recognised values for FormatType.
const (
	FormatANSI   FormatType = "ansi"
	FormatText   FormatType = "text"
	FormatGitHub FormatType = "github"
	FormatJSON   FormatType = "json"
)

// UnmarshalText implements encoding.TextUnmarshaler.
func (f *FormatType) UnmarshalText(text []byte) error {
	switch FormatType(text) {
	case FormatANSI, FormatText, FormatGitHub, FormatJSON:
		*f = FormatType(text)
		return nil
	default:
		return fmt.Errorf("invalid format type: %s", text)
	}
}

// LoggerFlag represents a flag that can be used to control the behavior of a
// logger.
type LoggerFlag struct {
	format  FormatType
	noColour bool
}

type formatValue FormatType

func (f *formatValue) String() string {
	return string(*f)
}

func (f *formatValue) Set(s string) error {
	return (*FormatType)(f).UnmarshalText([]byte(s))
}

func (f *formatValue) Type() string {
	return "format"
}

var _ pflag.Value = (*formatValue)(nil)

// RegisterFlags registers the LoggerFlag's flags with the provided FlagSet.
func (lf *LoggerFlag) RegisterFlags(fs *pflag.FlagSet) {
	fs.Var((*formatValue)(&lf.format),
		"output-format",
		"The format to use for diagnostics (text, ansi, json, github)",
	)
	fs.BoolVar(&lf.noColour, "no-colour", false, "Disable colour output in diagnostics")
}

// ColourEnabled returns true if the logger should produce colour output.
func (lf *LoggerFlag) ColourEnabled() bool {
	return !lf.noColour
}

// Logger returns a new Logger based on the flag's configuration.
func (lf *LoggerFlag) Logger(w io.Writer) *Logger {
	switch lf.format {
	case FormatJSON:
		return NewJSONLogger(w)
	case FormatText:
		return NewLogger(NewTextHandler(w, ansi.FixedEnabler(false), term.DefaultSizer))
	case FormatANSI:
		return NewLogger(NewTextHandler(w, ansi.FixedEnabler(lf.ColourEnabled()), term.DefaultSizer))
	case FormatGitHub:
		return NewGitHubLogger(w)
	default:
		if !lf.ColourEnabled() {
			return NewLogger(NewTextHandler(w, ansi.FixedEnabler(false), term.DefaultSizer))
		}
		return NewTextLogger(w)
	}
}
