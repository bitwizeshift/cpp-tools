package gitdiff

import (
	"bufio"
	"encoding"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Sentinel errors returned by parsing operations.
var (
	// ErrInvalidFileHeader indicates that a "diff --git" header could not be
	// parsed into a pair of paths.
	ErrInvalidFileHeader = errors.New("invalid file header")

	// ErrInvalidHunkHeader indicates that a "@@ ... @@" header could not be
	// parsed into old and new line ranges.
	ErrInvalidHunkHeader = errors.New("invalid hunk header")

	// ErrInvalidMode indicates that a mode token was not a valid octal value.
	ErrInvalidMode = errors.New("invalid file mode")

	// ErrInvalidOperation indicates that an operation token was not one of the
	// defined [Operation] values.
	ErrInvalidOperation = errors.New("invalid operation")

	// ErrTruncatedHunk indicates that a hunk body ended before its declared
	// line counts were satisfied.
	ErrTruncatedHunk = errors.New("truncated hunk")

	// ErrUnexpectedLine indicates that a line within a hunk body did not begin
	// with one of ' ', '+', '-', or '\\'.
	ErrUnexpectedLine = errors.New("unexpected line in hunk body")
)

// Mode is a unix file mode token as it appears in a git diff header
// (for example 100644 or 100755).
type Mode uint32

// Common mode values seen in git diffs.
const (
	ModeFile       Mode = 0o100644
	ModeExecutable Mode = 0o100755
	ModeSymlink    Mode = 0o120000
	ModeSubmodule  Mode = 0o160000
)

// UnmarshalText parses an octal mode token (for example "100644") into m.
// Returns [ErrInvalidMode] wrapped with the offending text on failure.
func (m *Mode) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 8, 32)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidMode, text)
	}
	*m = Mode(v)
	return nil
}

var _ encoding.TextUnmarshaler = (*Mode)(nil)

// Operation classifies the change applied to a file in a diff.
type Operation string

// Operation values produced by [Parse].
const (
	OperationModify Operation = "modify"
	OperationCreate Operation = "create"
	OperationDelete Operation = "delete"
	OperationRename Operation = "rename"
	OperationCopy   Operation = "copy"
)

// UnmarshalText parses one of the defined operation tokens into op.
// Returns [ErrInvalidOperation] wrapped with the offending text on failure.
func (op *Operation) UnmarshalText(text []byte) error {
	switch s := Operation(text); s {
	case OperationModify, OperationCreate, OperationDelete, OperationRename, OperationCopy:
		*op = s
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidOperation, text)
	}
}

var _ encoding.TextUnmarshaler = (*Operation)(nil)

// LineRange is a contiguous run of line numbers.
// Start is 1-indexed and Count is at least 1.
type LineRange struct {
	Start int
	Count int
}

// Hunk holds the line-range deltas of a single "@@ ... @@" section.
// Added ranges reference line numbers in the new file; Removed ranges
// reference line numbers in the old file.
type Hunk struct {
	Added   []LineRange
	Removed []LineRange
}

// File holds the parsed metadata and hunks for one "diff --git" entry.
// OldMode and NewMode are nil when the diff does not specify them.
// Binary reports whether the diff is a binary patch, in which case the
// content is skipped and no Hunks are produced.
type File struct {
	OldPath   string
	NewPath   string
	Operation Operation
	OldMode   *Mode
	NewMode   *Mode
	Binary    bool
	Hunks     []Hunk
}

// Parse reads an uncolored unified git diff from r and returns one [File]
// per "diff --git" entry, in source order. Preamble or trailer text outside
// diff sections is silently skipped.
func Parse(r io.Reader) ([]File, error) {
	return newParser(r).parse()
}

// ParseString is equivalent to [Parse] over s.
func ParseString(s string) ([]File, error) {
	return Parse(strings.NewReader(s))
}

const scannerMaxLine = 1 << 20

type parser struct {
	scanner *bufio.Scanner
	peeked  string
	hasPeek bool
	files   []File
}

func newParser(r io.Reader) *parser {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), scannerMaxLine)
	return &parser{scanner: s}
}

func (p *parser) next() (string, bool) {
	if p.hasPeek {
		line := p.peeked
		p.peeked = ""
		p.hasPeek = false
		return line, true
	}
	if !p.scanner.Scan() {
		return "", false
	}
	return p.scanner.Text(), true
}

