package yamlpath

import (
	"iter"

	"go.yaml.in/yaml/v4"
)

type entry struct {
	name string
	node *yaml.Node
}

// Result is the collection of (path, node) pairs produced by evaluating a
// [Query] against a YAML document.
type Result struct {
	entries []entry
}

// Iter returns a [iter.Seq2] that yields each matched path together with its
// corresponding YAML node, in match order.
func (r *Result) Iter() iter.Seq2[string, *yaml.Node] {
	return func(yield func(string, *yaml.Node) bool) {
		for _, e := range r.entries {
			if !yield(e.name, e.node) {
				return
			}
		}
	}
}
