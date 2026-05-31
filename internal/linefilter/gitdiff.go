package linefilter

import "rodusek.dev/pkg/cpp-tools/internal/gitdiff"

// FromDiff returns the [LineFilters] that select the added line ranges of
// each non-binary, non-deleted [gitdiff.File] in files. Each emitted
// [LineFilter] uses the file's NewPath. Files with no added ranges are
// omitted.
func FromDiff(files ...gitdiff.File) LineFilters {
	var filters LineFilters
	for _, f := range files {
		if f.Operation == gitdiff.OperationDelete || f.Binary {
			continue
		}
		lines := addedRanges(f.Hunks)
		if len(lines) == 0 {
			continue
		}
		filters = append(filters, LineFilter{Name: f.NewPath, Lines: lines})
	}
	return filters
}

func addedRanges(hunks []gitdiff.Hunk) []Range {
	var ranges []Range
	for _, h := range hunks {
		for _, r := range h.Added {
			ranges = append(ranges, Range{Start: r.Start, End: r.Start + r.Count - 1})
		}
	}
	return ranges
}
