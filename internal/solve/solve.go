package solve

import (
	"fmt"
	"math"

	"convunits/internal/units"
)

type Interval struct {
	Min float64
	Max float64
}

func NewInterval(min, max float64) (Interval, error) {
	if !isFinite(min) || !isFinite(max) {
		return Interval{}, fmt.Errorf("values must be finite")
	}
	if min > max {
		return Interval{}, fmt.Errorf("range minimum %g exceeds maximum %g", min, max)
	}
	if min <= 0 {
		return Interval{}, fmt.Errorf("force relationship values must be greater than zero")
	}
	return Interval{Min: min, Max: max}, nil
}

func (i Interval) Scalar() bool { return i.Min == i.Max }

type Quantity struct {
	Range Interval
	Unit  string
}

type Result struct {
	Range    Interval
	Unit     string
	Variable string
}

type Solver struct{ Registry *units.Registry }

func New(registry *units.Registry) *Solver { return &Solver{Registry: registry} }

var variableDimensions = map[string]units.Dimension{
	"force":    {Mass: 1, Length: 1, Time: -2},
	"mass":     {Mass: 1},
	"distance": {Length: 1},
	"time":     {Time: 1},
}

var canonicalUnits = map[string]string{
	"force": "N", "mass": "kg", "distance": "m", "time": "s",
}

// Solve applies F=m*d/t^2 and solves for the target variable. The primary
// input and each named given must describe distinct variables.
func (s *Solver) Solve(input Quantity, outputUnit string, givens map[string]Quantity) (Result, error) {
	inputVariable, err := s.variableForUnit(input.Unit)
	if err != nil {
		return Result{}, fmt.Errorf("input: %w", err)
	}
	targetVariable, err := s.variableForUnit(outputUnit)
	if err != nil {
		return Result{}, fmt.Errorf("target: %w", err)
	}
	if inputVariable == targetVariable {
		return Result{}, fmt.Errorf("input already has the target dimension; use direct conversion instead")
	}

	known := make(map[string]Interval)
	canonical, err := s.toCanonical(inputVariable, input)
	if err != nil {
		return Result{}, fmt.Errorf("input: %w", err)
	}
	known[inputVariable] = canonical

	for name, quantity := range givens {
		if _, ok := variableDimensions[name]; !ok {
			return Result{}, fmt.Errorf("unknown given %q; expected force, mass, distance, or time", name)
		}
		actual, err := s.variableForUnit(quantity.Unit)
		if err != nil {
			return Result{}, fmt.Errorf("given %s: %w", name, err)
		}
		if actual != name {
			return Result{}, fmt.Errorf("given %s has %s dimensions", name, actual)
		}
		if name == targetVariable {
			return Result{}, fmt.Errorf("target variable %s must not also be supplied", name)
		}
		if _, exists := known[name]; exists {
			return Result{}, fmt.Errorf("%s was supplied more than once", name)
		}
		canonical, err := s.toCanonical(name, quantity)
		if err != nil {
			return Result{}, fmt.Errorf("given %s: %w", name, err)
		}
		known[name] = canonical
	}

	for name := range variableDimensions {
		if name != targetVariable {
			if _, ok := known[name]; !ok {
				return Result{}, fmt.Errorf("missing --given %s=VALUEUNIT", name)
			}
		}
	}

	base := solveInterval(targetVariable, known)
	converted, err := s.fromCanonical(targetVariable, base, outputUnit)
	if err != nil {
		return Result{}, err
	}
	return Result{Range: converted, Unit: outputUnit, Variable: targetVariable}, nil
}

func (s *Solver) variableForUnit(unit string) (string, error) {
	expressions, err := s.Registry.ParseCandidates(unit)
	if err != nil {
		return "", err
	}
	var found string
	for _, expression := range expressions {
		for name, dimension := range variableDimensions {
			if expression.Dimension == dimension {
				if found != "" && found != name {
					return "", fmt.Errorf("ambiguous dimensions for %q", unit)
				}
				found = name
			}
		}
	}
	if found == "" {
		return "", fmt.Errorf("unit %q is not force, mass, distance, or time", unit)
	}
	return found, nil
}

func (s *Solver) toCanonical(variable string, quantity Quantity) (Interval, error) {
	min, err := s.Registry.Convert(quantity.Range.Min, quantity.Unit, canonicalUnits[variable])
	if err != nil {
		return Interval{}, err
	}
	max, err := s.Registry.Convert(quantity.Range.Max, quantity.Unit, canonicalUnits[variable])
	if err != nil {
		return Interval{}, err
	}
	return NewInterval(min.Value, max.Value)
}

func (s *Solver) fromCanonical(variable string, value Interval, outputUnit string) (Interval, error) {
	min, err := s.Registry.Convert(value.Min, canonicalUnits[variable], outputUnit)
	if err != nil {
		return Interval{}, err
	}
	max, err := s.Registry.Convert(value.Max, canonicalUnits[variable], outputUnit)
	if err != nil {
		return Interval{}, err
	}
	return NewInterval(min.Value, max.Value)
}

func solveInterval(target string, v map[string]Interval) Interval {
	switch target {
	case "time":
		return Interval{
			Min: math.Sqrt(v["mass"].Min * v["distance"].Min / v["force"].Max),
			Max: math.Sqrt(v["mass"].Max * v["distance"].Max / v["force"].Min),
		}
	case "force":
		return Interval{
			Min: v["mass"].Min * v["distance"].Min / square(v["time"].Max),
			Max: v["mass"].Max * v["distance"].Max / square(v["time"].Min),
		}
	case "mass":
		return Interval{
			Min: v["force"].Min * square(v["time"].Min) / v["distance"].Max,
			Max: v["force"].Max * square(v["time"].Max) / v["distance"].Min,
		}
	default: // distance
		return Interval{
			Min: v["force"].Min * square(v["time"].Min) / v["mass"].Max,
			Max: v["force"].Max * square(v["time"].Max) / v["mass"].Min,
		}
	}
}

func square(v float64) float64 { return v * v }
func isFinite(v float64) bool  { return !math.IsNaN(v) && !math.IsInf(v, 0) }
