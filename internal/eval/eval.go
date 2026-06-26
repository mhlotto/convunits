package eval

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"convunits/internal/units"
)

const (
	speedOfLight          = 299792458.0
	gravitationalConstant = 6.67430e-11
	standardGravity       = 9.80665
)

type Result struct {
	Value       float64
	Dimension   units.Dimension
	Unit        string
	Approximate bool
}

type Evaluator struct {
	Registry *units.Registry
}

func New(registry *units.Registry) *Evaluator { return &Evaluator{Registry: registry} }

func (e *Evaluator) Evaluate(text string) (Result, error) {
	expression, output, hasOutput := splitArrow(text)
	parser := parser{s: strings.TrimSpace(expression), evaluator: e}
	result, err := parser.expression()
	if err != nil {
		return Result{}, err
	}
	parser.skipSpace()
	if parser.pos != len(parser.s) {
		return Result{}, fmt.Errorf("unexpected %q at position %d", parser.s[parser.pos:], parser.pos+1)
	}
	if hasOutput {
		converted, err := e.convertTo(result, output)
		if err != nil {
			return Result{}, err
		}
		return converted, nil
	}
	return Result{Value: result.value, Dimension: result.dimension, Approximate: result.approximate}, nil
}

func splitArrow(text string) (string, string, bool) {
	before, after, ok := strings.Cut(text, "->")
	return before, strings.TrimSpace(after), ok
}

type quantity struct {
	value                     float64
	dimension                 units.Dimension
	approximate               bool
	temperatureLookingLiteral bool
}

var temperatureDimension = units.Dimension{Temperature: 1}

const affineTemperatureEvalError = "eval does not treat Celsius/Fahrenheit as ordinary scalar units; use normal conversion mode, e.g. `convunits 100C F`"

func (e *Evaluator) convertTo(q quantity, output string) (Result, error) {
	if output == "" {
		return Result{}, fmt.Errorf("missing output unit after ->")
	}
	if q.temperatureLookingLiteral && temperatureLookingOutput(output) {
		return Result{}, fmt.Errorf(affineTemperatureEvalError)
	}
	outs, err := e.Registry.ParseCandidates(output)
	if err != nil {
		return Result{}, fmt.Errorf("output unit: %w", err)
	}
	var matches []units.Expression
	for _, out := range outs {
		if out.Affine != nil {
			continue
		}
		if out.Dimension == q.dimension {
			matches = append(matches, out)
		}
	}
	if len(matches) == 0 {
		return Result{}, fmt.Errorf("cannot convert %s to %s: incompatible dimensions (%s and %s)", q.dimension, output, q.dimension, outs[0].Dimension)
	}
	if len(matches) > 1 {
		return Result{}, fmt.Errorf("ambiguous output unit %q; use a long unit name", output)
	}
	out := matches[0]
	return Result{Value: out.FromBase(q.value), Dimension: q.dimension, Unit: output, Approximate: q.approximate || out.Approximate}, nil
}

type parser struct {
	s         string
	pos       int
	evaluator *Evaluator
}

func (p *parser) expression() (quantity, error) {
	left, err := p.term()
	if err != nil {
		return quantity{}, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) || (p.s[p.pos] != '+' && p.s[p.pos] != '-') {
			return left, nil
		}
		op := p.s[p.pos]
		p.pos++
		right, err := p.term()
		if err != nil {
			return quantity{}, err
		}
		if left.dimension != right.dimension {
			if temperatureLookingMismatch(left, right) {
				return quantity{}, fmt.Errorf(affineTemperatureEvalError)
			}
			return quantity{}, fmt.Errorf("cannot add/subtract incompatible dimensions %s and %s", left.dimension, right.dimension)
		}
		if op == '+' {
			left.value += right.value
		} else {
			left.value -= right.value
		}
		left.approximate = left.approximate || right.approximate
		left.temperatureLookingLiteral = left.temperatureLookingLiteral || right.temperatureLookingLiteral
	}
}

