package spdx

import "errors"

// ErrInvalidExpression indicates that an SPDX expression could not be parsed
// because it is structurally invalid.
var ErrInvalidExpression = errors.New("spdx: invalid expression")

// ErrUnexpectedToken indicates that the parser encountered a token that is
// not valid at its position in the expression.
var ErrUnexpectedToken = errors.New("spdx: unexpected token")

// ErrUnexpectedEOF indicates that the parser reached the end of the input
// while a sub-expression was still required.
var ErrUnexpectedEOF = errors.New("spdx: unexpected end of expression")

// ErrUnknownLicense indicates that an identifier used in a license position
// is not part of the embedded SPDX license list.
var ErrUnknownLicense = errors.New("spdx: unknown license identifier")

// ErrUnknownException indicates that an identifier used after a WITH
// operator is not part of the embedded SPDX exception list.
var ErrUnknownException = errors.New("spdx: unknown exception identifier")
