// package yamlpath compiles compact path expressions and evaluates them
// against parsed YAML node trees.
//
// A query is a dot-and-bracket expression such as foo.bar[0].* that walks
// mapping fields, sequence indices, and the corresponding wildcards. Use
// [Compile] to parse a query, then [Query.FindAll] to extract every matching
// (path, node) pair from a [yaml.Node]. Results carry canonical path strings
// reconstructed from the actual traversal so that wildcard matches report the
// concrete key or index they expanded to.
package yamlpath
