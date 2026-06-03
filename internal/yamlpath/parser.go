package yamlpath

import "fmt"

// tokenKind identifies a token category produced by the yamlpath tokenizer.
type tokenKind int

const (
	tEOF tokenKind = iota
	tIdent
	tDot
	tStar
	tLBracket
	tRBracket
	tInt
)

// token is a lexical unit produced by the yamlpath tokenizer.
type token struct {
	kind  tokenKind
	text  string
	value int
}

// tokenize converts the input query string into a sequence of tokens. It
// returns [ErrUnexpectedToken] wrapped with positional context when an
// unrecognized character is encountered.
func tokenize(input string) ([]token, error) {
	var tokens []token
	for i := 0; i < len(input); {
		c := input[i]
		switch {
		case c == '.':
			tokens = append(tokens, token{kind: tDot, text: "."})
			i++
		case c == '*':
			tokens = append(tokens, token{kind: tStar, text: "*"})
			i++
		case c == '[':
			tokens = append(tokens, token{kind: tLBracket, text: "["})
			i++
		case c == ']':
			tokens = append(tokens, token{kind: tRBracket, text: "]"})
			i++
		case isDigit(c):
			j := i
			n := 0
			for j < len(input) && isDigit(input[j]) {
				n = n*10 + int(input[j]-'0')
				j++
			}
			tokens = append(tokens, token{kind: tInt, text: input[i:j], value: n})
			i = j
		case isIdentStart(c):
			j := i
			for j < len(input) && isIdentCont(input[j]) {
				j++
			}
			tokens = append(tokens, token{kind: tIdent, text: input[i:j]})
			i = j
		default:
			return nil, fmt.Errorf("at position %d: %q: %w", i, string(c), ErrUnexpectedToken)
		}
	}
	tokens = append(tokens, token{kind: tEOF, text: "end of query"})
	return tokens, nil
}

// isDigit reports whether c is an ASCII decimal digit.
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isIdentStart reports whether c is a valid leading character for an
// identifier.
func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isIdentCont reports whether c is a valid continuation character for an
// identifier.
func isIdentCont(c byte) bool {
	return isIdentStart(c) || isDigit(c) || c == '-'
}

// parser consumes a token stream and produces a compoundExpression.
type parser struct {
	tokens []token
	pos    int
}

// peek returns the current token without advancing.
func (p *parser) peek() token {
	return p.tokens[p.pos]
}

// advance returns the current token and moves past it.
func (p *parser) advance() token {
	t := p.tokens[p.pos]
	p.pos++
	return t
}

// parse drives the Pratt-style loop, returning the parsed expression or a
// wrapped sentinel error.
func (p *parser) parse() (compoundExpression, error) {
	expr := compoundExpression{}
	first := p.peek()
	if first.kind == tEOF {
		return expr, nil
	}
	if first.kind != tIdent {
		return nil, unexpectedToken(first)
	}
	p.advance()
	expr = append(expr, fieldExpression{name: first.text})

	for {
		next, err := p.parsePostfix()
		if err != nil {
			return nil, err
		}
		if next == nil {
			return expr, nil
		}
		expr = append(expr, next)
	}
}

// parsePostfix consumes a single postfix operator and returns the resulting
// expression. It returns a nil expression with a nil error when the end of the
// token stream is reached.
func (p *parser) parsePostfix() (expression, error) {
	t := p.advance()
	switch t.kind {
	case tEOF:
		return nil, nil
	case tDot:
		return p.parseDotSuffix()
	case tLBracket:
		return p.parseIndexSuffix()
	default:
		return nil, unexpectedToken(t)
	}
}

// parseDotSuffix parses the operand that follows a `.` token.
func (p *parser) parseDotSuffix() (expression, error) {
	t := p.advance()
	switch t.kind {
	case tIdent:
		return fieldExpression{name: t.text}, nil
	case tStar:
		return wildcardFieldExpression{}, nil
	case tEOF:
		return nil, fmt.Errorf("after '.': %w", ErrUnexpectedEOF)
	default:
		return nil, unexpectedToken(t)
	}
}

// parseIndexSuffix parses the body and closing bracket of an `[...]` operator.
func (p *parser) parseIndexSuffix() (expression, error) {
	body := p.advance()
	var expr expression
	switch body.kind {
	case tInt:
		expr = indexExpression{index: body.value}
	case tStar:
		expr = wildcardIndexExpression{}
	case tEOF:
		return nil, fmt.Errorf("after '[': %w", ErrUnexpectedEOF)
	default:
		return nil, unexpectedToken(body)
	}
	closer := p.advance()
	switch closer.kind {
	case tRBracket:
		return expr, nil
	case tEOF:
		return nil, fmt.Errorf("expected ']': %w", ErrUnexpectedEOF)
	default:
		return nil, unexpectedToken(closer)
	}
}

// unexpectedToken returns an [ErrUnexpectedToken] wrapped with context about t.
func unexpectedToken(t token) error {
	return fmt.Errorf("%q: %w", t.text, ErrUnexpectedToken)
}

// parseQuery tokenizes and parses input into a compoundExpression. Returned
// errors wrap [ErrUnexpectedToken] or [ErrUnexpectedEOF].
func parseQuery(input string) (compoundExpression, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	return p.parse()
}
