package diagnostic

import (
	"log/slog"
	"time"
)

// Location represents the location of a diagnostic in the source code.
type Location struct {
	File        string `json:"file,omitempty"`
	LineStart   int    `json:"line_start,omitempty"`
	LineEnd     int    `json:"line_end,omitempty"`
	ColumnStart int    `json:"column_start,omitempty"`
	ColumnEnd   int    `json:"column_end,omitempty"`
}

// Diagnostic represents a diagnostic message, such as an error or warning, that
// can be reported to the user.
type Diagnostic struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`

	Location *Location `json:"location,omitempty"`

	// Err is a sentinel error that can be used to indicate the type of error this
	// diagnostic is associated with.
	Err error `json:"-"`
}

const (
	attrID          = "id"
	attrTitle       = "title"
	attrMessage     = "message"
	attrSeverity    = "severity"
	attrFile        = "file"
	attrLineStart   = "line_start"
	attrLineEnd     = "line_end"
	attrColumnStart = "column_start"
	attrColumnEnd   = "column_end"
	attrSource      = "source"
)

func (d *Diagnostic) toRecord(level slog.Level) slog.Record {
	attrs := []slog.Attr{
		slog.String(attrID, d.ID),
		slog.String(attrTitle, d.Title),
		slog.String(attrMessage, d.Message),
		slog.String(attrSeverity, level.String()),
	}
	attrs = append(attrs, d.locationToAttr(d.Location)...)

	r := slog.NewRecord(
		time.Now(),
		level,
		d.Message,
		0,
	)

	r.AddAttrs(attrs...)
	return r
}

func (d *Diagnostic) fromRecord(r slog.Record) {
	fields := map[string]any{}
	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindGroup {
			for _, inner := range a.Value.Group() {
				fields[inner.Key] = inner.Value.Any()
			}
		} else {
			fields[a.Key] = a.Value.Any()
		}
		return true
	})

	id, _ := fields[attrID].(string)
	title, _ := fields[attrTitle].(string)
	message, _ := fields[attrMessage].(string)

	file, _ := fields[attrFile].(string)
	lineStart, _ := fields[attrLineStart].(int64)
	lineEnd, _ := fields[attrLineEnd].(int64)
	colStart, _ := fields[attrColumnStart].(int64)
	colEnd, _ := fields[attrColumnEnd].(int64)

	d.ID = id
	d.Title = title
	d.Message = message

	if file != "" || lineStart != 0 || lineEnd != 0 || colStart != 0 || colEnd != 0 {
		d.Location = &Location{
			File:        file,
			LineStart:   int(lineStart),
			LineEnd:     int(lineEnd),
			ColumnStart: int(colStart),
			ColumnEnd:   int(colEnd),
		}
	}
}

func (d *Diagnostic) locationToAttr(loc *Location) []slog.Attr {
	if loc == nil {
		return nil
	}
	var attrs []any
	if loc.File != "" {
		attrs = append(attrs, slog.String(attrFile, loc.File))
	}
	if loc.LineStart != 0 {
		attrs = append(attrs, slog.Int(attrLineStart, loc.LineStart))
	}
	if loc.LineEnd != 0 {
		attrs = append(attrs, slog.Int(attrLineEnd, loc.LineEnd))
	}
	if loc.ColumnStart != 0 {
		attrs = append(attrs, slog.Int(attrColumnStart, loc.ColumnStart))
	}
	if loc.ColumnEnd != 0 {
		attrs = append(attrs, slog.Int(attrColumnEnd, loc.ColumnEnd))
	}
	if len(attrs) == 0 {
		return nil
	}
	return []slog.Attr{slog.Group(attrSource, attrs...)}
}
