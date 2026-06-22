package units

import (
	"fmt"
	"math"
)

type Result struct {
	Value         float64
	Input, Output Expression
	Approximate   bool
	Category      string
}

func (r *Registry) Convert(value float64, input, output string) (Result, error) {
	if isFuelConsumption(input) || isFuelConsumption(output) {
		return r.convertFuel(value, input, output)
	}
	return r.convertDirect(value, input, output)
}

func (r *Registry) convertDirect(value float64, input, output string) (Result, error) {
	ins, err := r.ParseCandidates(input)
	if err != nil {
		return Result{}, fmt.Errorf("input unit: %w", err)
	}
	outs, err := r.ParseCandidates(output)
	if err != nil {
		return Result{}, fmt.Errorf("output unit: %w", err)
	}
	var matches []Result
	for _, in := range ins {
		for _, out := range outs {
			if in.Dimension == out.Dimension {
				converted := out.FromBase(in.ToBase(value))
				if out.Affine != nil && math.Abs(converted) < 1e-12 {
					converted = 0
				}
				matches = append(matches, Result{Value: converted, Input: in, Output: out, Category: in.Dimension.Category()})
			}
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return Result{}, fmt.Errorf("ambiguous conversion from %q to %q; use long unit names", input, output)
	}
	if len(ins) > 1 && len(outs) > 1 {
		return Result{}, fmt.Errorf("ambiguous input unit %q and output unit %q; use long unit names", input, output)
	}
	if len(ins) > 1 {
		return Result{}, fmt.Errorf("ambiguous input unit %q; use a long unit name", input)
	}
	if len(outs) > 1 {
		return Result{}, fmt.Errorf("ambiguous output unit %q; use a long unit name", output)
	}
	return Result{}, IncompatibleError{input, output, ins[0].Dimension, outs[0].Dimension}
}

const litersPer100KilometersScale = 1e-8 // (1 L)/(100 km), expressed as m^2.

func isFuelConsumption(unit string) bool {
	return unit == "L/100km" || unit == "l/100km"
}

func (r *Registry) convertFuel(value float64, input, output string) (Result, error) {
	consumptionDimension := Dimension{Length: 2}
	economyDimension := Dimension{Length: -2}
	consumption := Expression{Text: "L/100km", Dimension: consumptionDimension, Multiplier: litersPer100KilometersScale}

	if isFuelConsumption(input) && isFuelConsumption(output) {
		return Result{Value: value, Input: consumption, Output: consumption, Category: "fuel consumption"}, nil
	}
	if value <= 0 {
		return Result{}, fmt.Errorf("fuel economy and consumption values must be greater than zero")
	}

	if isFuelConsumption(input) {
		outs, err := r.ParseCandidates(output)
		if err != nil {
			return Result{}, fmt.Errorf("output unit: %w", err)
		}
		for _, out := range outs {
			if out.Dimension == economyDimension {
				baseEconomy := 1 / (value * litersPer100KilometersScale)
				return Result{Value: out.FromBase(baseEconomy), Input: consumption, Output: out, Category: "fuel economy"}, nil
			}
		}
		return Result{}, fmt.Errorf("cannot convert L/100km to %s: output must be a distance-per-volume fuel economy unit", output)
	}

	ins, err := r.ParseCandidates(input)
	if err != nil {
		return Result{}, fmt.Errorf("input unit: %w", err)
	}
	for _, in := range ins {
		if in.Dimension == economyDimension {
			baseEconomy := in.ToBase(value)
			converted := (1 / baseEconomy) / litersPer100KilometersScale
			return Result{Value: converted, Input: in, Output: consumption, Category: "fuel consumption"}, nil
		}
	}
	return Result{}, fmt.Errorf("cannot convert %s to L/100km: input must be a distance-per-volume fuel economy unit", input)
}
