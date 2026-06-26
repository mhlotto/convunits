package units

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ParseCandidates parses an expression. A plain ambiguous symbol (notably C
// or F) returns candidates so conversion can resolve it by target dimension.
func (r *Registry) ParseCandidates(text string) ([]Expression, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ParseError{"empty expression"}
	}
	if isPlainName(text) {
		all, err := r.LookupAll(text)
		if err != nil {
			return nil, err
		}
		out := make([]Expression, 0, len(all))
		for _, u := range all {
			e := Expression{Text: text, Dimension: u.Dimension, Multiplier: u.Multiplier, Approximate: u.Approximate}
			if u.Affine {
				e.Affine = u
			}
			out = append(out, e)
		}
		return out, nil
	}
	p := parser{s: text, registry: r}
	e, err := p.expression()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	if p.pos != len(p.s) {
		return nil, ParseError{fmt.Sprintf("unexpected %q at position %d", p.s[p.pos:], p.pos+1)}
	}
	e.Text = text
	return []Expression{e}, nil
}

func isPlainName(s string) bool {
	for _, ch := range s {
		if ch == '*' || ch == '/' || ch == '^' || ch == '(' || ch == ')' {
			return false
		}
	}
	return true
}

type parser struct {
	s        string
	pos      int
	registry *Registry
}

func (p *parser) expression() (Expression, error) {
	left, err := p.factor()
	if err != nil {
		return Expression{}, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) || (p.s[p.pos] != '*' && p.s[p.pos] != '/') {
			break
		}
		op := p.s[p.pos]
		p.pos++
		right, err := p.factor()
		if err != nil {
			return Expression{}, err
		}
		if left.Affine != nil || right.Affine != nil {
			return Expression{}, fmt.Errorf("affine temperature units cannot be used in compound expressions")
		}
		if op == '*' {
			left.Dimension = left.Dimension.Add(right.Dimension)
			left.Multiplier *= right.Multiplier
		} else {
			left.Dimension = left.Dimension.Sub(right.Dimension)
			left.Multiplier /= right.Multiplier
		}
		left.Approximate = left.Approximate || right.Approximate
	}
	return left, nil
}

func (p *parser) factor() (Expression, error) {
	p.skipSpace()
	var e Expression
	var err error
	if p.pos < len(p.s) && p.s[p.pos] == '(' {
		p.pos++
		e, err = p.expression()
		if err != nil {
			return e, err
		}
		p.skipSpace()
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return e, ParseError{"missing closing parenthesis"}
		}
		p.pos++
	} else {
		start := p.pos
		for p.pos < len(p.s) {
			ch := rune(p.s[p.pos])
			if ch == '*' || ch == '/' || ch == '^' || ch == '(' || ch == ')' || unicode.IsSpace(ch) {
				break
			}
			p.pos++
		}
		if start == p.pos {
			return e, ParseError{fmt.Sprintf("expected unit at position %d", p.pos+1)}
		}
		name := p.s[start:p.pos]
		u, lookupErr := p.registry.lookupCompound(name)
		if lookupErr != nil {
			return e, lookupErr
		}
		e = Expression{Dimension: u.Dimension, Multiplier: u.Multiplier, Approximate: u.Approximate}
	}
	p.skipSpace()
	if p.pos < len(p.s) && p.s[p.pos] == '^' {
		if e.Affine != nil {
			return e, fmt.Errorf("affine temperature units cannot be raised to a power")
		}
		p.pos++
		p.skipSpace()
		start := p.pos
		if p.pos < len(p.s) && (p.s[p.pos] == '-' || p.s[p.pos] == '+') {
			p.pos++
		}
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			p.pos++
		}
		if start == p.pos {
			return e, ParseError{"missing integer exponent"}
		}
		n, convErr := strconv.Atoi(p.s[start:p.pos])
		if convErr != nil {
			return e, ParseError{"invalid integer exponent"}
		}
		e.Dimension = e.Dimension.Mul(n)
		e.Multiplier = powInt(e.Multiplier, n)
	}
	return e, nil
}

func (p *parser) skipSpace() {
	for p.pos < len(p.s) && unicode.IsSpace(rune(p.s[p.pos])) {
		p.pos++
	}
}

func powInt(v float64, n int) float64 {
	if n < 0 {
		return 1 / powInt(v, -n)
	}
	r := 1.0
	for n > 0 {
		if n&1 == 1 {
			r *= v
		}
		v *= v
		n >>= 1
	}
	return r
}
