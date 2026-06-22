package formula

import (
	"fmt"
	"math"
	"sort"

	"convunits/internal/units"
)

const (
	gravitationalConstant = 6.67430e-11
	standardGravity       = 9.80665
)

type Input struct {
	Value float64
	Unit  string
}

type Result struct {
	Value       float64
	Unit        string
	Approximate bool
}

type Definition struct {
	Name, Arguments, OutputDimension string
}

var definitions = []Definition{
	{"bmi", "--mass MASS --height HEIGHT", "mass/length^2"},
	{"escape-velocity", "--mass MASS --radius RADIUS", "speed"},
	{"freefall-time", "--height HEIGHT [--gravity ACCELERATION]", "time"},
	{"orbital-period", "--mass MASS --radius RADIUS", "time"},
	{"orbital-speed", "--mass MASS --radius RADIUS", "speed"},
	{"pendulum-period", "--length LENGTH [--gravity ACCELERATION]", "time"},
}

func Definitions() []Definition {
	out := append([]Definition(nil), definitions...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

type Calculator struct{ Registry *units.Registry }

func New(registry *units.Registry) *Calculator { return &Calculator{Registry: registry} }

func (c *Calculator) Compute(name string, inputs map[string]Input, outputUnit string) (Result, error) {
	allowedByFormula := map[string]map[string]bool{
		"escape-velocity": {"mass": true, "radius": true},
		"orbital-period":  {"mass": true, "radius": true},
		"orbital-speed":   {"mass": true, "radius": true},
		"freefall-time":   {"height": true, "gravity": true},
		"pendulum-period": {"length": true, "gravity": true},
		"bmi":             {"mass": true, "height": true},
	}
	allowed, known := allowedByFormula[name]
	if !known {
		return Result{}, fmt.Errorf("unknown formula %q", name)
	}
	for argument := range inputs {
		if !allowed[argument] {
			return Result{}, fmt.Errorf("formula %s does not accept --%s", name, argument)
		}
	}
	var value float64
	var canonicalOutput string
	approximate := true
	switch name {
	case "escape-velocity":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = math.Sqrt(2 * gravitationalConstant * mass / radius)
		canonicalOutput = "m/s"
	case "orbital-period":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = 2 * math.Pi * math.Sqrt(radius*radius*radius/(gravitationalConstant*mass))
		canonicalOutput = "s"
	case "orbital-speed":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = math.Sqrt(gravitationalConstant * mass / radius)
		canonicalOutput = "m/s"
	case "freefall-time":
		height, err := c.required(inputs, "height", "m")
		if err != nil {
			return Result{}, err
		}
		gravity, err := c.optional(inputs, "gravity", "m/s^2", standardGravity)
		if err != nil {
			return Result{}, err
		}
		value = math.Sqrt(2 * height / gravity)
		canonicalOutput = "s"
	case "pendulum-period":
		length, err := c.required(inputs, "length", "m")
		if err != nil {
			return Result{}, err
		}
		gravity, err := c.optional(inputs, "gravity", "m/s^2", standardGravity)
		if err != nil {
			return Result{}, err
		}
		value = 2 * math.Pi * math.Sqrt(length/gravity)
		canonicalOutput = "s"
	case "bmi":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		height, err := c.required(inputs, "height", "m")
		if err != nil {
			return Result{}, err
		}
		value = mass / (height * height)
		canonicalOutput = "kg/m^2"
		approximate = false
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Result{}, fmt.Errorf("formula %s produced a non-finite result", name)
	}
	requested := outputUnit
	convertUnit := outputUnit
	if outputUnit == "bmi" {
		if name != "bmi" {
			return Result{}, fmt.Errorf("output alias bmi is only valid for the bmi formula")
		}
		convertUnit = "kg/m^2"
	}
	converted, err := c.Registry.Convert(value, canonicalOutput, convertUnit)
	if err != nil {
		return Result{}, fmt.Errorf("output unit: %w", err)
	}
	return Result{Value: converted.Value, Unit: requested, Approximate: approximate}, nil
}

func (c *Calculator) required(inputs map[string]Input, name, canonical string) (float64, error) {
	input, ok := inputs[name]
	if !ok {
		return 0, fmt.Errorf("missing required argument --%s", name)
	}
	return c.convertPositive(name, input, canonical)
}

func (c *Calculator) optional(inputs map[string]Input, name, canonical string, fallback float64) (float64, error) {
	input, ok := inputs[name]
	if !ok {
		return fallback, nil
	}
	return c.convertPositive(name, input, canonical)
}

func (c *Calculator) convertPositive(name string, input Input, canonical string) (float64, error) {
	converted, err := c.Registry.Convert(input.Value, input.Unit, canonical)
	if err != nil {
		return 0, fmt.Errorf("--%s: %w", name, err)
	}
	if converted.Value <= 0 {
		return 0, fmt.Errorf("--%s must be greater than zero", name)
	}
	return converted.Value, nil
}
