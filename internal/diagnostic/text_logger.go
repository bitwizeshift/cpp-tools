package diagnostic

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"rodusek.dev/pkg/cpp-tools/internal/ansi"
	"rodusek.dev/pkg/cpp-tools/internal/format"
	"rodusek.dev/pkg/cpp-tools/internal/term"
)

// TextHandler renders Diagnostics in a rustc-style human-readable form,
// optionally with ANSI colour.
type TextHandler struct {
	w           io.Writer
	enableColour bool
	columns     int
}

// NewTextLogger returns a Logger that writes to w using TextHandler with the
// package-default colour policy ([ansi.DefaultEnabler]).
func NewTextLogger(w io.Writer) *Logger {
	return NewLogger(NewTextHandler(w, ansi.DefaultEnabler, term.DefaultSizer))
}

// NewTextHandler returns a TextHandler that writes to w. Colour is enabled
// when enabler returns true for w.
func NewTextHandler(w io.Writer, enabler ansi.Enabler, sizer term.Sizer) *TextHandler {
	return &TextHandler{
		w:           w,
		enableColour: enabler.EnableColour(w),
		columns:     sizer.Columns(w),
	}
}

// Enabled implements slog.Handler.
func (h *TextHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (h *TextHandler) Handle(_ context.Context, r slog.Record) error {
	var diagnostic Diagnostic
	diagnostic.fromRecord(r)

	severity := strings.ToLower(r.Level.String())
	var location string
	if diagnostic.Location != nil {
		location = h.formatLocation(
			diagnostic.Location.File,
			diagnostic.Location.LineStart,
			diagnostic.Location.LineEnd,
			diagnostic.Location.ColumnStart,
			diagnostic.Location.ColumnEnd,
		)
	}
	id := diagnostic.ID
	title := diagnostic.Title
	message := diagnostic.Message

	severityFmt := h.formatterFor(severityColour(r.Level))
	accentFmt := h.formatterFor(ansi.BrightBlue)
	idFmt := h.formatterFor(ansi.BrightWhite)

	/*
	   Example output:
	   error[E0499]: cannot borrow `x` as mutable more than once
	     --> src/main.rs:10:5-10
	*/
	headline := h.firstNonEmpty(title, message)
	if id != "" {
		header := severityFmt.Format("%s[", severity) + idFmt.Format("%s", id) + severityFmt.Format("]")
		_, _ = fmt.Fprintf(h.w, "%s: %s\n", header, headline)
	} else {
		_, _ = fmt.Fprintf(h.w, "%s: %s\n", severityFmt.Format("%s", severity), headline)
	}
	if location != "" {
		_, _ = fmt.Fprintf(
			h.w,
			"  %s %s\n",
			accentFmt.Format("-->"),
			ansi.Underline.Format("%s", location),
		)
	}
	if title != "" && message != "" && title != message {
		_, _ = fmt.Fprintf(h.w, "   %s\n", accentFmt.Format("|"))
		text := format.Resize(message, h.columns-5)
		for line := range strings.SplitSeq(text, "\n") {
			_, _ = fmt.Fprintf(h.w, "   %s %s\n", accentFmt.Format("="), line)
		}
	}
	_, _ = fmt.Fprintln(h.w)
	return nil
}

func (h *TextHandler) formatterFor(c ansi.Colour) formatter {
	var f formatter = c
	if !h.enableColour {
		f = noFormatter{}
	}
	return f
}

func severityColour(level slog.Level) ansi.Colour {
	switch level {
	case slog.LevelError:
		return ansi.BrightRed
	case slog.LevelWarn:
		return ansi.BrightYellow
	case slog.LevelInfo:
		return ansi.BrightGreen
	default:
		return ansi.BrightCyan
	}
}

// WithAttrs implements slog.Handler. Attrs are ignored; Diagnostics carry
// their own attributes.
func (h *TextHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

// WithGroup implements slog.Handler. Groups are ignored; Diagnostics carry
// their own attributes.
func (h *TextHandler) WithGroup(string) slog.Handler {
	return h
}

var _ slog.Handler = (*TextHandler)(nil)

func (h *TextHandler) formatLocation(file string, lineStart int, lineEnd int, colStart int, colEnd int) string {
	var b strings.Builder
	if file != "" {
		b.WriteString(file)
	}
	if lineStart > 0 {
		b.WriteString(":")
		fmt.Fprintf(&b, "%d", lineStart)
		if lineEnd > 0 && lineEnd != lineStart {
			b.WriteString("-")
			fmt.Fprintf(&b, "%d", lineEnd)
		}
		if colStart > 0 {
			b.WriteString(":")
			fmt.Fprintf(&b, "%d", colStart)
			if colEnd > 0 && colEnd != colStart {
				b.WriteString("-")
				fmt.Fprintf(&b, "%d", colEnd)
			}
		}
	}
	return b.String()
}
func (h *TextHandler) firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

type formatter interface {
	Format(format string, args ...any) string
}

type noFormatter struct{}

func (f noFormatter) Format(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
