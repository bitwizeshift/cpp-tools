package gitdiff_test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/gitdiff"
)

func readDiffFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = err %v, want nil", path, err)
	}
	return string(data)
}

func TestParseConformance(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		path string
	}{
		{
			name: "Additions",
			path: "testdata/additions.diff",
		}, {
			name: "AdditionsAndDeletions",
			path: "testdata/additions_and_deletions.diff",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			input := readDiffFile(t, tc.path)
			reader := strings.NewReader(input)

			// Act
			_, err := gitdiff.Parse(reader)

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Errorf("Parse(%s) = %v, want %v", tc.path, got, want)
			}
		})
	}
}