func (p *parser) term() (quantity, error) {
	left, err := p.power()
	if err != nil {
		return quantity{}, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) || (p.s[p.pos] != '*' && p.s[p.pos] != '/') {
			return left, nil
		}
		op := p.s[p.pos]
		p.pos++
		right, err := p.power()
		if err != nil {
			return quantity{}, err
		}
		if op == '*' {
			left.value *= right.value
			left.dimension = left.dimension.Add(right.dimension)
		} else {
			if right.value == 0 {
				return quantity{}, fmt.Errorf("division by zero")
			}
			left.value /= right.value
			left.dimension = left.dimension.Sub(right.dimension)
		}
		left.approximate = left.approximate || right.approximate
		left.temperatureLookingLiteral = left.temperatureLookingLiteral || right.temperatureLookingLiteral
	}
}

func (p *parser) power() (quantity, error) {
	left, err := p.unary()
	if err != nil {
		return quantity{}, err
	}
	p.skipSpace()
	if p.pos >= len(p.s) || p.s[p.pos] != '^' {
		return left, nil
	}
	p.pos++
	right, err := p.power()
	if err != nil {
		return quantity{}, err
	}
	if right.dimension != (units.Dimension{}) {
		return quantity{}, fmt.Errorf("power exponent must be dimensionless")
	}
	if left.dimension == (units.Dimension{}) {
		left.value = math.Pow(left.value, right.value)
		left.approximate = left.approximate || right.approximate
		left.temperatureLookingLiteral = left.temperatureLookingLiteral || right.temperatureLookingLiteral
		return left, nil
	}
	exponent := math.Round(right.value)
	if math.Abs(right.value-exponent) > 1e-12 {
		return quantity{}, fmt.Errorf("dimensioned powers require an integer exponent")
	}
	left.value = math.Pow(left.value, exponent)
	left.dimension = left.dimension.Mul(int(exponent))
	left.approximate = left.approximate || right.approximate
	left.temperatureLookingLiteral = left.temperatureLookingLiteral || right.temperatureLookingLiteral
	return left, nil
}

func (p *parser) unary() (quantity, error) {
	p.skipSpace()
	if p.pos < len(p.s) && p.s[p.pos] == '-' {
		p.pos++
		q, err := p.unary()
		q.value = -q.value
		return q, err
	}
	return p.primary()
}

func (p *parser) primary() (quantity, error) {
	p.skipSpace()
	if p.pos >= len(p.s) {
		return quantity{}, fmt.Errorf("expected expression at position %d", p.pos+1)
	}
	if p.s[p.pos] == '(' {
		p.pos++
		q, err := p.expression()
		if err != nil {
			return quantity{}, err
		}
		p.skipSpace()
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return quantity{}, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return q, nil
	}
	if isNumberStart(p.s[p.pos]) {
		return p.numberWithOptionalUnit()
	}
	if isNameStart(rune(p.s[p.pos])) {
		name := p.readName()
		return p.namedQuantity(name)
	}
	return quantity{}, fmt.Errorf("unexpected %q at position %d", p.s[p.pos:p.pos+1], p.pos+1)
}

func (p *parser) numberWithOptionalUnit() (quantity, error) {
	start := p.pos
	for p.pos < len(p.s) {
		ch := p.s[p.pos]
		if (ch >= '0' && ch <= '9') || ch == '.' || ch == 'e' || ch == 'E' || ((ch == '+' || ch == '-') && p.pos > start && (p.s[p.pos-1] == 'e' || p.s[p.pos-1] == 'E')) {
			p.pos++
			continue
		}
		break
	}
	value, err := strconv.ParseFloat(p.s[start:p.pos], 64)
	if err != nil {
		return quantity{}, fmt.Errorf("invalid number %q", p.s[start:p.pos])
	}
	unitStart := p.pos
	if unitStart < len(p.s) && isUnitStart(rune(p.s[unitStart])) {
		p.pos = p.attachedUnitEnd(unitStart)
		unitText := p.s[unitStart:p.pos]
		expr, err := p.singleUnitExpression(unitText)
		if err != nil {
			return quantity{}, err
		}
		return quantity{value: expr.ToBase(value), dimension: expr.Dimension, approximate: expr.Approximate, temperatureLookingLiteral: unitText == "C" || unitText == "F"}, nil
	}
	return quantity{value: value}, nil
}

