package gitdiff_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/command"
	"rodusek.dev/pkg/cpp-tools/internal/command/commandtest"
	"rodusek.dev/pkg/cpp-tools/internal/gitdiff"
)

const validDiff = `diff --git a/foo b/foo
index aaa..bbb 100644
--- a/foo
+++ b/foo
@@ -1 +1 @@
-old
+new
`

const invalidDiff = `diff --git invalid header
`

func TestListerList(t *testing.T) {
	t.Parallel()

	errStart := errors.New("start failed")
	errStdoutPipe := errors.New("stdout pipe failed")
	errWait := errors.New("wait failed")

	testCases := []struct {
		name    string
		creator command.Creator
		opts    *gitdiff.ListOptions
		want    []gitdiff.File
		wantErr error
	}{
		{
			name:    "nil options succeeds",
			creator: commandtest.Pipes(validDiff, ""),
			opts:    nil,
			want: []gitdiff.File{
				{
					OldPath:   "foo",
					NewPath:   "foo",
					Operation: gitdiff.OperationModify,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added:   []gitdiff.LineRange{{Start: 1, Count: 1}},
							Removed: []gitdiff.LineRange{{Start: 1, Count: 1}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "indexed with paths succeeds",
			creator: commandtest.Pipes(validDiff, ""),
			opts: &gitdiff.ListOptions{
				Indexed: true,
				Paths:   []string{"foo", "bar"},
			},
			want: []gitdiff.File{
				{
					OldPath:   "foo",
					NewPath:   "foo",
					Operation: gitdiff.OperationModify,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added:   []gitdiff.LineRange{{Start: 1, Count: 1}},
							Removed: []gitdiff.LineRange{{Start: 1, Count: 1}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "stdout pipe error",
			creator: commandtest.ErrOnStdoutPipe(errStdoutPipe),
			opts:    &gitdiff.ListOptions{},
			want:    nil,
			wantErr: errStdoutPipe,
		},
		{
			name:    "start error",
			creator: commandtest.ErrOnStart(errStart),
			opts:    &gitdiff.ListOptions{},
			want:    nil,
			wantErr: errStart,
		},
		{
			name:    "parse error",
			creator: commandtest.Pipes(invalidDiff, ""),
			opts:    &gitdiff.ListOptions{},
			want:    nil,
			wantErr: gitdiff.ErrInvalidFileHeader,
		},
		{
			name:    "wait error",
			creator: commandtest.PipesAndWaitErr(validDiff, "", errWait),
			opts:    &gitdiff.ListOptions{},
			want:    nil,
			wantErr: errWait,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			sut := &gitdiff.Lister{
				CommandCreator: tc.creator,
			}

			// Act
			files, err := sut.List(ctx, tc.opts)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Lister.List(...) error = %v, want %v", got, want)
			}
			opts := []cmp.Option{cmpopts.EquateEmpty()}
			if got, want := files, tc.want; !cmp.Equal(got, want, opts...) {
				t.Errorf("Lister.List(...) = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts...))
			}
		})
	}
}
