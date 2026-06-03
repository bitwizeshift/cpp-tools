package yamlpath

import "errors"

// ErrUnexpectedToken indicates that the tokenizer or parser encountered a token
// that is not valid in the current position.
var ErrUnexpectedToken = errors.New("yamlpath: unexpected token")

// ErrUnexpectedEOF indicates that the parser reached the end of the query while
// still expecting additional tokens.
var ErrUnexpectedEOF = errors.New("yamlpath: unexpected end of query")
