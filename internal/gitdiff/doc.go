// Package gitdiff parses uncolored unified git diff output and reports
// per-file change metadata along with the line-number ranges that were added
// or removed in each hunk.
//
// Use [Parse] for an [io.Reader] or [ParseString] for a string. Both return
// one [File] per "diff --git" entry, in source order. Preamble or trailer
// text outside diff sections is silently skipped, as are diff headers the
// parser does not recognize. Binary patches are reported via [File.Binary]
// with no [Hunk] entries.
//
// Exact hunk line content is intentionally not retained; only structural
// metadata (paths, [Operation], [Mode], and [LineRange] entries on each
// [Hunk]) is exposed.
package gitdiff
