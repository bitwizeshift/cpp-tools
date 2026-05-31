package linefilter

import (
	"encoding/json"
	"fmt"
)

// LineFilters represents a set of filters that can be applied to lines in files.
type LineFilters []LineFilter

// Includes returns true if the provided file name and line number match any of
// the filters in the set.
func (lfs LineFilters) Includes(fileName string, line int) bool {
	for _, lf := range lfs {
		if lf.Includes(fileName, line) {
			return true
		}
	}
	return false
}

// LineFilter represents a filter that can be applied to lines in a file. It
// matches lines in a file with a specific name, and optionally restricts the
// match to a set of line ranges.
type LineFilter struct {
	Name  string  `json:"name"`
	Lines []Range `json:"lines,omitempty"`
}

// Includes returns true if the provided file name and line number match the
// filter.
// path inclusion is done by exact file-name match; so ensure that paths are
// normalized to the same form as the filter's Name field.
func (lf *LineFilter) Includes(fileName string, line int) bool {
	if lf.Name != fileName {
		return false
	}
	if len(lf.Lines) == 0 {
		return true
	}
	for _, r := range lf.Lines {
		if line >= r.Start && line <= r.End {
			return true
		}
	}
	return false
}

// Range represents a range of line numbers, inclusive.
type Range struct {
	// Start is the first line in the range (inclusive).
	Start int

	// End is the last line in the range (inclusive).
	End int
}

var (
	ErrDomain = fmt.Errorf("linefilter: invalid domain")
	ErrFormat = fmt.Errorf("linefilter: invalid format")
)

// UnmarshalJSON implements [json.Unmarshaler]. It expects the JSON to be an
// array of two integers, representing the start and end of the range,
// respectively.
func (r *Range) UnmarshalJSON(data []byte) error {
	var rng []int
	if err := json.Unmarshal(data, &rng); err != nil {
		return fmt.Errorf("%w: %w", ErrFormat, err)
	}
	if len(rng) != 2 {
		return fmt.Errorf(
			"%w: expected an array of two integers, got %d elements",
			ErrFormat,
			len(rng),
		)
	}
	r.Start = rng[0]
	r.End = rng[1]
	if r.Start < 1 || r.End < 1 {
		return fmt.Errorf(
			"%w: line numbers must be positive; got start=%d, end=%d",
			ErrDomain,
			r.Start,
			r.End,
		)
	}
	if r.Start > r.End {
		return fmt.Errorf(
			"%w: start line number must be less than or equal to end line number; got start=%d, end=%d",
			ErrDomain,
			r.Start,
			r.End,
		)
	}
	return nil
}
