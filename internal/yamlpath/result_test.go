package yamlpath_test

import (
	"iter"
	"maps"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.yaml.in/yaml/v4"
)

func TestResult_Iter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		collect func(iter.Seq2[string, *yaml.Node]) map[string]*yaml.Node
		want    map[string]*yaml.Node
	}{
		{
			name:    "collect all",
			collect: maps.Collect[string, *yaml.Node],
			want: map[string]*yaml.Node{
				"foo.a": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"},
				"foo.b": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "2"},
				"foo.c": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "3"},
			},
		}, {
			name:    "collect 2 then stop",
			collect: collect2,
			want: map[string]*yaml.Node{
				"foo.a": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"},
				"foo.b": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			query := mustCompile(t, "foo.*")
			node := mustParseYAML(t, "foo:\n  a: 1\n  b: 2\n  c: 3\n")
			result := query.FindAll(node)

			// Act
			collected := tc.collect(result.Iter())

			// Assert
			opts := compareNodes()
			if got, want := collected, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Result.Iter with %s = mismatch (-want +got):\n%s", tc.name, cmp.Diff(want, got, opts))
			}
		})
	}
}

func collect2(iter iter.Seq2[string, *yaml.Node]) map[string]*yaml.Node {
	out := make(map[string]*yaml.Node)
	i := 0
	for path, node := range iter {
		if i == 2 {
			break
		}
		out[path] = node
		i++
	}
	return out
}
