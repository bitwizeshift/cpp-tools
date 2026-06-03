package yamlpath

import (
	"strconv"

	"go.yaml.in/yaml/v4"
)

// expression is one step of a compiled yamlpath query that produces a new
// [Result] from a given input [Result].
type expression interface {
	Evaluate(*Result) *Result
}

// compoundExpression chains sub-expressions left-to-right, feeding the output
// of each step into the next.
type compoundExpression []expression

// Evaluate applies each contained expression in order, returning the final
// [Result].
func (c compoundExpression) Evaluate(in *Result) *Result {
	current := in
	for _, e := range c {
		current = e.Evaluate(current)
	}
	return current
}

// fieldExpression selects a single named field from each mapping entry in the
// input.
type fieldExpression struct {
	name string
}

// Evaluate emits one entry for every input mapping node that contains the
// requested key.
func (f fieldExpression) Evaluate(in *Result) *Result {
	out := &Result{}
	for _, e := range in.entries {
		value, ok := mappingValue(e.node, f.name)
		if !ok {
			continue
		}
		out.entries = append(out.entries, entry{
			name: joinField(e.name, f.name),
			node: value,
		})
	}
	return out
}

// wildcardFieldExpression selects every key-value pair from each mapping entry
// in the input.
type wildcardFieldExpression struct{}

// Evaluate emits one entry per (key, value) pair contained in each mapping
// node from the input.
func (wildcardFieldExpression) Evaluate(in *Result) *Result {
	out := &Result{}
	for _, e := range in.entries {
		if e.node.Kind != yaml.MappingNode {
			continue
		}
		for i := 0; i+1 < len(e.node.Content); i += 2 {
			key := e.node.Content[i]
			value := e.node.Content[i+1]
			out.entries = append(out.entries, entry{
				name: joinField(e.name, key.Value),
				node: value,
			})
		}
	}
	return out
}

// indexExpression selects a single element from each sequence entry in the
// input by zero-based position.
type indexExpression struct {
	index int
}

// Evaluate emits one entry for every input sequence node whose length exceeds
// the configured index.
func (i indexExpression) Evaluate(in *Result) *Result {
	out := &Result{}
	for _, e := range in.entries {
		if e.node.Kind != yaml.SequenceNode {
			continue
		}
		if i.index >= len(e.node.Content) {
			continue
		}
		out.entries = append(out.entries, entry{
			name: joinIndex(e.name, i.index),
			node: e.node.Content[i.index],
		})
	}
	return out
}

// wildcardIndexExpression selects every element from each sequence entry in the
// input.
type wildcardIndexExpression struct{}

// Evaluate emits one entry per element contained in each sequence node from
// the input.
func (wildcardIndexExpression) Evaluate(in *Result) *Result {
	out := &Result{}
	for _, e := range in.entries {
		if e.node.Kind != yaml.SequenceNode {
			continue
		}
		for i, child := range e.node.Content {
			out.entries = append(out.entries, entry{
				name: joinIndex(e.name, i),
				node: child,
			})
		}
	}
	return out
}

// mappingValue returns the value for key in node when node is a mapping node
// containing that key, along with whether the key was found.
func mappingValue(node *yaml.Node, key string) (*yaml.Node, bool) {
	if node.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1], true
		}
	}
	return nil, false
}

// joinField returns the path produced by appending key as a field step to
// parent.
func joinField(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

// joinIndex returns the path produced by appending i as an index step to
// parent.
func joinIndex(parent string, i int) string {
	return parent + "[" + strconv.Itoa(i) + "]"
}
