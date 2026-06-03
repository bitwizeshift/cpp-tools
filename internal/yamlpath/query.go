package yamlpath

import "go.yaml.in/yaml/v4"

// Query is a compiled yamlpath expression that can be applied to YAML nodes to
// extract matching sub-nodes together with their canonical paths.
type Query struct {
	query string
	root  expression
}

// Compile parses path into a [Query]. It may return [ErrUnexpectedToken], or
// [ErrUnexpectedEOF] when path is not a valid query.
func Compile(path string) (*Query, error) {
	root, err := parseQuery(path)
	if err != nil {
		return nil, err
	}
	return &Query{query: path, root: root}, nil
}

// String returns the original query text that was passed to [Compile].
func (q *Query) String() string {
	return q.query
}

// FindAll evaluates the query against node and returns the matched entries as
// a [Result]. A document node is unwrapped to its underlying root before
// evaluation.
func (q *Query) FindAll(node *yaml.Node) *Result {
	seed := node
	if seed.Kind == yaml.DocumentNode && len(seed.Content) > 0 {
		seed = seed.Content[0]
	}
	in := &Result{entries: []entry{{name: "", node: seed}}}
	return q.root.Evaluate(in)
}