func (p *parser) attachedUnitEnd(start int) int {
	pos := start
	for pos < len(p.s) {
		ch := rune(p.s[pos])
		if isUnitSuffixStop(ch) {
			break
		}
		if (ch == '*' || ch == '/') && p.operatorStartsEvalFactor(pos+1) {
			break
		}
		pos++
	}
	return pos
}

func (p *parser) operatorStartsEvalFactor(pos int) bool {
	for pos < len(p.s) && unicode.IsSpace(rune(p.s[pos])) {
		pos++
	}
	if pos >= len(p.s) {
		return false
	}
	if p.s[pos] == '(' || isNumberStart(p.s[pos]) {
		return true
	}
	if isNameStart(rune(p.s[pos])) {
		start := pos
		for pos < len(p.s) && isNamePart(rune(p.s[pos])) {
			pos++
		}
		return isEvalConstant(p.s[start:pos])
	}
	return false
}

func (p *parser) namedQuantity(name string) (quantity, error) {
	switch name {
	case "pi":
		return quantity{value: math.Pi}, nil
	case "c":
		return quantity{value: speedOfLight, dimension: units.Dimension{Length: 1, Time: -1}}, nil
	case "G":
		return quantity{value: gravitationalConstant, dimension: units.Dimension{Length: 3, Mass: -1, Time: -2}, approximate: true}, nil
	case "g0":
		return quantity{value: standardGravity, dimension: units.Dimension{Length: 1, Time: -2}}, nil
	}
	expr, err := p.singleUnitExpression(name)
	if err != nil {
		return quantity{}, err
	}
	return quantity{value: expr.ToBase(1), dimension: expr.Dimension, approximate: expr.Approximate}, nil
}

func (p *parser) singleUnitExpression(text string) (units.Expression, error) {
	expressions, err := p.evaluator.Registry.ParseCandidates(text)
	if err != nil {
		return units.Expression{}, fmt.Errorf("unit %q: %w", text, err)
	}
	var usable []units.Expression
	for _, expr := range expressions {
		if expr.Affine == nil {
			usable = append(usable, expr)
		}
	}
	if len(usable) == 1 {
		return usable[0], nil
	}
	if len(usable) == 0 {
		return units.Expression{}, fmt.Errorf("eval does not treat Celsius/Fahrenheit as ordinary scalar units")
	}
	return units.Expression{}, fmt.Errorf("ambiguous unit %q; use a long unit name", text)
}

func (p *parser) readName() string {
	start := p.pos
	for p.pos < len(p.s) && isNamePart(rune(p.s[p.pos])) {
		p.pos++
	}
	return p.s[start:p.pos]
}

func (p *parser) skipSpace() {
	for p.pos < len(p.s) && unicode.IsSpace(rune(p.s[p.pos])) {
		p.pos++
	}
}

func isNumberStart(ch byte) bool {
	return (ch >= '0' && ch <= '9') || ch == '.'
}

func isNameStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == 'µ' || ch == 'Ω'
}

func isNamePart(ch rune) bool {
	return isNameStart(ch) || unicode.IsDigit(ch) || ch == '-'
}

func isUnitStart(ch rune) bool {
	return isNameStart(ch)
}

func isUnitSuffixStop(ch rune) bool {
	return unicode.IsSpace(ch) || ch == '+' || ch == '-' || ch == ')' || ch == '('
}

func isEvalConstant(name string) bool {
	return name == "pi" || name == "c" || name == "G" || name == "g0"
}

func temperatureLookingMismatch(left, right quantity) bool {
	return (left.temperatureLookingLiteral && right.dimension == temperatureDimension) ||
		(right.temperatureLookingLiteral && left.dimension == temperatureDimension)
}

func temperatureLookingOutput(output string) bool {
	output = strings.TrimSpace(output)
	return output == "K" || output == "C" || output == "F" || output == "celsius" || output == "fahrenheit"
}