func (p *parser) pushBack(line string) {
	p.peeked = line
	p.hasPeek = true
}

func (p *parser) parse() ([]File, error) {
	for {
		line, ok := p.next()
		if !ok {
			break
		}
		if !strings.HasPrefix(line, "diff --git ") {
			continue
		}
		if err := p.parseFile(line); err != nil {
			return nil, err
		}
	}
	if err := p.scanner.Err(); err != nil {
		return nil, err
	}
	return p.files, nil
}

func (p *parser) parseFile(headerLine string) error {
	f := File{}
	if err := assignHeaderPaths(&f, headerLine); err != nil {
		return err
	}
	for {
		line, ok := p.next()
		if !ok {
			break
		}
		if strings.HasPrefix(line, "diff --git ") {
			p.pushBack(line)
			break
		}
		if err := p.handleFileLine(&f, line); err != nil {
			return err
		}
		if f.Binary {
			break
		}
	}
	if f.Operation == "" {
		f.Operation = OperationModify
	}
	p.files = append(p.files, f)
	return nil
}

func assignHeaderPaths(f *File, line string) error {
	rest := line[len("diff --git "):]
	oldPart, newPath, ok := strings.Cut(rest, " b/")
	if !ok || !strings.HasPrefix(oldPart, "a/") {
		return fmt.Errorf("%w: %q", ErrInvalidFileHeader, line)
	}
	f.OldPath = oldPart[len("a/"):]
	f.NewPath = newPath
	if f.OldPath == "" || f.NewPath == "" {
		return fmt.Errorf("%w: %q", ErrInvalidFileHeader, line)
	}
	return nil
}

func (p *parser) handleFileLine(f *File, line string) error {
	switch {
	case strings.HasPrefix(line, "@@"):
		return p.parseHunk(f, line)
	case isBinaryMarker(line):
		f.Binary = true
		p.skipBinaryBody()
		return nil
	case strings.HasPrefix(line, "old mode "):
		return assignMode(&f.OldMode, line[len("old mode "):])
	case strings.HasPrefix(line, "new mode "):
		return assignMode(&f.NewMode, line[len("new mode "):])
	case strings.HasPrefix(line, "new file mode "):
		if err := assignMode(&f.NewMode, line[len("new file mode "):]); err != nil {
			return err
		}
		f.Operation = OperationCreate
		return nil
	case strings.HasPrefix(line, "deleted file mode "):
		if err := assignMode(&f.OldMode, line[len("deleted file mode "):]); err != nil {
			return err
		}
		f.Operation = OperationDelete
		return nil
	case strings.HasPrefix(line, "rename from "):
		f.Operation = OperationRename
		f.OldPath = line[len("rename from "):]
		return nil
	case strings.HasPrefix(line, "rename to "):
		f.Operation = OperationRename
		f.NewPath = line[len("rename to "):]
		return nil
	case strings.HasPrefix(line, "copy from "):
		f.Operation = OperationCopy
		f.OldPath = line[len("copy from "):]
		return nil
	case strings.HasPrefix(line, "copy to "):
		f.Operation = OperationCopy
		f.NewPath = line[len("copy to "):]
		return nil
	case strings.HasPrefix(line, "index "):
		applyIndexMode(f, line)
		return nil
	default:
		return nil
	}
}

func isBinaryMarker(line string) bool {
	if strings.HasPrefix(line, "Binary files ") && strings.HasSuffix(line, " differ") {
		return true
	}
	return line == "GIT binary patch"
}

func applyIndexMode(f *File, line string) {
	rest := line[len("index "):]
	sp := strings.LastIndexByte(rest, ' ')
	if sp < 0 {
		return
	}
	var m Mode
	if err := m.UnmarshalText([]byte(rest[sp+1:])); err != nil {
		return
	}
	if f.OldMode == nil {
		old := m
		f.OldMode = &old
	}
	if f.NewMode == nil {
		nw := m
		f.NewMode = &nw
	}
}

func assignMode(dst **Mode, token string) error {
	var m Mode
	if err := m.UnmarshalText([]byte(token)); err != nil {
		return err
	}
	*dst = &m
	return nil
}

func (p *parser) skipBinaryBody() {
	for {
		line, ok := p.next()
		if !ok {
			return
		}
		if strings.HasPrefix(line, "diff --git ") {
			p.pushBack(line)
			return
		}
	}
}

