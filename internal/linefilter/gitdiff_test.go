package linefilter_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/gitdiff"
	"rodusek.dev/pkg/cpp-tools/internal/linefilter"
)

func TestFromDiff(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		files []gitdiff.File
		want  linefilter.LineFilters
	}{
		{
			name:  "no files",
			files: nil,
			want:  nil,
		},
		{
			name: "single modify with added range",
			files: []gitdiff.File{
				{
					OldPath:   "foo.txt",
					NewPath:   "foo.txt",
					Operation: gitdiff.OperationModify,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 2, Count: 3}}},
					},
				},
			},
			want: linefilter.LineFilters{
				{
					Name:  "foo.txt",
					Lines: []linefilter.Range{{Start: 2, End: 4}},
				},
			},
		},
		{
			name: "create with multi-line added range",
			files: []gitdiff.File{
				{
					OldPath:   "new.txt",
					NewPath:   "new.txt",
					Operation: gitdiff.OperationCreate,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 1, Count: 5}}},
					},
				},
			},
			want: linefilter.LineFilters{
				{
					Name:  "new.txt",
					Lines: []linefilter.Range{{Start: 1, End: 5}},
				},
			},
		},
		{
			name: "single-line added range",
			files: []gitdiff.File{
				{
					OldPath:   "foo.txt",
					NewPath:   "foo.txt",
					Operation: gitdiff.OperationModify,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 7, Count: 1}}},
					},
				},
			},
			want: linefilter.LineFilters{
				{
					Name:  "foo.txt",
					Lines: []linefilter.Range{{Start: 7, End: 7}},
				},
			},
		},
		{
			name: "rename with content uses NewPath",
			files: []gitdiff.File{
				{
					OldPath:   "old.txt",
					NewPath:   "new.txt",
					Operation: gitdiff.OperationRename,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 5, Count: 1}}},
					},
				},
			},
			want: linefilter.LineFilters{
				{
					Name:  "new.txt",
					Lines: []linefilter.Range{{Start: 5, End: 5}},
				},
			},
		},
		{
			name: "delete skipped",
			files: []gitdiff.File{
				{
					OldPath:   "del.txt",
					NewPath:   "del.txt",
					Operation: gitdiff.OperationDelete,
					Hunks: []gitdiff.Hunk{
						{Removed: []gitdiff.LineRange{{Start: 1, Count: 2}}},
					},
				},
			},
			want: nil,
		},
		{
			name: "binary skipped",
			files: []gitdiff.File{
				{
					OldPath:   "img.png",
					NewPath:   "img.png",
					Operation: gitdiff.OperationModify,
					Binary:    true,
				},
			},
			want: nil,
		},
		{
			name: "rename without content skipped",
			files: []gitdiff.File{
				{
					OldPath:   "old.txt",
					NewPath:   "new.txt",
					Operation: gitdiff.OperationRename,
				},
			},
			want: nil,
		},
		{
			name: "modify with only removals skipped",
			files: []gitdiff.File{
				{
					OldPath:   "foo.txt",
					NewPath:   "foo.txt",
					Operation: gitdiff.OperationModify,
					Hunks: []gitdiff.Hunk{
						{Removed: []gitdiff.LineRange{{Start: 1, Count: 2}}},
					},
				},
			},
			want: nil,
		},
		{
			name: "multiple files",
			files: []gitdiff.File{
				{
					OldPath:   "a.txt",
					NewPath:   "a.txt",
					Operation: gitdiff.OperationModify,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 1, Count: 2}}},
					},
				},
				{
					OldPath:   "b.txt",
					NewPath:   "b.txt",
					Operation: gitdiff.OperationCreate,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 1, Count: 3}}},
					},
				},
			},
			want: linefilter.LineFilters{
				{
					Name:  "a.txt",
					Lines: []linefilter.Range{{Start: 1, End: 2}},
				},
				{
					Name:  "b.txt",
					Lines: []linefilter.Range{{Start: 1, End: 3}},
				},
			},
		},
		{
			name: "multiple hunks aggregated",
			files: []gitdiff.File{
				{
					OldPath:   "foo.txt",
					NewPath:   "foo.txt",
					Operation: gitdiff.OperationModify,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 1, Count: 1}}},
						{
							Added: []gitdiff.LineRange{
								{Start: 10, Count: 2},
								{Start: 15, Count: 1},
							},
						},
					},
				},
			},
			want: linefilter.LineFilters{
				{
					Name: "foo.txt",
					Lines: []linefilter.Range{
						{Start: 1, End: 1},
						{Start: 10, End: 11},
						{Start: 15, End: 15},
					},
				},
			},
		},
		{
			name: "delete and modify mixed keeps only modify",
			files: []gitdiff.File{
				{
					OldPath:   "del.txt",
					NewPath:   "del.txt",
					Operation: gitdiff.OperationDelete,
					Hunks: []gitdiff.Hunk{
						{Removed: []gitdiff.LineRange{{Start: 1, Count: 2}}},
					},
				},
				{
					OldPath:   "keep.txt",
					NewPath:   "keep.txt",
					Operation: gitdiff.OperationModify,
					Hunks: []gitdiff.Hunk{
						{Added: []gitdiff.LineRange{{Start: 4, Count: 1}}},
					},
				},
			},
			want: linefilter.LineFilters{
				{
					Name:  "keep.txt",
					Lines: []linefilter.Range{{Start: 4, End: 4}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			files := tc.files

			// Act
			filters := linefilter.FromDiff(files...)

			// Assert
			opts := cmp.Options{cmpopts.EquateEmpty()}
			if got, want := filters, tc.want; !cmp.Equal(got, want, opts...) {
				t.Errorf("FromDiff(...) = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts...))
			}
		})
	}
}
