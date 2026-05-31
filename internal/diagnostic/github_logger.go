package diagnostic

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// GitHubHandler renders Diagnostics as GitHub Actions workflow commands so
// the runner surfaces them as annotations on the job log and PR review.
type GitHubHandler struct {
	w io.Writer
}

// NewGitHubLogger returns a Logger that writes GitHub workflow commands to w.
func NewGitHubLogger(w io.Writer) *Logger {
	return NewLogger(NewGitHubHandler(w))
}

// NewGitHubHandler returns a GitHubHandler that writes to w.
func NewGitHubHandler(w io.Writer) *GitHubHandler {
	return &GitHubHandler{w: w}
}

// Enabled implements [slog.Handler].
func (h *GitHubHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *GitHubHandler) toSeverityString(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelWarn:
		return "warning"
	case slog.LevelError:
		return "error"
	default:
		return "info"
	}
}

// Handle implements [slog.Handler].
func (h *GitHubHandler) Handle(_ context.Context, r slog.Record) error {
	var diagnostic Diagnostic
	diagnostic.fromRecord(r)

	severity := h.toSeverityString(r.Level)
	id := diagnostic.ID
	title := diagnostic.Title
	message := diagnostic.Message

	var props []string
	if title != "" {
		titleValue := title
		if id != "" {
			titleValue = fmt.Sprintf("[%s] %s", id, title)
		}
		props = append(props, "title="+githubEscapeProperty(titleValue))
	}
	if diagnostic.Location != nil {
		file := diagnostic.Location.File
		lineStart := diagnostic.Location.LineStart
		lineEnd := diagnostic.Location.LineEnd
		colStart := diagnostic.Location.ColumnStart
		colEnd := diagnostic.Location.ColumnEnd
		if file != "" {
			props = append(props, "file="+githubEscapeProperty(file))
		}
		if lineStart != 0 {
			props = append(props, fmt.Sprintf("line=%d", lineStart))
		}
		if lineEnd != 0 {
			props = append(props, fmt.Sprintf("endLine=%d", lineEnd))
		}
		if colStart != 0 {
			props = append(props, fmt.Sprintf("col=%d", colStart))
		}
		if colEnd != 0 {
			props = append(props, fmt.Sprintf("endColumn=%d", colEnd))
		}
	}

	var sb strings.Builder
	sb.WriteString("::")
	sb.WriteString(severity)
	if len(props) > 0 {
		sb.WriteByte(' ')
		sb.WriteString(strings.Join(props, ","))
	}
	sb.WriteString("::")
	sb.WriteString(githubEscapeData(message))
	sb.WriteByte('\n')

	_, err := h.w.Write([]byte(sb.String()))
	return err
}

// WithGroup implements slog.Handler. Groups are ignored; Diagnostics carry
// their own attributes.
func (h *GitHubHandler) WithGroup(string) slog.Handler {
	return h
}

// WithAttrs implements slog.Handler. Attrs are ignored; Diagnostics carry
// their own attributes.
func (h *GitHubHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

var _ slog.Handler = (*GitHubHandler)(nil)

// GitHub workflow commands use percent-encoded escapes for control characters
// and (in property values) for the delimiters `:` and `,`. See the encoding
// applied by the official actions/toolkit core library:
// https://github.com/actions/toolkit/blob/main/packages/core/src/command.ts
var (
	githubDataReplacer = strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
	)
	githubPropertyReplacer = strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
		":", "%3A",
		",", "%2C",
	)
)

func githubEscapeData(s string) string {
	return githubDataReplacer.Replace(s)
}

func githubEscapeProperty(s string) string {
	return githubPropertyReplacer.Replace(s)
}
