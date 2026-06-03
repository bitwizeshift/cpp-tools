package yamlpath_test

import (
	"maps"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.yaml.in/yaml/v4"

	"rodusek.dev/pkg/cpp-tools/internal/yamlpath"
)

// compareNodes returns a cmp.Option that compares yaml.Node values only by
// Kind, Tag, and Value.
func compareNodes() cmp.Option {
	return cmp.Comparer(func(x, y yaml.Node) bool {
		return x.Kind == y.Kind && x.Tag == y.Tag && x.Value == y.Value
	})
}

// mustCompile compiles a yamlpath query, failing the test on error.
func mustCompile(t *testing.T, input string) *yamlpath.Query {
	t.Helper()
	query, err := yamlpath.Compile(input)
	if err != nil {
		t.Fatalf("Compile(%q) error = %v, want nil", input, err)
	}
	return query
}

// mustParseYAML parses src as YAML, failing the test on error.
func mustParseYAML(t *testing.T, src string) *yaml.Node {
	t.Helper()
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(src), &n); err != nil {
		t.Fatalf("yaml.Unmarshal(...) error = %v", err)
	}
	return &n
}

func TestCompile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:  "empty query",
			input: "",
		},
		{
			name:  "single field",
			input: "foo",
		},
		{
			name:  "field with hyphen",
			input: "foo-bar",
		},
		{
			name:  "nested fields",
			input: "foo.bar",
		},
		{
			name:  "indexed field",
			input: "foo[0]",
		},
		{
			name:  "wildcard field",
			input: "foo.*",
		},
		{
			name:  "wildcard index",
			input: "foo[*]",
		},
		{
			name:  "mixed",
			input: "foo[0].bar.*[1]",
		},
		{
			name:    "leading dot",
			input:   ".foo",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "trailing dot",
			input:   "foo.",
			wantErr: yamlpath.ErrUnexpectedEOF,
		},
		{
			name:    "bare star",
			input:   "*",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "unterminated bracket",
			input:   "foo[0",
			wantErr: yamlpath.ErrUnexpectedEOF,
		},
		{
			name:    "empty bracket",
			input:   "foo[]",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "non integer index",
			input:   "foo[abc]",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "unknown character",
			input:   "foo$",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "leading bracket",
			input:   "[0]",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "double dot",
			input:   "foo..bar",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "bracket then eof",
			input:   "foo[",
			wantErr: yamlpath.ErrUnexpectedEOF,
		},
		{
			name:    "missing close bracket",
			input:   "foo[0bar]",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "stray close bracket after field",
			input:   "foo]",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
		{
			name:    "leading integer",
			input:   "0foo",
			wantErr: yamlpath.ErrUnexpectedToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange

			// Act
			_, err := yamlpath.Compile(tc.input)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Compile(%q) error = %v, want %v", tc.input, got, want)
			}
		})
	}
}

func TestQuery_String(t *testing.T) {
	t.Parallel()

	// Arrange
	input := "foo[0].bar.*[1]"
	query := mustCompile(t, input)

	// Act
	str := query.String()

	// Assert
	if got, want := str, input; !cmp.Equal(got, want) {
		t.Errorf("Query.String() = %q, want %q", got, want)
	}
}

func TestQuery_FindAll(t *testing.T) {
	t.Parallel()

	const sampleYAML = `
foo:
  bar: hello
  baz: world
  nums:
    - 10
    - 20
    - 30
list:
  - name: a
    value: 1
  - name: b
    value: 2
plain: scalar
`

	testCases := []struct {
		name  string
		query string
		input string
		want  map[string]*yaml.Node
	}{
		{
			name:  "empty query returns root mapping",
			query: "",
			input: sampleYAML,
			want: map[string]*yaml.Node{
				"": {Kind: yaml.MappingNode, Tag: "!!map"},
			},
		},
		{
			name:  "field hit on scalar",
			query: "plain",
			input: sampleYAML,
			want: map[string]*yaml.Node{
				"plain": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "scalar"},
			},
		},
		{
			name:  "missing field returns empty",
			query: "missing",
			input: sampleYAML,
			want:  nil,
		},
		{
			name:  "nested field",
			query: "foo.bar",
			input: sampleYAML,
			want: map[string]*yaml.Node{
				"foo.bar": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "hello"},
			},
		},
		{
			name:  "nested field missing leaf",
			query: "foo.missing",
			input: sampleYAML,
			want:  nil,
		},
		{
			name:  "field then index",
			query: "foo.nums[1]",
			input: sampleYAML,
			want: map[string]*yaml.Node{
				"foo.nums[1]": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "20"},
			},
		},
		{
			name:  "index out of range",
			query: "foo.nums[9]",
			input: sampleYAML,
			want:  nil,
		},
		{
			name:  "index on non-sequence",
			query: "foo[0]",
			input: sampleYAML,
			want:  nil,
		},
		{
			name:  "wildcard field expands mapping",
			query: "foo.*",
			input: "foo:\n  a: 1\n  b: 2\n",
			want: map[string]*yaml.Node{
				"foo.a": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"},
				"foo.b": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "2"},
			},
		},
		{
			name:  "wildcard field on non-mapping",
			query: "plain.*",
			input: sampleYAML,
			want:  nil,
		},
		{
			name:  "field on non-mapping",
			query: "plain.bar",
			input: sampleYAML,
			want:  nil,
		},
		{
			name:  "wildcard index expands sequence",
			query: "foo.nums[*]",
			input: sampleYAML,
			want: map[string]*yaml.Node{
				"foo.nums[0]": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "10"},
				"foo.nums[1]": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "20"},
				"foo.nums[2]": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "30"},
			},
		},
		{
			name:  "wildcard index on non-sequence",
			query: "foo[*]",
			input: sampleYAML,
			want:  nil,
		},
		{
			name:  "wildcard chain via mapping then field",
			query: "list[*].name",
			input: sampleYAML,
			want: map[string]*yaml.Node{
				"list[0].name": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "a"},
				"list[1].name": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "b"},
			},
		},
		{
			name:  "wildcard field followed by mapping",
			query: "list.*",
			input: "list:\n  a:\n    name: one\n  b:\n    name: two\n",
			want: map[string]*yaml.Node{
				"list.a": {Kind: yaml.MappingNode, Tag: "!!map"},
				"list.b": {Kind: yaml.MappingNode, Tag: "!!map"},
			},
		},
		{
			name:  "first-step field on root",
			query: "foo",
			input: sampleYAML,
			want: map[string]*yaml.Node{
				"foo": {Kind: yaml.MappingNode, Tag: "!!map"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := mustCompile(t, tc.query)
			node := mustParseYAML(t, tc.input)

			// Act
			result := maps.Collect(sut.FindAll(node).Iter())

			// Assert
			opts := cmp.Options{cmpopts.EquateEmpty(), compareNodes()}
			if got, want := result, tc.want; !cmp.Equal(got, want, opts) {
				t.Errorf("Query.FindAll(...) = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func TestQuery_FindAll_NonDocumentNode(t *testing.T) {
	t.Parallel()

	// Arrange
	doc := mustParseYAML(t, "foo: 1\n")
	root := doc.Content[0]
	sut := mustCompile(t, "foo")
	want := map[string]*yaml.Node{
		"foo": {Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"},
	}

	// Act
	result := maps.Collect(sut.FindAll(root).Iter())

	// Assert
	opts := cmp.Options{compareNodes()}
	if got, want := result, want; !cmp.Equal(got, want, opts) {
		t.Errorf("Query.FindAll(non-document node) = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
	}
}
