// Package ansi provides ANSI SGR escape codes for foreground colours and
// basic text styling.
package ansi

import "fmt"

// Colour is an ANSI SGR escape sequence applied as a prefix and paired with
// [Reset] when wrapping text.
type Colour string

// Foreground colour escape sequences. Reset clears any active SGR state.
const (
	Reset Colour = "\x1b[0m"

	Black   Colour = "\x1b[30m"
	Red     Colour = "\x1b[31m"
	Green   Colour = "\x1b[32m"
	Yellow  Colour = "\x1b[33m"
	Blue    Colour = "\x1b[34m"
	Magenta Colour = "\x1b[35m"
	Cyan    Colour = "\x1b[36m"
	White   Colour = "\x1b[37m"

	BrightBlack   Colour = "\x1b[90m"
	BrightRed     Colour = "\x1b[91m"
	BrightGreen   Colour = "\x1b[92m"
	BrightYellow  Colour = "\x1b[93m"
	BrightBlue    Colour = "\x1b[94m"
	BrightMagenta Colour = "\x1b[95m"
	BrightCyan    Colour = "\x1b[96m"
	BrightWhite   Colour = "\x1b[97m"
)

// Text style escape sequences.
const (
	Underline Colour = "\x1b[4m"
)

// Format formats the given string with the ANSI escape code for this colour,
// and resets the colour at the end.
func (c Colour) Format(format string, args ...any) string {
	return string(c) + fmt.Sprintf(format, args...) + string(Reset)
}
