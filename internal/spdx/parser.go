package spdx

import (
	"fmt"
)

// Parse converts an SPDX license expression into its disjunctive normal form.
// Parsing supports the AND, OR, and WITH operators, the "+" suffix on license
// identifiers, and parenthesised grouping.
//
// All license identifiers must be present in the embedded SPDX license list
// and all exception identifiers must be present in the embedded SPDX
// exception list. The returned error wraps one of [ErrInvalidExpression],
// [ErrUnexpectedToken], [ErrUnexpectedEOF], [ErrUnknownLicense], or
// [ErrUnknownException].
func Parse(expr string) (*Expression, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	choices, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if !p.atEnd() {
		return nil, fmt.Errorf("%w: %q", ErrUnexpectedToken, p.peek().value)
	}
	return &Expression{Choices: choices}, nil
}

// tokenKind enumerates the kinds of token the SPDX tokenizer produces.
type tokenKind int

const (
	tokenIdent tokenKind = iota
	tokenAnd
	tokenOr
	tokenWith
	tokenLParen
	tokenRParen
	tokenPlus
)

type token struct {
	kind  tokenKind
	value string
}

// tokenize splits expr into a slice of tokens. It returns
// [ErrInvalidExpression] when expr contains no tokens, and
// [ErrUnexpectedToken] when an invalid character is encountered.
func tokenize(expr string) ([]token, error) {
	var tokens []token
	for i := 0; i < len(expr); {
		ch := expr[i]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			i++
		case ch == '(':
			tokens = append(tokens, token{kind: tokenLParen, value: "("})
			i++
		case ch == ')':
			tokens = append(tokens, token{kind: tokenRParen, value: ")"})
			i++
		case ch == '+':
			tokens = append(tokens, token{kind: tokenPlus, value: "+"})
			i++
		case isIdentStart(ch):
			end := i + 1
			for end < len(expr) && isIdentPart(expr[end]) {
				end++
			}
			value := expr[i:end]
			tokens = append(tokens, classifyIdent(value))
			i = end
		default:
			return nil, fmt.Errorf("%w: %q", ErrUnexpectedToken, string(ch))
		}
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w: empty input", ErrInvalidExpression)
	}
	return tokens, nil
}

// classifyIdent converts a raw identifier string into its corresponding
// [token], distinguishing the keywords AND, OR, and WITH from plain
// identifiers. Keyword matching is case-sensitive per the SPDX spec.
func classifyIdent(value string) token {
	switch value {
	case "AND":
		return token{kind: tokenAnd, value: value}
	case "OR":
		return token{kind: tokenOr, value: value}
	case "WITH":
		return token{kind: tokenWith, value: value}
	}
	return token{kind: tokenIdent, value: value}
}

// isIdentStart reports whether ch may start an SPDX identifier.
func isIdentStart(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

// isIdentPart reports whether ch may appear inside an SPDX identifier
// after its initial character.
func isIdentPart(ch byte) bool {
	return isAlpha(ch) || isDigit(ch) || ch == '-' || ch == '.'
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// parser holds the state for a recursive-descent SPDX expression parse.
type parser struct {
	tokens []token
	pos    int
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.tokens)
}

func (p *parser) peek() token {
	return p.tokens[p.pos]
}

func (p *parser) advance() token {
	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *parser) parseExpr() ([]Choice, error) {
	return p.parseOr()
}

func (p *parser) parseOr() ([]Choice, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for !p.atEnd() && p.peek().kind == tokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = append(left, right...)
	}
	return left, nil
}

func (p *parser) parseAnd() ([]Choice, error) {
	left, err := p.parseWith()
	if err != nil {
		return nil, err
	}
	for !p.atEnd() && p.peek().kind == tokenAnd {
		p.advance()
		right, err := p.parseWith()
		if err != nil {
			return nil, err
		}
		left = andChoices(left, right)
	}
	return left, nil
}

func (p *parser) parseWith() ([]Choice, error) {
	primary, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	if p.atEnd() || p.peek().kind != tokenWith {
		return primary, nil
	}
	p.advance()
	if !isSingleLicense(primary) {
		return nil, fmt.Errorf("%w: WITH requires a license operand", ErrInvalidExpression)
	}
	if p.atEnd() {
		return nil, fmt.Errorf("%w: expected exception identifier after WITH", ErrUnexpectedEOF)
	}
	next := p.peek()
	if next.kind != tokenIdent {
		return nil, fmt.Errorf("%w: expected exception identifier after WITH, got %q", ErrUnexpectedToken, next.value)
	}
	p.advance()
	if !IsKnownExceptionID(next.value) {
		return nil, fmt.Errorf("%w: %q", ErrUnknownException, next.value)
	}
	primary[0].Licenses[0].Exception = next.value
	return primary, nil
}

func (p *parser) parsePrimary() ([]Choice, error) {
	if p.atEnd() {
		return nil, fmt.Errorf("%w: expected license identifier", ErrUnexpectedEOF)
	}
	tok := p.advance()
	switch tok.kind {
	case tokenLParen:
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.atEnd() {
			return nil, fmt.Errorf("%w: expected %q", ErrUnexpectedEOF, ")")
		}
		closing := p.advance()
		if closing.kind != tokenRParen {
			return nil, fmt.Errorf("%w: expected %q, got %q", ErrUnexpectedToken, ")", closing.value)
		}
		return inner, nil
	case tokenIdent:
		return p.finishLicense(tok)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnexpectedToken, tok.value)
	}
}

// finishLicense produces a single-choice, single-license result from an
// identifier token, consuming a trailing "+" if present and verifying the
// identifier against the embedded SPDX license list.
func (p *parser) finishLicense(tok token) ([]Choice, error) {
	orLater := false
	if !p.atEnd() && p.peek().kind == tokenPlus {
		p.advance()
		orLater = true
	}
	if !IsKnownLicenseID(tok.value) {
		return nil, fmt.Errorf("%w: %q", ErrUnknownLicense, tok.value)
	}
	return []Choice{{Licenses: []License{{ID: tok.value, OrLater: orLater}}}}, nil
}

// isSingleLicense reports whether choices represents exactly one [Choice]
// containing exactly one [License], i.e. a primary expression suitable as the
// left operand of WITH.
func isSingleLicense(choices []Choice) bool {
	return len(choices) == 1 && len(choices[0].Licenses) == 1
}

// andChoices returns the disjunctive normal form of the AND of left and
// right via the identity (A | B) & (C | D) == (A & C) | (A & D) | (B & C) |
// (B & D). Each output [Choice] concatenates the [License] lists of one
// [Choice] from left with one from right.
func andChoices(left, right []Choice) []Choice {
	out := make([]Choice, 0, len(left)*len(right))
	for _, l := range left {
		for _, r := range right {
			merged := make([]License, 0, len(l.Licenses)+len(r.Licenses))
			merged = append(merged, l.Licenses...)
			merged = append(merged, r.Licenses...)
			out = append(out, Choice{Licenses: merged})
		}
	}
	return out
}