func (p *parser) parseHunk(f *File, header string) error {
	oldStart, oldCount, newStart, newCount, err := parseHunkHeader(header)
	if err != nil {
		return err
	}
	body := hunkBody{
		oldRem:  oldCount,
		newRem:  newCount,
		oldLine: oldStart,
		newLine: newStart,
	}
	for body.oldRem > 0 || body.newRem > 0 {
		line, ok := p.next()
		if !ok {
			return fmt.Errorf("%w: %q", ErrTruncatedHunk, header)
		}
		if err := body.consume(line); err != nil {
			return err
		}
	}
	body.flush()
	p.consumeTrailingNoNewline()
	f.Hunks = append(f.Hunks, body.hunk)
	return nil
}

func (p *parser) consumeTrailingNoNewline() {
	line, ok := p.next()
	if !ok {
		return
	}
	if !strings.HasPrefix(line, `\`) {
		p.pushBack(line)
	}
}

type hunkBody struct {
	hunk      Hunk
	oldRem    int
	newRem    int
	oldLine   int
	newLine   int
	addedOpen LineRange
	addedSet  bool
	rmOpen    LineRange
	rmSet     bool
}

func (b *hunkBody) consume(line string) error {
	var lead byte
	if len(line) > 0 {
		lead = line[0]
	}
	switch lead {
	case 0, ' ':
		b.flush()
		b.oldRem--
		b.newRem--
		b.oldLine++
		b.newLine++
	case '+':
		b.flushRemoved()
		if b.addedSet {
			b.addedOpen.Count++
		} else {
			b.addedOpen = LineRange{Start: b.newLine, Count: 1}
			b.addedSet = true
		}
		b.newRem--
		b.newLine++
	case '-':
		b.flushAdded()
		if b.rmSet {
			b.rmOpen.Count++
		} else {
			b.rmOpen = LineRange{Start: b.oldLine, Count: 1}
			b.rmSet = true
		}
		b.oldRem--
		b.oldLine++
	case '\\':
		// "\ No newline at end of file" — no line-count change.
	default:
		return fmt.Errorf("%w: %q", ErrUnexpectedLine, line)
	}
	return nil
}

func (b *hunkBody) flush() {
	b.flushAdded()
	b.flushRemoved()
}

func (b *hunkBody) flushAdded() {
	if b.addedSet {
		b.hunk.Added = append(b.hunk.Added, b.addedOpen)
		b.addedSet = false
	}
}

func (b *hunkBody) flushRemoved() {
	if b.rmSet {
		b.hunk.Removed = append(b.hunk.Removed, b.rmOpen)
		b.rmSet = false
	}
}

func parseHunkHeader(line string) (oldStart, oldCount, newStart, newCount int, err error) {
	rest, ok := strings.CutPrefix(line, "@@ ")
	if !ok {
		return 0, 0, 0, 0, fmt.Errorf("%w: %q", ErrInvalidHunkHeader, line)
	}
	before, _, ok0 := strings.Cut(rest, " @@")
	if !ok0 {
		return 0, 0, 0, 0, fmt.Errorf("%w: %q", ErrInvalidHunkHeader, line)
	}
	parts := strings.SplitN(before, " ", 2)
	if len(parts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("%w: %q", ErrInvalidHunkHeader, line)
	}
	oldStart, oldCount, ok1 := parseRangeSpec(parts[0], '-')
	newStart, newCount, ok2 := parseRangeSpec(parts[1], '+')
	if !ok1 || !ok2 {
		return 0, 0, 0, 0, fmt.Errorf("%w: %q", ErrInvalidHunkHeader, line)
	}
	return oldStart, oldCount, newStart, newCount, nil
}

func parseRangeSpec(s string, prefix byte) (start, count int, ok bool) {
	if len(s) == 0 || s[0] != prefix {
		return 0, 0, false
	}
	body := s[1:]
	if before, after, ok0 := strings.Cut(body, ","); ok0 {
		startVal, err1 := strconv.Atoi(before)
		countVal, err2 := strconv.Atoi(after)
		if err1 != nil || err2 != nil {
			return 0, 0, false
		}
		return startVal, countVal, true
	}
	startVal, err := strconv.Atoi(body)
	if err != nil {
		return 0, 0, false
	}
	return startVal, 1, true
}
