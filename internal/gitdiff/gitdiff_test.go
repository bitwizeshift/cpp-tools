package gitdiff_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/gitdiff"
)

//go:fix inline
func modePtr(m gitdiff.Mode) *gitdiff.Mode {
	return new(m)
}

type errReader struct {
	err error
}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, r.err
}

var _ io.Reader = (*errReader)(nil)

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    []gitdiff.File
		wantErr error
	}{
		{
			name: "single file modify",
			input: `diff --git a/foo b/foo
index aaa..bbb 100644
--- a/foo
+++ b/foo
@@ -1,3 +1,3 @@
 line1
-line2
+changed
 line3
`,
			want: []gitdiff.File{
				{
					OldPath:   "foo",
					NewPath:   "foo",
					Operation: gitdiff.OperationModify,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added:   []gitdiff.LineRange{{Start: 2, Count: 1}},
							Removed: []gitdiff.LineRange{{Start: 2, Count: 1}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "multi-file modify",
			input: `diff --git a/a.txt b/a.txt
index aaa..bbb 100644
--- a/a.txt
+++ b/a.txt
@@ -1 +1 @@
-old-a
+new-a
diff --git a/b.txt b/b.txt
index ccc..ddd 100644
--- a/b.txt
+++ b/b.txt
@@ -1 +1 @@
-old-b
+new-b
`,
			want: []gitdiff.File{
				{
					OldPath:   "a.txt",
					NewPath:   "a.txt",
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
				{
					OldPath:   "b.txt",
					NewPath:   "b.txt",
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
			name: "preamble and trailer noise",
			input: `commit abc12345
Author: Foo Bar
Date: Today

    Some message body

diff --git a/foo b/foo
index aaa..bbb 100644
--- a/foo
+++ b/foo
@@ -1 +1 @@
-old
+new
trailer line one
trailer line two
`,
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
			name: "new file",
			input: `diff --git a/new.txt b/new.txt
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/new.txt
@@ -0,0 +1,2 @@
+line1
+line2
`,
			want: []gitdiff.File{
				{
					OldPath:   "new.txt",
					NewPath:   "new.txt",
					Operation: gitdiff.OperationCreate,
					OldMode:   nil,
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added: []gitdiff.LineRange{{Start: 1, Count: 2}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "deleted file",
			input: `diff --git a/del.txt b/del.txt
deleted file mode 100644
index abc1234..0000000
--- a/del.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-line1
-line2
`,
			want: []gitdiff.File{
				{
					OldPath:   "del.txt",
					NewPath:   "del.txt",
					Operation: gitdiff.OperationDelete,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   nil,
					Hunks: []gitdiff.Hunk{
						{
							Removed: []gitdiff.LineRange{{Start: 1, Count: 2}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "rename without content change",
			input: `diff --git a/old.txt b/new.txt
similarity index 100%
rename from old.txt
rename to new.txt
`,
			want: []gitdiff.File{
				{
					OldPath:   "old.txt",
					NewPath:   "new.txt",
					Operation: gitdiff.OperationRename,
				},
			},
			wantErr: nil,
		},
		{
			name: "rename with content change",
			input: `diff --git a/old.txt b/new.txt
similarity index 87%
rename from old.txt
rename to new.txt
index abc..def 100644
--- a/old.txt
+++ b/new.txt
@@ -1,3 +1,3 @@
 a
-b
+c
 d
`,
			want: []gitdiff.File{
				{
					OldPath:   "old.txt",
					NewPath:   "new.txt",
					Operation: gitdiff.OperationRename,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added:   []gitdiff.LineRange{{Start: 2, Count: 1}},
							Removed: []gitdiff.LineRange{{Start: 2, Count: 1}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "copy",
			input: `diff --git a/src.txt b/dst.txt
similarity index 100%
copy from src.txt
copy to dst.txt
`,
			want: []gitdiff.File{
				{
					OldPath:   "src.txt",
					NewPath:   "dst.txt",
					Operation: gitdiff.OperationCopy,
				},
			},
			wantErr: nil,
		},
		{
			name: "mode change only",
			input: `diff --git a/script.sh b/script.sh
old mode 100644
new mode 100755
`,
			want: []gitdiff.File{
				{
					OldPath:   "script.sh",
					NewPath:   "script.sh",
					Operation: gitdiff.OperationModify,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeExecutable),
				},
			},
			wantErr: nil,
		},
		{
			name: "binary marker only",
			input: `diff --git a/img.png b/img.png
index abc..def
Binary files a/img.png and b/img.png differ
`,
			want: []gitdiff.File{
				{
					OldPath:   "img.png",
					NewPath:   "img.png",
					Operation: gitdiff.OperationModify,
					Binary:    true,
				},
			},
			wantErr: nil,
		},
		{
			name: "binary git patch followed by text file",
			input: `diff --git a/bin.dat b/bin.dat
new file mode 100644
index 0000000..abc1234
GIT binary patch
literal 12
zcmd6FJ+abc

literal 0
HcmV?d00001
diff --git a/text.txt b/text.txt
index aaa..bbb 100644
--- a/text.txt
+++ b/text.txt
@@ -1 +1 @@
-old
+new
`,
			want: []gitdiff.File{
				{
					OldPath:   "bin.dat",
					NewPath:   "bin.dat",
					Operation: gitdiff.OperationCreate,
					NewMode:   new(gitdiff.ModeFile),
					Binary:    true,
				},
				{
					OldPath:   "text.txt",
					NewPath:   "text.txt",
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
			name: "no newline at end of file",
			input: `diff --git a/foo b/foo
index abc..def 100644
--- a/foo
+++ b/foo
@@ -1,2 +1,2 @@
 a
-b
\ No newline at end of file
+c
\ No newline at end of file
`,
			want: []gitdiff.File{
				{
					OldPath:   "foo",
					NewPath:   "foo",
					Operation: gitdiff.OperationModify,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added:   []gitdiff.LineRange{{Start: 2, Count: 1}},
							Removed: []gitdiff.LineRange{{Start: 2, Count: 1}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "hunk header without counts",
			input: `diff --git a/foo b/foo
index abc..def 100644
--- a/foo
+++ b/foo
@@ -8 +8 @@
-old
+new
`,
			want: []gitdiff.File{
				{
					OldPath:   "foo",
					NewPath:   "foo",
					Operation: gitdiff.OperationModify,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added:   []gitdiff.LineRange{{Start: 8, Count: 1}},
							Removed: []gitdiff.LineRange{{Start: 8, Count: 1}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "multiple non-adjacent hunks",
			input: `diff --git a/foo b/foo
index abc..def 100644
--- a/foo
+++ b/foo
@@ -1,2 +1,3 @@
+inserted
 a
 b
@@ -10,3 +11,2 @@
 x
-y
 z
`,
			want: []gitdiff.File{
				{
					OldPath:   "foo",
					NewPath:   "foo",
					Operation: gitdiff.OperationModify,
					OldMode:   new(gitdiff.ModeFile),
					NewMode:   new(gitdiff.ModeFile),
					Hunks: []gitdiff.Hunk{
						{
							Added: []gitdiff.LineRange{{Start: 1, Count: 1}},
						},
						{
							Removed: []gitdiff.LineRange{{Start: 11, Count: 1}},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "unknown extended header tolerated",
			input: `diff --git a/foo b/foo
some-new-future-header value
index abc..def 100644
--- a/foo
+++ b/foo
@@ -1 +1 @@
-old
+new
`,
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
			name: "index trailing mode unparseable",
			input: `diff --git a/foo b/foo
index abc..def garbage
--- a/foo
+++ b/foo
@@ -1 +1 @@
-old
+new
`,
			want: []gitdiff.File{
				{
					OldPath:   "foo",
					NewPath:   "foo",
					Operation: gitdiff.OperationModify,
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
			name:    "invalid file header missing b path",
			input:   "diff --git no-prefix-pair\n",
			want:    nil,
			wantErr: gitdiff.ErrInvalidFileHeader,
		},
		{
			name:    "invalid file header empty path",
			input:   "diff --git a/ b/\n",
			want:    nil,
			wantErr: gitdiff.ErrInvalidFileHeader,
		},
		{
			name: "invalid hunk header non-numeric range",
			input: `diff --git a/foo b/foo
@@ -bogus +x @@
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidHunkHeader,
		},
		{
			name: "invalid hunk header missing space after marker",
			input: `diff --git a/foo b/foo
@@nospace
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidHunkHeader,
		},
		{
			name: "invalid hunk header missing closing marker",
			input: `diff --git a/foo b/foo
@@ -1,2 +1,2
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidHunkHeader,
		},
		{
			name: "invalid hunk header single token",
			input: `diff --git a/foo b/foo
@@ -1,2 @@
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidHunkHeader,
		},
		{
			name: "invalid hunk header wrong sign prefixes",
			input: `diff --git a/foo b/foo
@@ x y @@
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidHunkHeader,
		},
		{
			name: "invalid hunk header non-numeric count after comma",
			input: `diff --git a/foo b/foo
@@ -1,bad +1,bad @@
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidHunkHeader,
		},
		{
			name: "invalid mode token",
			input: `diff --git a/foo b/foo
new file mode 99Z644
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidMode,
		},
		{
			name: "invalid mode token deleted file",
			input: `diff --git a/foo b/foo
deleted file mode 99Z644
`,
			want:    nil,
			wantErr: gitdiff.ErrInvalidMode,
		},
		{
			name: "truncated hunk body",
			input: `diff --git a/foo b/foo
@@ -1,3 +1,3 @@
 a
`,
			want:    nil,
			wantErr: gitdiff.ErrTruncatedHunk,
		},
		{
			name: "unexpected line in hunk body",
			input: `diff --git a/foo b/foo
@@ -1,2 +1,2 @@
 a
?weird
+b
`,
			want:    nil,
			wantErr: gitdiff.ErrUnexpectedLine,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			reader := strings.NewReader(tc.input)

			// Act
			files, err := gitdiff.Parse(reader)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Parse(...) = err %v, want %v", got, want)
			}
			opts := cmp.Options{cmpopts.EquateEmpty()}
			if got, want := files, tc.want; !cmp.Equal(got, want, opts...) {
				t.Errorf("Parse(...) = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts...))
			}
		})
	}
}

func TestParseString(t *testing.T) {
	t.Parallel()

	// Arrange
	input := `diff --git a/foo b/foo
index abc..def 100644
--- a/foo
+++ b/foo
@@ -1 +1 @@
-old
+new
`
	wantFiles := []gitdiff.File{
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
	}

	// Act
	files, err := gitdiff.ParseString(input)

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ParseString(...) = err %v, want %v", got, want)
	}
	opts := cmp.Options{cmpopts.EquateEmpty()}
	if got, want := files, wantFiles; !cmp.Equal(got, want, opts...) {
		t.Errorf("ParseString(...) = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts...))
	}
}

func TestParseReaderError(t *testing.T) {
	t.Parallel()

	// Arrange
	errBoom := errors.New("boom")
	reader := &errReader{err: errBoom}

	// Act
	files, err := gitdiff.Parse(reader)

	// Assert
	if got, want := err, errBoom; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Parse(errReader) = err %v, want %v", got, want)
	}
	if got, want := files, []gitdiff.File(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Parse(errReader) = %v, want %v", got, want)
	}
}

func TestModeUnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		text    string
		want    gitdiff.Mode
		wantErr error
	}{
		{
			name:    "regular file",
			text:    "100644",
			want:    gitdiff.ModeFile,
			wantErr: nil,
		},
		{
			name:    "executable",
			text:    "100755",
			want:    gitdiff.ModeExecutable,
			wantErr: nil,
		},
		{
			name:    "symlink",
			text:    "120000",
			want:    gitdiff.ModeSymlink,
			wantErr: nil,
		},
		{
			name:    "submodule",
			text:    "160000",
			want:    gitdiff.ModeSubmodule,
			wantErr: nil,
		},
		{
			name:    "non-octal characters",
			text:    "99Z644",
			want:    0,
			wantErr: gitdiff.ErrInvalidMode,
		},
		{
			name:    "empty",
			text:    "",
			want:    0,
			wantErr: gitdiff.ErrInvalidMode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var m gitdiff.Mode

			// Act
			err := m.UnmarshalText([]byte(tc.text))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Mode.UnmarshalText(%q) = err %v, want %v", tc.text, got, want)
			}
			if got, want := m, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Mode.UnmarshalText(%q) = %v, want %v", tc.text, got, want)
			}
		})
	}
}

func TestOperationUnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		text    string
		want    gitdiff.Operation
		wantErr error
	}{
		{
			name:    "modify",
			text:    "modify",
			want:    gitdiff.OperationModify,
			wantErr: nil,
		},
		{
			name:    "create",
			text:    "create",
			want:    gitdiff.OperationCreate,
			wantErr: nil,
		},
		{
			name:    "delete",
			text:    "delete",
			want:    gitdiff.OperationDelete,
			wantErr: nil,
		},
		{
			name:    "rename",
			text:    "rename",
			want:    gitdiff.OperationRename,
			wantErr: nil,
		},
		{
			name:    "copy",
			text:    "copy",
			want:    gitdiff.OperationCopy,
			wantErr: nil,
		},
		{
			name:    "unknown token",
			text:    "remove",
			want:    "",
			wantErr: gitdiff.ErrInvalidOperation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var op gitdiff.Operation

			// Act
			err := op.UnmarshalText([]byte(tc.text))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Operation.UnmarshalText(%q) = err %v, want %v", tc.text, got, want)
			}
			if got, want := op, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Operation.UnmarshalText(%q) = %v, want %v", tc.text, got, want)
			}
		})
	}
}
